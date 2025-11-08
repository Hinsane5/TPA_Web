package handlers

import (
	"context"
	"net/http"
	"time"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthHandler struct{
	UserClient pb.UserServiceClient
}

type registerRequestJSON struct {
	Name              string `json:"name"`
	Username          string `json:"username"`
	Email             string `json:"email"`
	Password          string `json:"password"`
	DateOfBirth       string `json:"date_of_birth"`
	Gender            string `json:"gender"`
	ProfilePictureUrl string `json:"profile_picture_url"`
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
		DateOfBirth:       dobPb,
		Gender:            jsonReq.Gender,
		ProfilePictureUrl: jsonReq.ProfilePictureUrl,
	}

	res, err := h.UserClient.RegisterUser(context.Background(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}