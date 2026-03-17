# SENT Backend Logic Flow (Go-Gin & PostgreSQL)

Tài liệu mô tả luồng vận hành của Backend SENT. Kiến trúc được thiết kế theo tiêu chuẩn RESTful API, phục vụ như một trạm điều phối trung tâm giữa Agent, Web Dashboard và AI Engine.

## 1. Kiến trúc Xác thực & Phân quyền (RBAC)
Hệ thống sử dụng mô hình RBAC tinh gọn với 2 cấp độ:
- **Level 1 (Nhân viên/User):** Chỉ có quyền Read-only đối với thiết bị, được phép dùng AI Copilot để tra cứu thông tin nội bộ.
- **Level 2 (Quản trị viên/Admin):** Full quyền Read/Write/Delete. Có thể nạp tài liệu cho AI học (Docs), cấu hình luật kỹ thuật (Policy Center). (Chức năng Quản trị SME tạm ẩn).
- Mọi Request từ React đều phải kèm theo `JWT Token`. Middleware của Go sẽ tự động giải mã và chặn các request trái phép.

## 2. Trạm điều phối Dữ liệu (Data Mediator)
- Nhận luồng dữ liệu viễn trắc từ các Agent.
- Dùng `GORM` để mapping dữ liệu vào cấu trúc bảng của PostgreSQL một cách an toàn.
- Xử lý đối soát: So sánh ngay với bảng `universal_policies` để sinh ra cảnh báo.

## 3. Kiến trúc Tri thức AI (Knowledge Base - /docs)
- Backend Go nhận file (PDF/Word) từ Admin, lưu trữ vật lý tại thư mục `uploads/policies/`.
- Lưu thông tin Metadata vào bảng `policy_documents`.

## 4. Trung tâm Chính sách (Policy Center - /policies)
- Quản lý tập trung mọi quy tắc kỹ thuật (Software, USB, Network) trong bảng `universal_policies`.
- Hỗ trợ cơ chế phân phối linh hoạt: Toàn cầu (GLOBAL) hoặc từng máy trạm (SPECIFIC HWIDs).

---
## 6. Kiến trúc Phê duyệt Tập trung (Approval Center / Maker-Checker)
Để đảm bảo an toàn tuyệt đối (Zero Trust), mọi thay đổi quan trọng trong hệ thống đều phải đi qua luồng phê duyệt 2 lớp:
- **Nguyên lý:** Khi có tác vụ nhạy cảm (Cài mới Agent, Thêm luật cấm USB, Xóa tài liệu), hệ thống KHÔNG áp dụng ngay mà sinh ra một `ApprovalTicket` (Vé chờ duyệt) với trạng thái `PENDING`.
- **Bảo vệ Agent:** Agent mới kết nối sẽ có trạng thái `PENDING` và bị giam trong vùng cách ly. Chỉ khi Admin bấm `APPROVED`, Agent mới chuyển sang `ACTIVE` và được phép nhận Policy hoặc gửi Log.
- **Workflow:** 1. `Maker` (Nhân viên SOC / Hoặc tự động từ Agent) -> Gọi API thêm mới -> Backend tạo Record (Status: PENDING) + Tạo `ApprovalTicket`.
  2. `Checker` (Trưởng ca SOC / Admin) -> Gọi API `/approvals/:id/review` -> Đổi Status của Ticket thành `APPROVED` -> Trigger update Status của Record gốc thành `ACTIVE/APPROVED`.
## 📂 Cấu Trúc Mã Nguồn Thực Tế (Project Structure)
Hệ thống tuân thủ tiêu chuẩn **Standard Go Project Layout**.

```text
SENT_backend/
├── cmd/
│   └── server/
│       └── main.go               # Khởi chạy router, kết nối DB, định tuyến /docs và /policies
├── internal/
│   ├── api/v1/                   # HANDLER: Tiếp nhận HTTP request
│   │   ├── auth/
│   │   │   └── auth_handlers.go     (Xử lý Login, Register)
│   │   ├── users/
│   │   │   └── user_handlers.go     (Đổi tên từ admin_handlers.go, chứa CreateUser, GetUsers...)
│   │   ├── approvals/
│   │   │   └── approval_handlers.go.go  
│   │   ├── agents/
│   │   │   └── agent_handlers.go    (Đổi tên từ asset_handlers.go)
│   │   ├── policies/
│   │   │   └── policy_handlers.go   (Quản lý file AI Docs và Luật kỹ thuật)
│   │   ├── incidents/
│   │   │   └── incident_handlers.go
│   │   └── dashboard/
│   │       └── dashboard_handlers.go(Thống kê tổng quan)   # Xử lý file tài liệu (Docs) và luật (Universal Policy)
│   ├── auth/
│   │   └── security.go                      # Cấu hình xác thực mở rộng
│   ├── database/           
│   │   └── db.go                 # Kết nối PostgreSQL, AutoMigrate
│   ├── middleware/           
│   │   └── auth.go               # Check JWT và Role Level
│   ├── models/               
│   │   └── models.go             # Struct DB (Agent, UniversalPolicy, PolicyDocument...)
│   ├── repository/               # REPOSITORY: Truy vấn DB 
│   │   └── policy_repo.go        
│   └── service/                  # SERVICE: Xử lý logic nghiệp vụ
│       ├── ai/         
│       │    └── chat_service.go 
│       ├── scoring/           
│       │    └── score_service.go 
│       ├── agent_data/           
│       │    ├── inventory.go      # Xử lý thông tin phần cứng
│       │    ├── software.go       # Xử lý danh sách phần mềm
│       │    ├── usb.go            # Xử lý log USB
│       │    └── telemetry.go      # Xử lý Port và mạng
│       ├── security/             # Chuyên logic nghiệp vụ an ninh
│       │   ├── ai_analysis.go    # Phân tích hành vi bằng AI
│       │   ├── alerts.go         # Logic tạo và quản lý cảnh báo
│       │   └── policies.go       # Kiểm tra chính sách (Whitelist/Blacklist)
│       └── processor.go          # File "điều phối" chính (Entry point)
├── pkg/                          # Các package dùng chung
├── uploads/                      # Thư mục chứa file vật lý (PDF, DOCX)
├── .env                          # SECRET_KEY, DB_URL, PORT
├── go.mod
└── go.sum
