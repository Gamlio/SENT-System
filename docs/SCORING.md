# SENTINEX SOC - MÔ HÌNH ĐÁNH GIÁ UY TÍN & RỦI RO NGỮ CẢNH (Contextual Trust & Risk Scoring v6.0)

Tài liệu này mô tả thuật toán Sentinex v6.0. Hệ thống không chỉ nhìn vào những gì đang xảy ra, mà còn đánh giá dựa trên bối cảnh bộ phận (Ngữ cảnh), lịch sử hành vi (Uy tín), và khả năng xâu chuỗi sự kiện (Event Correlation).

## PHẦN 1: TRIẾT LÝ "LÝ LỊCH AN NINH" (THE REPUTATION PHILOSOPHY)

- **Contextual Risk (Rủi ro theo bối cảnh):** Một hành vi không có "điểm chết" cố định. Mức độ nghiêm trọng (P1-P4) thay đổi dựa trên Department Tag (Bộ phận) của máy đó.
- **Event Correlation (Tương quan sự kiện):** Một cảnh báo đơn lẻ có thể là P3, nhưng nếu xảy ra liên tiếp trong 5 phút (VD: Cắm USB -> Ghi đĩa tốc độ cao), hệ thống sẽ tự động gộp thành Incident P1.
- **Exponential Escalation (Thang rủi ro lũy thừa):** Khi một máy trạm dính nhiều lỗi cùng lúc, khả năng bị chiếm quyền hoàn toàn tăng theo cấp số nhân ($E^n$), khiến điểm rủi ro "dựng đứng".
- **Long-term Trust (Uy tín tích lũy):** Máy trạm có một "điểm uy tín" gốc. Những vi phạm trong quá khứ để lại "vết sẹo". Một máy có tiền sử xấu sẽ bị hệ thống "lì lợm" hơn khi tính điểm an toàn.

## PHẦN 2: MA TRẬN RỦI RO THEO BỘ PHẬN (DEPARTMENT MATRIX)

Mức độ khẩn cấp (Priority) được xác định bằng phép giao giữa Loại Sensor và Tag Bộ phận.

| Loại Sự kiện (Sensor) | Nhóm DEV (Lập trình) | Nhóm FINANCE (Tài chính) | Nhóm PROD (Sản xuất) |
| :--- | :--- | :--- | :--- |
| **PROC_START** (Chạy app lạ) | P3 (Bình thường) | P1 (Vi phạm chính sách) | P1 (Nguy cơ dừng máy) |
| **USB_PLUG** (Cắm USB lạ) | P3 (Bình thường) | P1 (Cấm tuyệt đối) | P2 (Nguy hiểm) |
| **PORT_OPEN** (Mở cổng RDP/SSH)| P2 (Cần kiểm tra) | P1 (Nghiêm trọng) | P1 (Cấm tuyệt đối) |
| **NET_EXFILTRATION** (Gửi Data lớn)| P3 (Push Code/Docker)| P1 (Nghi đánh cắp Data) | P1 (Nghi đánh cắp Data) |
| **DISK_HOARDING** (Ghi đĩa > 500MB)| P3 (Build dự án) | P1 (Nghi Ransomware) | P2 (Bất thường) |

## PHẦN 3: CÔNG THỨC TOÁN HỌC TIỆM CẬN (ASYMPTOTIC MODEL)

Sentinex v6 sử dụng hàm Tiệm cận Logarit để tránh điểm số chạm trần 100đ quá sớm.

### 1. Điểm rủi ro tức thời ($R_{current}$)
Tính dựa trên các sự cố đang MỞ (Open Incidents), áp dụng hệ số lũy thừa cho số lượng lỗi ($n$):

$$R_{current} = 100 \times \left( 1 - e^{-\frac{\sum (S_{i} \times E^{n})}{k}} \right)$$

- **$S_{i}$**: Điểm gốc của sự cố (P1=50, P2=25, P3=10).
- **$E^{n}$**: Hệ số lũy thừa (Số lượng lỗi càng nhiều, độ dốc càng cao).
- **$k$**: Hệ số điều chỉnh độ nhạy (Enterprise chuẩn thường chọn $k=40$ đến $60$).

### 2. Nợ rủi ro dài hạn ($D_{debt}$)
Dựa trên chỉ số Trust Score (Lịch sử 1 năm) lưu trong DB.
- **Trust Score:** Mặc định 100 điểm.
- **Trừ điểm:** Mỗi lỗi P1 trong 30 ngày qua trừ 15đ, lỗi P2 trừ 5đ.
- **Hồi phục:** Sau mỗi 7 ngày "sạch", cộng lại 2đ uy tín.

## PHẦN 4: VÍ DỤ THỰC TẾ (KỊCH BẢN CHUỖI TẤN CÔNG RANSOMWARE)

- **Kịch bản:** Máy trạm `PC-ACCOUNTING` (Tag: FINANCE). Đang có TrustScore = 100.
- **Phút 01:** Nhân viên cắm USB lạ.
  - Theo ma trận FINANCE: USB = P1. Điểm vọt lên 65đ. SOC nhận cảnh báo đỏ.
- **Phút 03:** Cảm biến Viễn trắc (Telemetry) ghi nhận Disk Write tăng vọt > 1GB/phút (`DISK_HOARDING`).
  - Correlation Engine (Động cơ tương quan) kích hoạt: Gộp sự kiện `USB_PLUG` + `DISK_HOARDING` thành Siêu sự cố: **"Nghi ngờ Ransomware lây lan qua USB"**.
  - Hệ số lũy thừa $E^n$ kích hoạt. Điểm rủi ro vọt lên 98đ.
- **Phản ứng tự động:** Kênh WebSocket tự động bắn lệnh `ISOLATE` xuống asset để ngắt mạng LAN, khóa đứng máy trạm trước khi mã độc lây sang máy tính Giám đốc.
- **Kết quả:** Sau khi Admin diệt virus và mở khóa, điểm $R_{current}$ về 0. Nhưng TrustScore giảm còn 70đ. Lần sau máy này chỉ cần cắm USB là điểm tự động nhảy thẳng lên mức báo động.

## PHẦN 5: KHẢ NĂNG MỞ RỘNG (EXPANDABILITY)

Thuật toán được thiết kế dưới dạng Framework mở. Khi thêm các Cảm biến (Sensors) mới ở phía asset như:
- **Telemetry Monitoring:** Đo đếm Network Bytes Sent và Disk Bytes Written.
- **Registry Monitoring:** Định nghĩa `REG_CHANGE` vào Ma trận rủi ro.

Hệ thống chấm điểm sẽ tự động nạp các Alert mới này, chạy qua Correlation Engine và đưa vào công thức tính mà không cần sửa đổi lõi toán học của Backend.