package routes

import (
	"github.com/Hinsane5/hoshiBmaTchi/backend/api-gateway/handlers"
	"github.com/gin-gonic/gin"
)

func SetupAiRoutes(router *gin.Engine, aiHandler *handlers.AIHandler){
	api := router.Group("/api/v1") 
    {
        api.POST("/summarize", aiHandler.SummarizeCaption) 
    }
}