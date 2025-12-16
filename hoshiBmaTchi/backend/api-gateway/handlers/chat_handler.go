package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/chat"
)

type ChatHandler struct {
	client pb.ChatServiceClient
}

func NewChatHandler(client pb.ChatServiceClient) *ChatHandler {
	return &ChatHandler{client: client}
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