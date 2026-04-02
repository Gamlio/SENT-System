---
name: system-analyzer-pro
description: "Phân tích hệ thống và vẽ biểu đồ chuẩn hóa (Usecase, Sequence, Activity) theo phong cách chuyên nghiệp. Lưu kết quả vào docs/architecture/."
---

# Kỹ năng: Phân tích & Thiết kế Hệ thống Chuyên sâu

## 1. Tiêu chuẩn Vẽ biểu đồ (Mermaid Standard)
Để đảm bảo biểu đồ "đẹp" và "chuẩn", hãy tuân thủ:
* **Usecase:** Phải có ranh giới hệ thống (`subgraph`), phân biệt rõ Actor chính (Admin/User) và Actor hệ thống (Agent/AI). Sử dụng đúng `-->` cho tương tác và `-.->` cho `<<include>>/<<extend>>`.
* **Sequence:** Phải có `autonumber`. Sử dụng `box` để nhóm các thành phần (Frontend, Backend, Database). Luôn có mũi tên phản hồi (dotted line `-->`).
* **Activity:** Sử dụng `start/stop`, các khối điều kiện `if/else` rõ ràng và phân làn (swimlanes) nếu quy trình đi qua nhiều bộ phận.
* **Giao diện:** Sử dụng cấu trúc `direction LR` (Trái sang Phải) cho Usecase để tránh biểu đồ quá dài.

## 2. Quy trình Kiểm kê (Inventory)
* Quét `models.go` (Postgres) và `mongo_models.go` (Mongo) để liệt kê usecase.
* **Phân loại:** [Tên Usecase] -> [Loại DB] -> [Vùng hiển thị (Agent/Incident/Admin)].
* **Cảnh báo:** Chỉ ra các usecase thiếu `binding:"required"` (lỗi bảo mật).

## 3. Quản lý File (File Management)
* Mọi biểu đồ phải được xuất ra file riêng biệt trong thư mục `docs/architecture/diagrams/`.
* Tên file định dạng: `[loại_biểu_đồ]_[tên_chức_năng].mmd`.