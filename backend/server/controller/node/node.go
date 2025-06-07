package node 

import (
	"context"
	"math/rand"
	"time"
)

// NodeStatus 定义了单个节点的状态结构
type NodeStatus struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	Term          int    `json:"term"`
	LogIndex      int    `json:"logIndex"`
	LastHeartbeat string `json:"lastHeartbeat"`
	Status        string `json:"status"`
}


// GetNodeStatus 获取节点状态的业务逻辑 (当前为模拟)
func GetNodeStatus(ctx context.Context) ([]NodeStatus, error) {
	// 在真实应用中，这里会连接到Raft集群并查询每个节点的状态
	// 目前，我们返回模拟数据
	roles := []string{"领导者", "跟随者", "候选者"}
	statuses := []string{"活跃", "离线"}

	nodes := []NodeStatus{
		{ID: 1, Name: "Node1", Role: roles[0], Term: 12, LogIndex: 1256, LastHeartbeat: time.Now().Format("15:04:05"), Status: statuses[0]},
		{ID: 2, Name: "Node2", Role: roles[1], Term: 12, LogIndex: 1255, LastHeartbeat: time.Now().Add(-2 * time.Second).Format("15:04:05"), Status: statuses[0]},
		{ID: 3, Name: "Node3", Role: roles[1], Term: 12, LogIndex: 1255, LastHeartbeat: time.Now().Add(-30 * time.Second).Format("15:04:05"), Status: statuses[1]},
		{ID: 4, Name: "Node4", Role: roles[1], Term: 11, LogIndex: 1198, LastHeartbeat: time.Now().Add(-5 * time.Second).Format("15:04:05"), Status: statuses[0]},
	}

	// 为数据添加一些随机性
	for i := range nodes {
		if i > 0 {
			nodes[i].Role = roles[rand.Intn(2)+1] // 只有Node1是领导者
		}
		nodes[i].LogIndex += rand.Intn(20)
	}

	return nodes, nil
}

