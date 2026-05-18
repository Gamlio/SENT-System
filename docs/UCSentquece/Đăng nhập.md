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
    actor User as 👤 Thành viên SOC / QTV
    participant FE as 🖥️ Frontend (ReactJS)
    participant BE as ⚙️ Backend (Gin API)
    participant DB as 🗄️ Database (PostgreSQL)
    participant Cache as ⚡ Cache tầng đệm (Redis)

    User->>FE: Truy cập cổng Đăng nhập hệ thống
    activate FE
    FE-->>User: Hiển thị form yêu cầu điền thông tin xác thực
    deactivate FE
    
    User->>FE: Nhập Username, Password, Captcha, Workspace Code & xác nhận
    activate FE
    FE->>BE: POST /api/v1/auth/login (Payload)
    activate BE
    
    BE->>BE: 1. Xác thực mã Captcha (Chống Replay Attack)
    BE->>BE: 2. Chuẩn hóa Workspace Code về định dạng IN HOA
    
    BE->>DB: Truy xuất User thông qua Username & OrgID
    activate DB
    DB-->>BE: Trả về thông tin User + Chuỗi băm mật khẩu
    deactivate DB
    
    BE->>BE: 3. So khớp mật khẩu qua Bcrypt (CheckPasswordHash)
    BE->>BE: 4. Khởi tạo khóa phiên JWT (Thời hạn mã ký 8 giờ)
    
    BE->>Cache: Lưu trữ trạng thái phiên làm việc (Session State)
    activate Cache
    Cache-->>BE: Ghi nhận Session thành công
    deactivate Cache
    
    BE-->>FE: Trả về Access Token (JWT) và điều hướng theo đặc quyền
    deactivate BE
    
    FE-->>User: Hiển thị giao diện điều hành Dashboard trung tâm
    deactivate FE