package ports

import (
	"context"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/domain"
)


type PostRepository interface {
	CreatePost(ctx context.Context, post *domain.Post) error
	GetPostByID(ctx context.Context, postID string) (*domain.Post, error)
	GetPostsByUserID(ctx context.Context, userID string) ([]*domain.Post, error)

	LikePost(ctx context.Context, like *domain.PostLike) error
	UnlikePost(ctx context.Context, userID, postID string) error

	CreateComment(ctx context.Context, comment *domain.PostComment) error
    GetCommentsForPost(ctx context.Context, postID string) ([]*domain.PostComment, error)
	GetFeedPosts(ctx context.Context, userIDs []string, currentUserID string, limit, offset int) ([]*domain.Post, error)

	CreateCollection(ctx context.Context, collection *domain.Collection) error
    GetUserCollections(ctx context.Context, userID string) ([]*domain.Collection, error)
    
    ToggleSavePost(ctx context.Context, userID, postID, collectionID string) (bool, error)

	CreatePostWithMentions(ctx context.Context, post *domain.Post, mentions []domain.UserMention) error
	GetPostsByMention(ctx context.Context, targetUserID string, limit, offset int) ([]domain.Post, error)
}