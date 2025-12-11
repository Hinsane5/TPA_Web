package http

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Hinsane5/hoshiBmaTchi/backend/services/chat/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/chat/internal/repositories"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/chat/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type CreateGroupRequest struct {
	Name   string   `json:"name" binding:"required"`
	UserIDs []string `json:"user_ids" binding:"required"`
}

type ParticipantRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type ChatHandler struct {
	Repo *repositories.ChatRepository
	Hub  *ws.Hub
}

type ShareContentRequest struct {
	RecipientID string `json:"recipient_id" binding:"required"`
	ContentID   string `json:"content_id" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Thumbnail   string `json:"thumbnail"`
}

func NewChatHandler(repo *repositories.ChatRepository, hub *ws.Hub) *ChatHandler {
	return &ChatHandler{Repo: repo, Hub: hub}
}

func (h *ChatHandler) RegisterRoutes(r *gin.Engine){
	r.GET("/ws", h.ServeWS)

	chatGroup := r.Group("/chats")
	{
		chatGroup.GET("", h.GetConversations)
		chatGroup.POST("", h.CreateGroupChat) 
		chatGroup.POST("/share", h.ShareContent)

		chatGroup.GET("/:id/messages", h.GetMessageHistory)
		chatGroup.GET("/search", h.SearchMessages) 

		chatGroup.POST("/:id/participants", h.AddParticipant)
		chatGroup.DELETE("/:id/participants", h.RemoveParticipant)

		chatGroup.POST("/upload", h.UploadMedia) 
        chatGroup.DELETE("/:id", h.DeleteConversation)
	}
}

func (h *ChatHandler) ServeWS(c *gin.Context) {

	userID := c.GetHeader("X-User-ID")
	
	if userID == "" {
		userID = c.Query("user_id")
	}

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: User ID missing"})
		return
	}

	ws.ServeWs(h.Hub, h.Repo, c.Writer, c.Request, userID)
}

func (h *ChatHandler) GetConversations(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	convs, err := h.Repo.GetConversations(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversations"})
		return
	}

	c.JSON(http.StatusOK, convs)
}

func (h *ChatHandler) CreateGroupChat(c *gin.Context) {
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if len(req.UserIDs) == 1 {
		targetUserID := req.UserIDs[0]
		
		existingConv, err := h.Repo.FindDirectConversation(c, userID, targetUserID)
		if err == nil && existingConv != nil {
			c.JSON(http.StatusOK, gin.H{
				"conversation_id": existingConv.ID.String(),
				"message":         "Chat already exists",
				"is_existing":     true,
			})
			return
		}
	}

	userIDs := append(req.UserIDs, userID)

	conv := &domain.Conversation{
		Name:      req.Name,
		IsGroup:   true,
		CreatedAt: time.Now(),
	}

	if err := h.Repo.CreateConversation(c, conv, userIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"conversation_id": conv.ID.String(),
		"message":         "Group created successfully",
		"is_existing":     false,
	})
}

func (h *ChatHandler) GetMessageHistory(c *gin.Context) {
	conversationID := c.Param("id")
	
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	msgs, err := h.Repo.GetMessageHistory(c, conversationID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
		return
	}

	c.JSON(http.StatusOK, msgs)
}

func (h *ChatHandler) SearchMessages(c *gin.Context) {
	conversationID := c.Query("conversation_id")
	query := c.Query("q")

	if conversationID == "" || query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id and q parameters are required"})
		return
	}

	msgs, err := h.Repo.SearchMessages(c, conversationID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(http.StatusOK, msgs)
}

func (h *ChatHandler) AddParticipant(c *gin.Context) {
	conversationID := c.Param("id")
	var req ParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Repo.AddParticipant(c, conversationID, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add participant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Participant added"})
}

func (h *ChatHandler) RemoveParticipant(c *gin.Context) {
	conversationID := c.Param("id")
	var req ParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Repo.RemoveParticipant(c, conversationID, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove participant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Participant removed"})
}

func (h *ChatHandler) UploadMedia(c *gin.Context){
	file, header, err := c.Request.FormFile("file")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	defer file.Close()

	accessKey := os.Getenv("MINIO_ROOT_USER")   
    secretKey := os.Getenv("MINIO_ROOT_PASSWORD")
	bucketName := os.Getenv("MINIO_CHAT_BUCKET_NAME")

	endpoint := "localhost:9000"
	minioClient, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: false,
    })

    objectName := uuid.New().String() + filepath.Ext(header.Filename)
    contentType := header.Header.Get("Content-Type")

	ctx := context.Background()
    if exists, _ := minioClient.BucketExists(ctx, bucketName); !exists {
        minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
    }

    _, err = minioClient.PutObject(ctx, bucketName, objectName, file, header.Size, minio.PutObjectOptions{
        ContentType: contentType,
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to storage"})
        return
    }


    fileURL := "http://" + endpoint + "/" + bucketName + "/" + objectName
    
    c.JSON(http.StatusOK, gin.H{
        "media_url": fileURL,
        "type":      getHeaderType(contentType),
    })
}

func (h *ChatHandler) DeleteConversation(c *gin.Context) {
	conversationID := c.Param("id")
	userID := c.GetHeader("X-User-ID")

	if err := h.Repo.RemoveParticipant(c, conversationID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation deleted"})
}

func getHeaderType(contentType string) string {
    if contentType == "video/mp4" { return "video" }
    return "image"
}

func (h *ChatHandler) ShareContent(c *gin.Context) {
	var req ShareContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	senderID := c.GetHeader("X-User-ID")
	if senderID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var conversationID uuid.UUID
	
	existingConv, err := h.Repo.FindDirectConversation(c, senderID, req.RecipientID)
	if err == nil && existingConv != nil {
		conversationID = existingConv.ID
	} else {
		conv := &domain.Conversation{
			Name:      "Direct Message",
			IsGroup:   false,
			CreatedAt: time.Now(),
		}
		userIDs := []string{senderID, req.RecipientID}
		if err := h.Repo.CreateConversation(c, conv, userIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation for sharing"})
			return
		}
		conversationID = conv.ID
	}

	// 2. Create Message
	sID, _ := uuid.Parse(senderID)
	
	msg := &domain.Message{
		ID:             uuid.New(),
		ConversationID: conversationID,
		SenderID:       sID,
		Content:        req.ContentID, 
		MediaType:      req.Type + "_share",
		MediaURL:       req.Thumbnail,      
		CreatedAt:      time.Now(),
	}

	if err := h.Repo.SaveMessage(c, msg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send share message"})
		return
	}

	wsMsg := map[string]interface{}{
		"type":            "new_message",
		"id":              msg.ID,
		"conversation_id": msg.ConversationID,
		"sender_id":       msg.SenderID,
		"content":         msg.Content,
		"media_url":       msg.MediaURL,
		"media_type":      msg.MediaType,
		"created_at":      msg.CreatedAt,
	}

	if msgBytes, err := json.Marshal(wsMsg); err == nil {
		h.Hub.Broadcast(msgBytes)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Content shared successfully", "conversation_id": conversationID})
}