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

func SetupRouter(authService services.AuthService, userService services.UserService, memeService services.MemeService) *gin.Engine {
	r := gin.Default()

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
	memeHandler := handlers.NewMemeHandler(memeService)

	apiRoot := r.Group("/api")
	{
		apiRoot.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		apiRoot.GET("/openapi.json", func(c *gin.Context) {
			openapi3JSON, err := GetOpenAPI3Spec(docs.SwaggerJSON)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.Header("Content-Type", "application/json")
			c.String(http.StatusOK, string(openapi3JSON))
		})
	}

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
			users.DELETE("/account", userHandler.DeleteAccount)
		}

		memes := api.Group("/memes")
		{
			memes.GET("", memeHandler.GetAllMemes)
			memes.GET("/public", memeHandler.GetPublicMemes)
			memes.GET("/styles", memeHandler.GetAvailableStyles)
			memes.GET("/:id", memeHandler.GetMeme)
			memes.GET("/:id/status", memeHandler.CheckMemeStatus)

			memes.Use(middleware.JWTAuth(authService))
			memes.POST("/generate", memeHandler.GenerateMeme)
			memes.GET("/my", memeHandler.GetMyMemes)
			memes.DELETE("/:id", memeHandler.DeleteMeme)
		}
	}

	return r
}
