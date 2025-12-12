package handlers

import (
	"net/http"

	pb "github.com/Hinsane5/hoshiBmaTchi/backend/proto/users"
	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	userClient pb.UserServiceClient
}

func NewSettingsHandler(userClient pb.UserServiceClient) *SettingsHandler {
	return &SettingsHandler{userClient: userClient}
}

func (h *SettingsHandler) UpdateProfile(c *gin.Context) {
	var req pb.UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	req.UserId = userID.(string)

	res, err := h.userClient.UpdateUserProfile(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *SettingsHandler) GetSettings(c *gin.Context) {
	userID, _ := c.Get("userID")
	
	res, err := h.userClient.GetSettings(c, &pb.GetSettingsRequest{UserId: userID.(string)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *SettingsHandler) UpdateNotifications(c *gin.Context) {
	var req pb.UpdateNotificationSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("userID")
	req.UserId = userID.(string)

	res, err := h.userClient.UpdateNotificationSettings(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *SettingsHandler) UpdatePrivacy(c *gin.Context) {
	var req pb.UpdatePrivacySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("userID")
	req.UserId = userID.(string)

	res, err := h.userClient.UpdatePrivacySettings(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *SettingsHandler) GetCloseFriends(c *gin.Context) {
	userID, _ := c.Get("userID")
	res, err := h.userClient.GetCloseFriends(c, &pb.GetListRequest{UserId: userID.(string)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *SettingsHandler) AddCloseFriend(c *gin.Context) {
	targetID := c.Param("id")
	userID, _ := c.Get("userID")

	res, err := h.userClient.AddCloseFriend(c, &pb.ManageRelationRequest{
		UserId:       userID.(string),
		TargetUserId: targetID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *SettingsHandler) RemoveCloseFriend(c *gin.Context) {
	targetID := c.Param("id")
	userID, _ := c.Get("userID")

	res, err := h.userClient.RemoveCloseFriend(c, &pb.ManageRelationRequest{
		UserId:       userID.(string),
		TargetUserId: targetID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *SettingsHandler) GetHiddenStoryUsers(c *gin.Context) {
	userID, _ := c.Get("userID")
	res, err := h.userClient.GetHiddenStoryUsers(c, &pb.GetListRequest{UserId: userID.(string)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *SettingsHandler) HideStoryFromUser(c *gin.Context) {
	targetID := c.Param("id")
	userID, _ := c.Get("userID")

	res, err := h.userClient.HideStoryFromUser(c, &pb.ManageRelationRequest{
		UserId:       userID.(string),
		TargetUserId: targetID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *SettingsHandler) UnhideStoryFromUser(c *gin.Context) {
	targetID := c.Param("id")
	userID, _ := c.Get("userID")

	res, err := h.userClient.UnhideStoryFromUser(c, &pb.ManageRelationRequest{
		UserId:       userID.(string),
		TargetUserId: targetID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *SettingsHandler) RequestVerification(c *gin.Context) {
	var req pb.RequestVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("userID")
	req.UserId = userID.(string)

	res, err := h.userClient.RequestVerification(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}