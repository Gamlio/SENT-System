Luồng Xử lý Backend (BACKEND_FLOW.md)
Cơ chế Mediator (Điều phối):

Backend Go đóng vai trò là "Trạm kiểm soát" trung tâm.

Mọi dữ liệu từ máy trạm (Go Agent) đẩy về đều phải qua lớp Middleware xác thực trước khi được đưa vào xử lý tại service.

Xác thực Đa tầng (Security Flow):


Bảo mật 1: Xác thực Username/Password bằng bcrypt (Native Go).


Bảo mật 2: Yêu cầu mã TOTP (2FA) cho các tài khoản quản trị (Level 1, 2, 3).


Cấp quyền: Trả về JWT Token chứa thông tin level và org_id.

Cô lập Dữ liệu (Multi-tenancy):

Sử dụng cơ chế lọc cứng tại lớp repository.

Mọi câu lệnh SQL đều tự động đính kèm điều kiện WHERE org_id = ? dựa trên dữ liệu trích xuất từ Token.