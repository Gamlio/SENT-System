Dưới đây là toàn bộ phần cơ sở lý thuyết, công thức toán học và giải trình logic của mô hình tính điểm rủi ro tổng hợp ($Risk\_Score$) được chuyển đổi sang định dạng Markdown chuẩn hóa theo đúng nội dung đồ án tốt nghiệp của bạn:

---

## 1.5 Cơ sở toán học cho mô hình đánh giá rủi ro

Trong các giải pháp EDR truyền thống, việc đánh giá độ nguy hại của một máy trạm thường phụ thuộc vào các tập luật tĩnh hoặc đếm số lượng cảnh báo một cách tuyến tính, dẫn đến hiện tượng quá tải cảnh báo (Alert Fatigue) cho quản trị viên.

Để giải quyết triệt để bài toán này, trong khuôn khổ xây dựng hệ thống mã nguồn mở SENT, tác giả đã tự nghiên cứu, thiết kế và thực hiện cấu trúc một mô hình toán học định lượng rủi ro tổng hợp. Mô hình này không sao chép từ bất kỳ giải pháp thương mại nào, được xây dựng dựa trên sự kết hợp động giữa ba thành phần: Trạng thái sự cố tức thời, lịch sử biến động tài nguyên và bề mặt phơi nhiễm mạng của tài sản máy trạm.

$$Risk\_Score = \min\left(100.0, (R_{active} + R_{history}) \times C \times V\right)$$

---

### 1.5.1 Định lượng mức độ ưu tiên và trọng số sự cố

Thành phần rủi ro tức thời $R_{active}$ phản ánh các mối đe dọa đang diễn ra trực tiếp tại thiết bị đầu cuối. Để chuyển đổi các sự kiện an ninh từ dạng định tính sang định lượng phục vụ tính toán, tác giả thực hiện phân cấp sự cố và gán các trọng số cấu hình thực nghiệm ($T_i$) dựa trên mức độ tác động an ninh thực tế:

* 
**Sự cố mức Chí mạng (Priority P1 - Trọng số $T_{P1} = 5.0$):** Áp dụng cho các hành vi xâm nhập nghiêm trọng, mã độc fileless thực thi mã từ xa, hoặc thiết bị ngoại vi độc hại được cắm trực tiếp vào máy trạm thuộc phòng ban nhạy cảm.


* 
**Sự cố mức Cao (Priority P2 - Trọng số $T_{P2} = 2.5$):** Áp dụng cho các hành vi vi phạm chính sách an toàn thông tin nghiêm trọng, tiến trình lạ cố gắng chiếm quyền điều khiển người dùng, hoặc mở cổng mạng trái phép.


* 
**Sự cố mức Thấp / Thông tin (Priority P3 - Trọng số $T_{P3} = 1.0$):** Áp dụng cho các hành vi mang tính chất thăm dò, cài đặt ứng dụng không nằm trong danh mục cho phép, hoặc bất thường nhẹ trong tiến trình hệ thống.



> **Giải trình logic cấu hình của tác giả:**
> Các thông số trọng số này ($5.0 : 2.5 : 1.0$) là các hằng số cấu hình hệ thống (Configuration Values) được tác giả đề xuất theo mô hình lũy tiến hình học (tỷ lệ mã hóa 2:1). Khoảng cách trọng số được thiết lập đủ rộng để đảm bảo thuật toán phân tách rõ ràng mức độ ưu tiên xử lý giữa một máy trạm đang bị tấn công APT với một máy trạm chỉ phát sinh các cảnh báo thông tin thông thường. Các giá trị này được lưu trữ trong file cấu hình môi trường của hệ thống SENT, cho phép người quản trị tùy biến thay đổi tùy theo khẩu vị rủi ro của từng doanh nghiệp.
> 
> 

---

### 1.5.2 Thuật toán Rủi ro Tức thời và Quy luật Cận biên

Một thách thức lớn trong giám sát an ninh là hiện tượng "bùng nổ cảnh báo" (Alert Fatigue) khi một hành vi vi phạm lặp lại liên tục. Nếu tính toán theo cấp số cộng tuyến tính sẽ làm điểm số rủi ro bị phình to quá mức, gây nhiễu hệ thống.

Do đó, tác giả đề xuất công thức nén cảnh báo sử dụng hàm Logarit cơ số 2 dựa trên nguyên lý **"Lợi ích cận biên giảm dần"** (lỗi thứ $n$ cùng loại thường không làm tăng mức độ nguy hiểm lên gấp $n$ lần so với lỗi đầu tiên) nhằm bảo vệ tài nguyên tính toán của trung tâm điều hành SOC:

$$R_{active} = \sum_{i \in \{P1, P2, P3\}} T_i \times \log_2(n_i + 1)$$

Trong đó:

* 
$n_i$: Số lượng cảnh báo chủ động chưa xử lý thuộc nhóm ưu tiên $i$ thu thập được từ máy trạm trong chu kỳ giám sát.


* 
$T_i$: Trọng số tĩnh, cố định của từng phân cấp sự cố ($T_{P1} = 5.0, T_{P2} = 2.5, T_{P3} = 1.0$) , không nhồi các biến động thô vào $T_i$.



Việc áp dụng hàm $\log_2(n_i + 1)$ giúp đồ thị rủi ro tiệm cận dần về một ngưỡng bão hòa khi một hành vi độc hại lặp lại liên tục.

---

### 1.5.3 Mô hình rủi ro lịch sử và hàm suy giảm theo thời gian (Time Decay)

Khả năng "ghi nhớ" lịch sử bảo mật là yếu tố then chốt để phân loại các máy trạm có nguy cơ cao. Tuy nhiên, trọng số của các sự cố cũ cần giảm dần theo thời gian để phản ánh đúng trạng thái hiện tại của thiết bị. Mô hình áp dụng hàm phân thức suy giảm (Time Decay) trong chu kỳ quét 180 ngày để tính toán cấu phần rủi ro lịch sử $R_{history}$:

$$R_{history} = \sum_{j=1}^{m} \frac{H_j}{1.0 + 0.01 \times \Delta t_j}$$

Trong đó:

* 
$H_j$: Trọng số tác động lịch sử của sự cố cũ thứ $j$ (quy đổi mặc định: P1 = 5.0, P2 = 2.5, P3 = 1.0). Đối với sự cố Trung bình (P3), hệ thống ngầm định gán giá trị khởi tạo cơ sở $H_j = 1.0$.


* 
$\Delta t_j$: Khoảng thời gian tính bằng ngày kể từ khi sự cố lịch sử $j$ phát sinh cho đến thời điểm tính toán hiện tại.


* 
**Hệ số $0.01$:** Đảm bảo điểm số suy giảm một cách an toàn và giữ vết lâu (ví dụ: sau 90 ngày, một lỗi P1 cũ vẫn để lại 'vết sẹo' xấp xỉ 5 điểm rủi ro lịch sử trên hệ thống).



---

### 1.5.4 Hệ số ngữ cảnh tài sản (C) và bề mặt phơi nhiễm (V)

Điểm rủi ro cuối cùng phải được điều chỉnh bởi ngữ cảnh vận hành của tài sản (Contextual Awareness):

* 
**Hệ số quan trọng tài sản ($C$):** Tài sản được phân tầng theo mức độ quan trọng của phòng ban và chức năng vận hành:


* Thiết bị đóng vai trò là máy chủ dữ liệu (SERVER): $C = 2.0$ 


* Máy trạm của quản trị viên hệ thống (ADMIN): $C = 1.5$ 


* Máy trạm thông thường (USER): $C = 1.0$ 




* 
**Hệ số phơi nhiễm bề mặt mạng ($V$):** Quyết định khả năng máy trạm bị khai thác hoặc trở thành bàn đạp tấn công leo thang đặc quyền trong mạng nội bộ. Hệ số này do tác giả thiết kế dựa trên số lượng các cổng mạng mở kết nối thực tế, thiết bị ngoại vi không rõ nguồn gốc và biến động dữ liệu đọc ghi trên máy trạm:



$$V = 1.0 + (\text{Số cổng mạng mở} \times 0.05) + \log_{10}(n_{usb\_unknown} + 1) + \log_{10}\left(\frac{\Delta \text{Disk I/O}}{10^6} + 1\right)$$