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
    actor QTV as 👤 IT HD
    participant FE as 🖥️ Giao diện (Frontend)
    participant BE as ⚙️ Hệ thống trung tâm (Backend)
    participant DB as 🗄️ Cơ sở dữ liệu (Database)

    %% Giai đoạn 1: Truy cập danh sách tài sản
    QTV->>FE: Truy cập phân hệ quản lý tài sản
    activate FE
    FE->>BE: Yêu cầu tải danh sách thiết bị
    activate BE
    BE-->>FE: Phản hồi danh sách thiết bị
    deactivate BE
    FE-->>QTV: Hiển thị danh sách tài sản trên giao diện
    deactivate FE

    %% Giai đoạn 2: Thay đổi phân hạng
    QTV->>FE: Chọn thiết bị hoặc nhóm thiết bị cần thay đổi phân hạng
    activate FE
    
    FE->>BE: Yêu cầu thay đổi phân hạng tài sản (SERVER, FINANCE...)
    activate BE
    
    BE->>BE: 1. Kiểm tra quyền hạn điều chuyển của người thực hiện
    BE->>BE: 2. Xác định hệ số rủi ro tương ứng của phân hạng mới
    
    BE->>DB: Cập nhật thông tin phân hạng mới của thiết bị đầu cuối
    activate DB
    DB-->>BE: Xác nhận lưu dữ liệu thành công
    deactivate DB
    
    BE->>BE: 3. Kích hoạt tiến trình tính toán lại điểm rủi ro tổng thể
    
    BE-->>FE: Thông báo hoàn tất thiết lập phân hạng
    deactivate BE
    
    FE-->>QTV: Cập nhật chỉ số rủi ro mới lên bảng điều hành thời gian thực
    deactivate FE