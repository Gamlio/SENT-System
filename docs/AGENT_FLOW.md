# SENT Agent Logic Flow (v3.2 - Compliance & Telemetry Focus)

Tài liệu này mô tả cơ chế hoạt động của Agent (được viết bằng Go) triển khai trên các thiết bị đầu cuối (Endpoint). Agent được thiết kế theo hướng nhẹ (lightweight), đóng vai trò như một "Sensor" thu thập viễn trắc để phục vụ bài toán Giám sát Tuân thủ An ninh (Security Compliance) và IT Audit.

## 1. Kích hoạt ngầm đa khách thuê (Multi-tenant Silent Enrollment)
- Agent sử dụng `Enrollment Token` (được cấp bởi Admin của từng công ty).
- Khi cài đặt, Agent tự động trích xuất HWID (Hardware ID) và gửi request kèm Token về API `/api/v1/agents/push`.
- Backend tự động đối chiếu Token để mapping máy trạm vào đúng `OrgID` (Công ty) và Vùng/Chi nhánh tương ứng.

## 2. Báo cáo Sai khác tối ưu Băng thông (Differential Reporting)
- Agent định kỳ quét 3 nhóm tài nguyên tĩnh: Phần cứng (Inventory), Danh sách phần mềm (Software), và Cổng mạng mở (Open Ports).
- Thay vì gửi toàn bộ dữ liệu, Agent băm (Hash) trạng thái hiện tại và so sánh với Hash của lần gửi trước. 
- Dữ liệu chỉ được đóng gói JSON và đẩy về Backend khi có sự thay đổi (sai khác), giúp giảm tải tối đa cho hệ thống mạng nội bộ.

## 3. Giám sát Sự kiện Thời gian thực (Real-time Event Hooking)
- Agent giám sát Windows Event Log, Registry và I/O Device (USB).
- Khi phát hiện hành vi vi phạm (Cắm USB không được phép, đăng nhập sai nhiều lần - Event 4625), Agent gửi ngay `SecurityAlert` về Backend.
- Backend tiếp nhận, đối chiếu với `Universal Policy` và tự động gom cụm các cảnh báo này thành một Hồ sơ Sự cố (`Incident`) hoàn chỉnh.
sent_agent/
├── cmd/
│   └── main.go              // Điểm khởi chạy (Entry point) - Chỉ gọi các module khác
├── internal/
│   ├── config/              // Quản lý file cấu hình, nhập mã công ty
│   │   └── config.go
│   ├── collector/           // Logic thu thập dữ liệu (Inventory, USB, Soft...)
│   │   ├── system.go        // Inventory (CPU, RAM, OS)
│   │   ├── software.go      // Software list
│   │   ├── network.go       // Telemetry (IP, Port)
│   │   └── usb.go           // USB devices
│   ├── transport/           // Gửi dữ liệu đi (HTTP, Hashing)
│   │   └── client.go
│   └── utils/               // Các hàm tiện ích chung
│       └── helpers.go
├── agent_config.json        // File cấu hình (tự sinh)
├── go.mod
└── go.sum