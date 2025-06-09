package block

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
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

// GetBlocks 获取分页的区块列表 (模拟)
func GetBlocks(ctx context.Context, page, limit int) (*PaginatedBlockResult, error) {
	totalPages := int(math.Ceil(float64(totalMockBlocks) / float64(limit)))
	
	blocks := make([]Block, 0, limit)
	startHeight := latestBlockHeight - int64((page-1)*limit)

	for i := 0; i < limit; i++ {
		currentHeight := startHeight - int64(i)
		if currentHeight <= 0 {
			break
		}
		blocks = append(blocks, Block{
			Height:           currentHeight,
			Timestamp:        time.Now().Add(time.Duration(-i*10) * time.Second).Format("2006-01-02 15:04:05"),
			Hash:             fmt.Sprintf("0x%x", rand.Int63()),
			PreviousHash:     fmt.Sprintf("0x%x", rand.Int63()),
			TransactionCount: rand.Intn(10),
		})
	}

	result := &PaginatedBlockResult{
		Blocks: blocks,
		Pagination: Pagination{
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalItems:  totalMockBlocks,
			PerPage:     limit,
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