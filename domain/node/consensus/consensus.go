package consensus

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/haomiao000/raftchain/domain/node/rpc/raft"
	"github.com/haomiao000/raftchain/library/model"
	"github.com/haomiao000/raftchain/library/query"
	"github.com/haomiao000/raftchain/library/resource"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm/clause"
)

type NodeState int

const (
	Follower NodeState = iota
	Candidate
	Leader
)

const (
	heartbeatInterval      = 100 * time.Millisecond
	electionTimeoutMin     = 250
	electionTimeoutMax     = 400
	blockPackagingInterval = 2 * time.Second
)

type ConsensusNode struct {
	raft.UnimplementedRaftServiceServer
	mu          sync.Mutex
	id          int
	peer        map[int]string
	peerClients map[int]raft.RaftServiceClient
	state       NodeState
	currentTerm int
	votedFor    int

	commitIndex int
	lastApplied int
	log         []*raft.Block

	nextIndex  map[int]int
	matchIndex map[int]int

	electionTimer *time.Timer
	blockPacker   *time.Ticker

	txPool []*raft.Transaction
}

func NewConsensusNode(id int, peer map[int]string) *ConsensusNode {
	node := &ConsensusNode{
		id:          id,
		peer:        peer,
		peerClients: make(map[int]raft.RaftServiceClient),
		log:         []*raft.Block{{Term: 0}},
		commitIndex: 0,
		lastApplied: 0,
		state:       Follower,
		currentTerm: 0,
		votedFor:    -1,
		txPool:      make([]*raft.Transaction, 0),
	}
	return node
}

// applyLogs_nl 将已提交但未应用的日志写入数据库
func (cn *ConsensusNode) applyLogs_nl() {
	q := query.Use(resource.GormServe)

	for cn.lastApplied < cn.commitIndex {
		cn.lastApplied++
		appliedIndex := cn.lastApplied
		blockToApply := cn.log[appliedIndex]

		cn.logEventToDBAsync("INFO", fmt.Sprintf("开始应用区块 %d 到状态机", appliedIndex))

		var transactions []*raft.Transaction
		if len(blockToApply.Data) > 0 {
			if err := json.Unmarshal(blockToApply.Data, &transactions); err != nil {
				cn.logEventToDBAsync("ERROR", fmt.Sprintf("反序列化区块 %d 中的交易失败: %v", appliedIndex, err))
				continue
			}
		}

		err := q.Transaction(func(tx *query.Query) error {
			dbBlock := &model.Block{
				Height:         int64(appliedIndex),
				Term:           blockToApply.Term,
				Hash:           blockToApply.Hash,
				PrevHash:       blockToApply.PrevHash,
				Timestamp:      time.Now(),
				ProposerNodeID: 0,
				Data:           blockToApply.Data,
			}
			if err := tx.Block.Create(dbBlock); err != nil {
				return fmt.Errorf("创建区块记录失败: %w", err)
			}

			if len(transactions) > 0 {
				dbTransactions := make([]*model.Tx, len(transactions))
				for i, tr := range transactions {
					txHashStr := fmt.Sprintf("%s%d", tr.SenderPublicKey, time.Now().UnixNano())
					hash := sha256.Sum256([]byte(txHashStr))
					dbTransactions[i] = &model.Tx{
						TxHash:          hex.EncodeToString(hash[:]),
						BlockHeight:     int64(appliedIndex),
						BlockHash:       blockToApply.Hash,
						SenderPublicKey: tr.SenderPublicKey,
						Type:            tr.Type,
						Payload:         tr.Payload,
					}
				}
				if err := tx.Tx.CreateInBatches(dbTransactions, 100); err != nil {
					return fmt.Errorf("创建交易记录失败: %w", err)
				}
			}
			return nil
		})

		if err != nil {
			cn.logEventToDBAsync("ERROR", fmt.Sprintf("应用区块 %d 到数据库失败: %v。应用中止！", appliedIndex, err))
			cn.lastApplied--
			return
		}

		cn.logEventToDBAsync("SUCCESS", fmt.Sprintf("成功应用并持久化区块 %d，包含 %d 条交易", appliedIndex, len(transactions)))
	}
}

func (cn *ConsensusNode) ConnectToPeers() {
	cn.mu.Lock()
	defer cn.mu.Unlock()
	for peerID, addr := range cn.peer {
		if peerID != cn.id {
			conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				// 这种基础连接日志保留在标准输出
				log.Printf("[Node %d] 连接 gRPC 节点 %d (%s) 失败: %v。稍后将重试。\n", cn.id, peerID, addr, err)
				continue
			}
			cn.peerClients[peerID] = raft.NewRaftServiceClient(conn)
			cn.logEventToDBAsync("SUCCESS", fmt.Sprintf("node%v成功连接到node%d (%s)", cn.id, peerID, addr))
		}
	}
}

func (cn *ConsensusNode) Run() {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			cn.updateNodeStatusInDBAsync()
		}
	}()
	cn.resetElectionTimer()
}

func (cn *ConsensusNode) Propose(data []byte) (bool, int) {
	cn.mu.Lock()
	defer cn.mu.Unlock()

	if cn.state != Leader {
		return false, -1
	}

	lastLogIndex, _ := cn.getLastLogInfo_nl()
	prevHash := ""
	if lastLogIndex > 0 {
		prevHash = cn.log[lastLogIndex].Hash
	}

	newBlock := &raft.Block{
		Term:     int64(cn.currentTerm),
		Data:     data,
		PrevHash: prevHash,
	}
	hashStr := fmt.Sprintf("%d%s%s", newBlock.Term, string(newBlock.Data), newBlock.PrevHash)
	hash := sha256.Sum256([]byte(hashStr))
	newBlock.Hash = hex.EncodeToString(hash[:])

	cn.log = append(cn.log, newBlock)
	cn.logEventToDBAsync("INFO", fmt.Sprintf("提议新区块 (index: %d, term: %d)。日志长度: %d", len(cn.log)-1, cn.currentTerm, len(cn.log)))

	cn.matchIndex[cn.id] = len(cn.log) - 1
	cn.nextIndex[cn.id] = len(cn.log)

	return true, cn.currentTerm
}

func (cn *ConsensusNode) becomeFollower_nl(term int) {
	if cn.state == Leader {
		if cn.blockPacker != nil {
			cn.blockPacker.Stop()
			cn.logEventToDBAsync("WARNING", "Leader 状态终止，停止打包区块")
		}
	}
	cn.state = Follower
	cn.currentTerm = term
	cn.votedFor = -1
	cn.logEventToDBAsync("INFO", fmt.Sprintf("node%v转变为 Follower，任期为 %d",cn.id, cn.currentTerm))
	cn.resetElectionTimer_nl()
	go cn.updateNodeStatusInDBAsync()
}

func (cn *ConsensusNode) resetElectionTimer_nl() {
	if cn.electionTimer != nil {
		cn.electionTimer.Stop()
	}
	timeout := time.Duration(electionTimeoutMin+rand.Intn(electionTimeoutMax-electionTimeoutMin)) * time.Millisecond
	cn.electionTimer = time.AfterFunc(timeout, cn.startElection)
}

func (cn *ConsensusNode) resetElectionTimer() {
	cn.mu.Lock()
	defer cn.mu.Unlock()
	cn.resetElectionTimer_nl()
}

func (cn *ConsensusNode) startElection() {
	cn.mu.Lock()
	if cn.state == Leader {
		cn.mu.Unlock()
		return
	}
	cn.state = Candidate
	cn.currentTerm++
	cn.votedFor = cn.id
	votes := 1
	cn.logEventToDBAsync("INFO", fmt.Sprintf("选举计时器超时，开始为任期 %d 进行选举", cn.currentTerm))

	cn.resetElectionTimer_nl()

	lastLogIndex, lastLogTerm := cn.getLastLogInfo_nl()
	args := &raft.RequestVoteArgs{
		Term:         int32(cn.currentTerm),
		CandidateId:  int32(cn.id),
		LastLogIndex: int32(lastLogIndex),
		LastLogTerm:  int32(lastLogTerm),
	}
	currentTerm := cn.currentTerm
	cn.mu.Unlock()

	for peerID := range cn.peer {
		if peerID != cn.id {
			go func(peerID int) {
				reply, err := cn.sendRequestVote(peerID, args)
				if err != nil {
					return
				}

				cn.mu.Lock()
				defer cn.mu.Unlock()

				if cn.state != Candidate || cn.currentTerm != currentTerm {
					return
				}

				if int(reply.Term) > cn.currentTerm {
					cn.becomeFollower_nl(int(reply.Term))
					return
				}

				if reply.VoteGranted {
					votes++
					cn.logEventToDBAsync("INFO", fmt.Sprintf("获得来自节点 %d 的选票，总票数: %d", peerID, votes))
					if votes > len(cn.peer)/2 {
						cn.becomeLeader_nl()
					}
				}
			}(peerID)
		}
	}
}

func (cn *ConsensusNode) becomeLeader_nl() {
	if cn.state != Candidate {
		return
	}
	cn.state = Leader
	if cn.electionTimer != nil {
		cn.electionTimer.Stop()
	}

	cn.nextIndex = make(map[int]int)
	cn.matchIndex = make(map[int]int)
	lastLogIndex, _ := cn.getLastLogInfo_nl()
	for peerID := range cn.peer {
		cn.nextIndex[peerID] = lastLogIndex + 1
		cn.matchIndex[peerID] = 0
	}
	cn.matchIndex[cn.id] = lastLogIndex

	cn.logEventToDBAsync("SUCCESS", fmt.Sprintf("node%v 当选，成为任期 %d 的 Leader！", cn.id, cn.currentTerm))

	for peerID := range cn.peer {
		if peerID != cn.id {
			go cn.replicationLoopForPeer(peerID)
		}
	}

	cn.blockPacker = time.NewTicker(blockPackagingInterval)
	go cn.blockPackingLoop()
	go cn.updateNodeStatusInDBAsync()
}

func (cn *ConsensusNode) blockPackingLoop() {
	for {
		cn.mu.Lock()
		if cn.state != Leader {
			cn.blockPacker.Stop()
			cn.mu.Unlock()
			return
		}
		cn.mu.Unlock()

		<-cn.blockPacker.C

		cn.mu.Lock()
		if len(cn.txPool) == 0 {
			cn.mu.Unlock()
			continue
		}

		txsToPack := make([]*raft.Transaction, len(cn.txPool))
		copy(txsToPack, cn.txPool)
		cn.txPool = make([]*raft.Transaction, 0)
		cn.mu.Unlock()

		cn.logEventToDBAsync("INFO", fmt.Sprintf("打包 %d 条交易到新区块中", len(txsToPack)))
		data, err := json.Marshal(txsToPack)
		if err != nil {
			cn.logEventToDBAsync("ERROR", fmt.Sprintf("序列化交易失败: %v", err))
			continue
		}
		cn.Propose(data)
	}
}

func (cn *ConsensusNode) replicationLoopForPeer(peerID int) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		cn.mu.Lock()
		if cn.state != Leader {
			cn.mu.Unlock()
			return
		}

		nextIdx := cn.nextIndex[peerID]
		prevLogIndex := nextIdx - 1

		if prevLogIndex < 0 {
			cn.logEventToDBAsync("ERROR", fmt.Sprintf("对节点 %d 的 prevLogIndex 为 %d，这是异常情况！", peerID, prevLogIndex))
			cn.mu.Unlock()
			time.Sleep(1 * time.Second)
			continue
		}

		prevLogTerm := cn.log[prevLogIndex].Term
		var entriesToSend []*raft.Block
		if len(cn.log)-1 >= nextIdx {
			entriesToSend = cn.log[nextIdx:]
		}

		args := &raft.AppendEntriesArgs{
			Term:         int32(cn.currentTerm),
			LeaderId:     int32(cn.id),
			PrevLogIndex: int32(prevLogIndex),
			PrevLogTerm:  int32(prevLogTerm),
			Entries:      entriesToSend,
			LeaderCommit: int32(cn.commitIndex),
		}
		currentTerm := cn.currentTerm
		cn.mu.Unlock()

		reply, err := cn.sendAppendEntries(peerID, args)

		cn.mu.Lock()
		if cn.state != Leader || cn.currentTerm != currentTerm {
			cn.mu.Unlock()
			return
		}

		if err != nil {
			cn.mu.Unlock()
			<-ticker.C
			continue
		}

		if reply.Success {
			cn.matchIndex[peerID] = prevLogIndex + len(entriesToSend)
			cn.nextIndex[peerID] = cn.matchIndex[peerID] + 1
			cn.updateCommitIndex_nl()
			cn.mu.Unlock()
			<-ticker.C
		} else {
			if int(reply.Term) > cn.currentTerm {
				cn.becomeFollower_nl(int(reply.Term))
				cn.mu.Unlock()
				return
			}

			if cn.nextIndex[peerID] > 1 {
				cn.nextIndex[peerID]--
				cn.logEventToDBAsync("WARNING", fmt.Sprintf("节点 %d 拒绝 AppendEntries，重试 nextIndex = %d", peerID, cn.nextIndex[peerID]))
			} else {
				cn.logEventToDBAsync("WARNING", fmt.Sprintf("节点 %d 拒绝 AppendEntries，且 nextIndex 已为1，暂停复制", peerID))
			}
			cn.mu.Unlock()
		}
	}
}

func (cn *ConsensusNode) updateCommitIndex_nl() {
	majority := len(cn.peer)/2 + 1

	matchIndexes := make([]int, 0, len(cn.peer))
	for _, mi := range cn.matchIndex {
		matchIndexes = append(matchIndexes, mi)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(matchIndexes)))

	newCommitIndex := matchIndexes[majority-1]

	if newCommitIndex > cn.commitIndex && cn.log[newCommitIndex].Term == int64(cn.currentTerm) {
		cn.logEventToDBAsync("SUCCESS", fmt.Sprintf("多数节点已匹配，推进 commitIndex 从 %d 到 %d", cn.commitIndex, newCommitIndex))
		cn.commitIndex = newCommitIndex
		cn.applyLogs_nl()
	}
}

func (cn *ConsensusNode) RequestVote(ctx context.Context, args *raft.RequestVoteArgs) (*raft.RequestVoteReply, error) {
	cn.mu.Lock()
	defer cn.mu.Unlock()

	reply := &raft.RequestVoteReply{Term: int32(cn.currentTerm), VoteGranted: false}

	if int(args.Term) < cn.currentTerm {
		return reply, nil
	}

	if int(args.Term) > cn.currentTerm {
		cn.becomeFollower_nl(int(args.Term))
	}

	myLastLogIndex, myLastLogTerm := cn.getLastLogInfo_nl()
	isLogOk := int(args.LastLogTerm) > myLastLogTerm || (int(args.LastLogTerm) == myLastLogTerm && int(args.LastLogIndex) >= myLastLogIndex)

	if (cn.votedFor == -1 || cn.votedFor == int(args.CandidateId)) && isLogOk {
		cn.votedFor = int(args.CandidateId)
		reply.VoteGranted = true
		cn.logEventToDBAsync("INFO", fmt.Sprintf("node%v投票给节点 %d (任期 %d)",cn.id, args.CandidateId, cn.currentTerm))
		cn.resetElectionTimer_nl()
	} else {
		cn.logEventToDBAsync("INFO", fmt.Sprintf("node%v拒绝为节点 %d 投票 (votedFor=%d, isLogOk=%t)",cn.id, args.CandidateId, cn.votedFor, isLogOk))
	}

	reply.Term = int32(cn.currentTerm)
	return reply, nil
}

func (cn *ConsensusNode) AppendEntries(ctx context.Context, args *raft.AppendEntriesArgs) (*raft.AppendEntriesReply, error) {
	cn.mu.Lock()
	defer cn.mu.Unlock()

	reply := &raft.AppendEntriesReply{Term: int32(cn.currentTerm), Success: false}

	if int(args.Term) < cn.currentTerm {
		return reply, nil
	}
	cn.resetElectionTimer_nl()

	if int(args.Term) > cn.currentTerm || cn.state == Candidate {
		cn.becomeFollower_nl(int(args.Term))
	}

	reply.Term = int32(cn.currentTerm)

	lastLogIndex, _ := cn.getLastLogInfo_nl()
	if int(args.PrevLogIndex) > lastLogIndex {
		cn.logEventToDBAsync("WARNING", fmt.Sprintf("拒绝 Leader %d 的 AppendEntries: PrevLogIndex %d 超出本地日志范围 (本地最新: %d)", args.LeaderId, args.PrevLogIndex, lastLogIndex))
		return reply, nil
	}

	if cn.log[args.PrevLogIndex].Term != int64(args.PrevLogTerm) {
		cn.logEventToDBAsync("WARNING", fmt.Sprintf("拒绝 Leader %d 的 AppendEntries: 在 index %d 上任期不匹配 (我的: %d, Leader的: %d)", args.LeaderId, args.PrevLogIndex, cn.log[args.PrevLogIndex].Term, args.PrevLogTerm))
		return reply, nil
	}

	if len(args.Entries) > 0 {
		conflictIndex := -1
		for i, entry := range args.Entries {
			logIndex := int(args.PrevLogIndex) + 1 + i
			if logIndex > len(cn.log)-1 || cn.log[logIndex].Term != entry.Term {
				conflictIndex = logIndex
				break
			}
		}

		if conflictIndex != -1 {
			cn.log = cn.log[:conflictIndex]
			entriesToAppend := args.Entries[conflictIndex-(int(args.PrevLogIndex)+1):]
			cn.log = append(cn.log, entriesToAppend...)
		}
		cn.logEventToDBAsync("INFO", fmt.Sprintf("从 Leader %d 同步日志。当前日志长度: %d", args.LeaderId, len(cn.log)))
	}

	if args.LeaderCommit > int32(cn.commitIndex) {
		lastNewEntryIndex, _ := cn.getLastLogInfo_nl()
		cn.commitIndex = int(math.Min(float64(args.LeaderCommit), float64(lastNewEntryIndex)))
		cn.applyLogs_nl()
	}

	reply.Success = true
	return reply, nil
}

func (cn *ConsensusNode) SubmitTransaction(ctx context.Context, tx *raft.Transaction) (*raft.TransactionReply, error) {
	cn.mu.Lock()
	defer cn.mu.Unlock()

	if cn.state != Leader {
		return &raft.TransactionReply{Success: false, LeaderHint: ""}, nil
	}

	cn.txPool = append(cn.txPool, tx)
	cn.logEventToDBAsync("INFO", fmt.Sprintf("收到新交易，交易池大小变为 %d", len(cn.txPool)))

	txHashStr := fmt.Sprintf("%s%s", tx.SenderPublicKey, time.Now().String())
	hash := sha256.Sum256([]byte(txHashStr))

	return &raft.TransactionReply{
		Success: true,
		TxHash:  hex.EncodeToString(hash[:]),
	}, nil
}

func (cn *ConsensusNode) GetBlocks(ctx context.Context, args *raft.GetBlocksArgs) (*raft.GetBlocksReply, error) {
	cn.mu.Lock()
	defer cn.mu.Unlock()

	lastIndex := len(cn.log) - 1
	if lastIndex == 0 {
		return &raft.GetBlocksReply{Blocks: []*raft.Block{}, Pagination: &raft.Pagination{TotalItems: 0}}, nil
	}

	page := int(args.Page)
	limit := int(args.Limit)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	totalItems := int64(lastIndex)
	totalPages := int32(math.Ceil(float64(totalItems) / float64(limit)))

	startIndex := lastIndex - (page-1)*limit
	if startIndex < 1 {
		startIndex = 0
	}

	endIndex := startIndex - limit
	if endIndex < 0 {
		endIndex = 0
	}

	var blocksForAPI []*raft.Block
	for i := startIndex; i > endIndex && i > 0; i-- {
		blocksForAPI = append(blocksForAPI, cn.log[i])
	}

	return &raft.GetBlocksReply{
		Blocks: blocksForAPI,
		Pagination: &raft.Pagination{
			CurrentPage: int32(page),
			TotalPages:  totalPages,
			TotalItems:  totalItems,
			PerPage:     int32(limit),
		},
	}, nil
}

func (cn *ConsensusNode) sendRequestVote(peerID int, args *raft.RequestVoteArgs) (*raft.RequestVoteReply, error) {
	cn.mu.Lock()
	client, ok := cn.peerClients[peerID]
	cn.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("no client connection for peer %d", peerID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	return client.RequestVote(ctx, args)
}

func (cn *ConsensusNode) sendAppendEntries(peerID int, args *raft.AppendEntriesArgs) (*raft.AppendEntriesReply, error) {
	cn.mu.Lock()
	client, ok := cn.peerClients[peerID]
	cn.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("no client connection for peer %d", peerID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), heartbeatInterval)
	defer cancel()
	return client.AppendEntries(ctx, args)
}

func (cn *ConsensusNode) getLastLogInfo_nl() (int, int) {
	lastIndex := len(cn.log) - 1
	lastTerm := int(cn.log[lastIndex].Term)
	return lastIndex, lastTerm
}

func (cn *ConsensusNode) IsLeader() (bool, int) {
	cn.mu.Lock()
	defer cn.mu.Unlock()
	return cn.state == Leader, cn.currentTerm
}

func (cn *ConsensusNode) updateNodeStatusInDBAsync() {
	cn.mu.Lock()
	var roleStr string
	switch cn.state {
	case Follower:
		roleStr = "Follower"
	case Candidate:
		roleStr = "Candidate"
	case Leader:
		roleStr = "Leader"
	}

	nodeInfo := &model.Node{
		NodeID:        int32(cn.id),
		Address:       cn.peer[cn.id],
		Role:          roleStr,
		Term:          int32(cn.currentTerm),
		CommitIndex:   int32(cn.commitIndex),
		LastApplied:   int32(cn.lastApplied),
		Status:        "active",
		LastHeartbeat: time.Now(),
	}
	cn.mu.Unlock()

	go func() {
		q := query.Use(resource.GormServe)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := q.Node.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "node_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"role", "term", "commit_index", "last_applied", "status", "last_heartbeat", "address"}),
		}).Create(nodeInfo)

		if err != nil {
			log.Printf("[Node %d] (async) Failed to upsert node status to DB: %v", nodeInfo.NodeID, err)
		}
	}()
}

// logEventToDBAsync 异步地将事件记录到数据库，不会阻塞主流程
func (cn *ConsensusNode) logEventToDBAsync(level, message string) {
	nodeID := cn.id
	go func() {
		if resource.GormServe == nil {
			log.Printf("[Node %d] Gorm DB not initialized. DB Log skipped: %s - %s", nodeID, level, message)
			return
		}
		q := query.Use(resource.GormServe)
		logEntry := &model.SystemLog{
			Timestamp: time.Now(),
			NodeID:    int32(nodeID),
			Level:     level,
			Message:   message,
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := q.SystemLog.WithContext(ctx).Create(logEntry); err != nil {
			// 如果数据库日志自身失败，回退到标准输出
			log.Printf("[Node %d] (async) Failed to write event log to DB: %v", nodeID, err)
		}
	}()
}