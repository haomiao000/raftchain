package control

import (
	"context"
	"fmt"
	"math/rand"
)

// NodeControlParams 定义了节点控制的参数结构
type NodeControlParams struct {
	Action       string `json:"action" binding:"required"`
	TargetNodeID string `json:"targetNodeId" binding:"required"`
	DelayMs      int    `json:"delay_ms,omitempty"`
}

// ControlResult 定义了操作结果的返回结构
type ControlResult struct {
	Message string `json:"message"`
}

type PerformanceStats struct {
	TPS           string `json:"tps"`
	BlockInterval string `json:"blockInterval"`
	AvgLatency    string `json:"avgLatency"`
	ElectionRate  string `json:"electionRate"`
}

// ExecuteNodeControl 执行节点控制指令 (当前为模拟)
func ExecuteNodeControl(ctx context.Context, params NodeControlParams) (*ControlResult, error) {
	// 在真实应用中，这里会通过RPC或其他方式向目标Raft节点发送真实的控制指令
	// 目前，我们仅在后端控制台打印收到的指令，并返回成功信息

	fmt.Printf("收到控制指令:\n")
	fmt.Printf("  - 操作: %s\n", params.Action)
	fmt.Printf("  - 目标节点ID: %s\n", params.TargetNodeID)
	if params.Action == "delay_network" {
		fmt.Printf("  - 延迟时间: %d ms\n", params.DelayMs)
	}

	message := fmt.Sprintf("'%s' 指令已成功模拟执行于节点 %s", params.Action, params.TargetNodeID)
	result := &ControlResult{
		Message: message,
	}

	return result, nil
}

// ... GetAdvancedLogs 和 generateAllMockLogs 函数保持不变 ...

// GetPerformanceStats 获取性能统计数据的业务逻辑 (当前为模拟)
func GetPerformanceStats(ctx context.Context) (*PerformanceStats, error) {
	// 在真实应用中，这些数据会从监控系统、指标数据库(如Prometheus)或应用内部的计时器计算得出
	// 目前，我们返回随机生成的模拟数据
	stats := &PerformanceStats{
		TPS:           fmt.Sprintf("%.1f", rand.Float64()*30+5),   // 5-35
		BlockInterval: fmt.Sprintf("%.1f", rand.Float64()*4+2),    // 2-6s
		AvgLatency:    fmt.Sprintf("%.0f", rand.Float64()*50+15),  // 15-65ms
		ElectionRate:  fmt.Sprintf("%.1f", rand.Float64()*2),      // 0-2 per hour
	}
	return stats, nil
}