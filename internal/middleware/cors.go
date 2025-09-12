package middleware

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

<<<<<<< HEAD
		// 允许 localhost 任意端口
		matchLocalhost, _ := regexp.MatchString(`^http://(localhost|127\.0\.0\.1)(:\d+)?$`, origin)
=======
		// 允许localhost的所有端口
		if origin == "http://localhost:3000" ||
			origin == "http://localhost:8080" ||
			origin == "http://localhost:3001" ||
			origin == "https://go.ylmz.com.cn" ||
			origin == "http://127.0.0.1:3000" ||
			origin == "http://127.0.0.1:8080" ||
			origin == "http://127.0.0.1:5173" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
>>>>>>> 9e0c830 (fix cors)

		if matchLocalhost || origin == "https://go.ylmz.com.cn" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
