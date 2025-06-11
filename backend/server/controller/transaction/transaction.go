package transaction

import (
	"context"
	"fmt"
	"math/rand"
	"encoding/json"
	"log"

	"github.com/haomiao000/raftchain/backend/client" // --- 新增import
	"github.com/haomiao000/raftchain/domain/node/rpc/raft" // --- 新增import
)

// SubmissionData 定义了接收交易提交的数据结构
// 使用 map[string]interface{} 来灵活处理不同类型的payload
type SubmissionData struct {
	Type            string                 `json:"type" binding:"required"`
	SenderPublicKey string                 `json:"senderPublicKey" binding:"required"`
	Payload         map[string]interface{} `json:"payload" binding:"required"`
}

// SubmissionResult 定义了交易提交成功后的返回结构
type SubmissionResult struct {
	Message         string `json:"message"`
	TransactionHash string `json:"transactionHash"`
}

// PoolEntry 定义了交易池中单个条目的结构
type PoolEntry struct {
	ID       string  `json:"id"`
	Sender   string  `json:"sender"`
	Receiver string  `json:"receiver,omitempty"` // omitempty表示如果字段为空，则在JSON中省略
	Amount   float64 `json:"amount,omitempty"`
	Type     string  `json:"type"`
	Summary  string  `json:"summary"` // 为不同类型交易提供统一的摘要
}


// SubmitTransaction 处理交易提交的业务逻辑 (当前为模拟)
func SubmitTransaction(ctx context.Context, txData SubmissionData) (*SubmissionResult, error) {
	// 1. 将前端传来的payload（map[string]interface{}）序列化为bytes
	payloadBytes, err := json.Marshal(txData.Payload)
	if err != nil {
		log.Printf("Failed to marshal payload: %v\n", err)
		return nil, fmt.Errorf("invalid payload format")
	}

	// 2. 创建gRPC格式的Transaction对象
	grpcTx := &raft.Transaction{
		Type:            txData.Type,
		SenderPublicKey: txData.SenderPublicKey,
		Payload:         payloadBytes,
	}

	// 3. 通过Raft客户端管理器将交易提交到集群
	raftClient := client.GetRaftClient()
	reply, err := raftClient.SubmitTransaction(grpcTx)
	if err != nil {
		return nil, err
	}

	// 4. 组装返回给前端的结果
	result := &SubmissionResult{
		Message:         "交易已成功提交到Raft集群",
		TransactionHash: reply.TxHash,
	}

	return result, nil
}

// GetTransactionPool 获取交易池数据的业务逻辑 (当前为模拟)
func GetTransactionPool(ctx context.Context) ([]PoolEntry, error) {
	// 在真实应用中，这里会从内存中的交易池数据结构中读取数据
	pool := make([]PoolEntry, 0)
	
	// 模拟生成一些不同类型的交易
	pool = append(pool, PoolEntry{
		ID:       fmt.Sprintf("tx_pool_%x", rand.Int31()),
		Sender:   "user_a_pubkey",
		Receiver: "user_b_pubkey",
		Amount:   100.5,
		Type:     "transfer",
		Summary:  "转账: 100.5",
	})
	pool = append(pool, PoolEntry{
		ID:      fmt.Sprintf("tx_pool_%x", rand.Int31()),
		Sender:  "user_c_pubkey",
		Type:    "data_storage",
		Summary: "存储数据: key=user_profile",
	})
	pool = append(pool, PoolEntry{
		ID:       fmt.Sprintf("tx_pool_%x", rand.Int31()),
		Sender:   "user_d_pubkey",
		Receiver: "user_e_pubkey",
		Amount:   50,
		Type:     "transfer",
		Summary:  "转账: 50",
	})

	return pool, nil
}