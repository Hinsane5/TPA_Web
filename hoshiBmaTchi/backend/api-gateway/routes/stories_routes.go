package routes

import (
    "github.com/Hinsane5/hoshiBmaTchi/backend/api-gateway/handlers"
    "github.com/gin-gonic/gin"
)

func SetupStoriesRoutes(router *gin.Engine, storiesHandler *handlers.StoriesHandler, authHandler *handlers.AuthHandler) {
    stories := router.Group("/api/stories") 
    
    stories.Use(authHandler.AuthMiddleware())
    {
        stories.POST("/", storiesHandler.CreateStory)
        
        stories.GET("/upload-url", storiesHandler.GenerateUploadURL) 

        stories.GET("/archive", storiesHandler.GetArchivedStories)

        stories.GET("", storiesHandler.GetStory) 
        stories.DELETE("", storiesHandler.DeleteStory) 

        stories.GET("/user", storiesHandler.GetUserStories)
        
        stories.GET("/following", storiesHandler.GetFollowingStories)

        stories.POST("/view", storiesHandler.ViewStory)
        stories.POST("/like", storiesHandler.LikeStory)
        
        stories.POST("/unlike", storiesHandler.UnlikeStory) 
        
        stories.POST("/reply", storiesHandler.ReplyToStory)
        stories.GET("/replies", storiesHandler.GetStoryReplies)
        
        stories.POST("/share", storiesHandler.ShareStory)
        stories.GET("/viewers", storiesHandler.GetStoryViewers)
    }
}