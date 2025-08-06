package main

import (
	"log"

	"testogo/internal/router"
	"testogo/pkg/config"
	"testogo/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	// 初始化数据库
	if err := database.Init(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 创建 Gin 引擎
	app := gin.Default()

	// 初始化路由
	router.InitRouter(app)

	// 启动服务器
	port := config.GetString("server.port")
	if err := app.Run(":" + port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
