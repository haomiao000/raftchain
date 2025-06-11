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
)

type NodeState int

const (
	Follower  NodeState = iota
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
	// 直接使用全局的 resource.GormServe
	q := query.Use(resource.GormServe)

	for cn.lastApplied < cn.commitIndex {
		cn.lastApplied++
		appliedIndex := cn.lastApplied
		blockToApply := cn.log[appliedIndex]

		log.Printf("[Node %d] Applying block %d to state machine...\n", cn.id, appliedIndex)

		var transactions []*raft.Transaction
		if len(blockToApply.Data) > 0 {
			if err := json.Unmarshal(blockToApply.Data, &transactions); err != nil {
				log.Printf("[Node %d] ERROR: Failed to unmarshal transactions in block %d: %v. Skipping application.", cn.id, appliedIndex, err)
				continue
			}
		}

		// 使用数据库事务保证原子性
		err := q.Transaction(func(tx *query.Query) error {
			// 创建区块记录
			dbBlock := &model.Block{
				Height:         int64(appliedIndex),
				Term:           blockToApply.Term,
				Hash:           blockToApply.Hash,
				PrevHash:       blockToApply.PrevHash,
				Timestamp:      time.Now(),
				ProposerNodeID: 0, // 暂时为0
				Data:           blockToApply.Data,
			}
			if err := tx.Block.Create(dbBlock); err != nil {
				return fmt.Errorf("failed to create block in db: %w", err)
			}

			// 创建交易记录
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
					return fmt.Errorf("failed to create transactions in db: %w", err)
				}
			}
			return nil
		})

		if err != nil {
			log.Printf("[Node %d] FATAL: Failed to apply block %d to database: %v. Halting application.", cn.id, appliedIndex, err)
			cn.lastApplied--
			return
		}

		log.Printf("[Node %d] Successfully applied and persisted block %d with %d transaction(s).\n", cn.id, appliedIndex, len(transactions))
	}
}

// --- Raft核心逻辑 (保持不变) ---

func (cn *ConsensusNode) ConnectToPeers() {
	cn.mu.Lock()
	defer cn.mu.Unlock()
	for peerID, addr := range cn.peer {
		if peerID != cn.id {
			conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
			if err != nil {
				log.Printf("[Node %d] Failed to dial peer %d at %s: %v. Will retry later.\n", cn.id, peerID, addr, err)
				continue
			}
			cn.peerClients[peerID] = raft.NewRaftServiceClient(conn)
			log.Printf("[Node %d] Successfully connected to peer %d at %s.\n", cn.id, peerID, addr)
		}
	}
}

func (cn *ConsensusNode) Run() {
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
	log.Printf("[Node %d Leader] Proposed new block at index %d, term %d. Log length: %d\n", cn.id, len(cn.log)-1, cn.currentTerm, len(cn.log))

	cn.matchIndex[cn.id] = len(cn.log) - 1
	cn.nextIndex[cn.id] = len(cn.log)

	return true, cn.currentTerm
}

func (cn *ConsensusNode) becomeFollower_nl(term int) {
	if cn.state == Leader {
		if cn.blockPacker != nil {
			cn.blockPacker.Stop()
		}
	}
	cn.state = Follower
	cn.currentTerm = term
	cn.votedFor = -1
	log.Printf("[Node %d] Becoming Follower in term %d.\n", cn.id, cn.currentTerm)
	cn.resetElectionTimer_nl()
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
	log.Printf("[Node %d] Election timer expired. Starting election for term %d.\n", cn.id, cn.currentTerm)

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
					log.Printf("[Node %d] Received vote from Node %d. Total votes: %d\n", cn.id, peerID, votes)
					if votes > len(cn.peer)/2 {
						log.Printf("[Node %d] Won election with %d votes. Becoming Leader!\n", cn.id, votes)
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

	log.Printf("[Node %d] Became Leader for term %d!\n", cn.id, cn.currentTerm)

	for peerID := range cn.peer {
		if peerID != cn.id {
			go cn.replicationLoopForPeer(peerID)
		}
	}

	cn.blockPacker = time.NewTicker(blockPackagingInterval)
	go cn.blockPackingLoop()
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

		log.Printf("[Node %d Leader] Packing %d transactions into a new block.\n", cn.id, len(txsToPack))
		data, err := json.Marshal(txsToPack)
		if err != nil {
			log.Printf("[Node %d Leader] ERROR: Failed to marshal transactions: %v\n", cn.id, err)
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
			log.Printf("[Node %d Leader] ERROR: prevLogIndex for peer %d is %d. This should not happen.", cn.id, peerID, prevLogIndex)
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
				log.Printf("[Node %d Leader] Node %d rejected AppendEntries. Retrying with nextIndex = %d\n", cn.id, peerID, cn.nextIndex[peerID])
			} else {
				log.Printf("[Node %d Leader] Node %d rejected AppendEntries, and nextIndex is already 1. Holding.", cn.id, peerID)
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
		log.Printf("[Node %d Leader] Advancing commitIndex from %d to %d\n", cn.id, cn.commitIndex, newCommitIndex)
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
		log.Printf("[Node %d] Voted for Node %d in term %d.\n", cn.id, args.CandidateId, cn.currentTerm)
		cn.resetElectionTimer_nl()
	} else {
		log.Printf("[Node %d] Rejected vote for Node %d. (votedFor=%d, isLogOk=%t)\n", cn.id, args.CandidateId, cn.votedFor, isLogOk)
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
		log.Printf("[Node %d Follower] Rejected AppendEntries from Leader %d: PrevLogIndex %d is out of bounds (my last index is %d)\n", cn.id, args.LeaderId, args.PrevLogIndex, lastLogIndex)
		return reply, nil
	}

	if cn.log[args.PrevLogIndex].Term != int64(args.PrevLogTerm) {
		log.Printf("[Node %d Follower] Rejected AppendEntries from Leader %d: Term mismatch at index %d. (MyTerm: %d, LeaderTerm: %d)\n", cn.id, args.LeaderId, args.PrevLogIndex, cn.log[args.PrevLogIndex].Term, args.PrevLogTerm)
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
		log.Printf("[Node %d Follower] Appended/updated logs from Leader %d. My log length is now %d.\n", cn.id, args.LeaderId, len(cn.log))
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
	log.Printf("[Node %d Leader] Received a new transaction, txPool size is now %d\n", cn.id, len(cn.txPool))

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