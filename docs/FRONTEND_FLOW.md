Luồng Truy cập (Access Flow): 
- Trang Login: Giao diện Split-screen chuyên nghiệp. Sau khi đăng nhập thành công, Token và thông tin User được lưu vào AuthContext.
- Bảo vệ tuyến đường: ProtectedRoute kiểm tra trạng thái đăng nhập và level trước khi cho phép vào các trang Dashboard hoặc quản lý.

Hiển thị theo quyền hạn (Role-based UI):
- Admin Level 1 (Global): Hiển thị menu quản lý các công ty SME, trạng thái License và sức khỏe hệ thống.
- Admin Level 2 (SME): Hiển thị danh sách Agent, phê duyệt máy mới và tạo Enroll Token cho nhân viên cài đặt.
Cấu trúc Modular CSS: Sử dụng các file CSS riêng biệt (auth.css, dashboard.css) để dễ dàng bảo trì giao diện SOC.