package server

import (
	"net/http"
	"strconv"

	"github.com/haomiao000/raftchain/backend/server/controller/node"
	"github.com/haomiao000/raftchain/backend/server/controller/log"
	
	"github.com/gin-gonic/gin"
)

// GetNodeStatus 获取所有节点的状态
// @Router /api/v1/node/GetNodeStatus [GET]
func GetNodeStatus(c *gin.Context) {
	// 调用controller层获取数据
	nodes, err := node.GetNodeStatus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取节点状态失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, nodes)
}

// GetSystemLogs 获取系统日志
// @Router /api/v1/node/GetSystemLogs [GET]
func GetSystemLogs(c *gin.Context) {
	// 从查询参数中获取 'limit'，如果未提供则默认为50
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的limit参数"})
		return
	}

	// 调用controller层获取数据
	logs, err := log.GetSystemLogs(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取系统日志失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

// GetAdvancedLogs 处理高级日志筛选请求
// @Router /api/v1/log/GetAdvancedLogs [GET]
func GetAdvancedLogs(c *gin.Context) {
	// 将所有查询参数绑定到一个 map 或结构体中
	var filterParams log.LogFilterParams
	if err := c.ShouldBindQuery(&filterParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的查询参数: " + err.Error()})
		return
	}

	// 调用 controller 层进行筛选和查询
	result, err := log.GetAdvancedLogs(c.Request.Context(), filterParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询日志失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}