```mermaid

sequenceDiagram
    actor QTV as 👤 Quản trị viên
    participant FE as 🖥️ Giao diện (Frontend)
    participant BE as ⚙️ Hệ thống trung tâm (Backend)
    participant DB as 🗄️ Cơ sở dữ liệu (Database)

    %% Giai đoạn 1: Truy cập phân hệ chính sách
    QTV->>FE: Truy cập phân hệ quản lý chính sách
    activate FE
    FE->>BE: Yêu cầu tải danh sách các chính sách hiện hành
    activate BE
    BE-->>FE: Phản hồi danh sách dữ liệu
    deactivate BE
    FE-->>QTV: Hiển thị bảng danh mục chính sách
    deactivate FE

    %% Giai đoạn 2: Thiết lập quy tắc
    QTV->>FE: Chọn lệnh thêm mới quy tắc an ninh
    activate FE
    FE-->>QTV: Hiển thị giao diện thiết lập chính sách
    deactivate FE
    
    QTV->>FE: Nhập lẻ định danh hoặc tải lên tệp tin chứa danh sách hàng loạt
    activate FE
    
    FE->>BE: Gửi danh sách định danh quy tắc cần thiết lập
    activate BE
    
    BE->>BE: 1. Tự động bóc tách và kiểm tra tính hợp lệ của định dạng chuỗi nhập vào
    BE->>BE: 2. Loại bỏ các dữ liệu sai quy chuẩn định dạng cấu trúc ngay tại chỗ
    BE-->>FE: Phản hồi danh sách dữ liệu hợp lệ
    deactivate BE
    FE-->>QTV: Hiển thị danh sách để tiếp tục cấu hình
    deactivate FE
    
    QTV->>FE: Cấu hình loại quy tắc (Cho phép/Ngăn chặn) & xác nhận lưu
    activate FE
    
    FE->>BE: Gửi yêu cầu lưu trữ cấu hình thiết lập chính sách an ninh
    activate BE
    
    BE->>DB: Ghi nhận thông tin quy tắc chính sách mới vào hệ thống
    activate DB
    DB-->>BE: Xác nhận ghi dữ liệu thành công
    deactivate DB
    
    BE->>BE: 3. Khởi tạo Yêu cầu phê duyệt tự động gửi tới trung tâm kiểm soát
    
    BE-->>FE: Phản hồi thông báo lưu cấu hình thành công
    deactivate BE
    
    FE-->>QTV: Hiển thị chính sách mới ở trạng thái "Đóng băng / Chờ duyệt"
    deactivate FE