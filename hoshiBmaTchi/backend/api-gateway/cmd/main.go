package main

import (
	"log"

	"github.com/Hinsane5/hoshiBmaTchi/backend/api-gateway/handlers"
	"github.com/Hinsane5/hoshiBmaTchi/backend/api-gateway/routes"
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

	router := gin.Default()

	authHandler := handlers.NewAuthHandler(userClient)

	postsHandler := handlers.NewPostsHandler(postsClient, userClient)

	routes.SetupAuthRoutes(router, userClient)

	routes.SetupPostsRoutes(router, postsHandler, authHandler)
	
	log.Println("API Gateway listening on port :8080")
	router.Run(":8080")
}
 