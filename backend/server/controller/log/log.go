package log

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

// LogEntry 定义了单条日志的结构
type LogEntry struct {
	NodeID    string `json:"node_id"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Level     string `json:"level"`
}

type LogFilterParams struct {
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=20"`
	StartDate string `form:"start_date"` // 格式 YYYY-MM-DD
	EndDate   string `form:"end_date"`   // 格式 YYYY-MM-DD
	Level     string `form:"level"`
	NodeID    string `form:"node_id"`
	Keyword   string `form:"keyword"`
}

// PaginatedLogResult 定义了包含分页信息的日志查询结果结构
type PaginatedLogResult struct {
	Logs       []LogEntry `json:"logs"`
	Pagination Pagination `json:"pagination"`
}

// Pagination 定义了分页元数据
type Pagination struct {
	CurrentPage int   `json:"currentPage"`
	TotalPages  int   `json:"totalPages"`
	TotalItems  int64 `json:"totalItems"`
	PerPage     int   `json:"perPage"`
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

func GetAdvancedLogs(ctx context.Context, params LogFilterParams) (*PaginatedLogResult, error) {
	// 在真实应用中，这些筛选条件会被转换成SQL查询
	// 当前，我们在模拟数据上执行这些筛选

	// 1. 生成全量模拟数据
	allLogs := generateAllMockLogs(10) // 生成一个较大的数据集用于筛选

	// 2. 执行筛选
	var filteredLogs []LogEntry
	for _, log := range allLogs {
		// 关键字筛选
		if params.Keyword != "" && !strings.Contains(strings.ToLower(log.Message), strings.ToLower(params.Keyword)) {
			continue
		}
		// 级别筛选
		if params.Level != "" && log.Level != params.Level {
			continue
		}
		// 节点筛选 (注意：我们的模拟数据node_id是 "Node-X" 格式)
		if params.NodeID != "" && log.NodeID != fmt.Sprintf("Node-%s", params.NodeID) {
			continue
		}
		// 日期筛选
		logTime, _ := time.Parse("2006-01-02 15:04:05", log.Timestamp)
		if params.StartDate != "" {
			startTime, err := time.Parse("2006-01-02", params.StartDate)
			if err == nil && logTime.Before(startTime) {
				continue
			}
		}
		if params.EndDate != "" {
			endTime, err := time.Parse("2006-01-02", params.EndDate)
			if err == nil {
				// 包含当天
				endTime = endTime.Add(24*time.Hour - 1*time.Nanosecond)
				if logTime.After(endTime) {
					continue
				}
			}
		}

		filteredLogs = append(filteredLogs, log)
	}

	// 3. 执行分页
	totalItems := int64(len(filteredLogs))
	totalPages := int(math.Ceil(float64(totalItems) / float64(params.Limit)))
	startIndex := (params.Page - 1) * params.Limit
	endIndex := startIndex + params.Limit
	if startIndex > len(filteredLogs) {
		startIndex = len(filteredLogs)
	}
	if endIndex > len(filteredLogs) {
		endIndex = len(filteredLogs)
	}
	paginatedLogs := filteredLogs[startIndex:endIndex]

	// 4. 组装返回结果
	result := &PaginatedLogResult{
		Logs: paginatedLogs,
		Pagination: Pagination{
			CurrentPage: params.Page,
			TotalPages:  totalPages,
			TotalItems:  totalItems,
			PerPage:     params.Limit,
		},
	}

	return result, nil
}

// generateAllMockLogs 是一个辅助函数，用于生成测试数据
func generateAllMockLogs(count int) []LogEntry {
	// ... (这个函数和 GetSystemLogs 里的模拟逻辑类似, 但用于生成一个固定的大集合)
	logs := make([]LogEntry, count)
	levels := []string{"INFO", "WARNING", "ERROR", "SUCCESS"}
	messages := []string{"心跳检测成功", "收到投票请求", "日志条目已复制", "Leader选举超时", "成功应用区块", "数据库连接失败", "配置已重新加载"}
	for i := 0; i < count; i++ {
		logs[i] = LogEntry{
			Timestamp: time.Now().Add(time.Duration(-i*15) * time.Minute).Format("2006-01-02 15:04:05"),
			Level:     levels[rand.Intn(len(levels))],
			Message:   fmt.Sprintf("Node%d: %s #%d", rand.Intn(4)+1, messages[rand.Intn(len(messages))], rand.Intn(1000)),
			NodeID:    fmt.Sprintf("Node-%d", rand.Intn(4)+1), // 添加 NodeID 以便筛选
		}
	}
	return logs
}
