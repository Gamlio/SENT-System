```mermaid

sequenceDiagram
    actor QTV as 👤 QTV đại diện Tổ chức
    participant FE as 🖥️ Frontend (ReactJS)
    participant BE as ⚙️ Backend (Gin API)
    participant DB as 🗄️ Database (PostgreSQL)

    QTV->>FE: Truy cập cổng Đăng ký hệ thống
    activate FE
    FE-->>QTV: Hiển thị form khai báo thông tin tổ chức và admin
    deactivate FE
    
    QTV->>FE: Khai báo thông tin đầy đủ & xác nhận đăng ký
    activate FE
    FE->>BE: POST /api/v1/workspace/register (Payload thô)
    activate BE
    
    BE->>BE: 1. Kiểm tra độ mạnh mật khẩu (>= 8 ký tự)
    BE->>BE: 2. Tự động sinh Workspace ID duy nhất
    BE->>BE: 3. Băm mật khẩu quản trị bằng Bcrypt
    
    BE->>DB: Khởi tạo Tổ chức & Tài khoản Root Admin (DB Transaction)
    activate DB
    DB-->>BE: Xác nhận lưu dữ liệu thành công
    deactivate DB
    
    BE-->>FE: Cấp phát Workspace thành công (JSON Response)
    deactivate BE
    
    FE-->>QTV: Hiển thị mã Workspace Code & chuyển hướng đăng nhập
    deactivate FE