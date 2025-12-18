package handlers

import (
	"context"
	"encoding/json"
	"fmt"

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
	amqp "github.com/rabbitmq/amqp091-go"
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
	amqpChan       *amqp.Channel
}

func NewGRPCServer(
    repo ports.PostRepository, 
    service *services.PostService, 
    minio *minio.Client, 
    presignClient *minio.Client, 
    bucketName string, 
    publicEndPoint string, 
    userClient userPb.UserServiceClient,
    amqpChan *amqp.Channel, 
) *Server {
	return &Server{
		repo:           repo,
		service:        service,
		minio:          minio,
		presignClient:  presignClient,
		bucketName:     bucketName,
		publicEndPoint: publicEndPoint,
		userClient:     userClient,
        amqpChan:       amqpChan, 
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
    if _, err := uuid.Parse(req.GetUserId()); err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "Invalid User ID")
    }

    newPost, err := s.service.CreatePost(ctx, req)
    if err != nil {
        log.Printf("Service failed to create post: %v", err)
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

func (s *Server) SearchHashtags(ctx context.Context, req *pb.SearchHashtagsRequest) (*pb.SearchHashtagsResponse, error) {
    results, err := s.service.SearchHashtags(ctx, req.Query)
    if err != nil {
        return nil, status.Error(codes.Internal, "Failed to search hashtags")
    }

    var pbResults []*pb.HashtagResult
    for _, r := range results {
        pbResults = append(pbResults, &pb.HashtagResult{
            Name:  r.Name,
            Count: r.Count,
        })
    }

    return &pb.SearchHashtagsResponse{Hashtags: pbResults}, nil
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
	expiry := time.Hour * 1
	for _, c := range collections {
		var covers []string
		for _, sp := range c.SavedPosts {
			if len(sp.Post.Media) > 0 {
				objectName := sp.Post.Media[0].MediaObjectName
				mediaType := sp.Post.Media[0].MediaType

				reqParams := make(url.Values)
				reqParams.Set("response-content-type", mediaType)

				presignedURL, err := h.presignClient.PresignedGetObject(ctx, h.bucketName, objectName, expiry, reqParams)
				
				if err == nil {
					finalURL := strings.Replace(presignedURL.String(), "minio:9000", "localhost:9000", 1)
					finalURL = strings.Replace(finalURL, "http://backend:9000", "http://localhost:9000", 1)
					
					covers = append(covers, finalURL)
				}
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
            
            presignedURL, err := s.presignClient.PresignedGetObject(ctx, s.bucketName, m.MediaObjectName, expiry, reqParams)
            
            mediaURLString := ""
            if err == nil {
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

func (s *Server) GetExplorePosts(ctx context.Context, req *pb.GetExplorePostsRequest) (*pb.GetExplorePostsResponse, error) {
    posts, err := s.service.GetExplorePosts(ctx, int(req.Limit), int(req.Offset), req.Hashtag)
    if err != nil {
        log.Printf("Failed to fetch explore posts: %v", err)
        return nil, status.Error(codes.Internal, "Failed to fetch explore posts")
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
            if err == nil {
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

func (s *Server) GetUserReels(ctx context.Context, req *pb.GetUserReelsRequest) (*pb.GetPostsResponse, error) {
    if req.GetUserId() == "" {
        return nil, status.Error(codes.InvalidArgument, "User ID is required")
    }

    // Call the service we created in Step 2
    posts, err := s.service.GetUserReels(ctx, req.GetUserId())
    if err != nil {
        log.Printf("Failed to fetch user reels: %v", err)
        return nil, status.Error(codes.Internal, "Failed to fetch user reels")
    }

    // Map Domain Posts to Proto Response (with Presigned URLs)
    var pbPosts []*pb.PostResponse
    expiry := time.Hour * 1

    for _, post := range posts {
        var pbMedia []*pb.PostMediaResponse
        
        for _, m := range post.Media {
            reqParams := make(url.Values)
            reqParams.Set("response-content-type", m.MediaType)
            
            // Generate MinIO Presigned URL
            presignedURL, err := s.presignClient.PresignedGetObject(ctx, s.bucketName, m.MediaObjectName, expiry, reqParams)
            
            mediaURLString := ""
            if err == nil {
                // Docker networking fix for browser (localhost)
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

    return &pb.GetPostsResponse{Posts: pbPosts}, nil
}

func (s *Server) GetCollectionPosts(ctx context.Context, req *pb.GetCollectionPostsRequest) (*pb.GetCollectionPostsResponse, error) {
    posts, err := s.service.GetCollectionPosts(ctx, req.CollectionId, int(req.Limit), int(req.Offset))
    if err != nil {
        log.Printf("Failed to get collection posts: %v", err)
        return nil, status.Error(codes.Internal, "Failed to fetch collection posts")
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
            if err == nil {
                // Docker networking fix
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

    return &pb.GetCollectionPostsResponse{Posts: pbPosts}, nil
}

func (s *Server) UpdateCollection(ctx context.Context, req *pb.UpdateCollectionRequest) (*pb.CollectionResponse, error) {
    col, err := s.service.UpdateCollection(ctx, req.CollectionId, req.Name, req.UserId)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    return &pb.CollectionResponse{
        Id:     col.ID.String(),
        Name:   col.Name,
        UserId: col.UserID.String(),
    }, nil
}

func (s *Server) DeleteCollection(ctx context.Context, req *pb.DeleteCollectionRequest) (*pb.DeleteCollectionResponse, error) {
    err := s.service.DeleteCollection(ctx, req.CollectionId, req.UserId)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    return &pb.DeleteCollectionResponse{Success: true, Message: "Collection deleted"}, nil
}

func (s *Server) GetPostByID(ctx context.Context, req *pb.GetPostByIDRequest) (*pb.PostResponse, error) {
    // 1. Fetch from Repo
    post, err := s.repo.GetPostByID(ctx, req.GetPostId())
    if err != nil {
        log.Printf("Failed to fetch post %s: %v", req.GetPostId(), err)
        return nil, status.Error(codes.NotFound, "Post not found")
    }

	isLiked := false
    if req.UserId != "" {
        isLiked, _ = s.repo.IsPostLikedByUser(ctx, req.GetPostId(), req.UserId)
    }

    // 2. Generate Presigned URLs for Media
    var pbMedia []*pb.PostMediaResponse
    expiry := time.Hour * 1

    for _, m := range post.Media {
        reqParams := make(url.Values)
        reqParams.Set("response-content-type", m.MediaType)

        presignedURL, err := s.presignClient.PresignedGetObject(ctx, s.bucketName, m.MediaObjectName, expiry, reqParams)
        
        mediaURLString := ""
        if err == nil {
            mediaURLString = strings.Replace(presignedURL.String(), "minio:9000", "localhost:9000", 1)
            mediaURLString = strings.Replace(mediaURLString, "http://backend:9000", "http://localhost:9000", 1)
        }

        pbMedia = append(pbMedia, &pb.PostMediaResponse{
            MediaUrl:  mediaURLString,
            MediaType: m.MediaType,
        })
    }

    return &pb.PostResponse{
        Id:            post.ID.String(),
        UserId:        post.UserID.String(),
        Media:         pbMedia,
        Caption:       post.Caption,
        Location:      post.Location,
        CreatedAt:     post.CreatedAt.Format(time.RFC3339),
        LikesCount:    post.LikesCount,
        CommentsCount: post.CommentsCount,
        IsLiked:       isLiked,
        IsReel:        post.IsReel,
    }, nil
}

func (s *Server) GetPostReports(ctx context.Context, req *pb.Empty) (*pb.PostReportListResponse, error) {
    reports, err := s.repo.GetPendingPostReports(ctx)
    if err != nil {
        log.Printf("Failed to fetch reports: %v", err)
        return nil, status.Error(codes.Internal, "Failed to fetch reports")
    }

    var pbReports []*pb.PostReportItem
    for _, r := range reports {
        pbReports = append(pbReports, &pb.PostReportItem{
            Id:             r.ID.String(),
            ReporterId:     r.ReporterID.String(),
            PostId:         r.PostID.String(),
            Reason:         r.Reason,
            Status:         r.Status,
            CreatedAt:      r.CreatedAt.Format(time.RFC3339),
        })
    }

    return &pb.PostReportListResponse{Reports: pbReports}, nil
}

func (s *Server) ReviewPostReport(ctx context.Context, req *pb.ReviewReportRequest) (*pb.Response, error) {
    // 1. Fetch Report to get PostID
    report, err := s.repo.GetPostReportByID(ctx, req.ReportId)
    if err != nil {
        return nil, status.Error(codes.NotFound, "Report not found")
    }

    statusStr := "REJECTED"
    
    // 2. Handle Actions
    if req.Action == "DELETE_POST" {
        statusStr = "RESOLVED"
        if err := s.repo.DeletePost(ctx, report.PostID.String()); err != nil {
            log.Printf("Failed to delete post %s: %v", report.PostID, err)
            return nil, status.Error(codes.Internal, "Failed to delete post")
        }
    }

    // 3. Update Report Status
    if err := s.repo.UpdatePostReportStatus(ctx, req.ReportId, statusStr); err != nil {
        return nil, status.Error(codes.Internal, "Failed to update report status")
    }

	if req.Action == "DELETE_POST" {
        // 1. Get Reporter Email
        emailRes, err := s.userClient.GetUserEmail(ctx, &userPb.GetUserEmailRequest{UserId: report.ReporterID.String()})
        if err == nil && emailRes.Email != "" {
            // 2. Publish Email Task
            type EmailTask struct {
                Email   string `json:"email"`
                Subject string `json:"subject"`
                Body    string `json:"body"`
            }
            emailBody := "We have reviewed your report and removed the content that violated our guidelines."
            task := EmailTask{Email: emailRes.Email, Subject: "Report Action Taken", Body: emailBody}
            taskBody, _ := json.Marshal(task)

			err = s.amqpChan.PublishWithContext(ctx, 
                "email_exchange", 
                "send_email", 
                false, 
                false, 
                amqp.Publishing{
                    ContentType: "application/json", 
                    Body: taskBody,
                })
            
            if err != nil {
                log.Printf("Failed to publish email task: %v", err)
            }
        }
    }

	if err := s.repo.UpdatePostReportStatus(ctx, req.ReportId, statusStr); err != nil {
        return nil, status.Error(codes.Internal, "Failed to update report status")
    }
	
    return &pb.Response{Success: true, Message: "Report processed"}, nil
}

func (s *Server) ReportPost(ctx context.Context, req *pb.ReportPostRequest) (*pb.Response, error) {
    report := &domain.PostReport{
        PostID:     uuid.MustParse(req.PostId),
        ReporterID: uuid.MustParse(req.UserId),
        Reason:     req.Reason,
        Status:     "PENDING",
    }

    if err := s.repo.CreatePostReport(report); err != nil {
        return nil, status.Error(codes.Internal, "Failed to report post")
    }
    return &pb.Response{Success: true, Message: "Post reported"}, nil
}

func (s *Server) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostResponse, error) {
    err := s.service.DeletePost(ctx, req.GetPostId(), req.GetUserId())
    if err != nil {
        log.Printf("[ERROR] DeletePost failed: %v", err) // This prints to docker logs

        // Check for specific errors
        if strings.Contains(err.Error(), "unauthorized") {
            return nil, status.Error(codes.PermissionDenied, "You are not authorized to delete this post")
        }
        if strings.Contains(err.Error(), "record not found") || strings.Contains(err.Error(), "not found") {
            return nil, status.Error(codes.NotFound, "Post not found")
        }
        
        // Return the actual error string during development so you see it in the frontend/Postman
        return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to delete post: %v", err))
    }

    return &pb.DeletePostResponse{
        Success: true, 
        Message: "Post deleted successfully",
    }, nil
}