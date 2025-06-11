package client

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/haomiao000/raftchain/domain/node/rpc/raft"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"
	"os"
)

type RaftClientManager struct {
	mu          sync.RWMutex
	peers       map[int]string
	clients     map[int]raft.RaftServiceClient
	currentLeaderID int
}

var raftClient *RaftClientManager
var once sync.Once

// GetRaftClient 使用单例模式获取Raft客户端管理器
func GetRaftClient() *RaftClientManager {
	once.Do(func() {
		raftClient = &RaftClientManager{
			peers:       make(map[int]string),
			clients:     make(map[int]raft.RaftServiceClient),
			currentLeaderID: -1, // -1 表示未知
		}
	})
	return raftClient
}

// Init 从配置文件加载Raft节点信息并建立连接
func (rcm *RaftClientManager) Init(configFile string) error {
	rcm.mu.Lock()
	defer rcm.mu.Unlock()

	type Config struct {
		Nodes map[int]string `yaml:"nodes"`
	}

	file, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return err
	}

	rcm.peers = config.Nodes
	for id, addr := range rcm.peers {
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Failed to connect to Raft node %d at %s: %v\n", id, addr, err)
			continue
		}
		rcm.clients[id] = raft.NewRaftServiceClient(conn)
		log.Printf("Successfully established gRPC connection to Raft node %d\n", id)
	}
	return nil
}

// SubmitTransaction 尝试将交易提交给Leader节点
func (rcm *RaftClientManager) SubmitTransaction(tx *raft.Transaction) (*raft.TransactionReply, error) {
	rcm.mu.RLock()
	defer rcm.mu.RUnlock()

	// 简单的Leader发现和重试策略
	// 1. 如果已知Leader，直接发给Leader
	if rcm.currentLeaderID != -1 {
		client, ok := rcm.clients[rcm.currentLeaderID]
		if ok {
			reply, err := client.SubmitTransaction(context.Background(), tx)
			if err == nil && reply.Success {
				return reply, nil // 提交成功
			}
			// 如果失败（可能是Leader已变更），重置已知Leader，进入随机尝试
			rcm.currentLeaderID = -1
		}
	}
	
	// 2. 随机选择一个节点进行尝试
	for _, id := range rcm.getRandomPeerOrder() {
		client, ok := rcm.clients[id]
		if !ok {
			continue
		}
		log.Printf("Trying to submit transaction to potential leader: Node %d\n", id)
		reply, err := client.SubmitTransaction(context.Background(), tx)
		if err == nil && reply.Success {
			rcm.currentLeaderID = id // 提交成功，缓存LeaderID
			log.Printf("Found new leader: Node %d\n", id)
			return reply, nil
		}
	}
	
	return nil, errors.New("failed to submit transaction: could not find leader")
}

func (rcm *RaftClientManager) getRandomPeerOrder() []int {
	ids := make([]int, 0, len(rcm.peers))
	for id := range rcm.peers {
		ids = append(ids, id)
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })
	return ids
}

// GetBlocks 从任意一个节点查询区块列表
func (rcm *RaftClientManager) GetBlocks(page, limit int32) (*raft.GetBlocksReply, error) {
	rcm.mu.RLock()
	defer rcm.mu.RUnlock()

	// 对于只读请求，可以随机选择任何一个节点进行查询
	for _, id := range rcm.getRandomPeerOrder() {
		client, ok := rcm.clients[id]
		if !ok {
			continue
		}

		args := &raft.GetBlocksArgs{Page: page, Limit: limit}
		reply, err := client.GetBlocks(context.Background(), args)
		if err == nil {
			return reply, nil // 只要有一个节点成功返回即可
		}
	}

	return nil, errors.New("failed to get blocks from any node")
}