# HỆ THỐNG CƠ SỞ DỮ LIỆU SENT v3.2 (Optimized for Multi-tenant)

## I. NHÓM 1: TỔ CHỨC GLOBAL (Multi-tenant Core)
1. **organizations**: Định danh doanh nghiệp thuê dịch vụ.
   - `company_code`: Mã đăng nhập độc nhất (VD: SME-A1B2C3).
   - `enroll_token_prefix`: Tiền tố cho chuỗi cài đặt Agent.
2. **regions**: Chi nhánh của từng công ty (dùng để phân cụm máy trạm).

## II. NHÓM 2: NGƯỜI DÙNG & PHÂN QUYỀN (Granular RBAC)
3. **users**: Tài khoản đăng nhập hệ thống.
   - Index: Khóa kết hợp `(org_id, username)` đảm bảo tính độc lập giữa các công ty.
   - `role`: Phân định "ADMIN" (Chủ công ty) và "USER" (Nhân viên).
   - Các trường boolean phân quyền: `can_manage_agents`, `can_manage_incidents`, `can_manage_docs`, `can_manage_users`.

## III. NHÓM 3: THIẾT BỊ (Asset & Telemetry)
4. **agents**: Bảng định danh máy trạm (Primary Key: `hw_id`).
5. **agent_inventories**: Cấu hình phần cứng (CPU, RAM, OS).
6. **software_items**: Danh sách phần mềm thu thập từ Agent.
7. **agent_snapshots**: Lưu trữ mã băm (Hash) để thực hiện Differential Reporting.

## IV. NHÓM 4: GIÁM SÁT AN NINH & SỰ CỐ (SOC/IR)
8. **incidents**: Hồ sơ điều tra sự cố tập trung.
   - Tổng hợp nhiều cảnh báo thuộc về 1 máy trạm.
   - `ai_analysis`: Trường lưu trữ kịch bản Playbook do AI (RAG) tự động sinh ra.
   - `status`: Quản lý luồng (Open / Resolved).
9. **security_alerts**: Cảnh báo kỹ thuật chi tiết (Bằng chứng số). Liên kết với bảng `incidents` qua `incident_id`.
10. **security_events**, **open_ports**, **usb_logs**: Các nhật ký thô (Raw logs) từ Endpoint.

## V. NHÓM 5: CHÍNH SÁCH TẬP TRUNG (Compliance & RAG)
11. **universal_policies**: Luật kỹ thuật dạng cứng (Cấm cổng, cấm phần mềm).
12. **policy_documents**: Hồ sơ tài liệu vật lý (PDF/Word). Đóng vai trò là Vector Knowledge Base để AI đọc và sinh Playbook.