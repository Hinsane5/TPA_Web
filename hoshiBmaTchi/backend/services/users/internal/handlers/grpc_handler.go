package handlers

import (
	"context"
	"errors"
	"regexp"
	"time"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/ports"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/utils"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)



type UserHandler struct{
	pb.UnimplementedUserServiceServer
	repo ports.UserRepository
}

func NewUserHandler(repo ports.UserRepository) *UserHandler{
	return &UserHandler{repo: repo}
}

func (h *UserHandler) validateRegisterRequest(req *pb.RegisterUserRequest) error{
	if len(req.Name) <= 4 {
		return status.Error(codes.InvalidArgument, "name must be at least 5 characters long")
	}

	if !regexp.MustCompile(`^[a-zA-Z ]+$`).MatchString(req.Name) {
		return status.Error(codes.InvalidArgument, "name must not contain symbols or numbers")
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.com$`).MatchString(req.Email){
		return status.Error(codes.InvalidArgument, "Invalid email format")
	}

	if req.Password != req.ConfirmPassword{
		return status.Error(codes.InvalidArgument, "password and confirm password must be the same")
	}

	if len(req.Password) < 8 {
		return status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}

	if !regexp.MustCompile(`[A-Z]`).MatchString(req.Password) {
		return status.Error(codes.InvalidArgument, "password must contain at least one uppercase letter")
	}

	if !regexp.MustCompile(`[0-9]`).MatchString(req.Password) {
		return status.Error(codes.InvalidArgument, "password must contain at least one number")
	}

	if !regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(req.Password) {
		return status.Error(codes.InvalidArgument, "password must contain at least one special character")
	}

	if req.Gender != "male" && req.Gender != "female" {
		return status.Error(codes.InvalidArgument, "gender must be male or female")
	}

	if req.DateOfBirth == nil || !req.DateOfBirth.IsValid() {
		return status.Error(codes.InvalidArgument, "date of birth is required")
	}

	birth_date := req.DateOfBirth.AsTime()
	thirteenthBirthday := birth_date.AddDate(13, 0, 0)
	if time.Now().Before(thirteenthBirthday) {
		return status.Error(codes.InvalidArgument, "you must be at least 13 years old to register")
	}

	return nil
}

func (h *UserHandler) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error){

	if err := h.validateRegisterRequest(req); err != nil {
		return nil, err
	}

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
		SubscribedToNewsletters: req.SubscribeToNewsletter,
		TwoFactorEnabled:      req.Enable_2Fa,
	}

	err = h.repo.Save(newUser)
	if err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505"{
			return nil, status.Error(codes.AlreadyExists, "username or email already exists")
		}

		return nil, status.Error(codes.Internal, "failed to save user")
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

