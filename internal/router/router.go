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

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
