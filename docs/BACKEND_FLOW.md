# SENT Backend Logic Flow (v4.0 - Centralized SOC & Auto-Triage)

Tài liệu mô tả luồng vận hành của Backend SENT. Kiến trúc được thiết kế theo tiêu chuẩn RESTful API, đóng vai trò là "Bộ não trung tâm" xử lý hàng triệu log từ các Ninja Agent, tự động phân luồng sự cố và cung cấp AI Cố vấn cho đội ngũ SOC.

## 1. Kiến trúc Xác thực & Phân quyền (RBAC)
Hệ thống sử dụng mô hình RBAC tinh gọn với 2 cấp độ:
- **Level 1 (Nhân viên/User):** Chỉ có quyền Read-only đối với thiết bị, được phép dùng AI Copilot để tra cứu thông tin nội bộ.
- **Level 2 (Quản trị viên/SOC Admin):** Full quyền Read/Write/Delete. Nhận cảnh báo, đóng/mở Case, thiết lập Policy và upload tài liệu tri thức (RAG).
- Mọi Request từ React đều phải kèm theo `JWT Token`.

## 2. Trạm điều phối Dữ liệu (Data Mediator)
- Nhận luồng dữ liệu JSON Viễn trắc (Telemetry, Software, USB) từ các Agent.
- Dùng cơ chế **Goroutine** (Xử lý bất đồng bộ) để parse JSON từ Router vào thẳng các module chuyên trách (`agent_data`).
- Check đối soát với Bảng `UniversalPolicy` (Luật bảo mật) và Cơ sở dữ liệu YARA/Threat Intel để phát hiện dị thường.

## 3. Kiến trúc Tri thức AI (Knowledge Base - /docs)
- Backend Go nhận file (PDF/Word) từ Admin, lưu trữ vật lý tại `uploads/policies/`.
- AI sử dụng dữ liệu này để làm RAG (Retrieval-Augmented Generation) trả lời nội quy cho nhân viên.

## 4. Động cơ Phân luồng Sự cố (Auto-Triage Engine)
Hệ thống áp dụng thuật toán **Exact Matching Correlation**:
- Không ném Alert (Cảnh báo) rời rạc cho Admin đọc.
- Gom nhóm các Alert có **Cùng Máy Trạm + Khớp 100% Loại Lỗi + Xảy ra trong 24h** vào thành 1 **Hồ sơ Sự cố (Incident/Case) duy nhất**.
- Tự động nâng cấp độ nghiêm trọng (Severity) của Case nếu xuất hiện log nguy hiểm hơn.

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
ÁC LUỒNG VẬN HÀNH CHÍNH (WORKFLOWS)
Luồng 1: Tiếp nhận và Lọc Dữ liệu (Ingestion Pipeline)
## 1. Kiến trúc Gate -> Brain (Separation of Concerns)
Hệ thống được tách bạch rạch ròi giữa tầng Tiếp nhận và tầng Xử lý:
- **API Gate (Handlers):** Chỉ làm nhiệm vụ xác thực HMAC, giải mã JSON thô và kiểm tra quyền cơ bản. Phản hồi Agent ngay lập tức (<10ms).
- **Service Brain (Processor):** Trung tâm điều phối duy nhất. Mọi dữ liệu từ Gate được đẩy vào đây để AI hoặc các module chuyên trách phân tích.

## 2. Luồng Zero-Trust & Whitelist-First
Thay vì chặn cái xấu (Blacklist), SENT chuyển sang chỉ cho phép cái tốt (Whitelist):
1. **Enrollment:** Máy mới mặc định ở trạng thái `PENDING`.
2. **Baseline Discovery:** Admin ra lệnh quét máy. Agent gửi toàn bộ thông tin "sạch" hiện tại lên.
3. **Lockdown:** Backend nạp Baseline vào bảng `whitelist_items`. Chuyển máy sang `is_zero_trust = true`.
4. **Enforcement:** Bất kỳ USB lạ (Hash khác) hoặc Software lạ (Publisher/Hash khác) xuất hiện -> Bắn Alert P1 và tự động tạo đơn phê duyệt.

## 3. Real-time Command Hub (WebSocket Push)
Thay thế cơ chế Pull cũ để đạt độ trễ mili giây:
- **Persistent Connection:** Agent duy trì kết nối WebSocket tới Server.
- **Instant Push:** Khi Admin nhấn "Duyệt" hoặc "Update Policy", Server đẩy lệnh JSON trực tiếp xuống HWID tương ứng qua Hub.
- **Workflow:** Admin Action -> DB Update -> WebSocket Dispatcher -> Agent Execution.

## 4. Multi-tenancy & Data Isolation
- Dữ liệu hoàn toàn cách ly theo `org_id`.
- Mọi truy vấn từ Handler/Service bắt buộc phải có điều kiện `WHERE org_id = ?` lấy từ Context của JWT.