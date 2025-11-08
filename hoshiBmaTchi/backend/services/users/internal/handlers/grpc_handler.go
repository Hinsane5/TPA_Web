package handlers

import (
	"context"
	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/ports"
	"strconv"
)



type UserHandler struct{
	pb.UnimplementedUserServiceServer
	repo ports.UserRepository
}

func NewUserHandler(repo ports.UserRepository) *UserHandler{
	return &UserHandler{repo: repo}
}

func (h *UserHandler) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error){

	hashedPassword := req.Password

	newUser := &domain.User{
		Name: req.Name,
		Username: req.Username,
        Email:    req.Email,
        Password: hashedPassword,
	}

	err := h.repo.Save(newUser)
	if err != nil {
		return nil, err
	}

	userIDString := strconv.FormatUint(uint64(newUser.ID), 10)

	return &pb.RegisterUserResponse{
		UserId:  userIDString,
		Message: "User registered successfully",
	}, nil
}

