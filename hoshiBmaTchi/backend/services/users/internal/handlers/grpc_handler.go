package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
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

type NotificationEvent struct {
	RecipientID string `json:"recipient_id"`
	SenderID    string `json:"sender_id"`
	SenderName  string `json:"sender_name"`
	SenderImage string `json:"sender_image"`
	Type        string `json:"type"`
	EntityID    string `json:"entity_id"`
	Message     string `json:"message"`
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

	accessToken, refreshToken, err := utils.GenerateTokens(user.ID.String(), user.Email, user.Role)
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

	accessToken, refreshToken, err := utils.GenerateTokens(user.ID.String(), user.Email, user.Role)
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
	user, err := h.repo.FindByEmailOrUsername(req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.InvalidArgument, "Invalid or expired 2FA code")
		}
		return nil, status.Error(codes.Internal, "failed to find user")
	}

	otpKey := "2fa:" + user.Email
	
	storedOtp, err := h.redis.Get(ctx, otpKey).Result()
	if err == redis.Nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid or expired 2FA code")
	} else if err != nil {
		return nil, status.Error(codes.Internal, "failed to get 2FA code")
	}

	if storedOtp != req.OtpCode {
		return nil, status.Error(codes.InvalidArgument, "Invalid or expired 2FA code")
	}

	h.redis.Del(ctx, otpKey)

	accessToken, refreshToken, err := utils.GenerateTokens(user.ID.String(), user.Email, user.Role)
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

	role, _ := claims["role"].(string)
	return &pb.ValidateTokenResponse{
		Valid:  true,
		UserId: userID,
		Role: role,
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

	isFollowing := false
    if req.ViewerId != "" && req.ViewerId != req.UserId {
        status, err := h.repo.IsFollowing(req.ViewerId, req.UserId)
        if err == nil {
            isFollowing = status
        }
    }

	if req.ViewerId != "" && req.ViewerId != req.UserId {
        isBlocked, err := h.repo.IsBlocked(req.ViewerId, req.UserId)
        if err == nil && isBlocked {
            return nil, status.Error(codes.NotFound, "User not found")
        }
    }

	return &pb.GetUserProfileResponse{
        Id:                user.ID.String(),
        Username:          user.Username,
        Name:              user.Name,
        Bio:               user.Bio,
        ProfilePictureUrl: user.ProfilePictureURL,
        FollowersCount:    followers,
        FollowingCount:    following,
		IsFollowing:       isFollowing,
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

	go func() {
        follower, _, _, err := h.repo.GetUserProfileWithStats(req.FollowerId)
        if err != nil { return }

        event := NotificationEvent{
            RecipientID: req.FollowingId,       
            SenderID:    req.FollowerId,        
            SenderName:  follower.Username,
            SenderImage: follower.ProfilePictureURL,
            Type:        "follow",
            EntityID:    req.FollowerId,        
            Message:     "started following you",
        }

        body, _ := json.Marshal(event)
        h.amqpChan.PublishWithContext(context.Background(),
            "notification_exchange", 
            "notification.event",    
            false, false,
            amqp.Publishing{
                ContentType: "application/json",
                Body:        body,
            },
        )
    }()

    return &pb.FollowUserResponse{Message: "Successfully followed user"}, nil
}

func (h *UserHandler) UnfollowUser(ctx context.Context, req *pb.UnfollowUserRequest) (*pb.UnfollowUserResponse, error) {
    err := h.repo.DeleteFollow(req.FollowerId, req.FollowingId)
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to unfollow user")
    }
    return &pb.UnfollowUserResponse{Message: "Successfully unfollowed user"}, nil
}

func (h *UserHandler) GetFollowingList (ctx context.Context, req *pb.GetFollowingListRequest) (*pb.GetFollowingListResponse, error){
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	followingIDs, err := h.repo.GetFollowing(req.UserId)

	if err != nil {
		log.Printf("Failed to fetch following list for user %s: %v", req.UserId, err)
        return nil, status.Error(codes.Internal, "Failed to fetch following list")
	}

	return &pb.GetFollowingListResponse{
		FollowingIds: followingIDs,
	}, nil
}

func min(a, b int) int { 
	if a < b { 
		return a 
	}; 
	return b 
}

func max(a, b int) int { 
	if a > b { 
		return a 
	}; 

	return b 
}

func jaroWinkler(s1, s2 string) float64{
	s1, s2 = strings.ToLower(s1), strings.ToLower(s2)
	if s1 == s2 {
		return 1.0
	}

	matchDistance := int(math.Floor(float64(max(len(s1), len(s2)))/2.0) - 1)
	matches := 0
	transpositions := 0

	s1Matches := make([]bool, len(s1))
	s2Matches := make([]bool, len(s2))

	for i := 0; i < len(s1); i++{
		start := max(0, i - matchDistance)
		end := min(i + matchDistance + 1, len(s2))

		for j := start; j < end; j++ {
			if s2Matches[j] { 
				continue 
			}

            if s1[i] != s2[j] { 
				continue 
			}

            s1Matches[i] = true
            s2Matches[j] = true
            matches++
            break
		}
	}

	if matches == 0 { 
		return 0.0 
	}
    
    k := 0
    for i := 0; i < len(s1); i++ {
        if !s1Matches[i] { 
			continue 
		}

        for !s2Matches[k] { k++ }
        if s1[i] != s2[k] { transpositions++ }
        k++
    }
    
    jaro := (float64(matches)/float64(len(s1)) + 
             float64(matches)/float64(len(s2)) + 
             float64(matches-transpositions/2)/float64(matches)) / 3.0

    prefix := 0
    for i := 0; i < min(len(s1), len(s2)); i++ {
        if s1[i] == s2[i] { prefix++ } else { break }
    }
    prefix = min(prefix, 4)
    
    return jaro + 0.1*float64(prefix)*(1.0-jaro)
}

func (h *UserHandler) SearchUsers(ctx context.Context, req *pb.SearchUsersRequest) (*pb.SearchUsersResponse, error){
	candidates, err := h.repo.SearchUsers(ctx, req.Query, req.UserId)
	if err != nil {
		return nil, err
	}

	type match struct {
		User *domain.User
		Score float64
	}

	var matches []match

	for _, u := range candidates {
		score := jaroWinkler(req.Query, u.Username)
		nameScore := jaroWinkler(req.Query, u.Name)

		if nameScore > score {
			score = nameScore
		}

		matches = append(matches, match{User: u, Score: score})
	}

    sort.Slice(matches, func(i, j int) bool {
        return matches[i].Score > matches[j].Score
    })

    limit := 5
    if len(matches) < 5 {
        limit = len(matches)
    }

    var responseUsers []*pb.UserProfile
    for i := 0; i < limit; i++ {
        u := matches[i].User
        responseUsers = append(responseUsers, &pb.UserProfile{
            UserId:            u.ID.String(),
            Username:          u.Username,
            Name:              u.Name,
            ProfilePictureUrl: u.ProfilePictureURL,
        })
    }

    return &pb.SearchUsersResponse{Users: responseUsers}, nil
}

func (h *UserHandler) GetUserByUsername(ctx context.Context, req *pb.GetUserByUsernameRequest) (*pb.GetUserProfileResponse, error) {
    if req.Username == "" {
        return nil, status.Error(codes.InvalidArgument, "Username is required")
    }

    user, err := h.repo.FindByEmailOrUsername(req.Username)
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, status.Error(codes.NotFound, "User not found")
        }
        return nil, status.Error(codes.Internal, "Failed to find user")
    }

    return h.GetUserProfile(ctx, &pb.GetUserProfileRequest{
        UserId: user.ID.String(),
    })
}

func (h *UserHandler) GetSuggestedUsers(ctx context.Context, req *pb.GetSuggestedUsersRequest) (*pb.GetSuggestedUsersResponse, error) {
    if req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "User ID is required")
    }

    users, err := h.repo.GetSuggestedUsers(ctx, req.UserId)
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to fetch suggested users")
    }

    var responseUsers []*pb.UserProfile
    for _, u := range users {
        responseUsers = append(responseUsers, &pb.UserProfile{
            UserId:            u.ID.String(),
            Username:          u.Username,
            Name:              u.Name,
            ProfilePictureUrl: u.ProfilePictureURL,
            IsFollowing:       false, // By definition, these are not followed
        })
    }

    return &pb.GetSuggestedUsersResponse{
        Users: responseUsers,
    }, nil
}

func (h *UserHandler) GetFollowingProfiles(ctx context.Context, req *pb.GetFollowingListRequest) (*pb.GetFollowingProfilesResponse, error) {
    if req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "User ID is required")
    }

    users, err := h.repo.GetFollowingUsers(req.UserId)
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to fetch following profiles")
    }

    var responseUsers []*pb.UserProfile
    for _, u := range users {
        responseUsers = append(responseUsers, &pb.UserProfile{
            UserId:            u.ID.String(),
            Username:          u.Username,
            Name:              u.Name,
            ProfilePictureUrl: u.ProfilePictureURL,
            IsFollowing:       true,
        })
    }

    return &pb.GetFollowingProfilesResponse{
        Users: responseUsers,
    }, nil
}

func (h *UserHandler) BlockUser(ctx context.Context, req *pb.BlockUserRequest) (*pb.BlockUserResponse, error) {
    if req.BlockerId == req.BlockedId {
        return nil, status.Error(codes.InvalidArgument, "Cannot block yourself")
    }

    err := h.repo.CreateBlock(req.BlockerId, req.BlockedId)
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to block user")
    }

    return &pb.BlockUserResponse{Message: "User blocked successfully"}, nil
}

func (h *UserHandler) UnblockUser(ctx context.Context, req *pb.UnblockUserRequest) (*pb.UnblockUserResponse, error) {
    err := h.repo.DeleteBlock(req.BlockerId, req.BlockedId)
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to unblock user")
    }

    return &pb.UnblockUserResponse{Message: "User unblocked successfully"}, nil
}

func (h *UserHandler) GetBlockedList(ctx context.Context, req *pb.GetBlockedListRequest) (*pb.GetBlockedListResponse, error) {
    if req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "User ID required")
    }

    users, err := h.repo.GetBlockedUsers(req.UserId)
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to fetch blocked users")
    }

    var responseUsers []*pb.UserProfile
    for _, u := range users {
        responseUsers = append(responseUsers, &pb.UserProfile{
            UserId:            u.ID.String(),
            Username:          u.Username,
            Name:              u.Name,
            ProfilePictureUrl: u.ProfilePictureURL,
        })
    }

    return &pb.GetBlockedListResponse{Users: responseUsers}, nil
}

func (h *UserHandler) GetSettings(ctx context.Context, req *pb.GetSettingsRequest) (*pb.GetSettingsResponse, error) {
    user, err := h.repo.FindByID(req.UserId)
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to fetch user settings")
    }

    return &pb.GetSettingsResponse{
        IsPrivate:   user.IsPrivate,            // Ensure this exists in domain.User
        EnablePush:  user.PushNotificationsEnabled,    // Ensure this exists in domain.User
        EnableEmail: user.SubscribedToNewsletter, 
    }, nil
}

func (h *UserHandler) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileRequest) (*pb.UpdateUserProfileResponse, error) {
    user, err := h.repo.FindByID(req.UserId)
    if err != nil {
        return nil, status.Error(codes.NotFound, "User not found")
    }

    // Update fields
    user.Name = req.Name
    user.Bio = req.Bio
    user.Gender = req.Gender
    
    // Only update picture if a new URL is provided
    if req.ProfilePictureUrl != "" {
        user.ProfilePictureURL = req.ProfilePictureUrl
    }

    err = h.repo.UpdateUser(user)
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to update profile")
    }

    return &pb.UpdateUserProfileResponse{
        User: &pb.UserProfile{
            UserId:            user.ID.String(),
            Username:          user.Username,
            Name:              user.Name,
            ProfilePictureUrl: user.ProfilePictureURL,
        },
    }, nil
}

func (h *UserHandler) UpdateNotificationSettings(ctx context.Context, req *pb.UpdateNotificationSettingsRequest) (*pb.UpdateNotificationSettingsResponse, error) {
    user, err := h.repo.FindByID(req.UserId)
    if err != nil {
        return nil, status.Error(codes.NotFound, "User not found")
    }

    user.PushNotificationsEnabled = req.EnablePush
    user.SubscribedToNewsletter = req.EnableEmail

    err = h.repo.UpdateUser(user)
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to update notification settings")
    }

    return &pb.UpdateNotificationSettingsResponse{Success: true}, nil
}

func (h *UserHandler) UpdatePrivacySettings(ctx context.Context, req *pb.UpdatePrivacySettingsRequest) (*pb.UpdatePrivacySettingsResponse, error) {
    user, err := h.repo.FindByID(req.UserId)
    if err != nil {
        return nil, status.Error(codes.NotFound, "User not found")
    }

    user.IsPrivate = req.IsPrivate

    err = h.repo.UpdateUser(user)
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to update privacy settings")
    }

    return &pb.UpdatePrivacySettingsResponse{Success: true}, nil
}

func (h *UserHandler) GetCloseFriends(ctx context.Context, req *pb.GetListRequest) (*pb.GetListResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID required")
	}

    // Call the repo method (already exists in your gorm_repository.go)
	users, err := h.repo.GetCloseFriends(utils.ParseUUID(req.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to fetch close friends")
	}

	var responseUsers []*pb.UserProfile
	for _, u := range users {
		responseUsers = append(responseUsers, &pb.UserProfile{
			UserId:            u.ID.String(),
			Username:          u.Username,
			Name:              u.Name,
			ProfilePictureUrl: u.ProfilePictureURL,
		})
	}

	return &pb.GetListResponse{Users: responseUsers}, nil
}

func (h *UserHandler) AddCloseFriend(ctx context.Context, req *pb.ManageRelationRequest) (*pb.ManageRelationResponse, error) {
	err := h.repo.AddCloseFriend(utils.ParseUUID(req.UserId), utils.ParseUUID(req.TargetUserId))
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to add close friend")
	}
	return &pb.ManageRelationResponse{Success: true}, nil
}

func (h *UserHandler) RemoveCloseFriend(ctx context.Context, req *pb.ManageRelationRequest) (*pb.ManageRelationResponse, error) {
	err := h.repo.RemoveCloseFriend(utils.ParseUUID(req.UserId), utils.ParseUUID(req.TargetUserId))
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to remove close friend")
	}
	return &pb.ManageRelationResponse{Success: true}, nil
}

func (h *UserHandler) GetHiddenStoryUsers(ctx context.Context, req *pb.GetListRequest) (*pb.GetListResponse, error) {
	users, err := h.repo.GetHiddenStoryUsers(utils.ParseUUID(req.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to fetch hidden story users")
	}

	var responseUsers []*pb.UserProfile
	for _, u := range users {
		responseUsers = append(responseUsers, &pb.UserProfile{
			UserId:            u.ID.String(),
			Username:          u.Username,
			Name:              u.Name,
			ProfilePictureUrl: u.ProfilePictureURL,
		})
	}

	return &pb.GetListResponse{Users: responseUsers}, nil
}

func (h *UserHandler) HideStoryFromUser(ctx context.Context, req *pb.ManageRelationRequest) (*pb.ManageRelationResponse, error) {
	err := h.repo.HideStoryFromUser(utils.ParseUUID(req.UserId), utils.ParseUUID(req.TargetUserId))
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to hide story from user")
	}
	return &pb.ManageRelationResponse{Success: true}, nil
}

func (h *UserHandler) UnhideStoryFromUser(ctx context.Context, req *pb.ManageRelationRequest) (*pb.ManageRelationResponse, error) {
	err := h.repo.UnhideStoryFromUser(utils.ParseUUID(req.UserId), utils.ParseUUID(req.TargetUserId))
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to unhide story from user")
	}
	return &pb.ManageRelationResponse{Success: true}, nil
}

// --- Verification ---

func (h *UserHandler) RequestVerification(ctx context.Context, req *pb.RequestVerificationRequest) (*pb.RequestVerificationResponse, error) {
	verificationReq := &domain.VerificationRequest{
		UserID:           utils.ParseUUID(req.UserId),
		NationalIDNumber: req.NationalIdNumber,
		Reason:           req.Reason,
		SelfieURL:        req.SelfieUrl,
        Status:           "PENDING", // Default status
	}

	err := h.repo.CreateVerificationRequest(verificationReq)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to submit verification request")
	}

	return &pb.RequestVerificationResponse{Message: "Verification request submitted"}, nil
}

func (h *UserHandler) GetAllUsers(ctx context.Context, req *pb.Empty) (*pb.UserListResponse, error) {
    users, err := h.repo.GetAllUsers()
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to fetch users")
    }

    var pbUsers []*pb.UserProfile
    for _, u := range users {
        pbUsers = append(pbUsers, &pb.UserProfile{
            UserId:            u.ID.String(),
            Username:          u.Username,
            Name:              u.Name,
            Email:             u.Email, 
            IsBanned:          u.IsBanned, 
            ProfilePictureUrl: u.ProfilePictureURL,
        })
    }
    return &pb.UserListResponse{Users: pbUsers}, nil
}

func (h *UserHandler) ToggleUserBan(ctx context.Context, req *pb.ToggleUserBanRequest) (*pb.Response, error) {
	if req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "User ID is required")
    }

    err := h.repo.UpdateUserBanStatus(req.UserId, req.IsBanned)
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to update ban status")
    }
    
    // Optional: Send Email Notification about Ban/Unban
    
    return &pb.Response{Success: true, Message: "User status updated"}, nil
}

func (h *UserHandler) GetSubscribedEmails(ctx context.Context, req *pb.Empty) (*pb.EmailListResponse, error) {
    emails, err := h.repo.GetSubscribedEmails()
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to fetch emails")
    }
    return &pb.EmailListResponse{Emails: emails}, nil
}

func (h *UserHandler) GetVerificationRequests(ctx context.Context, req *pb.Empty) (*pb.VerificationListResponse, error) {
    reqs, err := h.repo.GetPendingVerificationRequests()
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to fetch requests")
    }

    var pbReqs []*pb.VerificationRequestItem
    for _, r := range reqs {
        // You might need to fetch User details here if not preloaded
        // user, _ := h.repo.FindByID(r.UserID.String())
        
        pbReqs = append(pbReqs, &pb.VerificationRequestItem{
            Id:               r.ID.String(),
            UserId:           r.UserID.String(),
			Username:          r.User.Username,          
            ProfilePictureUrl: r.User.ProfilePictureURL,
            NationalIdNumber: r.NationalIDNumber,
            Reason:           r.Reason,
            SelfieUrl:        r.SelfieURL,
            Status:           r.Status,
            CreatedAt:        r.CreatedAt.Format(time.RFC3339),
        })
    }
    return &pb.VerificationListResponse{Requests: pbReqs}, nil
}

func (h *UserHandler) ReviewVerification(ctx context.Context, req *pb.ReviewVerificationRequest) (*pb.Response, error) {
    // 1. Update Request Status
    verificationReq, err := h.repo.UpdateVerificationStatus(req.RequestId, req.Action)
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to update status")
    }

    // 2. If Accepted, Verify User and Send Email
    if req.Action == "ACCEPTED" {
        h.repo.VerifyUser(verificationReq.UserID)

        // Fetch User to get Email
        user, _ := h.repo.FindByID(verificationReq.UserID.String())
        
        // Send Email
        emailBody := "Congratulations! Your account has been verified."
        task := EmailTask{Email: user.Email, Subject: "Verification Approved", Body: emailBody}
        taskBody, _ := json.Marshal(task)
        
        h.amqpChan.PublishWithContext(ctx, "email_exchange", "send_email", false, false, 
            amqp.Publishing{ContentType: "application/json", Body: taskBody})
    }

    return &pb.Response{Success: true, Message: "Review processed"}, nil
}

func (h *UserHandler) GetUserReports(ctx context.Context, req *pb.Empty) (*pb.UserReportListResponse, error) {
    reports, err := h.repo.GetPendingUserReports()
    if err != nil { return nil, err }
    
    var pbReports []*pb.UserReportItem
    for _, r := range reports {
        pbReports = append(pbReports, &pb.UserReportItem{
            Id:             r.ID.String(),
            ReporterId:     r.ReporterID.String(),
            ReportedUserId: r.ReportedUserID.String(),
            Reason:         r.Reason,
            Status:         r.Status,
        })
    }
    return &pb.UserReportListResponse{Reports: pbReports}, nil
}

func (h *UserHandler) ReviewUserReport(ctx context.Context, req *pb.ReviewReportRequest) (*pb.Response, error) {
    report, err := h.repo.FindUserReportByID(req.ReportId)
    if err != nil {
        return nil, status.Error(codes.NotFound, "Report not found")
    }

    statusStr := "REJECTED"
    
    if req.Action == "BAN_USER" {
        statusStr = "RESOLVED"
        
        if err := h.repo.UpdateUserBanStatus(report.ReportedUserID.String(), true); err != nil {
            return nil, status.Error(codes.Internal, "Failed to ban user")
        }

        reporter, err := h.repo.FindByID(report.ReporterID.String())
        if err == nil {
            emailBody := fmt.Sprintf("We have reviewed your report against user %s and have taken action.", report.ReportedUserID)
            task := EmailTask{Email: reporter.Email, Subject: "Report Update", Body: emailBody}
            
            taskBody, _ := json.Marshal(task)
            
            h.amqpChan.PublishWithContext(ctx, 
                "email_exchange", 
                "send_email", 
                false, 
                false, 
                amqp.Publishing{
                    ContentType: "application/json", 
                    Body: taskBody, 
                })
        }
    }

    if err := h.repo.UpdateUserReportStatus(req.ReportId, statusStr); err != nil {
         return nil, status.Error(codes.Internal, "Failed to update report status")
    }

    return &pb.Response{Success: true}, nil
}

func (h *UserHandler) ReportUser(ctx context.Context, req *pb.ReportUserRequest) (*pb.Response, error) {
    if req.ReporterId == req.ReportedUserId {
        return nil, status.Error(codes.InvalidArgument, "Cannot report yourself")
    }

    report := &domain.UserReport{
        ReporterID:     utils.ParseUUID(req.ReporterId),
        ReportedUserID: utils.ParseUUID(req.ReportedUserId),
        Reason:         req.Reason,
        Status:         "PENDING",
    }

    if err := h.repo.CreateUserReport(report); err != nil {
        return nil, status.Error(codes.Internal, "Failed to submit report")
    }

    return &pb.Response{Success: true, Message: "User reported successfully"}, nil
}

// Implement GetUserEmail (Internal)
func (h *UserHandler) GetUserEmail(ctx context.Context, req *pb.GetUserEmailRequest) (*pb.GetUserEmailResponse, error) {
    user, err := h.repo.FindByID(req.UserId)
    if err != nil {
        return nil, status.Error(codes.NotFound, "User not found")
    }
    return &pb.GetUserEmailResponse{Email: user.Email}, nil
}
