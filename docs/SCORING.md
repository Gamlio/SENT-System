PHẦN 1: MÔ HÌNH TOÁN HỌC CỦA THUẬT TOÁN RISK SCORINGĐể điểm số thực sự phản ánh đúng rủi ro mà không bị lạm phát (Alert Fatigue), một sự kiện bảo mật (Event) từ khi sinh ra đến khi tính vào tổng điểm của máy trạm phải đi qua 4 Trọng số (Weights).
1. Các Trọng số thành phần (The 4 Pillars of Risk)
$W_{base}$ (Base Severity - Mức độ cơ bản): Điểm gốc của hành vi vi phạm.Critical (Tắt Antivirus, Phát hiện Malware): 80 điểm.High (Cắm USB lạ, Mở Port rủi ro cao như 3389): 60 điểm.Medium (Cài phần mềm Crack/Torrent): 30 điểm.Low (Lỗi cấu hình nhỏ): 10 điểm.

$W_{time}$ (Temporal Escalation - Thang độ thời gian): Lỗi càng để lâu không ai xử lý (Helpdesk chưa vào cuộc), rủi ro càng tăng.Dưới 4 giờ: 1.0 (Chưa phạt).Từ 4h - 24h: 1.2 (Phạt x1.2).Trên 24h: 1.5 (Cảnh báo đỏ, Helpdesk đang bỏ sót việc).

$W_{freq}$ (Frequency - Tần suất vi phạm): Đánh giá tính "cố tình" của người dùng.Vi phạm lần đầu: 1.0.Lặp lại cùng 1 lỗi (VD: Cắm đi cắm lại cái USB bị cấm 5 lần trong ngày): 1.3 (Hệ số ngoan cố).

$W_{asset}$ (Asset Criticality - Tầm quan trọng của tài sản): Máy của Giám đốc (CEO) hoặc Máy chủ (Server) dính 1 lỗi Medium sẽ nguy hiểm hơn máy của nhân viên thực tập dính lỗi Medium.Máy chủ / Lãnh đạo (Tier 1): 1.5Nhân viên chính thức (Tier 2): 1.0Máy Public / Khách (Tier 3): 0.8

2. Công thức tính điểm của MỘT Sự kiện (Event Risk Score - $E$)
Điểm của một cảnh báo cụ thể sẽ biến thiên theo thời gian và tần suất:

    $$E_i = W_{base} \times W_{time} \times W_{freq}$$3.

Công thức tính Tổng điểm của Máy trạm (Total Asset Risk Score - $R$)Áp dụng mô hình Weighted Max-Score (Lấy lỗi nặng nhất làm gốc, các lỗi phụ chỉ đóng góp một phần nhỏ để chống lạm phát),
  sau đó nhân với mức độ quan trọng của thiết bị:
  
   $$R_{total} = \left( \max(E_1, E_2, ..., E_n) + \alpha \sum_{j \neq \max} E_j \right) \times W_{asset}$$
Trong đó:
$\max(E)$ là sự kiện có điểm cao nhất hiện tại.
$\alpha$ là hệ số suy giảm cho các lỗi phụ (Thường chọn $\alpha = 0.15$ hoặc $15\%$).
$R_{total}$ luôn được giới hạn (Cap) tối đa là 100 điểm.

PHẦN 2: LUỒNG VẬN HÀNH "HUMAN-IN-THE-LOOP" (Quy trình nghiệp vụ Helpdesk)
Vì Agent chỉ đóng vai trò "Camera giám sát", luồng rủi ro sẽ diễn ra như sau:
Phát hiện: Agent gửi log "Nhân viên A cài uTorrent" về Backend.

Tính điểm lần 1: Backend chạy thuật toán, gán điểm $E_1 = 30$. Điểm máy của A lên 30 (Mức Vàng).
Leo thang (Escalation): Qua 1 ngày (24h), Helpdesk lười không xử lý, nhân viên A vẫn không gỡ uTorrent. Thuật toán tự chạy lại: $E_1 = 30 \times 1.5 = 45$. Máy chuyển sang mức Cam.

Can thiệp (Helpdesk Action): Helpdesk thấy máy A đỏ chót, bèn gọi điện: "Anh A gỡ ngay phần mềm và viết báo cáo giải trình trên Web cho em".

Giảm trừ (Decay): Nhân viên gỡ app. Helpdesk lên Dashboard bấm "Đã xử lý (Resolved)". Sự kiện $E_1$ bị loại khỏi thuật toán. Điểm máy A lập tức tụt về 0.