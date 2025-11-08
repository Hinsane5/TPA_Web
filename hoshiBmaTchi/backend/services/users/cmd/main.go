package main

import (
	"fmt"
	"log"
	"net"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/handlers"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/repositories"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main(){

	dsn := "host=postgres user=admin password=postgres_password_123 dbname=hoshi_users_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil{
		log.Fatalf("Failed to connect to database: %v", err)
	}

	repo := repositories.NewGormUserRepository(db)

	handler := handlers.NewUserHandler(repo)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil{
		log.Fatalf("Failed to listed %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, handler)

	fmt.Println("User gRPC server listening on port :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}