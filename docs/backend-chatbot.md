# Backend Flow: AI Chatbot (SENT Copilot)

## Tổng quan

Tính năng AI Chatbot (SENT Copilot) là trợ lý AI tích hợp trong hệ thống SENT SOC, giúp SOC Analyst tư vấn về an ninh mạng, phân tích sự cố và tra cứu chính sách. Chatbot sử dụng mô hình ngôn ngữ lớn (LLM) Ollama để xử lý câu hỏi tự nhiên và cung cấp câu trả lời có cấu trúc.

## Kiến trúc tổng thể

```
Frontend (React) → Backend (Gin) → Ollama API → Database (PostgreSQL)
     ↓                ↓              ↓              ↓
- Chat UI        - REST API      - LLM Inference - Chat Logs
- Session Mgmt   - Auth Check    - Prompt Eng.   - Sessions
- Real-time WS   - Rate Limiting - Response Parse
```

## Luồng xử lý chính (Backend Flow)

### 1. Khởi tạo phiên chat (Session Management)

**Endpoint:** `POST /api/v1/ai/sessions`

**Flow:**
1. User gửi request tạo phiên mới
2. Middleware `AuthRequired()` kiểm tra JWT token
3. Handler `CreateSession()`:
   - Lấy `userID` từ context (thử nhiều key: "userID", "user_id", "sub")
   - Tạo record `AIChatSession` với `user_id` và `title` mặc định
   - Trả về session ID

**Database Model:**
```go
type AIChatSession struct {
    ID        uint      `gorm:"primarykey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    UserID    uint      `gorm:"index"`
    Title     string
}
```

### 2. Gửi tin nhắn chat (Chat Processing)

**Endpoint:** `POST /api/v1/ai/chat/:session_id`

**Flow chi tiết:**

#### 2.1 Authentication & Validation
- Middleware kiểm tra JWT
- Parse `session_id` từ URL param
- Lấy `userID` từ context
- Validate session thuộc về user

#### 2.2 Lưu tin nhắn user
```go
userLog := models.AIChatLog{
    UserID:    userID,
    SessionID: uint(sessionID),
    Role:      "user",
    Content:   req.Message,
}
database.DB.Create(&userLog)
```

#### 2.3 Xử lý AI (Core Logic)
Gọi `service_ai.ChatWithPolicy(req.Message)`

**Bước 3.1: Thu thập dữ liệu context**
- **Policies:** Lấy tất cả `UniversalPolicy` có `is_active = true`
- **Incidents:** Lấy 5 sự cố gần nhất, sắp xếp theo `created_at desc`
- **Playbook:** Knowledge base cố định về quy trình xử lý sự cố

**Bước 3.2: Tạo prompt có cấu trúc**
```go
finalPrompt := fmt.Sprintf(`<system_instructions>
Bạn là SENT Copilot - Trợ lý AI An ninh mạng cấp cao.
...
</system_instructions>

<system_data>
<policies>%s</policies>
<reent_incidents>%s</recent_incidents>
<playbooks>%s</playbooks>
</system_data>

<user_question>%s</user_question>`, ...)
```

**Bước 3.3: Gọi Ollama API**
```go
reqBody := OllamaRequest{
    Model:  AI_MODEL, // Từ env OLLAMA_MODEL
    Prompt: finalPrompt,
    Stream: false,
}
resp, err := http.Post(OLLAMA_URL, "application/json", bytes.NewBuffer(jsonData))
```

**Bước 3.4: Parse response**
- Parse JSON response từ Ollama
- Tách `thought` và `answer` từ raw response (dùng thẻ `<think>...</think>`)
- Nếu không có thẻ think, toàn bộ là answer

#### 2.4 Lưu phản hồi AI
```go
aiLog := models.AIChatLog{
    SessionID: uint(sessionID),
    UserID:    userID,
    Role:      "ai",
    Content:   answer,
    Thought:   thought,
}
database.DB.Create(&aiLog)
```

#### 2.5 Cập nhật session
```go
database.DB.Model(&AIChatSession{}).
    Where("id = ?", sessionID).
    Update("updated_at", time.Now())
```

#### 2.6 Response
```json
{
    "response": "Câu trả lời cho user",
    "thought": "Suy nghĩ nội bộ của AI"
}
```

### 3. Quản lý phiên (Session Operations)

#### 3.1 Lấy danh sách phiên
**Endpoint:** `GET /api/v1/ai/sessions`
- Lọc theo `user_id`
- Sắp xếp theo `updated_at desc`

#### 3.2 Lấy lịch sử chat
**Endpoint:** `GET /api/v1/ai/chat/:session_id`
- Lọc theo `session_id` và `user_id`
- Sắp xếp theo `created_at asc`

#### 3.3 Đổi tên phiên
**Endpoint:** `PUT /api/v1/ai/sessions/:id`
- Validate quyền sở hữu
- Update `title`

#### 3.4 Xóa phiên
**Endpoint:** `DELETE /api/v1/ai/sessions/:id`
- Transaction: Xóa `AIChatLog` trước, sau đó `AIChatSession`

### 4. Phân tích sự cố bằng AI (Incident Analysis)

**Endpoint:** `POST /api/v1/incidents/:id/ai-analyze`

**Flow:**
1. Lấy incident với preload Agent và Alerts
2. Tạo prompt cố vấn (không cho phép hành động)
3. Gọi Ollama với model khác (nếu cần)
4. Parse và trả về phân tích

## Cấu hình và Dependencies

### Environment Variables
```bash
OLLAMA_URL=http://127.0.0.1:11434/api/generate
AI_MODEL=qwen3.5:4b
```

### Database Tables
- `ai_chat_sessions`: Lưu phiên chat
- `ai_chat_logs`: Lưu lịch sử tin nhắn (user + AI)

### External Dependencies
- **Ollama**: LLM server chạy locally
- **PostgreSQL**: Lưu trữ sessions và logs

## Điểm mạnh và hạn chế

### Điểm mạnh
- **Context-aware**: Sử dụng dữ liệu real-time từ hệ thống
- **Structured prompts**: Đảm bảo AI tuân thủ quy tắc
- **Session management**: Hỗ trợ nhiều cuộc hội thoại
- **Audit trail**: Lưu toàn bộ lịch sử chat

### Hạn chế hiện tại
- **Performance**: Load toàn bộ policies (có thể chậm với nhiều policies)
- **Scalability**: Không có caching cho prompts
- **Security**: Không validate input prompt
- **Error handling**: Thiếu retry mechanism cho Ollama API

## Đề xuất cải tiến

### 1. Performance Optimization
- Implement RAG (Retrieval-Augmented Generation) thay vì load toàn bộ policies
- Cache prompts và responses
- Async processing cho long-running AI calls

### 2. Security Enhancements
- Input sanitization cho user messages
- Rate limiting cho AI calls
- Audit logging cho sensitive queries

### 3. Feature Extensions
- Multi-turn conversations với memory
- Integration với external knowledge bases
- Voice input/output support

## Monitoring và Debugging

### Logs
- AI request/response logs trong service
- Database errors trong handlers
- Authentication failures

### Metrics
- Response time cho AI calls
- Success rate của Ollama API
- Session creation/deletion rates

---

*Tài liệu này được tạo dựa trên phân tích code backend tính năng AI Chatbot. Cập nhật lần cuối: April 1, 2026*</content>
<parameter name="filePath">d:\HACOM\Code\SENT-System\docs\backend-chatbot.md