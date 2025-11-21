package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/chat/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/chat/internal/repositories"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/chat/internal/ws"
	chatHttp "github.com/Hinsane5/hoshiBmaTchi/backend/services/chat/internal/delivery/http"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Configuration
const (
	ChatServiceURL = "http://chat-service:8080"
	AuthServiceURL = "http://auth-service:50051"
)

func main() {
	// 1. Load Configuration & Connect to Database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Optional: AutoMigrate your models to ensure tables exist
	// Make sure your domain models are correctly defined
	err = db.AutoMigrate(&domain.Conversation{}, &domain.Participant{}, &domain.Message{})
	if err != nil {
		log.Printf("Warning: AutoMigration failed: %v", err)
	}

	// 2. Connect to Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// 3. Initialize Architecture Layers
	// Repository
	chatRepo := repositories.NewChatRepository(db)

	// WebSocket Hub
	hub := ws.NewHub(rdb)
	go hub.Run() // Run the hub in a separate goroutine to handle broadcasting

	// Handler
	chatHandler := chatHttp.NewChatHandler(chatRepo, hub)

	// 4. Setup Router
	r := gin.Default()

	// 5. Register Routes
	// This will register /chats, /ws, etc. matching what the Gateway forwards
	chatHandler.RegisterRoutes(r)

	// 6. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Chat Service running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
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