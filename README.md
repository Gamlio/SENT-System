# 🛡️ SENT - AI-Powered Endpoint Detection & Response (EDR)

Hệ thống giám sát an ninh đầu cuối và quản trị chính sách tập trung, tích hợp **Trợ lý AI Copilot** hỗ trợ phân tích và phản ứng sự cố theo tiêu chuẩn SOC hiện đại.

##  Tổng quan hệ thống
SENT (Security Endpoint Network Tracker) là một giải pháp EDR toàn diện được phát triển để phát hiện, phân tích và phản ứng với các mối đe dọa an ninh trên máy trạm. Hệ thống áp dụng mô hình **Adaptive SOC** với khả năng tự động chấm điểm rủi ro và tư vấn xử lý bằng Trí tuệ nhân tạo.

##  Kiến trúc kỹ thuật (3-Tier Architecture)
* **Endpoint Agent (Golang):** Tác tử chạy ngầm hiệu năng cao, thu thập viễn trắc (Telemetry) và thực hiện phản xạ bảo mật tại chỗ (Local Rules).
* **Backend SOC (Golang - Gin):** Trạm điều phối trung tâm, xử lý thuật toán Risk Scoring (P1, P2, P3) và quản lý Incident War Room.
* **AI Copilot (LLM Integration):** Sử dụng mô hình ngôn ngữ lớn (Qwen/Phi3) để phân tích log và hướng dẫn xử lý theo Playbook.
* **Frontend (ReactJS - Tailwind):** Giao diện quản trị SOC trực quan, theo dõi trạng thái Real-time và tương tác AI.

## 5 Use-cases Giám sát cốt lõi
1.  **Malware Detection:** Phát hiện mã độc thông qua tích hợp dữ liệu từ Windows Defender.
2.  **Security Compliance:** Kiểm tra trạng thái bản vá Windows Update và tường lửa hệ thống.
3.  **Network Monitoring:** Giám sát các cổng mạng (Open Ports) lạ đang lắng nghe trên thiết bị.
4.  **Software Control:** Đối soát danh mục phần mềm cài đặt với chính sách Whitelist/Blacklist của tổ chức.
5.  **Peripheral Security:** Giám sát lịch sử kết nối thiết bị ngoại vi USB lạ.

## Tính năng nổi bật
* **Adaptive Risk Scoring:** Tự động tính toán điểm nguy hiểm dựa trên mức độ vi phạm Playbook.
* **AI Incident Analysis:** AI tự động đọc log sự cố và đưa ra kết luận tư vấn kỹ thuật.
* **Real-time Alerting:** Cảnh báo Popup trực tiếp tại máy trạm và đẩy thông báo về trung tâm SOC trong < 5 giây.