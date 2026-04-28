---
name: system-auditor
description: "Sử dụng để đọc hiểu toàn bộ nội dung trong thư mục docs/, đối chiếu tài liệu với code thực tế và kiểm tra tính nhất quán của hệ thống."
---

# Kỹ năng: Kiểm tra & Đối chiếu Hệ thống (Audit)

Khi được yêu cầu kiểm tra hoặc đọc hiểu hệ thống:

## 1. Tiếp nhận Nguồn tin (Source of Truth)
* Ưu tiên đọc tất cả các file trong thư mục `docs/` để hiểu quy định, tiêu chuẩn và thiết kế ban đầu của dự án.
* Sử dụng thông tin trong `docs/` làm tiêu chuẩn để đánh giá code hiện tại.

## 2. Kiểm tra tính nhất quán (Consistency Check)
* So sánh logic code thực tế với những gì đã viết trong tài liệu.
* Phát hiện các điểm "lệch pha": Ví dụ tài liệu bảo có API A nhưng code thực tế không có, hoặc logic xử lý khác nhau.
* Kiểm tra toàn bộ hệ thống để đảm bảo các module kết nối với nhau đúng như tài liệu mô tả.

## 3. Báo cáo & Cảnh báo
* Liệt kê các tài liệu đã cũ (Outdated) cần cập nhật.
* Đưa ra danh sách các file code đang vi phạm quy chuẩn đã đề ra trong `docs/`.