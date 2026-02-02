Giai đoạn 1: Đăng ký (Enrollment):
- Nếu máy chưa có Agent_Secret_Key, Agent yêu cầu người dùng nhập Enroll_Token từ Admin cung cấp.
- Agent tự động thu thập mã phần cứng (HWID) và tên máy (Hostname) gửi về Backend để xin phê duyệt.
Giai đoạn 2: Giám sát (Monitoring):
- Sau khi được phê duyệt, Agent chạy ngầm theo chu kỳ được cấu hình trong file .yml.
- Quét cổng mạng (Network Ports), phần mềm đã cài đặt và trạng thái thiết bị ngoại vi (USB).
Giai đoạn 3: Truyền tải (Transporter):
- Dữ liệu được đóng gói dạng JSON và đẩy lên API Gateway qua HTTPS để đảm bảo an toàn thông tin.