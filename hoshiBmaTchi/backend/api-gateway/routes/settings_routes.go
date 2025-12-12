package routes

import (
	"github.com/Hinsane5/hoshiBmaTchi/backend/api-gateway/handlers"
	"github.com/gin-gonic/gin"
)

func SetupSettingsRoutes(router *gin.Engine, settingsHandler *handlers.SettingsHandler, authHandler *handlers.AuthHandler) {
	settings := router.Group("/v1/settings")
	
	// Ensure Auth Middleware is applied
	// Assuming you have a middleware that extracts the token and sets "userID" in context
	settings.Use(authHandler.AuthMiddleware()) 
	{
		settings.PUT("/profile", settingsHandler.UpdateProfile)
		settings.GET("/preferences", settingsHandler.GetSettings)
		settings.PUT("/notifications", settingsHandler.UpdateNotifications)
		settings.PUT("/privacy", settingsHandler.UpdatePrivacy)

		settings.GET("/close-friends", settingsHandler.GetCloseFriends)
		settings.POST("/close-friends/:id", settingsHandler.AddCloseFriend)
		settings.DELETE("/close-friends/:id", settingsHandler.RemoveCloseFriend)

		settings.GET("/story-hide", settingsHandler.GetHiddenStoryUsers)
		settings.POST("/story-hide/:id", settingsHandler.HideStoryFromUser)
		settings.DELETE("/story-hide/:id", settingsHandler.UnhideStoryFromUser)

		settings.POST("/verification-request", settingsHandler.RequestVerification)
	}
}