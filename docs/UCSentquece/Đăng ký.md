```mermaid
%%{init: {
  'theme': 'base',
  'themeVariables': {
    'background': '#FFFFFF',
    'actorBkg': '#F8FAFC',
    'actorBorder': '#475569',
    'actorTextColor': '#1E293B',
    'actorLineColor': '#94A3B8',
    'signalColor': '#334155',
    'signalLineColor': '#64748B',
    'labelBoxBkgColor': '#F1F5F9',
    'labelBoxBorderColor': '#CBD5E1',
    'labelTextColor': '#1E293B',
    'activationBkgColor': '#4F46E5',
    'activationBorderColor': '#4F46E5'
  }
}}%%
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