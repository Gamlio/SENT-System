---
name: fetch
description: "fetch": {
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-fetch"]
}
---

# Fetch

## Instructions

Sử dụng skill này để thực hiện các yêu cầu HTTP (GET, POST, PUT) tới Backend API chạy tại http://localhost:8000.

Kiểm tra các endpoint như /health, /api/v1/alerts, và các route xử lý dữ liệu từ Agent.

Xác thực rằng các phản hồi (response) từ Backend trả về đúng định dạng JSON cho Frontend React tại cổng 3000.

Khi sửa đổi code Go, hãy dùng Fetch để đảm bảo các thay đổi không làm hỏng (break) giao thức kết nối hiện tại.
