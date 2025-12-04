package repositories

import (
	"context"
	"time"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/core/domain"
	"gorm.io/gorm"
)

type GormStoryRepository struct {
	db *gorm.DB
}

func NewGormStoryRepository(db *gorm.DB) *GormStoryRepository {
	return &GormStoryRepository{db: db}
}

func (r *GormStoryRepository) CreateStory(ctx context.Context, story *domain.Story) error {
	return r.db.WithContext(ctx).Create(story).Error
}

func (r *GormStoryRepository) GetStoryByID(ctx context.Context, storyID string) (*domain.Story, error) {
	var story domain.Story
	err := r.db.WithContext(ctx).
		Where("id = ? AND expires_at > ?", storyID, time.Now()).
		First(&story).Error

	if err != nil {
		return nil, err
	}
	return &story, nil
}

func (r *GormStoryRepository) GetUserStories(ctx context.Context, userID string, includeExpired bool) ([]*domain.Story, error) {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	// Only filter by expiration if we are NOT looking at the archive
	if !includeExpired {
		query = query.Where("expires_at > ?", time.Now())
	}

	var stories []*domain.Story
	err := query.Order("created_at DESC").Find(&stories).Error
	return stories, err
}

func (r *GormStoryRepository) GetFollowingStories(ctx context.Context, userIDs []string, viewerID string,limit int) (map[string][]*domain.Story, error) {
	if len(userIDs) == 0 {
		return make(map[string][]*domain.Story), nil
	}

	var stories []*domain.Story
	err := r.db.WithContext(ctx).
		Where("user_id IN ? AND expires_at > ?", userIDs, time.Now()).
		Where("user_id NOT IN (?)", r.db.Table("story_visibilities").Select("user_id").Where("hidden_viewer_id = ?", viewerID)).
		Order("created_at DESC").
		Limit(limit).
		Find(&stories).Error

	if err != nil {
		return nil, err
	}

	storiesByUser := make(map[string][]*domain.Story)
	for _, story := range stories {
		storiesByUser[story.UserID] = append(storiesByUser[story.UserID], story)
	}

	return storiesByUser, nil
}

func (r *GormStoryRepository) DeleteStory(ctx context.Context, storyID, userID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", storyID, userID).
		Delete(&domain.Story{}).Error
}

func (r *GormStoryRepository) DeleteExpiredStories(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&domain.Story{}).Error
}

func (r *GormStoryRepository) ViewStory(ctx context.Context, view *domain.StoryView) error {
	return r.db.WithContext(ctx).
		Where("story_id = ? AND user_id = ?", view.StoryID, view.UserID).
		Assign(domain.StoryView{
			StoryID:  view.StoryID,
			UserID:   view.UserID,
			ViewedAt: time.Now(),
		}).
		FirstOrCreate(&view).Error
}

func (r *GormStoryRepository) IsStoryViewed(ctx context.Context, storyID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.StoryView{}).
		Where("story_id = ? AND user_id = ?", storyID, userID).
		Count(&count).Error

	return count > 0, err
}

func (r *GormStoryRepository) GetStoryViewers(ctx context.Context, storyID string) ([]*domain.StoryView, error) {
	var viewers []*domain.StoryView
	err := r.db.WithContext(ctx).
		Where("story_id = ?", storyID).
		Order("viewed_at DESC").
		Find(&viewers).Error

	return viewers, err
}

func (r *GormStoryRepository) LikeStory(ctx context.Context, like *domain.StoryLike) error {
	return r.db.WithContext(ctx).Create(like).Error
}

func (r *GormStoryRepository) UnlikeStory(ctx context.Context, storyID, userID string) error {
	return r.db.WithContext(ctx).
		Where("story_id = ? AND user_id = ?", storyID, userID).
		Delete(&domain.StoryLike{}).Error
}

func (r *GormStoryRepository) IsStoryLiked(ctx context.Context, storyID, userID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.StoryLike{}).
		Where("story_id = ? AND user_id = ?", storyID, userID).
		Count(&count).Error

	return count > 0, err
}

func (r *GormStoryRepository) CreateReply(ctx context.Context, reply *domain.StoryReply) error {
	return r.db.WithContext(ctx).Create(reply).Error
}

func (r *GormStoryRepository) GetStoryReplies(ctx context.Context, storyID string) ([]*domain.StoryReply, error) {
	var replies []*domain.StoryReply
	err := r.db.WithContext(ctx).
		Where("story_id = ?", storyID).
		Order("created_at DESC").
		Find(&replies).Error

	return replies, err
}

func (r *GormStoryRepository) ShareStory(ctx context.Context, share *domain.StoryShare) error {
	return r.db.WithContext(ctx).Create(share).Error
}
