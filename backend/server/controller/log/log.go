package log

import (
	"context"
	"fmt"
	"math"
	"github.com/haomiao000/raftchain/library/query"
	"github.com/haomiao000/raftchain/library/resource"
	"time"
)

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
    q := query.Use(resource.GormServe)

    // 从数据库查询最新的日志
    dbLogs, err := q.SystemLog.WithContext(ctx).Order(q.SystemLog.Timestamp.Desc()).Limit(limit).Find()
    if err != nil {
        return nil, fmt.Errorf("从数据库获取日志失败: %w", err)
    }

    // 格式化为 API 响应
    apiLogs := make([]LogEntry, len(dbLogs))
    for i, log := range dbLogs {
        apiLogs[i] = LogEntry{
            NodeID:    fmt.Sprintf("Node-%d", log.NodeID),
            Timestamp: log.Timestamp.Format("2006-01-02 15:04:05"),
            Message:   log.Message,
            Level:     log.Level,
        }
    }

    return apiLogs, nil
}
type LogRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Query    string `form:"query"`
	Level    string `form:"level"`
	NodeID   string `form:"node_id"`
}

// LogEntry 定义了返回给前端的单条日志格式 (保持不变)
type LogEntry struct {
	NodeID    string `json:"node_id"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Level     string `json:"level"`
}
// GetAdvancedLogs 从数据库中获取并筛选日志 (核心逻辑)
func GetAdvancedLogs(ctx context.Context, params LogFilterParams) (*PaginatedLogResult, error) {
	// 1. 初始化 GORM 查询
	q := query.Use(resource.GormServe)
	logQuery := q.SystemLog.WithContext(ctx)

	// 2. 根据前端参数，动态构建 WHERE 查询条件
	if params.Keyword != "" {
		logQuery = logQuery.Where(q.SystemLog.Message.Like("%" + params.Keyword + "%"))
	}
	if params.Level != "" {
		logQuery = logQuery.Where(q.SystemLog.Level.Eq(params.Level))
	}
	if params.NodeID != "" {
		// 从 "Node-1" 或 "1" 这种格式中解析出数字 ID
		var nodeNum int
		// 优先尝试 "Node-1" 格式
		if _, err := fmt.Sscanf(params.NodeID, "Node-%d", &nodeNum); err != nil {
			// 如果失败，尝试直接转换数字 "1"
			if _, err := fmt.Sscanf(params.NodeID, "%d", &nodeNum); err == nil {
				logQuery = logQuery.Where(q.SystemLog.NodeID.Eq(int32(nodeNum)))
			}
		} else {
			logQuery = logQuery.Where(q.SystemLog.NodeID.Eq(int32(nodeNum)))
		}
	}
    if params.StartDate != "" {
        startTime, err := time.Parse("2006-01-02", params.StartDate)
        if err == nil {
            // Gte = Greater than or equal to (>=)
            logQuery = logQuery.Where(q.SystemLog.Timestamp.Gte(startTime))
        }
    }
    if params.EndDate != "" {
        endTime, err := time.Parse("2006-01-02", params.EndDate)
        if err == nil {
            // 为了包含当天，将时间设为当天的 23:59:59
            endTime = endTime.Add(24*time.Hour - 1*time.Nanosecond)
            // Lte = Less than or equal to (<=)
            logQuery = logQuery.Where(q.SystemLog.Timestamp.Lte(endTime))
        }
    }


	// 3. 获取符合筛选条件的总日志数 (用于分页)
	totalItems, err := logQuery.Count()
	if err != nil {
		return nil, fmt.Errorf("获取日志总数失败: %w", err)
	}

	// 4. 根据分页参数，获取当前页的日志数据
    if params.Page <= 0 {
        params.Page = 1
    }
    if params.Limit <= 0 {
        params.Limit = 20
    }
	offset := (params.Page - 1) * params.Limit
	// 按时间倒序排列，最新的日志在最前面
	dbLogs, err := logQuery.Order(q.SystemLog.Timestamp.Desc()).Offset(offset).Limit(params.Limit).Find()
	if err != nil {
		return nil, fmt.Errorf("查询日志失败: %w", err)
	}

	// 5. 将数据库模型转换为前端需要的 JSON 格式
	apiLogs := make([]LogEntry, len(dbLogs))
	for i, log := range dbLogs {
		apiLogs[i] = LogEntry{
			NodeID:    fmt.Sprintf("Node-%d", log.NodeID),
			Timestamp: log.Timestamp.Format("2006-01-02 15:04:05"),
			Message:   log.Message,
			Level:     log.Level,
		}
	}
    
    // 6. 组装最终返回结果
    result := &PaginatedLogResult{
        Logs:       apiLogs,
        Pagination: Pagination{
            CurrentPage: params.Page,
            TotalPages:  int(math.Ceil(float64(totalItems) / float64(params.Limit))),
            TotalItems:  totalItems,
            PerPage:     params.Limit,
        },
    }

	return result, nil
}