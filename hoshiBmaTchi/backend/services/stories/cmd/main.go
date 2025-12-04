package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/stories"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/clients"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/core/ports"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/handlers"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/repositories"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		getEnv("DB_HOST", "postgres"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "stories_db"),
		getEnv("DB_PORT", "5432"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(
		&domain.Story{},
		&domain.StoryView{},
		&domain.StoryLike{},
		&domain.StoryReply{},
		&domain.StoryShare{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_story_user_view ON story_views(story_id, user_id) WHERE deleted_at IS NULL")
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_story_user_like ON story_likes(story_id, user_id) WHERE deleted_at IS NULL")

	repo := repositories.NewGormStoryRepository(db)

	userClient, err := clients.NewUserServiceClient(getEnv("USER_SERVICE_URL", "users-service:50051"))
	if err != nil {
		log.Fatalf("Failed to connect to user service: %v", err)
	}
	defer userClient.Close()

	chatClient, err := clients.NewChatServiceClient(getEnv("CHAT_SERVICE_URL", "chat-service:50052"))
	if err != nil {
		log.Fatalf("Failed to connect to chat service: %v", err)
	}
	defer chatClient.Close()

	handler := handlers.NewGRPCHandler(repo, userClient, chatClient)

	go startCleanupRoutine(repo)

	port := getEnv("GRPC_PORT", "50053")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterStoriesServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	go func() {
		log.Printf("Stories service listening on :%s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	grpcServer.GracefulStop()
	log.Println("Server stopped")
}

func startCleanupRoutine(repo ports.StoryRepository) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		if err := repo.DeleteExpiredStories(ctx); err != nil {
			log.Printf("Error cleaning up expired stories: %v", err)
		} else {
			log.Println("Successfully cleaned up expired stories")
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

