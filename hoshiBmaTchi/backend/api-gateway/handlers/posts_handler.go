package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"

	postsProto "github.com/Hinsane5/hoshiBmaTchi/backend/proto/posts"
	usersProto "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

type PostsHandler struct {
    postsClient postsProto.PostsServiceClient
    usersClient usersProto.UserServiceClient 
}

type createPostJSON struct {
    Media    []mediaItemJSON `json:"media" binding:"required"` 
	Caption  string          `json:"caption"`
	Location string          `json:"location"`
    IsReel          bool   `json:"is_reel"`
}

type mediaItemJSON struct {
    MediaObjectName string `json:"media_object_name" binding:"required"`
    MediaType       string `json:"media_type" binding:"required"`
}

func NewPostsHandler(postsClient postsProto.PostsServiceClient, usersClient usersProto.UserServiceClient) *PostsHandler {
    return &PostsHandler{
        postsClient: postsClient,
        usersClient: usersClient,
    }
}

type createCommentJSON struct {
	Content string `json:"content" binding:"required"`
}

type toggleSaveJSON struct {
    CollectionID string `json:"collection_id"`
}

type createCollectionJSON struct {
    Name string `json:"name" binding:"required"`
}

type UserSummary struct {
    ID             string `json:"id"`
    Username       string `json:"username"`
    ProfilePicture string `json:"profile_picture"`
}

type ReelResponse struct {
    ID            string      `json:"id"`
    Caption       string      `json:"caption"`
    MediaUrl      string      `json:"video_url"`
    ThumbnailUrl  string      `json:"thumbnail_url"`
    User          UserSummary `json:"user"`
    LikesCount    int32       `json:"likes_count"`
    CommentsCount int32       `json:"comments_count"`
    IsLiked       bool        `json:"is_liked"`
    IsSaved       bool        `json:"is_saved"`
    CreatedAt     string      `json:"created_at"`
}

func (h *PostsHandler) GenerateUploadURL (c *gin.Context){
    fileName := c.Query("file_name")
	fileType := c.Query("file_type")
	if fileName == "" || fileType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query 'file_name' and 'file_type' diperlukan"})
		return
	}

    res, err := h.postsClient.GenerateUploadURL(context.Background(), &postsProto.GenerateUploadURLRequest{
        FileName: fileName,
        FileType: fileType,
    })

    if err != nil {
        if s, ok := status.FromError(err); ok {
            c.JSON(http.StatusInternalServerError, gin.H{"error": s.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to called gRPC: " + err.Error()})
		}
		return
    }

    c.JSON(http.StatusOK, gin.H {
        "upload_url" : res.UploadUrl,
        "object_name": res.ObjectName,
    })
}

func (h *PostsHandler) CreatePost(c *gin.Context){
    var jsonReq createPostJSON
    if err := c.ShouldBindJSON(&jsonReq); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error" : err.Error()})
        return
    }

    userID, exists := c.Get("userID")
    if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "UserID not found in token"})
		return
	}

    var protoMedia []*postsProto.PostMediaItem
    for _, m := range jsonReq.Media {
        protoMedia = append(protoMedia, &postsProto.PostMediaItem{
            MediaObjectName: m.MediaObjectName,
            MediaType:       m.MediaType,
        })
    }

    res, err := h.postsClient.CreatePost(context.Background(), &postsProto.CreatePostRequest{
		UserId:   userID.(string),
		Media:    protoMedia, 
		Caption:  jsonReq.Caption,
		Location: jsonReq.Location,
        IsReel:    jsonReq.IsReel,
	})

    if err != nil {
		if s, ok := status.FromError(err); ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": s.Message()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call gRPC: " + err.Error()})
		}
		return
	}

    c.JSON(http.StatusCreated, res.Post)
}

func (h *PostsHandler) GetPostsByUserID (c *gin.Context){
    userID := c.Param("userID") 
    if userID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error" : "Parameter userID diperlukan"})
        return
    }

    res, err := h.postsClient.GetPostsByUserID(context.Background(), &postsProto.GetPostsByUserIDRequest{
        UserId: userID,
    })

    if err != nil {
        if s, oke := status.FromError(err); oke {
            c.JSON(http.StatusInternalServerError, gin.H{"error": s.Message()})
        } else{
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to called gRPC: " + err.Error()})
        }
        return
    }

    c.JSON(http.StatusOK, res.Posts)
}

func (h *PostsHandler) LikePost(c *gin.Context) {
    postID := c.Param("postID")
    userID, _ := c.Get("userID")

    res, err := h.postsClient.LikePost(context.Background(), &postsProto.LikePostRequest{
        UserId: userID.(string),
        PostId: postID,
    })
    if err != nil {
        if s, ok := status.FromError(err); ok {
            c.JSON(http.StatusInternalServerError, gin.H{"error": s.Message()})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memanggil gRPC: " + err.Error()})
        }
        return
    }
    c.JSON(http.StatusCreated, res)
}

func (h *PostsHandler) UnlikePost(c *gin.Context) {
    postID := c.Param("postID")
    userID, _ := c.Get("userID")

    res, err := h.postsClient.UnlikePost(context.Background(), &postsProto.UnlikePostRequest{
        UserId: userID.(string),
        PostId: postID,
    })
    if err != nil {
        if s, ok := status.FromError(err); ok {
            c.JSON(http.StatusInternalServerError, gin.H{"error": s.Message()})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memanggil gRPC: " + err.Error()})
        }
        return
    }
    c.JSON(http.StatusOK, res)
}

func (h *PostsHandler) CreateComment(c *gin.Context) {
    postID := c.Param("postID")
    userID, _ := c.Get("userID")

    var jsonReq createCommentJSON
	if err := c.ShouldBindJSON(&jsonReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

    res, err := h.postsClient.CreateComment(context.Background(), &postsProto.CreateCommentRequest{
        UserId:  userID.(string),
        PostId:  postID,
        Content: jsonReq.Content,
    })
    if err != nil {
        if s, ok := status.FromError(err); ok {
            c.JSON(http.StatusInternalServerError, gin.H{"error": s.Message()})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memanggil gRPC: " + err.Error()})
        }
        return
    }
    c.JSON(http.StatusCreated, res)
}

func (h *PostsHandler) GetCommentsForPost(c *gin.Context) {
    postID := c.Param("postID")

    res, err := h.postsClient.GetCommentsForPost(context.Background(), &postsProto.GetCommentsForPostRequest{
        PostId: postID,
    })
    if err != nil {
        if s, ok := status.FromError(err); ok {
            c.JSON(http.StatusInternalServerError, gin.H{"error": s.Message()})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal memanggil gRPC: " + err.Error()})
        }
        return
    }
    c.JSON(http.StatusOK, res.Comments)
}

func (h *PostsHandler) GetHomeFeed(c *gin.Context) {
    userID, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    limitStr := c.DefaultQuery("limit", "10")
    offsetStr := c.DefaultQuery("offset", "0")
    limit, _ := strconv.Atoi(limitStr)
    offset, _ := strconv.Atoi(offsetStr)

    res, err := h.postsClient.GetHomeFeed(context.Background(), &postsProto.GetHomeFeedRequest{
        UserId: userID.(string),
        Limit:  int32(limit),
        Offset: int32(offset),
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch feed: " + err.Error()})
        return
    }

    var enrichedPosts []gin.H

    for _, post := range res.Posts {
        userRes, err := h.usersClient.GetUserProfile(context.Background(), &usersProto.GetUserProfileRequest{
            UserId: post.UserId,
        })

        username := "Unknown"
        profilePic := ""
        
        if err == nil {
            username = userRes.Username
            profilePic = userRes.ProfilePictureUrl
        }


        var mediaList []gin.H
        for _, m := range post.Media {
            mediaList = append(mediaList, gin.H{
                "media_url":  m.MediaUrl,
                "media_type": m.MediaType,
            })
        }

        enrichedPosts = append(enrichedPosts, gin.H{
            "id":              post.Id,
            "user_id":         post.UserId,
            "media":           mediaList, 
            "caption":         post.Caption,
            "location":        post.Location,
            "created_at":      post.CreatedAt,
            "username":        username,          
            "profile_picture": profilePic,        
            "likes_count":     post.LikesCount,
            "comments_count":  post.CommentsCount,
            "is_liked":        post.IsLiked,
        })
    }

    c.JSON(http.StatusOK, gin.H{"data": enrichedPosts})
}

func (h *PostsHandler) ToggleSavePost(c *gin.Context) {
    postID := c.Param("postID")
    userID, _ := c.Get("userID")
    
    var req toggleSaveJSON
    c.ShouldBindJSON(&req)

    res, err := h.postsClient.ToggleSavePost(context.Background(), &postsProto.ToggleSavePostRequest{
        UserId:       userID.(string),
        PostId:       postID,
        CollectionId: req.CollectionID,
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, res)
}

func (h *PostsHandler) CreateCollection(c *gin.Context) {
    userID, _ := c.Get("userID")
    var req createCollectionJSON
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    res, err := h.postsClient.CreateCollection(context.Background(), &postsProto.CreateCollectionRequest{
        UserId: userID.(string),
        Name:   req.Name,
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, res)
}

func (h *PostsHandler) GetUserCollections(c *gin.Context) {
    userID, _ := c.Get("userID")

    res, err := h.postsClient.GetUserCollections(context.Background(), &postsProto.GetUserCollectionsRequest{
        UserId: userID.(string),
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, res.Collections)
}

func (h *PostsHandler) GetUserMentions(c *gin.Context) {
    targetID := c.Param("target_id")
    log.Printf("[DEBUG-GATEWAY] GetUserMentions Request. TargetID: %s", targetID)
    
    // Pagination
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

    // Call gRPC Service
    res, err := h.postsClient.GetUserMentions(context.Background(), &postsProto.GetUserMentionsRequest{
        TargetUserId: targetID,
        Limit:        int32(limit),
        Offset:       int32(offset),
    })
    
    if err != nil {
        if s, ok := status.FromError(err); ok {
            c.JSON(http.StatusInternalServerError, gin.H{"error": s.Message()})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    // Enrich Data (similar to GetHomeFeed)
    var enrichedPosts []gin.H

    for _, post := range res.Posts {
        userRes, err := h.usersClient.GetUserProfile(context.Background(), &usersProto.GetUserProfileRequest{
            UserId: post.UserId,
        })
        
        username := "Unknown"
        profilePic := ""
        if err == nil {
            username = userRes.Username
            profilePic = userRes.ProfilePictureUrl
        }

        var mediaList []gin.H
        for _, m := range post.Media {
            mediaList = append(mediaList, gin.H{
                "media_url":  m.MediaUrl,
                "media_type": m.MediaType,
            })
        }

        enrichedPosts = append(enrichedPosts, gin.H{
            "id":              post.Id,
            "user_id":         post.UserId,
            "username":        username,
            "profile_picture": profilePic,
            "media":           mediaList,
            "caption":         post.Caption,
            "likes_count":     post.LikesCount,
            "comments_count":  post.CommentsCount,
            "created_at":      post.CreatedAt,
            "is_liked":        post.IsLiked,
        })
    }
    
    // Return standard "data" wrapper
    c.JSON(http.StatusOK, gin.H{"data": enrichedPosts})
}

func (h *PostsHandler) GetReelsFeed(c *gin.Context) {
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

    // 1. Call gRPC to get raw reels data
    res, err := h.postsClient.GetReels(context.Background(), &postsProto.GetReelsRequest{
        Limit:  int32(limit),
        Offset: int32(offset),
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reels: " + err.Error()})
        return
    }

    var reels []ReelResponse

    for _, post := range res.Posts {
        userRes, err := h.usersClient.GetUserProfile(context.Background(), &usersProto.GetUserProfileRequest{
            UserId: post.UserId,
        })

        userSummary := UserSummary{
            ID:       post.UserId,
            Username: "Unknown",
        }

        if err == nil {
            userSummary.Username = userRes.Username
        }

        // Get the first media item as the main video
        mainVideo := ""
        if len(post.Media) > 0 {
            mainVideo = post.Media[0].MediaUrl
        }

        reels = append(reels, ReelResponse{
            ID:            post.Id,
            Caption:       post.Caption,
            MediaUrl:      mainVideo,
            ThumbnailUrl:  mainVideo,
            User:          userSummary,
            LikesCount:    post.LikesCount,
            CommentsCount: post.CommentsCount,
            IsLiked:       post.IsLiked,
            CreatedAt:     post.CreatedAt,
            // IsSaved:    post.IsSaved, // Add this to proto if needed later
        })
    }

    c.JSON(http.StatusOK, gin.H{"data": reels})
}

func (h *PostsHandler) GetExplorePosts(c *gin.Context) {
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
    hashtag := c.Query("hashtag") 

    res, err := h.postsClient.GetExplorePosts(context.Background(), &postsProto.GetExplorePostsRequest{
        Limit:   int32(limit),
        Offset:  int32(offset),
        Hashtag: hashtag,
        UserId:  "",
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    enrichedPosts := []gin.H{}
    for _, post := range res.Posts {
        userRes, err := h.usersClient.GetUserProfile(context.Background(), &usersProto.GetUserProfileRequest{
            UserId: post.UserId,
        })
        
        username := "Unknown"
        profilePic := ""
        if err == nil {
            username = userRes.Username
            profilePic = userRes.ProfilePictureUrl
        }

        var mediaList []gin.H
        for _, m := range post.Media {
            mediaList = append(mediaList, gin.H{
                "media_url":  m.MediaUrl,
                "media_type": m.MediaType,
            })
        }

        enrichedPosts = append(enrichedPosts, gin.H{
            "id":              post.Id,
            "user_id":         post.UserId,
            "username":        username,
            "profile_picture": profilePic,
            "media":           mediaList,
            "caption":         post.Caption,
            "likes_count":     post.LikesCount,
            "comments_count":  post.CommentsCount,
            "is_reel":         post.IsReel,
            "created_at":      post.CreatedAt,
            "is_liked":        post.IsLiked,
        })
    }

    c.JSON(http.StatusOK, gin.H{"data": enrichedPosts})
}

func (h *PostsHandler) GetUserReels(c *gin.Context) {
    targetUserID := c.Param("userID")
    if targetUserID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
        return
    }

    res, err := h.postsClient.GetUserReels(context.Background(), &postsProto.GetUserReelsRequest{
        UserId: targetUserID,
    })

    if err != nil {
        if s, ok := status.FromError(err); ok {
            c.JSON(http.StatusInternalServerError, gin.H{"error": s.Message()})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call gRPC: " + err.Error()})
        }
        return
    }

    c.JSON(http.StatusOK, res.Posts)
}

func (h *PostsHandler) GetCollectionPosts(c *gin.Context) {
    collectionID := c.Param("collectionID")
    userID, _ := c.Get("userID")
    
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

    res, err := h.postsClient.GetCollectionPosts(context.Background(), &postsProto.GetCollectionPostsRequest{
        CollectionId: collectionID,
        UserId:       userID.(string),
        Limit:        int32(limit),
        Offset:       int32(offset),
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // Reuse logic to enrich with user profiles if necessary, or return raw
    // For brevity, returning enriched structure similar to Explore
    enrichedPosts := []gin.H{}
    for _, post := range res.Posts {
         // (Optional: fetch user profile for each post author if needed)
         var mediaList []gin.H
        for _, m := range post.Media {
            mediaList = append(mediaList, gin.H{
                "media_url":  m.MediaUrl,
                "media_type": m.MediaType,
            })
        }
        
        enrichedPosts = append(enrichedPosts, gin.H{
            "id":              post.Id,
            "user_id":         post.UserId,
            "media":           mediaList,
            "caption":         post.Caption,
            "likes_count":     post.LikesCount,
            "comments_count":  post.CommentsCount,
            "created_at":      post.CreatedAt,
            "is_liked":        post.IsLiked,
             // Add username/profile_picture here if fetched from UsersService
        })
    }
    
    c.JSON(http.StatusOK, gin.H{"data": enrichedPosts})
}

func (h *PostsHandler) UpdateCollection(c *gin.Context) {
    collectionID := c.Param("collectionID")
    userID, _ := c.Get("userID")
    var req struct {
        Name string `json:"name" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    res, err := h.postsClient.UpdateCollection(context.Background(), &postsProto.UpdateCollectionRequest{
        CollectionId: collectionID,
        UserId:       userID.(string),
        Name:         req.Name,
    })
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, res)
}

func (h *PostsHandler) DeleteCollection(c *gin.Context) {
    collectionID := c.Param("collectionID")
    userID, _ := c.Get("userID")

    _, err := h.postsClient.DeleteCollection(context.Background(), &postsProto.DeleteCollectionRequest{
        CollectionId: collectionID,
        UserId:       userID.(string),
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Collection deleted"})
}

func (h *PostsHandler) GetPostByID(c *gin.Context) {
    postID := c.Param("postID")
    if postID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Post ID is required"})
        return
    }

    userID := ""
    if val, exists := c.Get("userID"); exists {
        userID = val.(string)
    }

    // 1. Get Post Data
    res, err := h.postsClient.GetPostByID(context.Background(), &postsProto.GetPostByIDRequest{
        PostId: postID,
        UserId: userID,
    })

    if err != nil {
        if s, ok := status.FromError(err); ok {
            c.JSON(http.StatusInternalServerError, gin.H{"error": s.Message()})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch post: " + err.Error()})
        }
        return
    }

    // 2. Get Author Data
    username := "Unknown"
    profilePic := ""
    
    userRes, err := h.usersClient.GetUserProfile(context.Background(), &usersProto.GetUserProfileRequest{
        UserId: res.UserId,
    })
    if err == nil {
        username = userRes.Username
        profilePic = userRes.ProfilePictureUrl
    }

    // 3. Format Response
    var mediaList []gin.H
    for _, m := range res.Media {
        mediaList = append(mediaList, gin.H{
            "media_url":  m.MediaUrl,
            "media_type": m.MediaType,
        })
    }

    c.JSON(http.StatusOK, gin.H{
        "id":              res.Id,
        "user_id":         res.UserId,
        "username":        username,
        "profile_picture": profilePic,
        "media":           mediaList,
        "caption":         res.Caption,
        "location":        res.Location,
        "created_at":      res.CreatedAt,
        "likes_count":     res.LikesCount,
        "comments_count":  res.CommentsCount,
        "is_liked":        res.IsLiked,
        "is_reel":         res.IsReel,
    })
}

func (h *PostsHandler) ReportPost(c *gin.Context) {
    postID := c.Param("id")
    var req struct {
        Reason string `json:"reason" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Reason is required"})
        return
    }
    userID := c.GetString("userID")

    _, err := h.postsClient.ReportPost(context.Background(), &postsProto.ReportPostRequest{
        PostId: postID,
        UserId: userID,
        Reason: req.Reason,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Post reported"})
}