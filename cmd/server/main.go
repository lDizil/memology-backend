package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "memology-backend/docs"
	"memology-backend/internal/config"
	"memology-backend/internal/database"
	"memology-backend/internal/repository"
	"memology-backend/internal/router"
	"memology-backend/internal/services"
	"memology-backend/pkg/auth"
)

// @title Memology API
// @version 1.0
// @description API for meme generation platform
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token
func main() {
	cfg := config.Load()

	db := database.Connect(&cfg.Database)
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	jwtManager := auth.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	memeRepo := repository.NewMemeRepository(db)

	minioService, err := services.NewMinIOService(&cfg.MinIO)
	if err != nil {
		log.Fatal("Failed to initialize MinIO:", err)
	}

	aiService := services.NewAIService(&cfg.AI)

	authService := services.NewAuthService(userRepo, sessionRepo, jwtManager)
	userService := services.NewUserService(userRepo)

	taskProcessor := services.NewTaskProcessor(cfg, memeRepo, aiService, minioService)
	taskProcessor.Start()
	defer taskProcessor.Stop()

	memeService := services.NewMemeServiceWithProcessor(memeRepo, minioService, aiService, taskProcessor)

	r := router.SetupRouter(authService, userService, memeService)

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
