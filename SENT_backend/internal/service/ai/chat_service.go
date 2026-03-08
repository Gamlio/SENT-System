package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sent_backend/internal/database"
	"sent_backend/internal/models"
	"strings"
)

const OLLAMA_URL = "http://127.0.0.1:11434/api/generate"

// const AI_MODEL = "qwen2.5:3b"
const AI_MODEL = "qwen3:4b"

// const AI_MODEL = "phi3"

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
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
	var policies []models.UniversalPolicy
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
			incidentContext += fmt.Sprintf("- [%s] Loại: %s | Mức độ: %s | Agent: %s | Trạng thái: %s\n",
				inc.CreatedAt.Format("15:04 02/01"), inc.Type, inc.Severity, inc.AgentHWID, inc.Status)
		}
	}

	playbookKnowledge := `
    QUY TRÌNH XỬ LÝ SỰ CỐ:
    1. Firewall Disabled (P1): Xác minh người dùng -> Bật lại -> Quét mã độc.
    2. Malware/AV Alert (P2): Cô lập máy -> Quét Full Disk -> Cài lại OS nếu cần.
    3. Unpatched OS (P3): Kiểm tra version -> Chạy Windows Update.
    4. Unauthorized (P3): Gỡ phần mềm / Rút USB -> Cảnh báo.
    `

	// TẠO PROMPT MỚI: ÉP BUỘC RÕ RÀNG HƠN
	finalPrompt := fmt.Sprintf(`Bạn là SENT Copilot - Trợ lý AI An ninh mạng cấp cao.

			QUY TẮC ỨNG XỬ (TUÂN THỦ TUYỆT ĐỐI):
			1. GIAO TIẾP THÔNG THƯỜNG: Nếu người dùng gửi lời chào (VD: "Hello", "Hi", "Chào"),
		 	bạn BẮT BUỘC chỉ trả lời lại bằng 1 câu chào ngắn gọn đúng ngôn ngữ đó.
			KHÔNG ĐƯỢC báo cáo sự cố. KHÔNG ĐƯỢC dùng thẻ <tool_call>.
			2. TRẢ LỜI NGHIỆP VỤ: Chỉ khi người dùng hỏi về lỗi, sự cố hoặc chính sách,
			bạn mới dùng thẻ <tool_call> suy luận <tool_call> và đọc dữ liệu bên dưới để tư vấn.

			DỮ LIỆU HỆ THỐNG:
			%s
			%s
			%s

			Câu hỏi của người dùng: "%s"`, policyContext, incidentContext, playbookKnowledge, userQuestion)

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
