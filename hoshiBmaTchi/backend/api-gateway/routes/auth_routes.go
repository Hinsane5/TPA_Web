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
	}
}
