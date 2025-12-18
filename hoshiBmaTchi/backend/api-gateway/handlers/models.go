package handlers

// Swagger response models
// These are used for documentation only, actual responses use protobuf types

// LoginResponse represents the login response
// @Description Login response with user token
type LoginResponse struct {
    Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
    User  UserInfo `json:"user"`
}

// UserInfo represents user information
type UserInfo struct {
    ID       string `json:"id" example:"123"`
    Username string `json:"username" example:"john_doe"`
    Email    string `json:"email" example:"john@example.com"`
    FullName string `json:"full_name" example:"John Doe"`
}

// RegisterResponse represents the registration response
type RegisterResponse struct {
    Message string `json:"message" example:"User registered successfully"`
    UserID  string `json:"user_id" example:"123"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
    Error   string `json:"error" example:"Invalid credentials"`
    Message string `json:"message,omitempty" example:"Please check your username and password"`
}

// PostResponse represents a single post
type PostResponse struct {
    ID          string   `json:"id" example:"post123"`
    UserID      string   `json:"user_id" example:"user123"`
    Username    string   `json:"username" example:"john_doe"`
    Caption     string   `json:"caption" example:"Beautiful sunset"`
    MediaURLs   []string `json:"media_urls"`
    LikesCount  int32    `json:"likes_count" example:"42"`
    IsLiked     bool     `json:"is_liked" example:"false"`
    CreatedAt   string   `json:"created_at" example:"2025-12-18T10:30:00Z"`
}

// PostsListResponse represents a list of posts
type PostsListResponse struct {
    Posts []PostResponse `json:"posts"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
    Message string `json:"message" example:"Operation successful"`
}