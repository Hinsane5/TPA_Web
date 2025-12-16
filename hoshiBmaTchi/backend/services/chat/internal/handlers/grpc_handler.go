package handlers

import (
	"context"
	"hash/crc32"
	"os"
	"time"

	"github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/rtctokenbuilder2"
	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/chat"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/chat/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/chat/internal/repositories"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
) 

type ChatGRPCServer struct {
	pb.UnimplementedChatServiceServer
	repo *repositories.ChatRepository
}

func NewChatGRPCServer(repo *repositories.ChatRepository) *ChatGRPCServer {
	return &ChatGRPCServer{repo: repo}
}

func (s *ChatGRPCServer) CreateGroupChat(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupResponse, error){
	conv := &domain.Conversation{
		Name:      req.Name,
		IsGroup:   true,
		CreatedAt: time.Now(),
	}


	err := s.repo.CreateConversation(ctx, conv, req.UserIds)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to create group")
	}

	return &pb.CreateGroupResponse{ConversationId: conv.ID.String()}, nil
}

func (s *ChatGRPCServer) GetConversations(ctx context.Context, req *pb.GetConversationsRequest) (*pb.GetConversationsResponse, error) {
	convs, err := s.repo.GetConversations(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to fetch conversations")
	}

	var pbConvs []*pb.Conversation
	for _, c := range convs {
		pbConvs = append(pbConvs, &pb.Conversation{
			Id:        c.ID.String(),
			Name:      c.Name,
			IsGroup:   c.IsGroup,
		})
	}

	return &pb.GetConversationsResponse{Conversations: pbConvs}, nil
}

func (s *ChatGRPCServer) GetMessageHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	msgs, err := s.repo.GetMessageHistory(ctx, req.ConversationId, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to fetch history")
	}

	var pbMsgs []*pb.Message
	for _, m := range msgs {
		pbMsgs = append(pbMsgs, &pb.Message{
			Id:        m.ID.String(),
			SenderId:  m.SenderID.String(),
			Content:   m.Content,
			MediaUrl:  m.MediaURL,
			CreatedAt: m.CreatedAt.Format(time.RFC3339),
			IsUnsent:  m.IsUnsent,
		})
	}

	return &pb.GetHistoryResponse{Messages: pbMsgs}, nil
}

func (s *ChatGRPCServer) GetCallToken(ctx context.Context, req *pb.GetCallTokenRequest) (*pb.GetCallTokenResponse, error) {
	appID := os.Getenv("AGORA_APP_ID")
	appCertificate := os.Getenv("AGORA_APP_CERTIFICATE")

	if appID == "" || appCertificate == "" {
		return nil, status.Error(codes.Internal, "Agora credentials not configured on server")
	}

	uid := crc32.ChecksumIEEE([]byte(req.UserId))

	tokenExpireTimeInSeconds := uint32(86400)
	privilegeExpireTimeInSeconds := uint32(86400)

	// 4. Generate Token
	channelName := req.ConversationId
	token, err := rtctokenbuilder2.BuildTokenWithUid(
		appID,
		appCertificate,
		channelName,
		uid,
		rtctokenbuilder2.RolePublisher,
		tokenExpireTimeInSeconds,
		privilegeExpireTimeInSeconds,
	)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate token: %v", err)
	}

	return &pb.GetCallTokenResponse{
		Token:       token,
		ChannelName: channelName,
		AppId:       appID,
	}, nil
}