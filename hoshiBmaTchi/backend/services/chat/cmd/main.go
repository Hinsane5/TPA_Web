package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Configuration
const (
	ChatServiceURL = "http://chat-service:8080"
	AuthServiceURL = "http://auth-service:50051"
)

func main() {
	r := gin.Default()

	// 1. CORS Middleware (Crucial for Vue)
	r.Use(CORSMiddleware())

	// 2. Public Routes (Auth)
	// Note: Authentication usually happens via JSON REST, so we proxy normally.
	r.POST("/login", reverseProxy(AuthServiceURL))
	r.POST("/signup", reverseProxy(AuthServiceURL))

	// 3. Protected API Routes
	api := r.Group("/api")
	api.Use(AuthMiddleware())
	{
		// Forward REST requests to Chat Service
		// The Chat Service must have matching HTTP handlers (e.g., GET /chats)
		api.GET("/chats", reverseProxy(ChatServiceURL))
		api.GET("/chats/search", reverseProxy(ChatServiceURL)) // Search Implementation
		api.GET("/chats/:id/messages", reverseProxy(ChatServiceURL))
		
		// Forward other service requests here...
	}

	// 4. WebSocket Route
	// Use AuthMiddleware to ensure they have a token before upgrading connection
	r.GET("/ws", AuthMiddleware(), func(c *gin.Context) {
		proxyWebSocket(c, ChatServiceURL)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Printf("Gateway running on port %s", port)
	log.Fatal(r.Run(":" + port))
}

func getJWTSecret() string {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        // Fallback for development only - remove in production!
        return "YOUR_SUPER_SECRET_KEY" 
    }
    return secret
}

// AuthMiddleware validates JWT and injects User ID into headers
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1. Check Authorization Header
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 2. Fallback: Check Query Param (Common for WebSockets)
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			return
		}

		// 3. Parse & Validate Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			secret := getJWTSecret()
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// 4. Extract Claims & Set Header for Downstream Services
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Assuming the JWT has a "sub" or "user_id" field
			if userID, ok := claims["user_id"].(string); ok {
				// IMPORTANT: Add X-User-ID header so Chat Service knows who this is
				c.Request.Header.Set("X-User-ID", userID)
			} else {
				// Try "sub" if user_id is missing
				if sub, ok := claims["sub"].(string); ok {
					c.Request.Header.Set("X-User-ID", sub)
				}
			}
		}

		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // In production, change * to your frontend domain
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Upgrade, Connection")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		
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
		// Update the request to match the target
		c.Request.Host = targetURL.Host
		c.Request.URL.Host = targetURL.Host
		c.Request.URL.Scheme = targetURL.Scheme
		
		// X-User-ID header set in AuthMiddleware will be forwarded automatically
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func proxyWebSocket(c *gin.Context, target string) {
	targetURL, err := url.Parse(target)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid target URL"})
		return
	}

	// Switch protocol to ws/wss
	if targetURL.Scheme == "https" {
		targetURL.Scheme = "wss"
	} else {
		targetURL.Scheme = "ws"
	}
	
	// Ensure the path matches the Chat Service's WS endpoint
	targetURL.Path = "/ws" 

	// Create a custom Director to forward the handshake details
	director := func(req *http.Request) {
		req.URL = targetURL
		req.Host = targetURL.Host
		
		// Forward the User ID query param if strictly needed by the backend WS handler
		// (Since WS handshakes can't always easily read custom headers added by middleware)
		userID := c.Request.Header.Get("X-User-ID")
		q := req.URL.Query()
		q.Add("user_id", userID)
		req.URL.RawQuery = q.Encode()
	}

	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(c.Writer, c.Request)
}