package routes

import (
	"github.com/Hinsane5/hoshiBmaTchi/backend/api-gateway/handlers"
	"github.com/gin-gonic/gin"
)

func SetupPostsRoutes(r *gin.Engine, postsHandler *handlers.PostsHandler, authHandler *handlers.AuthHandler) {
    
    
    postsRoutes := r.Group("/api/v1/posts")
    postsRoutes.Use(authHandler.AuthMiddleware())
    {

        postsRoutes.GET("/generate-upload-url", postsHandler.GenerateUploadURL)

        postsRoutes.POST("", postsHandler.CreatePost)

        postsRoutes.GET("/user/:userID", postsHandler.GetPostsByUserID)

        postsRoutes.POST("/:postID/like", postsHandler.LikePost)
        postsRoutes.DELETE("/:postID/like", postsHandler.UnlikePost)

        postsRoutes.POST("/:postID/comments", postsHandler.CreateComment)
        postsRoutes.GET("/:postID/comments", postsHandler.GetCommentsForPost)

        postsRoutes.GET("/feed", postsHandler.GetHomeFeed)

        postsRoutes.POST("/:postID/save", postsHandler.ToggleSavePost)
        postsRoutes.GET("/collections", postsHandler.GetUserCollections)
        postsRoutes.POST("/collections", postsHandler.CreateCollection)
        postsRoutes.GET("/mentions/:target_id", postsHandler.GetUserMentions)
    }

    reelsRoutes := r.Group("/api/v1/reels")
    reelsRoutes.Use(authHandler.AuthMiddleware()) 
    {
        reelsRoutes.GET("/feed", postsHandler.GetReelsFeed)
    }
}