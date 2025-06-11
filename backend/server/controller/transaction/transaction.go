package transaction

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/haomiao000/raftchain/backend/client"
	"github.com/haomiao000/raftchain/domain/node/rpc/raft"
	"github.com/haomiao000/raftchain/library/model"
	"github.com/haomiao000/raftchain/library/query"
	"github.com/haomiao000/raftchain/library/resource"
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

// SubmitTransaction 处理交易提交的业务逻辑
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

// GetTransactionPool 从数据库获取交易池数据，并根据CreatedAt进行去重
func GetTransactionPool(ctx context.Context) ([]PoolEntry, error) {
	// 1. 初始化数据库查询构造器
	q := query.Use(resource.GormServe)

	// 2. 从数据库查询所有交易，按ID降序排列（最新的在最前面）
	// 降序确保我们为每个时间戳保留的是最新插入的记录
	dbTxs, err := q.Tx.WithContext(ctx).Order(q.Tx.ID.Desc()).Find()
	if err != nil {
		log.Printf("Error fetching transactions from DB: %v", err)
		return nil, fmt.Errorf("无法从数据库获取交易数据")
	}

	// 3. 将数据库模型转换为前端需要的 PoolEntry 模型，并进行去重
	pool := make([]PoolEntry, 0)
	// 使用map来记录已经添加的CreatedAt时间戳（使用Unix秒级时间戳），确保唯一性
	seenTimestamps := make(map[int64]bool)
	seenBlock := make(map[string]bool)
	for _, tx := range dbTxs {
		ts := tx.CreatedAt.Unix()
		flag := true
		for i := range int64(10) {
			if _, seen := seenTimestamps[ts + i]; seen {
				// 如果这个时间戳的交易已经添加过，则跳过，实现去重
				flag = false
			}
		}
		if !flag {
			continue
		}
		if _, seen2 := seenBlock[tx.BlockHash]; seen2 {
			continue
		}
		entry, err := convertTxToPoolEntry(tx)
		if err != nil {
			// 记录转换过程中的错误，但继续处理下一条，保证接口的健壮性
			log.Printf("Failed to convert transaction %s: %v", tx.TxHash, err)
			continue
		}

		pool = append(pool, entry)
		// 标记这个时间戳已经处理过
		seenTimestamps[ts] = true
		seenBlock[tx.BlockHash] = true
	}

	return pool, nil
}

// convertTxToPoolEntry 是一个辅助函数，负责将数据库的Tx模型转换为API的PoolEntry模型
func convertTxToPoolEntry(tx *model.Tx) (PoolEntry, error) {
	entry := PoolEntry{
		ID:     tx.TxHash,
		Sender: tx.SenderPublicKey,
		Type:   tx.Type,
	}

	// 解析存储在 Payload 中的 JSON 数据
	var payload map[string]interface{}
	// 如果 payload 为空或者解析失败，我们依然可以返回部分信息
	if len(tx.Payload) > 0 {
		if err := json.Unmarshal(tx.Payload, &payload); err != nil {
			log.Printf("Could not unmarshal payload for tx %s: %v", tx.TxHash, err)
			entry.Summary = "载荷数据解析失败"
			return entry, nil
		}
	} else {
		payload = make(map[string]interface{}) // Ensure payload is not nil
	}

	// 根据交易类型，提取关键信息并生成摘要
	switch tx.Type {
	case "transfer":
		receiver, _ := payload["receiver"].(string)
		// JSON 数字默认被解析为 float64
		amount, _ := payload["amount"].(float64)

		entry.Receiver = receiver
		entry.Amount = amount
		entry.Summary = fmt.Sprintf("转账: 金额 %.2f", amount)

	case "data_storage":
		key, _ := payload["key"].(string)
		entry.Summary = fmt.Sprintf("数据存储: 键名='%s'", key)

	default:
		entry.Summary = fmt.Sprintf("未知类型 '%s' 交易", tx.Type)
	}

	return entry, nil
}
