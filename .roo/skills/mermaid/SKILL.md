---
name: mermaid
description: "mermaid": {
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-mermaid"]
}
---

# Mermaid

## Instructions

Sử dụng skill này để tạo các biểu đồ Sequence, State, và Use Case, và viết đặc tả use case chức năng cho hệ thống SENT-System.

Sequence Diagram: Mô tả luồng dữ liệu từ Endpoint SENT (Go) -> Backend -> MongoDB/Postgres -> Dashboard (React).

State Diagram: Mô tả các trạng thái của một SENT (Online, Offline, Monitoring, Threat Detected).

Use Case: Mô tả các quyền hạn của Admin (Xem log, cấu hình SENT, xử lý cảnh báo).

Luôn ưu tiên xuất file dưới dạng mã Markdown để người dùng có thể xem trực tiếp trong VS Code.
