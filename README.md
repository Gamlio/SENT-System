# 🚀 SENT - AI-Powered Security Operations Center (SOC)
Hệ thống giám sát an ninh đầu cuối (EDR) và quản trị chính sách tập trung, tích hợp Trợ lý AI (Knowledge RAG) dành cho doanh nghiệp.

## 🌟 Tính năng cốt lõi
1. **Trạm điều phối trung tâm (Backend Go-Gin):**
   - Xử lý viễn trắc (Telemetry) theo thời gian thực từ Agent.
   - Quản trị Rule-base tự động đối soát Phần mềm, Thiết bị USB và Mạng.
2. **Trung tâm Tri thức AI (Knowledge Base):**
   - Hỗ trợ nạp tài liệu (PDF/Word), tiêu chuẩn ISO 27001 để xây dựng Vector Database.
   - Cung cấp dữ liệu ngữ cảnh cho AI Security Assistant.
3. **Giao diện SOC chuyên nghiệp (ReactJS - Tailwind):**
   - Cấu trúc Component-based với Custom Hooks (`useAgents`, `usePolicies`).
   - Theo dõi trạng thái Online/Offline của thiết bị theo Real-time.
   - Quản lý chính sách bảo mật tập trung (Policy Center) áp dụng toàn cầu (Global) hoặc đích danh máy trạm (Specific).
4. **Agent tự động (Go):**
   - Chạy ngầm nhẹ nhàng, tự động thu thập Inventory, Software, Port và USB Logs.

## 🛠️ Công nghệ sử dụng
- **Backend:** Golang, Gin Framework, GORM, PostgreSQL, JWT, Bcrypt.
- **Frontend:** ReactJS, Tailwind CSS, Axios, Lucide Icons.
- **AI/ML Integration:** Tích hợp RAG (Retrieval-Augmented Generation) để tư vấn quy định nội bộ.