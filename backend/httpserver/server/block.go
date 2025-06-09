package server

import (
	"net/http"
	"strconv"

	"github.com/haomiao000/raftchain/backend/server/controller/block"
	"github.com/gin-gonic/gin"
)

// GetBlocks 获取分页的区块列表
// @Router /api/v1/blocks [GET]
func GetBlocks(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "15")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 15
	}

	result, err := block.GetBlocks(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取区块列表失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetBlockByHeight 根据高度获取单个区块的详细信息
// @Router /api/v1/blocks/:height [GET]
func GetBlockByHeight(c *gin.Context) {
	heightStr := c.Param("height")
	height, err := strconv.ParseInt(heightStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的区块高度"})
		return
	}

	result, err := block.GetBlockByHeight(c.Request.Context(), height)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "获取区块详情失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}