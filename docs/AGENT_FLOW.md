# SENT Agent Logic Flow (v3.1)

Tài liệu này mô tả cách thức Agent (Viết bằng Go) hoạt động trên máy trạm điểm cuối (Endpoint) và giao tiếp với hệ thống Backend.

## 1. Silent Enrollment (Kích hoạt ngầm)
- **Admin (Level 2)** tạo mã `Enrollment Token` cho một Chi nhánh/Vùng cụ thể trên Web Dashboard.
- Khi cài đặt, Agent tự động đọc mã này từ file cấu hình (hoặc tham số dòng lệnh), thu thập mã định danh phần cứng (`HWID`) và tự động gọi API đăng ký về Backend.
- Nếu thành công, Backend tự động map máy trạm đó vào đúng công ty và chi nhánh mà không cần thao tác thủ công.

## 2. Giám sát Tài sản & Tuân thủ (Asset & Compliance Monitoring)
- Agent sử dụng cơ chế **Differential Reporting (Báo cáo sai khác)**.
- Quét định kỳ cấu hình phần cứng (Inventory), danh sách phần mềm (Software) và các thiết bị ngoại vi (USB/Port).
- Dữ liệu được băm (Hash) và so sánh. Agent chỉ đẩy dữ liệu về API `/api/v1/agents/push` khi có sự thay đổi, giúp tiết kiệm tối đa băng thông mạng.

## 3. Cảnh báo Thời gian thực (Real-time Alerts)
- Thay vì tự ra quyết định khóa máy, Agent làm nhiệm vụ "Tai mắt" (Sensor). 
- Bất kỳ khi nào có một thiết bị USB lạ cắm vào, Agent ngay lập tức gửi Log về Backend. Backend sẽ kiểm tra chéo với bảng `Rules` và trả về lệnh xử lý (nếu có).
## 4. Giám sát Hành vi & Kết nối (Nâng cấp v4.0)
- **Security Event Monitoring**: Agent theo dõi các Event ID nhạy cảm (4625 - Brute force, 1102 - Clear Logs). Khi phát hiện, Log được đẩy về AI để phân tích hành vi tấn công.
- **Outbound Telemetry**: Thu thập các kết nối `ESTABLISHED`. Giúp hệ thống phát hiện máy trạm đang kết nối tới các máy chủ C2 (Command & Control) hoặc gửi dữ liệu ra ngoài lãnh thổ.