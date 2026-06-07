## TIẾN TRÌNH ĐỘC HẠI - MALICIOUS PROCESS
Từ khóa: tiến trình lạ, malicious process, tệp thực thi cấm, pid, path, sha256, blacklist.

PHÁT HIỆN: Agent SENT-SYSTEM bắt sự kiện khởi chạy tệp tin ngoài whitelist, trích xuất thông tin PID, đường dẫn và mã băm SHA256 để tạo hồ sơ cảnh báo thời gian thực.

ĐIỀU TRA: IT HD và SOC phối hợp kiểm tra lịch sử vận hành để xác định rõ: Ai là người thực thi tệp tin, Thời gian kích hoạt chính xác, Lý do người dùng tải hoặc chạy tệp tin này, và Tại sao tệp tin độc hại có thể bypass qua các bộ lọc ban đầu.

XỬ LÝ THỰC ĐỊA: Do hệ thống SENT-SYSTEM là Passive, sử dụng công cụ EDR Active từ xa hoặc đến trực tiếp máy trạm để Kill tiến trình theo PID. Tiến hành xóa triệt để tệp thực thi trong các thư mục tạm Temp, AppData Local, AppData Roaming. Khóa và xóa các Registry Key khởi động ngầm tại HKLM\Software\Microsoft\Windows\CurrentVersion\Run và RunOnce.

KHẮC PHỤC VÀ NHÂN SỰ: Yêu cầu người dùng vi phạm viết bản kiểm điểm tường trình sự việc, tổ chức đào tạo lại nhận thức an toàn thông tin. Thực hiện khôi phục kết nối mạng, kích hoạt tính năng quét mã độc toàn diện Full Scan bằng Windows Defender trên Windows 11 và cập nhật chữ ký số Antivirus mới nhất.

## CỔNG MẠNG TRÁI PHÉP - UNAUTHORIZED PORT
Từ khóa: cổng mạng, mở cổng, unauthorized port, network snapshot, listen, ip blacklist, reverse shell, tiến trình lạ.

PHÁT HIỆN: SENT-SYSTEM ghi nhận Snapshot mạng có tiến trình đang mở cổng kết nối LISTEN trái phép hoặc kết nối tới IP nguy hại thuộc danh sách đen, đặc biệt là nhóm 18 cổng nhạy cảm như TCP 21, 22, 23, 445, 3389 hoặc Reverse Shell 4444, 8080.

ĐIỀU TRA: Xác định danh tính nhân sự sở hữu máy trạm mở cổng lậu, Thời điểm cổng mạng bắt đầu mở, Lý do tự ý cấu hình hoặc cài đặt dịch vụ chiếm dụng cổng, và Tại sao không tuân thủ quy trình xin phê duyệt an ninh mạng của tổ chức.

XỬ LÝ THỰC ĐỊA: Sử dụng quyền quản trị Active EDR để đóng cổng mạng bằng cách Kill tiến trình lạ đang chiếm dụng tài nguyên ngay lập tức. Rà soát phần cứng vật lý máy trạm, kiểm tra và gỡ bỏ các thiết bị cắm ngoài bypass mạng như USB 4G hoặc Card Wifi lậu. Truy tìm vị trí file thực thi gốc để cách ly hoàn toàn.

KHẮC PHỤC VÀ NHÂN SỰ: Chuyển thông tin cho phòng nhân sự phối hợp xử lý viết bản kiểm điểm cá nhân, đưa vào danh sách đào tạo lại nhận thức. Cấu hình lại quy tắc lọc mạng cục bộ Firewall Rules trên Windows và chặn triệt để các IP Blacklist ngoại vi.

## THIẾT BỊ NGOẠI VI TRÁI PHÉP - USB PERIPHERAL
Từ khóa: usb, cắm usb, thiết bị ngoại vi, unauthorized peripheral, lưu trữ ngoại vi, vid, pid, serial number, badusb.

PHÁT HIỆN: SENT-SYSTEM phát hiện sự kiện người dùng cắm thiết bị lưu trữ USB lạ vào máy trạm và đối chiếu thấy thất bại với danh sách whitelist VID, PID, Serial Number hệ thống.

ĐIỀU TRA: Truy vết chính xác nhân sự đã cắm thiết bị, Thời gian cắm rút, Lý do sử dụng thiết bị ngoại vi không được cấp phép, và Tại sao mang thiết bị chưa qua kiểm duyệt vào môi trường nội bộ.

XỬ LÝ THỰC ĐỊA: Cán bộ IT HD trực tiếp tới hiện trường tịch thu thu hồi USB vật lý để bàn giao cho bộ phận giám định. Sử dụng EDR Active hoặc lệnh hệ điều hành truy vấn lịch sử cắm rút qua Windows Registry tại HKLM\SYSTEM\CurrentControlSet\Enum\USBSTOR và HKLM\SYSTEM\CurrentControlSet\Enum\USB để trích xuất sâu thông tin phần cứng.

KHẮC PHỤC VÀ NHÂN SỰ: Áp dụng hình thức kỷ luật phê bình, viết bản kiểm điểm và bắt buộc tham gia khóa đào tạo nhận thức an ninh. Tiến hành chạy trình quét virus Full Scan của Windows 11 trên máy trạm để kiểm tra tính toàn vẹn hệ thống sau khi rút thiết bị.

## PHẦN MỀM LẬU KHÔNG ĐƯỢC PHÉP - UNAUTHORIZED SOFTWARE
Từ khóa: phần mềm lậu, cài phần mềm, unauthorized software, phần mềm cấm, crack, keygen, portable.

PHÁT HIỆN: Bộ quét định kỳ của SENT-SYSTEM phát hiện ứng dụng nằm trong danh mục Software Blacklist hoặc các tệp tin chạy ngay Portable lậu không có bản quyền.

ĐIỀU TRA: Làm rõ ai là người cài đặt hoặc sao chép phần mềm vào máy, Thời gian cài đặt, Lý do sử dụng phần mềm bẻ khóa phục vụ công việc hay mục đích cá nhân, và Tại sao không sử dụng các phần mềm tiêu chuẩn do công ty cấp.

XỬ LÝ THỰC ĐỊA: Sử dụng công cụ quản trị Active EDR hoặc trực tiếp thao tác gỡ bỏ cài đặt Uninstall phần mềm lậu qua trình quản lý gói của hệ điều hành. Thực hiện xóa tận gốc các file crack, keygen trong thư mục tạm và dọn sạch các Registry rác còn sót lại tại HKCU\Software và HKLM\Software.

KHẮC PHỤC VÀ NHÂN SỰ: Yêu cầu nhân sự vi phạm viết bản kiểm điểm cam kết không tái phạm, hướng dẫn họ đăng ký phần mềm theo quy trình chuẩn. Đồng bộ lại danh mục phần mềm tiêu chuẩn của máy trạm về trạng thái an toàn.

## VÔ HIỆU HÓA HỆ THỐNG BẢO VỆ - FIREWALL DISABLEMENT
Từ khóa: tắt tường lửa, tắt firewall, firewall evasion, disable firewall, ngưng dịch vụ bảo vệ, service monitor.

PHÁT HIỆN: Hệ thống Passive SENT-SYSTEM đưa ra cảnh báo khẩn cấp khi dịch vụ tường lửa cục bộ Windows Firewall bị tắt hoặc các dịch vụ an ninh nền tảng bị ngưng hoạt động trái phép.

ĐIỀU TRA: Xác định ai đã thực hiện thao tác tắt dịch vụ (người dùng hay mã độc), Thời gian dịch vụ bị ngưng, Lý do tại sao phải tắt tường lửa (ví dụ để cài phần mềm lậu hoặc bypass network), và Tại sao cơ chế tự bảo vệ của máy trạm không ngăn chặn được.

XỬ LÝ THỰC ĐỊA: Kích hoạt công cụ điều khiển Active để cưỡng bức tái kích hoạt lại tường lửa Windows bằng lệnh netsh advfirewall set allprofiles state on. Kiểm tra Registry tại HKLM\SYSTEM\CurrentControlSet\Services\SharedAccess\Parameters\Firewalls\Policy để tìm và xóa bỏ các cấu hình mã độc can thiệp bypass.

KHẮC PHỤC VÀ NHÂN SỰ: Lập biên bản sự cố, yêu cầu IT HD hướng dẫn người dùng viết bản kiểm điểm và đào tạo lại nhận thức về tầm quan trọng của tường lửa. Thực hiện quy trình quét mã độc toàn diện Full Scan bằng Windows Defender để đảm bảo không có Trojan ẩn náu, sau đó chạy Patch Management cập nhật các bản vá lỗi hệ điều hành.

## CẬP NHẬT WINDOWS BỊ CHẶN HOẶC LỖI - WINDOWS UPDATE BLOCK/FAILURE
Từ khóa: cập nhật hệ điều hành, windows update, bản vá lỗi, patch management, tắt update, wuauserv, kb patch.

PHÁT HIỆN: SENT-SYSTEM cảnh báo dịch vụ Windows Update (wuauserv) bị vô hiệu hóa, xuất hiện lỗi liên tiếp khi cài đặt các bản vá bảo mật quan trọng (KB patch), hoặc hệ thống lỗi thời không đồng bộ với máy chủ WSUS.

ĐIỀU TRA: Tìm hiểu ai là người can thiệp chặn cập nhật, Thời gian hệ thống bắt đầu dừng cập nhật, Lý do chặn (ví dụ sợ lỗi win hoặc giảm hiệu năng), và Tại sao chính sách cập nhật tập trung của tổ chức bị vô hiệu hóa trên máy trạm này.

XỬ LÝ THỰC ĐỊA: Sử dụng công cụ Active quản trị để kiểm tra và xóa bỏ các cấu hình chặn Update trong Registry tại HKLM\SOFTWARE\Policies\Microsoft\Windows\WindowsUpdate. Cưỡng bức tái khởi động và chuyển chế độ dịch vụ về tự động bằng lệnh sc config wuauserv start= auto và net start wuauserv. Thực hiện dọn sạch thư mục bộ đệm lỗi tại C:\Windows\SoftwareDistribution.

KHẮC PHỤC VÀ NHÂN SỰ: Yêu cầu nhân sự vận hành máy trạm viết bản kiểm điểm giải trình, tham gia đào tạo lại quy định an toàn thiết bị đầu cuối. Tiến hành cho máy trạm vào cập nhật lại đầy đủ các bản vá, đồng thời chạy lệnh quét kiểm tra toàn diện quy tắc tường lửa Windows và tính toàn vẹn hệ thống bằng sfc /scannow.