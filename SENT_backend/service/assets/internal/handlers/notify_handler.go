package handlers

import (
	"net/http"

	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"SENT_backend/pkg/websocket"

	"github.com/gin-gonic/gin"
)

// InternalNotify: internal endpoint used by other services to request an ASSET_UPDATE broadcast
// POST /api/v1/assets/internal/notify { "asset_hwid": "HWID-..." }
func InternalNotify(c *gin.Context) {
	var req struct {
		AssetHWID string `json:"asset_hwid" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "asset_hwid is required"})
		return
	}

	var asset models.Asset
	if err := database.DB.Select("org_id").Where("asset_hwid = ?", req.AssetHWID).First(&asset).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	// Broadcast to all clients in the organization that asset has updated
	websocket.GlobalHub.BroadcastToOrg(asset.OrgID, map[string]interface{}{"type": "ASSET_UPDATE", "hwid": req.AssetHWID})

	c.JSON(http.StatusOK, gin.H{"message": "notified"})
}
