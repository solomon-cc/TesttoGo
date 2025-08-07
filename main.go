package main

import (
	"log"

	"testogo/docs"
	"testogo/internal/middleware"
	"testogo/internal/router"
	"testogo/pkg/config"
	"testogo/pkg/database"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	// 设置 Swagger 信息
	docs.SwaggerInfo.BasePath = "/api/v1"

	// 添加带认证的 Swagger 路由
	swaggerGroup := app.Group("/swagger")
	swaggerGroup.Use(middleware.BasicAuth())
	swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 初始化路由
	router.InitRouter(app)

	// 启动服务器
	port := config.GetString("server.port")
	if err := app.Run(":" + port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
