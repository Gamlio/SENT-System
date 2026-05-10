package service

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"fmt"
	"net/http"
)

type AssetTypeService struct{}

// GetTypes lấy danh sách các loại máy của công ty
func (s *AssetTypeService) GetTypes(orgID uint) ([]models.AssetType, error) {
	var types []models.AssetType
	err := database.DB.Where("org_id = ?", orgID).Order("risk_weight desc").Find(&types).Error
	return types, err
}

// UpdateType cập nhật thông tin và hệ số rủi ro
func (s *AssetTypeService) UpdateType(id uint, orgID uint, req models.AssetType) error {
	var existing models.AssetType
	if err := database.DB.Where("id = ? AND org_id = ?", id, orgID).First(&existing).Error; err != nil {
		return err
	}

	// 1. Cập nhật dữ liệu
	err := database.DB.Model(&existing).Updates(map[string]interface{}{
		"name":        req.Name,
		"risk_weight": req.RiskWeight,
		"description": req.Description,
		"icon":        req.Icon,
	}).Error

	if err != nil {
		return err
	}

	// 2. LOGIC QUAN TRỌNG: Tính lại điểm cho tất cả Asset thuộc loại này
	// Vì Weight thay đổi nên Risk Score của máy trạm sẽ thay đổi theo
	go s.recalculateAllAssetsInType(id)

	return nil
}

func (s *AssetTypeService) recalculateAllAssetsInType(typeID uint) {
	var assetHWIDs []string
	database.DB.Model(&models.Asset{}).Where("asset_type_id = ?", typeID).Pluck("asset_hwid", &assetHWIDs)

	for _, hwid := range assetHWIDs {
		// Gọi API cho từng máy trạm thay vì gọi hàm scoring.RecalculateRiskScore
		go func(id string) {
			url := fmt.Sprintf("http://scoring-service:8010/api/v1/scoring/recalculate/%s", id)
			_, err := http.Post(url, "application/json", nil)
			if err != nil {
				fmt.Printf("⚠️ Lỗi gọi Scoring API cho máy %s: %v\n", id, err)
			}
		}(hwid)
	}
}
