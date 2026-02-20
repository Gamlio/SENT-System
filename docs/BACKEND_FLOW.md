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

## 📂 Cấu Trúc Mã Nguồn Thực Tế (Project Structure)
Hệ thống tuân thủ tiêu chuẩn **Standard Go Project Layout**.

```text
SENT_backend/
├── cmd/
│   └── server/
│       └── main.go               # Khởi chạy router, kết nối DB, định tuyến /docs và /policies
├── internal/
│   ├── api/v1/                   # HANDLER: Tiếp nhận HTTP request
│   │   ├── asset_handlers.go     # Tiếp nhận Log từ Agent, quản lý Whitelist
│   │   ├── auth_handlers.go      # Đăng nhập, cấp Token
│   │   ├── dashboard_handlers.go # API thống kê tổng quan
│   │   └── policy_handlers.go    # Xử lý file tài liệu (Docs) và luật (Universal Policy)
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
│       ├── asset_service.go      # Phân tích Log, check Differential
│       └── policy_service.go     # Xử lý lưu file vật lý an toàn (MkdirAll, WriteFile)
├── pkg/                          # Các package dùng chung
├── uploads/                      # Thư mục chứa file vật lý (PDF, DOCX)
├── .env                          # SECRET_KEY, DB_URL, PORT
├── go.mod
└── go.sum