package main

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"time"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/posts"
	userPb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/clients"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/handlers"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/repositories"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed at connect to database: %v", err)
	}
	log.Println("Connected to database")

	err = db.AutoMigrate(&domain.Post{}, &domain.PostLike{}, &domain.PostComment{})
	if err != nil {
		log.Fatalf("Failed to automigrate: %v", err)
	}

	log.Println("Automigrate successfully")

	var minioClient *minio.Client // Use the *minio.Client type
	var minioErr error
	
	const maxRetries = 10
	for i := 0; i < maxRetries; i++ {
		// This calls the code you just posted
		minioClient, minioErr = clients.NewMinIOClient()
		if minioErr == nil {
			// Success!
			break
		}
		
		log.Printf("Gagal terhubung ke MinIO (attempt %d/%d): %v. Retrying in 5 seconds...", i+1, maxRetries, minioErr)
		time.Sleep(5 * time.Second)
	}
	
	// If it still failed after all retries, then exit
	if minioErr != nil {
		log.Fatalf("Failed to initialize MinIO client after multiple retries: %v", minioErr)
	}

	bucketName := os.Getenv("MINIO_BUCKET_NAME")
	if bucketName == "" {
		log.Fatal("FATAL: MINIO_BUCKET_NAME environment variable is not set.")
	}

	publicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
	if publicEndpoint == "" {
		log.Fatal("FATAL: MINIO_PUBLIC_ENDPOINT environment variable is not set.")
	}

	publicURL, err := url.Parse(publicEndpoint)
	if err != nil {
		log.Fatalf("Failed to parse MINIO_PUBLIC_ENDPOINT: %v", err)
	}

	presignClient, err := minio.New(publicURL.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ACCESS_KEY_ID"), os.Getenv("MINIO_SECRET_ACCESS_KEY"), ""),
		Secure: publicURL.Scheme == "https",
		Region: "us-east-1",
	})
	if err != nil {
		log.Fatalf("Failed to create presign MinIO client: %v", err)
	}

	userServiceAddr := "users-service:50051" 
	userConn, err := grpc.Dial(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to users-service: %v", err)
	}
	defer userConn.Close()
	
	userClient := userPb.NewUserServiceClient(userConn)

	postRepo := repositories.NewGormPostRepository(db)
	grpcServer := handlers.NewGRPCServer(postRepo, minioClient, presignClient, bucketName, publicEndpoint, userClient)

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50052"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen port %s: %v", grpcPort, err)
	}

	s := grpc.NewServer()
	pb.RegisterPostsServiceServer(s, grpcServer)

	log.Printf("Server gRPC 'posts-service' running in port: %s", grpcPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to run gRPC server: %v", err)
	}
	

}