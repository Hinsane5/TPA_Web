package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/chat"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/chat/internal/core/domain"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/chat/internal/repositories"
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/chat/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/AgoraIO/Tools/DynamicKey/AgoraDynamicKey/go/src/rtctokenbuilder2"
)

type CreateGroupRequest struct {
	Name   string   `json:"name"`
	UserIDs []string `json:"user_ids" binding:"required"`
}

type ParticipantRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type ChatHandler struct {
	Repo *repositories.ChatRepository
	Hub  *ws.Hub
	client pb.ChatServiceClient
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

		chatGroup.GET("/:id/call-token", h.GenerateCallToken)

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

	allUserIDs := append(req.UserIDs, userID)
    
    groupName := req.Name
    if groupName == "" && len(req.UserIDs) > 1 {
        groupName = "New Group"
    }

	conv := &domain.Conversation{
		ID:        uuid.New(),
		Name:      groupName,
		IsGroup:   len(req.UserIDs) > 1, // Logic: >1 target user means group
		CreatedAt: time.Now(),
	}

	if err := h.Repo.CreateConversation(c, conv, allUserIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	wsMsg := map[string]interface{}{
		"type":            "group_created",
		"conversation_id": conv.ID.String(),
		"name":            conv.Name,
		"participants":    allUserIDs,
		"created_by":      userID,
		"created_at":      conv.CreatedAt,
	}

	if msgBytes, err := json.Marshal(wsMsg); err == nil {
		h.Hub.Broadcast(msgBytes)
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

	wsMsg := map[string]interface{}{
		"type":            "participant_added",
		"conversation_id": conversationID,
		"user_id":         req.UserID,
	}
	if msgBytes, err := json.Marshal(wsMsg); err == nil {
		h.Hub.Broadcast(msgBytes)
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

	wsMsg := map[string]interface{}{
		"type":            "participant_removed",
		"conversation_id": conversationID,
		"user_id":         req.UserID,
	}
	if msgBytes, err := json.Marshal(wsMsg); err == nil {
		h.Hub.Broadcast(msgBytes)
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

	accessKey := os.Getenv("MINIO_ACCESS_KEY_ID")   
    secretKey := os.Getenv("MINIO_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("MINIO_BUCKET_NAME")

	endpoint := os.Getenv("MINIO_ENDPOINT") 
    if endpoint == "" {
        endpoint = "localhost:9000" 
    }

	publicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
    if publicEndpoint == "" {
        publicEndpoint = "http://localhost:9000"
    }

	minioClient, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: os.Getenv("MINIO_USE_SSL") == "true",
    })

	if err != nil {
		fmt.Printf("MinIO Connection Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to storage"})
		return
    }

    objectName := uuid.New().String() + filepath.Ext(header.Filename)
    contentType := header.Header.Get("Content-Type")

	ctx := context.Background()

	exists, err := minioClient.BucketExists(ctx, bucketName)
    if err != nil {
         fmt.Printf("Bucket Check Error: %v\n", err)
         c.JSON(http.StatusInternalServerError, gin.H{"error": "Storage bucket check failed"})
         return
    }
    if !exists {
        err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
        if err != nil {
             c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create storage bucket"})
             return
        }
    }

    _, err = minioClient.PutObject(ctx, bucketName, objectName, file, header.Size, minio.PutObjectOptions{
        ContentType: contentType,
    })

    if err != nil {
        fmt.Printf("Upload Error: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to storage"})
        return
    }

    fileURL := fmt.Sprintf("%s/%s/%s", publicEndpoint, bucketName, objectName)
    
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

	if err := h.Repo.DeleteConversation(c, conversationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete conversation"})
		return
	}

	wsMsg := map[string]interface{}{
		"type":            "conversation_deleted",
		"conversation_id": conversationID,
		"deleted_by":      userID,
	}

	if msgBytes, err := json.Marshal(wsMsg); err == nil {
		h.Hub.Broadcast(msgBytes)
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

func (h *ChatHandler) GetCallToken(c *gin.Context) {
	var req pb.GetCallTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.client.GetCallToken(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ChatHandler) GenerateCallToken(c *gin.Context) {
	conversationID := c.Param("id")
	userID := c.GetHeader("X-User-ID")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	appID := os.Getenv("AGORA_APP_ID")
	appCertificate := os.Getenv("AGORA_APP_CERTIFICATE")

	if appID == "" || appCertificate == "" {
		fmt.Println("Error: AGORA_APP_ID or AGORA_APP_CERTIFICATE not set")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Voice service configuration error"})
		return
	}

	channelName := conversationID
	
	tokenUID := uint32(0) 
	
	tokenExpirationInSeconds := uint32(86400) 
	privilegeExpirationInSeconds := uint32(86400)

	token, err := rtctokenbuilder2.BuildTokenWithUid(
		appID,
		appCertificate,
		channelName,
		tokenUID,
		rtctokenbuilder2.RolePublisher,
		tokenExpirationInSeconds,
		privilegeExpirationInSeconds,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate token: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":        token,
		"app_id":       appID,
		"channel_name": channelName,
	})
}

