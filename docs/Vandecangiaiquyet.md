4. KIẾN TRÚC DB: Phình to vì lưu History vào chung bảng Chính



Tình trạng: Chấp nhận rủi ro tạm thời (Accepted Risk for MVP).

Chi tiết: Ở các hàm như AnalyzeSoftwareBehavior, để cập nhật danh sách phần mềm, Backend đang chạy lệnh Delete xóa sạch dữ liệu cũ rồi Create dữ liệu mới.

Đánh giá: Ở quy mô 50 - 100 máy trạm đang Test, cách này hoạt động hoàn hảo và dễ code. Tuy nhiên, nó vi phạm nguyên tắc "Giữ lại lịch sử theo thời gian" (Time-series).

Khi nào cần sửa: Việc đẩy data sang ElasticSearch hay ClickHouse đòi hỏi setup cơ sở hạ tầng phức tạp. Bạn khoan hãy làm bước này. Hãy đợi đến khi phần mềm của bạn có khách hàng đầu tiên, hoặc số lượng máy quản lý vượt quá 500 máy, lúc đó hẵng cấu trúc lại DB.

Lỗ hổng "Tự tung tự tác" (Self-Approval): Hiện tại, hệ thống Maker-Checker chưa có hàm chặn việc người tạo đơn cũng chính là người duyệt đơn. Nếu một người có cả quyền user_manage và approval_manage, họ có thể tạo đơn đổi quyền rồi tự sang tab Phê duyệt để ấn "Duyệt". Chúng ta phải thêm một dòng logic: if ticket.RequestedBy == currentUsername { return error("Không thể tự duyệt đơn của chính mình") }.

Rủi ro từ Snapshot: Dữ liệu SnapshotData đang được lưu dưới dạng Plain JSON text. Dù rất khó, nhưng nếu một ai đó can thiệp trực tiếp vào Database để sửa chuỗi JSON này trước khi nó được duyệt, hệ thống sẽ thực thi dữ liệu sai. (Giải pháp tương lai: Mã hóa hoặc tạo mã Hash cho chuỗi Snapshot).

Điểm cần tối ưu (Nút thắt cổ chai):

Rò rỉ Goroutine (Goroutine Leak): Hàm go func() gửi email hiện tại không có cơ chế Timeout hoặc Retry. Nếu server SMTP (Gmail) bị nghẽn, các Goroutine này sẽ treo vĩnh viễn trên RAM của máy chủ. (Cần dùng thư viện Queue như RabbitMQ hoặc Redis nếu scale lớn, hoặc thêm Context timeout).

Thiếu Index Database: Bảng ApprovalTicket đang được query liên tục với điều kiện Where("status = ?", "PENDING"). Khi công ty có hàng vạn đơn đã xử lý, câu query này sẽ rất chậm nếu cột status và module_type không được đánh Index trong cấu trúc GORM.

Frontend Parse JSON: Hàm renderContent đang dùng JSON.parse(ticket.snapshot_data) bên trong một vòng lặp map để render bảng. Nếu danh sách đơn phê duyệt lên tới 500 dòng, trình duyệt sẽ hơi giật. (Nên phân trang bảng Phê duyệt ở Backend).