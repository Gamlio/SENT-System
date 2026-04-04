---
name: backend-sync-expert
description: "Sử dụng khi cần đồng bộ dữ liệu từ Client lên Backend trong hệ thống Hybrid DB (PostgreSQL + MongoDB), đảm bảo bảo mật và đúng kiến trúc."
---

# Kỹ năng: Đồng bộ hóa & Bảo mật Backend (Hybrid DB)

Khi nhận yêu cầu cập nhật Backend dựa trên file thu thập dữ liệu hoặc tính năng mới, hãy thực hiện theo các quy tắc nghiêm ngặt sau:

## 1. Phân loại Database (Database Classification)
Dựa trên kiến trúc hiện tại, hãy chọn đúng nơi lưu trữ:
* **PostgreSQL (GORM):** Dành cho các tính năng cấu hình, tổ chức, người dùng và phê duyệt (ví dụ: `Organization`, `User`, `asset`, `ApprovalTicket`, `UniversalPolicy`).
* **MongoDB (NoSQL):** Dành cho dữ liệu phát sinh liên tục, log giám sát, và dữ liệu không cấu hình (ví dụ: `SoftwareItem`, `USBLog`, `assetIOActivity`, `AIChatLog`).

## 2. Strong Field Validation (Bắt buộc)
Để tránh các lỗi bảo mật như Injection, Malformed Request hoặc DoS qua asset Flooding, mọi struct nhận dữ liệu (Payload) phải:
* Luôn thêm tag `binding:"required"` cho các trường bắt buộc để chặn dữ liệu rỗng ở tầng API layer.
* Sử dụng các ràng buộc bổ sung như `min`, `max`, `email`, `alphanum` để kiểm soát chặt chẽ nội dung đầu vào.
* **Ví dụ:** `Username string `json:"username" binding:"required,alphanum,min=3"``.

## 3. Thực thi Logic & Cấu trúc (Implementation)
* **Request Structs:** Tạo các `Payload` struct riêng biệt trong file thích hợp để nhận dữ liệu từ Client.
* **Database Mapping:** * Nếu dùng Postgres: Sử dụng `database.DB` để `Create` hoặc `Updates`.
    * Nếu dùng Mongo: Sử dụng các Collection tương ứng như `SoftwareCollection` hoặc `USBCollection` từ `database.go`.
* **Error Handling:** Trả về lỗi 400 (Bad Request) ngay lập tức nếu Validation thất bại, kèm theo thông báo chi tiết về trường dữ liệu bị lỗi.

> **Ghi chú:** Tuyệt đối không để xảy ra tình trạng parse JSON mà thiếu kiểm duyệt (validation). Luôn ưu tiên tính toàn vẹn của dữ liệu và hiệu năng hệ thống.