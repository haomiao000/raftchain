package transaction

import (
	"context"
	"fmt"
	"math/rand"
	"time"
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
	// 在真实应用中，这里会进行：
	// 1. 验证签名 (使用 SenderPublicKey)
	// 2. 校验交易数据的合法性 (例如，检查账户余额)
	// 3. 将交易广播到网络中的交易池

	// 目前，我们仅打印接收到的数据并返回一个模拟的成功响应
	fmt.Printf("接收到新交易: \n")
	fmt.Printf("  - 类型: %s\n", txData.Type)
	fmt.Printf("  - 发送者公钥: %s\n", txData.SenderPublicKey)
	fmt.Printf("  - 交易内容 (Payload): %+v\n", txData.Payload)

	// 模拟生成一个交易哈希
	rand.Seed(time.Now().UnixNano())
	txHash := fmt.Sprintf("0x%x", rand.Int63())

	result := &SubmissionResult{
		Message:         "交易已成功提交到交易池",
		TransactionHash: txHash,
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