package handlers

import (
	"context"
	"net/http"
	"strings"
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
	TurnstileToken	   string `json:"turnstile_token"`
}

type googleLoginJson struct {
	IdToken string `json:"id_token" binding:"required"`
}

type loginUserJSON struct {
	EmailOrUsername string `json:"email_or_username" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

type verify2FAJSON struct {
	Email   string `json:"email" binding:"required"`
	OtpCode string `json:"otp_code" binding:"required"`
}

type requestPasswordResetJSON struct {
	Email string `json:"email" binding:"required"`
}

type performPasswordResetJSON struct {
	Token           string `json:"token" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

func (h *AuthHandler) LoginUser(c *gin.Context){
	var jsonReq loginUserJSON
	if err := c.ShouldBindJSON(&jsonReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and Password are required"})
		return;
	}

	res, err := h.UserClient.LoginUser(context.Background(), &pb.LoginUserRequest{
		EmailOrUsername:    jsonReq.EmailOrUsername,
		Password: jsonReq.Password,
	})

	if err != nil {
		if s, ok := status.FromError(err); ok {
			switch s.Code() {
				case codes.Unauthenticated:
					c.JSON(http.StatusUnauthorized, gin.H{"error": s.Message()})
					return
				case codes.PermissionDenied:
					c.JSON(http.StatusForbidden, gin.H{"error": s.Message()})
					return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// if res.TwoFaRequired {
	// 	c.JSON(http.StatusOK, gin.H{"message": "2FA code sent to your email."})
	// 	return
	// }

	c.JSON(http.StatusOK, res)
}

func (h *AuthHandler) VerifyLogin2FA(c *gin.Context) {
	var jsonReq verify2FAJSON
	if err := c.ShouldBindJSON(&jsonReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and OTP code are required"})
		return
	}

	res, err := h.UserClient.VerifyLogin2FA(context.Background(), &pb.VerifyLogin2FARequest{
		Email:   jsonReq.Email,
		OtpCode: jsonReq.OtpCode,
	})

	if err != nil {
		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.InvalidArgument {
				c.JSON(http.StatusBadRequest, gin.H{"error": s.Message()})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *AuthHandler) LoginWithGoogle(c *gin.Context){
	var jsonReq googleLoginJson
	if err := c.ShouldBindJSON(&jsonReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id_token is required"})
		return
	}

	res, err := h.UserClient.LoginWithGoogle(context.Background(), &pb.LoginWithGoogleRequest{
		IdToken: jsonReq.IdToken,
	})

	if err != nil {
		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.Unauthenticated {
				c.JSON(http.StatusUnauthorized, gin.H{"error": s.Message()})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, res)
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
		TurnstileToken: jsonReq.TurnstileToken,
	}

	res, err := h.UserClient.RegisterUser(context.Background(), grpcReq)
	if err != nil {

		if s, ok := status.FromError(err); ok {
			switch s.Code(){
			case codes.InvalidArgument:
				c.JSON((http.StatusBadRequest), gin.H{"error": s.Message()})
				return
			
			case codes.AlreadyExists:
				c.JSON(http.StatusConflict, gin.H{"error": s.Message()})
				return
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC error: " + s.Message()})
				return
					
			}
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var jsonReq requestPasswordResetJSON
	if err := c.ShouldBindJSON(&jsonReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	_, err := h.UserClient.RequestPasswordReset(context.Background(), &pb.RequestPasswordResetRequest{
		Email: jsonReq.Email,
	})

	if err != nil {
		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.PermissionDenied {
				c.JSON(http.StatusForbidden, gin.H{"error": s.Message()})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "If your account exists, a password reset link has been sent."})
}

func (h *AuthHandler) PerformPasswordReset(c *gin.Context) {
	var jsonReq performPasswordResetJSON
	if err := c.ShouldBindJSON(&jsonReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token and new password are required"})
		return
	}

	res, err := h.UserClient.PerformPasswordReset(context.Background(), &pb.PerformPasswordResetRequest{
		Token:           jsonReq.Token,
		NewPassword:     jsonReq.NewPassword,
		ConfirmPassword: jsonReq.ConfirmPassword,
	})

	if err != nil {
		if s, ok := status.FromError(err); ok {
			if s.Code() == codes.InvalidArgument || s.Code() == codes.Unauthenticated {
				c.JSON(http.StatusBadRequest, gin.H{"error": s.Message()})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization token"})
			c.Abort()
			return
		}

		res, err := h.UserClient.ValidateToken(context.Background(), &pb.ValidateTokenRequest{
			Token: tokenString,
		})
		
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token or internal error", "details": err.Error()})
			c.Abort()
			return
		}

		if !res.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("userID", res.UserId)

		c.Next()
	}
}

func NewAuthHandler(userClient pb.UserServiceClient) *AuthHandler {
	return &AuthHandler{
		UserClient: userClient,
	}
}

func (h *AuthHandler) GetUserProfile(c *gin.Context) {
    userID := c.Param("id")

	var viewerID string
    if val, exists := c.Get("userID"); exists {
        viewerID = val.(string)
    }

    res, err := h.UserClient.GetUserProfile(context.Background(), &pb.GetUserProfileRequest{
        UserId:   userID,
        ViewerId: viewerID,
    })

    if err != nil {
        if s, ok := status.FromError(err); ok {
            if s.Code() == codes.NotFound {
                c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
                return
            }
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile"})
        return
    }

    c.JSON(http.StatusOK, res)
}

func (h *AuthHandler) FollowUser(c *gin.Context) {
    targetUserID := c.Param("id")
    currentUserID, _ := c.Get("userID")

    _, err := h.UserClient.FollowUser(context.Background(), &pb.FollowUserRequest{
        FollowerId:  currentUserID.(string),
        FollowingId: targetUserID,
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Followed successfully"})
}

func (h *AuthHandler) UnfollowUser(c *gin.Context) {
    targetUserID := c.Param("id")
    currentUserID, _ := c.Get("userID")

    _, err := h.UserClient.UnfollowUser(context.Background(), &pb.UnfollowUserRequest{
        FollowerId:  currentUserID.(string),
        FollowingId: targetUserID,
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Unfollowed successfully"})
}

func (h *AuthHandler) SearchUsers (c *gin.Context){
	query := c.Query("q")
	if query == ""{
		c.JSON(http.StatusOK, gin.H{"users": []interface{}{}})
		return
	}

	userID, _ := c.Get("userID")

    res, err := h.UserClient.SearchUsers(context.Background(), &pb.SearchUsersRequest{
        Query:  query,
        UserId: userID.(string),
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"users": res.Users})
}