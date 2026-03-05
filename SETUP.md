# ⚙️ Hướng dẫn Triển khai Hệ thống SENT SOC

## 📋 Yêu cầu hệ thống
* **Backend/Agent:** Golang 1.21+
* **Frontend:** NodeJS 18+ (React 18)
* **Database:** PostgreSQL 15+

## 📦 Bước 1: Thiết lập Cơ sở dữ liệu
1.  Khởi chạy PostgreSQL và tạo database tên: `sent_db`.
2.  Backend sử dụng GORM để tự động khởi tạo các bảng (AutoMigrate).

## 🖥️ Bước 2: Cài đặt Backend SOC
1.  Di chuyển vào thư mục: `cd SENT_backend`
2.  Tạo file `.env` với nội dung mẫu:
    ```env
    PORT=8000
    DATABASE_URL=host=localhost user=postgres password=YOUR_PASS dbname=sent_db port=5432 sslmode=disable
    JWT_SECRET=thanh_soc_security_key
    ```
3.  Cài đặt thư viện: `go mod tidy`
4.  Chạy server: `go run main.go`

## 🌐 Bước 3: Cài đặt Giao diện Frontend
1.  Di chuyển vào thư mục: `cd SENT_frontend`
2.  Cài đặt dependencies: `npm install`
3.  Khởi chạy: `npm start` (Giao diện sẽ chạy tại `http://localhost:3000`)

## 🛡️ Bước 4: Cài đặt Endpoint Agent
1.  Di chuyển vào thư mục: `cd SENT_agent`
2.  Cấu hình `config.json` hoặc nhập **Company Code** khi khởi chạy lần đầu.
3.  Build và chạy: `go run main.go`
    * *Lưu ý: Agent cần quyền Admin trên Windows để thu thập thông tin phần mềm và USB.*

