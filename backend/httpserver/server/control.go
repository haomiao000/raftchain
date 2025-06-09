package server

import (
	"net/http"

	"github.com/haomiao000/raftchain/backend/server/controller/control"
	"github.com/gin-gonic/gin"
)

// ControlNode 接收并处理前端发送的节点控制指令
// @Router /api/v1/control/node [POST]
func ControlNode(c *gin.Context) {
	var params control.NodeControlParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的控制参数: " + err.Error()})
		return
	}

	result, err := control.ExecuteNodeControl(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func GetPerformanceStats(c *gin.Context) {
	stats, err := control.GetPerformanceStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取性能数据失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}