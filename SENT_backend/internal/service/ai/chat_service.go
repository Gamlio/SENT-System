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

const AI_MODEL = "qwen2.5:3b"

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

	// 2. [MỚI] LẤY 5 SỰ CỐ MỚI NHẤT (Để AI biết tình hình hiện tại)
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

	// 3. [MỚI] NHÚNG KIẾN THỨC PLAYBOOK (Tóm tắt từ ảnh bạn gửi)
	playbookKnowledge := `
    QUY TRÌNH XỬ LÝ SỰ CỐ (PLAYBOOKS):
    1. Unauthorized Software (P3):
       - B1: Xác định thiết bị và người dùng qua Woodpecker.
       - B2: Gỡ bỏ phần mềm trái phép.
       - B3: Cảnh báo người dùng. Nếu tái phạm -> Leo thang.
    2. Virus Infection (P2):
       - B1: Cô lập máy trạm (Ngắt mạng).
       - B2: Dùng Anti-virus quét full disk.
       - B3: Nếu không sạch -> Cài lại Win (Re-image).
    3. Unauthorized USB (P3):
       - B1: Kiểm tra Device ID.
       - B2: Yêu cầu rút USB.
       - B3: Nếu USB lạ -> Tịch thu kiểm tra malware.
    `

	// 4. TẠO PROMPT NÂNG CAO
	finalPrompt := fmt.Sprintf(`
	Bạn là Chuyên gia SOC tại Trung tâm điều hành an ninh SENT.
	
	DỮ LIỆU HỆ THỐNG HIỆN TẠI:
	%s
	%s
	
	KIẾN THỨC XỬ LÝ (PLAYBOOKS):
	%s
	
	YÊU CẦU:
	1. Dựa vào "CÁC SỰ CỐ GẦN ĐÂY", hãy đưa ra nhận định nếu người dùng hỏi về tình hình an ninh.
	2. Nếu người dùng hỏi cách xử lý, hãy dùng "KIẾN THỨC PLAYBOOK" để hướng dẫn từng bước.
	3. Luôn SUY LUẬN trong thẻ <think> trước khi trả lời.

	Câu hỏi: "%s"
	`, policyContext, incidentContext, playbookKnowledge, userQuestion)

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
