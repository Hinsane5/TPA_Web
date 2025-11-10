package handlers

import (
	"context"
	"net/http"
	"time"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthHandler struct{
	UserClient pb.UserServiceClient
}

type sendOtpJSON struct {
	Email string `json:"email" binding:"required"`
}

type registerRequestJSON struct {
	Name              string `json:"name"`
	Username          string `json:"username"`
	Email             string `json:"email"`
	Password          string `json:"password"`
	ConfirmPassword   string `json:"confirm_password"`
	DateOfBirth       string `json:"date_of_birth"`
	Gender            string `json:"gender"`
	ProfilePictureUrl string `json:"profile_picture_url"`
	SubscribeToNewsletter bool `json:"subscribe_to_newsletter"`
	Enable2FA             bool `json:"enable_2fa"`
	OtpCode               string `json:"otp_code"`
}

func (h *AuthHandler) SendOtp(c *gin.Context){
	var jsonReq sendOtpJSON
	if err := c.ShouldBindJSON(&jsonReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	res, err := h.UserClient.SendOtp(context.Background(), &pb.SendOtpRequest{
		Email: jsonReq.Email,
	})

	if err != nil {
		// Handle gRPC errors (like rate limit)
		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{"error": s.Message()})
				return
			case codes.ResourceExhausted:
				c.JSON(http.StatusTooManyRequests, gin.H{"error": s.Message()})
				return
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC error: " + s.Message()})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var jsonReq registerRequestJSON
	if err := c.ShouldBindJSON(&jsonReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	dob, err := time.Parse(time.RFC3339, jsonReq.DateOfBirth)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_of_birth format. Use ISO 8601 (RFC3339)."})
		return
	}

	dobPb := timestamppb.New(dob)

	grpcReq := &pb.RegisterUserRequest{
		Name:              jsonReq.Name,
		Username:          jsonReq.Username,
		Email:             jsonReq.Email,
		Password:          jsonReq.Password,
		ConfirmPassword:   jsonReq.ConfirmPassword,
		DateOfBirth:       dobPb,
		Gender:            jsonReq.Gender,
		ProfilePictureUrl: jsonReq.ProfilePictureUrl,
		SubscribeToNewsletter: jsonReq.SubscribeToNewsletter,
		Enable_2Fa:            jsonReq.Enable2FA,
		OtpCode:               jsonReq.OtpCode,
	}

	res, err := h.UserClient.RegisterUser(context.Background(), grpcReq)
	if err != nil {

		if s, ok := status.FromError(err); ok {
			switch s.Code(){
			case codes.InvalidArgument:
				c.JSON((http.StatusBadRequest), gin.H{"error": s.Message()})
				return
			
			case codes.AlreadyExists:
				// If user exists, return a 409 Conflict (or 400)
				c.JSON(http.StatusConflict, gin.H{"error": s.Message()})
				return
			default:
				// For any other gRPC error, return a 500
				c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC error: " + s.Message()})
				return
					
			}
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}