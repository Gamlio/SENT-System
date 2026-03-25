# Luồng Hoạt Động Chi Tiết - SENT Backend

Tài liệu này mô tả chi tiết luồng xử lý của backend hệ thống SENT, được xây dựng bằng Go. Kiến trúc được thiết kế theo mô hình nhiều lớp (multi-layer) để đảm bảo sự rõ ràng, dễ bảo trì và mở rộng.

---

## 1. Luồng Khởi Động (Application Bootstrap)

Luồng xử lý bắt đầu từ file entry-point của ứng dụng.

- **File:** `cmd/server/main.go`
- **Nhiệm vụ:**
    1.  **Tải Cấu Hình:** Đọc các biến môi trường (ví dụ: chuỗi kết nối database, secret key cho JWT) từ file `.env`.
    2.  **Kết Nối Database:** Gọi hàm từ `internal/database/db.go` để khởi tạo kết nối đến cơ sở dữ liệu (PostgreSQL) và chạy `AutoMigrate` để tự động tạo/cập nhật các bảng dựa trên các `models`.
    3.  **Khởi Tạo WebSocket Hub:** Chạy `hub` từ `internal/websocket/hub.go` trong một goroutine riêng để quản lý các kết nối real-time.
    4.  **Khởi Tạo Router (Gin):** Thiết lập một Gin router mới.
    5.  **Định Nghĩa Middleware:** Gắn các middleware toàn cục vào router, ví dụ: middleware cho CORS, logging.
    6.  **Định Nghĩa Routes:** Khai báo tất cả các API endpoint. Các route được nhóm theo chức năng và trỏ đến các hàm xử lý (handler) tương ứng trong `internal/api/v1/`.
    7.  **Khởi Chạy Server:** Lắng nghe và phục vụ các request HTTP trên một port đã định.

---

## 2. Luồng Xử Lý Một HTTP Request (Ví dụ: User Request)

Khi một người dùng đã đăng nhập yêu cầu lấy danh sách các sự cố (`GET /api/v1/incidents`).

1.  **Middleware (`internal/middleware/auth.go`):**
    *   Request đầu tiên đi qua middleware xác thực.
    *   Middleware trích xuất `JWT Token` từ header `Authorization`.
    *   Giải mã và xác thực token. Nếu token không hợp lệ hoặc hết hạn, request bị từ chối với lỗi `401 Unauthorized`.
    *   Nếu hợp lệ, thông tin người dùng (ví dụ: `user_id`, `role`) được lấy từ token và lưu vào `context` của request để các handler sau có thể sử dụng.

2.  **Handler (`internal/api/v1/incidents/incident_handlers.go`):**
    *   Hàm xử lý (ví dụ: `GetAllIncidents`) được gọi.
    *   Nó không chứa logic nghiệp vụ phức tạp. Nhiệm vụ chính là:
        *   Parse các tham số query (ví dụ: `?page=1&limit=10`).
        *   Gọi hàm tương ứng trong `Service Layer` (ví dụ: `incidentService.GetAll(...)`).

3.  **Service (`internal/service/incidents/incident_service.go`):**
    *   Đây là nơi chứa logic nghiệp vụ.
    *   Hàm `GetAll(...)` thực hiện các công việc như xây dựng câu truy vấn dựa trên các tham số, kiểm tra quyền hạn (nếu cần).
    *   Nó gọi đến `Repository Layer` (hoặc trực tiếp sử dụng ORM) để truy vấn cơ sở dữ liệu.

4.  **Database Interaction (`internal/models/models.go` & DB connection):**
    *   Service sử dụng các struct trong `models.go` (ví dụ: `Incident`) để tương tác với DB.
    *   Thực hiện câu lệnh `SELECT * FROM incidents...` để lấy dữ liệu.

5.  **Hồi Đáp (Response):**
    *   Dữ liệu từ DB được trả về cho `Service`.
    *   `Service` trả dữ liệu về cho `Handler`.
    *   `Handler` đóng gói dữ liệu thành cấu trúc JSON và trả về cho client với status `200 OK`.

---

## 3. Luồng Dữ Liệu từ Agent (Agent Data Ingestion)

Đây là luồng quan trọng nhất, xử lý dữ liệu từ các agent được cài đặt trên máy người dùng.

1.  **Agent Gửi Dữ Liệu:**
    *   `SENT_agent` trên máy client thu thập dữ liệu (USB, phần mềm, firewall...).
    *   Nó gửi một request `POST` đến endpoint, ví dụ: `/api/v1/agents/data`. Request này chứa một token xác thực riêng của agent.

2.  **Middleware (`internal/middleware/agent_auth.go`):**
    *   Middleware này được áp dụng riêng cho các route của agent.
    *   Nó xác thực token của agent, đảm bảo chỉ các agent hợp lệ mới được gửi dữ liệu.

3.  **Handler (`internal/api/v1/agents/receiver_handler.go`):**
    *   Nhận dữ liệu JSON thô từ agent.
    *   Gọi `data_service` để xử lý. Phản hồi cho agent ngay lập tức để giảm độ trễ. Việc xử lý sâu hơn được thực hiện bất đồng bộ.

4.  **Service (`internal/service/agents/data_service.go`):**
    *   Nhận dữ liệu và sử dụng các hàm helper trong `internal/service/agents/data/` (ví dụ: `antivirus.go`, `software.go`) để parse và chuẩn hóa dữ liệu theo từng loại.
    *   Lưu dữ liệu đã được xử lý vào database.

5.  **Event Engine (`internal/service/incidents/event_engine.go`):**
    *   Sau khi `data_service` lưu dữ liệu, nó có thể tạo ra một "sự kiện" và đẩy vào `Event Engine`.
    *   `Event Engine` so sánh sự kiện này với các `Policy` đã được định nghĩa trong hệ thống (lấy từ `policy_service`).
    *   Ví dụ: Sự kiện là "phát hiện USB mới cắm vào", Policy là "cấm tất cả USB lạ". `Event Engine` sẽ thấy sự trùng khớp.

6.  **Incident Creation (`internal/service/incidents/incident_service.go`):**
    *   Khi `Event Engine` phát hiện vi phạm, nó sẽ gọi `incident_service` để tạo một bản ghi `Incident` mới trong DB.

7.  **Real-time Notification (`internal/websocket/hub.go`):**
    *   `incident_service` sau khi tạo sự cố thành công sẽ gửi một thông điệp đến `WebSocket Hub`.
    *   `Hub` sẽ phát thông điệp này đến tất cả các client (trình duyệt của admin) đang kết nối.

---

## 4. Luồng Phê Duyệt (Approval Workflow)

Hệ thống áp dụng mô hình "Maker-Checker" để tăng cường bảo mật.

- **File chính:** `internal/service/approvals/approval_service.go` và các `strategies` trong cùng thư mục.
- **Luồng hoạt động:**
    1.  **Tạo yêu cầu:** Một "Maker" (ví dụ: user, hoặc agent tự động) thực hiện một hành động cần phê duyệt (ví dụ: đăng ký agent mới). Một request được gửi đến handler tương ứng.
    2.  **Tạo Ticket:** Thay vì thực hiện ngay, service sẽ tạo một bản ghi trong DB với trạng thái `PENDING` và tạo một `ApprovalTicket` liên kết với nó.
    3.  **Thông báo:** Admin ("Checker") nhận được thông báo (có thể qua UI real-time).
    4.  **Phê duyệt/Từ chối:** Admin vào `Approval Center`, xem ticket và nhấn "Approve" hoặc "Reject". Request được gửi đến `approval_handlers.go`.
    5.  **Thực thi:** `approval_service` nhận yêu cầu. Nó sử dụng `Strategy Pattern` (`agent_strategy.go`, `user_strategy.go`...) để biết hành động cụ thể cần làm khi ticket được duyệt. Ví dụ, với `agent_strategy`, nó sẽ cập nhật trạng thái của agent từ `PENDING` thành `ACTIVE`.
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
│   │   │   ├── strategies/
│   │   │   │     ├── agent_strategy.go      
│   │   │   │     ├── document_strategy.go   
│   │   │   │     ├── policy_strategy.go 
│   │   │   │     ├── strategy.go  
│   │   │   │     └── user_strategy.go
│   │   │   └── approval_handlers.go
│   │   ├── agents/
│   │   │   ├── enrollment_handlers.go   
│   │   │   ├── receiver_handler.go   
│   │   │   ├── dashboard_handler.go   
│   │   │   └── receiver_handler.go    
│   │   ├── policies/
│   │   │   ├── policy_engine.go   
│   │   │   └── policy_handlers.go   (Quản lý file AI Docs và Luật kỹ thuật)
│   │   ├── incidents/
│   │   │   └── incident_handlers.go
│   │   └── dashboard/
│   │       └── dashboard_handlers.go(Thống kê tổng quan) 
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
│   ├── utils/               # REPOSITORY: Truy vấn DB 
│   │   └── email.go 
│   ├── websocket/               # REPOSITORY: Truy vấn DB 
│   │   └── hub.go 
│   └── service/                  # SERVICE: Xử lý logic nghiệp vụ
│       ├── ai/         
│       │    └── chat_service.go 
│       ├── scoring/           
│       │    └── score_service.go 
│       ├── agent_data/     
│       │    ├── inventory.go      # Xử lý thông tin phần cứng
│       │    ├── software.go       # Xử lý danh sách phần mềm
│       │    ├── network.go            # Xử lý log USB      
│       │    ├── port.go      # Xử lý thông tin phần cứng
│       │    ├── usb.go            # Xử lý log USB
│       │    └── helpers.go      
│       ├── security/             # Chuyên logic nghiệp vụ an ninh
│       │   ├── ai_analysis.go    # Phân tích hành vi bằng AI
│       │   ├── alerts.go         # Logic tạo và quản lý cảnh báo
│       │   ├── compliance.go         
│       │   └── policies.go       # Kiểm tra chính sách (Whitelist/Blacklist)
│       └── processor.go          # File "điều phối" chính (Entry point)
├── pkg/                          # Các package dùng chung
├── uploads/                      # Thư mục chứa file vật lý (PDF, DOCX)
├── .env                          # SECRET_KEY, DB_URL, PORT
├── Dockerfile
├── go.mod
└── go.sum