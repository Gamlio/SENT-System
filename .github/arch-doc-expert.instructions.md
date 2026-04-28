---
name: arch-doc-expert
description: "Sử dụng khi cần viết tài liệu kỹ thuật (Backend/Frontend flow), phân tích cấu trúc folder hoặc yêu cầu thay đổi cấu trúc dự án hiện tại."
---

# Kỹ năng: Tài liệu hóa & Tối ưu Cấu trúc

Khi nhận yêu cầu về tài liệu hoặc cấu trúc, hãy thực hiện:

## 1. Phân tích Luồng (Flow Mapping)
* **Backend Flow:** Mô tả chi tiết từ Request -> Middleware -> Controller -> Service -> Repository -> Database.
* **Frontend Flow:** Mô tả từ Component -> State Management (Redux/Context) -> API Call -> Render.
* **Cấu trúc:** Xuất ra sơ đồ dạng Markdown (Mermaid) hoặc danh sách phân cấp file rõ ràng.
Cross-DB Mapping: Hiểu rằng các field có tag gorm:"-" trong models.go (như Software, Alerts, Inventory) là dữ liệu ảo, cần được truy vấn từ MongoDB dựa trên AssetHWID .

Phân vùng dữ liệu: Dữ liệu quản trị (User, Org, Policy) nằm ở Postgres; dữ liệu giám sát/logs (USB, Software, Port, IO) nằm ở MongoDB.
## 2. Cập nhật & Sửa đổi Cấu trúc (Refactoring)
* Nếu người dùng muốn sửa cấu trúc, hãy phân tích `@workspace` để đề xuất cách di chuyển file, tách logic (ví dụ: tách Business Logic ra khỏi Controller) sao cho chuẩn Clean Architecture.
* Đảm bảo sau khi đổi cấu trúc, các đường dẫn `import` và `export` được cập nhật chính xác.
"Nếu ghi vào Postgres thành công nhưng ghi vào MongoDB thất bại, phải có cơ chế Retry hoặc đánh dấu lỗi (Inconsistency Flag) để xử lý sau."
## 3. Định dạng Tài liệu
* Luôn viết tài liệu dưới định dạng Markdown chuyên nghiệp.
* Bao gồm các phần: Tổng quan, Sơ đồ luồng, Danh sách API/Component liên quan, và Lưu ý bảo mật/Hiệu suất.