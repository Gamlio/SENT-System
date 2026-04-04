# ĐÁNH GIÁ KIẾN TRÚC & LỘ TRÌNH MỞ RỘNG (Sentinex SOC)

## 1. ƯU ĐIỂM VƯỢT TRỘI (The Good)
Hệ thống của bạn có nền tảng rất vững chắc so với nhiều dự án mã nguồn mở khác:

- **Kiến trúc lõi hiệu năng cao:** Sử dụng Go (Golang) cho cả Backend và asset là một lựa chọn xuất sắc. Nó giúp asset (Ninja Thin-Client) tiêu tốn rất ít RAM/CPU, trong khi Backend có thể xử lý hàng nghìn luồng (Goroutines) đồng thời.
- **Đồng bộ Trạng thái Thực (Real-time State Diffing) [MỚI]:** Hệ thống không chỉ ghi nhận log tĩnh (Stateless) mà đã sở hữu thuật toán Diffing để theo dõi vòng đời thực của thiết bị (Cắm/Rút USB, Bật/Tắt tiến trình, Mở/Đóng Port). Điều này giúp triệt tiêu hoàn toàn vấn nạn Spam Alert và phản ánh đúng thực trạng máy trạm (Digital Twin).
- **Mô hình Chấm điểm Tiên tiến (Sentinex v6.0):** Áp dụng cơ chế Trust Score (Uy tín dài hạn), Contextual Matrix (Ngữ cảnh phòng ban) và Hàm tiệm cận (Asymptotic) đưa hệ thống này tiệm cận với logic của các giải pháp Enterprise (như CrowdStrike hay SentinelOne).
- **Thiết kế Cảm biến Mô-đun (Plug-and-Play):** Cấu trúc `collector.Registry` ở asset cho phép viết thêm tính năng giám sát mới (như Giám sát Network I/O, quét RAM) cực kỳ dễ dàng.
- **Giao thức Kết hợp (Hybrid Protocol):** Dùng HTTP REST để đẩy log lớn và WebSocket để truyền lệnh Real-time (Zero Trust Baseline, Isolate) là một thiết kế thông minh, tối ưu băng thông.

## 2. NHƯỢC ĐIỂM & ĐỘ KHÓ BẢO TRÌ (The Bad)
Hệ thống hiện tại là một khối Monolithic (Nguyên khối) khá chặt chẽ, điều này sinh ra một số khó khăn:

- **Phụ thuộc quá nhiều vào GORM (RDBMS):** Đang lưu trữ mọi thứ (từ thông tin máy trạm, cấu hình, cho đến hàng triệu dòng Log USB, Process) vào chung một CSDL quan hệ (PostgreSQL/MySQL). Khi số lượng máy trạm lên mức 1,000+, truy vấn lịch sử sẽ bị thắt cổ chai.
- **Nút thắt cổ chai ở WebSocket (Hub):** Lưu các kết nối vào bộ nhớ RAM (`map[string]*websocket.Conn`). Khi muốn chạy 2 server Backend để chịu tải (Load Balancing), các server này sẽ không biết máy trạm nào đang kết nối ở server kia.

## 3. LỖ HỔNG BẢO MẬT HIỆN TẠI (Security Vulnerabilities)
Là một giải pháp an ninh, Sentinex đang có 3 rủi ro kỹ thuật cần vá:

- **Lộ lọt SecretKey qua URL:** asset gọi WebSocket bằng `ws://.../ws?token=SECRET_KEY`. Tham số này dễ bị ghi lại dạng rõ (plaintext) trong log proxy. -> *Cách sửa: Đưa SecretKey vào HTTP Header.*
- **Tấn công phát lại (Replay Attack):** Hàm tạo chữ ký HMAC-SHA256 thiếu Timestamp (Thời gian) hoặc Nonce. Hacker có thể "nghe lén" gói tin hợp lệ và gửi lại (Replay) liên tục. -> *Cách sửa: Thêm Timestamp vào Payload, chỉ nhận gói tin lệch không quá 5 phút.*
- **Lưu SecretKey dạng rõ:** File `sent_config.json` ở máy trạm đang lưu dạng Text. -> *Cách sửa: Mã hóa bằng DPAPI (Windows) hoặc dùng Credential Manager.*

## 4. BÀI TOÁN MỞ RỘNG TƯƠNG LAI (EDR ROADMAP)
Để Sentinex phục vụ 10,000+ máy trạm và trở thành một EDR thực thụ, lộ trình nâng cấp gồm 4 Trụ cột:

### Giai đoạn 1 (Nâng cấp Não bộ SOC - Đang triển khai)
- **Động cơ Tương quan Sự kiện (Correlation Engine):** Gom các cảnh báo đơn lẻ (VD: Cắm USB + Chạy file lạ + Mở port) thành một Siêu sự cố (Incident) để phát hiện chuỗi tấn công (Kill Chain).
- **Phát hiện Data Exfiltration & Ransomware:** Bổ sung cảm biến đo lường Network I/O và Disk I/O để phát hiện hành vi tuồn dữ liệu ra ngoài hoặc mã hóa ổ cứng hàng loạt.

### Giai đoạn 2 (Tối ưu Hạ tầng Backend)
- **Tách cơ sở dữ liệu (Database Split):** Dữ liệu nghiệp vụ giữ ở SQL. Dữ liệu Viễn trắc (Telemetry, Diffing logs) chuyển sang Time-series Database (Elasticsearch, ClickHouse).
- **Redis Pub/Sub:** Thay thế bộ nhớ RAM của `hub.go` bằng Redis để hỗ trợ Load Balancing cho WebSocket.

### Giai đoạn 3 (Tối ưu asset)
- **Cơ chế Phản ứng Chủ động (Active Response):** asset có khả năng tự động "Kill Process" hoặc "Isolate Network" ngay khi nhận lệnh từ WebSocket.