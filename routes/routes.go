package routes

import (
	"imp101/controllers"
	"imp101/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	// Health route
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Keep backward-compatible root routes and add versioned routes.
	registerRoutes(router.Group("/"))
	registerRoutes(router.Group("/api/v1"))
}

func registerRoutes(group *gin.RouterGroup) {
	// Public routes
	group.POST("/signup", controllers.Signup)
	group.POST("/login", controllers.Login)
	group.GET("/jobs", controllers.ListJobs)
	group.GET("/jobs/:id", controllers.GetJob)

	// Protected routes
	protected := group.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/users", controllers.GetUser)
		protected.GET("/users/:id", controllers.GetUser)
		protected.DELETE("/users/me", controllers.DeleteMyData)
		protected.GET("/applications", controllers.ListMyApplications)
		protected.GET("/applications/:id", controllers.GetMyApplication)
		protected.POST("/applications", controllers.ApplyToJob)
	}

	// Recruiter/Admin routes
	recruiter := group.Group("/")
	recruiter.Use(middleware.AuthMiddleware())
	recruiter.Use(middleware.RecruiterOrAdminMiddleware())
	{
		recruiter.POST("/jobs", controllers.CreateJob)
		recruiter.PUT("/jobs/:id", controllers.UpdateJob)
		recruiter.POST("/jobs/:id/publish", controllers.PublishJob)
		recruiter.POST("/jobs/:id/close", controllers.CloseJob)
		recruiter.POST("/jobs/:id/archive", controllers.ArchiveJob)
		recruiter.GET("/job-rankings/:jobId", controllers.RankCandidatesByJob)
		recruiter.GET("/application-explainability/:id", controllers.GetApplicationExplainability)
	}

	// Admin routes
	admin := group.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/users", controllers.GetAllUsers)
		admin.PATCH("/users/:id/role", controllers.UpdateUserRole)
		admin.PATCH("/users/:id/status", controllers.UpdateUserStatus)
	}
}
