---
name: mongodb
description: "mongodb": {
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-mongodb", "--uri", "mongodb://localhost:27017"]
}
---

# Mongodb

## Instructions
Sử dụng skill này để đọc và phân tích các log thô (raw logs) không cấu trúc được gửi về từ các Endpoint Agent.

Thực hiện các truy vấn đếm (count) hoặc tìm kiếm theo thời gian thực để xác định tần suất xuất hiện của các sự kiện bảo mật.

Địa chỉ kết nối mặc định là mongodb://localhost:27017.

Khi cần kiểm tra hiệu năng của Agent, hãy truy vấn các bản ghi có chứa thông tin về CPU và RAM trong collection log.