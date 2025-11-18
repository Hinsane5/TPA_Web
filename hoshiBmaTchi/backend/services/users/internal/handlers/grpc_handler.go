package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/ports"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)


type UserHandler struct{
	pb.UnimplementedUserServiceServer
	repo ports.UserRepository
	redis *redis.Client
	amqpChan *amqp.Channel
}

type turnstileResponse struct{
	Success bool `json:"success"`
}

type EmailTask struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func NewUserHandler(repo ports.UserRepository, redis *redis.Client, amqpChan *amqp.Channel) *UserHandler{
	return &UserHandler{
		repo: repo,
		redis: redis,
		amqpChan: amqpChan,
	}
}

func (h *UserHandler) SendOtp(ctx context.Context, req *pb.SendOtpRequest) (*pb.SendOtpResponse, error){
	if !regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.com$`).MatchString(req.Email) {
		return nil, status.Error(codes.InvalidArgument, "Invalid email format")
	}

	rateLimitKey := "ratelimit:" + req.Email

	set, err := h.redis.SetNX(ctx, rateLimitKey, 1, 60*time.Second).Result()
	if err != nil{
		return nil, status.Error(codes.Internal, "failed to check rate limit")
	}

	if !set{
		return nil, status.Error(codes.ResourceExhausted, "Please wait 60 seconds before resending OTP")
	}

	otp := fmt.Sprintf("%06d", rand.Intn(1000000))

	otpKey := "otp:" + req.Email

	err = h.redis.Set(ctx, otpKey, otp, 5*time.Minute).Err()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to store OTP")
	}

	emailBody := fmt.Sprintf("Your verification code is %s", otp)
	task := EmailTask{Email: req.Email, Subject: "Verify your email", Body: emailBody}
	taskBody, _ := json.Marshal(task)

	err = h.amqpChan.PublishWithContext(ctx, 
		"email_exchange",
		"send_email",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        taskBody,
		},
	)

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to publish email task")
	}

	return &pb.SendOtpResponse{
		Message: "OTP sent successfully",
	}, nil
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

func (h *UserHandler) validateTurnstile (token string) error{
	const turnstileSecretKey = "CLOUDFLARE_SECRET"

	resp, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify",
		url.Values{
			"secret": {turnstileSecretKey},
			"response": {token},
		})
	
	if err != nil {
		return status.Error(codes.Internal, "failed to verify turnstile")
	}

	if err != nil{
		return status.Error(codes.Internal, "failed to verify turnstile")
	}

	defer resp.Body.Close()

	var body turnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return status.Error(codes.Internal, "failed to parse turnstile response")
	}

	if !body.Success {
		return status.Error(codes.Unauthenticated, "Invalid CAPTCHA token")
	}

	return nil
}

func (h *UserHandler) LoginWithGoogle(ctx context.Context, req *pb.LoginWithGoogleRequest) (*pb.TokenResponse, error){
	
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		log.Println("FATAL ERROR: GOOGLE_CLIENT_ID environment variable is not set.")
		return nil, status.Error(codes.Internal, "Google login is not configured")
	}
	
	payload, err := idtoken.Validate(ctx, req.IdToken, googleClientID)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid Google token")
	}

	email := payload.Claims["email"].(string)
	name := payload.Claims["name"].(string)
	picture := payload.Claims["picture"].(string)

	user, err := h.repo.FindByEmail(email)
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, status.Error(codes.Internal, "failed to find user")
	}

	if err == gorm.ErrRecordNotFound{
		newUser := &domain.User{
			Name: name,
			Email: email,
			ProfilePictureURL: picture,
			Password: "__GOOGLE_AUTH_USER__",
			TwoFactorEnabled: false,
			SubscribedToNewsletter: false,
		}

		err = h.repo.Save(newUser)
		if err != nil {
			return nil, status.Error(codes.Internal, "Failed to create new user")
		}
		
		user = newUser
	}

	accessToken, refreshToken, err := utils.GenerateTokens(user.ID.String(), user.Email)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	return &pb.TokenResponse{
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *UserHandler) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error){

	user, err := h.repo.FindByEmailOrUsername(req.EmailOrUsername)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.Unauthenticated, "Invalid credentials")
		}
		return nil, status.Error(codes.Internal, "failed to find user")
	}

	if !user.IsActive {
		return nil, status.Error(codes.Unauthenticated, "Your account has been deactivated.")
	}
	if user.IsBanned {
		return nil, status.Error(codes.PermissionDenied, "Your account has been banned.")
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	if user.TwoFactorEnabled {
		otp := fmt.Sprintf("%06d", rand.Intn(1000000))
		otpKey := "2fa:" + user.Email 
		err = h.redis.Set(ctx, otpKey, otp, 5*time.Minute).Err()
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to store 2FA code")
		}

		emailBody := fmt.Sprintf("Your 2FA login code is: %s It expires in 5 minutes.", otp)
		task := EmailTask{Email: user.Email, Subject: "Your Login Code", Body: emailBody}
		taskBody, _ := json.Marshal(task)

		err = h.amqpChan.PublishWithContext(ctx,
			"email_exchange",
			"send_email",
			false, false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        taskBody,
			},
		)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to publish 2FA email")
		}

		return &pb.LoginUserResponse{
			TwoFaRequired: true,
			Tokens:        nil,
		}, nil
	}

	accessToken, refreshToken, err := utils.GenerateTokens(user.ID.String(), user.Email)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	return &pb.LoginUserResponse{
		TwoFaRequired: false,
		Tokens: &pb.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil

}

func (h *UserHandler) VerifyLogin2FA(ctx context.Context, req *pb.VerifyLogin2FARequest) (*pb.TokenResponse, error){
	otpKey := "2fa:" + req.Email
	storedOtp, err := h.redis.Get(ctx, otpKey).Result()
	if err == redis.Nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid or expired 2FA code")
	} else if err != nil {
		return nil, status.Error(codes.Internal, "failed to get 2FA code")
	}

	if storedOtp != req.OtpCode {
		return nil, status.Error(codes.InvalidArgument, "Invalid or expired 2FA code")
	}

	user, err := h.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to find user post-2FA")
	}

	h.redis.Del(ctx, otpKey)

	accessToken, refreshToken, err := utils.GenerateTokens(user.ID.String(), user.Email)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate tokens")
	}

	return &pb.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *UserHandler) RequestPasswordReset(ctx context.Context, req *pb.RequestPasswordResetRequest) (*pb.SendOtpResponse, error) {

	user, err := h.repo.FindByEmail(req.Email)
	if err != nil {
		log.Printf("Password reset request for non-existent user: %s", req.Email)
		return &pb.SendOtpResponse{Message: "If your account exists, a password reset link has been sent."}, nil
	}

	if user.IsBanned {
		return nil, status.Error(codes.PermissionDenied, "Banned accounts cannot reset passwords.")
	}

	resetClaims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"exp":     time.Now().Add(time.Minute * 15).Unix(), 
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, resetClaims)

	resetToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate reset token")
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	resetLink := fmt.Sprintf("http://localhost:5173/reset-password?token=%s", resetToken)
	emailBody := fmt.Sprintf("Click here to reset your password: <a href=\"%s\">%s</a>. This link expires in 15 minutes.", resetLink, resetLink)
	
	task := EmailTask{Email: user.Email, Subject: "Reset Your Password", Body: emailBody}
	taskBody, _ := json.Marshal(task)

	err = h.amqpChan.PublishWithContext(ctx,
		"email_exchange",
		"send_email", 
		false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        taskBody,
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to publish reset email")
	}

	return &pb.SendOtpResponse{Message: "If your account exists, a password reset link has been sent."}, nil
}

func (h *UserHandler) PerformPasswordReset(ctx context.Context, req *pb.PerformPasswordResetRequest) (*pb.SendOtpResponse, error) {

	if req.NewPassword != req.ConfirmPassword {
		return nil, status.Error(codes.InvalidArgument, "New passwords do not match")
	}
	if len(req.NewPassword) < 8 {
		return nil, status.Error(codes.InvalidArgument, "Password must be at least 8 characters")
	}

	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, status.Error(codes.Unauthenticated, "Invalid or expired token")
	}

	userID := claims["user_id"].(string)

	user, err := h.repo.FindByID(userID) 
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to find user")
	}

	
	if utils.CheckPasswordHash(req.NewPassword, user.Password) {
		return nil, status.Error(codes.InvalidArgument, "New password cannot be the same as the old one")
	}

	newHashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash new password")
	}

	err = h.repo.UpdatePassword(userID, newHashedPassword)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to update password")
	}

	return &pb.SendOtpResponse{Message: "Password has been reset successfully."}, nil
}

func (h *UserHandler) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error){

	// if err := h.validateTurnstile(req.TurnstileToken); err != nil {
	// 	return nil, err
	// }

	if err := h.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	otpKey := "otp:" + req.Email
	storedOtp, err := h.redis.Get(ctx, otpKey).Result()

	if err == redis.Nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid or expired OTP")
	} else if err != nil {
		return nil, status.Error(codes.Internal, "failed to get OTP")
	}

	if storedOtp != req.OtpCode{
		return nil, status.Error(codes.InvalidArgument, "Invalid or expired OTP code")
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
		SubscribedToNewsletter: req.SubscribeToNewsletter,
		TwoFactorEnabled:      req.Enable_2Fa,
	}

	err = h.repo.Save(newUser)
	if err != nil {

		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) && pqErr.Code == "23505"{
			return nil, status.Error(codes.AlreadyExists, "username or email already exists")
		}

		return nil, status.Error(codes.Internal, "failed to save user")
	}

	h.redis.Del(ctx, otpKey)

	welcomeBody := fmt.Sprintf("Hi %s, welcome to hoshiBmaTchi!", req.Name)
	task := EmailTask{Email: req.Email, Subject: "Welcome to hoshiBmaTchi!", Body: welcomeBody}
	taskBody, _ := json.Marshal(task)

	err = h.amqpChan.PublishWithContext(ctx, 
		"email_exchange",
		"email.welcome",
		false, 
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:       taskBody,
		},
	)	

	if err != nil{
		log.Printf("Failed to publish welcome email task: %v", err)
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

func (h *UserHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Token diperlukan")
	}

	claims, err := utils.ValidateToken(req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid:  false,
			UserId: "",
		}, nil 
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Token claims tidak valid")
	}

	return &pb.ValidateTokenResponse{
		Valid:  true,
		UserId: userID,
	}, nil
}

func (h *UserHandler) GetUserProfile(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.GetUserProfileResponse, error){
	if req.UserId == ""{
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	user, followers, following, err := h.repo.GetUserProfileWithStats(req.UserId)
	if err != nil {
		if err == gorm.ErrRecordNotFound{
			return nil, status.Error(codes.NotFound, "User not found")
		}

		return nil, status.Error(codes.Internal, "Failed to fetch user profile")
	}

	return &pb.GetUserProfileResponse{
        Id:                user.ID.String(),
        Username:          user.Username,
        Name:              user.Name,
        Bio:               user.Bio,
        ProfilePictureUrl: user.ProfilePictureURL,
        FollowersCount:    followers,
        FollowingCount:    following,
        IsFollowing:       false, 
    }, nil
}

func (h *UserHandler) FollowUser(ctx context.Context, req *pb.FollowUserRequest) (*pb.FollowUserResponse, error) {
    if req.FollowerId == req.FollowingId {
        return nil, status.Error(codes.InvalidArgument, "You cannot follow yourself")
    }

    exists, _ := h.repo.IsFollowing(req.FollowerId, req.FollowingId)
    if exists {
        return &pb.FollowUserResponse{Message: "Already following"}, nil
    }

    err := h.repo.CreateFollow(req.FollowerId, req.FollowingId)
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to follow user")
    }

    return &pb.FollowUserResponse{Message: "Successfully followed user"}, nil
}

func (h *UserHandler) UnfollowUser(ctx context.Context, req *pb.UnfollowUserRequest) (*pb.UnfollowUserResponse, error) {
    err := h.repo.DeleteFollow(req.FollowerId, req.FollowingId)
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to unfollow user")
    }
    return &pb.UnfollowUserResponse{Message: "Successfully unfollowed user"}, nil
}
