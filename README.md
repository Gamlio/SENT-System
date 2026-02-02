# SENT - Security Monitoring & Asset Management
Hệ thống giám sát an ninh tập trung dành cho doanh nghiệp SME.

## 🚀 Công nghệ sử dụng
- **Backend:** Python FastAPI, SQLAlchemy, PostgreSQL.
- **Frontend:** React, Tailwind CSS, Lucide Icons.
- **Agent:** Golang (Hiệu suất cao cho máy trạm).

## 🛠 Luồng hoạt động trung gian (Mediator Flow)
Mọi kết nối trong hệ thống đều đi qua **Backend FastAPI** làm trung gian điều hướng:
1. **Agent (Go):** Thu thập thông tin máy trạm -> Gửi HTTPS Request kèm Token về Backend.
2. **Backend (Python):** Xác thực Token -> Phân tích dữ liệu -> Lưu vào PostgreSQL -> Đẩy thông báo qua WebSocket.
3. **Frontend (React):** Nhận dữ liệu từ Backend -> Hiển thị Dashboard thời gian thực.