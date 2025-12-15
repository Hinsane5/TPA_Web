package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"github.com/gin-gonic/gin"
)

type AIHandler struct{}

func NewAIHandler() *AIHandler {
	return &AIHandler{}
}

type SummarizeRequest struct {
	Text string `json:"text" binding:"required"`
}

type SummarizeResponse struct {
	Summary string `json:"summary"`
}

func (h *AIHandler) SummarizeCaption(c *gin.Context) {
	var req SummarizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	jsonData, _ := json.Marshal(map[string]string{
		"text": req.Text,
	})

	resp, err := http.Post("http://ai-service:8000/summarize", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Failed to connect to AI Service"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "AI Service returned an error"})
		return
	}

	var aiResp SummarizeResponse
	if err := json.NewDecoder(resp.Body).Decode(&aiResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse AI response"})
		return
	}

	c.JSON(http.StatusOK, aiResp)
}