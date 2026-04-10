
# SENT asset Logic Flow (v4.0 - Ninja asset & Multi-OS)

Tài liệu này mô tả kiến trúc của asset sau quá trình tái cấu trúc (Refactor). asset hiện tại hoạt động theo triết lý **"Ninja Sensor"**: Hoàn toàn im lặng, cực kỳ nhẹ, không tự đưa ra quyết định (Không tích hợp AI tại biên), chỉ làm nhiệm vụ thu thập, băm dữ liệu và gửi về Backend.

## 1. Cấu trúc Source Code (Refactored Structure)
Dự án asset được thiết kế theo chuẩn module hóa:

```text
SENT/
├── cmd/
│   └── main.go              // Điểm khởi chạy (Entry point)
├── internal/
│   ├── config/              // Quản lý asset_config.json, mã công ty
│   ├── collector/           // Module Thu thập dữ liệu Đa nền tảng
│   │   ├── system.go        // Lấy Inventory (CPU, RAM, OS, Hostname)
│   │   ├── software.go      // Quét phần mềm, tính File Hash, check Ghost Registry
│   │   ├── usb.go           // Bắt sự kiện USB, lấy VID/PID, tính Device Hash
│   │   └── telemetry.go     // Lấy trạng thái Tường lửa, Quét Port mạng (Nmap local)
│   ├── network/             // Module Giao tiếp Backend
│   │   ├── client.go        // Đóng gói JSON, xử lý HTTP/HTTPS Request
│   │   └── diff.go          // Thuật toán Differential Reporting (So sánh Hash)
│   └── utils/               // Các hàm tiện ích (Lấy Outbound IP, Hash...)
├── build.sh
├── go.mod
└── go.sum

2. Luồng vận hành chính (The "Ninja" Flow)
A. Khởi động & Định danh (Bootstrap)
Load Config: Đọc asset_config.json. Nếu chưa có, yêu cầu nhập Company Code.

Identify Host: Sinh ra AssetHWID  (Hardware ID) độc nhất dựa trên Mainboard/MAC Address. Xác định IP LAN thực tế qua hàm GetOutboundIP().

B. Thu thập Đa nền tảng (Multi-OS Collectors)
Tự động nhận diện OS (runtime.GOOS) để chạy lệnh tương ứng:

Software: - Windows: Quét Registry. Phát hiện Ghost Registry (Có key nhưng mất file vật lý). Tính SHA-256 của file thực thi.

Linux: Quét dpkg-query / rpm.

USB Monitoring: Lấy chính xác VID, PID và tính Device Hash để định danh phần cứng chống giả mạo.

Telemetry: Kiểm tra trạng thái OS Firewall và các Port đang mở rủi ro (VD: 3389, 22).

C. Gửi dữ liệu thông minh (Differential Reporting)
Để tối ưu 90% băng thông mạng cho Doanh nghiệp, asset KHÔNG gửi toàn bộ log liên tục:

Hashing: Băm toàn bộ cục dữ liệu vừa thu thập thành chuỗi SHA-256.

Compare: So sánh với Hash của chu kỳ trước (lưu trong RAM).

Decision:

Khác nhau: Gửi gói Type: "DATA" (Chứa toàn bộ log mới).

Giống nhau: Chỉ gửi gói Type: "HEARTBEAT" (Body rỗng) để duy trì trạng thái Online.

3. Cấu trúc Payload JSON (Gửi lên Backend)
Payload được thiết kế bọc trong thẻ Data để Backend dễ dàng làm "Router" điều phối.

A. Gói Software (Kèm Hash & Trạng thái Lẩn tránh)
JSON
{
  "type": "DATA",
  "log_type": "software",
  "hwid": "WIN-ABC123XYZ",
  "hostname": "PC-KETOAN-01",
  "data": [
    {
      "software_name": "Unikey",
      "version": "4.3",
      "publisher": "Pham Kim Long",
      "install_location": "C:\\Program Files\\Unikey",
      "file_hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
      "status": "INSTALLED", 
      "is_running": true
    }
  ]
}
(Lưu ý: Nếu status là GHOST_REGISTRY, Backend sẽ tự động lập Case [Defense Evasion]).

B. Gói USB (Kèm Định danh Phần cứng)
JSON
{
  "type": "DATA",
  "log_type": "usb",
  "hwid": "WIN-ABC123XYZ",
  "hostname": "PC-KETOAN-01",
  "data": [
    {
      "device_name": "Kingston DataTraveler 3.0",
      "vid": "0951",
      "pid": "1666",
      "serial_number": "E0D55EA573E4F",
      "device_hash": "a1b2c3d4e5f6...",
      "event_type": "active"
    }
  ]
}
C. Gói Telemetry (Mạng & Tường lửa)
JSON
{
  "type": "DATA",
  "log_type": "telemetry",
  "hwid": "WIN-ABC123XYZ",
  "hostname": "PC-KETOAN-01",
  "data": {
    "ip_address": "192.168.1.15",
    "firewall_off": false,
    "open_ports": [
      { "port": 3389, "protocol": "TCP", "service": "ms-wbt-server" }
    ]
  }
}
4. Quản lý Vòng đời asset (Lifecycle)
Enrollment: Cài đặt lần đầu -> Trạng thái PENDING trên SOC. (Zero-Trust: Mọi dữ liệu gửi lên lúc này đều bị Backend vứt bỏ).

Active: Trưởng ca SOC bấm Duyệt -> Trạng thái ACTIVE. Bắt đầu phân tích log và tính Điểm Rủi ro.

Response: Khi có lệnh từ SOC (Cách ly mạng, Kill Process), asset sẽ thực thi thông qua cơ chế Polling (hoặc WebSocket/MQTT trong tương lai).

**(Các lỗ hổng bảo mật cấp Enterprise còn tồn tại):**

#### Lỗ hổng 1: Giả mạo asset (asset Spoofing)
- **Tình trạng:** Hiện tại asset gửi API lên Backend chỉ dựa vào `hwid` (Ví dụ: `WIN-ABC123XYZ`). AssetHWID  này là mã tĩnh dễ dàng bị lộ hoặc đoán được.
- **Rủi ro:** Hacker (hoặc một nhân viên nội bộ) có thể dùng Postman giả mạo AssetHWID  của máy tính "Giám đốc", sau đó gửi một gói JSON Telemetry chứa `{"firewall_off": true}`. Lập tức máy Giám đốc bị nhảy điểm rủi ro và SOC phát báo động giả.
- **Giải pháp khắc phục:** Tại bước Enrollment, sau khi Admin bấm duyệt, Backend phải cấp cho asset một **JWT Token** (hoặc TLS Certificate). Từ đó trở đi, asset gửi dữ liệu phải đính kèm Token này vào Header `Authorization: Bearer <Token>`.

#### Lỗ hổng 2: Thay đổi dữ liệu trên đường truyền (Man-in-the-Middle)
- **Tình trạng:** Dữ liệu JSON `{"is_running": true}` đang được gửi trần trụi.
- **Rủi ro:** Dù có dùng HTTPS, nếu máy tính bị dính mã độc ở mức proxy nội bộ, mã độc có thể đánh tráo gói tin, sửa `{"file_hash": "mã_độc"}` thành `{"file_hash": "mã_an_toàn"}` trước khi nó bay ra khỏi máy.
- **Giải pháp khắc phục:** Cần áp dụng **HMAC (Hash-based Message Authentication Code)**. asset dùng một Secret Key bí mật (chỉ asset và Backend biết) để ký lên toàn bộ chuỗi JSON. Backend nhận được sẽ kiểm tra chữ ký, nếu sai 1 ký tự -> Vứt bỏ.

#### Lỗ hổng 3: Bảo mật chức năng Upload File (Post-Mortem)
- **Tình trạng:** Trong hàm `AddIncidentAudit`, Backend đang lưu BẤT KỲ file nào có trong FormData vào thư mục `uploads/incidents/`.
- **Rủi ro:** Một Kỹ sư SOC bị hack tài khoản (hoặc Hacker chiếm phiên) có thể đính kèm một file `shell.php` hoặc `backdoor.sh` vào ô đính kèm ảnh của Case. Dù file được đổi tên (`timestamp_shell.php`), nhưng nếu thư mục `/uploads/` cho phép thực thi mã, Server Go sẽ bị chiếm quyền điều khiển.
- **Giải pháp khắc phục:** Trong hàm xử lý file của `incident_handlers.go`, bạn phải kiểm tra đuôi file (Chỉ cho phép `.png`, `.jpg`, `.pdf`) VÀ kiểm tra MIME Type thực sự của file đó trước khi lệnh `SaveUploadedFile` được chạy.