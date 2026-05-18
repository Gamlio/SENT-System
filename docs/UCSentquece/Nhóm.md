```mermaid

sequenceDiagram
    actor OpStaff as 👤 Nhân sự cấp vận hành
    participant FE as 🖥️ Giao diện (Frontend)
    participant BE as ⚙️ Hệ thống trung tâm (Backend)
    participant DB as 🗄️ Cơ sở dữ liệu (Database)

    OpStaff->>FE: Truy cập danh mục phòng ban & chọn thêm mới hoặc cập nhật đơn vị
    activate FE
    FE-->>OpStaff: Hiển thị form khai báo thông tin
    deactivate FE
    
    OpStaff->>FE: Nhập tên đơn vị phòng ban mới và mô tả mục đích sử dụng tương ứng
    activate FE
    
    FE->>BE: Gửi thông tin cấu trúc tổ chức đề xuất thay đổi
    activate BE
    
    BE->>BE: Kiểm tra tính duy nhất của tên phòng ban trong phạm vi tổ chức hiện tại
    
    alt Phát hiện trùng tên phòng ban
        BE-->>FE: Trả về thông báo lỗi xung đột định danh phòng ban
    else Tên phòng ban là duy nhất
        BE->>BE: Xác định vai trò của người thực hiện thuộc cấp vận hành trung gian
        BE->>DB: Khởi tạo Yêu cầu phê duyệt chứa ảnh chụp toàn bộ dữ liệu thay đổi
        activate DB
        DB-->>BE: Xác nhận ghi nhận trạng thái hồ sơ chờ duyệt thành công
        deactivate DB
        BE-->>FE: Phản hồi thông tin hồ sơ chờ duyệt cấu trúc
    end
    deactivate BE
    
    FE-->>OpStaff: Hiển thị trạng thái "Chờ duyệt" và thông báo cho cấp quản lý rà soát cơ cấu
    deactivate FE