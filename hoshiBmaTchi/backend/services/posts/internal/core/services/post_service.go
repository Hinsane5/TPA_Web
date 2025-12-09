package services

import (
	"context"
	"regexp"

	"github.com/google/uuid"
	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/posts"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/ports"
)

var mentionRegex = regexp.MustCompile(`@([a-zA-Z0-9._]+)`)

type PostService struct {
	repo ports.PostRepository
}

func NewPostService(repo ports.PostRepository) *PostService {
	return &PostService{
		repo: repo,
	}
}

func (s *PostService) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*domain.Post, error) {
	userID, _ := uuid.Parse(req.UserId)

	post := &domain.Post{
		UserID:   userID,
		Caption:  req.Caption,
		Location: req.Location,
	}


	uniqueUsernames := make(map[string]bool)
	matches := mentionRegex.FindAllStringSubmatch(req.Caption, -1)
	for _, match := range matches {
		if len(match) > 1 {
			uniqueUsernames[match[1]] = true 
		}
	}

	var mentions []domain.UserMention
	
	err := s.repo.CreatePostWithMentions(ctx, post, mentions)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostService) GetUserMentions(ctx context.Context, req *pb.GetUserMentionsRequest) ([]domain.Post, error) {

    return s.repo.GetPostsByMention(ctx, req.TargetUserId, int(req.Limit), int(req.Offset))
}

func (s *PostService) GetReelsFeed(ctx context.Context, limit, offset int) ([]*domain.Post, error) {
    return s.repo.GetReels(ctx, limit, offset)
}

func (s *PostService) GetExplorePosts(ctx context.Context, limit, offset int, hashtag string) ([]*domain.Post, error) {
    return s.repo.GetExplorePosts(ctx, limit, offset, hashtag)
}