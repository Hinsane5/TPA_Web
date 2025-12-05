package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/Hinsane5/hoshiBmaTchi/backend/proto/stories"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/clients"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/core/ports"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/events"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/handlers"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/repositories"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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
		&domain.StoryReply{},
		&domain.StoryShare{},
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

	rabbitURL := os.Getenv("RABBITMQ_URL")
    if rabbitURL == "" {
        rabbitURL = "amqp://guest:guest@localhost:5672/"
    }

	publisher, err := events.NewEventPublisher(rabbitURL)
    if err != nil {
        log.Printf("Failed to initialize RabbitMQ publisher: %v", err)
    } else {
        defer publisher.Close()
        log.Println("Connected to RabbitMQ")
    }

	chatClient, err := clients.NewChatServiceClient(os.Getenv("CHAT_SERVICE_URL"))
	if err != nil {
		log.Fatalf("Failed to create chat client: %v", err)
	}

	minioClient, err := minio.New(os.Getenv("MINIO_ENDPOINT"), &minio.Options{
        Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ACCESS_KEY_ID"), os.Getenv("MINIO_SECRET_ACCESS_KEY"), ""),
        Secure: os.Getenv("MINIO_USE_SSL") == "true",
    })
    if err != nil {
        log.Fatal(err)
    }

	handler := handlers.NewGRPCHandler(repo, redisRepo, userClient, chatClient, publisher, minioClient, os.Getenv("MINIO_BUCKET_NAME"),)

	go startCleanupRoutine(repo)

	lis, err := net.Listen("tcp", ":50054")
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

