package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/galyym/fcm_push/internal/config"
	"github.com/galyym/fcm_push/internal/database"
	"github.com/galyym/fcm_push/internal/handler"
	"github.com/galyym/fcm_push/internal/middleware"
	"github.com/galyym/fcm_push/internal/repository"
	"github.com/galyym/fcm_push/internal/service"
	"github.com/galyym/fcm_push/internal/worker"
	"github.com/galyym/fcm_push/pkg/fcm"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.NewDB(database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(cfg); err != nil {
		log.Printf("Warning: Migration failed: %v", err)
	}

	// Initialize FCM client
	ctx := context.Background()
	fcmClient, err := fcm.NewClient(ctx, cfg.FCM.CredentialsPath)
	if err != nil {
		log.Fatalf("Failed to initialize FCM client: %v", err)
	}

	// Initialize repositories
	queueRepo := repository.NewQueueRepository(db)

	// Initialize services
	pushService := service.NewPushService(fcmClient)
	queueService := service.NewQueueService(queueRepo)

	// Parse worker configuration
	pollInterval, err := time.ParseDuration(cfg.Worker.PollInterval)
	if err != nil {
		log.Fatalf("Invalid poll interval: %v", err)
	}

	retryIntervals, err := parseRetryIntervals(cfg.Worker.RetryIntervals)
	if err != nil {
		log.Fatalf("Invalid retry intervals: %v", err)
	}

	// Initialize and start queue worker
	queueWorker := worker.NewQueueWorker(queueRepo, fcmClient, worker.Config{
		WorkerCount:    cfg.Worker.WorkerCount,
		PollInterval:   pollInterval,
		RetryIntervals: retryIntervals,
		CleanupAfter:   time.Duration(cfg.Worker.CleanupAfterDays) * 24 * time.Hour,
	})
	queueWorker.Start()
	defer queueWorker.Stop()

	// Initialize handlers
	pushHandler := handler.NewPushHandler(pushService, queueService)
	queueHandler := handler.NewQueueHandler(queueService)

	// Setup router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	// Health check endpoint
	router.GET("/health", pushHandler.HealthCheck)

	// API routes
	api := router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())
	{
		// Push endpoints
		push := api.Group("/push")
		{
			push.POST("/send", pushHandler.SendPush)
			push.POST("/send-batch", pushHandler.SendBatchPush)
		}

		// Queue management endpoints
		queue := api.Group("/queue")
		{
			queue.GET("/status/:id", queueHandler.GetTaskStatus)
			queue.GET("/history", queueHandler.GetHistory)
			queue.GET("/stats", queueHandler.GetStats)
		}
	}

	// Start HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	go func() {
		log.Printf("Starting FCM Push Service on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// runMigrations runs database migrations
func runMigrations(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	m, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

// parseRetryIntervals parses retry intervals from comma-separated string
func parseRetryIntervals(intervalsStr string) ([]time.Duration, error) {
	parts := strings.Split(intervalsStr, ",")
	intervals := make([]time.Duration, 0, len(parts))

	for _, part := range parts {
		duration, err := time.ParseDuration(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("invalid duration %q: %w", part, err)
		}
		intervals = append(intervals, duration)
	}

	return intervals, nil
}
