package services

import (
	"context"
	"encoding/json"
	"log"
	"regexp"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/posts"
	userPb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/ports"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

var mentionRegex = regexp.MustCompile(`@([a-zA-Z0-9._]+)`)

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

	go func() {
		// Fetch Sender Profile for Notification Payload
		senderProfile, err := s.userClient.GetUserProfile(context.Background(), &userPb.GetUserProfileRequest{UserId: req.UserId})
		if err != nil {
			log.Printf("Failed to fetch sender profile for notification: %v", err)
			return
		}
		
		for username := range uniqueUsernames {
			searchRes, err := s.userClient.SearchUsers(context.Background(), &userPb.SearchUsersRequest{Query: username})
			if err == nil && len(searchRes.Users) > 0 {
				targetUser := searchRes.Users[0] 
				if targetUser.Username == username {
					
					event := NotificationEvent{
						RecipientID: targetUser.UserId, 
						SenderID:    req.UserId,
						SenderName:  senderProfile.Username,
						SenderImage: senderProfile.ProfilePictureUrl,
						Type:        "mention",
						EntityID:    post.ID.String(), 
						Message:     "mentioned you in a post",
					}

					s.publishNotification(event)
				}
			}
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
	err := s.amqpChan.PublishWithContext(context.Background(),
		"notification_exchange",
		"notification.event",
		false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("Failed to publish notification: %v", err)
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