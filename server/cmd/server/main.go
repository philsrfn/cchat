package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gotext/server/internal/api"
	"github.com/gotext/server/internal/db"
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
	
	// Get DB connection for passing to router setup
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.SSLMode,
	)
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Fatalf("Failed to create DB connection: %v", err)
	}
	
	// Create router using the Gin implementation from api package
	router := api.SetupRouter(dbConn)
	
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