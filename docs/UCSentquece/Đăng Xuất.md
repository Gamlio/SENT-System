```mermaid
sequenceDiagram
    actor User as 👤 Người dùng (Admin/Analyst)
    participant FE as 🖥️ Giao diện (Frontend)
    participant BE as ⚙️ Hệ thống trung tâm (Backend)
    participant Cache as ⚡ Bộ nhớ đệm phiên (Cache)

    User->>FE: Chọn lệnh Đăng xuất từ menu cá nhân
    activate FE
    
    FE->>BE: Gửi yêu cầu kết thúc phiên làm việc
    deactivate FE
    activate BE
    
    BE->>BE: Ghi nhận nhật ký về việc người dùng chủ động thoát
    
    BE->>Cache: Thu hồi quyền truy cập của phiên hiện tại
    activate Cache
    Cache-->>BE: Xác nhận thu hồi phiên thành công
    deactivate Cache
    
    BE-->>FE: Phản hồi kết thúc phiên làm việc an toàn
    activate FE
    deactivate BE
    
    %% Các hành động tự làm sạch giao diện và bộ nhớ tạm phía Client
    FE->>FE: 1. Xóa thông tin đăng nhập khỏi trình duyệt
    FE->>FE: 2. Xóa sạch các dữ liệu hiển thị sự cố tạm thời
    FE->>FE: 3. Giữ lại mã định danh tổ chức để tối ưu trải nghiệm
    
    FE-->>User: Điều hướng về cổng đăng nhập & khóa toàn bộ trang nội bộ
    deactivate FE