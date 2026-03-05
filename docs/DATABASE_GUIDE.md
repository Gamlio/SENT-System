# HỆ THỐNG CƠ SỞ DỮ LIỆU SENT v3.2 (Optimized for Multi-tenant)
Hệ thống sử dụng PostgreSQL với ORM Gorm. Dưới đây là lược đồ CSDL cập nhật phục vụ cho hệ thống giám sát đa nền tảng.

## I. NHÓM QUẢN TRỊ (Core Admin)
1.  **organizations**: Quản lý đa khách hàng (Multi-tenant).
2.  **users**: Người dùng hệ thống (Admin/Staff).
3.  **regions**: Phân vùng địa lý/chi nhánh của máy trạm.

## II. NHÓM GIÁM SÁT THIẾT BỊ (Device Monitoring)
Đây là nhóm bảng cốt lõi, lưu trữ dữ liệu từ Agent gửi về.

### 1. agents (Bảng chủ)
Lưu thông tin định danh và trạng thái sống.
* `hw_id` (PK): Mã phần cứng duy nhất (VD: UUID của Mainboard).
* `hostname`: Tên máy tính.
* `ip_address`: IP LAN (Được cập nhật bởi Telemetry Service).
* `status`: Trạng thái (online/offline).
* `last_seen`: Thời điểm cuối cùng nhận được gói tin bất kỳ.

### 2. agent_inventories
Lưu cấu hình phần cứng tĩnh.
* `os_info`: Hệ điều hành (VD: "Windows 11 Pro", "Ubuntu 22.04 LTS").
* `cpu_model`: Tên Chip xử lý.
* `ram_total_gb`: Dung lượng RAM thực tế.

### 3. open_ports (Telemetry)
Lưu trạng thái mạng thời gian thực.
* `port`: Số hiệu cổng đang mở (Listen).
* `process_name`: Tên tiến trình chiếm dụng cổng (VD: `sshd`, `nginx`, `svchost.exe`).

### 4. software_items
Lưu danh sách phần mềm đã cài đặt.
* `software_name`: Tên ứng dụng.
* `version`: Phiên bản.
* *Lưu ý:* Bảng này được làm mới (Refresh) hoàn toàn mỗi khi Agent báo cáo thay đổi.

### 5. agent_snapshots
Bảng phụ trợ kỹ thuật, dùng để tối ưu băng thông.
* `last_port_hash`: Mã băm của lần gửi Port cuối cùng.
* `last_software_hash`: Mã băm của lần gửi Software cuối cùng.

## III. NHÓM AN NINH & SỰ CỐ (Security)

### 1. usb_logs
Lưu lịch sử kết nối thiết bị ngoại vi.
* `device_name`: Tên thiết bị (VD: "Kingston DataTraveler").
* `device_id`: VID/PID của USB.
* `event_type`: "plugged" (cắm) hoặc "unplugged" (rút).
* `is_whitelisted`: Cờ đánh dấu thiết bị có nằm trong danh sách tin cậy không.

### 2. incidents & security_alerts
* **security_alerts:** Các cảnh báo lẻ tẻ (VD: Cắm USB lạ, Phần mềm đen).
* **incidents:** Sự cố tổng hợp (Gom nhiều alert lại để xử lý theo quy trình).
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

