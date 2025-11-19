package repositories

import (
	"context"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/domain"
	"gorm.io/gorm"
)

type GormPostRepository struct {
	db *gorm.DB
}

func NewGormPostRepository(db *gorm.DB) *GormPostRepository{
	return &GormPostRepository{db: db}
}

func (r *GormPostRepository) CreatePost(ctx context.Context, post *domain.Post) error {
	
	result := r.db.WithContext(ctx).Create(post)
    return result.Error 
}

func (r *GormPostRepository) GetPostByID(ctx context.Context, postID string) (*domain.Post, error) {
	return nil, nil 
}

func (r *GormPostRepository) GetPostsByUserID(ctx context.Context, userID string) ([]*domain.Post, error) {
	var posts []*domain.Post

	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&posts).Error

	if err != nil {
		return nil, err
	}

	return posts, nil 
}

func (r *GormPostRepository) LikePost(ctx context.Context, like *domain.PostLike) error {
    result := r.db.WithContext(ctx).Create(like)
    return result.Error
}

func (r *GormPostRepository) UnlikePost(ctx context.Context, userID, postID string) error {
    result := r.db.WithContext(ctx).
        Where("user_id = ? AND post_id = ?", userID, postID).
        Delete(&domain.PostLike{})

    return result.Error
}

func (r *GormPostRepository) CreateComment(ctx context.Context, comment *domain.PostComment) error {
	result := r.db.WithContext(ctx).Create(comment)
	return result.Error
}

func (r *GormPostRepository) GetCommentsForPost(ctx context.Context, postID string) ([]*domain.PostComment, error){
	var comments []*domain.PostComment
	err := r.db.WithContext(ctx).Where("post_id = ?", postID).Order("created_at asc").Find(&comments).Error
	return comments, err
}

func (r *GormPostRepository) GetFeedPosts(ctx context.Context, userIDs []string, limit, offset int) ([]*domain.Post, error) {
	var posts []*domain.Post
	
	err := r.db.WithContext(ctx).
		Where("user_id IN ?", userIDs).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error

	if err != nil {
		return nil, err
	}

	return posts, nil
}
