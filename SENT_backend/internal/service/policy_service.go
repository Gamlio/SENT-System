package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sent_backend/internal/models"
	"sent_backend/internal/repository"
	"time"
)

func SavePolicyFile(orgID uint, title, category, fileName string, fileData []byte) error {
	// 1. Tự động tạo cây thư mục uploads/policies nếu chưa có
	uploadDir := filepath.Join("uploads", "policies")
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return err
	}

	// 2. Tạo tên file duy nhất bằng timestamp để không bị ghi đè
	safeFileName := fmt.Sprintf("%d_%s", time.Now().Unix(), fileName)
	filePath := filepath.Join(uploadDir, safeFileName)

	// 3. Ghi dữ liệu file ra ổ cứng
	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return err
	}

	// 4. Lưu vào DB với đường dẫn web-friendly
	doc := &models.PolicyDocument{
		OrgID:    orgID,
		Title:    title,
		FileName: safeFileName,
		FilePath: filepath.ToSlash(filePath), // Chuyển \ thành / để Frontend dễ đọc
		Category: category,
	}
	return repository.CreatePolicy(doc)
}

func DeletePolicyAndFile(id string, orgID uint) error {
	// Lấy thông tin để biết đường dẫn file
	doc, err := repository.GetPolicyByID(id, orgID)
	if err != nil {
		return err
	}

	// Xóa file vật lý
	os.Remove(doc.FilePath)

	// Xóa khỏi DB
	return repository.DeletePolicy(id, orgID)
}
