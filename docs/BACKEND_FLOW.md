Vai trò Mediator: Tiếp nhận mọi yêu cầu từ Agent (Go) và Frontend (React), sau đó điều phối vào Database PostgreSQL.

Luồng Đăng nhập (Auth Flow):
- Tiếp nhận thông tin định danh: username, password, captcha, và otp (Mã tạm 000000).
- Xác thực mật khẩu: Sử dụng thư viện bcrypt thuần để kiểm tra mật khẩu đã băm (Fix lỗi 72 bytes trên Python 3.13).
- Cấp quyền: Trả về JWT Token kèm theo level (1: Global Admin, 2: SME Admin) và org_id.
Luồng Đa công ty (Multi-tenancy):
- Mọi truy vấn dữ liệu đều phải lọc qua org_id để đảm bảo tính cô lập (Logic Isolation).
- Admin Level 1 có quyền truy cập toàn bộ organizations, trong khi Level 2 chỉ thấy dữ liệu thuộc công ty của mình.