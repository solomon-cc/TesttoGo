package docs

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SwaggerHandler returns gin handler for swagger
func SwaggerHandler() gin.HandlerFunc {
	return ginSwagger.WrapHandler(swaggerFiles.Handler)
}