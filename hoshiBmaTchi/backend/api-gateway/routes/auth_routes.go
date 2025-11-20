package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/Hinsane5/hoshiBmaTchi/backend/api-gateway/handlers" 
	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"    
)

func SetupAuthRoutes(router *gin.Engine, userClient pb.UserServiceClient){

	authHandler := &handlers.AuthHandler{
		UserClient: userClient,
	}

	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/send-otp", authHandler.SendOtp)
		authGroup.POST("/google/login", authHandler.LoginWithGoogle)
		authGroup.POST("/login", authHandler.LoginUser)
		authGroup.POST("/verify-2fa", authHandler.VerifyLogin2FA)
		authGroup.POST("/request-password-reset", authHandler.RequestPasswordReset)
		authGroup.POST("/reset-password", authHandler.PerformPasswordReset)
	}

	usersGroup := router.Group("/api/v1/users")
	usersGroup.Use(authHandler.AuthMiddleware())
	{
		usersGroup.GET("/:id", authHandler.GetUserProfile)
		usersGroup.POST("/:id/follow", authHandler.FollowUser) 
    	usersGroup.DELETE("/:id/follow", authHandler.UnfollowUser)
		usersGroup.GET("/search", authHandler.SearchUsers)
	}
}
