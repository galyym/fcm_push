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

	"github.com/galyym/fcm_push/internal/config"
	"github.com/galyym/fcm_push/internal/handler"
	"github.com/galyym/fcm_push/internal/middleware"
	"github.com/galyym/fcm_push/internal/service"
	"github.com/galyym/fcm_push/pkg/fcm"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализация FCM клиента
	ctx := context.Background()
	fcmClient, err := fcm.NewClient(ctx, cfg.FCM.CredentialsPath)
	if err != nil {
		log.Fatalf("Failed to initialize FCM client: %v", err)
	}

	// Инициализация сервисов
	pushService := service.NewPushService(fcmClient)

	// Инициализация handlers
	pushHandler := handler.NewPushHandler(pushService)

	// Настройка Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	// Health check endpoint (без аутентификации)
	router.GET("/health", pushHandler.HealthCheck)

	// API routes с аутентификацией
	api := router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())
	{
		push := api.Group("/push")
		{
			push.POST("/send", pushHandler.SendPush)
			push.POST("/send-batch", pushHandler.SendBatchPush)
		}
	}

	// Создание HTTP сервера
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Запуск сервера в горутине
	go func() {
		log.Printf("Starting FCM Push Service on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
