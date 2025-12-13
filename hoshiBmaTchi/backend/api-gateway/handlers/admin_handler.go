package handlers

import (
	"context"
	"net/http"
	
	"github.com/gin-gonic/gin"
	postPb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/posts"
	userPb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
)

type AdminHandler struct {
	UserClient userPb.UserServiceClient
	PostClient postPb.PostsServiceClient
}

func NewAdminHandler(userClient userPb.UserServiceClient, postClient postPb.PostsServiceClient) *AdminHandler {
	return &AdminHandler{
		UserClient: userClient,
		PostClient: postClient,
	}
}

// --- User Management ---

func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	res, err := h.UserClient.GetAllUsers(context.Background(), &userPb.Empty{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res.Users)
}

func (h *AdminHandler) BanUser(c *gin.Context) {
	userID := c.Param("id")
	var req struct {
		IsBanned bool `json:"is_banned"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body"})
		return
	}

	_, err := h.UserClient.ToggleUserBan(context.Background(), &userPb.ToggleUserBanRequest{
		UserId:   userID,
		IsBanned: req.IsBanned,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User status updated"})
}

// --- Newsletter ---

func (h *AdminHandler) SendNewsletter(c *gin.Context) {
	var req struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subject and Body required"})
		return
	}

	// 1. Get emails from User Service
	res, err := h.UserClient.GetSubscribedEmails(context.Background(), &userPb.Empty{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscribers"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Newsletter queued for " + string(rune(len(res.Emails))) + " subscribers"})
}

// --- Verification ---

func (h *AdminHandler) GetVerificationRequests(c *gin.Context) {
	res, err := h.UserClient.GetVerificationRequests(context.Background(), &userPb.Empty{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res.Requests)
}

func (h *AdminHandler) ReviewVerification(c *gin.Context) {
	reqID := c.Param("id")
	var req struct {
		Action string `json:"action"` // ACCEPT or REJECT
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Action required"})
		return
	}

	_, err := h.UserClient.ReviewVerification(context.Background(), &userPb.ReviewVerificationRequest{
		RequestId: reqID,
		Action:    req.Action,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Review processed"})
}

// --- Reports ---

func (h *AdminHandler) GetReports(c *gin.Context) {
	reportType := c.Query("type") // "post" or "user"

	if reportType == "post" {
		res, err := h.PostClient.GetPostReports(context.Background(), &postPb.Empty{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res.Reports)
	} else {
		res, err := h.UserClient.GetUserReports(context.Background(), &userPb.Empty{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res.Reports)
	}
}

func (h *AdminHandler) ReviewReport(c *gin.Context) {
	reportID := c.Param("id")
	var req struct {
		Type   string `json:"type"`   // "post" or "user"
		Action string `json:"action"` // "DELETE_POST", "BAN_USER", "IGNORE"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type and Action required"})
		return
	}

	if req.Type == "post" {
		_, err := h.PostClient.ReviewPostReport(context.Background(), &postPb.ReviewReportRequest{
			ReportId: reportID,
			Action:   req.Action,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		_, err := h.UserClient.ReviewUserReport(context.Background(), &userPb.ReviewReportRequest{
			ReportId: reportID,
			Action:   req.Action,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Report handled"})
}