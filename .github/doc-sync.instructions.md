---
name: doc-sync-specialist
description: "Sử dụng khi cần cập nhật tài liệu BACKEND.md và FRONTEND.md dựa trên mã nguồn thực tế hiện tại của dự án."
---

# Kỹ năng: Đồng bộ hóa Tài liệu Kỹ thuật (Reverse Engineering Docs)

Khi nhận yêu cầu cập nhật BACKEND.md hoặc FRONTEND.md, hãy thực hiện quy trình sau:

## 1. Quét Mã nguồn (Codebase Scanning)
* **Đối với BACKEND.md:** Quét các file Controller, Routes, Middleware, Service và Database Schema. Xác định các API Endpoint mới nhất, các tham số đầu vào/đầu ra và các logic xử lý chính.
* **Đối với FRONTEND.md:** Quét các Components, Hooks, State Management (Redux/Context), và các luồng gọi API. Xác định cấu trúc thư mục, các thư viện chính và luồng điều hướng (Routing).

## 2. Đối chiếu & Phát hiện thay đổi (Diff Analysis)
* Đọc nội dung cũ của `BACKEND.md` và `FRONTEND.md`.
* Liệt kê các phần đã cũ, không còn khớp với code (ví dụ: API cũ đã xóa, biến đã đổi tên, hoặc cấu trúc folder đã thay đổi).
* Phát hiện các tính năng mới có trong code nhưng chưa có trong tài liệu.

## 3. Cập nhật Nội dung (Content Rewriting)
* Viết lại tài liệu theo cấu trúc chuyên nghiệp, bao gồm:
    * **Kiến trúc hệ thống:** Sơ đồ luồng (nếu cần).
    * **Chi tiết kỹ thuật:** Danh sách API, Data Models, Components chính.
    * **Hướng dẫn:** Cách cài đặt, chạy và test.
* Giữ nguyên phong cách trình bày của file cũ nhưng cập nhật 100% số liệu và logic mới.

> **Yêu cầu:** Tài liệu phải chính xác tuyệt đối với code thực tế. Nếu có phần nào trong code không rõ ràng, hãy ghi chú lại để người dùng kiểm tra.
> **Lưu ý:** Nếu có sự khác biệt giữa code và tài liệu, ưu tiên mô tả chính xác code trong tài liệu, đồng thời đánh dấu phần nào đã thay đổi để người dùng dễ dàng nhận biết.
"Luôn thêm một mục  'Git Hash' vào đầu file tài liệu để biết bản docs này tương ứng với phiên bản code nào."