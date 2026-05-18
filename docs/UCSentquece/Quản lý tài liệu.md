```mermaid

sequenceDiagram
    actor Mgr as 👤 Người quản lý tài liệu
    participant FE as 🖥️ Giao diện (Frontend)
    participant BE as ⚙️ Hệ thống trung tâm (Backend)
    participant DB as 🗄️ Cơ sở dữ liệu (Database)

    %% Giai đoạn 1: Truy cập phân hệ
    Mgr->>FE: Truy cập phân hệ Quản lý tài liệu
    activate FE
    FE->>BE: Yêu cầu tải danh sách tài liệu
    activate BE
    BE->>DB: Truy xuất dữ liệu
    activate DB
    DB-->>BE: Trả về danh sách
    deactivate DB
    BE-->>FE: Phản hồi danh mục tài liệu
    deactivate BE
    FE-->>Mgr: Hiển thị danh sách tài liệu hiện hành
    deactivate FE

    %% Giai đoạn 2: Đề xuất thay đổi
    Mgr->>FE: Chọn tài liệu cụ thể & nhấn lệnh "Cập nhật" hoặc "Xóa bỏ"
    activate FE
    FE-->>Mgr: Hiển thị form bắt buộc khai báo lý do thay đổi
    deactivate FE
    
    Mgr->>FE: Thay đổi thông tin (nếu sửa) + nhập lý do giải trình & xác nhận
    activate FE
    
    FE->>BE: Gửi yêu cầu thay đổi tài liệu kèm lý do giải trình chi tiết
    activate BE
    
    BE->>BE: 1. Kiểm tra lý do giải trình (Ngăn chặn xử lý hồ sơ trống)
    BE->>BE: 2. Ghi nhận hành động và lý do vào nhật ký giám sát để lưu vết
    
    BE->>DB: Đóng gói thông tin cũ/mới và lưu thành một Hồ sơ phê duyệt thay đổi
    activate DB
    DB-->>BE: Xác nhận khởi tạo hồ sơ phê duyệt thành công
    deactivate DB
    
    BE->>BE: 3. Khóa trạng thái tài liệu hiện tại sang "Chờ duyệt thay đổi"
    
    BE-->>FE: Phản hồi kết quả khởi tạo luồng xử lý thành công
    deactivate BE
    
    FE->>FE: Thiết lập hiệu ứng đóng băng tạm thời các lệnh thao tác trên tài liệu này
    FE-->>Mgr: Thông báo tài liệu đã chuyển sang trạng thái chờ cấp quản lý phê chuẩn
    deactivate FE