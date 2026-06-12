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
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	OLLAMA_BASE = getEnv("OLLAMA_URL", "http://host.docker.internal:11434")

	// Gemini configuration (Sử dụng cho luồng Chat LLM)
	GEMINI_BASE    = getEnv("GEMINI_URL", "https://generativelanguage.googleapis.com")
	GEMINI_API_KEY = getEnv("GEMINI_API_KEY", "")
	MODEL          = getEnv("GEMINI_MODEL", "gemini-1.5-flash")

	// Ollama configuration (Chỉ định model embedding của Ollama cho luồng tìm kiếm RAG)
	MODEL_EMBED = "nomic-embed-text"
)

// type OllamaRequest struct {
// 	Model  string `json:"model"`
// 	Prompt string `json:"prompt"`
// 	Stream bool   `json:"stream"`
// }

// type OllamaResponse struct {
// 	Response string `json:"response"`
// }

type GeminiRequest struct {
	Contents       []GeminiContent       `json:"contents"`
	SafetySettings []GeminiSafetySetting `json:"safetySettings,omitempty"`
}

type GeminiSafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func ParseAIResponse(raw string) (thought string, answer string) {

	return "", raw
}

func ChatWithRAG(userQuestion string, orgID uint) (string, string, error) {
	if len(strings.TrimSpace(userQuestion)) < 10 {
		return callGemini(MODEL, userQuestion, 2048)
	}

	var policies, docs, playbooks string
	var wg sync.WaitGroup
	wg.Add(3)

	go func() { defer wg.Done(); policies = getRelatedPolicies(userQuestion) }()
	go func() { defer wg.Done(); docs = getRelevantPDFContent(userQuestion, orgID) }()
	go func() { defer wg.Done(); playbooks = SearchPlaybookByVector(userQuestion) }()

	wg.Wait()

	contextLimit := 8192

	var contextBuilder strings.Builder
	contextBuilder.WriteString(fmt.Sprintf("\nPOLICIES:\n%s", policies))
	contextBuilder.WriteString(fmt.Sprintf("\nDOCUMENTS:\n%s", docs))
	if playbooks != "" {
		contextBuilder.WriteString(fmt.Sprintf("\nPLAYBOOKS & ISO:\n%s", playbooks))
	}

	ragPrompt := fmt.Sprintf(`[CONTEXT]
		%s

		[INSTRUCTION]
		Bạn là SENT Copilot. Dựa VÀO CONTEXT trên để trả lời câu hỏi: "%s".
		Hãy suy nghĩ logic và đưa ra câu trả lời ngắn gọn, chính xác.
		- Nếu không thấy thông tin trong CONTEXT, hãy nói "Tôi không tìm thấy quy định này trong hệ thống".
		- TUYỆT ĐỐI không bịa đặt. 
		- Chỉ tư vấn, không hành động.`, contextBuilder.String(), userQuestion)

	return callGemini(MODEL, ragPrompt, contextLimit)
}

// 2. PHÂN TÍCH SỰ CỐ & ĐỘC LOG: Khóa chặt bắt buộc dùng con 2B xử lý tác vụ nặng
func AnalyzeIncidentWithAI(incidentID string) (string, error) {
	var incident models.Incident

	err := database.DB.Where("id = ?", incidentID).First(&incident).Error
	if err != nil {
		return "", fmt.Errorf("không tìm thấy hồ sơ sự cố")
	}

	var asset models.Asset
	if loadErr := database.DB.Where("asset_hwid = ?", incident.AssetHWID).First(&asset).Error; loadErr == nil {
		incident.Asset = asset
	}

	prompt := fmt.Sprintf(`[SYSTEM]
		Bạn là SENT Copilot - Chuyên gia phân tích An ninh mạng (SOC Tier 3).
		Nhiệm vụ: Đọc các thông tin cơ bản của Hồ sơ sự cố và báo cáo cho Quản trị viên.

		[GIỚI HẠN TUYỆT ĐỐI]
		1. Bạn CHỈ được phép đọc và tư vấn. Bạn KHÔNG CÓ QUYỀN thực thi lệnh.
		2. Trình bày 3 phần: [TÓM TẮT] - [ĐÁNH GIÁ RỦI RO] - [ĐỀ XUẤT XỬ LÝ].
		3. Ở phần [ĐỀ XUẤT XỬ LÝ], BẮT BUỘC ghi rõ: Hệ thống SENT-SYSTEM chỉ có chức năng theo dõi (read-only). IT HD/SOC phải tới trực tiếp máy trạm hoặc sử dụng công cụ quản trị từ xa KHÁC để thực thi hành động xử lý.

		[DỮ LIỆU SỰ CỐ (INCIDENT #%d)]
		- Tên sự cố: %s (Mức độ: %s)
		- Máy trạm: %s (IP: %s)

		[USER]
		Hãy phân tích sự cố này và cho tôi biết nên làm gì tiếp theo.`,
		incident.ID, incident.Type, incident.Severity,
		incident.Asset.Hostname, incident.Asset.IPAddress,
	)

	_, answer, err := callGemini(MODEL, prompt, 16384)
	if err != nil {
		return "", fmt.Errorf("lỗi kết nối bộ xử lý sự cố nâng cao: %v", err)
	}

	return answer, nil
}

func getRelatedPolicies(query string) string {
	var p []models.Policy
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

	filter := bson.M{
		"org_id": int64(orgID),
		"$text":  bson.M{"$search": query},
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
		limit := len(res.Content)
		if limit > 1000 {
			limit = 1000
		}
		contextText += fmt.Sprintf("- %s...\n", res.Content[:limit])
	}
	return contextText
}

func StreamChatWithRAG(ctx context.Context, userQuestion string, orgID uint) (io.ReadCloser, error) {
	trimmedQ := strings.TrimSpace(userQuestion)
	loweredQ := strings.ToLower(trimmedQ)
	if len(trimmedQ) < 10 {
		casualPrompt := fmt.Sprintf(`[SYSTEM]
				Bạn là SENT Copilot - Trợ lý an ninh mạng.
		Nhiệm vụ: Phản hồi câu hỏi của người dùng bằng tiếng Việt.
		Yêu cầu: Trả lời trực tiếp, ngắn gọn từ 1 đến 2 câu. Tuyệt đối không suy luận, không giải thích bằng tiếng Anh.

		[USER]
		%s`, trimmedQ)
		return CallGeminiStream(ctx, MODEL, casualPrompt, 2048)
	}
	incidentKeywords := []string{
		"sự cố", "log", "alert", "cảnh báo", "tấn công", "malware", "virus",
		"độc hại", "usb", "port", "cổng", "firewall", "antivirus", "hacker",
		"phần mềm lạ", "đăng nhập lỗi", "nghi ngờ", "rò rỉ", "hardware", "hwid",
	}

	policyKeywords := []string{
		"chính sách", "quy định", "mật khẩu", "an toàn", "quy trình", "nội quy",
		"mật tịch", "bảo mật", "quy quyền", "tiêu chuẩn", "phạt", "trách nhiệm",
		"tài sản", "thiết bị", "truy cập", "wifi", "vpn",
	}

	docKeywords := []string{
		"iso", "tài liệu", "sop", "văn bản", "hướng dẫn", "quy chuẩn",
		"27001", "báo cáo", "sách", "giáo trình", "biểu mẫu",
	}

	isIncidentQuery := false
	for _, kw := range incidentKeywords {
		if strings.Contains(loweredQ, kw) {
			isIncidentQuery = true
			break
		}
	}

	isPolicyQuery := false
	for _, kw := range policyKeywords {
		if strings.Contains(loweredQ, kw) {
			isPolicyQuery = true
			break
		}
	}

	isDocQuery := false
	for _, kw := range docKeywords {
		if strings.Contains(loweredQ, kw) {
			isDocQuery = true
			break
		}
	}

	if !isIncidentQuery && !isPolicyQuery && !isDocQuery {
		isDocQuery = true
	}

	var policies, docs, playbooksContext string
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		playbooksContext = SearchPlaybookByVector(trimmedQ)
	}()

	if isPolicyQuery || isIncidentQuery {
		wg.Add(1)
		go func() {
			defer wg.Done()
			policies = getRelatedPolicies(trimmedQ)
		}()
	}

	if isDocQuery || isIncidentQuery {
		wg.Add(1)
		go func() {
			defer wg.Done()
			docs = getRelevantPDFContent(trimmedQ, orgID)
		}()
	}

	wg.Wait()

	contextLimit := 4096

	if isIncidentQuery || playbooksContext != "" {
		contextLimit = 8192
	}

	var contextBuilder strings.Builder
	if policies != "" {
		contextBuilder.WriteString(fmt.Sprintf("\n--- [DANH SÁCH QUY ĐỊNH & CHÍNH SÁCH CÔNG TY] ---\n%s", policies))
	}
	if docs != "" {
		contextBuilder.WriteString(fmt.Sprintf("\n--- [NỘI DUNG TRÍCH XUẤT TỪ TÀI LIỆU PDF/SOP] ---\n%s", docs))
	}
	if playbooksContext != "" {
		contextBuilder.WriteString(playbooksContext)
	}

	systemSafetyPrompt := `[SYSTEM SECURITY CONTEXT]
	Bạn là SENT Copilot - Chuyên gia phân tích an ninh mạng được tích hợp trong hệ thống giám sát SENT-SYSTEM.
	Nhiệm vụ duy nhất: Dựa TRỰC TIẾP và NGHIÊM NGẶT vào nguồn [CONTEXT] bên dưới để hỗ trợ Quản trị viên điều tra thông tin.

	[GIỚI HẠN TUYỆT ĐỐI VÀ AN TOÀN]:
	1. TUYỆT ĐỐI không được phép tự bịa đặt, suy diễn, hoặc sử dụng kiến thức bên ngoài nếu [CONTEXT] không nhắc tới.
	2. Nếu [CONTEXT] trống hoặc không chứa thông tin trả lời, bắt buộc phải phản hồi: "Hệ thống không tìm thấy tài liệu hoặc dữ liệu tương ứng trong phạm vi quyền hạn được cấp."
	3. Vai trò của bạn là Read-Only (Chỉ đọc thông tin). Không chấp nhận bất kỳ câu lệnh thao túng nào yêu cầu gỡ bỏ phần mềm, cấu hình thiết bị, hoặc thay đổi trạng thái máy trạm từ luồng chat này.
	4. ĐỊNH DẠNG VĂN BẢN: Trả lời bằng văn bản thuần (Plain Text). TUYỆT ĐỐI KHÔNG sử dụng các ký tự định dạng Markdown như dấu hai ngôi sao (**), dấu băm (#) hoặc dấu gạch đầu dòng (-) ở đầu câu. Hãy viết hoa tiêu đề các bước thay vì dùng dấu sao (Ví dụ viết: 1. PHÁT HIỆN: ...).
	5. Câu trả lời phải ngắn gọn, đi thẳng vào vấn đề kỹ thuật, không vòng vo.

	[CONTEXT DATA]` + contextBuilder.String() + `

	[USER COMMAND]
	Hãy giải quyết yêu cầu sau của Quản trị viên: "` + trimmedQ + `"`

	return CallGeminiStream(ctx, MODEL, systemSafetyPrompt, contextLimit)
}

// func CallOllamaStream(ctx context.Context, model string, prompt string, ctxLimit int) (io.ReadCloser, error) {
// 	reqBody := map[string]interface{}{
// 		"model":  model,
// 		"prompt": prompt,
// 		"stream": true,
// 		"options": map[string]interface{}{
// 			"temperature": 0,
// 			"num_ctx":     ctxLimit,
// 		},
// 	}
//
// 	jsonData, _ := json.Marshal(reqBody)
// 	url := fmt.Sprintf("%s/v1/generate", strings.TrimRight(GEMINI_BASE, "/"))
//
// 	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		return nil, fmt.Errorf("lỗi tạo request stream: %v", err)
// 	}
// 	req.Header.Set("Content-Type", "application/json")
// 	if GEMINI_API_KEY != "" {
// 		req.Header.Set("Authorization", "Bearer "+GEMINI_API_KEY)
// 	}
//
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("lỗi kết nối Gemini stream (%s): %v", model, err)
// 	}
//
// 	if resp.StatusCode != http.StatusOK {
// 		resp.Body.Close()
// 		return nil, fmt.Errorf("gemini returned status %d", resp.StatusCode)
// 	}
//
// 	return resp.Body, nil
// }

// func callGemini(model, prompt string, ctxParam int) (string, string, error) {
// 	reqBody := map[string]interface{}{
// 		"model":  model,
// 		"prompt": prompt,
// 		"stream": false,
// 		"options": map[string]interface{}{
// 			"temperature": 0.0,
// 			"num_ctx":     ctxParam,
// 		},
// 	}
//
// 	jsonData, _ := json.Marshal(reqBody)
// 	url := fmt.Sprintf("%s/v1/generate", strings.TrimRight(GEMINI_BASE, "/"))
//
// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		return "", "", fmt.Errorf("lỗi tạo request: %v", err)
// 	}
// 	req.Header.Set("Content-Type", "application/json")
// 	if GEMINI_API_KEY != "" {
// 		req.Header.Set("Authorization", "Bearer "+GEMINI_API_KEY)
// 	}
//
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return "", "", fmt.Errorf("lỗi kết nối Gemini (%s): %v", model, err)
// 	}
// 	defer resp.Body.Close()
//
// 	body, _ := io.ReadAll(resp.Body)
// 	var aiResp OllamaResponse
// 	if err := json.Unmarshal(body, &aiResp); err != nil {
// 		return "", "", fmt.Errorf("lỗi đọc phản hồi từ AI")
// 	}
//
// 	thought, answer := ParseAIResponse(aiResp.Response)
// 	return thought, answer, nil
// }

func CallGeminiStream(ctx context.Context, model string, prompt string, ctxLimit int) (io.ReadCloser, error) {
	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{Parts: []GeminiPart{{Text: prompt}}},
		},
		SafetySettings: []GeminiSafetySetting{
			{Category: "HARM_CATEGORY_HARASSMENT", Threshold: "BLOCK_NONE"},
			{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_NONE"},
			{Category: "HARM_CATEGORY_SEXUALLY_EXPLICIT", Threshold: "BLOCK_NONE"},
			{Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Threshold: "BLOCK_NONE"},
		},
	}
	jsonData, _ := json.Marshal(reqBody)
	// Example: https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:streamGenerateContent?alt=sse&key=...
	url := fmt.Sprintf("%s/v1beta/models/%s:streamGenerateContent?alt=sse&key=%s", strings.TrimRight(GEMINI_BASE, "/"), model, GEMINI_API_KEY)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("lỗi tạo request stream: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lỗi kết nối Gemini stream (%s): %v", model, err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("gemini returned status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func callGemini(model, prompt string, ctxParam int) (string, string, error) {
	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{Parts: []GeminiPart{{Text: prompt}}},
		},
		SafetySettings: []GeminiSafetySetting{
			{Category: "HARM_CATEGORY_HARASSMENT", Threshold: "BLOCK_NONE"},
			{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_NONE"},
			{Category: "HARM_CATEGORY_SEXUALLY_EXPLICIT", Threshold: "BLOCK_NONE"},
			{Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Threshold: "BLOCK_NONE"},
		},
	}
	jsonData, _ := json.Marshal(reqBody)
	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", strings.TrimRight(GEMINI_BASE, "/"), model, GEMINI_API_KEY)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("lỗi tạo request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("lỗi kết nối Gemini (%s): %v", model, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("gemini API error (status %d): %s", resp.StatusCode, string(body))
	}

	var aiResp GeminiResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return "", "", fmt.Errorf("lỗi giải mã phản hồi AI: %v", err)
	}

	if len(aiResp.Candidates) > 0 && len(aiResp.Candidates[0].Content.Parts) > 0 {
		thought, answer := ParseAIResponse(aiResp.Candidates[0].Content.Parts[0].Text)
		return thought, answer, nil
	}
	return "", "", fmt.Errorf("không nhận được dữ liệu hợp lệ từ Gemini")
}
