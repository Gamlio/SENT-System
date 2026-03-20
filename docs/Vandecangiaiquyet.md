🚨 1. LỖ HỔNG CHÍNH MẠNG: Phá hủy Bằng chứng Pháp y (Forensic Evidence)
Vấn đề:
Ở bước giải quyết lỗi khóa ngoại (SQLSTATE 23503) vừa rồi, chúng ta đã viết hàm deleteAgentSafely dùng lệnh DELETE FROM để xóa sạch sành sanh từ Hoạt động (Activity) -> Sự cố (Incident) -> Cảnh báo (Alert) -> Máy trạm (Agent).
Trong ngành An toàn thông tin, ĐÂY LÀ ĐIỀU TỐI KỴ! Nếu một máy trạm bị hacker xâm nhập (gây ra Incident), sau đó hacker hoặc Admin nội gián xóa máy trạm đó đi, toàn bộ dấu vết tấn công sẽ bốc hơi khỏi Database. Hệ thống SIEM/SOC của bạn bị "mù" hoàn toàn.

Giải pháp (Soft Delete - Xóa Mềm):

Tuyệt đối KHÔNG BAO GIỜ xóa vật lý thiết bị và sự cố.

Thay vì gọi .Delete(), khi Admin duyệt Đơn Xóa, ta chỉ đổi trạng thái máy trạm thành UNINSTALLED hoặc RETIRED.

Dữ liệu vẫn nằm trong DB để phục vụ điều tra (Audit), nhưng không hiển thị trên giao diện Dashboard hoặc Bảng điều khiển nữa.

Cách sửa trong agent_strategy.go:

Go
func (s *AgentDeleteStrategy) OnApprove(tx *gorm.DB, ticket *models.ApprovalTicket) error {
	// XÓA MỀM: Chỉ đổi trạng thái, ngắt kết nối, giữ lại toàn bộ Log và Incident
	return tx.Model(&models.Agent{}).Where("hw_id = ?", ticket.TargetName).
		Updates(map[string]interface{}{
			"status": "UNINSTALLED", // Hoặc "RETIRED"
			"secret_key": "",        // Thu hồi khóa luôn, máy này vĩnh viễn không gửi được data nữa
		}).Error
}
🚨 2. LỖ HỔNG XÁC THỰC: Agent gửi rác mà không bị kiểm tra (Spoofing)
Vấn đề:
Lúc đăng ký (Enrollment), ta đã cấp cho Agent một cái SecretKey rất xịn để cất giấu. NHƯNG ở API nhận dữ liệu (PushDataHandler), backend chỉ kiểm tra xem HWID và CompanyCode có tồn tại không rồi nhận data luôn:

Go
database.DB.Where("hw_id = ? AND ...", req.HWID, req.CompanyCode).First(&agent)
Nếu nhân viên biết được HWID của sếp, họ có thể dùng Postman bắn data giả (Ví dụ: gửi log báo máy sếp đang cài phần mềm cấm) lên Server. Backend hoàn toàn tin tưởng.

Giải pháp (HMAC Signature):

Agent phải dùng SecretKey để băm (Hash) nội dung Payload thành một chữ ký (Signature) gửi kèm trong Header (X-Sent-Signature).

Backend nhận được, lấy SecretKey của HWID đó trong DB, băm thử nội dung. Nếu 2 chữ ký khớp nhau thì mới nhận Data. (Giống hệt cách Webhook của GitHub hay Stripe hoạt động).

🚨 3. LỖ HỔNG TỪ CHỐI DỊCH VỤ (DoS): Re-Enrollment bừa bãi
Vấn đề:
Trong hàm EnrollAgent (đăng ký), nếu một HWID đã tồn tại, chúng ta tự động đẩy trạng thái của máy đó về lại PENDING và cấp SecretKey mới.
Kẻ gian chỉ cần lấy được Enrollment Token (còn hạn 15 phút) và liên tục gửi API đăng ký lại với HWID của các máy chủ quan trọng. Các máy chủ này sẽ lập tức bị đá văng khỏi mạng (chuyển về PENDING) và ngừng đẩy log cho đến khi Admin vào duyệt lại.

Giải pháp (Lock Active Agents):

Máy nào đã ACTIVE thì CẤM đăng ký lại.

Nếu máy bị cài lại Win thật, Admin phải chủ động vào Web bấm nút "Cho phép đăng ký lại (Revoke & Re-enroll)", đưa trạng thái máy về WAITING_REINSTALL, lúc đó API Enroll mới chấp nhận.

🚨 4. KIẾN TRÚC DB: Phình to vì lưu History vào chung bảng Chính
Vấn đề:
Mỗi lần quét (30s), nếu có cổng mạng mới mở (OpenPorts) hoặc phần mềm cài mới (SoftwareItem), backend đang dùng tx.Where(...).Delete() để xóa sạch cái cũ và tx.Create() cái mới.
Điều này gây lãng phí tài nguyên ghi (Write) của Database, đồng thời bạn không lưu lại được Lịch sử (Time-series). (Ví dụ: 2h sáng máy sếp có mở cổng 3389 không? Hiện tại không biết được vì data đã bị ghi đè mất).

Giải pháp (Đưa dữ liệu động vào ElasticSearch hoặc DB Time-series):

Về lâu dài (sau giai đoạn Test), các dữ liệu tĩnh như Cấu hình (RAM/CPU/Tên máy) lưu ở PostgreSQL.

Các dữ liệu nhảy múa liên tục như Port, Process, USB Logs nên được đẩy thẳng vào một Database chuyên chứa Time-series (như ElasticSearch, ClickHouse, hoặc VictoriaMetrics) để phục vụ việc search tốc độ cao và vẽ biểu đồ lịch sử.