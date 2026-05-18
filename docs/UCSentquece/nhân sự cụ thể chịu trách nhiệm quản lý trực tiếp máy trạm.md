```mermaid

sequenceDiagram
    actor QTV as 👤 Quản trị viên SOC
    participant FE as 🖥️ Giao diện (Frontend)
    participant BE as ⚙️ Hệ thống trung tâm (Backend)
    participant DB as 🗄️ Cơ sở dữ liệu (Database)

    %% Giai đoạn 1: Truy cập phân hệ tài sản
    QTV->>FE: Truy cập phân hệ quản lý máy trạm
    activate FE
    FE->>BE: Yêu cầu tải danh sách thiết bị
    activate BE
    BE-->>FE: Phản hồi danh sách thiết bị
    deactivate BE
    FE-->>QTV: Hiển thị danh sách tài sản
    deactivate FE

    %% Giai đoạn 2: Phân công nhân sự
    QTV->>FE: Chọn các thiết bị cần phân công & nhấn lệnh điều chuyển
    activate FE
    
    FE->>BE: Yêu cầu tải danh sách nhân sự đủ năng lực tiếp nhận
    activate BE
    BE->>DB: Tra cứu danh sách nhân sự thuộc cùng tổ chức hợp lệ
    activate DB
    DB-->>BE: Trả về danh sách thông tin nhân sự
    deactivate DB
    BE-->>FE: Phản hồi thông tin danh mục nhân sự khả dụng
    deactivate BE
    
    FE-->>QTV: Trình diện danh sách nhân sự trên màn hình điều hành
    deactivate FE
    
    QTV->>FE: Chọn nhân sự đích & xác nhận lệnh phân công người quản lý
    activate FE
    FE->>BE: Gửi yêu cầu chuyển giao trách nhiệm tài sản
    activate BE
    
    BE->>BE: 1. Kiểm tra đặc quyền điều chuyển tài sản của người thực hiện
    BE->>BE: 2. Kiểm tra ranh giới dữ liệu để chống gán chéo tổ chức
    
    BE->>DB: Cập nhật mối quan hệ sở hữu mới và ghi lịch sử điều chuyển
    activate DB
    DB-->>BE: Xác nhận cập nhật thông tin sở hữu thành công
    deactivate DB
    
    BE-->>FE: Phản hồi kết quả chuyển giao hoàn tất
    deactivate BE
    
    FE-->>QTV: Cập nhật tên người chịu trách nhiệm mới lên bảng điều hành thời gian thực
    deactivate FE