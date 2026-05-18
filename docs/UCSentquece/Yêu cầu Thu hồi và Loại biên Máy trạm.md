```mermaid

sequenceDiagram
    actor AssetMgr as 👤 Chuyên viên quản lý tài sản
    participant FE as 🖥️ Giao diện (Frontend)
    participant BE as ⚙️ Hệ thống trung tâm (Backend)
    participant DB as 🗄️ Cơ sở dữ liệu (Database)

    %% Giai đoạn 1: Truy cập phân hệ tài sản
    AssetMgr->>FE: Truy cập danh mục máy trạm
    activate FE
    FE->>BE: Yêu cầu tải dữ liệu thiết bị
    activate BE
    BE-->>FE: Phản hồi danh sách
    deactivate BE
    FE-->>AssetMgr: Hiển thị danh sách thiết bị
    deactivate FE

    %% Giai đoạn 2: Đề xuất loại biên
    AssetMgr->>FE: Chọn thiết bị cần loại biên trên giao diện
    activate FE
    FE-->>AssetMgr: Hiển thị yêu cầu nhập lý do giải trình nghiệp vụ
    deactivate FE
    
    AssetMgr->>FE: Nhập lý do gỡ bỏ và xác nhận gửi đề xuất
    activate FE
    
    FE->>BE: Gửi yêu cầu gỡ bỏ kèm lý do giải trình
    activate BE
    
    BE->>BE: 1. Kiểm tra sự tồn tại của hồ sơ chờ duyệt cũ
    BE->>BE: 2. Đóng gói thông tin và lý do vào Hồ sơ phê duyệt mới
    
    BE->>DB: Đánh dấu thiết bị sang trạng thái "Chờ loại biên" & lưu hồ sơ duyệt
    activate DB
    DB-->>BE: Xác nhận cập nhật dữ liệu thành công
    deactivate DB
    
    BE->>BE: 3. Kích hoạt cơ chế niêm phong tạm thời các thao tác cấu hình khác
    
    BE-->>FE: Phản hồi trạng thái xử lý hồ sơ thành công
    deactivate BE
    
    FE-->>AssetMgr: Cập nhật danh sách điều hành & thông báo hồ sơ đã chuyển tới cấp thẩm quyền
    deactivate FE