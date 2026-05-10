package service

import (
	"SENT_backend/pkg/models"
	"SENT_backend/pkg/models/database"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Cấu hình linh hoạt cho 2 model của bạn
var (
	OLLAMA_BASE = getEnv("OLLAMA_URL", "http://sent_ollama:11434")
	MODEL_SMALL = "qwen3.5:0.8b"
	MODEL_LARGE = "qwen3.5:2b"
)

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Hàm tách Suy nghĩ và Câu trả lời
func parseAIResponse(raw string) (thought string, answer string) {
	if strings.Contains(raw, "<think>") && strings.Contains(raw, "</think>") {
		parts := strings.Split(raw, "</think>")
		thoughtPart := strings.ReplaceAll(parts[0], "<think>", "")
		answerPart := parts[1]
		return strings.TrimSpace(thoughtPart), strings.TrimSpace(answerPart)
	}
	return "", strings.TrimSpace(raw)
}

func ChatWithRAG(userQuestion string, orgID uint) (string, string, error) {
	// BƯỚC 1: RETRIEVAL (TÌM KIẾM)
	// Thay vì lấy sạch DB, ta chỉ lấy những thứ "dính" tới câu hỏi
	policies := getRelatedPolicies(userQuestion)
	incidents := getRelatedIncidents(userQuestion)
	docs := getRelevantPDFContent(userQuestion, orgID)

	// BƯỚC 2: AUGMENTATION (LÀM GIÀU PROMPT)
	// Prompt được thiết kế cực gọn cho Qwen 2b
	ragPrompt := fmt.Sprintf(`[CONTEXT]
POLICIES:
%s
INCIDENTS:
%s
DOCUMENTS:
%s

[INSTRUCTION]
Bạn là SENT Copilot. Dựa VÀO CONTEXT trên để trả lời câu hỏi: "%s".
- Nếu không thấy thông tin trong CONTEXT, hãy nói "Tôi không tìm thấy quy định này trong hệ thống".
- TUYỆT ĐỐI không bịa đặt. 
- Chỉ tư vấn, không hành động.`, policies, incidents, docs, userQuestion)

	// BƯỚC 3: GENERATION (TẠO CÂU TRẢ LỜI)
	return callOllama(MODEL_LARGE, ragPrompt)
}

func getRelatedPolicies(query string) string {
	var p []models.Policy
	// Tìm kiếm mờ (Fuzzy search) trong Postgres
	database.DB.Where("is_active = ? AND (title ILIKE ? OR value ILIKE ?)",
		true, "%"+query+"%", "%"+query+"%").Limit(5).Find(&p)

	res := ""
	for _, v := range p {
		res += fmt.Sprintf("- %s: %s\n", v.Title, v.Value)
	}
	return res
}

func getRelatedIncidents(query string) string {
	var incidents []models.Incident
	database.DB.Where("type ILIKE ? OR status ILIKE ?", "%"+query+"%", "%"+query+"%").Order("created_at desc").Limit(5).Find(&incidents)
	res := ""
	for _, inc := range incidents {
		res += fmt.Sprintf("- [%s] Loại: %s | Mức độ: %s | Máy trạm: %s | Trạng thái: %s\n",
			inc.CreatedAt.Format("15:04 02/01"), inc.Type, inc.Severity, inc.AssetHWID, inc.Status)
	}
	return res
}

func getRelevantPDFContent(query string, orgID uint) string {
	if database.DocumentContentCollection == nil {
		return ""
	}

	// Tìm kiếm mờ nội dung tài liệu trong MongoDB
	filter := bson.M{
		"org_id": int64(orgID),
		"$text":  bson.M{"$search": query}, // Yêu cầu đã tạo Text Index trong Mongo
	}

	cursor, err := database.DocumentContentCollection.Find(context.TODO(), filter, options.Find().SetLimit(2))
	if err != nil {
		return ""
	}

	var results []struct {
		Content string `bson:"content"`
	}
	cursor.All(context.TODO(), &results)

	contextText := "\nTHÔNG TIN TỪ TÀI LIỆU PDF (SOP):\n"
	for _, res := range results {
		// Cắt nhỏ văn bản (Chỉ lấy 1000 ký tự đầu liên quan để tránh tràn VRAM 4GB)
		limit := len(res.Content)
		if limit > 1000 {
			limit = 1000
		}
		contextText += fmt.Sprintf("- %s...\n", res.Content[:limit])
	}
	return contextText
}

func callOllama(model, prompt string) (string, string, error) {
	reqBody := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.3,  // Thấp để tránh AI "sáng tạo" lung tung
			"num_ctx":     2048, // Qwen 2b chạy tốt ở mức này
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/api/generate", strings.TrimRight(OLLAMA_BASE, "/"))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("lỗi kết nối AI: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var aiResp OllamaResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return "", "", fmt.Errorf("lỗi đọc phản hồi AI")
	}

	// 4. Tách suy nghĩ và câu trả lời
	thought, answer := parseAIResponse(aiResp.Response)
	return thought, answer, nil
}

// AnalyzeIncidentWithAI: Trợ lý AI chuyên phân tích Sự cố (Chỉ đọc, không hành động)
func AnalyzeIncidentWithAI(incidentID string) (string, error) {
	var incident models.Incident

	// 1. Lấy dữ liệu Sự cố và liên quan
	err := database.DB.Where("id = ?", incidentID).First(&incident).Error
	if err != nil {
		return "", fmt.Errorf("không tìm thấy hồ sơ sự cố")
	}

	var asset models.Asset
	if loadErr := database.DB.Where("asset_hwid = ?", incident.AssetHWID).First(&asset).Error; loadErr == nil {
		incident.Asset = asset
	}

	if database.SecurityAlertCollection != nil {
		cursor, _ := database.SecurityAlertCollection.Find(context.TODO(), bson.M{"incident_id": incident.ID})
		var alerts []models.SecurityAlert
		cursor.All(context.TODO(), &alerts)
		incident.Alerts = alerts
	}

	// 2. Đóng gói Alerts thành JSON
	alertsJSON, _ := json.Marshal(incident.Alerts)

	// 3. Prompt "Cố vấn" (Khóa quyền hành động)
	prompt := fmt.Sprintf(`[SYSTEM]
Bạn là SENT Copilot - Chuyên gia phân tích An ninh mạng (SOC Tier 3).
Nhiệm vụ: Đọc các cảnh báo trong Hồ sơ sự cố và báo cáo cho Quản trị viên.

[GIỚI HẠN TUYỆT ĐỐI]
1. Bạn CHỈ được phép đọc và tư vấn. Bạn KHÔNG CÓ QUYỀN thực thi lệnh.
2. CẤM sử dụng thẻ <tool_call>.
3. Trình bày 3 phần: [TÓM TẮT] - [ĐÁNH GIÁ RỦI RO] - [ĐỀ XUẤT XỬ LÝ].
4. Ở phần [ĐỀ XUẤT XỬ LÝ], BẮT BUỘC ghi rõ: Hệ thống SENT-SYSTEM chỉ có chức năng theo dõi (read-only). IT HD/SOC phải tới trực tiếp máy trạm hoặc sử dụng công cụ quản trị từ xa KHÁC (ngoài SENT-SYSTEM) để thực thi hành động xử lý (VD: gỡ phần mềm, rút USB, cấu hình Firewall).

[DỮ LIỆU SỰ CỐ (INCIDENT #%d)]
- Tên sự cố: %s (Mức độ: %s)
- Máy trạm: %s (IP: %s)
- Danh sách cảnh báo chi tiết:
%s

[USER]
Hãy phân tích sự cố này và cho tôi biết nên làm gì tiếp theo.`,
		incident.ID, incident.Type, incident.Severity,
		incident.Asset.Hostname, incident.Asset.IPAddress,
		string(alertsJSON),
	)

	// 4. Gọi tới Ollama (Nên dùng model 3b/4b với nhiệt độ 0.2)
	reqBody := OllamaRequest{
		Model:  MODEL_LARGE, // Dùng model chuyên sâu hơn cho việc phân tích
		Prompt: prompt,
		Stream: false,
	}

	reqBytes, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/api/generate", strings.TrimRight(OLLAMA_BASE, "/"))
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return "", fmt.Errorf("lỗi kết nối Ollama")
	}
	defer resp.Body.Close()

	var oResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&oResp); err != nil {
		return "", fmt.Errorf("lỗi đọc phản hồi")
	}

	_, answer := parseAIResponse(oResp.Response)
	return answer, nil
}
