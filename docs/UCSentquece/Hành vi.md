```mermaid

sequenceDiagram
    actor Analyst as 👤 Chuyên viên SOC
    participant FE as 🖥️ Giao diện (Frontend)
    participant BE as ⚙️ Hệ thống trung tâm (Backend)
    participant DB as 🗄️ Cơ sở dữ liệu (Database)

    %% Giai đoạn 1: Đọc danh sách và tra cứu chi tiết bằng chứng
    Analyst->>FE: Truy cập màn hình Phân tích hành vi đáng ngờ
    activate FE
    FE->>BE: Yêu cầu tải danh sách cảnh báo hành vi
    activate BE
    BE->>DB: Truy xuất dữ liệu hành vi thuộc phạm vi tổ chức
    activate DB
    DB-->>BE: Trả về danh sách các bản ghi hành vi đáng ngờ
    deactivate DB
    BE-->>FE: Phản hồi danh sách dữ liệu (kèm mức độ phân loại)
    deactivate BE
    FE-->>Analyst: Hiển thị danh sách các thẻ hành vi vi phạm chính sách
    deactivate FE
    
    %% Giai đoạn 2: Lập hồ sơ sự cố an ninh
    Analyst->>FE: Xem chi tiết bằng chứng vi phạm & nhấn "Lập hồ sơ sự cố"
    activate FE
    FE->>BE: Gửi yêu cầu khởi tạo hồ sơ sự cố từ cảnh báo
    activate BE
    BE->>BE: Kiểm tra tính duy nhất của hồ sơ (Tránh trùng lặp dữ liệu)
    BE->>DB: Khởi tạo bản ghi sự cố mới trong hệ thống quản lý
    activate DB
    DB-->>BE: Xác nhận khởi tạo hồ sơ thành công
    deactivate DB
    BE-->>FE: Phản hồi kết quả lập hồ sơ thành công
    deactivate BE
    FE->>FE: Cập nhật trạng thái thẻ hành vi thành "Đã lập hồ sơ"
    FE-->>Analyst: Thông báo thành công và hiển thị mã số sự cố mới
    deactivate FE