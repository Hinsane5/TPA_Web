package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/stories"
	usersPb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StoriesHandler struct {
	client     pb.StoriesServiceClient
	userClient usersPb.UserServiceClient
}

func NewStoriesHandler(client pb.StoriesServiceClient, userClient usersPb.UserServiceClient) *StoriesHandler {
	return &StoriesHandler{
		client:     client,
		userClient: userClient,
	}
}

func (h *StoriesHandler) CreateStory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		MediaObjectName string `json:"media_object_name" binding:"required"`
		MediaType       string `json:"media_type" binding:"required"`
		Duration        int32  `json:"duration"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var mediaType pb.MediaType
	if req.MediaType == "VIDEO" {
		mediaType = pb.MediaType_VIDEO
	} else {
		mediaType = pb.MediaType_IMAGE
	}

	if req.Duration == 0 {
		req.Duration = 5
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.CreateStory(ctx, &pb.CreateStoryRequest{
		UserId:    userID.(string),
		MediaUrl:  req.MediaObjectName,
		MediaType: mediaType,
		Duration:  req.Duration,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create story"})
		return
	}

	c.JSON(http.StatusCreated, resp.Story)
}

func (h *StoriesHandler) GenerateUploadURL(c *gin.Context) {
	fileName := c.Query("file_name")
	fileType := c.Query("file_type")

	if fileName == "" || fileType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_name and file_type are required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GenerateUploadURL(ctx, &pb.GenerateUploadURLRequest{
		FileName: fileName,
		FileType: fileType,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate upload URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"upload_url":  resp.UploadUrl,
		"object_name": resp.ObjectName,
	})
}

func (h *StoriesHandler) GetStory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	storyID := c.Query("id")
	if storyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Story ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetStory(ctx, &pb.GetStoryRequest{
		StoryId: storyID,
		UserId:  userID.(string),
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Story not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get story"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *StoriesHandler) GetUserStories(c *gin.Context) {
	viewerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	targetUserID := c.Query("user_id")
	if targetUserID == "" {
		targetUserID = viewerID.(string)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetUserStories(ctx, &pb.GetUserStoriesRequest{
		UserId:   targetUserID,
		ViewerId: viewerID.(string),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stories": resp.Stories})
}

func (h *StoriesHandler) DeleteStory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	storyID := c.Query("id")
	if storyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Story ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := h.client.DeleteStory(ctx, &pb.DeleteStoryRequest{
		StoryId: storyID,
		UserId:  userID.(string),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete story"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *StoriesHandler) ViewStory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		StoryID string `json:"story_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := h.client.ViewStory(ctx, &pb.ViewStoryRequest{
		StoryId: req.StoryID,
		UserId:  userID.(string),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to view story"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *StoriesHandler) LikeStory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		StoryID string `json:"story_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := h.client.LikeStory(ctx, &pb.LikeStoryRequest{
		StoryId: req.StoryID,
		UserId:  userID.(string),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like story"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *StoriesHandler) UnlikeStory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	storyID := c.Query("story_id")
	if storyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Story ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := h.client.UnlikeStory(ctx, &pb.UnlikeStoryRequest{
		StoryId: storyID,
		UserId:  userID.(string),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlike story"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *StoriesHandler) ReplyToStory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		StoryID string `json:"story_id" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.ReplyToStory(ctx, &pb.ReplyToStoryRequest{
		StoryId: req.StoryID,
		UserId:  userID.(string),
		Content: req.Content,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reply to story"})
		return
	}

	c.JSON(http.StatusCreated, resp.Reply)
}

func (h *StoriesHandler) GetStoryReplies(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	storyID := c.Query("story_id")
	if storyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Story ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetStoryReplies(ctx, &pb.GetStoryRepliesRequest{
		StoryId: storyID,
		UserId:  userID.(string),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get replies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"replies": resp.Replies})
}

func (h *StoriesHandler) GetStoryViewers(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	storyID := c.Query("story_id")
	if storyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Story ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetStoryViewers(ctx, &pb.GetStoryViewersRequest{
		StoryId: storyID,
		UserId:  userID.(string),
	})
	if err != nil {
		if status.Code(err) == codes.PermissionDenied {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get viewers"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"viewers": resp.Viewers})
}

func (h *StoriesHandler) ShareStory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		StoryID     string `json:"story_id" binding:"required"`
		RecipientID string `json:"recipient_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.ShareStory(ctx, &pb.ShareStoryRequest{
		StoryId:     req.StoryID,
		SenderId:    userID.(string),
		RecipientId: req.RecipientID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share story"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *StoriesHandler) GetFollowingStories(c *gin.Context) {

	fmt.Println("--- DEBUG: GetFollowingStories Hit ---")
    keys := c.Keys
    fmt.Printf("Available Keys in Context: %v\n", keys)

    userID, exists := c.Get("userID") 
    fmt.Printf("Trying to get 'userID': exists=%v, value=%v\n", exists, userID)

    if !exists {
        fmt.Println("--- DEBUG: Unauthorized (Key missing) ---")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	followedResp, err := h.userClient.GetFollowingList(ctx, &usersPb.GetFollowingListRequest{
		UserId: userID.(string),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch followed users"})
		return
	}

	authorIDs := followedResp.FollowingIds

	if len(authorIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"user_stories": []interface{}{}})
		return
	}

	storiesResp, err := h.client.GetStoriesByAuthors(ctx, &pb.GetStoriesByAuthorsRequest{
		AuthorIds: authorIDs,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stories"})
		return
	}

	uniqueUserIDs := make(map[string]bool)
	for _, s := range storiesResp.Stories {
		uniqueUserIDs[s.UserId] = true
	}

	usersMap := make(map[string]*usersPb.GetUserProfileResponse)

	for uid := range uniqueUserIDs {
		uResp, err := h.userClient.GetUserProfile(ctx, &usersPb.GetUserProfileRequest{
			UserId:   uid,
			ViewerId: userID.(string),
		})
		if err == nil {
			usersMap[uid] = uResp
		}
	}

	var enrichedStories []gin.H

	for _, story := range storiesResp.Stories {
		user, found := usersMap[story.UserId]

		username := "Unknown"
		avatar := ""
		uid := story.UserId

		if found {
			username = user.Username
			avatar = user.ProfilePictureUrl
			uid = user.Id
		}

		storyData := gin.H{
			"id":          story.Id,
			"user_id":     story.UserId,
			"media_url":   story.MediaUrl,
			"media_type":  story.MediaType.String(), 
			"duration":    story.Duration,
			"created_at":  story.CreatedAt.AsTime().Format(time.RFC3339),
			"expires_at":  story.ExpiresAt.AsTime().Format(time.RFC3339),
			"is_viewed":   false, 
			"is_liked":    false, 
			"likes_count": 0,
			"user": gin.H{
				"id":         uid,
				"username":   username,
				"userAvatar": avatar,
			},
		}

		enrichedStories = append(enrichedStories, storyData)
	}

	c.JSON(http.StatusOK, gin.H{"user_stories": enrichedStories})
}