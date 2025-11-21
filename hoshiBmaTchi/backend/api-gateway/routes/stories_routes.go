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

		stories.GET("/user/:userId", storiesHandler.GetUserStories)
		stories.GET("/following", storiesHandler.GetFollowingStories)
		stories.GET("/:storyId", storiesHandler.GetStory)

		stories.POST("/:storyId/view", storiesHandler.ViewStory)
		stories.POST("/:storyId/like", storiesHandler.LikeStory)
		stories.DELETE("/:storyId/like", storiesHandler.UnlikeStory)
		stories.POST("/:storyId/reply", storiesHandler.ReplyToStory)
		stories.GET("/:storyId/replies", storiesHandler.GetStoryReplies)
		stories.POST("/:storyId/share", storiesHandler.ShareStory)

		stories.GET("/:storyId/viewers", storiesHandler.GetStoryViewers)

		stories.DELETE("/:storyId", storiesHandler.DeleteStory)
	}
}
