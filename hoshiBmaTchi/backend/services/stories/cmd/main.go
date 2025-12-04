package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/Hinsane5/hoshiBmaTchi/backend/proto/stories"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/clients"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/core/ports"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/handlers"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/repositories"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(
        &domain.Story{},
        &domain.StoryView{},
        &domain.StoryLike{},
        &domain.StoryVisibility{},
    )
    if err != nil {
        log.Fatalf("Failed to migrate database: %v", err)
    }

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})

	repo := repositories.NewGormStoryRepository(db)
	redisRepo := repositories.NewRedisStoryRepository(rdb)

	userClient, err := clients.NewUserServiceClient(os.Getenv("USER_SERVICE_URL"))
	if err != nil {
		log.Fatalf("Failed to create user client: %v", err)
	}

	chatClient, err := clients.NewChatServiceClient(os.Getenv("CHAT_SERVICE_URL"))
	if err != nil {
		log.Fatalf("Failed to create chat client: %v", err)
	}

	handler := handlers.NewGRPCHandler(repo, redisRepo, userClient, chatClient)

	// 6. Start Cleanup Routine
	// Fixed "does not implement" error by updating ports (see step 2 below)
	go startCleanupRoutine(repo)

	// 7. Start Server
	lis, err := net.Listen("tcp", ":50054") // Ensure port matches docker-compose
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	stories.RegisterStoriesServiceServer(s, handler)

	log.Printf("Stories service listening on %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func startCleanupRoutine(repo ports.StoryRepository) {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		// Ensure this method exists in your repo interface
		repo.DeleteExpiredStories(context.Background())
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

