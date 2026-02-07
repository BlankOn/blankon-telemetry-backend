package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	delivery "github.com/herpiko/blankon-telemetry-backend/internal/delivery/http"
	"github.com/herpiko/blankon-telemetry-backend/internal/repo"
	"github.com/herpiko/blankon-telemetry-backend/internal/usecase"
)

func main() {
	// Get config from environment
	// DATABASE_URL takes precedence; otherwise build from individual env vars
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		dbUser := getEnvOrDefault("POSTGRES_USER", "postgres")
		dbPassword := getEnvOrDefault("POSTGRES_PASSWORD", "postgres")
		dbHost := getEnvOrDefault("POSTGRES_HOST", "localhost")
		dbPort := getEnvOrDefault("POSTGRES_PORT", "5432")
		dbName := getEnvOrDefault("POSTGRES_DB", "telemetry")
		dbSSLMode := getEnvOrDefault("POSTGRES_SSLMODE", "disable")
		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode)
	}

	log.Printf("Database URL: %s", databaseURL)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Connect to database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	log.Println("Connected to TimescaleDB")

	// Initialize layers
	eventRepo := repo.NewEventRepository(pool)
	analyticsRepo := repo.NewAnalyticsRepository(pool)
	
	eventUC := usecase.NewEventUsecase(eventRepo)
	analyticsUC := usecase.NewAnalyticsUsecase(analyticsRepo)
	
	handler := delivery.NewHandler(eventUC, analyticsUC)
	router := delivery.NewRouter(handler)

	// Create server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
