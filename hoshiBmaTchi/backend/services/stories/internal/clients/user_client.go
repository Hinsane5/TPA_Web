package clients

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type UserServiceClient struct {
	client pb.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUserServiceClient(address string) (*UserServiceClient, error) {
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
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	client := pb.NewUserServiceClient(conn)
	log.Printf("✅ Connected to user service at %s", address)

	return &UserServiceClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *UserServiceClient) GetFollowing(ctx context.Context, userID string) ([]string, error) {
    if c == nil || c.client == nil {
        log.Println("⚠️ User service client not initialized")
        return []string{}, nil 
    }


    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    resp, err := c.client.GetFollowingList(ctx, &pb.GetFollowingListRequest{
        UserId: userID,
    })
    if err != nil {
        log.Printf("⚠️ Failed to get following list for user %s: %v", userID, err)
        return []string{}, nil 
    }

    if resp == nil {
        log.Printf("⚠️ Empty response from GetFollowingList for user %s", userID)
        return []string{}, nil
    }

    followingIDs := resp.FollowingIds
    if followingIDs == nil {
        followingIDs = []string{}
    }

    log.Printf("✅ Found %d following users for user %s", len(followingIDs), userID)
    return followingIDs, nil
}

func (c *UserServiceClient) Close() error {
	if c.conn != nil {
		log.Println("Closing user service connection")
		return c.conn.Close()
	}
	return nil
}
