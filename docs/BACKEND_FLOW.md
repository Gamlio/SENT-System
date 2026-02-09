# SENT Backend Logic Flow (Go-Gin Edition)

Tài liệu này mô tả luồng vận hành chi tiết của hệ thống Backend SENT, được tối ưu hóa cho mô hình Multi-tenant (đa công ty) và quản lý thiết bị quy mô lớn.

---

## 📂 Cấu Trúc Mã Nguồn (Project Structure)
Hệ thống tuân thủ tiêu chuẩn **Standard Go Project Layout** để đảm bảo tính mở rộng.

```text
SENT_backend/
├── cmd/
│   └── server/
│       └── main.go           # Khởi chạy router, kết nối DB
├── internal/
│   ├── api/v1/               # HANDLER: Tiếp nhận HTTP request
│   │   ├── auth_handlers.go  # Đăng nhập 2FA, cấp Token
│   │   ├── asset_handlers.go # Tiếp nhận Log từ Agent (Mediator Entry)
│   │   └── admin_handlers.go # R1, R3 quản lý R2, R4 và Rules
│   ├── service/              # SERVICE: "Trạm trung gian" xử lý logic
│   │   ├── asset_service.go  # Phân tích Log, check Differential
│   │   ├── auth_service.go   # Logic xác thực, phân quyền R1-R4
│   │   └── policy_service.go # Xử lý các Rules/Chính sách cho từng Org
│   ├── repository/           # REPOSITORY: Truy vấn DB (Lọc org_id cứng)
│   │   ├── asset_repo.go     # GORM queries cho Assets/Logs
│   │   └── user_repo.go      # Quản lý User và Permissions theo Org
│   ├── models/               
│   │   └── models.go         # Định nghĩa Struct DB (Hệ thống 18 bảng)
│   ├── middleware/           
│   │   ├── auth.go           # Check JWT và Level (R1-R4)
│   │   └── tenant.go         # Đảm bảo request luôn có OrgID/RegionID
│   └── pkg/                  # Các công cụ dùng chung (TOTP, Bcrypt, JWT)
├── .env                      # SECRET_KEY, DB_URL
└── go.mod

1. Vai trò Mediator (Bộ điều phối trung tâm)
Backend Go đóng vai trò là trạm trung chuyển dữ liệu duy nhất trong hệ thống.
$$Tiếp nhận: Mọi yêu cầu từ React Frontend và Go Agent đều đổ về đây qua giao thức HTTPS.

$$Xác thực: Kiểm tra Token (JWT cho User hoặc Enrollment Token cho Agent) trước khi cho phép dữ liệu đi sâu vào lớp Service.

$$Điều phối: Phân loại Log (Inventory, Telemetry, Software, Events) và lưu trữ vào các bảng tương ứng trong PostgreSQL.

2. Luồng Xác thực & Bảo mật (Auth Flow)
Hệ thống áp dụng cơ chế xác thực đa tầng để đảm bảo an toàn tuyệt đối.Hệ thống áp dụng cơ chế xác thực đa tầng để đảm bảo an toàn tuyệt đối.
$$Bước 1 - Mật khẩu: Sử dụng thư viện bcrypt nguyên bản của Go để kiểm tra mật khẩu đã băm.$$Bước 2 - 2FA (TOTP): Đối với các cấp quản trị (R1, R2, R3), hệ thống yêu cầu mã OTP từ ứng dụng xác thực (như Google Authenticator).
$$Bước 3 - Cấp quyền: Trả về JWT Token chứa thông tin định danh:$$\text{Payload} = \{ \text{user\_id, org\_id, role\_level, iat, exp} \}
$$Role-level: Xác định rõ quyền hạn từ R1 (Global Admin) xuống R4 (SME Staff).

3. Luồng Đa công ty (Multi-tenancy Isolation)
Đảm bảo hàng chục công ty dùng chung hệ thống mà dữ liệu không bị chồng chéo.
$$Middleware Lọc: Mọi Request đều đi qua trạm lọc tenant.go để trích xuất org_id từ Token.

$$Logic Isolation: Tại lớp Repository, mọi truy vấn SQL đều được đính kèm điều kiện lọc cứng: SELECT * FROM agents WHERE org_id = ? AND region_id = ?

$$R4 Region Filter: Nhân viên Level 4 chỉ được truy cập dữ liệu thuộc các Region_ID (Vùng) được gán quyền trong bảng user_permissions.

4. Luồng xử lý dữ liệu Agent (Agent Telemetry Flow)
Tối ưu hóa hiệu năng bằng cơ chế Báo cáo sai khác (Differential Reporting).
A. Giai đoạn Enrollment:
$$Agent gửi gói tin đầu tiên kèm Enroll_Token.
$$Backend đối soát Token để tự động gán máy trạm vào đúng Org_ID và Region_ID.
B. Giai đoạn Xử lý Log & Tối ưu lưu trữ:
$$So sánh Snapshot: Backend lấy mã Hash cũ từ bảng agent_snapshots. Nếu $Hash_{new} = Hash_{old}$, Backend chỉ cập nhật last_seen mà không ghi đè dữ liệu bảng Inventory/Software.
$$Xử lý Sự kiện (Events): Các sự kiện bảo mật từ Windows Event Log được đưa vào lớp phân tích Rules để phát hiện hành vi bất thường (như Logon sai nhiều lần).

5. Quy trình Phân quyền & Quản lý Rule (R1,3 ➔ R2,4)
$$Cấp quyền: R1 và R3 sử dụng giao diện quản lý để tạo bản ghi trong bảng user_permissions.

$$Gán Vùng: R3 gán danh sách Region_ID cho R4. Khi R4 gọi API, Backend chỉ trả về các máy thuộc các vùng đó.

$$Thực thi Chính sách: R3 thiết lập các Rules (ví dụ: Chặn USB lạ). Rules này được lưu vào DB và gửi ngược lại cho Agent trong bản tin phản hồi của Heartbeat.
## 7. Cơ chế Tuân thủ & Truy cứu trách nhiệm (Compliance Flow)
Đây là tầng logic quan trọng nhất để R3 (SME Admin) quản lý kỷ luật thiết bị.

### A. Quản lý USB (USB Accountability)
1. **Thiết lập**: R3 đăng ký các mã `device_id` (VID/PID/Serial) vào bảng `usb_whitelist` và gán tên nhân viên cụ thể.
2. **Đối soát**: Khi Agent gửi log `telemetry` (chứa `usb_devices`), Backend sẽ so sánh ID thực tế với Whitelist của công ty đó.
3. **Truy cứu**: Nếu phát hiện sai khác, hệ thống tạo `security_alerts` mức độ **High**. R3 có thể xuất báo cáo bằng chứng bao gồm: Máy vi phạm, Thời gian cắm, và ID của thiết bị lạ.

### B. Quản lý Phần mềm (Software Compliance)
1. **Chính sách**: R3 định nghĩa danh sách phần mềm bị cấm (Prohibited) hoặc yêu cầu phiên bản tối thiểu.
2. **Phát hiện**: Lớp `policy_service.go` sẽ quét danh sách `software_inventory` gửi về. Bất kỳ phần mềm nào nằm trong danh sách cấm sẽ bị đánh dấu vi phạm ngay lập tức.
6. Cấu trúc Triển khai (Docker Deployment)
Hệ thống được đóng gói hoàn toàn bằng Docker để chạy Server thật ổn định.

$$Container 1: Go Backend (Mediator).

$$Container 2: PostgreSQL (Hệ thống 18 bảng dữ liệu).

$$Container 3: Nginx (Proxy ngược & SSL).