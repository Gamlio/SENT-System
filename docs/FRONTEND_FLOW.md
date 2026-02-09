SENT Frontend Flow (React-Tailwind Edition)
Tài liệu này mô tả cấu trúc và luồng vận hành của giao diện người dùng hệ thống SENT, tối ưu cho việc quản lý đa công ty và giám sát an ninh đầu cuối.

📂 Cấu trúc Thư mục (Project Structure)
Cấu trúc này được mở rộng từ bộ khung hiện tại để hỗ trợ các tính năng quản lý tuân thủ (Compliance) và phân quyền vùng (Region).

SENT_frontend/
├── public/
│   └── index.html
├── src/
│   ├── api/
│   │   └── axios.js            # Cấu hình Axios, đính kèm JWT Token
│   ├── components/
│   │   ├── common/             # Các component dùng chung (Button, Modal, Card)
│   │   ├── Sidebar.jsx         # Menu động thay đổi theo Role Level
│   │   ├── Navbar.jsx          # Thanh điều hướng, thông báo Alert mới
│   │   └── ProtectedRoute.jsx  # Chặn truy cập trái phép cấp độ Route
│   ├── context/
│   │   └── AuthContext.js      # Quản lý trạng thái Login, Role R1-R4
│   ├── pages/
│   │   ├── Auth/
│   │   │   ├── Login.jsx       # Đăng nhập 2FA
│   │   │   └── Register.jsx    # Đăng ký SME mới (Level 3)
│   │   ├── Dashboard/
│   │   │   └── Dashboard.jsx   # Tổng quan Alert và tình trạng Agent
│   │   ├── Admin/              # Dành cho R1, R2 quản lý hệ thống
│   │   │   └── Organizations.jsx # Quản lý danh sách các công ty SME
│   │   ├── Agents/             # Quản lý thiết bị
│   │   │   ├── AgentList.jsx   # Danh sách máy trạm (Lọc theo Region)
│   │   │   └── AgentDetail/    # Chi tiết máy trạm (Tabs: HW, SW, USB, Ports)
│   │   ├── Policies/           # Quản lý tuân thủ & Truy cứu trách nhiệm
│   │   │   ├── USBWhitelist.jsx # Danh sách trắng thiết bị USB
│   │   │   └── SoftwareRules.jsx # Danh sách phần mềm cấm/cho phép
│   │   ├── Alerts/
│   │   │   └── AlertList.jsx   # Danh sách cảnh báo an ninh
│   │   └── Regions/
│   │       └── RegionList.jsx  # Quản lý vùng/văn phòng
│   ├── styles/
│   │   └── App.css             # Tailwind CSS
│   ├── App.js                  # Cấu hình React Router v6
│   └── index.js
├── tailwind.config.js          # Cấu hình giao diện
└── docker-compose.yml          # Triển khai môi trường Docker

1. Luồng Xác thực và Phân quyền (RBAC Flow)
Frontend xử lý hiển thị dựa trên role_level trả về từ Backend.

R1/R2 (Global): Thấy menu quản lý toàn bộ các Organization và License.

R3 (SME Admin): Có toàn quyền trong một công ty, tạo được User R4 và gán vùng.

R4 (SME Staff): Chỉ thấy menu Agent và Alerts thuộc các Region_ID được chỉ định.

2. Luồng Giám sát Thiết bị (Agent Monitoring)
Hiển thị dữ liệu thu thập từ Agent v3.1 một cách trực quan.

Agent Detail View: Sử dụng Tab-view để tách biệt các nhóm dữ liệu:

Hardware: CPU, RAM, OS (Dữ liệu từ bảng agent_inventory).

Software: Highlight màu đỏ các phần mềm vi phạm chính sách.

Network: Hiển thị bảng Port kèm tên tiến trình thực tế.

USB History: Danh sách thiết bị ngoại vi, đánh dấu Whitelisted vs Unauthorized.

3. Luồng Quản lý Tuân thủ (Compliance & Enforcement)
Hỗ trợ R3/R4 thiết lập luật và truy cứu trách nhiệm.

USB Whitelisting: Form nhập VID/PID/Serial và gán tên nhân viên (Assigned To) để định danh chủ sở hữu thiết bị.

Software Compliance: Dashboard so sánh danh sách phần mềm máy trạm với SoftwarePolicy để tự động tạo Ticket vi phạm.

4. Luồng Xử lý Cảnh báo (Alert Response)
Real-time Alert: Sử dụng thông báo (Toast) khi có Alert mức độ Critical đổ về từ Backend.

Audit Trail: Mỗi Alert ghi rõ thời gian (ScanTimestamp) và định danh máy (HWID) để làm bằng chứng truy cứu.