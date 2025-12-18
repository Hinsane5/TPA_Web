package ports

import (
	"context"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/domain"
	"github.com/google/uuid"
)

type UserRepository interface{
	Save(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindByEmailOrUsername(identifier string) (*domain.User, error)
	FindByID(userID string) (*domain.User, error)
	UpdatePassword(userID string, newPassword string) error
	GetUserProfileWithStats(userID string) (*domain.User, int64, int64, error)
	CreateFollow(followerID, followingID string) error
    DeleteFollow(followerID, followingID string) error
    IsFollowing(followerID, followingID string) (bool, error)
	GetFollowing(userID string) ([]string, error)
	SearchUsers(ctx context.Context, query string, userID string) ([]*domain.User, error)
	GetSuggestedUsers(ctx context.Context, userID string) ([]*domain.User, error)
	GetFollowingUsers(userID string) ([]*domain.User, error)
	CreateBlock(blockerID, blockedID string) error
    DeleteBlock(blockerID, blockedID string) error
    GetBlockedUsers(userID string) ([]*domain.User, error)
    IsBlocked(userA, userB string) (bool, error)
	UpdateUser(user *domain.User) error
	
	AddCloseFriend(userID, targetID uuid.UUID) error
	RemoveCloseFriend(userID, targetID uuid.UUID) error
	GetCloseFriends(userID uuid.UUID) ([]domain.User, error)
	
	HideStoryFromUser(userID, targetID uuid.UUID) error
	UnhideStoryFromUser(userID, targetID uuid.UUID) error
	GetHiddenStoryUsers(userID uuid.UUID) ([]domain.User, error)

	CreateVerificationRequest(req *domain.VerificationRequest) error
	GetSubscribedEmails() ([]string, error) 
	GetPendingVerificationRequests() ([]*domain.VerificationRequest, error)
	UpdateVerificationStatus(requestID string, status string) (*domain.VerificationRequest, error)
	VerifyUser(userID uuid.UUID) error
	GetPendingUserReports() ([]*domain.UserReport, error)
	UpdateUserReportStatus(reportID string, status string) error
	UpdateUserBanStatus(userID string, isBanned bool) error
	GetAllUsers() ([]*domain.User, error)

	FindUserReportByID(reportID string) (*domain.UserReport, error)
	CreateUserReport(report *domain.UserReport) error
}