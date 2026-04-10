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

## 3. Luồng Dữ Liệu từ asset (asset Data Ingestion)

Đây là luồng quan trọng nhất, xử lý dữ liệu từ các asset được cài đặt trên máy người dùng.

1.  **asset Gửi Dữ Liệu:**
    *   `SENT` trên máy client thu thập dữ liệu (USB, phần mềm, firewall...).
    *   Nó gửi một request `POST` đến endpoint, ví dụ: `/api/v1/assets/data`. Request này chứa một token xác thực riêng của asset.

2.  **Middleware (`internal/middleware/asset_auth.go`):**
    *   Middleware này được áp dụng riêng cho các route của asset.
    *   Nó xác thực token của asset, đảm bảo chỉ các asset hợp lệ mới được gửi dữ liệu.

3.  **Handler (`internal/api/v1/assets/receiver_handler.go`):**
    *   Nhận dữ liệu JSON thô từ asset.
    *   Gọi `data_service` để xử lý. Phản hồi cho asset ngay lập tức để giảm độ trễ. Việc xử lý sâu hơn được thực hiện bất đồng bộ.

4.  **Service (`internal/service/assets/data_service.go`):**
    *   Nhận dữ liệu và sử dụng các hàm helper trong `internal/service/assets/data/` (ví dụ: `antivirus.go`, `software.go`) để parse và chuẩn hóa dữ liệu theo từng loại.
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
    1.  **Tạo yêu cầu:** Một "Maker" (ví dụ: user, hoặc asset tự động) thực hiện một hành động cần phê duyệt (ví dụ: đăng ký asset mới). Một request được gửi đến handler tương ứng.
    2.  **Tạo Ticket:** Thay vì thực hiện ngay, service sẽ tạo một bản ghi trong DB với trạng thái `PENDING` và tạo một `ApprovalTicket` liên kết với nó.
    3.  **Thông báo:** Admin ("Checker") nhận được thông báo (có thể qua UI real-time).
    4.  **Phê duyệt/Từ chối:** Admin vào `Approval Center`, xem ticket và nhấn "Approve" hoặc "Reject". Request được gửi đến `approval_handlers.go`.
    5.  **Thực thi:** `approval_service` nhận yêu cầu. Nó sử dụng `Strategy Pattern` (`asset_strategy.go`, `user_strategy.go`...) để biết hành động cụ thể cần làm khi ticket được duyệt. Ví dụ, với `asset_strategy`, nó sẽ cập nhật trạng thái của asset từ `PENDING` thành `ACTIVE`.

---

## 5. Chi Tiết Các Loại Dữ Liệu Giám Sát Cụ Thể

Hệ thống thu thập và xử lý 8 loại dữ liệu giám sát chính từ assets, được lưu trữ trong MongoDB để xử lý dữ liệu lớn.

### 5.1 Software Inventory (SoftwareItem)
- **Mục đích:** Theo dõi phần mềm cài đặt trên máy asset
- **Dữ liệu thu thập:** Tên phần mềm, phiên bản, nhà phát hành, đường dẫn cài đặt, hash file, trạng thái (INSTALLED/GHOST_REGISTRY), trạng thái running
- **Luồng xử lý:** asset gửi danh sách → `data_service` parse → Lưu MongoDB → Hiển thị trên dashboard asset
- **File liên quan:** `internal/service/assets/data/software.go`, `mongo_models.go`

### 5.2 Open Ports Monitoring (OpenPort)
- **Mục đích:** Giám sát cổng mạng mở trên máy asset
- **Dữ liệu thu thập:** Số port, tên process đang sử dụng, trạng thái (OPEN/CLOSED)
- **Luồng xử lý:** asset scan ports → Gửi dữ liệu → `data_service` validate → Lưu MongoDB → Hiển thị trên asset details
- **File liên quan:** `internal/service/assets/data/port.go`, `mongo_models.go`

### 5.3 USB Device Logging (USBLog)
- **Mục đích:** Ghi log cắm/rút thiết bị USB
- **Dữ liệu thu thập:** Tên thiết bị, device ID, VID/PID, serial number, device hash, event type (CONNECT/DISCONNECT), whitelist status
- **Luồng xử lý:** asset detect USB events → Gửi log → `data_service` check whitelist → Tạo incident nếu vi phạm → Lưu MongoDB
- **File liên quan:** `internal/service/assets/data/usb.go`, `mongo_models.go`

### 5.4 I/O Activity Monitoring (assetIOActivity)
- **Mục đích:** Theo dõi hoạt động truyền tải dữ liệu
- **Dữ liệu thu thập:** Bytes sent/recv mạng, bytes written/read ổ đĩa, timestamp
- **Luồng xử lý:** asset thu thập metrics → Gửi định kỳ → `data_service` aggregate → Lưu MongoDB → Hiển thị charts trên dashboard
- **File liên quan:** `internal/service/assets/data/network.go`, `mongo_models.go`

### 5.5 Hardware Inventory (assetInventory)
- **Mục đích:** Thu thập thông tin phần cứng máy asset
- **Dữ liệu thu thập:** Model CPU, RAM total, thông tin OS
- **Luồng xử lý:** asset thu thập lúc khởi động → Gửi enrollment → Lưu MongoDB → Hiển thị trên asset profile
- **File liên quan:** `internal/service/assets/data/inventory.go`, `mongo_models.go`

### 5.6 Security Alerts (SecurityAlert)
- **Mục đích:** Cảnh báo bảo mật từ asset hoặc hệ thống
- **Dữ liệu thu thập:** Priority (P1-P4), alert type, title, description, severity (Low-Critical), trạng thái resolved
- **Luồng xử lý:** Event engine detect → Tạo alert → Link với incident → Lưu MongoDB → Push notification real-time
- **File liên quan:** `internal/service/security/alerts.go`, `mongo_models.go`

### 5.7 Incident Activities (IncidentAudit)
- **Mục đích:** Timeline chi tiết quá trình xử lý incident
- **Dữ liệu thu thập:** Action type (COMMENT, STATUS_CHANGE, AI_ANALYSIS), content, old/new status, images, user info
- **Luồng xử lý:** User actions → Tạo activity log → Lưu MongoDB → Hiển thị timeline trong incident details
- **File liên quan:** `internal/service/incidents/incident_service.go`, `mongo_models.go`

### 5.8 AI Chat Logs (AIChatSession, AIChatLog)
- **Mục đích:** Lưu trữ lịch sử trò chuyện với AI assistant
- **Dữ liệu thu thập:** Session info, user messages, AI responses, thought process
- **Luồng xử lý:** User chat → AI process → Lưu MongoDB → Hiển thị chat history
- **File liên quan:** `internal/service/ai/chat_service.go`, `mongo_models.go`

---

## 6. Nghiệp Vụ Quản Lý Tổ Chức (Organization/Region)

Hệ thống hỗ trợ quản lý đa tổ chức với cấu trúc phân cấp.

- **Organization:** Đơn vị công ty cao nhất, chứa users, regions, policies
- **Region:** Vùng địa lý trong tổ chức, chứa assets với enrollment token riêng
- **Luồng hoạt động:**
  1. Admin tạo organization với company code và enroll token prefix
  2. Tạo regions với token riêng cho từng khu vực
  3. assets sử dụng token region để enroll
  4. Policies và users được scope theo organization
- **File liên quan:** `models.go` (Organization, Region), `internal/api/v1/users/`, `internal/service/assets/lifecycle_service.go`

---

## 7. Hệ Thống Permissions Chi Tiết Của User

Hệ thống sử dụng role-based access control với 8 permissions cụ thể:

- **PermassetView/Action/Delete:** Xem/thao tác/xóa assets
- **PermPolicyView/Action:** Xem/quản lý policies
- **PermIncidentView/Action:** Xem/xử lý incidents
- **PermDocView/Manage:** Xem/quản lý documents
- **PermUserManage:** Quản lý users
- **PermApprovalManage:** Phê duyệt requests

- **Luồng xử lý:**
  1. User login → JWT token chứa permissions
  2. Middleware check permissions cho từng API
  3. Service validate quyền trước khi thực thi
  4. UI ẩn/hiện features theo permissions
- **File liên quan:** `models.go` (User struct), `internal/middleware/auth.go`, `internal/auth/security.go`

---

## 8. asset Lifecycle Management

Quản lý vòng đời asset từ enrollment đến decommission.

- **Trạng thái:** PENDING → APPROVED → ACTIVE → SUSPENDED
- **Luồng enrollment:**
  1. asset gửi request với AssetHWID , hostname, IP, token
  2. Validate token region → Tạo record PENDING
  3. Tạo approval ticket → Admin approve
  4. Status chuyển ACTIVE → Gửi secret key
- **Baseline & Trust Score:**
  - Baseline: Snapshot ban đầu sau enrollment
  - Trust Score: 0-100, giảm khi vi phạm, tự phục hồi theo thời gian
- **File liên quan:** `internal/api/v1/assets/enrollment_handlers.go`, `internal/service/assets/lifecycle_service.go`, `models.go`

---

## 9. Incident Management Workflow

Luồng xử lý sự cố end-to-end với AI support.

- **Tạo incident:** Từ alerts, policy violations, hoặc manual
- **Phân loại:** Type (Malware, DDoS), Severity (Low-Critical), Priority (P1-P4)
- **Workflow:**
  1. Tạo incident → Assign assignee
  2. Investigation → AI analysis
  3. Status updates: Open → Investigating → Resolved
  4. Resolution summary khi đóng case
- **Timeline:** Mọi action được log trong IncidentAudit
- **File liên quan:** `internal/service/incidents/incident_service.go`, `internal/api/v1/incidents/`, `models.go`

---

## 10. Universal Policy System

Hệ thống chính sách tập trung cho tất cả loại threats.

- **Policy Types:** SOFTWARE, USB, NETWORK
- **Policy Values:** BLACKLIST/WHITELIST
- **Target Scope:** GLOBAL hoặc specific AssetHWID s
- **Luồng thực thi:**
  1. Admin tạo policy → Status PENDING
  2. Approval workflow → Active
  3. Event engine check violations
  4. Tạo incident nếu match
- **File liên quan:** `models.go` (Policy), `internal/service/policies/`, `internal/api/v1/policies/`

---

## 11. Policy Documents Management

Quản lý tài liệu chính sách và quy trình.

- **Upload Process:** Word/PDF → Convert to PDF → Store files
- **Approval:** Maker-Checker cho documents
- **AI Processing:** Parse nội dung → Extract rules → Auto-create policies
- **File liên quan:** `models.go` (Document), `internal/api/v1/policies/policy_handlers.go`, `uploads/`

---

## 12. AI Chat Functionality

Trợ lý AI tích hợp cho analysis và support.

- **Features:** Chat về incidents, policy suggestions, security analysis
- **Architecture:** Sessions + Logs, context-aware responses
- **Integration:** Với incident management, policy engine
- **File liên quan:** `internal/service/ai/chat_service.go`, `internal/api/v1/ai/`, `mongo_models.go`

---

## 13. WebSocket Real-time Notifications

Hệ thống thông báo real-time cho dashboard.

- **Hub Architecture:** Central hub quản lý connections
- **Events:** asset status changes, new incidents, approvals
- **Broadcast:** To all connected clients theo permissions
- **File liên quan:** `internal/websocket/hub.go`, `internal/service/incidents/incident_service.go`

---

## 14. Scoring/Risk Calculation Logic

Hệ thống tính điểm rủi ro động.

- **Risk Score:** asset-level risk (0-100)
- **Factors:** Software violations, USB events, alerts, trust score
- **Calculation:** Weighted algorithm, real-time updates
- **Recovery:** Trust score tự tăng theo thời gian không vi phạm
- **File liên quan:** `internal/service/scoring/score_service.go`, `models.go` (asset.RiskScore)

---
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
│   │   │   │     ├── asset_strategy.go      
│   │   │   │     ├── document_strategy.go   
│   │   │   │     ├── policy_strategy.go 
│   │   │   │     ├── strategy.go  
│   │   │   │     └── user_strategy.go
│   │   │   └── approval_handlers.go
│   │   ├── assets/
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
│   │   └── models.go             # Struct DB (asset, Policy, Document...)
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
│       ├── asset_data/     
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