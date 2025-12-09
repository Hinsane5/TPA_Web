package handlers

import (
	"context"
	// "errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/posts"
	userPb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/ports"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/posts/internal/core/services"
	"github.com/google/uuid"
	// "github.com/jackc/pgx/v5/pgconn"
	"github.com/minio/minio-go/v7"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedPostsServiceServer 
	repo           ports.PostRepository   
	service        *services.PostService      
	minio          *minio.Client 
	presignClient  *minio.Client  
	bucketName     string  
	publicEndPoint string    
	userClient     userPb.UserServiceClient     
}

func NewGRPCServer(repo ports.PostRepository, service *services.PostService, minio *minio.Client, presignClient *minio.Client, bucketName string, publicEndPoint string, userClient userPb.UserServiceClient) *Server {
	return &Server{
		repo:           repo,
		service: service,
		minio:          minio,
		presignClient:  presignClient,
		bucketName:     bucketName,
		publicEndPoint: publicEndPoint,
		userClient:     userClient,
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

	return &pb.GenerateUploadURLResponse{
		UploadUrl:  presignedURL.String(),
		ObjectName: objectName,
	}, nil
}

func (s *Server) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.CreatePostResponse, error){
	userID, err := uuid.Parse(req.GetUserId())

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "UserID tidak valid")
	}

	var mediaItems []domain.PostMedia
	for i, item := range req.GetMedia() {
		mediaItems = append(mediaItems, domain.PostMedia{
			MediaObjectName: item.MediaObjectName,
			MediaType:       item.MediaType,
			Sequence:        i,
		})
	}

	newPost := &domain.Post{
		UserID:   userID,
		Caption:  req.GetCaption(),
		Location: req.GetLocation(),
		Media:    mediaItems, 
		IsReel:    req.IsReel,
	}

	err = s.repo.CreatePost(ctx, newPost)
	if err != nil{
		log.Printf("Failed to save into DB: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create post")
	}

	var pbMedia []*pb.PostMediaResponse

	for _, m := range newPost.Media {
		pbMedia = append(pbMedia, &pb.PostMediaResponse{

			MediaType: m.MediaType,
		})
	}

	return &pb.CreatePostResponse{
		Post: &pb.PostResponse{
			Id:        newPost.ID.String(),
			UserId:    newPost.UserID.String(),
			Media:     pbMedia, 
			Caption:   newPost.Caption,
			Location:  newPost.Location,
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
		var pbMedia []*pb.PostMediaResponse
		
		for _, m := range post.Media {
			reqParams := make(url.Values)
			reqParams.Set("response-content-type", m.MediaType)

			presignedURL, err := s.presignClient.PresignedGetObject(ctx, s.bucketName, m.MediaObjectName, expiry, reqParams)
			
			mediaURLString := ""
			if err != nil {
				log.Printf("Failed to make Presigned GET URL for %s: %v", m.MediaObjectName, err)
			} else {
				mediaURLString = presignedURL.String()
			}

			pbMedia = append(pbMedia, &pb.PostMediaResponse{
				MediaUrl:  mediaURLString,
				MediaType: m.MediaType,
			})
		}

		pbPosts = append(pbPosts, &pb.PostResponse{
			Id:            post.ID.String(),
			UserId:        post.UserID.String(),
			Media:         pbMedia, 
			Caption:       post.Caption,
			Location:      post.Location,
			CreatedAt:     post.CreatedAt.Format(time.RFC3339),
			LikesCount:    post.LikesCount,
			CommentsCount: post.CommentsCount,
			IsLiked:       post.IsLiked,
		})
	}

	return &pb.GetPostsResponse{Posts: pbPosts}, nil
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

	var pbPosts []*pb.PostResponse
	expiry := time.Hour * 1

	for _, post := range posts {
		var pbMedia []*pb.PostMediaResponse

		for _, m := range post.Media {
			reqParams := make(url.Values)
			reqParams.Set("response-content-type", m.MediaType)

			presignedURL, err := s.presignClient.PresignedGetObject(ctx, s.bucketName, m.MediaObjectName, expiry, reqParams)
			
			mediaURLString := ""
			if err != nil {
				log.Printf("Failed to make Presigned GET URL for %s: %v", m.MediaObjectName, err)
			} else {
				mediaURLString = presignedURL.String()
			}

			pbMedia = append(pbMedia, &pb.PostMediaResponse{
				MediaUrl:  mediaURLString,
				MediaType: m.MediaType,
			})
		}

		pbPosts = append(pbPosts, &pb.PostResponse{
			Id:            post.ID.String(),
			UserId:        post.UserID.String(),
			Media:         pbMedia,
			Caption:       post.Caption,
			Location:      post.Location,
			CreatedAt:     post.CreatedAt.Format(time.RFC3339),
			LikesCount:    post.LikesCount,
			CommentsCount: post.CommentsCount,
			IsLiked:       post.IsLiked,
		})
	}

	return &pb.GetHomeFeedResponse{Posts: pbPosts}, nil
}

func (s *Server) LikePost(ctx context.Context, req *pb.LikePostRequest) (*pb.LikePostResponse, error){
	// FIX: Call the Service (which handles Notifications + Toggle logic), NOT the Repo directly.
	err := s.service.LikePost(ctx, req)
	if err != nil {
		log.Printf("LikePost service failed: %v", err)
		return nil, status.Error(codes.Internal, "Failed to toggle like")
	}

	return &pb.LikePostResponse{Message: "Success"}, nil
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
	// FIX: Call the Service (which handles Notifications), NOT the Repo directly.
	comment, err := s.service.CreateComment(ctx, req)
	if err != nil {
		log.Printf("CreateComment service failed: %v", err)
		return nil, status.Error(codes.Internal, "Failed to create comment")
	}

	return &pb.CommentResponse{
		Id:        comment.ID.String(),
		PostId:    comment.PostID.String(),
		UserId:    comment.UserID.String(),
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
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

func (h *Server) CreateCollection(ctx context.Context, req *pb.CreateCollectionRequest) (*pb.CollectionResponse, error) {
	collection := &domain.Collection{
		UserID: uuid.MustParse(req.UserId),
		Name:   req.Name,
	}

	if err := h.repo.CreateCollection(ctx, collection); err != nil {
		return nil, err
	}

	return &pb.CollectionResponse{
		Id:     collection.ID.String(),
		Name:   collection.Name,
		UserId: collection.UserID.String(),
	}, nil
}

func (h *Server) GetUserCollections(ctx context.Context, req *pb.GetUserCollectionsRequest) (*pb.GetUserCollectionsResponse, error) {
	collections, err := h.repo.GetUserCollections(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	var protoCollections []*pb.CollectionResponse
	for _, c := range collections {
		var covers []string
		for _, sp := range c.SavedPosts {
			if len(sp.Post.Media) > 0 {
				covers = append(covers, sp.Post.Media[0].MediaObjectName)
			}
		}

		protoCollections = append(protoCollections, &pb.CollectionResponse{
			Id:          c.ID.String(),
			Name:        c.Name,
			UserId:      c.UserID.String(),
			CoverImages: covers,
		})
	}

	return &pb.GetUserCollectionsResponse{Collections: protoCollections}, nil
}

func (h *Server) ToggleSavePost(ctx context.Context, req *pb.ToggleSavePostRequest) (*pb.ToggleSavePostResponse, error) {
	isSaved, err := h.repo.ToggleSavePost(ctx, req.UserId, req.PostId, req.CollectionId)
	if err != nil {
		return nil, err
	}

	msg := "Post unsaved"
	if isSaved {
		msg = "Post saved"
	}

	return &pb.ToggleSavePostResponse{
		IsSaved: isSaved,
		Message: msg,
	}, nil
}

func (s *Server) GetPostByID(ctx context.Context, req *pb.GetPostByIDRequest) (*pb.PostResponse, error) {
	return nil, nil
}

func (s *Server) GetUserMentions(ctx context.Context, req *pb.GetUserMentionsRequest) (*pb.GetPostsResponse, error) {
    posts, err := s.service.GetUserMentions(ctx, req)
    if err != nil {
        log.Printf("Failed to fetch mentions: %v", err)
        return nil, status.Error(codes.Internal, "Failed to fetch mentions")
    }

    var pbPosts []*pb.PostResponse
    expiry := time.Hour * 1

    for _, post := range posts {
        var pbMedia []*pb.PostMediaResponse
        
        for _, m := range post.Media {
            reqParams := make(url.Values)
            reqParams.Set("response-content-type", m.MediaType)
            
            presignedURL, err := s.presignClient.PresignedGetObject(ctx, s.bucketName, m.MediaObjectName, expiry, reqParams)
            
            finalURL := ""
            if err == nil {
                // --- THE MISSING PART ---
                // Replace internal Docker name 'minio' with 'localhost' so the browser can see it
                finalURL = strings.Replace(presignedURL.String(), "minio:9000", "localhost:9000", 1)
                finalURL = strings.Replace(finalURL, "http://backend:9000", "http://localhost:9000", 1)
            }
            
            pbMedia = append(pbMedia, &pb.PostMediaResponse{
                MediaUrl:  finalURL,
                MediaType: m.MediaType,
            })
        }

        pbPosts = append(pbPosts, &pb.PostResponse{
            Id:            post.ID.String(),
            UserId:        post.UserID.String(),
            Media:         pbMedia,
            Caption:       post.Caption,
            Location:      post.Location,
            CreatedAt:     post.CreatedAt.Format(time.RFC3339),
            LikesCount:    post.LikesCount,
            CommentsCount: post.CommentsCount,
        })
    }

    return &pb.GetPostsResponse{Posts: pbPosts}, nil
}

// Add this new function to the file
func (s *Server) GetReels(ctx context.Context, req *pb.GetReelsRequest) (*pb.GetReelsResponse, error) {
    posts, err := s.service.GetReelsFeed(ctx, int(req.Limit), int(req.Offset))
    if err != nil {
        log.Printf("Failed to fetch reels from DB: %v", err)
        return nil, status.Error(codes.Internal, "Failed to fetch reels")
    }

    var pbPosts []*pb.PostResponse
    expiry := time.Hour * 1

    for _, post := range posts {
        var pbMedia []*pb.PostMediaResponse
        
        for _, m := range post.Media {
            reqParams := make(url.Values)
            reqParams.Set("response-content-type", m.MediaType)
            
            // Generate Presigned URL
            presignedURL, err := s.presignClient.PresignedGetObject(ctx, s.bucketName, m.MediaObjectName, expiry, reqParams)
            
            mediaURLString := ""
            if err == nil {
                // Fix Docker networking for browser access
                mediaURLString = strings.Replace(presignedURL.String(), "minio:9000", "localhost:9000", 1)
                mediaURLString = strings.Replace(mediaURLString, "http://backend:9000", "http://localhost:9000", 1)
            }

            pbMedia = append(pbMedia, &pb.PostMediaResponse{
                MediaUrl:  mediaURLString,
                MediaType: m.MediaType,
            })
        }

        pbPosts = append(pbPosts, &pb.PostResponse{
            Id:            post.ID.String(),
            UserId:        post.UserID.String(),
            Media:         pbMedia,
            Caption:       post.Caption,
            Location:      post.Location,
            CreatedAt:     post.CreatedAt.Format(time.RFC3339),
            LikesCount:    post.LikesCount,
            CommentsCount: post.CommentsCount,
            IsLiked:       post.IsLiked,
            IsReel:        post.IsReel,
        })
    }

    return &pb.GetReelsResponse{Posts: pbPosts}, nil
}

// ... inside Server struct methods ...

func (s *Server) GetExplorePosts(ctx context.Context, req *pb.GetExplorePostsRequest) (*pb.GetExplorePostsResponse, error) {
    // 1. Call Service
    posts, err := s.service.GetExplorePosts(ctx, int(req.Limit), int(req.Offset), req.Hashtag)
    if err != nil {
        log.Printf("Failed to fetch explore posts: %v", err)
        return nil, status.Error(codes.Internal, "Failed to fetch explore posts")
    }

    var pbPosts []*pb.PostResponse
    expiry := time.Hour * 1

    // 2. Map Domain to Proto & Generate Presigned URLs
    for _, post := range posts {
        var pbMedia []*pb.PostMediaResponse
        
        for _, m := range post.Media {
            reqParams := make(url.Values)
            reqParams.Set("response-content-type", m.MediaType)
            
            presignedURL, err := s.presignClient.PresignedGetObject(ctx, s.bucketName, m.MediaObjectName, expiry, reqParams)
            
            mediaURLString := ""
            if err == nil {
                // Fix Docker networking for browser access (replace minio host with localhost)
                mediaURLString = strings.Replace(presignedURL.String(), "minio:9000", "localhost:9000", 1)
                mediaURLString = strings.Replace(mediaURLString, "http://backend:9000", "http://localhost:9000", 1)
            }

            pbMedia = append(pbMedia, &pb.PostMediaResponse{
                MediaUrl:  mediaURLString,
                MediaType: m.MediaType,
            })
        }

        pbPosts = append(pbPosts, &pb.PostResponse{
            Id:            post.ID.String(),
            UserId:        post.UserID.String(),
            Media:         pbMedia,
            Caption:       post.Caption,
            Location:      post.Location,
            CreatedAt:     post.CreatedAt.Format(time.RFC3339),
            LikesCount:    post.LikesCount,
            CommentsCount: post.CommentsCount,
            IsLiked:       post.IsLiked,
            IsReel:        post.IsReel,
        })
    }

    return &pb.GetExplorePostsResponse{Posts: pbPosts}, nil
}