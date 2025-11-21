package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/stories"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StoriesHandler struct {
	client pb.StoriesServiceClient
}

func NewStoriesHandler(client pb.StoriesServiceClient) *StoriesHandler {
	return &StoriesHandler{client: client}
}

func (h *StoriesHandler) CreateStory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := c.Request.ParseMultipartForm(100 << 20); err != nil { 
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	mediaURL := c.PostForm("media_url")
	mediaTypeStr := c.PostForm("media_type")
	durationStr := c.PostForm("duration")

	if mediaURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Media URL is required"})
		return
	}

	var mediaType pb.MediaType
	if mediaTypeStr == "VIDEO" {
		mediaType = pb.MediaType_VIDEO
	} else {
		mediaType = pb.MediaType_IMAGE
	}

	duration := int32(5) 
	if durationStr != "" {
		if d, err := strconv.Atoi(durationStr); err == nil {
			duration = int32(d)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := h.client.CreateStory(ctx, &pb.CreateStoryRequest{
		UserId:    userID.(string),
		MediaUrl:  mediaURL,
		MediaType: mediaType,
		Duration:  duration,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create story"})
		return
	}

	c.JSON(http.StatusCreated, resp.Story)
}

func (h *StoriesHandler) GetStory(c *gin.Context) {
	userID, exists := c.Get("user_id")
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
	viewerID, exists := c.Get("user_id")
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

func (h *StoriesHandler) GetFollowingStories(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := h.client.GetFollowingStories(ctx, &pb.GetFollowingStoriesRequest{
		UserId: userID.(string),
		Limit:  int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_stories": resp.UserStories})
}

func (h *StoriesHandler) DeleteStory(c *gin.Context) {
	userID, exists := c.Get("user_id")
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
	userID, exists := c.Get("user_id")
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
	userID, exists := c.Get("user_id")
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
	userID, exists := c.Get("user_id")
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
	userID, exists := c.Get("user_id")
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
	userID, exists := c.Get("user_id")
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
	userID, exists := c.Get("user_id")
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
	userID, exists := c.Get("user_id")
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