package repository

import (
	"sent_backend/internal/database"
	"sent_backend/internal/models"
)

func CreatePolicy(doc *models.PolicyDocument) error {
	return database.DB.Create(doc).Error
}

func GetPoliciesByOrg(orgID uint) ([]models.PolicyDocument, error) {
	var docs []models.PolicyDocument
	// Sắp xếp mới nhất lên đầu
	err := database.DB.Where("org_id = ?", orgID).Order("created_at desc").Find(&docs).Error
	return docs, err
}

func GetPolicyByID(id string, orgID uint) (models.PolicyDocument, error) {
	var doc models.PolicyDocument
	err := database.DB.Where("id = ? AND org_id = ?", id, orgID).First(&doc).Error
	return doc, err
}

func DeletePolicy(id string, orgID uint) error {
	return database.DB.Where("id = ? AND org_id = ?", id, orgID).Delete(&models.PolicyDocument{}).Error
}

func UpdatePolicyStatus(id string, title, category string) error {
	return database.DB.Model(&models.PolicyDocument{}).Where("id = ?", id).
		Updates(map[string]interface{}{"title": title, "category": category}).Error
}
