Silent Enrollment (Kích hoạt ngầm):

Admin Level 3 tạo mã Enroll_Token trên Web Dashboard.

Agent tự động đọc mã này từ file cấu hình, thu thập mã phần cứng (HWID) và tự đăng ký về Backend.

Giám sát Tài sản (Asset Monitoring):

Agent quét định kỳ trạng thái USB và phần mềm cài đặt.

Dữ liệu được đẩy về API /api/v1/assets/push dưới dạng JSON tinh gọn.

Phát hiện Bất thường:

Backend so sánh dữ liệu Agent đẩy về với Rules đã được Admin thiết lập.

Nếu phát hiện thiết bị lạ hoặc hành vi bất thường, Backend ghi log và đẩy cảnh báo tức thì về Dashboard.