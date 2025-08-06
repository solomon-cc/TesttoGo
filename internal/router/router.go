package router

import (
	"testogo/internal/controller"
	"testogo/internal/middleware"

	"github.com/gin-gonic/gin"
)

func InitRouter(r *gin.Engine) {
	// 公共路由组
	public := r.Group("/api/v1")
	{
		// 认证相关路由
		auth := public.Group("/auth")
		{
			auth.POST("/register", controller.Register)
			auth.POST("/login", controller.Login)
		}
	}

	// 需要认证的路由组
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())
	{
		// 题目相关路由
		questions := protected.Group("/questions")
		{
			questions.GET("", controller.ListQuestions)
			questions.GET("/:id", controller.GetQuestion)
			questions.POST("", middleware.RoleMiddleware("teacher", "admin"), controller.CreateQuestion)
			questions.PUT("/:id", middleware.RoleMiddleware("teacher", "admin"), controller.UpdateQuestion)
			questions.DELETE("/:id", middleware.RoleMiddleware("teacher", "admin"), controller.DeleteQuestion)
		}

		// 试卷相关路由
		papers := protected.Group("/papers")
		{
			papers.GET("", controller.ListPapers)
			papers.GET("/:id", controller.GetPaper)
			papers.POST("", middleware.RoleMiddleware("teacher", "admin"), controller.CreatePaper)
			papers.POST("/:id/submit", controller.SubmitPaper)
			papers.GET("/:id/result", controller.GetPaperResult)
		}

		// 用户管理路由（仅管理员）
		users := protected.Group("/users")
		users.Use(middleware.RoleMiddleware("admin"))
		{
			users.GET("", controller.ListUsers)
			users.PUT("/:id/role", controller.UpdateUserRole)
			users.DELETE("/:id", controller.DeleteUser)
		}
	}
}
