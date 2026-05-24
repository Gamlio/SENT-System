
## TIẾN TRÌNH ĐỘC HẠI - MALICIOUS PROCESS
Từ khóa: tiến trình lạ, malicious process, tệp thực thi cấm, pid, path, sha256, blacklist.
- Phát hiện: Agent bắt sự kiện khởi chạy tệp tin ngoài whitelist, trích xuất pid, hash sha256, tạo hồ sơ sự cố.
- Cô lập: Thực thi quy tắc tường lửa tại chỗ cách ly máy trạm, cô lập host, chặn lateral movement.
- Xử lý từ xa: Kill process theo PID, hủy tiến trình độc hại qua API hệ điều hành.
- Xử lý thực địa: Xóa tệp thực thi trong Temp, AppData Local, AppData Roaming. Khóa xóa Registry Key khởi động ngầm Persistence tại HKLM\Software\Microsoft\Windows\CurrentVersion\Run và RunOnce.
- Phục hồi: Khôi phục mạng, quét mã độc toàn diện, cập nhật chữ ký số antivirus.

## CỔNG MẠNG TRÁI PHÉP - UNAUTHORIZED PORT
Từ khóa: cổng mạng, mở cổng, unauthorized port, network snapshot, listen, ip blacklist, reverse shell.
- Phát hiện: Phát hiện tiến trình mở cổng kết nối LISTEN trái phép hoặc kết nối IP nguy hại. Nhóm 18 cổng nhạy cảm: TCP 21 FTP, TCP 22 SSH, TCP 23 Telnet, TCP 80 443 http https lậu, TCP 445 SMB, TCP 1433 MSSQL, TCP 3306 MySQL, TCP 3389 RDP, TCP 5900 VNC, Reverse Shell Backdoor 4444, 8080, 8888.
- Cô lập: Gọi tường lửa Windows Firewall hoặc iptables đóng chặn cổng dịch vụ vi phạm và IP đích, chặn kênh C2.
- Xử lý từ xa: Truy vết đảo ngược Reverse Mapping tìm PID đang chiếm dụng cổng và thực hiện kill process từ xa.
- Xử lý thực địa: Rà soát phần cứng vật lý máy trạm, kiểm tra thiết bị cắm ngoài bypass mạng như USB 4G, card Wifi lậu.
- Phục hồi: Gỡ bỏ lệnh chặn, định cấu hình lại quy tắc lọc mạng cục bộ firewall rules.

## THIẾT BỊ NGOẠI VI TRÁI PHÉP - USB PERIPHERAL
Từ khóa: usb, cắm usb, thiết bị ngoại vi, unauthorized peripheral, lưu trữ ngoại vi, vid, pid, serial number, badusb.
- Phát hiện: Người dùng cắm thiết bị lưu trữ usb lạ vào máy trạm, đối chiếu danh sách whitelist vid pid serial thất bại.
- Cô lập: Active Response tự động vô hiệu hóa driver logic cổng lưu trữ USB (Disable USB Storage Device Driver), chặn hành vi đọc ghi hoặc nạp payload từ BadUSB.
- Xử lý từ xa: Khóa cứng không cho phép driver ngoại vi hoạt động lại thông qua cơ chế Policy tập trung.
- Xử lý thực địa: Tịch thu thu hồi USB vật lý. Truy vấn lịch sử cắm rút qua Windows Registry tại HKLM\SYSTEM\CurrentControlSet\Enum\USBSTOR và HKLM\SYSTEM\CurrentControlSet\Enum\USB để trích xuất thông tin.
- Phục hồi: Kiểm tra tính toàn vẹn hệ thống, quét virus tại chỗ, chỉ mở khóa driver khi có phê duyệt.

## PHẦN MỀM LẬU KHÔNG ĐƯỢC PHÉP - UNAUTHORIZED SOFTWARE
Từ khóa: phần mềm lậu, cài phần mềm, unauthorized software, phần mềm cấm, crack, keygen, portable.
- Phát hiện: Quét hệ thống tệp tin phát hiện ứng dụng trong Software Blacklist hoặc tệp chạy ngay Portable lậu.
- Cô lập: Gửi chỉ thị Suspend Process đóng băng tiến trình giao diện, bật popup cảnh báo.
- Xử lý từ xa: Khóa quyền thực thi đối với tệp tin nhị phân (binary files) của phần mềm lậu.
- Xử lý thực địa: Gỡ bỏ cài đặt Uninstall qua trình quản lý gói hệ điều hành, xóa file crack keygen trong thư mục tạm, dọn sạch Registry rác tại HKCU\Software và HKLM\Software.
- Phục hồi: Đồng bộ danh mục phần mềm tiêu chuẩn, giải phóng trạng thái khóa thực thi.

## VÔ HIỆU HÓA HỆ THỐNG BẢO VỆ - FIREWALL DISABLEMENT
Từ khóa: tắt tường lửa, tắt firewall, firewall evasion, disable firewall, ngưng dịch vụ bảo vệ, service monitor.
- Phát hiện: Dịch vụ tường lửa cục bộ bị tắt hoặc các dịch vụ an ninh nền tảng bị ngưng hoạt động.
- Cô lập: Cưỡng bức tái kích hoạt tường lửa qua lệnh netsh advfirewall set allprofiles state on (Windows) hoặc ufw enable (Linux).
- Xử lý từ xa: Kiểm tra quyền cấu hình dịch vụ, ép thiết lập trạng thái khởi động service về Tự động (Automatic).
- Xử lý thực địa: Rà soát Event Viewer. Kiểm tra sâu các nhánh registry cấu hình trạng thái tường lửa tại HKLM\SYSTEM\CurrentControlSet\Services\SharedAccess\Parameters\Firewalls\Policy để tìm mã độc can thiệp bypass.
- Phục hồi: Reset firewall rules về mặc định an toàn, chạy Patch Management cập nhật bản vá lỗi hệ điều hành.