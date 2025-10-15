package router

import (
	"memology-backend/docs"
	"memology-backend/internal/handlers"
	"memology-backend/internal/middleware"
	"memology-backend/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(authService services.AuthService, userService services.UserService) *gin.Engine {
	r := gin.Default()

	// CORS middleware with credentials support
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
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

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// OpenAPI 3.0 endpoint
	r.GET("/openapi.json", func(c *gin.Context) {
		openapi3JSON, err := GetOpenAPI3Spec(docs.SwaggerJSON)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Type", "application/json")
		c.String(http.StatusOK, string(openapi3JSON))
	})

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
