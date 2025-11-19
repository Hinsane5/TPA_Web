package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	// "strings"
	"time"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/posts"
	userPb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/minio/minio-go/v7"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedPostsServiceServer 
	repo ports.PostRepository         
	minio *minio.Client 
	presignClient  *minio.Client  
	bucketName string  
	publicEndPoint string    
	userClient userPb.UserServiceClient     
}

func NewGRPCServer(repo ports.PostRepository, minio *minio.Client, presignClient *minio.Client, bucketName string, publicEndPoint string, userClient userPb.UserServiceClient) *Server {
	return &Server{
		repo:           repo,
		minio:          minio,
		presignClient:  presignClient,
		bucketName:     bucketName,
		publicEndPoint: publicEndPoint,
		userClient: userClient,
	}
}

func (s *Server) GenerateUploadURL(ctx context.Context, req *pb.GenerateUploadURLRequest) (*pb.GenerateUploadURLResponse, error){
	bucketName := s.bucketName

	objectName := uuid.New().String() + "-" + req.GetFileName()

	expiry := 15 * time.Minute

	headers := make(http.Header)
	if req.GetFileType() != "" {
		headers.Set("Content-Type", req.GetFileType())
	}

	reqParams := make(url.Values)

	presignedURL, err := s.presignClient.PresignHeader(ctx, "PUT", bucketName, objectName, expiry, reqParams, headers)
	if err != nil {
		log.Printf("Failed to generate presigned URL: %v", err)
		return nil, status.Error(codes.Internal, "Failed to generate posts url")
	}

	// internalEndpoint := s.minio.EndpointURL().String()

	// publicURL := strings.Replace(presignedURL.String(), internalEndpoint, s.publicEndPoint, 1)


	return &pb.GenerateUploadURLResponse{
		UploadUrl: presignedURL.String(),
		ObjectName: objectName,
	}, nil
}

func (s *Server) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.CreatePostResponse, error){
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "UserID tidak valid")
	}


	newPost := &domain.Post{
		UserID:          userID,
		MediaObjectName: req.GetMediaObjectName(),
		MediaType:       req.GetMediaType(),
		Caption:         req.GetCaption(),
		Location:        req.GetLocation(),

	}

	err = s.repo.CreatePost(ctx, newPost)
	if err != nil{
		log.Printf("Failed to save into DB: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create post")
	}

	return &pb.CreatePostResponse{
		Post: &pb.PostResponse{
			Id: newPost.ID.String(),
			UserId: newPost.UserID.String(),
			MediaType: newPost.MediaType,
			Caption: newPost.Caption,
			Location: newPost.Location,
			CreatedAt: newPost.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (s *Server) GetPostsByUserID(ctx context.Context, req *pb.GetPostsByUserIDRequest) (*pb.GetPostsResponse, error){
	if req.GetUserId() == ""{
		return nil, status.Error(codes.InvalidArgument, "UserID needed")
	}

	posts, err := s.repo.GetPostsByUserID(ctx, req.GetUserId())

	if err != nil {
		log.Printf("Failed to takes post from DB: %v", err)
		return nil, status.Error(codes.Internal, "Failed to takes post")
	}

	var pbPosts []*pb.PostResponse
	expiry := time.Hour * 1

	for _, post := range posts{
		reqParams := make(url.Values)
		reqParams.Set("response-content-type", post.MediaType)

		presignedURL, err := s.presignClient.PresignedGetObject(ctx, s.bucketName, post.MediaObjectName, expiry, reqParams)
		
		if err != nil {
			log.Printf("Failed to to make Presigned GET URL for %s: %v", post.MediaObjectName, err)
			presignedURL = nil
		}

		mediaURLString := ""
		if presignedURL != nil {
			mediaURLString = presignedURL.String()
		}

		pbPosts = append(pbPosts, &pb.PostResponse{
			Id:        post.ID.String(),
			UserId:    post.UserID.String(),
			MediaUrl:  mediaURLString, 
			MediaType: post.MediaType,
			Caption:   post.Caption,
			Location:  post.Location,
			CreatedAt: post.CreatedAt.Format(time.RFC3339),
		})
	}

	return &pb.GetPostsResponse{Posts: pbPosts}, nil
}

func (s *Server) LikePost(ctx context.Context, req *pb.LikePostRequest) (*pb.LikePostResponse, error){
	userID, err := uuid.Parse(req.GetUserId())

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid UserID")
	}
	postID, err := uuid.Parse(req.GetPostId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid PostID")
	}

	newLike := &domain.PostLike{
		UserID: userID,
		PostID: postID,
	}

	err = s.repo.LikePost(ctx, newLike)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505"{
			return &pb.LikePostResponse{Message: "Already liked this post"}, nil
		}

		log.Printf("Likedpost failed: %v", err)
		return nil, status.Error(codes.Internal, "Failed to like post")
	}

	return &pb.LikePostResponse{Message: "Successfully like post"}, nil
}

func (s *Server) UnlikePost(ctx context.Context, req *pb.UnlikePostRequest) (*pb.UnlikePostResponse, error) {
	err := s.repo.UnlikePost(ctx, req.GetUserId(), req.GetPostId())
	if err != nil {
		log.Printf("Failed UnlikePost: %v", err)
		return nil, status.Error(codes.Internal, "Falied to unlike postingan")
	}

	return &pb.UnlikePostResponse{Message: "Successfully unlike post"}, nil
}

func (s *Server) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*pb.CommentResponse, error){
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "UserID tidak valid")
	}
	postID, err := uuid.Parse(req.GetPostId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "PostID tidak valid")
	}
	if req.GetContent() == "" {
		return nil, status.Error(codes.InvalidArgument, "Komentar tidak boleh kosong")
	}

	newComment := &domain.PostComment{
		UserID:  userID,
		PostID:  postID,
		Content: req.GetContent(),
	}

	err = s.repo.CreateComment(ctx, newComment)
	if err != nil {
		log.Printf("Gagal CreateComment: %v", err)
		return nil, status.Error(codes.Internal, "Gagal menyimpan komentar")
	}

	return &pb.CommentResponse{
		Id:        newComment.ID.String(),
		PostId:    newComment.PostID.String(),
		UserId:    newComment.UserID.String(),
		Content:   newComment.Content,
		CreatedAt: newComment.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Server) GetCommentsForPost(ctx context.Context, req *pb.GetCommentsForPostRequest) (*pb.GetCommentsForPostResponse, error) {
	comments, err := s.repo.GetCommentsForPost(ctx, req.GetPostId())
	if err != nil {
		log.Printf("Failed to GetCommentsForPost: %v", err)
		return nil, status.Error(codes.Internal, "Failed to get comments")
	}

	var pbComments []*pb.CommentResponse
	for _, comment := range comments {
		pbComments = append(pbComments, &pb.CommentResponse{
			Id:        comment.ID.String(),
			PostId:    comment.PostID.String(),
			UserId:    comment.UserID.String(),
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		})
	}

	return &pb.GetCommentsForPostResponse{Comments: pbComments}, nil
}

func (s *Server) GetHomeFeed(ctx context.Context, req *pb.GetHomeFeedRequest) (*pb.GetHomeFeedResponse, error){
	userRes, err := s.userClient.GetFollowingList(ctx, &userPb.GetFollowingListRequest{
		UserId: req.UserId,
	})

	if err != nil {
		log.Printf("Failed to get following list from user service: %v", err)
		return nil, status.Error(codes.Internal, "Failed to fetch feed configuration")
	}

	authorIDs := append(userRes.FollowingIds, req.UserId)

	posts, err := s.repo.GetFeedPosts(ctx, authorIDs, req.UserId, int(req.Limit), int(req.Offset))
	if err != nil {
		log.Printf("Failed to fetch feed posts from DB: %v", err)
		return nil, status.Error(codes.Internal, "Failed to fetch posts")
	}

	// 4. Convert to Protobuf response
	var pbPosts []*pb.PostResponse
	expiry := time.Hour * 1

	for _, post := range posts {
		// Generate Presigned URL for viewing
		reqParams := make(url.Values)
		reqParams.Set("response-content-type", post.MediaType)

		presignedURL, err := s.presignClient.PresignedGetObject(ctx, s.bucketName, post.MediaObjectName, expiry, reqParams)
		mediaURLString := ""
		if err == nil {
			mediaURLString = presignedURL.String()
		}

		pbPosts = append(pbPosts, &pb.PostResponse{
			Id:        post.ID.String(),
			UserId:    post.UserID.String(),
			MediaUrl:  mediaURLString,
			MediaType: post.MediaType,
			Caption:   post.Caption,
			Location:  post.Location,
			CreatedAt: post.CreatedAt.Format(time.RFC3339),
		})
	}

	return &pb.GetHomeFeedResponse{Posts: pbPosts}, nil


}
