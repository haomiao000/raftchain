package router

import (
	"github.com/haomiao000/raftchain/backend/httpserver/server"
	"github.com/haomiao000/raftchain/backend/server/middleware"

	"net/http"

	"github.com/gin-gonic/gin"
)

func InitRouters(r *gin.Engine) {
	r.Use(middleware.CORSMiddleware())
	registerPingRoute(r)

	// 需要身份认证
	apiV1 := r.Group("/api/v1")
	apiV1.Use(middleware.VerifyToken())
	{
		registerNodeRoute(apiV1)
		registeLogRoute(apiV1)
		registeBlockRoute(apiV1)
		registeTransactionRoute(apiV1)
		registeControlRoute(apiV1)
	}

	// 不需要身份认证
	apiV2 := r.Group("/api/v2")
	{
		registerUserRoute(apiV2)
	}
}

func registerPingRoute(r *gin.Engine) {
	r.POST("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
}

func registerUserRoute(rg *gin.RouterGroup) {
	r := rg.Group("/user")
	r.POST("/register", server.Register)
	r.POST("/login", server.Login)
	// 使用新的密码修改路由
	r.POST("/change-password", server.ChangePassword)
}

func registerNodeRoute(rg *gin.RouterGroup) {
	r := rg.Group("/node")
	r.GET("/GetNodeStatus", server.GetNodeStatus)
}

func registeLogRoute(rg *gin.RouterGroup) {
	r := rg.Group("/log")
	r.GET("/GetSystemLogs", server.GetSystemLogs)
	r.GET("/GetAdvancedLogs",server.GetAdvancedLogs)
}

func registeBlockRoute(rg *gin.RouterGroup) {
	r := rg.Group("/block")
	r.GET("", server.GetBlocks) // 获取区块列表
	r.GET("/:height", server.GetBlockByHeight) // 获取区块详情
}

func registeTransactionRoute(rg *gin.RouterGroup) {
	r := rg.Group("/transactions")
	r.POST("/SubmitTransaction", server.SubmitTransaction)
	r.GET("/GetTransactionPool", server.GetTransactionPool)
}

func registeControlRoute(rg *gin.RouterGroup) {
	r := rg.Group("/control")
	r.GET("/GetPerformanceStats", server.GetPerformanceStats)
	r.POST("/ControlNode", server.ControlNode)
}