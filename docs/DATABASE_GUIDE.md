# HỆ THỐNG CƠ SỞ DỮ LIỆU SENT v4.0 (Enterprise SOC Architecture)

Hệ thống sử dụng PostgreSQL với ORM Gorm, thiết kế theo chuẩn Multi-tenant (Đa khách hàng) và tối ưu hóa cho quy trình Auto-Triage (Tự động phân luồng sự cố).

## I. NHÓM QUẢN TRỊ GLOBAL (Core Admin)
1. **organizations**: Quản lý đa khách hàng/công ty.
2. **users**: Người dùng hệ thống (Admin, Trưởng ca SOC, Helpdesk).
3. **regions**: Phân vùng địa lý/chi nhánh của máy trạm.

## II. NHÓM QUẢN LÝ TÀI SẢN (Asset Management)
Lưu trữ thông tin định danh và trạng thái thiết bị từ Ninja Agent gửi về.
1. **agents** (Bảng chủ): 
   - `hw_id` (PK): Mã phần cứng duy nhất.
   - `status`: Trạng thái Zero-Trust (`PENDING`, `ACTIVE`, `REJECTED`, `ISOLATED`).
   - `risk_score`: Điểm rủi ro động (Tính toán realtime dựa trên Sự cố).
   - `device_type`: Chức vụ thiết bị (SERVER, IT_ADMIN, GUEST) để tính trọng số rủi ro.
2. **agent_inventories**: Cấu hình phần cứng (OS, CPU, RAM).

## III. NHÓM DỮ LIỆU VIỄN TRẮC (Telemetry & Logs)
Nhận dữ liệu thô từ Agent qua thuật toán Differential Reporting.
1. **software_items**: Danh sách phần mềm (Kèm `file_hash` và trạng thái `GHOST_REGISTRY`).
2. **usb_logs**: Lịch sử cắm USB (Kèm `vid`, `pid`, `device_hash`).
3. **open_ports**: Các cổng mạng đang mở (Phát hiện Tường lửa tắt, mở port 3389/22).

## IV. NHÓM VẬN HÀNH SOC (Incident Response)
Trái tim của hệ thống, áp dụng thuật toán Exact Matching Correlation.
1. **security_alerts**: Cảnh báo đơn lẻ (Log thô vi phạm). Luôn được gắn vào một `Incident` thông qua `incident_id`.
2. **incidents**: Hồ sơ Sự cố (Gom nhóm các Alert cùng loại trong 24h).
   - `type`: Tên gia tộc lỗi (VD: `Malware Detected`, `Firewall Disabled`).
   - `status`: Tiến trình xử lý (`Open`, `Investigating`, `Resolved`).
3. **incident_activities**: Nhật ký tương tác (Timeline).
   - Lưu trữ lệnh điều tra, File/Ảnh bằng chứng hậu kiểm (`images`), và báo cáo từ AI Copilot.

## V. NHÓM TRI THỨC & AI (AI RAG & Policies)
1. **universal_policies**: Bộ luật bảo mật do Admin cấu hình (Blacklist/Whitelist).
2. **policy_documents**: Tài liệu PDF/Word đã upload (Làm Knowledge Base cho AI).
3. **ai_chat_sessions** & **ai_chat_logs**: Lịch sử hỏi đáp giữa Admin và SENT Copilot.