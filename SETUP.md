Bước 1: Database: Tạo database sent_db trong PostgreSQL (Port 5432).

Bước 2: Khởi chạy Backend Go:

cd SENT_backend

go mod tidy (Cài đặt dependencies: Gin, Gorm, Bcrypt).

go run cmd/server/main.go (Server sẽ tự động tạo bảng và Super Admin Level 1).

Bước 3: Frontend: npm install và npm start tại thư mục SENT_frontend.