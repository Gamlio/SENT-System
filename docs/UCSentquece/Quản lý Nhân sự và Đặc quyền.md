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
    actor HRAdmin as 👤 Quản trị viên nhân sự
    participant FE as 🖥️ Giao diện (Frontend)
    participant BE as ⚙️ Hệ thống trung tâm (Backend)
    participant DB as 🗄️ Cơ sở dữ liệu (Database)

    %% Giai đoạn 1: Truy cập phân hệ nhân sự
    HRAdmin->>FE: Truy cập phân hệ quản lý nhân sự
    activate FE
    FE->>BE: Yêu cầu tải danh sách tài khoản
    activate BE
    BE-->>FE: Phản hồi danh sách người dùng
    deactivate BE
    FE-->>HRAdmin: Hiển thị danh sách nhân sự
    deactivate FE

    %% Giai đoạn 2: Khởi tạo hồ sơ mới
    HRAdmin->>FE: Chọn chức năng thêm thành viên mới
    activate FE
    FE-->>HRAdmin: Hiển thị form khai báo hồ sơ nhân sự
    deactivate FE
    
    HRAdmin->>FE: Khai báo danh tính, chọn nhóm đặc quyền & xác nhận
    activate FE
    
    FE->>BE: Gửi thông tin tài khoản và ma trận quyền hạn đề xuất
    activate BE
    
    BE->>BE: 1. Kiểm tra trùng lặp thông tin định danh (Tên đăng nhập / Email)
    BE->>BE: 2. Kiểm tra quy tắc kiểm soát thứ bậc đặc quyền hành chính
    
    alt Người thực hiện cấp quyền vượt quá đặc quyền của chính họ
        BE-->>FE: Từ chối yêu cầu và thông báo vi phạm chính sách cấp phát
    else Kiểm tra thứ bậc hợp lệ
        BE->>BE: 3. Thực hiện chuyển đổi mật mã tài khoản sang dạng chuỗi băm bảo mật
        BE->>DB: Lưu hồ sơ nhân sự và tự động sinh một Phiếu yêu cầu phê duyệt độc lập
        activate DB
        DB-->>BE: Xác nhận tạo tài khoản ở trạng thái khóa chờ duyệt thành công
        deactivate DB
        BE-->>FE: Phản hồi kết quả khởi tạo luồng duyệt nhân sự
    end
    deactivate BE
    
    FE-->>HRAdmin: Thông báo hồ sơ nhân sự đã được chuyển giao cho cấp quản lý cấp cao rà soát
    deactivate FE