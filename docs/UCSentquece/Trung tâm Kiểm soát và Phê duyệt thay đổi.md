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
    actor Approver as 👤 Người phê duyệt chuyên trách
    participant FE as 🖥️ Cổng phê duyệt (Frontend)
    participant BE as ⚙️ Hệ thống trung tâm (Backend)
    participant DB as 🗄️ Cơ sở dữ liệu (Database)

    Approver->>FE: Truy cập vào phân hệ Trung tâm Phê duyệt
    activate FE
    
    FE->>BE: Yêu cầu tải danh sách các hồ sơ đang "Chờ xử lý"
    activate BE
    BE->>DB: Lấy các yêu cầu thay đổi (Tài sản, Nhân sự, Chính sách)
    activate DB
    DB-->>BE: Trả về danh sách hồ sơ chờ duyệt
    deactivate DB
    BE-->>FE: Phản hồi danh sách dữ liệu hồ sơ kèm mức độ ưu tiên
    deactivate BE
    
    FE-->>Approver: Hiển thị danh sách hồ sơ kèm các bộ lọc nghiệp vụ
    deactivate FE
    
    Approver->>FE: Chọn một hồ sơ cụ thể để xem bối cảnh chi tiết
    activate FE
    FE->>BE: Yêu cầu trích xuất dữ liệu nghiệp vụ đặc thù của hồ sơ
    activate BE
    BE-->>FE: Trả về thông số chi tiết tại thời điểm yêu cầu
    deactivate BE
    FE-->>Approver: Trình diện các thông số đối soát ngữ cảnh trực quan
    deactivate FE
    
    Approver->>FE: Nhập ghi chú đánh giá (bắt buộc) & nhấn lệnh "Phê duyệt"/"Từ chối"
    activate FE
    FE->>BE: Gửi quyết định xử lý hồ sơ kèm lý do giải trình
    activate BE
    
    BE->>BE: Kiểm tra xung đột xử lý (Tránh trường hợp người khác đã duyệt trước)
    
    BE->>DB: Cập nhật trạng thái hồ sơ vĩnh viễn và thực thi thay đổi tự động
    activate DB
    DB-->>BE: Xác nhận lưu trữ quyết định phê duyệt thành công
    deactivate DB
    
    BE-->>FE: Phản hồi kết quả thực thi nghiệp vụ thành công
    deactivate BE
    
    FE-->>Approver: Thông báo hoàn tất xử lý hồ sơ và đồng bộ trạng thái toàn hệ thống
    deactivate FE