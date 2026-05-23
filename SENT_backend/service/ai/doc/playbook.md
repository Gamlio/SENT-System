# DANH SÁCH PLAYBOOKS ỨNG PHÓ SỰ CỐ TIÊU CHUẨN CAO HỆ THỐNG SENT SYSTEM

## [PLAYBOOK-01] ỨNG PHÓ TIẾN TRÌNH ĐỘC HẠI MALICIOUS PROCESS DETECTED
Mã kịch bản: PB-01
Phân cấp sự cố: Mức độ Chí mạng Priority P1
Trọng số rủi ro: 5.0
Dấu hiệu kích hoạt: Phát hiện tệp thực thi lạ khởi chạy ngoài danh mục Whitelist, tự động trích xuất PID, Path, User, và mã băm SHA256 từ hạ tầng sự kiện nhân Microsoft Windows Kernel Process hoặc bảng tiến trình Linux.

Quy trình xử lý sự cố từng bước SOP:
- Chuẩn bị Preparation: Đồng bộ danh sách đen Blacklist SHA256 nguy hại xuống phân hệ lưu trữ cục bộ của Agent SENT.
- Phát hiện Detection: Agent bắt giữ sự kiện khởi chạy, Backend tính toán điểm rủi ro chủ động, đẩy cảnh báo khẩn cấp lên Dashboard và khởi tạo Hồ sơ sự cố mức P1.
- Phân tách Cô lập Containment: Chuyên viên SOC duyệt lệnh trực tuyến. Backend gửi chỉ thị xuống Agent thực thi quy tắc tường lửa logical tại chỗ cách ly máy trạm Isolate Host khỏi mạng nội bộ, bẻ gãy luồng càn quét ngang Lateral Movement.
- Triệt hạ Eradication Từ xa: Agent phát tín hiệu API hệ điều hành cưỡng bức hủy tiến trình độc hại Kill Process theo PID cấu hình.
- Triệt hạ Eradication Thực địa: Chuyên viên hỗ trợ kỹ thuật tiếp cận máy trạm bị nhiễm, thực hiện sao lưu tệp tin độc hại phục vụ phân tích pháp y Forensics. Tiến hành lục quét và xóa bỏ triệt để các tệp thực thi ngầm trong thư mục ẩn Temp, AppData Local, AppData Roaming. Thực hiện khóa và xóa bỏ các Registry Key độc hại cố tình tạo cơ chế duy trì thiết lập khởi động ngầm Persistence tại vị trí: HKLM Software Microsoft Windows CurrentVersion Run và HKLM Software Microsoft Windows CurrentVersion RunOnce.
- Phục hồi Recovery: Khôi phục trạng thái kết nối mạng logical cho máy trạm sau khi xác nhận cây tiến trình đã sạch hoàn toàn. Kích hoạt trình quét mã độc toàn diện tại chỗ, cập nhật chữ ký số Antivirus và theo dõi chỉ số rủi ro hệ thống trong 15 phút kế tiếp.
- Rút kinh nghiệm Lessons Learned: Cập nhật mã băm SHA256 của mẫu mã độc vừa phát hiện vào danh sách cấm chung của toàn hệ thống, nhập báo cáo giải trình thực địa và tiến hành đóng hồ sơ sự cố.

## [PLAYBOOK-02] ỨNG PHÓ CỔNG MẠNG TRÁI PHÉP UNAUTHORIZED PORT OPENED
Mã kịch bản: PB-02
Phân cấp sự cố: Mức độ Cao Priority P2
Trọng số rủi ro: 2.5
Dấu hiệu kích hoạt: Agent chụp ảnh nhanh bảng mạng Network Snapshot, phát hiện một tiến trình lạ đang mở cổng kết nối logic trái phép LISTEN không thuộc danh mục Whitelist Ports hoặc kết nối tới IP thuộc danh sách đen.

Quy trình xử lý sự cố từng bước SOP:
- Chuẩn bị Preparation: Khởi tạo danh mục 18 kiểu cổng mạng dịch vụ nhạy cảm, bao gồm các cổng quản trị quản lý từ xa TCP 21 FTP, TCP 22 SSH, TCP 23 Telnet, TCP 80 443 HTTP HTTPS lậu, TCP 445 SMB, TCP 1433 MSSQL, TCP 3306 MySQL, TCP 3389 RDP, TCP 5900 VNC, và các cổng nhận kết nối ngược Reverse Shell Backdoor phổ biến như TCP 4444, TCP 8080, TCP 8888.
- Phát hiện Detection: Server tự động ghi nhận dữ liệu kết nối dị thường, phân loại chính xác kiểu cổng vi phạm trong 18 nhóm nguy cơ, chấm điểm bề mặt phơi nhiễm mạng và khởi tạo Hồ sơ sự cố mức P2.
- Phân tách Cô lập Containment: Từ trung tâm SOC, Chuyên viên duyệt lệnh cách ly kết nối trực tuyến. Agent SENT thực thi lệnh gọi tường lửa nội tại của máy trạm Windows Firewall hoặc iptables để đóng chặn logic cổng dịch vụ vi phạm và IP đích lập tức, cô lập hoàn toàn kênh điều khiển và lệnh C2 của hacker.
- Triệt hạ Eradication Từ xa: Agent áp dụng kỹ thuật truy vết đảo ngược Reverse Mapping để định danh chính xác mã PID đang sở hữu kết nối mạng độc hại đó và tiến hành Kill Process từ xa.
- Triệt hạ Eradication Thực địa: Chuyên viên kỹ thuật di chuyển trực tiếp đến vị trí máy trạm mục tiêu. Thực hiện rà soát phần cứng vật lý, kiểm tra xem có dấu hiệu cắm các thiết bị mạng ngoại vi phần cứng trái phép như USB 4G, card Wifi lậu tạo mạng riêng độc lập để bypass hệ thống mạng công ty hay không.
- Phục hồi Recovery: Gỡ bỏ lệnh chặn card mạng sau khi kiểm tra cây tiến trình sạch hoàn toàn. Định cấu hình lại quy tắc lọc mạng cục bộ để ngăn chặn vĩnh viễn cổng logic độc hại tái diễn, đưa máy trạm về hạng an toàn.
- Rút kinh nghiệm Lessons Learned: Tổng hợp nhật ký kết nối mạng vi phạm từ cơ sở dữ liệu MongoDB, lập biên bản xử lý thực địa đối với các trường hợp cố tình vi phạm chính sách mạng và đóng ticket sự cố.

## [PLAYBOOK-03] ỨNG PHÓ THIẾT BỊ NGOẠI VI TRÁI PHÉP UNAUTHORIZED PERIPHERAL DEVICE
Mã kịch bản: PB-03
Phân cấp sự cố: Mức độ Chí mạng Priority P1
Trọng số rủi ro: 5.0
Dấu hiệu kích hoạt: Người dùng cắm một thiết bị lưu trữ ngoại vi USB lạ vào máy trạm, Agent SENT bắt giữ sự kiện phần cứng, bóc tách vân tay số của USB VID PID Serial Number và đối chiếu danh sách trắng Whitelist thất bại.

Quy trình xử lý sự cố từng bước SOP:
- Chuẩn bị Preparation: Định chuẩn và cấu hình danh sách mã nhận diện phần cứng thiết bị lưu trữ USB được phép hoạt động Whitelist VID PID Serial. Agent SENT kích hoạt trình lắng nghe sự kiện thay đổi driver thiết bị phần cứng thời gian thực.
- Phát hiện Detection: Hệ thống phát cảnh báo Popup cho người dùng tại máy trạm và khởi tạo Hồ sơ sự cố mức độ Chí mạng P1 trên hệ thống trung tâm.
- Phân tách Cô lập Containment: Kích hoạt phản xạ tự động tại chỗ Active Response: Agent lập tức phát tín hiệu điều khiển hệ điều hành để vô hiệu hóa driver logic của cổng lưu trữ USB Disable USB Storage Device Driver lập tức, chặn đứng mọi hành vi đọc ghi hoặc nạp payload từ phần cứng độc hại BadUSB.
- Triệt hạ Eradication Từ xa: Hệ thống khóa cứng không cho phép driver ngoại vi hoạt động lại thông qua cơ chế Policy tập trung.
- Triệt hạ Eradication Thực địa: Sau khi mối nguy hiểm logic đã bị cô lập bằng phần mềm, Chuyên viên hỗ trợ kỹ thuật di chuyển trực tiếp đến máy trạm mục tiêu để thu hồi thiết bị USB lạ về mặt vật lý, tịch thu tang vật phục vụ điều tra chuyên sâu. Tiến hành truy vấn kho lưu trữ lịch sử tĩnh qua Windows Registry tại các phân nhánh bất biến HKLM SYSTEM CurrentControlSet Enum USBSTOR và HKLM SYSTEM CurrentControlSet Enum USB để trích xuất toàn bộ lịch sử cắm rút phần cứng của thiết bị này trên máy trạm, phục vụ công tác giám định pháp lý.
- Phục hồi Recovery: Chạy công cụ kiểm tra tính toàn vẹn hệ thống và quét virus toàn diện tại chỗ để đảm bảo không có mã độc nào kịp lây nhiễm trước thời điểm driver bị khóa. Chỉ mở khóa logic driver nếu có yêu cầu phê duyệt Whitelist bổ sung hợp lệ từ quản lý.
- Rút kinh nghiệm Lessons Learned: Lưu trữ thông tin định danh phần cứng USB vi phạm vào hệ thống phân tích, ghi nhận biên bản bàn giao tang vật phần cứng, hoàn tất quy trình và đóng hồ sơ lịch sử sự cố.

## [PLAYBOOK-04] ỨNG PHÓ PHẦN MỀM LẬU KHÔNG ĐƯỢC PHÉP UNAUTHORIZED SOFTWARE INSTALLED
Mã kịch bản: PB-04
Phân cấp sự cố: Mức độ Thấp Thông tin Priority P3
Trọng số rủi ro: 1.0
Dấu hiệu kích hoạt: Agent SENT quét hệ thống tệp tin và tiến trình, phát hiện người dùng thực hiện cài đặt hoặc chạy ứng dụng cấm trong Software Blacklist hoặc các dạng tệp chạy ngay Portable không qua cài đặt chính thống.

Quy trình xử lý sự cố từng bước SOP:
- Chuẩn bị Preparation: Thiết lập và đồng bộ danh mục chính sách phần mềm cấm cài đặt hoặc thực thi Software Blacklist từ máy chủ xuống phân hệ quản lý của Agent SENT.
- Phát hiện Detection: Ghi nhận sự kiện, tự động tính điểm rủi ro tài sản vật lý dựa trên thuật toán Risk Scoring và khởi tạo Sự cố mức độ Thấp Thông tin P3.
- Phân tách Cô lập Containment: Agent SENT gửi lệnh từ xa cấu hình trạng thái Đình chỉ hoạt động Suspend Process của ứng dụng vi phạm, đóng băng tạm thời tiến trình giao diện phần mềm. Đẩy thông báo Popup cảnh báo vi phạm chính sách lên màn hình yêu cầu người dùng dừng tương tác.
- Triệt hạ Eradication Từ xa: Agent khóa quyền thực thi đối với tệp tin nhị phân của phần mềm lậu.
- Triệt hạ Eradication Thực địa: Chuyên viên kỹ thuật tiếp cận máy trạm tại hiện trường, chạy trình gỡ bỏ chính thống Uninstall phần mềm lậu qua trình quản lý gói của hệ điều hành. Thực hiện lục soát và xóa sạch mã độc Crack Keygen đi kèm ẩn trong các thư mục tạm, tiến hành quét sạch và gỡ bỏ các khóa cấu hình Registry rác do phần mềm lậu tạo ra tại đường dẫn HKCU Software và HKLM Software nhằm dọn sạch hoàn toàn bề mặt phơi nhiễm.
- Phục hồi Recovery: Đồng bộ lại danh mục phần mềm sạch tiêu chuẩn từ Dashboard, giải phóng trạng thái khóa thực thi, đưa máy trạm về trạng thái tuân thủ chính sách bảo mật của doanh nghiệp.
- Rút kinh nghiệm Lessons Learned: Kết xuất lịch sử cài đặt ứng dụng lỗi của người dùng lưu trữ vào MongoDB để phục vụ tổng hợp báo cáo vi phạm chính sách tài sản phần mềm, tiến hành đóng sự cố.

## [PLAYBOOK-05] ỨNG PHÓ VÔ HIỆU HÓA HỆ THỐNG BẢO VỆ FIREWALL SECURITY SERVICE DISABLEMENT
Mã kịch bản: PB-05
Phân cấp sự cố: Mức độ Chí mạng Priority P1
Trọng số rủi ro: 5.0
Dấu hiệu kích hoạt: Dịch vụ tường lửa cục bộ bị tắt Firewall Evasion hoặc các dịch vụ bảo vệ nền tảng của hệ điều hành bị ngưng hoạt động, Agent bắt giữ sự kiện thay đổi trạng thái dịch vụ tức thời.

Quy trình xử lý sự cố từng bước SOP:
- Chuẩn bị Preparation: Thiết lập tham số kiểm toán trạng thái an ninh Security Baseline và kích hoạt module giám sát trạng thái dịch vụ Service Monitor cốt lõi của Agent đối với các dịch vụ bảo vệ nền tảng của hệ điều hành.
- Phát hiện Detection: Hệ thống ghi nhận sự kiện, điểm rủi ro chủ động tăng vọt, đẩy cảnh báo đỏ về trung tâm điều hành và tự động tạo Hồ sơ sự cố mức độ Chí mạng P1.
- Phân tách Cô lập Containment: Từ trung tâm SOC, lệnh cấu hình khẩn cấp được phê duyệt đẩy xuống bất đồng bộ qua Message Broker. Agent SENT lập tức gọi lệnh hệ thống netsh advfirewall set allprofiles state on trên Windows hoặc ufw enable trên Linux để cưỡng bức tái kích hoạt lại hệ thống Tường lửa cục bộ ngay lập tức nhằm thu hẹp vùng phơi nhiễm.
- Triệt hạ Eradication Từ xa: Agent kiểm tra quyền cấu hình dịch vụ, thiết lập chế độ khởi động dịch vụ tường lửa về trạng thái tự động Automatic.
- Triệt hạ Eradication Thực địa: Chuyên viên kỹ thuật di chuyển đến vị trí máy trạm để rà soát trực tiếp nhật ký chỉnh sửa của người dùng Event Viewer. Tiến hành kiểm tra sâu các nhánh cấu hình Registry kiểm soát trạng thái tường lửa tại: HKLM SYSTEM CurrentControlSet Services SharedAccess Parameters Firewalls Policy để xác minh nguyên nhân dịch vụ bị vô hiệu hóa do người dùng cố tình tắt hay do có mã độc ẩn sâu can thiệp chỉnh sửa registry để bypass, triệt phá tận gốc dịch vụ ngầm hoặc mã độc gây xung đột.
- Phục hồi Recovery: Kiểm tra và cập nhật các quy tắc Rules tường lửa về trạng thái mặc định an toàn. Chạy trình kiểm tra bản vá lỗi hệ điều hành Patch Management còn thiếu để gia cố an ninh, đưa máy trạm về trạng thái an toàn.
- Rút kinh nghiệm Lessons Learned: Trợ lý AI Copilot thực hiện tóm tắt lại diễn biến và nguyên nhân gây tắt tường lửa dựa trên dữ liệu log thu thập, đóng sự cố và lưu bản ghi bảo toàn lịch sử tác chiến vào hệ thống.