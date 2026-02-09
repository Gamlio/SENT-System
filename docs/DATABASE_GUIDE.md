# HỆ THỐNG CƠ SỞ DỮ LIỆU SENT v3.2 (Optimized for Multi-tenant)

## I. NHÓM QUẢN TRỊ GLOBAL (Level 1 & 2)
1. **organizations**: Lưu thông tin doanh nghiệp SME.
   - `id`: Primary Key.
   - `name`: Tên công ty.
   - `enroll_token_prefix`: Tiền tố mã cài đặt riêng cho từng công ty.
2. **license_plans**: Định nghĩa các gói giới hạn thiết bị.
   - `max_agents`: Số máy tối đa được quản lý.

## II. NHÓM NGƯỜI DÙNG & PHÂN QUYỀN (R1-R4)
3. **users**: Tài khoản đăng nhập.
   - `org_id`: Khóa ngoại (NULL nếu là Level 1, 2).
   - `role_id`: Cấp độ R1, R2, R3, hoặc R4.
   - `totp_secret`: Mã bí mật cho 2FA.
4. **user_permissions**: Bảng phân quyền chi tiết (Trạm lọc logic).
   - `user_id`: Người được gán quyền.
   - `region_id`: Vùng được phép quản lý (Cực kỳ quan trọng cho R4).
   - `permission_code`: Mã quyền (ví dụ: 'VIEW_LOGS', 'BLOCK_USB').

## III. NHÓM THIẾT BỊ & AGENT (Asset Management)
5. **agents**: Bảng định danh máy trạm.
   - `hwid`: Mã phần cứng duy nhất thu thập từ Agent.
   - `status`: Online/Offline dựa trên Heartbeat.
6. **agent_inventory**: Thông tin tĩnh (Quét chậm).
   - `cpu_model`, `ram_total`, `os_info`.
7. **software_inventory**: Danh sách phần mềm đã cài.
   - `software_name`, `version`.
8. **agent_snapshots**: Lưu Hash để so sánh sai khác (Differential Reporting).
   - `software_hash`, `port_hash`: Dùng để Agent so sánh trước khi gửi.

## IV. NHÓM GIÁM SÁT AN NINH (Security Logs)
9. **usb_logs**: Nhật ký thiết bị ngoại vi thực tế.
   - `device_id`: VID/PID/Serial thực tế.
10. **open_ports**: Danh sách cổng và tiến trình.
    - `port_number`, `process_name`.
11. **security_events**: Log từ Windows Event Log.
    - `event_id`, `message` (Logon, Privileges).
12. **security_alerts**: Cảnh báo bắn về Dashboard Level 3/4.