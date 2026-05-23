package service

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"fmt"
)

type AssetTypeService struct{}

func (s *AssetTypeService) GetTypes(orgID uint) ([]models.AssetType, error) {
	var types []models.AssetType
	err := database.DB.Where("org_id = ?", orgID).Order("name asc").Find(&types).Error
	return types, err
}

func (s *AssetTypeService) CreateType(orgID uint, req models.AssetType) error {
	newType := models.AssetType{
		OrgID:       orgID,
		Name:        req.Name,
		Description: req.Description,
		RiskWeight:  req.RiskWeight,
	}
	if newType.RiskWeight == 0 {
		newType.RiskWeight = 1.0
	}
	return database.DB.Create(&newType).Error
}

func (s *AssetTypeService) UpdateType(id uint, orgID uint, req models.AssetType) (*models.AssetType, error) {
	var existingType models.AssetType
	if err := database.DB.Where("id = ? AND org_id = ?", id, orgID).First(&existingType).Error; err != nil {
		return nil, fmt.Errorf("không tìm thấy loại tài sản")
	}

	existingType.Name = req.Name
	existingType.Description = req.Description
	existingType.RiskWeight = req.RiskWeight

	if err := database.DB.Save(&existingType).Error; err != nil {
		return nil, err
	}

	return &existingType, nil
}
