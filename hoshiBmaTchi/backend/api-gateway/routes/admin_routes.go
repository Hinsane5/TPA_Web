package routes

import (
	"github.com/Hinsane5/hoshiBmaTchi/backend/api-gateway/handlers"
	"github.com/Hinsane5/hoshiBmaTchi/backend/api-gateway/middleware"
	"github.com/gin-gonic/gin"
)

func SetupAdminRoutes(router *gin.Engine, adminHandler *handlers.AdminHandler, authHandler *handlers.AuthHandler) {
	admin := router.Group("/admin")
	
	admin.Use(authHandler.AuthMiddleware(), middleware.AdminMiddleware())
	
	{
		admin.GET("/users", adminHandler.GetAllUsers)
		admin.POST("/users/:id/ban", adminHandler.BanUser)

		admin.POST("/newsletter", adminHandler.SendNewsletter)

		admin.GET("/verifications", adminHandler.GetVerificationRequests)
		admin.POST("/verifications/:id/review", adminHandler.ReviewVerification)

		admin.GET("/reports", adminHandler.GetReports)        
		admin.POST("/reports/:id/review", adminHandler.ReviewReport) 
	}
}