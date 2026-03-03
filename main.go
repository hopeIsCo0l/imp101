package main

import (
	"imp101/database"
	"imp101/middleware"
	"imp101/routes"
	"imp101/utils"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize JWT secret
	utils.InitializeJWTSecret()

	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Create Gin router
	r := gin.Default()
	r.Use(middleware.RateLimit(100))
	r.Use(middleware.AuditMiddleware())

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		allowedOrigin := os.Getenv("CORS_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "http://localhost:3000"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup routes
	routes.SetupRoutes(r)

	// Print all registered routes for debugging
	log.Println("=== Registered Routes ===")
	for _, route := range r.Routes() {
		log.Printf("%-6s %s", route.Method, route.Path)
	}
	log.Println("========================")

	// Start the server
	port := "8080"
	log.Println("Server is running on port", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
