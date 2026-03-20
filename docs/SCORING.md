# SENT SOC - MÔ HÌNH CHẤM ĐIỂM RỦI RO ĐỘNG (Dynamic Risk Scoring v4.0)

Tài liệu này mô tả thuật toán chấm điểm rủi ro (Risk Score) của hệ thống SENT SOC. 
Phiên bản 4.0 đánh dấu sự chuyển đổi từ việc "Chấm điểm từng Cảnh báo (Alert)" sang "Chấm điểm theo Hồ sơ Sự cố (Incident)", giúp phản ánh chính xác tình trạng sức khỏe của máy trạm theo thời gian thực và chống lạm phát điểm.

---

## PHẦN 1: TRIẾT LÝ THIẾT KẾ (THE PHILOSOPHY)

1. **Incident-Based (Dựa trên Sự cố):** Một máy trạm bị nhiễm virus có thể sinh ra 100 cảnh báo (Alerts) giống nhau. Hệ thống SENT sẽ gom chúng thành 1 Sự cố (Incident) duy nhất. Điểm rủi ro được tính trên Sự cố này, do đó điểm không bị cộng dồn vô lý lên hàng nghìn điểm.
2. **Real-time Cooling (Tự động hạ nhiệt):** Khi Trưởng ca SOC tiến hành điều tra và bấm "Đóng Case" (Resolved), điểm rủi ro của máy trạm sẽ ngay lập tức được tính toán lại và tụt giảm về mức an toàn.
3. **Asset-Aware (Nhận thức Tài sản):** Cùng một lỗi "Cắm USB lạ", nếu xảy ra trên máy Lễ tân thì rủi ro thấp, nhưng nếu xảy ra trên máy Domain Controller (Server) thì rủi ro cực kỳ nghiêm trọng.

---

## PHẦN 2: MÔ HÌNH TOÁN HỌC & CÁC TRỌNG SỐ

### 1. Phân loại Mức độ Sự cố ($S_{base}$)
Mỗi Sự cố (Incident) khi được tạo ra sẽ mang một mức độ nghiêm trọng (Severity/Priority) gốc:
- **Critical (P1):** 80 điểm *(VD: Mã độc, Tắt Tường lửa, Tắt Antivirus)*
- **High (P2):** 60 điểm *(VD: Ghost Registry, Mở Port 3389 trái phép)*
- **Medium (P3):** 30 điểm *(VD: Cắm USB chưa duyệt, Cài phần mềm Crack/Torrent)*
- **Low (P4):** 10 điểm *(VD: Lỗi cấu hình nhẹ)*

### 2. Trọng số Tài sản ($W_{asset}$)
Dựa trên chức vụ (Device Type) của thiết bị trong mạng Doanh nghiệp:
- **Máy chủ (SERVER - Tier 1):** $W_{asset} = 2.0$ (Nhân đôi rủi ro).
- **Máy Quản trị IT (IT_ADMIN - Tier 2):** $W_{asset} = 1.5$.
- **Máy Văn phòng chuẩn (STANDARD - Tier 3):** $W_{asset} = 1.0$.
- **Máy Khách/Lễ tân (GUEST - Tier 4):** $W_{asset} = 0.8$ (Giảm nhẹ rủi ro).

---

## PHẦN 3: CÔNG THỨC TÍNH TỔNG ĐIỂM (WEIGHTED MAX-SCORE)

Để chống lại việc cộng dồn điểm quá mức khi một máy dính nhiều Sự cố khác nhau cùng lúc, SENT SOC áp dụng công thức **Weighted Max-Score**.

**Công thức:**
$$R_{total} = \left( S_{max} + \alpha \sum S_{secondary} \right) \times W_{asset}$$

**Giải thích:**
- Lấy điểm của **Sự cố nặng nhất ($S_{max}$)** đang MỞ làm điểm mốc cơ sở.
- Các Sự cố phụ ($S_{secondary}$) đang MỞ khác chỉ đóng góp một hệ số rất nhỏ $\alpha$ (Mặc định SENT sử dụng $\alpha = 0.15$ tức $15\%$) vào tổng điểm.
- Cuối cùng nhân với Trọng số thiết bị ($W_{asset}$).
- **Giới hạn (Cap):** Điểm $R_{total}$ luôn được chặn tối đa là **100 điểm**.

---

## PHẦN 4: VÍ DỤ VẬN HÀNH THỰC TẾ (USE CASE)

**Bối cảnh:** Máy tính `PC-KETOAN` (Tier 3 -> $W_{asset} = 1.0$) đang hoạt động bình thường (0 điểm).

1. **08:00 AM:** Kế toán cắm USB lạ.
   - Hệ thống lập Incident `[USB Violation]` (P3 -> 30đ).
   - Điểm máy trạm: $30 \times 1.0 = 30$ điểm. (Màu Vàng - Cảnh báo).

2. **09:00 AM:** Virus từ USB lây vào máy, tự động tắt Tường lửa.
   - Hệ thống lập Incident `[Firewall Disabled]` (P1 -> 80đ).
   - Tính lại điểm: Lỗi nặng nhất là P1 (80). Lỗi phụ là P3 (30).
   - Công thức: $(80 + 15\% \times 30) \times 1.0 = 80 + 4.5 = 84.5$ điểm.
   - Điểm máy trạm: **85 điểm**. (Màu Đỏ - Nguy hiểm).

3. **10:00 AM:** Nhân viên SOC vào hệ thống, điều tra và đóng Case `[Firewall Disabled]`.
   - Case P1 chuyển sang Resolved (Bị loại khỏi công thức tính).
   - Máy chỉ còn Case P3 đang Open.
   - Điểm máy trạm tự động tụt về: **30 điểm**.

4. **10:30 AM:** SOC đóng nốt Case `[USB Violation]`.
   - Máy sạch bóng sự cố.
   - Điểm máy trạm: **0 điểm** (Màu Xanh - An toàn).