package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gotext/server/internal/auth"
	"github.com/gotext/server/internal/db"
	"github.com/gotext/server/internal/middleware"
)

const (
	defaultPort = "8080"
)

func main() {
	// Setup logger
	logger := log.New(os.Stdout, "GoText: ", log.LstdFlags|log.Lshortfile)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Initialize database connection
	dbConfig := db.DefaultConfig()
	if err := db.Init(dbConfig); err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	
	// Create router and register routes
	router := http.NewServeMux()
	
	// Basic health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Authentication routes
	router.HandleFunc("/api/auth/register", auth.RegisterHandler)
	router.HandleFunc("/api/auth/login", auth.LoginHandler)
	router.HandleFunc("/api/auth/logout", auth.LogoutHandler)
	router.Handle("/api/auth/validate", middleware.RequireAuth(http.HandlerFunc(auth.ValidateAuthHandler)))
	
	// Protected routes example
	router.Handle("/api/user/profile", middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a protected endpoint - only accessible with a valid JWT
		userID, err := middleware.GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		// Just a simple response to demonstrate the protected route
		response := map[string]interface{}{
			"message": "Protected endpoint accessed successfully",
			"user_id": userID,
		}
		
		w.Header().Set("Content-Type", "application/json")
		auth.RespondWithJSON(w, http.StatusOK, response)
	})))

	// Create server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Printf("Starting server on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Println("Server gracefully stopped")
} 