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
			questions.GET("/:id/statistics", controller.GetQuestionStatistics)
			questions.POST("", middleware.RoleMiddleware("teacher", "admin"), controller.CreateQuestion)
			questions.POST("/:id/answer", controller.AnswerQuestion)
			questions.PUT("/:id", middleware.RoleMiddleware("teacher", "admin"), controller.UpdateQuestion)
			questions.DELETE("/:id", middleware.RoleMiddleware("teacher", "admin"), controller.DeleteQuestion)
		}

		// 试卷相关路由
		papers := protected.Group("/papers")
		{
			papers.GET("", controller.ListPapers)
			papers.GET("/:id", controller.GetPaper)
			papers.POST("", middleware.RoleMiddleware("teacher", "admin"), controller.CreatePaper)
			papers.PUT("/:id", middleware.RoleMiddleware("teacher", "admin"), controller.UpdatePaper)
			papers.POST("/:id/submit", controller.SubmitPaper)
			papers.GET("/:id/result", controller.GetPaperResult)
		}

		// 用户相关路由
		users := protected.Group("/users")
		{
			// 答题历史（所有用户可访问自己的历史）
			users.GET("/answers/history", controller.GetUserAnswerHistory)
			// 用户答题表现统计
			users.GET("/performance", controller.GetUserPerformance)
			// 用户设置
			users.GET("/settings", controller.GetUserSettings)
			users.PUT("/settings", controller.UpdateUserSettings)
			users.POST("/settings/reset", controller.ResetUserSettings)
		}

		// 用户管理路由（仅管理员）
		admin := protected.Group("/users")
		admin.Use(middleware.RoleMiddleware("admin"))
		{
			admin.GET("", controller.ListUsers)
			admin.POST("", controller.CreateUser)
			admin.PUT("/:id", controller.UpdateUser)
			admin.PUT("/:id/role", controller.UpdateUserRole)
			admin.DELETE("/:id", controller.DeleteUser)
		}

		// 作业相关路由
		homework := protected.Group("/homework")
		{
			// 学生查看作业
			homework.GET("/student", controller.ListHomework)
			homework.GET("/:id", controller.GetHomework)
			homework.POST("/submit", controller.SubmitHomework)
			
			// 教师/管理员作业管理
			homework.GET("/teacher", middleware.RoleMiddleware("teacher", "admin"), controller.ListHomework)
			homework.POST("", middleware.RoleMiddleware("teacher", "admin"), controller.CreateHomework)
			homework.PUT("/:id", middleware.RoleMiddleware("teacher", "admin"), controller.UpdateHomework)
			homework.DELETE("/:id", middleware.RoleMiddleware("teacher", "admin"), controller.DeleteHomework)
			homework.POST("/:id/copy", middleware.RoleMiddleware("teacher", "admin"), controller.CopyHomework)
			homework.GET("/:id/submissions", middleware.RoleMiddleware("teacher", "admin"), controller.GetHomeworkSubmissions)
			homework.PUT("/:id/adjust", middleware.RoleMiddleware("teacher", "admin"), controller.AdjustHomework)
			homework.GET("/history", middleware.RoleMiddleware("teacher", "admin"), controller.GetHomeworkHistory)
		}

		// 强化学习相关路由
		reinforcement := protected.Group("/reinforcements")
		{
			// 强化物管理
			reinforcement.GET("", controller.ListReinforcementItems)
			reinforcement.GET("/:id", controller.GetReinforcementItem)
			reinforcement.POST("", middleware.RoleMiddleware("teacher", "admin"), controller.CreateReinforcementItem)
			reinforcement.PUT("/:id", middleware.RoleMiddleware("teacher", "admin"), controller.UpdateReinforcementItem)
			reinforcement.DELETE("/:id", middleware.RoleMiddleware("admin"), controller.DeleteReinforcementItem)
			
			// 强化物视频上传
			reinforcement.POST("/videos", middleware.RoleMiddleware("teacher", "admin"), controller.UploadRewardVideo)
			reinforcement.DELETE("/videos/:id", middleware.RoleMiddleware("teacher", "admin"), controller.DeleteRewardVideo)
		}

		// 强化设置相关路由
		reinforcementSettings := protected.Group("/reinforcement-settings")
		{
			reinforcementSettings.GET("", controller.ListReinforcementSettings)
			reinforcementSettings.GET("/:id", controller.GetReinforcementSetting)
			reinforcementSettings.POST("", middleware.RoleMiddleware("teacher", "admin"), controller.CreateReinforcementSetting)
			reinforcementSettings.PUT("/:id", middleware.RoleMiddleware("teacher", "admin"), controller.UpdateReinforcementSetting)
			reinforcementSettings.DELETE("/:id", middleware.RoleMiddleware("teacher", "admin"), controller.DeleteReinforcementSetting)
			reinforcementSettings.POST("/:id/copy", middleware.RoleMiddleware("teacher", "admin"), controller.CopyReinforcementSetting)
		}

		// 强化日志相关路由
		reinforcementLogs := protected.Group("/reinforcement-logs")
		{
			reinforcementLogs.POST("", controller.RecordReinforcementTrigger)
			reinforcementLogs.GET("/stats", controller.GetReinforcementStats)
		}

		// 文件上传相关路由
		upload := protected.Group("/upload")
		{
			upload.POST("/images", middleware.RoleMiddleware("teacher", "admin"), controller.UploadImages)
		}

		// 题目导入相关路由
		importGroup := protected.Group("/import")
		{
			importGroup.POST("/questions", middleware.RoleMiddleware("teacher", "admin"), controller.ImportQuestions)
			importGroup.POST("/questions/confirm", middleware.RoleMiddleware("teacher", "admin"), controller.ConfirmImportQuestions)
		}

		// 静态文件服务
		media := r.Group("/api/v1/media")
		{
			media.GET("/:filename", controller.ServeMedia)
		}

		// 年级、科目、主题相关路由
		content := protected.Group("/content")
		{
			// 年级管理
			grades := content.Group("/grades")
			{
				grades.GET("", controller.ListGrades)
				grades.GET("/:id", controller.GetGrade)
				grades.POST("", middleware.RoleMiddleware("admin"), controller.CreateGrade)
			}

			// 科目管理
			subjects := content.Group("/subjects")
			{
				subjects.GET("", controller.ListSubjects)
				subjects.GET("/:id", controller.GetSubject)
				subjects.GET("/by-code/:code", controller.GetSubjectByCode)
				subjects.POST("", middleware.RoleMiddleware("admin"), controller.CreateSubject)
				subjects.GET("/:id/topics", controller.ListTopics)
			}

			// 主题管理
			topics := content.Group("/topics")
			{
				topics.GET("/:id", controller.GetTopic)
				topics.GET("/by-code/:subject_code/:topic_code", controller.GetTopicByCode)
				topics.POST("", middleware.RoleMiddleware("admin"), controller.CreateTopic)
			}
		}
	}
}
