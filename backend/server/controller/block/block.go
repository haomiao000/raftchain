package block

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
	"github.com/haomiao000/raftchain/backend/client"
)

// Block 定义了区块列表中的区块结构
type Block struct {
	Height           int64  `json:"height"`
	Timestamp        string `json:"timestamp"`
	Hash             string `json:"hash"`
	PreviousHash     string `json:"previousHash"`
	TransactionCount int    `json:"transactionCount"`
}

// BlockDetail 定义了区块详情的结构
type BlockDetail struct {
	Block
	Term           int           `json:"term"`
	ProposerNodeID int           `json:"proposerNodeId"`
	Size           int           `json:"size"`
	Transactions   []Transaction `json:"transactions"`
}

// Transaction 定义了区块内交易的结构
type Transaction struct {
	ID       string  `json:"id"`
	Sender   string  `json:"sender"`
	Receiver string  `json:"receiver"`
	Amount   float64 `json:"amount"`
}

// PaginatedBlockResult 定义了分页结果的结构
type PaginatedBlockResult struct {
	Blocks     []Block    `json:"blocks"`
	Pagination Pagination `json:"pagination"`
}

// Pagination 定义了分页元数据
type Pagination struct {
	CurrentPage int   `json:"currentPage"`
	TotalPages  int   `json:"totalPages"`
	TotalItems  int64 `json:"totalItems"`
	PerPage     int   `json:"perPage"`
}

const totalMockBlocks = 500
const latestBlockHeight = 12850

// GetBlocks 获取分页的区块列表
func GetBlocks(ctx context.Context, page, limit int) (*PaginatedBlockResult, error) {
	// --- 修改：调用Raft客户端，不再使用模拟数据 ---
	raftClient := client.GetRaftClient()
	grpcReply, err := raftClient.GetBlocks(int32(page), int32(limit))
	if err != nil {
		return nil, err
	}
	
	// 将gRPC的响应格式转换为API的响应格式
	blocksForAPI := make([]Block, 0, len(grpcReply.Blocks))
	for _, b := range grpcReply.Blocks {
		blocksForAPI = append(blocksForAPI, Block{
			Height:           b.Term, // 注意：这里可能需要根据你的Block定义调整
			Timestamp:        "", // gRPC的Block没有Timestamp，可以后续添加
			Hash:             b.Hash,
			PreviousHash:     b.PrevHash,
			TransactionCount: 0, // gRPC的Block没有直接的tx count，可以后续添加
		})
	}
	
	result := &PaginatedBlockResult{
		Blocks: blocksForAPI,
		Pagination: Pagination{
			CurrentPage: int(grpcReply.Pagination.CurrentPage),
			TotalPages:  int(grpcReply.Pagination.TotalPages),
			TotalItems:  grpcReply.Pagination.TotalItems,
			PerPage:     int(grpcReply.Pagination.PerPage),
		},
	}

	return result, nil
}

// GetBlockByHeight 根据高度获取区块详情 (模拟)
func GetBlockByHeight(ctx context.Context, height int64) (*BlockDetail, error) {
	if height > latestBlockHeight || height <= 0 {
		return nil, errors.New("区块未找到")
	}

	txCount := rand.Intn(5)
	transactions := make([]Transaction, txCount)
	for i := 0; i < txCount; i++ {
		transactions[i] = Transaction{
			ID:       fmt.Sprintf("tx_%x", rand.Int63()),
			Sender:   fmt.Sprintf("sender_addr_%d", rand.Intn(100)),
			Receiver: fmt.Sprintf("receiver_addr_%d", rand.Intn(100)),
			Amount:   rand.Float64() * 100,
		}
	}

	detail := &BlockDetail{
		Block: Block{
			Height:           height,
			Timestamp:        time.Now().Add(time.Duration(-rand.Intn(10000)) * time.Second).Format("2006-01-02 15:04:05"),
			Hash:             fmt.Sprintf("0x%x_detail", height),
			PreviousHash:     fmt.Sprintf("0x%x_detail", height-1),
			TransactionCount: txCount,
		},
		Term:           int(height / 100),
		ProposerNodeID: rand.Intn(5) + 1,
		Size:           rand.Intn(2048) + 1024,
		Transactions:   transactions,
	}
	return detail, nil
}