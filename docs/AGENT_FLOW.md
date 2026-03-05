# SENT Agent Logic Flow (v3.5 - Refactored & Multi-OS)

Tài liệu này mô tả kiến trúc của Agent sau khi đã tái cấu trúc (Refactor) để dễ bảo trì và mở rộng. Agent đóng vai trò là "Sensor" đa nền tảng, chạy được trên Windows, Linux và macOS.

## 1. Cấu trúc Source Code (Refactored Structure)
Dự án Agent được chia theo kiến trúc **Domain-Driven Design (DDD)** đơn giản hóa:

2. Luồng vận hành chính (Operational Flow)
A. Khởi động & Cấu hình (Bootstrap)
Load Config: Kiểm tra file agent_config.json. Nếu chưa có, chuyển sang chế độ PromptForCompanyCode để người dùng nhập mã (VD: SME-701D60).

Identify Host: Lấy HWID (Hardware ID) và Hostname từ hệ điều hành.

B. Cơ chế Thu thập Đa luồng (Multi-OS Collectors)
Agent tự động phát hiện hệ điều hành (runtime.GOOS) để chạy logic tương ứng:

Inventory:

Windows: Lấy tên phiên bản qua Registry/WMI.

Linux: Đọc thông tin Kernel và Distro.

Software:

Windows: Quét HKLM\Software\Microsoft\Windows\CurrentVersion\Uninstall.

Linux: Chạy lệnh dpkg-query (Ubuntu/Debian).

USB Monitoring:

Windows: Dùng PowerShell Get-CimInstance Win32_PnPEntity.

Linux: Dùng lệnh lsusb.

C. Cơ chế Gửi thông minh (Smart Transportation)
Để tối ưu băng thông, Agent sử dụng cơ chế Differential Reporting:

Hashing: Dữ liệu thu thập (VD: Danh sách phần mềm) được băm thành chuỗi SHA256.

Compare: So sánh Hash mới với Hash cũ lưu trong RAM (LastHashes).

Decision:

Khác nhau: Gửi gói tin DATA (chứa toàn bộ dữ liệu mới) -> Backend cập nhật DB.

Giống nhau: Chỉ gửi gói tin HEARTBEAT (Body rỗng) -> Backend chỉ cập nhật LastSeen.

Ngoại lệ: Gói tin Telemetry (IP, Port) luôn được gửi để duy trì trạng thái Online thời gian thực.

3. Xử lý Mạng (Network Logic Fix)
Agent sử dụng hàm utils.GetOutboundIP() (kết nối UDP giả lập đến 8.8.8.8) để xác định chính xác IP LAN (VD: 192.168.1.15) thay vì 127.0.0.1, đảm bảo hiển thị đúng trên Dashboard.
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