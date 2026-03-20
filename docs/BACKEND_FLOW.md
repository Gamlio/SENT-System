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
│   │   │   └── approval_handlers.go.go  
│   │   ├── agents/
│   │   │   └── receiver_handler.go   
│   │   │   └── dashboard_handler.go   
│   │   │   └── receiver_handler.go    
│   │   ├── policies/
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
ÁC LUỒNG VẬN HÀNH CHÍNH (WORKFLOWS)
Luồng 1: Tiếp nhận và Lọc Dữ liệu (Ingestion Pipeline)
Agent gửi Payload (type: "DATA") lên API /api/v1/agents/push.

Hàm ProcessAgentData tại processor.go kiểm tra trạng thái Zero-Trust. Nếu máy tính đang bị khóa (PENDING/REJECTED), dữ liệu bị vứt bỏ.

Nếu hợp lệ, đẩy Payload.Data vào module tương ứng (VD: ProcessSoftware).

Module giải mã JSON, lưu CSDL hiện trạng thiết bị.

Luồng 2: Quy trình Xử lý Sự cố Thế hệ Mới (SOC Incident Lifecycle)
Đây là vòng đời chuẩn từ khi phát sinh lỗi đến khi Đóng Case:

Bước 1: Phát hiện (Detection)

Các file trong agent_data (VD: thấy mã băm lạ trong phần mềm) sẽ gọi hàm security.TriggerSecurityEvent().

Bước 2: Gom nhóm Tự động (Auto-Correlation)

Hệ thống tìm xem Nạn nhân (HWID) có Case nào mang tên Malware Detected đang mở không.

Nếu có: Nối cảnh báo (Alert) này vào Case đó. Làm mới khung thời gian 24h.

Nếu không: Lập một Case mới toanh có mã ID riêng.

Bước 3: Chấm điểm Rủi ro Động (Dynamic Risk Scoring)

Hàm scoring.RecalculateRiskScore() được gọi tự động.

Điểm rủi ro của máy trạm tăng lên dựa trên trọng số của Case nặng nhất đang Open. Giao diện đổi màu sang Vàng/Cam/Đỏ.

Bước 4: Điều tra bằng AI (AI Advisor)

Admin mở giao diện SOC, nhìn thấy Case. Thay vì phải đọc thủ công hàng chục cảnh báo, Admin bấm "Yêu cầu AI Phân tích".

API /api/v1/incidents/:id/ai-analyze thu thập toàn bộ log gom thành Prompt trói buộc (Không cho AI quyền thực thi) và gửi cho Ollama.

AI trả về báo cáo chuẩn format: [TÓM TẮT], [ĐÁNH GIÁ RỦI RO], [ĐỀ XUẤT XỬ LÝ].

Bước 5: Báo cáo Hậu kiểm & Đóng Case (Post-Mortem Resolution)

Quản trị viên xử lý xong sự cố, mở form nhập Báo cáo (Text) và đính kèm File/Ảnh Bằng chứng.

React gửi FormData (Multipart) lên API /activity.

Backend lưu Text và ghi File Ảnh vào ổ cứng (uploads/incidents/).

Trạng thái Case chuyển thành Resolved (Đã xử lý).

[QUAN TRỌNG] Backend tự động gọi lại hàm Risk Scoring. Lập tức điểm rủi ro của máy trạm tụt về 0 (Màu xanh an toàn).