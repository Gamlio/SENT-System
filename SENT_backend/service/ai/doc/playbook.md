
## TIẾN TRÌNH ĐỘC HẠI - MALICIOUS PROCESS
Từ khóa: tiến trình lạ, malicious process, tệp thực thi cấm, pid, path, sha256, blacklist.
- Phát hiện: Agent bắt sự kiện khởi chạy tệp tin ngoài whitelist, trích xuất pid, hash sha256, tạo hồ sơ sự cố.
- Xử lý thực địa: Xóa tệp thực thi trong Temp, AppData Local, AppData Roaming. Khóa xóa Registry Key khởi động ngầm Persistence tại HKLM\Software\Microsoft\Windows\CurrentVersion\Run và RunOnce.
- Phục hồi: Khôi phục mạng, quét mã độc toàn diện, cập nhật chữ ký số antivirus.

## CỔNG MẠNG TRÁI PHÉP - UNAUTHORIZED PORT
Từ khóa: cổng mạng, mở cổng, unauthorized port, network snapshot, listen, ip blacklist, reverse shell, tiến trình lạ.
- Phát hiện: Phát hiện tiến trình mở cổng kết nối LISTEN trái phép hoặc kết nối IP nguy hại. Nhóm 18 cổng nhạy cảm: TCP 21 FTP, TCP 22 SSH, TCP 23 Telnet, TCP 80 443 http https lậu, TCP 445 SMB, TCP 1433 MSSQL, TCP 3306 MySQL, TCP 3389 RDP, TCP 5900 VNC, Reverse Shell Backdoor 4444, 8080, 8888.
- Xử lý thực địa: Xác định PID (Process ID) đang chiếm dụng cổng, thực hiện kill tiến trình lạ ngay lập tức. Rà soát phần cứng vật lý máy trạm, kiểm tra thiết bị cắm ngoài bypass mạng như USB 4G, card Wifi lậu. Truy tìm vị trí file thực thi của tiến trình để cách ly hoặc xóa bỏ.
- Phục hồi: Cấu hình lại quy tắc lọc mạng cục bộ firewall rules, chặn IP blacklist ngoại vi.

## THIẾT BỊ NGOẠI VI TRÁI PHÉP - USB PERIPHERAL
Từ khóa: usb, cắm usb, thiết bị ngoại vi, unauthorized peripheral, lưu trữ ngoại vi, vid, pid, serial number, badusb.
- Phát hiện: Người dùng cắm thiết bị lưu trữ usb lạ vào máy trạm, đối chiếu danh sách whitelist vid pid serial thất bại.
- Xử lý thực địa: Tịch thu thu hồi USB vật lý. Truy vấn lịch sử cắm rút qua Windows Registry tại HKLM\SYSTEM\CurrentControlSet\Enum\USBSTOR và HKLM\SYSTEM\CurrentControlSet\Enum\USB để trích xuất thông tin.
- Phục hồi: Kiểm tra tính toàn vẹn hệ thống, quét virus tại chỗ.

## PHẦN MỀM LẬU KHÔNG ĐƯỢC PHÉP - UNAUTHORIZED SOFTWARE
Từ khóa: phần mềm lậu, cài phần mềm, unauthorized software, phần mềm cấm, crack, keygen, portable.
- Phát hiện: Quét hệ thống tệp tin phát hiện ứng dụng trong Software Blacklist hoặc tệp chạy ngay Portable lậu.
- Xử lý thực địa: Gỡ bỏ cài đặt Uninstall qua trình quản lý gói hệ điều hành, xóa file crack keygen trong thư mục tạm, dọn sạch Registry rác tại HKCU\Software và HKLM\Software.
- Phục hồi: Đồng bộ danh mục phần mềm tiêu chuẩn.

## VÔ HIỆU HÓA HỆ THỐNG BẢO VỆ - FIREWALL DISABLEMENT
Từ khóa: tắt tường lửa, tắt firewall, firewall evasion, disable firewall, ngưng dịch vụ bảo vệ, service monitor.
- Phát hiện: Dịch vụ tường lửa cục bộ bị tắt hoặc các dịch vụ an ninh nền tảng bị ngưng hoạt động.
- Cô lập: Cưỡng bức tái kích hoạt tường lửa qua lệnh netsh advfirewall set allprofiles state on (Windows) hoặc ufw enable (Linux).
HKLM\SYSTEM\CurrentControlSet\Services\SharedAccess\Parameters\Firewalls\Policy để tìm mã độc can thiệp bypass.
- Phục hồi: Reset firewall rules về mặc định an toàn, chạy Patch Management cập nhật bản vá lỗi hệ điều hành.

## CẬP NHẬT WINDOWS BỊ CHẶN HOẶC LỖI - WINDOWS UPDATE BLOCK/FAILURE
Từ khóa: cập nhật hệ điều hành, windows update, bản vá lỗi, patch management, tắt update, wuauserv, kb patch.
- Phát hiện: Phát hiện dịch vụ Windows Update (wuauserv) bị vô hiệu hóa trái phép, xuất hiện lỗi liên tiếp khi cài đặt các bản vá bảo mật quan trọng (KB patch), hoặc hệ thống lỗi thời không đồng bộ với máy chủ WSUS/Windows Update.
- Xử lý thực địa: Kiểm tra và xóa bỏ các cấu hình chặn Update trong Registry tại HKLM\SOFTWARE\Policies\Microsoft\Windows\WindowsUpdate. Cưỡng bức tái khởi động và chuyển chế độ dịch vụ về tự động bằng lệnh sc config wuauserv start= auto và net start wuauserv. Xóa thư mục bộ đệm lỗi tại C:\Windows\SoftwareDistribution để xóa các bản tải xuống bị hỏng.
- Phục hồi: Chạy công cụ kiểm tra tính toàn vẹn hệ thống sfc /scannow và DISM.