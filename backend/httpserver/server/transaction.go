package server

import (
	"net/http"

	"github.com/haomiao000/raftchain/backend/server/controller/transaction"
	"github.com/gin-gonic/gin"
)

// SubmitTransaction 接收前端提交的新交易
// @Router /api/v1/transactions [POST]
func SubmitTransaction(c *gin.Context) {
	var txData transaction.SubmissionData
	if err := c.ShouldBindJSON(&txData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的交易数据: " + err.Error()})
		return
	}

	// 从认证中间件中获取用户ID (可选，但推荐)
	// userID, _ := c.Get("userID")
	// txData.SubmitterID = userID.(int64)

	// 调用controller层处理交易
	result, err := transaction.SubmitTransaction(c.Request.Context(), txData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "处理交易失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// @Router /api/v1/transactions/pool [GET]
func GetTransactionPool(c *gin.Context) {
	result, err := transaction.GetTransactionPool(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取交易池数据失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}