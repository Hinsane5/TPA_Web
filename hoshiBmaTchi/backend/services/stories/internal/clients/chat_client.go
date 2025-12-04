package clients

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/chat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type ChatServiceClient struct {
	client pb.ChatServiceClient
	conn   *grpc.ClientConn
}

func NewChatServiceClient(address string) (*ChatServiceClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to chat service: %w", err)
	}

	client := pb.NewChatServiceClient(conn)
	log.Printf("✅ Connected to chat service at %s", address)

	return &ChatServiceClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *ChatServiceClient) SendMessage(ctx context.Context, senderID, recipientID, content, storyID string) (string, error) {
	if c == nil || c.client == nil {
		return "", fmt.Errorf("chat service client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.SendMessageRequest{
		SenderId:    senderID,
		RecipientId: recipientID,
		Content:     content,
		MessageType: pb.MessageType_TEXT,
	}

	if storyID != "" {
		req.MessageType = pb.MessageType_STORY_SHARE
		req.StoryId = storyID
	}

	resp, err := c.client.SendMessage(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	if resp == nil {
		return "", fmt.Errorf("empty response from chat service")
	}

	messageID := resp.MessageId
	if messageID == "" && resp.Message != nil {
		messageID = resp.Message.Id
	}

	log.Printf("✅ Message sent successfully: %s", messageID)
	return messageID, nil
}

func (c *ChatServiceClient) Close() error {
	if c.conn != nil {
		log.Println("Closing chat service connection")
		return c.conn.Close()
	}
	return nil
}