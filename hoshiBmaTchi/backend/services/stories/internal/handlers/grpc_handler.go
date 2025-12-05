package handlers

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/stories"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/clients"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/core/ports"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/events"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCHandler struct {
	pb.UnimplementedStoriesServiceServer
	repo              ports.StoryRepository
	redisRepo         ports.RedisRepository
	userServiceClient *clients.UserServiceClient
	chatServiceClient *clients.ChatServiceClient 
	publisher         *events.EventPublisher
	minioClient *minio.Client
    bucketName  string
}

func NewGRPCHandler(
	repo ports.StoryRepository,
	redisRepo ports.RedisRepository,
	userClient *clients.UserServiceClient,
	chatClient *clients.ChatServiceClient,
	publisher *events.EventPublisher,
	minioClient *minio.Client,
    bucketName string,
) *GRPCHandler {
	return &GRPCHandler{
		repo:              repo,
		redisRepo:         redisRepo,
		userServiceClient: userClient,
		chatServiceClient: chatClient,
		publisher:         publisher,
		minioClient: minioClient,
        bucketName:  bucketName,
	}
}

func (h *GRPCHandler) CreateStory(ctx context.Context, req *pb.CreateStoryRequest) (*pb.CreateStoryResponse, error) {

	fullMediaURL := fmt.Sprintf("http://%s/%s/%s", os.Getenv("MINIO_ENDPOINT"), h.bucketName, req.MediaUrl)
	
	story := &domain.Story{
		UserID:    req.UserId,
		MediaURL:  fullMediaURL,
		MediaType: domain.MediaType(req.MediaType.String()),
		Duration:  int(req.Duration),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), 
	}

	if err := h.repo.CreateStory(ctx, story); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create story: %v", err)
	}

	return &pb.CreateStoryResponse{
		Story: h.domainStoryToPb(story),
	}, nil
}

func (h *GRPCHandler) GenerateUploadURL(ctx context.Context, req *pb.GenerateUploadURLRequest) (*pb.GenerateUploadURLResponse, error) {
    objectName := uuid.New().String() + "-" + req.FileName
    expiry := 15 * time.Minute

    // Generate Presigned URL
    presignedURL, err := h.minioClient.PresignedPutObject(ctx, h.bucketName, objectName, expiry)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to generate upload url: %v", err)
    }

    return &pb.GenerateUploadURLResponse{
        UploadUrl:  presignedURL.String(),
        ObjectName: objectName,
    }, nil
}

func (h *GRPCHandler) GetStory(ctx context.Context, req *pb.GetStoryRequest) (*pb.GetStoryResponse, error) {
	story, err := h.repo.GetStoryByID(ctx, req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "story not found: %v", err)
	}

	isLiked, _ := h.repo.IsStoryLiked(ctx, req.StoryId, req.UserId)
	isViewed, _ := h.repo.IsStoryViewed(ctx, req.StoryId, req.UserId)

	return &pb.GetStoryResponse{
		Story:    h.domainStoryToPb(story),
		IsLiked:  isLiked,
		IsViewed: isViewed,
	}, nil
}

func (h *GRPCHandler) GetUserStories(ctx context.Context, req *pb.GetUserStoriesRequest) (*pb.GetUserStoriesResponse, error) {
    limit := int(req.Limit)
    if limit <= 0 {
        limit = 20 
    }
    offset := int(req.Offset)

    stories, err := h.repo.GetUserStories(ctx, req.UserId, req.IsArchive, limit, offset)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get user stories: %v", err)
    }

    pbStories := make([]*pb.Story, len(stories))
    for i, story := range stories {
        pbStories[i] = h.domainStoryToPb(story)
    }

    return &pb.GetUserStoriesResponse{
        Stories: pbStories,
    }, nil
}

func (h *GRPCHandler) GetFollowingStories(ctx context.Context, req *pb.GetFollowingStoriesRequest) (*pb.GetFollowingStoriesResponse, error) {
	cachedStories, err := h.redisRepo.GetUserFeed(ctx, req.UserId)
	if err == nil && cachedStories != nil {
		return h.groupStoriesToResponse(ctx, cachedStories, req.UserId), nil
	}

	followingIDs, err := h.userServiceClient.GetFollowing(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get following: %v", err)
	}

	storiesMap, err := h.repo.GetFollowingStories(ctx, followingIDs, req.UserId, int(req.Limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get following stories: %v", err)
	}

	var allStories []*domain.Story
	for _, userStories := range storiesMap {
		allStories = append(allStories, userStories...)
	}

	go func(uid string, stories []*domain.Story) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = h.redisRepo.SetUserFeed(bgCtx, uid, stories)
	}(req.UserId, allStories)

	return h.groupStoriesToResponse(ctx, allStories, req.UserId), nil
}

func (h *GRPCHandler) groupStoriesToResponse(ctx context.Context, stories []*domain.Story, viewerID string) *pb.GetFollowingStoriesResponse {
	grouped := make(map[string][]*domain.Story)
	for _, s := range stories {
		grouped[s.UserID] = append(grouped[s.UserID], s)
	}

	var userStories []*pb.UserStories

	for userID, userStoryList := range grouped {
		var pbStories []*pb.Story
		hasUnseen := false

		for _, story := range userStoryList {
			pbStories = append(pbStories, h.domainStoryToPb(story))

			// Check if viewed (This N+1 check might be slow; optimize by fetching all views for these stories in 1 query if needed)
			viewed, _ := h.repo.IsStoryViewed(ctx, story.ID, viewerID)
			if !viewed {
				hasUnseen = true
			}
		}

		userStories = append(userStories, &pb.UserStories{
			UserId:    userID,
			Stories:   pbStories,
			HasUnseen: hasUnseen,
		})
	}

	return &pb.GetFollowingStoriesResponse{
		UserStories: userStories,
	}
}

func (h *GRPCHandler) DeleteStory(ctx context.Context, req *pb.DeleteStoryRequest) (*pb.DeleteStoryResponse, error) {
	if err := h.repo.DeleteStory(ctx, req.StoryId, req.UserId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete story: %v", err)
	}

	return &pb.DeleteStoryResponse{Success: true}, nil
}

func (h *GRPCHandler) ViewStory(ctx context.Context, req *pb.ViewStoryRequest) (*pb.ViewStoryResponse, error) {
	view := &domain.StoryView{
		StoryID:  req.StoryId,
		UserID:   req.UserId,
		ViewedAt: time.Now(),
	}

	if err := h.repo.ViewStory(ctx, view); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to view story: %v", err)
	}

	return &pb.ViewStoryResponse{Success: true}, nil
}

func (h *GRPCHandler) LikeStory(ctx context.Context, req *pb.LikeStoryRequest) (*pb.LikeStoryResponse, error) {
	// 1. Check if story exists to get Owner ID
	story, err := h.repo.GetStoryByID(ctx, req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "story not found")
	}

	// 2. Save Like to DB
	like := &domain.StoryLike{
		StoryID:   req.StoryId,
		UserID:    req.UserId,
		LikedAt: time.Now(),
	}

	if err := h.repo.LikeStory(ctx, like); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to like story: %v", err)
	}

	// 3. Publish Notification Event (if not liking own story)
	if story.UserID != req.UserId && h.publisher != nil {
		go func() {
			err := h.publisher.PublishNotification(context.Background(), events.NotificationEvent{
				Type:       "story_like",
				ActorID:    req.UserId,
				TargetID:   story.UserID,
				ResourceID: story.ID,
				Message:    "liked your story.",
				CreatedAt:  time.Now(),
			})
			if err != nil {
				log.Printf("Failed to publish notification: %v", err)
			}
		}()
	}

	return &pb.LikeStoryResponse{Success: true}, nil
}

func (h *GRPCHandler) UnlikeStory(ctx context.Context, req *pb.UnlikeStoryRequest) (*pb.UnlikeStoryResponse, error) {
	if err := h.repo.UnlikeStory(ctx, req.StoryId, req.UserId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unlike story: %v", err)
	}

	return &pb.UnlikeStoryResponse{Success: true}, nil
}

func (h *GRPCHandler) ReplyToStory(ctx context.Context, req *pb.ReplyToStoryRequest) (*pb.ReplyToStoryResponse, error) {
	// 1. Get Story for Owner ID
	story, err := h.repo.GetStoryByID(ctx, req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "story not found")
	}

	reply := &domain.StoryReply{
		StoryID:   req.StoryId,
		UserID:    req.UserId,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	if err := h.repo.CreateReply(ctx, reply); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reply to story: %v", err)
	}

	// 2. Publish Notification Event
	if story.UserID != req.UserId && h.publisher != nil {
		go func() {
			err := h.publisher.PublishNotification(context.Background(), events.NotificationEvent{
				Type:       "story_reply",
				ActorID:    req.UserId,
				TargetID:   story.UserID,
				ResourceID: story.ID,
				Message:    fmt.Sprintf("replied to your story: %s", req.Content),
				CreatedAt:  time.Now(),
			})
			if err != nil {
				log.Printf("Failed to publish notification: %v", err)
			}
		}()
	}

	return &pb.ReplyToStoryResponse{
		Reply: &pb.StoryReply{
			Id:        reply.ID,
			StoryId:   reply.StoryID,
			UserId:    reply.UserID,
			Content:   reply.Content,
			CreatedAt: timestamppb.New(reply.CreatedAt),
		},
	}, nil
}

func (h *GRPCHandler) GetStoryReplies(ctx context.Context, req *pb.GetStoryRepliesRequest) (*pb.GetStoryRepliesResponse, error) {
	replies, err := h.repo.GetStoryReplies(ctx, req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get story replies: %v", err)
	}

	pbReplies := make([]*pb.StoryReply, len(replies))
	for i, reply := range replies {
		pbReplies[i] = &pb.StoryReply{
			Id:        reply.ID,
			StoryId:   reply.StoryID,
			UserId:    reply.UserID,
			Content:   reply.Content,
			CreatedAt: timestamppb.New(reply.CreatedAt),
		}
	}

	return &pb.GetStoryRepliesResponse{Replies: pbReplies}, nil
}

func (h *GRPCHandler) GetStoryViewers(ctx context.Context, req *pb.GetStoryViewersRequest) (*pb.GetStoryViewersResponse, error) {
	story, err := h.repo.GetStoryByID(ctx, req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "story not found: %v", err)
	}

	if story.UserID != req.UserId {
		return nil, status.Errorf(codes.PermissionDenied, "not authorized to view story viewers")
	}

	viewers, err := h.repo.GetStoryViewers(ctx, req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get story viewers: %v", err)
	}

	pbViewers := make([]*pb.StoryViewer, len(viewers))
	for i, viewer := range viewers {
		pbViewers[i] = &pb.StoryViewer{
			UserId:   viewer.UserID,
			ViewedAt: timestamppb.New(viewer.ViewedAt),
		}
	}

	return &pb.GetStoryViewersResponse{Viewers: pbViewers}, nil
}

func (h *GRPCHandler) ShareStory(ctx context.Context, req *pb.ShareStoryRequest) (*pb.ShareStoryResponse, error) {
	_, err := h.repo.GetStoryByID(ctx, req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "story not found: %v", err)
	}

	messageID, err := h.chatServiceClient.SendMessage(ctx, req.SenderId, req.RecipientId, "Shared a story", req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send story: %v", err)
	}

	share := &domain.StoryShare{
		StoryID:     req.StoryId,
		SenderID:    req.SenderId,
		RecipientID: req.RecipientId,
		MessageID:   messageID,
		CreatedAt:   time.Now(),
	}

	if err := h.repo.ShareStory(ctx, share); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to record share: %v", err)
	}

	return &pb.ShareStoryResponse{
		Success:   true,
		MessageId: messageID,
	}, nil
}

func (h *GRPCHandler) domainStoryToPb(story *domain.Story) *pb.Story {
	mediaType := pb.MediaType_IMAGE
	if story.MediaType == domain.MediaTypeVideo {
		mediaType = pb.MediaType_VIDEO
	}

	return &pb.Story{
		Id:         story.ID,
		UserId:     story.UserID,
		MediaUrl:   story.MediaURL,
		MediaType:  mediaType,
		Duration:   int32(story.Duration),
		CreatedAt:  timestamppb.New(story.CreatedAt),
		ExpiresAt:  timestamppb.New(story.ExpiresAt),
		ViewCount:  int32(len(story.Views)),
		LikeCount:  int32(len(story.Likes)),
		ReplyCount: int32(story.ReplyCount),
	}
}

func (h *GRPCHandler) ToggleStoryVisibility(ctx context.Context, req *pb.ToggleStoryVisibilityRequest) (*pb.ToggleStoryVisibilityResponse, error) {
    if req.UserId == "" || req.TargetId == "" {
        return nil, status.Error(codes.InvalidArgument, "user_id and target_id are required")
    }

    isHidden, err := h.repo.ToggleStoryVisibility(ctx, req.UserId, req.TargetId)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to toggle visibility: %v", err)
    }

    return &pb.ToggleStoryVisibilityResponse{
        IsHidden: isHidden,
    }, nil
}

func (h *GRPCHandler) GetHiddenUsers(ctx context.Context, req *pb.GetHiddenUsersRequest) (*pb.GetHiddenUsersResponse, error) {
    if req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "user_id is required")
    }

    hiddenIDs, err := h.repo.GetHiddenUsers(ctx, req.UserId)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get hidden users: %v", err)
    }

    return &pb.GetHiddenUsersResponse{
        HiddenUserIds: hiddenIDs,
    }, nil
}

// [ADD THIS FUNCTION]
func (h *GRPCHandler) GetStoriesByAuthors(ctx context.Context, req *pb.GetStoriesByAuthorsRequest) (*pb.GetStoriesResponse, error) {
    // 1. Call the repo
    stories, err := h.repo.GetStoriesByAuthors(req.AuthorIds)
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to fetch stories")
    }

    var protoStories []*pb.Story
    for _, s := range stories {
        protoStories = append(protoStories, &pb.Story{
            Id:        s.ID,
            UserId:    s.UserID,
            MediaUrl:  s.MediaURL,
            MediaType: pb.MediaType(pb.MediaType_value[string(s.MediaType)]),
            Duration:  int32(s.Duration),
            CreatedAt: timestamppb.New(s.CreatedAt),
            ExpiresAt: timestamppb.New(s.ExpiresAt),
        })
    }

    return &pb.GetStoriesResponse{Stories: protoStories}, nil
}