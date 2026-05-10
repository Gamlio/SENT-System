package handlers

import (
	"SENT_backend/pkg/models"
	assetSvc "SENT_backend/service/assets/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAssetTypes(c *gin.Context) {
	orgID := c.GetUint("org_id")
	svc := &assetSvc.AssetTypeService{}
	types, err := svc.GetTypes(orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy danh sách"})
		return
	}
	c.JSON(http.StatusOK, types)
}

func UpdateAssetType(c *gin.Context) {
	orgID := c.GetUint("org_id")
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req models.AssetType
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ"})
		return
	}

	svc := &assetSvc.AssetTypeService{}
	if err := svc.UpdateType(uint(id), orgID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cập nhật thành công và đang tính toán lại rủi ro"})
}
