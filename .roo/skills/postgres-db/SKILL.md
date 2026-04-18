---
name: postgres-db
description: "postgres-db": {
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-postgres", "postgresql://postgres:Thanh@123@localhost:5432/sent_db"]
}
modeSlugs:
  - ask
---

# Postgres Db

## Instructions
Sử dụng skill này để truy vấn dữ liệu quan hệ trong cơ sở dữ liệu sent_db.

Tập trung vào việc kiểm tra bảng alerts để theo dõi các cảnh báo bảo mật và bảng users để quản lý quyền truy cập.

Luôn kiểm tra cấu trúc bảng (schema) trước khi thực hiện các câu lệnh SELECT hoặc INSERT phức tạp.

Khi phát hiện có cảnh báo mới, hãy đối chiếu với dữ liệu từ MongoDB để xác định nguồn gốc của log.