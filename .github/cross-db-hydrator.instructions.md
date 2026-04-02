---
name: cross-db-hydrator
description: "Sử dụng khi cần viết hàm lấy dữ liệu từ MongoDB để lấp đầy các trường ảo (gorm:\"-\") trong các Model PostgreSQL."
---

# Kỹ năng: Hydration Dữ liệu Liên DB

Khi người dùng yêu cầu hiển thị thông tin chi tiết của một Agent hoặc Incident:

## 1. Xác định Nguồn (Sources)
* Lấy thông tin cơ bản từ PostgreSQL (bảng Agents, Incidents).
* Xác định các Collection cần truy vấn ở MongoDB dựa trên `AgentHWID` (ví dụ: `SoftwareCollection`, `USBCollection`).

## 2. Viết Logic Truy vấn (Query Logic)
* Sử dụng `database.MongoClient` để thực hiện các câu lệnh `Find` hoặc `Aggregate`.
* Chuyển đổi (Map) kết quả từ Mongo sang các Slice trong struct của Postgres.

## 3. Tối ưu Hiệu suất
* Đề xuất sử dụng `goroutines` (wait groups) để truy vấn song song nhiều collection từ Mongo cùng lúc nhằm giảm latency.