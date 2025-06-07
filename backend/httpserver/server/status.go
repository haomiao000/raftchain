package server

import (
	"net/http"
	"strconv"

	"github.com/haomiao000/raftchain/backend/server/controller/node"
	"github.com/haomiao000/raftchain/backend/server/controller/log"
	
	"github.com/gin-gonic/gin"
)

// GetNodeStatus 获取所有节点的状态
// @Router /api/v1/status/nodes [GET]
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
// @Router /api/v1/status/logs [GET]
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