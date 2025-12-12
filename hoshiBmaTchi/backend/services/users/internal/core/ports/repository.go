package ports

import (
	"context"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/users/internal/core/domain"
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
}