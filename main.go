package main

import (
	"log"
	"os"
	"main/bootstrap"
	"main/httpserver/router"

	"github.com/gin-gonic/gin"
)

var (
	AppServerPort = ":8080"
)

func main() {
	enableModelInit := false
	for _, arg := range os.Args[1:] {
		if arg == "--model_init" {
			enableModelInit = true
			break
		}
	}
	bootstrap.InitBootStrap(enableModelInit)
	r := gin.Default()
	router.InitRouters(r)
	
	err := r.Run(AppServerPort)
	if err != nil {
		log.Fatalf("启动 Gin 服务器失败: %v", err)
	}
}