#  SENT - AI-Powered Security Operations Center (SOC)
Hệ thống giám sát an ninh đầu cuối (EDR) và quản trị chính sách tập trung, tích hợp Trợ lý AI (Knowledge RAG) dành cho doanh nghiệp.

## Tính năng cốt lõi
1. **Trạm điều phối trung tâm (Backend Go-Gin):**
   - Xử lý viễn trắc (Telemetry) theo thời gian thực từ Agent.
   - Quản trị Rule-base tự động đối soát Phần mềm, Thiết bị USB và Mạng.
2. **Trung tâm Tri thức AI (Knowledge Base):**
   - Hỗ trợ nạp tài liệu (PDF/Word), tiêu chuẩn ISO 27001 để xây dựng Vector Database.
   - Cung cấp dữ liệu ngữ cảnh cho AI Security Assistant.
3. **Giao diện SOC chuyên nghiệp (ReactJS - Tailwind):**
   - Cấu trúc Component-based với Custom Hooks (`useAgents`, `usePolicies`).
   - Theo dõi trạng thái Online/Offline của thiết bị theo Real-time.
   - Quản lý chính sách bảo mật tập trung (Policy Center) áp dụng toàn cầu (Global) hoặc đích danh máy trạm (Specific).
4. **Agent tự động (Go):**
   - Chạy ngầm nhẹ nhàng, tự động thu thập Inventory, Software, Port và USB Logs.

## 🛠️ Công nghệ sử dụng
- **Backend:** Golang, Gin Framework, GORM, PostgreSQL, JWT, Bcrypt.
- **Frontend:** ReactJS, Tailwind CSS, Axios, Lucide Icons.
- **AI/ML Integration:** Tích hợp RAG (Retrieval-Augmented Generation) để tư vấn quy định nội bộ.
Luồng Nghiệp Vụ Cốt Lõi (Business Flow).

1. Luồng Xác thực Đa khách thuê (Multi-tenant Flow)
Đăng ký (Register): Người dùng tạo tổ chức mới. Backend tự sinh Company Code (VD: SME-A1B2C3) và trả về màn hình. User này trở thành Admin mặc định.

Đăng nhập (Login): Form hỗ trợ nhập mã công ty. Nếu Username tồn tại ở nhiều công ty khác nhau, hệ thống sẽ yêu cầu nhập Company Code để định tuyến chính xác vào không gian dữ liệu riêng biệt.

2. Luồng Quản lý Người dùng & Phân quyền (Granular RBAC)
Tại UserManagement.jsx, Admin cấp tài khoản cho nhân viên không dùng Level cứng, mà thông qua Ma trận Checkbox (VD: Chỉ cho Xem máy trạm, Không cho xử lý sự cố).

Context API lưu trữ cấu hình này, Sidebar tự động ẩn/hiện các menu (Agents, Incidents, Docs) một cách thông minh.

3. Luồng Điều tra Sự cố & AI Playbook (SOC Flow)
Dữ liệu vi phạm tuân thủ từ Endpoint tự động đổ về màn hình IncidentList.jsx theo thời gian thực.

Khi điều tra chi tiết (IncidentReport.jsx), hệ thống sử dụng thiết kế Tab:

Tổng quan: Thông tin HWID, Mức độ nghiêm trọng.

Chi tiết kỹ thuật: Raw logs, Event ID, Port/USB thực tế.

Playbook Ứng phó: Kịch bản xử lý tự động được sinh ra bởi AI (RAG) dựa trên việc đọc nội quy của công ty, giúp nhân viên SOC Level 1 biết chính xác cần bấm nút gì, thao tác ra sao.

4. Luồng Tri thức AI (Docs)
Admin tải tài liệu chuẩn hóa (ISO, Nội quy IT) lên /docs.

Nhân viên có thể dùng AIChatAssistant.jsx để tra cứu nhanh, hoặc AI sẽ ngầm sử dụng kho dữ liệu này để sinh Playbook cho luồng số 3.