---
name: incident-extension-flow
description: "Dùng khi thêm loại dữ liệu thu thập mới từ asset, đồng bộ qua Backend và chỉ hiển thị trong luồng xử lý Sự cố (Incidents)."
---

# Kỹ năng: Điều phối Mở rộng Usecase Sự cố

Khi nhận yêu cầu thêm một loại dữ liệu giám sát mới (ví dụ: Log tiến trình lạ, File dung lượng lớn):

## 1. Tầng Thu thập (asset & Payload)
* Phân tích cấu trúc dữ liệu mới từ phần mềm asset.
* Thiết kế `Payload` struct cho Backend với tag `binding:"required"` để chống DoS và Injection.

## 2. Tầng Lưu trữ (Hybrid Backend)
* **MongoDB:** Tạo struct mới trong `mongo_models.go` để lưu trữ dữ liệu lớn/log phát sinh.
* **PostgreSQL:** Không thay đổi bảng asset chung; chỉ tạo liên kết qua `IncidentID` nếu dữ liệu này dùng làm bằng chứng sự cố.

## 3. Tầng Hiển thị (Targeted UI)
* **Quy tắc vàng:** Không hiển thị dữ liệu này ở trang thông tin asset chung để tránh làm rối giao diện.
* **Thực thi:** Tạo một Component con (Evidence Widget) và chỉ nhúng vào màn hình chi tiết của `Incident`.

## 4. Tối ưu & Bảo mật
* Sử dụng Goroutines để xử lý ghi dữ liệu vào Mongo không gây nghẽn API.
* Kiểm tra quyền hạn `PermIncidentView` trước khi trả về dữ liệu này.