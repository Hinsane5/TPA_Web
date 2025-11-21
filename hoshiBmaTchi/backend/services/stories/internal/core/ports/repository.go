package ports

import (
	"context"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/core/domain"
)

type StoryRepository interface {
    CreateStory(ctx context.Context, story *domain.Story) error
    GetStoryByID(ctx context.Context, storyID string) (*domain.Story, error)
    GetUserStories(ctx context.Context, userID string) ([]*domain.Story, error)
    GetFollowingStories(ctx context.Context, userIDs []string, limit int) (map[string][]*domain.Story, error)
    DeleteStory(ctx context.Context, storyID, userID string) error
    DeleteExpiredStories(ctx context.Context) error
    
    ViewStory(ctx context.Context, view *domain.StoryView) error
    IsStoryViewed(ctx context.Context, storyID, userID string) (bool, error)
    GetStoryViewers(ctx context.Context, storyID string) ([]*domain.StoryView, error)
    
    LikeStory(ctx context.Context, like *domain.StoryLike) error
    UnlikeStory(ctx context.Context, storyID, userID string) error
    IsStoryLiked(ctx context.Context, storyID, userID string) (bool, error)
    
    CreateReply(ctx context.Context, reply *domain.StoryReply) error
    GetStoryReplies(ctx context.Context, storyID string) ([]*domain.StoryReply, error)
    
    ShareStory(ctx context.Context, share *domain.StoryShare) error
}

type UserServiceClient interface {
    GetFollowing(ctx context.Context, userID string) ([]string, error)
}

type ChatServiceClient interface {
    SendMessage(ctx context.Context, senderID, recipientID, content, storyID string) (string, error)
}
