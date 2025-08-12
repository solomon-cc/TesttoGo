package middleware

import (
	"os"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func BasicAuth() gin.HandlerFunc {
	// 检查是否在生产环境中启用Swagger
	if os.Getenv("GIN_MODE") == "release" && os.Getenv("ENABLE_SWAGGER") != "true" {
		return gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(404, gin.H{"message": "Not Found"})
			c.Abort()
		})
	}

	// 从环境变量获取凭据，如果不存在则使用配置文件
	username := os.Getenv("SWAGGER_USERNAME")
	if username == "" {
		username = viper.GetString("swagger.username")
	}

	password := os.Getenv("SWAGGER_PASSWORD")
	if password == "" {
		password = viper.GetString("swagger.password")
	}

	// 确保凭据不为空
	if username == "" || password == "" {
		return gin.HandlerFunc(func(c *gin.Context) {
			c.JSON(503, gin.H{"message": "Service Unavailable"})
			c.Abort()
		})
	}

	return gin.BasicAuth(gin.Accounts{
		username: password,
	})
}
