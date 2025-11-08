package handlers

import (
	"context"
	"time"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/ports"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)



type UserHandler struct{
	pb.UnimplementedUserServiceServer
	repo ports.UserRepository
}

func NewUserHandler(repo ports.UserRepository) *UserHandler{
	return &UserHandler{repo: repo}
}

func (h *UserHandler) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error){

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	var dob time.Time
	if req.DateOfBirth != nil && req.DateOfBirth.IsValid(){
		dob = req.DateOfBirth.AsTime()
	}

	newUser := &domain.User{
		Name: req.Name,
		Username: req.Username,
        Email:    req.Email,
        Password: hashedPassword,
		DateOfBirth: dob,
		Gender: req.Gender,
		ProfilePictureURL: req.ProfilePictureUrl,
	}

	err = h.repo.Save(newUser)
	if err != nil {
		return nil, err
	}

	var dobTimestamp *timestamppb.Timestamp
	if !newUser.DateOfBirth.IsZero(){
		dobTimestamp = timestamppb.New(newUser.DateOfBirth)
	}

	return &pb.RegisterUserResponse{
		UserId:            newUser.ID.String(),
		Name:              newUser.Name,
		Username:          newUser.Username,
		Email:             newUser.Email,
		DateOfBirth:       dobTimestamp,
		Gender:            newUser.Gender,
		ProfilePictureUrl: newUser.ProfilePictureURL,
	}, nil
}

