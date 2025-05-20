package router

import (
	"main/httpserver/server"
	"main/server/middleware"

	"net/http"

	"github.com/gin-gonic/gin"
)

func InitRouters(r *gin.Engine) {
	r.Use(middleware.CORSMiddleware())
	registerPingRoute(r)

	// 需要身份认证
	apiV1 := r.Group("/api/v1")
	apiV1.Use(middleware.VerifyToken())
	
	// 不需要身份认证
	apiV2 := r.Group("/api/v2")
	registerUserRoute(apiV2)
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
}
