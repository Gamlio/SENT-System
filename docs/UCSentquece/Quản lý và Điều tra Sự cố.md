```mermaid

sequenceDiagram
    actor Analyst as 👤 Chuyên viên SOC
    participant FE as 🖥️ Giao diện (Frontend)
    participant BE as ⚙️ Hệ thống trung tâm (Backend)
    participant DB as 🗄️ Cơ sở dữ liệu (Database)

    %% Giai đoạn 1: Thu thập bối cảnh sự cố
    Analyst->>FE: Chọn hồ sơ sự cố cần tiến hành điều tra chuyên sâu
    activate FE
    FE->>BE: Yêu cầu tải toàn bộ thông tin bối cảnh vụ việc
    activate BE
    BE->>DB: Truy xuất thông tin thiết bị, lịch sử cảnh báo & dòng thời gian pháp y
    activate DB
    DB-->>BE: Trả về dữ liệu chi tiết vụ việc
    deactivate DB
    BE-->>FE: Phản hồi thông tin sự cố sắp xếp theo mức độ ưu tiên ngữ cảnh
    deactivate BE
    FE-->>Analyst: Hiển thị giao diện điều hành điều tra sự cố tập trung
    deactivate FE
    
    %% Giai đoạn 2: Củng cố hồ sơ và niêm phong dữ liệu
    Analyst->>FE: Nhập nhận định chuyên môn & tải lên tệp tin hình ảnh bằng chứng
    activate FE
    FE->>BE: Yêu cầu cập nhật thông tin và niêm phong bằng chứng mới
    activate BE
    BE->>BE: Tự động chạy thuật toán bảo toàn tính toàn vẹn dữ liệu
    BE->>DB: Lưu trữ chuỗi bằng chứng không thể đứt gãy vào phân vùng bảo mật
    activate DB
    DB-->>BE: Xác nhận lưu trữ bằng chứng thành công
    deactivate DB
    BE-->>FE: Phản hồi dòng thời gian điều tra hiện tại
    deactivate BE
    FE-->>Analyst: Hiển thị bằng chứng mới trên dòng thời gian kèm chứng nhận toàn vẹn
    deactivate FE