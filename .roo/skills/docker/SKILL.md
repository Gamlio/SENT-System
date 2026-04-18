---
name: docker
description: "docker": {
  "command": "npx",
  "args": ["-y", "@pro-mcp/docker"]
}
---

# Docker

## Instructions

Sử dụng skill này để giám sát trạng thái của 4 container chính: sent_db, sent_mongo, sent_backend, và sent_frontend.

Nếu backend tại cổng 8000 không phản hồi, hãy dùng lệnh logs của Docker để kiểm tra lỗi runtime trong container sent_backend.

Kiểm tra mức chiếm dụng tài nguyên của các container để đảm bảo hệ thống SOC không làm treo máy (đặc biệt quan trọng khi chạy AI local trên GPU 4GB).

Hỗ trợ restart các dịch vụ khi có sự thay đổi trong file cấu hình .env hoặc mã nguồn Go.