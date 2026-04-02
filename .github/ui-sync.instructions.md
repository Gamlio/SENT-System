---
name: ui-sync-specialist
description: "Sử dụng khi cần đồng bộ giao diện, style, hoặc cấu trúc component giữa nhiều trang dựa trên một mẫu (reference) đã có sẵn."
---

# Kỹ năng: Đồng bộ hóa Giao diện & Trải nghiệm (UI/UX Sync)

Khi nhận được yêu cầu đồng bộ hóa 3 trang còn lại giống với 3 trang đã sửa, hãy thực hiện theo quy trình sau:

## 1. Trích xuất Mẫu chuẩn (Pattern Extraction)
* **Phân tích Reference:** Đọc 3 file đã sửa (Source of Truth) để tìm ra các điểm chung mới:
    * **Layout:** Cấu trúc Header, Footer, Sidebar, Grid/Flexbox.
    * **Styling:** Các class CSS/Tailwind mới, bảng màu, spacing (padding/margin).
    * **Components:** Các component dùng chung đã được cập nhật.
    * **Logic:** Cách xử lý dữ liệu hoặc state mới trên UI.

## 2. Lập danh sách Đối chiếu (Gap Analysis)
* Liệt kê 3 trang chưa sửa (Targets).
* So sánh từng trang Target với mẫu chuẩn để xác định những gì còn thiếu hoặc lỗi thời.
* **Đảm bảo:** Giữ lại các logic xử lý dữ liệu đặc thù (unique content) của trang Target, chỉ thay đổi phần "vỏ" giao diện.

## 3. Thực thi Đồng bộ (Systematic Update)
* Cập nhật code cho các trang Target theo phong cách của mẫu chuẩn.
* **Cấu trúc phản hồi:**
    * [Tên file]: Tóm tắt những thay đổi về giao diện.
    * [Code block]: Đoạn code đã được đồng bộ hóa hoàn chỉnh.
* **Kiểm tra:** Đảm bảo tính responsive (mobile/desktop) vẫn hoạt động tốt trên các trang mới update.

> **Lưu ý:** Nếu phát hiện các đoạn CSS lặp lại quá nhiều, hãy đề xuất tách chúng ra thành một file CSS chung hoặc một Component dùng chung (Refactoring).