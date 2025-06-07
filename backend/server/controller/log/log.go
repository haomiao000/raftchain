package log

import (
	"context"
	"math/rand"
	"time"
	"fmt"
)

// LogEntry 定义了单条日志的结构
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Level     string `json:"level"`
}

// GetSystemLogs 获取系统日志的业务逻辑 (当前为模拟)
func GetSystemLogs(ctx context.Context, limit int) ([]LogEntry, error) {
	// 在真实应用中，这里会从日志文件、数据库或日志服务中查询
	// 目前，我们返回模拟数据
	levels := []string{"INFO", "WARNING", "ERROR", "SUCCESS"}
	messages := []string{"心跳检测成功", "收到投票请求 from Node2", "日志条目已复制到大多数节点", "Leader选举超时，开始新一轮选举", "成功应用区块 #1256"}
	
	logs := make([]LogEntry, 0, limit)
	for i := 0; i < limit; i++ {
		log := LogEntry{
			Timestamp: time.Now().Add(time.Duration(-i*5) * time.Second).Format("2006-01-02 15:04:05"),
			Level:     levels[rand.Intn(len(levels))],
			Message:   fmt.Sprintf("Node%d: %s", rand.Intn(4)+1, messages[rand.Intn(len(messages))]),
		}
		logs = append(logs, log)
	}

	return logs, nil
}