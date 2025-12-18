package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/posts"
	userPb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/ports"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

var mentionRegex = regexp.MustCompile(`@([a-zA-Z0-9._]+)`)
var hashtagRegex = regexp.MustCompile(`#([a-zA-Z0-9_]+)`)

type PostService struct {
	repo       ports.PostRepository
	amqpChan   *amqp.Channel
	userClient userPb.UserServiceClient
}

type NotificationEvent struct {
	RecipientID string   `json:"recipient_id"`
	SenderID    string   `json:"sender_id"`
	SenderName  string `json:"sender_name"` 
	SenderImage string `json:"sender_image"` 
	Type        string `json:"type"`
	EntityID    string   `json:"entity_id"`
	Message     string `json:"message"`
}

func NewPostService(repo ports.PostRepository, amqpChan *amqp.Channel, userClient userPb.UserServiceClient) *PostService {
	return &PostService{
		repo:       repo,
		amqpChan:   amqpChan,
		userClient: userClient,
	}
}

func (s *PostService) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*domain.Post, error) {
	userID, _ := uuid.Parse(req.UserId)

	log.Printf("[DEBUG] ---------------- START CREATE POST ----------------")
	log.Printf("[DEBUG] Incoming Post from User: %s", req.UserId)
	log.Printf("[DEBUG] Caption: %s", req.Caption)

    var mediaItems []domain.PostMedia
    for i, item := range req.Media {
        mediaItems = append(mediaItems, domain.PostMedia{
            MediaObjectName: item.MediaObjectName,
            MediaType:       item.MediaType,
            Sequence:        i,
        })
    }

	post := &domain.Post{
		UserID:   userID,
		Caption:  req.Caption,
		Location: req.Location,
		IsReel:   req.IsReel,
        Media:    mediaItems,
	}

	uniqueUsernames := make(map[string]bool)
	matches := mentionRegex.FindAllStringSubmatch(req.Caption, -1)

	for _, match := range matches {
		if len(match) > 1 {
			uniqueUsernames[match[1]] = true
		}
	}

	var mentions []domain.UserMention

	for username := range uniqueUsernames {
		targetUser, err := s.userClient.GetUserByUsername(ctx, &userPb.GetUserByUsernameRequest{Username: username})
		if err == nil && targetUser != nil {
			mentions = append(mentions, domain.UserMention{
				MentionedUserID: uuid.MustParse(targetUser.Id),
				CreatedByUserID: userID,
			})
		}
	}

	uniqueTags := make(map[string]bool)
	tagMatches := hashtagRegex.FindAllStringSubmatch(req.Caption, -1)

	for _, match := range tagMatches {
		if len(match) > 1 {
			tagName := strings.ToLower(match[1])
			uniqueTags[tagName] = true
		}
	}

	var hashtags []domain.Hashtag
	for tagName := range uniqueTags {
		hashtags = append(hashtags, domain.Hashtag{
			Name: tagName,
		})
	}

	post.Hashtags = hashtags

	err := s.repo.CreateFullPost(ctx, post, mentions)
	if err != nil {
		return nil, err
	}

	go func() {
		senderProfile, err := s.userClient.GetUserProfile(context.Background(), &userPb.GetUserProfileRequest{UserId: req.UserId})
		if err != nil {
			return
		}

		for _, mention := range mentions {
			event := NotificationEvent{
				RecipientID: mention.MentionedUserID.String(),
				SenderID:    req.UserId,
				SenderName:  senderProfile.Username,
				SenderImage: senderProfile.ProfilePictureUrl,
				Type:        "mention",
				EntityID:    post.ID.String(),
				Message:     "mentioned you in a post",
			}
			s.publishNotification(event)
		}
	}()

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

func (s *PostService) publishNotification(event NotificationEvent) {
	body, _ := json.Marshal(event)

	// [FIX] Use dynamic routing key based on event type (e.g., "notification.mention")
	routingKey := fmt.Sprintf("notification.%s", event.Type)

	log.Printf("[DEBUG] Publishing to RabbitMQ. Key: %s, Recipient: %s", routingKey, event.RecipientID)

	err := s.amqpChan.PublishWithContext(context.Background(),
		"notification_exchange",
		routingKey, // [FIX] Replaced hardcoded "notification.event"
		false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	
	if err != nil {
		log.Printf("Failed to publish notification: %v", err)
	} else {
		log.Printf("[DEBUG] Successfully published to 'notification_exchange'")
	}
}

func (s *PostService) LikePost(ctx context.Context, req *pb.LikePostRequest) error {
    userID, _ := uuid.Parse(req.UserId)
    postID, _ := uuid.Parse(req.PostId)

    isLiked, err := s.repo.ToggleLike(ctx, postID.String(), userID.String())
    if err != nil {
        return err
    }

    if !isLiked {
        return nil
    }

    post, err := s.repo.GetPostByID(ctx, postID.String())
    if err != nil {
        return err
    }

    if post.UserID == userID {
        return nil
    }

    go func() {
        likerProfile, err := s.userClient.GetUserProfile(context.Background(), &userPb.GetUserProfileRequest{UserId: req.UserId})
        if err != nil {
            log.Printf("Failed to fetch liker profile: %v", err)
            return
        }

		log.Printf("[DEBUG] Liker Profile Fetched. Preparing event...")

        event := NotificationEvent{
            RecipientID: post.UserID.String(),
            SenderID:    req.UserId,          
            SenderName:  likerProfile.Username,
            SenderImage: likerProfile.ProfilePictureUrl,
            Type:        "like",
            EntityID:    post.ID.String(),    
            Message:     "liked your post",
        }

        s.publishNotification(event)
    }()

    return nil
}

func (s *PostService) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*domain.PostComment, error) {
    userID, _ := uuid.Parse(req.UserId)
    postID, _ := uuid.Parse(req.PostId)

    comment := &domain.PostComment{
        PostID:  postID,
        UserID:  userID,
        Content: req.Content,
    }

    err := s.repo.CreateComment(ctx, comment)
    if err != nil {
        return nil, err
    }

    post, err := s.repo.GetPostByID(ctx, postID.String())
    if err != nil {
        return comment, nil
    }

    if post.UserID == userID {
        return comment, nil
    }

    go func() {
        senderProfile, err := s.userClient.GetUserProfile(context.Background(), &userPb.GetUserProfileRequest{UserId: req.UserId})
        if err != nil { return }

        event := NotificationEvent{
            RecipientID: post.UserID.String(),
            SenderID:    req.UserId,
            SenderName:  senderProfile.Username,
            SenderImage: senderProfile.ProfilePictureUrl,
            Type:        "comment",
            EntityID:    post.ID.String(),
            Message:     "commented on your post",
        }

        s.publishNotification(event)
    }()

    return comment, nil
}

func (s *PostService) GetUserReels(ctx context.Context, userID string) ([]*domain.Post, error) {
    return s.repo.GetReelsByUserID(ctx, userID)
}

func (s *PostService) GetCollectionPosts(ctx context.Context, collectionID string, limit, offset int) ([]*domain.Post, error) {
    return s.repo.GetCollectionPosts(ctx, collectionID, limit, offset)
}

func (s *PostService) UpdateCollection(ctx context.Context, collectionID, name, userID string) (*domain.Collection, error) {
    return s.repo.UpdateCollection(ctx, collectionID, name, userID)
}

func (s *PostService) DeleteCollection(ctx context.Context, collectionID, userID string) error {
    return s.repo.DeleteCollection(ctx, collectionID, userID)
}

func (s *PostService) DeletePost(ctx context.Context, postID, userID string) error {
    post, err := s.repo.GetPostByID(ctx, postID)
    if err != nil {
        return fmt.Errorf("post not found or error fetching: %v", err)
    }

    if post.UserID.String() != userID {
        return fmt.Errorf("unauthorized: you are not the owner of this post")
    }

    return s.repo.DeletePost(ctx, postID)
}

func (s *PostService) SearchHashtags(ctx context.Context, query string) ([]ports.HashtagSearchParam, error) {
    cleanQuery := strings.TrimPrefix(query, "#")
    return s.repo.SearchHashtags(ctx, cleanQuery)
}