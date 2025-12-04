package routes

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/Hinsane5/hoshiBmaTchi/backend/api-gateway/handlers"
	"github.com/gin-gonic/gin"
)

func SetupChatRoutes(router *gin.Engine, authHandler *handlers.AuthHandler) {
	targetStr := "http://chat-service:8080"
	
	target, _ := url.Parse(targetStr)

	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host 
		
		if strings.HasPrefix(req.URL.Path, "/api/chats") {
			req.URL.Path = strings.Replace(req.URL.Path, "/api/chats", "/chats", 1)
		}
	}

	proxyHandler := func(c *gin.Context) {
		if userID, exists := c.Get("userID"); exists {
			c.Request.Header.Set("X-User-ID", userID.(string))
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}

	chatGroup := router.Group("/api/chats")
	chatGroup.Use(authHandler.AuthMiddleware())
	{
		chatGroup.GET("", proxyHandler)
		chatGroup.POST("", proxyHandler)

		chatGroup.Any("/*path", proxyHandler)
	}

	router.GET("/ws", authHandler.AuthMiddleware(), func(c *gin.Context) {
		if userID, exists := c.Get("userID"); exists {
			c.Request.Header.Set("X-User-ID", userID.(string))
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	})
}