package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ChatServiceURL = "http://chat-service:8080"
	AuthServiceURL = "http://auth-service:50051" 
)

func main() {
	r := gin.Default()

	r.Use(CORSMiddleware())

	r.POST("/login", reverseProxy(AuthServiceURL)) 
	r.POST("/signup", reverseProxy(AuthServiceURL))

	protected := r.Group("/api")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/chats", reverseProxy(ChatServiceURL))
		protected.GET("/chats/:id/messages", reverseProxy(ChatServiceURL))
	}

	r.GET("/ws", AuthMiddleware(), func(c *gin.Context) {
		proxyWebSocket(c, ChatServiceURL)
	})

	log.Fatal(r.Run(":8000")) 
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func reverseProxy(target string) gin.HandlerFunc {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatal("Invalid target URL:", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	return func(c *gin.Context) {
		c.Request.URL.Host = targetURL.Host
		c.Request.URL.Scheme = targetURL.Scheme
		c.Request.Header.Set("X-Forwarded-Host", c.Request.Header.Get("Host"))
		c.Request.Host = targetURL.Host

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func proxyWebSocket(c *gin.Context, target string) {
	targetURL, err := url.Parse(target)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid target URL"})
		return
	}

	if targetURL.Scheme == "https" {
		targetURL.Scheme = "wss"
	} else {
		targetURL.Scheme = "ws"
	}
	targetURL.Path = "/ws"

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Director = func(req *http.Request) {
		req.URL = targetURL
		req.Host = targetURL.Host
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}