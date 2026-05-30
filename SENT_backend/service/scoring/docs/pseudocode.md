THUẬT TOÁN QuyTrinh_XuLy_DiemRuiRo(MaDinhDanh_MayTram)

    Lấy thông tin tài sản và Phiếu đánh giá loại thiết bị từ Cơ sở dữ liệu
    Lấy danh sách các Cảnh báo an ninh đang hoạt động (Chưa xử lý)
    Lấy danh sách các Sự cố trong lịch sử hoạt động (Chu kỳ quét 180 ngày qua) 
    Lấy số lượng Cổng mạng đang ở trạng thái lắng nghe (LISTEN)
    Lấy số lượng Nhật ký hoạt động thiết bị ngoại vi (USB)
    Lấy tổng lượng dữ liệu Đọc/Ghi hiện tại của ổ đĩa (Bytes)

    // Nếu máy trạm ở trạng thái hoàn toàn "sạch" và không có phơi nhiễm
    NẾU (Không có cảnh báo active VÀ Không có cổng mạng mở VÀ Không có nhật ký USB VÀ Không có I/O ổ đĩa) THÌ
        Cập nhật Điểm rủi ro = 0.0 và Xếp hạng an ninh = "A" vào Cơ sở dữ liệu
        KẾT THÚC THUẬT TOÁN
    KẾT THÚC NEU

    Khởi tạo rActive = 0.0
    Đếm số lượng cảnh báo đang active theo từng Cấp độ ưu tiên (P1, P2, P3)

    VỚI MỖI (Cấp_Độ_Ưu_Tiên) TRONG (Danh_Sách_Đếm_Cảnh_Báo) THỰC HIỆN
        Lấy Trọng số static tương ứng (P1 = 5.0, P2 = 2.5, P3 = 1.0)
        Lấy Số lượng cảnh báo thuộc nhóm đó (n_i)
        
        // Áp dụng quy luật cận biên giảm dần dùng hàm Logarit để nén cảnh báo trùng lặp 
        rActive = rActive + Trọng_số_static * Logarit_Cơ_Số_2(Số_lượng_cảnh_báo + 1.0)
    KẾT THÚC VỚI MỖI


    Khởi tạo rHistory = 0.0

    VỚI MỖI (Sự_cố_cũ) TRONG (Danh_sách_sự_cố_lịch_sử_180_ngày) THỰC HIỆN 
        Tính Khoảng thời gian kể từ khi sự cố cũ phát sinh đến hiện tại (DaysOld)
        Xác định Trọng số tác động gốc của sự cố cũ (H_j: P1 = 5.0, P2 = 2.5, P3 = 1.0)
        
        // Áp dụng hàm phân thức suy giảm theo thời gian (Time Decay) 
        rHistory = rHistory + Trọng_số_gốc / (1.0 + 0.01 * DaysOld)
    KẾT THÚC VỚI MỖI


    // Hệ số quan trọng của tài sản C (Dựa trên kiểu thiết bị)
    Khởi tạo cFactor = 1.0
    NẾU (Loại tài sản được cấu hình có Trọng số đặc thù > 0) THÌ
        cFactor = Trọng_số_đặc_thù_của_loại_tài_sản
    KẾT THÚC NEU

    // Chuẩn hóa đơn vị dữ liệu ổ đĩa từ Bytes về Megabytes (MB) để cân bằng trọng số 
    Dung_lượng_ổ_đĩa_MB = Tổng_lượng_dữ_liệu_ổ_đĩa_Bytes / 1,000,000 

    // Tính hệ số phơi nhiễm bề mặt mạng V
    vFactor = 1.0 + 
              (Số_lượng_cổng_mạng_mở * 0.05) + 
              Logarit_Cơ_Số_10(Số_lượng_nhật_ký_USB + 1.0) + 
              Logarit_Cơ_Số_10(Dung_lượng_ổ_đĩa_MB + 1.0)


    Điểm_Số_Rủi_Ro = (rActive + rHistory) * cFactor * vFactor
    
    // Giới hạn trần rủi ro tuyệt đối không vượt quá Threshold của mô hình
    NẾU (Điểm_Số_Rủi_Ro > 100.0) THÌ
        Điểm_Số_Rủi_Ro = 100.0
    KẾT THÚC NEU

    // Phân cấp Hạng an ninh (Grade) dựa trên Điểm số tổng hợp
    Khởi tạo Hạng_An_Ninh = "A"
    NẾU Điểm_Số_Rủi_Ro > 60 THÌ Hạng_An_Ninh = "F"
    HOẶC NẾU Điểm_Số_Rủi_Ro > 35 THÌ Hạng_An_Ninh = "D"
    HOẶC NẾU Điểm_Số_Rủi_Ro > 15 THÌ Hạng_An_Ninh = "C"
    HOẶC NẾU Điểm_Số_Rủi_Ro > 5 THÌ Hạng_An_Ninh = "B"
    CÒN LẠI Hạng_An_Ninh = "A"


    // Kích hoạt kịch bản phản ứng tự động khẩn cấp thu thập pháp y nếu vượt ngưỡng nguy hiểm 
    NẾU (Điểm_Số_Rủi_Ro > 80.0) THÌ
        Ghi nhật ký cảnh báo hệ thống cần can thiệp khẩn cấp
        Gọi Lệnh chạy ngầm sang Phân_Hệ_Hành_Vi_Dịch_Vụ để:
            - Trích xuất danh sách tiến trình hệ thống hiện tại
            - Trích xuất danh sách cổng mạng và pháp y số
            - Phát định hướng thu thập pháp y mức độ Chí mạng (P1)
    KẾT THÚC NEU

    // Ghi nhận kết quả vào Hệ thống Lưu trữ trung tâm để hiển thị lên Dashboard
    Lưu các giá trị: Điểm_Số_Rủi_Ro, Hạng_An_Ninh, Mốc_thời_gian_sự_cố_mới_nhất vào Postgres
    
    // Khấu trừ điểm tin cậy (Trust Score) nếu phát sinh hành vi xâm nhập mức độ P1
    NẾU (Trong danh sách active có chứa Cảnh báo mức độ P1) THÌ
        Lấy Điểm_Tin_Cậy_Hiện_Tại của thiết bị từ Postgres
        Điểm_Tin_Cậy_Mới = Điểm_Tin_Cậy_Hiện_Tại - Khấu_Trừ_Điểm_P1_Quy_Định
        NẾU (Điểm_Tin_Cậy_Mới < 0.0) THÌ Điểm_Tin_Cậy_Mới = 0.0 KẾT THÚC NẾU
        Cập nhật Điểm_Tin_Cậy_Mới trở lại Cơ sở dữ liệu
    KẾT THÚC NEU

KẾT THÚC THUẬT TOÁN