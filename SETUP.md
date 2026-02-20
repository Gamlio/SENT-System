# ⚙️ Hướng dẫn Triển khai SENT SOC System

## Bước 1: Chuẩn bị Database
- Cài đặt và khởi chạy **PostgreSQL** (Mặc định Port `5432`).
- Tạo một database trống có tên: `sent_db`.

## Bước 2: Khởi chạy Backend (Trạm điều phối & AI Docs)
1. Mở Terminal, di chuyển vào thư mục Backend: `cd SENT_backend`
2. Tạo file `.env` ở thư mục gốc chứa các thông tin sau:
   ```env
   PORT=8000
   DATABASE_URL=host=localhost user=postgres password=YOUR_PASSWORD dbname=sent_db port=5432 sslmode=disable TimeZone=Asia/Ho_Chi_Minh
   JWT_SECRET=super_secret_key_cua_thanh
   ALLOWED_ORIGINS=http://localhost:3000
3. Chạy lệnh cài đặt thư viện: go mod tidy

4. Khởi chạy Server: go run cmd/server/main.go

Lưu ý: Quá trình này sẽ tự động AutoMigrate Database, tạo cây thư mục uploads/policies/ để lưu tài liệu AI và khởi tạo tài khoản Admin mặc định.

Bước 3: Đăng nhập vào Hệ thống
Backend đã tự động tạo một tài khoản Super Admin với thông tin sau:

Username: Admin

Password: Thanh@123

Khuyến nghị đổi mật khẩu sau lần đăng nhập đầu tiên.

Bước 4: Khởi chạy Giao diện Frontend
Mở một Terminal mới, di chuyển vào thư mục Frontend: cd SENT_frontend

Cài đặt các gói Node modules: npm install

Khởi chạy Dashboard: npm start (Mặc định sẽ chạy tại http://localhost:3000)