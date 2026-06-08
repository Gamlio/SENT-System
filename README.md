# 🛡️ SENT - AI-Powered Endpoint Detection & Response (EDR)

Hệ thống giám sát an ninh đầu cuối và quản trị chính sách tập trung, tích hợp **Trợ lý AI Copilot** hỗ trợ phân tích và phản ứng sự cố theo tiêu chuẩn SOC hiện đại.

##  Tổng quan hệ thống
SENT-SYSTEM (Sentinex)
Hệ thống Quản trị An ninh tập trung (SOC) & Tuân thủ rủi ro (GRC) tích hợp Trợ lý AI.
👉 [Xem Cấu trúc Thư mục Dự án SENT trực tuyến](https://htmlpreview.github.io/?https://github.com/Gamlio/SENT-System/blob/main/project_visual.html)
📌 Tổng quan
SENT-SYSTEM là một nền tảng giám sát an ninh mạng dành cho các doanh nghiệp SME, kết hợp giữa việc thu thập dữ liệu viễn trắc (Telemetry) từ máy trạm và khả năng tư vấn của Trí tuệ nhân tạo (AI Copilot). Hệ thống giúp đội ngũ IT HD/SOC phát hiện sớm các vi phạm chính sách, lệch chuẩn Baseline và nhận được hướng dẫn xử lý sự cố theo thời gian thực.

✨ Tính năng cốt lõi
Giám sát thiết bị (Asset Monitoring): Thu thập thông tin phần cứng, phần mềm, trạng thái cổng mạng, USB và lưu lượng I/O theo thời gian thực.

Quản lý tuân thủ (Baseline & Policy): Định nghĩa cấu hình chuẩn (Baseline) và các chính sách Whitelist/Blacklist để phát hiện sai lệch an ninh.

Trợ lý AI Copilot (RAG-based): Sử dụng Local LLM (Ollama) kết hợp với ngữ cảnh từ Chính sách và Playbook để tư vấn hướng xử lý sự cố cho IT HD.

Quản lý sự cố (Incident Management): Quy trình xử lý sự cố chặt chẽ với cơ chế xác minh tính toàn vẹn của bằng chứng (Hash Chain).

Quy trình phê duyệt (Maker-Checker): Mọi thay đổi nhạy cảm về hệ thống, tài liệu hoặc người dùng đều phải được Admin phê duyệt.

Phân tích rủi ro (Risk & Trust Scoring): Tự động tính toán điểm rủi ro tức thời và điểm uy tín dài hạn cho từng thiết bị.

🏗 Kiến trúc hệ thống
Hệ thống bao gồm hai thành phần chính:

SENT Backend (Golang): Trung tâm điều phối, xử lý dữ liệu viễn trắc, quản lý cơ sở dữ liệu (PostgreSQL & MongoDB) và giao tiếp với AI API.

Go-SENT  (Golang): Bộ thu thập dữ liệu siêu nhẹ (Ninja Thin-Client) chạy trên máy trạm Windows/Linux để đẩy dữ liệu về Server.

🛠 Công nghệ sử dụng
Ngôn ngữ: Golang (Gin Framework).

Cơ sở dữ liệu:

PostgreSQL: Quản lý cấu hình, người dùng, chính sách (Dữ liệu quan hệ).

MongoDB: Lưu trữ nhật ký sự cố, log viễn trắc, audit logs (Dữ liệu lớn).

AI: Ollama API (Model: Qwen/Phi3).

Bảo mật: JWT Authentication, HMAC Payload Signing, Hash Chaining.

Khác: WebSockets (Real-time updates), Redis (Planned for Caching)
```