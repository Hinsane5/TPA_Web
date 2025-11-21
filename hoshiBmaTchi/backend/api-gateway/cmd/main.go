package main

import (
	"log"

	"github.com/Hinsane5/hoshiBmaTchi/backend/api-gateway/handlers"
	"github.com/Hinsane5/hoshiBmaTchi/backend/api-gateway/routes"
	storiesProto "github.com/Hinsane5/hoshiBmaTchi/backend/proto/stories"
	postsProto "github.com/Hinsane5/hoshiBmaTchi/backend/proto/posts"
	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main(){
	conn, err := grpc.NewClient("users-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	
	if err != nil{
		log.Fatalf("Did not connect: %v", err)
	}

	defer conn.Close()

	userClient := pb.NewUserServiceClient(conn)

	postsConn, err := grpc.NewClient("posts-service:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect posts-service: %v", err)
	}
	defer postsConn.Close()
	postsClient := postsProto.NewPostsServiceClient(postsConn)
	log.Println("Connected to gRPC posts-service")

	storiesConn, err := grpc.NewClient("stories-service:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect to stories-service: %v", err)
    }
	defer storiesConn.Close()
    storiesClient := storiesProto.NewStoriesServiceClient(storiesConn)
    log.Println("Connected to gRPC stories-service")

	router := gin.Default()

	authHandler := handlers.NewAuthHandler(userClient)
	postsHandler := handlers.NewPostsHandler(postsClient, userClient)
	storiesHandler := handlers.NewStoriesHandler(storiesClient)

	routes.SetupAuthRoutes(router, userClient)

	routes.SetupPostsRoutes(router, postsHandler, authHandler)

	routes.SetupChatRoutes(router, authHandler)

	routes.SetupStoriesRoutes(router, storiesHandler, authHandler)
	
	log.Println("API Gateway listening on port :8080")
	router.Run(":8080")
}
 