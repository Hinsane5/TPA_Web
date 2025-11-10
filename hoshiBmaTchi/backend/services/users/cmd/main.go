package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/redis/go-redis/v9"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/handlers"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/repositories"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main(){

	dsn := "host=postgres user=admin password=postgres_password_123 dbname=hoshi_users_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil{
		log.Fatalf("Failed to connect to database: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
		Password: "",
		DB: 0,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil{
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	rabbitConn, err := amqp.Dial("amqp://admin:rabbitmq_password_123@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	defer rabbitConn.Close()

	amqpChan, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer amqpChan.Close()

	err = amqpChan.ExchangeDeclare(
		"email_exchange", "direct", true, false, false, false, nil,
	)

	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	repo := repositories.NewGormUserRepository(db)

	handler := handlers.NewUserHandler(repo, rdb, amqpChan)

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