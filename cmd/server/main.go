package main

import (
	"log"

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
// @host localhost:8080
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

	authService := services.NewAuthService(userRepo, sessionRepo, jwtManager)
	userService := services.NewUserService(userRepo)

	r := router.SetupRouter(authService, userService)

	log.Printf("Server starting on %s:%s", cfg.Server.Host, cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
