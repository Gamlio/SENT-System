# SENT-System Project Guide (ICTU Final Project)

## Tech Stack
- Backend: Go (Gin Framework, gRPC)
- Frontend: React (Vite, Tailwind CSS)
- Database: PostgreSQL (Relational), MongoDB (Logs)
- Infrastructure: Docker Compose

## Documentation Rules
- Khi cập nhật tài liệu, phải tuân thủ format báo cáo kỹ thuật của ICTU.
- Mọi thay đổi trong cấu trúc file phải được phản ánh vào mục "Project Structure" trong README.md.
- Sử dụng Mermaid.js để vẽ lại luồng dữ liệu nếu có module mới được thêm vào.

## Deployment Commands
- Build all: `docker-compose up --build -d`
- Test Backend: `go test ./...`