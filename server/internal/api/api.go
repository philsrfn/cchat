package api

import (
	"database/sql"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/gotext/server/internal/auth"
	"github.com/gotext/server/internal/messages"
	"github.com/gotext/server/internal/middleware"
	"github.com/gotext/server/internal/spaces"
)

// SetupRouter initializes and returns a configured Gin router
func SetupRouter(db *sql.DB) *gin.Engine {
	router := gin.Default()

	// Configure CORS - More permissive for development
	config := cors.DefaultConfig()
	// Allow all origins for development
	config.AllowAllOrigins = true
	// Alternative: Allow specific origins
	// config.AllowOrigins = []string{"http://localhost:3000", "http://client:3000"}
	config.AllowCredentials = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	
	corsMiddleware := cors.New(config)
	router.Use(corsMiddleware)
	
	log.Println("CORS configuration applied with AllowAllOrigins:", config.AllowAllOrigins)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(db)

	// Services
	spaceService := spaces.NewSpaceService(db)
	messageService := messages.NewMessageService(db)
	authService := auth.NewAuthService(db)

	// Public routes
	router.POST("/api/auth/register", authService.Register)
	router.POST("/api/auth/login", authService.Login)
	router.POST("/api/auth/logout", authService.Logout)

	// Add a validate session endpoint
	router.GET("/api/auth/validate", authService.ValidateSession)

	// Protected API routes
	api := router.Group("/api")
	api.Use(authMiddleware.GinAuthMiddleware())
	{
		// User routes
		api.GET("/users/profile", authService.GetProfile)
		api.GET("/users", authService.GetUsers)

		// Register Space routes
		spaces.RegisterSpaceRoutes(api, spaceService)

		// Register Message routes
		messages.RegisterMessageRoutes(api, messageService)
	}

	// WebSocket endpoint (protected)
	router.GET("/ws", authMiddleware.GinAuthMiddleware(), messageService.WebSocketHandler)

	return router
} 