package node

import (
	"context"
	"fmt"
	"time"

	"github.com/haomiao000/raftchain/library/query"
	"github.com/haomiao000/raftchain/library/resource"
)

// NodeStatus 定义了API响应的节点状态结构，与前端 other.html 对应
type NodeStatus struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	Term          int    `json:"term"`
	CommitIndex   int    `json:"commitIndex"`
	LastApplied   int    `json:"lastApplied"`
	LastHeartbeat string `json:"lastHeartbeat"`
	Status        string `json:"status"`
	Address       string `json:"address"`
}

// GetNodeStatus 从数据库获取所有节点的状态
func GetNodeStatus(ctx context.Context) ([]NodeStatus, error) {
	// 使用 bootstrap 中已初始化的全局数据库连接
	q := query.Use(resource.GormServe)

	// 从数据库查询所有节点信息，并按 node_id 排序
	dbNodes, err := q.Node.WithContext(ctx).Order(q.Node.NodeID).Find()
	if err != nil {
		return nil, fmt.Errorf("从数据库获取节点状态失败: %w", err)
	}

	// 准备API响应
	apiNodes := make([]NodeStatus, 0, len(dbNodes))
	for _, dbNode := range dbNodes {
		status := "活跃" // 前端显示中文 "活跃"

		// 如果最后心跳时间在15秒之前，则认为节点 "离线"
		if time.Since(dbNode.LastHeartbeat) > 3*time.Second {
			status = "离线"
		}

		// 将英文角色转换为中文以匹配前端
		var role string
		switch dbNode.Role {
		case "Leader":
			role = "领导者"
		case "Follower":
			role = "跟随者"
		case "Candidate":
			role = "候选者"
		default:
			role = "未知"
		}
		if status == "离线" {
			role = "未知"
		}
		apiNodes = append(apiNodes, NodeStatus{
			ID:            int(dbNode.NodeID),
			Name:          fmt.Sprintf("Node-%d", dbNode.NodeID),
			Address:       dbNode.Address,
			Role:          role,
			Term:          int(dbNode.Term),
			CommitIndex:   int(dbNode.CommitIndex),
			LastApplied:   int(dbNode.LastApplied),
			LastHeartbeat: dbNode.LastHeartbeat.Format("2006-01-02 15:04:05"),
			Status:        status,
		})
	}

	return apiNodes, nil
}