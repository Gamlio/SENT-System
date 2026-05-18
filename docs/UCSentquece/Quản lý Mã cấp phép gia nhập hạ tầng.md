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
    actor QTV as 👤 Quản trị viên SOC
    participant FE as 🖥️ Giao diện (Frontend)
    participant BE as ⚙️ Hệ thống trung tâm (Backend)
    participant DB as 🗄️ Cơ sở dữ liệu (Database)

    %% Giai đoạn 1: Truy cập màn hình
    QTV->>FE: Truy cập trung tâm điều hành máy trạm
    activate FE
    FE-->>QTV: Hiển thị giao diện quản lý thiết bị
    deactivate FE
    
    %% Giai đoạn 2: Lấy mã cấp phép
    QTV->>FE: Chọn chức năng "Lấy mã cấp phép cài đặt"
    activate FE
    
    FE->>BE: Yêu cầu truy xuất trạng thái mã cấp phép hiện hành
    activate BE
    
    BE->>DB: Tra cứu mã còn hiệu lực của tổ chức trong cơ sở dữ liệu
    activate DB
    DB-->>BE: Trả về trạng thái mã (nếu có)
    deactivate DB
    
    %% Nhánh xử lý nếu mã cũ hết hạn hoặc chưa tồn tại
    alt Mã đã hết hạn hoặc chưa tồn tại
        BE->>BE: Tự động khởi tạo chuỗi mã bảo mật mới có tiền tố định danh tổ chức
        BE->>DB: Lưu trữ mã mới kèm thời hạn hữu dụng tối đa 15 phút
        activate DB
        DB-->>BE: Xác nhận lưu trữ mã mới thành công
        deactivate DB
    end
    
    BE-->>FE: Trả về mã cấp phép hoạt động và cấu hình đồng hồ đếm ngược
    deactivate BE
    
    FE->>FE: Sử dụng font chữ đặc thù để hiển thị rõ các ký tự tương đồng
    FE-->>QTV: Trình diện mã cấp phép kèm đồng hồ đếm ngược thời gian thực
    deactivate FE
    
    %% Tiến trình tự động thu hồi khi hết chu kỳ thời gian
    Note over BE, DB: Khi đồng hồ đếm ngược chạm ngưỡng hết hạn (15 phút):
    activate BE
    BE->>DB: Ra lệnh thu hồi mã cũ và niêm phong tiến trình gia nhập của mã đó
    activate DB
    DB-->>BE: Xác nhận thu hồi hoàn tất
    deactivate DB
    deactivate BE