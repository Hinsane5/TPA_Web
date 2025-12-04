package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/core/domain"
	"github.com/redis/go-redis/v9"
)

type RedisStoryRepository struct {
	client *redis.Client
}

func NewRedisStoryRepository(client *redis.Client) *RedisStoryRepository{
	return &RedisStoryRepository{client: client}
}

func (r *RedisStoryRepository) SetUserFeed(ctx context.Context, userID string, stories []*domain.Story) error {
	data, err := json.Marshal(stories)
	if err != nil {
		return err
	}
	// Expire cache after 5 minutes (or invalidate on new story creation)
	return r.client.Set(ctx, fmt.Sprintf("feed:%s", userID), data, 5*time.Minute).Err()
}

func (r *RedisStoryRepository) GetUserFeed(ctx context.Context, userID string) ([]*domain.Story, error) {
	val, err := r.client.Get(ctx, fmt.Sprintf("feed:%s", userID)).Result()
	if err != nil {
		return nil, err
	}
	var stories []*domain.Story
	err = json.Unmarshal([]byte(val), &stories)
	return stories, err
}