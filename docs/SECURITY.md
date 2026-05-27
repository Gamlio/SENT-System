Chốt chặn Bảo mật (Anti-IDOR): Luôn nhắc Agent: "Mọi truy vấn tìm kiếm, cập nhật hoặc xóa trên SQL/MongoDB đều phải kẹp chặt cặp điều kiện asset_hwid và org_id để tránh lỗi rò rỉ dữ liệu chéo công ty"
Chốt chặn Bảo mật (Anti-IDOR): Luôn nhắc Agent: "Mọi truy vấn tìm kiếm, cập nhật hoặc xóa trên SQL/MongoDB đều phải kẹp chặt cặp điều kiện asset_hwid và org_id để tránh lỗi rò rỉ dữ liệu chéo công ty".

Chốt chặn Hiệu năng (Async Goroutine): Đối với các tác vụ gọi API liên dịch vụ (như gửi log sang behavior-service hoặc policy-service), ép Agent phải bọc trong một Goroutine bất đồng bộ (go func()) để tránh làm nghẽn (Block) luồng xử lý chính.

Chốt chặn Ép kiểu an toàn (Safe Type Assertion): Nhắc Agent kiểm tra kỹ dữ liệu khi bóc tách từ interface{} (dùng if val, ok := data.(type); ok) để tránh gây sập (Panic) toàn bộ hệ thống microserver khi Agent gửi sai định dạng dữ liệu.