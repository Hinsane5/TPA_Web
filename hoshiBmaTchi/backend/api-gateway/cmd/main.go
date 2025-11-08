package main

import (
	"log"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/Hinsane5/hoshiBmaTchi/backend/api-gateway/routes"
)

func main(){
	conn, err := grpc.NewClient("users-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil{
		log.Fatalf("Did not connect: %v", err)
	}

	defer conn.Close()

	userClient := pb.NewUserServiceClient(conn)

	router := gin.Default()

	routes.SetupAuthRoutes(router, userClient)
	
	log.Println("API Gateway listening on port :8080")
	router.Run(":8080")

}
 