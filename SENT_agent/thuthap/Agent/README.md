# 🛡️ HƯỚNG DẪN SỬ DỤNG - SENTINEL OFFLINE AGENT (Stealth Mode)

**Phiên bản:** 4.0 (Zero Footprint)
**Mục đích:** Thu thập dữ liệu máy trạm ngầm (Hardware, Software, Network Ports, Security Logs), tự động đồng bộ lên Google Drive và xóa dấu vết cục bộ.

---

## 📂 1. Cấu Trúc Thư Mục
Để Agent hoạt động hoàn hảo, thư mục triển khai cần có đủ 2 file cốt lõi sau:

* `offline_agent.exe`: File thực thi chính của Agent (Đã được build ở chế độ tàng hình).
* `Stop_Agent.bat`: Công cụ dùng để ép dừng tiến trình và dọn dẹp hệ thống.

---

## 🚀 2. Hướng Dẫn Kích Hoạt (Khởi chạy)

Agent được thiết kế để chạy hoàn toàn dưới nền hệ thống (Background Process) và **không hiển thị bất kỳ giao diện hay cửa sổ CMD nào** để tránh gây chú ý.

1.  Mở thư mục chứa Agent.
2.  Click **chuột phải** vào file `offline_agent.exe`.
3.  Bắt buộc chọn **"Run as administrator"** (Chạy dưới quyền Quản trị viên).
    * *Lưu ý: Nếu chỉ click đúp bình thường, Agent sẽ tự động hủy lệnh do không đủ quyền đọc Registry và System Ports.*
4.  Màn hình sẽ **KHÔNG CÓ GÌ XẢY RA**. Điều này là hoàn toàn bình thường (Chế độ tàng hình đã kích hoạt).

---

## 🔄 3. Cơ Chế Hoạt Động Ngầm

Sau khi được kích hoạt thành công, Agent sẽ thực hiện vòng lặp tự động như sau:

* **Chu kỳ:** Cứ đúng **5 phút** một lần.
* **Thu thập:** Quét toàn bộ Hardware, Software hiện tại, Ports đang mở và 10 sự kiện đăng nhập thất bại gần nhất từ Windows Event Log.
* **Đóng gói:** Tạo file `sent_data_[TenMay]_[ThoiGian].json` tại thư mục hiện tại.
* **Đồng bộ:** Gửi file JSON lên Google Drive của Admin qua mã hóa HTTPS.
* **Xóa dấu vết (Zero Footprint):** Ngay khi Google Drive báo nhận thành công, Agent lập tức **xóa vĩnh viễn** file JSON vừa tạo ra trên máy tính để không làm đầy ổ cứng và không để lại dấu vết thu thập.

> **💡 Mẹo kiểm tra:** Sau khi chạy file `.exe` khoảng 10 giây, hãy mở Google Drive của Admin để kiểm tra xem file JSON đầu tiên đã xuất hiện chưa. 

---

## 🛑 4. Hướng Dẫn Gỡ Bỏ & Dừng Thu Thập

Vì Agent chạy ẩn sâu trong hệ thống, bạn không thể tắt nó bằng dấu X như phần mềm bình thường. Khi muốn kết thúc quá trình giám sát máy trạm này, hãy làm theo bước sau:

1.  Mở thư mục chứa Agent.
2.  Nháy đúp chuột vào file **`Stop_Agent.bat`**.
3.  Một cửa sổ màu đỏ sẽ hiện lên, tự động dò tìm và "tiêu diệt" tiến trình `offline_agent.exe` đang chạy ngầm.
4.  Hệ thống đồng thời quét và xóa các file `.json` rác (nếu có do rớt mạng ngắt quãng quá trình upload).
5.  Cửa sổ tự đóng sau 3 giây. Trạng thái máy trạm trở về bình thường.

---

## ⚠️ 5. Xử Lý Sự Cố (Troubleshooting)

* **Không thấy file trên Google Drive:** Kiểm tra lại kết nối mạng của máy trạm. Nếu máy trạm mất mạng, file JSON cục bộ sẽ tạm thời **không bị xóa** để chờ lần quét 10 phút tiếp theo gửi lại.
* **Lỗi Access Denied trong file JSON:** Do lúc kích hoạt bạn quên chọn *Run as administrator*. Hãy chạy file `Stop_Agent.bat` để tắt đi, sau đó khởi chạy lại `.exe` với quyền Admin.