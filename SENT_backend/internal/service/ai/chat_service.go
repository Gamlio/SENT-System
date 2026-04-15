package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

var (
	// Lấy cấu hình từ biến môi trường, với giá trị mặc định
	OLLAMA_URL = getEnv("OLLAMA_URL", "http://127.0.0.1:11434/api/generate")
	AI_MODEL   = getEnv("AI_MODEL", "qwen3.5:4b")
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

// Logic chính gọi sang Ollama
func ChatWithPolicy(userQuestion string) (string, string, error) {
	// 1. LẤY POLICY (Giữ nguyên)
	// CẢNH BÁO HIỆU NĂNG: Lấy tất cả policy có thể gây chậm hệ thống.
	// ĐỀ XUẤT: Sử dụng Vector Search (RAG) để tìm các policy liên quan nhất đến câu hỏi.
	var policies []models.Policy
	database.DB.Where("is_active = ?", true).Find(&policies)
	policyContext := "CHÍNH SÁCH CÔNG TY:\n"
	for _, p := range policies {
		policyContext += fmt.Sprintf("- %s (%s): %s\n", p.Title, p.PolicyType, p.Value)
	}

	var incidents []models.Incident
	database.DB.Order("created_at desc").Limit(5).Find(&incidents)
	incidentContext := "\nCÁC SỰ CỐ GẦN ĐÂY:\n"
	if len(incidents) == 0 {
		incidentContext += "Hiện không có sự cố nào.\n"
	} else {
		for _, inc := range incidents {
			incidentContext += fmt.Sprintf("- [%s] Loại: %s | Mức độ: %s | asset: %s | Trạng thái: %s\n",
				inc.CreatedAt.Format("15:04 02/01"), inc.Type, inc.Severity, inc.AssetHWID, inc.Status)
		}
	}

	playbookKnowledge := `
    QUY TRÌNH XỬ LÝ SỰ CỐ:
    1. Firewall Disabled (P1): Xác minh người dùng -> Bật lại -> Quét mã độc.
    2. Malware/AV Alert (P2): Cô lập máy -> Quét Full Disk -> Cài lại OS nếu cần.
    3. Unpatched OS (P3): Kiểm tra version -> Chạy Windows Update.
    4. Unauthorized (P3): Gỡ phần mềm / Rút USB -> Cảnh báo.
    `

	// TẠO PROMPT MỚI: Sử dụng thẻ XML để cấu trúc hóa, giúp AI hoạt động ổn định hơn.
	finalPrompt := fmt.Sprintf(`<system_instructions>
Bạn là SENT Copilot - Trợ lý AI An ninh mạng cấp cao.

QUY TẮC ỨNG XỬ (TUÂN THỦ TUYỆT ĐỐI):
1. GIAO TIẾP THÔNG THƯỜNG: Nếu người dùng gửi lời chào (VD: "Hello", "Hi", "Chào"), bạn BẮT BUỘC chỉ trả lời lại bằng 1 câu chào ngắn gọn đúng ngôn ngữ đó. KHÔNG ĐƯỢC báo cáo sự cố hoặc dùng thẻ <tool_call>.
2. TRẢ LỜI NGHIỆP VỤ: Chỉ khi người dùng hỏi về lỗi, sự cố hoặc chính sách, bạn mới dùng thẻ <tool_call> để suy luận và đọc dữ liệu bên dưới để tư vấn.
3. QUY TRÌNH XỬ LÝ (QUAN TRỌNG): Hệ thống SENT-SYSTEM CHỈ CÓ CHỨC NĂNG PHÁT HIỆN VÀ BÁO CÁO (Read-only). Khi đưa ra hướng xử lý, bạn PHẢI nhấn mạnh rằng IT HD cần đi tới máy trạm hoặc dùng công cụ quản trị từ xa KHÁC (ngoài SENT-SYSTEM) để thực thi (VD: gỡ phần mềm, rút USB, cấu hình Firewall).
</system_instructions>

<system_data>
<policies>
%s
</policies>
<recent_incidents>
%s
</recent_incidents>
<playbooks>
%s
</playbooks>
</system_data>

<user_question>%s</user_question>`, policyContext, incidentContext, playbookKnowledge, userQuestion)

	// 3. GỬI REQUEST
	reqBody := OllamaRequest{
		Model:  AI_MODEL,
		Prompt: finalPrompt,
		Stream: false,
	}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post(OLLAMA_URL, "application/json", bytes.NewBuffer(jsonData))
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
		Model:  AI_MODEL, // Dùng AI_MODEL bạn đã định nghĩa ở trên
		Prompt: prompt,
		Stream: false,
	}

	reqBytes, _ := json.Marshal(reqBody)
	resp, err := http.Post(OLLAMA_URL, "application/json", bytes.NewBuffer(reqBytes))
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
