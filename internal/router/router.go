package router

import (
	"memology-backend/internal/handlers"
	"memology-backend/internal/middleware"
	"memology-backend/internal/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(authService services.AuthService, userService services.UserService) *gin.Engine {
	r := gin.Default()

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)

	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", middleware.JWTAuth(authService), authHandler.Logout)
			auth.POST("/logout-all", middleware.JWTAuth(authService), authHandler.LogoutAll)
		}

		users := api.Group("/users")
		users.Use(middleware.JWTAuth(authService))
		{
			users.GET("/profile", userHandler.GetProfile)
			users.PUT("/profile/update", userHandler.UpdateProfile)
			users.POST("/change-password", userHandler.ChangePassword)
			users.GET("/list", userHandler.GetUsers)
		}
	}

	return r
}
