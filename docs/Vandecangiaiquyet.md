4. KIẾN TRÚC DB: Phình to vì lưu History vào chung bảng Chính



Tình trạng: Chấp nhận rủi ro tạm thời (Accepted Risk for MVP).

Chi tiết: Ở các hàm như AnalyzeSoftwareBehavior, để cập nhật danh sách phần mềm, Backend đang chạy lệnh Delete xóa sạch dữ liệu cũ rồi Create dữ liệu mới.

Đánh giá: Ở quy mô 50 - 100 máy trạm đang Test, cách này hoạt động hoàn hảo và dễ code. Tuy nhiên, nó vi phạm nguyên tắc "Giữ lại lịch sử theo thời gian" (Time-series).

Khi nào cần sửa: Việc đẩy data sang ElasticSearch hay ClickHouse đòi hỏi setup cơ sở hạ tầng phức tạp. Bạn khoan hãy làm bước này. Hãy đợi đến khi phần mềm của bạn có khách hàng đầu tiên, hoặc số lượng máy quản lý vượt quá 500 máy, lúc đó hẵng cấu trúc lại DB.