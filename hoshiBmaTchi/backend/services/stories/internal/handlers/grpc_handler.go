package handlers

import (
	"context"
	"time"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/stories"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/stories/internal/core/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCHandler struct {
	pb.UnimplementedStoriesServiceServer
	repo              ports.StoryRepository
	userServiceClient ports.UserServiceClient
	chatServiceClient ports.ChatServiceClient
}

func NewGRPCHandler(repo ports.StoryRepository, userClient ports.UserServiceClient, chatClient ports.ChatServiceClient) *GRPCHandler {
	return &GRPCHandler{
		repo:              repo,
		userServiceClient: userClient,
		chatServiceClient: chatClient,
	}
}

func (h *GRPCHandler) CreateStory(ctx context.Context, req *pb.CreateStoryRequest) (*pb.CreateStoryResponse, error) {
	story := &domain.Story{
		UserID:    req.UserId,
		MediaURL:  req.MediaUrl,
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
	stories, err := h.repo.GetUserStories(ctx, req.UserId)
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
	followingIDs, err := h.userServiceClient.GetFollowing(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get following: %v", err)
	}

	storiesByUser, err := h.repo.GetFollowingStories(ctx, followingIDs, int(req.Limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get following stories: %v", err)
	}

	var userStories []*pb.UserStories
	for userID, stories := range storiesByUser {
		pbStories := make([]*pb.Story, len(stories))
		hasUnseen := false

		for i, story := range stories {
			pbStories[i] = h.domainStoryToPb(story)
			viewed, _ := h.repo.IsStoryViewed(ctx, story.ID, req.UserId)
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
	}, nil
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
	like := &domain.StoryLike{
		StoryID:   req.StoryId,
		UserID:    req.UserId,
		CreatedAt: time.Now(),
	}

	if err := h.repo.LikeStory(ctx, like); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to like story: %v", err)
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
	reply := &domain.StoryReply{
		StoryID:   req.StoryId,
		UserID:    req.UserId,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	if err := h.repo.CreateReply(ctx, reply); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reply to story: %v", err)
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
		ViewCount:  int32(story.ViewCount),
		LikeCount:  int32(story.LikeCount),
		ReplyCount: int32(story.ReplyCount),
	}
}
