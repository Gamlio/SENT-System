package repository

import (
	"sent_backend/internal/database"
	"sent_backend/internal/models"
)

func CreatePolicy(doc *models.Document) error {
	return database.DB.Create(doc).Error
}

func GetPoliciesByOrg(orgID uint) ([]models.Document, error) {
	var docs []models.Document
	// Sắp xếp mới nhất lên đầu
	err := database.DB.Where("org_id = ?", orgID).Order("created_at desc").Find(&docs).Error
	return docs, err
}

func GetPolicyByID(id string, orgID uint) (models.Document, error) {
	var doc models.Document
	err := database.DB.Where("id = ? AND org_id = ?", id, orgID).First(&doc).Error
	return doc, err
}

func DeletePolicy(id string, orgID uint) error {
	return database.DB.Where("id = ? AND org_id = ?", id, orgID).Delete(&models.Document{}).Error
}

func UpdatePolicyStatus(id string, title, category string) error {
	return database.DB.Model(&models.Document{}).Where("id = ?", id).
		Updates(map[string]interface{}{"title": title, "category": category}).Error
}
