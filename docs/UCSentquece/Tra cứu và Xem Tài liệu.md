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
    actor Analyst as 👤 Chuyên viên SOC
    participant FE as 🖥️ Giao diện (Frontend)
    participant BE as ⚙️ Hệ thống trung tâm (Backend)
    participant Storage as 🛡️ Kho lưu trữ bảo mật

    %% GIAI ĐOẠN 1: TẢI DANH SÁCH TÀI LIỆU
    Analyst->>FE: Truy cập phân hệ Trung tâm Tài liệu để tra cứu quy trình
    activate FE
    
    FE->>BE: Yêu cầu tải danh sách tài liệu hiện hành
    activate BE
    
    BE->>BE: Thực thi cơ chế tách biệt dữ liệu giữa các tổ chức
    
    BE->>Storage: Lấy danh sách văn bản ở trạng thái đã phê duyệt ban hành
    activate Storage
    Storage-->>BE: Trả về danh sách hồ sơ tài liệu hợp lệ
    deactivate Storage
    
    BE-->>FE: Phản hồi dữ liệu danh mục tài liệu
    deactivate BE
    
    FE-->>Analyst: Hiển thị danh sách văn bản quy trình (ISO, Playbooks...)
    deactivate FE %% KẾT THÚC TIẾN TRÌNH 1: Frontend giải phóng, đứng chờ user tương tác

    %% GIAI ĐOẠN 2: MỞ ĐỌC TÀI LIỆU CỤ THỂ (Phát sinh sau khi user click)
    Analyst->>FE: Nhấn chọn một tài liệu cụ thể để xem nội dung
    activate FE %% KHỞI TẠO TIẾN TRÌNH 2: Khối kích hoạt mới độc lập
    
    FE->>BE: Gửi yêu cầu truy xuất nội dung văn bản chọn xem
    activate BE
    
    BE->>Storage: Lấy bản tệp tin đã được số hóa chuyển đổi trực tuyến
    activate Storage
    Storage-->>BE: Trả về dữ tệp tin bảo mật
    deactivate Storage
    
    BE->>BE: Xác thực quyền xem văn bản của tài khoản yêu cầu
    
    BE-->>FE: Phản hồi tệp tin định dạng an toàn trực tuyến
    deactivate BE
    
    FE->>FE: Ép buộc hiển thị dưới dạng bản đọc chống chỉnh sửa nội dung gốc
    
    FE-->>Analyst: Trình diện nội dung văn bản trực quan trên màn hình điều hành
    deactivate FE %% KẾT THÚC TIẾN TRÌNH 2
```