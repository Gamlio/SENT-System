# SENTINEX SOC - MÔ HÌNH ĐÁNH GIÁ UY TÍN & RỦI RO NGỮ CẢNH (Contextual Trust & Risk Scoring v6.0)

Tài liệu này mô tả thuật toán Sentinex v6.0, được thiết kế để loại bỏ sự "mong manh" của các hệ thống chấm điểm truyền thống. Hệ thống không chỉ nhìn vào những gì đang xảy ra, mà còn đánh giá dựa trên bối cảnh bộ phận (Ngữ cảnh) và lịch sử hành vi của máy trạm trong 365 ngày (Uy tín).

## PHẦN 1: TRIẾT LÝ "LÝ LỊCH AN NINH" (THE REPUTATION PHILOSOPHY)

- **Contextual Risk (Rủi ro theo bối cảnh):** Một hành vi không có "điểm chết" cố định. Cùng một sự kiện nhưng mức độ nghiêm trọng (P1-P4) sẽ thay đổi dựa trên Department Tag (Bộ phận) của máy đó.
- **Exponential Escalation (Thang rủi ro lũy thừa):** Rủi ro không tăng tiến tuyến tính. Khi một máy trạm dính nhiều lỗi cùng lúc, khả năng máy đã bị chiếm quyền hoàn toàn tăng theo cấp số nhân ($E^n$), khiến điểm số "dựng đứng" để cảnh báo SOC.
- **Long-term Trust (Uy tín tích lũy):** Máy trạm có một "điểm uy tín" gốc. Những vi phạm trong quá khứ (1 tháng/1 năm) để lại "vết sẹo" điểm số. Một máy có tiền sử xấu sẽ bị hệ thống "lì lợm" hơn khi tính điểm an toàn.

## PHẦN 2: MA TRẬN RỦI RO THEO BỘ PHẬN (DEPARTMENT MATRIX)

Mức độ khẩn cấp (Priority) được xác định bằng phép giao giữa Loại Sensor và Tag Bộ phận của Agent.

| Loại Sự kiện (Sensor) | Nhóm DEV (Lập trình) | Nhóm FINANCE (Tài chính) | Nhóm PROD (Sản xuất) |
| :--- | :--- | :--- | :--- |
| **USB_PLUG** (Cắm USB lạ) | P4 (Chỉ ghi log) | P1 (Cấm tuyệt đối) | P2 (Nguy hiểm) |
| **PROC_START** (Chạy app lạ) | P3 (Bình thường) | P1 (Vi phạm chính sách) | P1 (Nguy cơ dừng máy) |
| **PORT_OPEN** (Mở cổng RDP/SSH)| P2 (Cần kiểm tra) | P1 (Nghiêm trọng) | P1 (Cấm tuyệt đối) |

## PHẦN 3: CÔNG THỨC TOÁN HỌC TIỆM CẬN (ASYMPTOTIC MODEL)

Để đảm bảo điểm số không bị kịch trần (100đ) quá sớm và có độ dốc thực tế, Sentinex v6 sử dụng hàm Tiệm cận Logarit.

### 1. Thành phần 1: Điểm rủi ro tức thời ($R_{current}$)
Tính dựa trên các sự cố đang MỞ (Open Incidents), áp dụng hệ số lũy thừa cho số lượng lỗi ($n$):

$$R_{current} = 100 \times \left( 1 - e^{-\frac{\sum (S_{i} \times E^{n})}{k}} \right)$$

- **$S_{i}$**: Điểm gốc của sự cố (P1=50, P2=25, P3=10, P4=5).
- **$E^{n}$**: Hệ số lũy thừa (Số lượng lỗi càng nhiều, độ dốc càng cao).
- **$k$**: Hệ số điều chỉnh độ nhạy (Enterprise chuẩn thường chọn $k=40$ đến $60$).

### 2. Thành phần 2: Nợ rủi ro dài hạn ($D_{debt}$)
Dựa trên chỉ số Trust Score (Lịch sử 1 năm) lưu trong `models.Agent`.

- **Trust Score:** Mặc định 100 điểm.
- **Trừ điểm:** Mỗi lỗi P1 trong 30 ngày qua trừ 15đ, lỗi P2 trừ 5đ.
- **Hồi phục:** Sau mỗi 7 ngày "sạch" (không có lỗi mới), máy được cộng lại 2đ uy tín.

## PHẦN 4: VÍ DỤ THỰC TẾ TRONG MÔI TRƯỜNG DOANH NGHIỆP

- **Kịch bản:** Máy trạm `PC-ACCOUNTING` (Tag: FINANCE).
- **Trạng thái ban đầu:** Máy có lịch sử sạch 1 năm (TrustScore = 100), đang hoạt động (0 điểm).
- **Sự cố 1:** Nhân viên cắm USB lạ.
  - Theo ma trận FINANCE: USB = P1 (Critical).
  - Điểm vọt lên: ~65 điểm (Cảnh báo đỏ ngay lập tức vì bối cảnh tài chính cực kỳ nhạy cảm với USB).
- **Sự cố 2 (Xảy ra đồng thời):** Tường lửa bị tắt.
  - Hệ thống nhận diện 2 lỗi P1 cùng lúc. Độ dốc lũy thừa kích hoạt.
  - Điểm vọt lên: 92 điểm (Gần kịch trần, báo động khẩn cấp cho toàn hệ thống SOC).
- **Xử lý xong (Resolved):** Admin đóng các Case.
  - $R_{current}$ về 0.
  - Tuy nhiên, TrustScore lúc này chỉ còn 70 điểm (Vết sẹo lịch sử).
- **Trên Dashboard:** Máy hiện màu Vàng (Warning) trong 30 ngày tới để SOC giám sát đặc biệt, thay vì hiện màu Xanh an toàn giả tạo.

## PHẦN 5: KHẢ NĂNG MỞ RỘNG (EXPANDABILITY)

Thuật toán được thiết kế dưới dạng Framework mở. Khi thêm các Sensor mới như:

- **Registry Monitoring:** Chỉ cần định nghĩa `REG_CHANGE` vào Ma trận rủi ro.
- **Network Traffic:** Định nghĩa `DDOS_PATTERN` và gán trọng số theo từng bộ phận.

Hệ thống chấm điểm sẽ tự động nạp các EventID này và tính toán mà không cần can thiệp vào mã nguồn lõi của bộ xử lý rủi ro.