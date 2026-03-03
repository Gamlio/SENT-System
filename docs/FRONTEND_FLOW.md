SENT Frontend Flow (React-Tailwind Edition)
Tài liệu này mô tả cấu trúc và luồng vận hành của giao diện người dùng hệ thống SENT, tối ưu cho việc quản lý đa công ty và giám sát an ninh đầu cuối.

📂 Cấu trúc Thư mục (Project Structure)
Cấu trúc này được mở rộng từ bộ khung hiện tại để hỗ trợ các tính năng quản lý tuân thủ (Compliance) và phân quyền vùng (Region).

SENT_frontend/
├── public/
├── src/
│   ├── api/
│   │   └── axios.js                # Cấu hình Axios, đính kèm JWT Token tự động
│   ├── components/
│   │   ├── common/                 # Component UI tái sử dụng
│   │   ├── AgentActions.jsx        # Component xử lý thao tác với máy trạm
│   │   ├── Navbar.jsx              # Thanh điều hướng chính (Chứa menu Agents, Docs, Policy Center)
│   │   └── Sidebar.jsx             
│   ├── context/
│   │   └── AuthContext.js          # Quản lý trạng thái Login, phân quyền (Level 1, Level 2)
│   ├── hooks/                      # CUSTOM HOOKS (Tách biệt logic gọi API)
│   │   ├── useAgents.js            # Xử lý data cho module Agents
│   │   ├── usePolicies.js          # Xử lý data cho module Policies và Docs
│   │   └── useUsers.js             # Xử lý data cho người dùng
│   ├── pages/
│   │   ├── Admin/                  # Module Cấu hình (Chỉ Admin Level 2)
│   │   │   ├── policies/           # Quản trị Tuân thủ & Tri thức
│   │   │   │   ├── PolicyCenter.jsx     # Giao diện tổng hợp luật
│   │   │   │   ├── PolicyDocuments.jsx  # Giao diện Nạp Tài Liệu AI (Docs)
│   │   │   │   ├── SoftwarePolicies.jsx # (Giữ làm backup)
│   │   │   │   └── USBWhitelist.jsx     # (Giữ làm backup)
│   │   │   └── System/             # Hệ thống
│   │   ├── Agents/                 # Quản lý Thiết bị
│   │   │   ├── AgentDetail.jsx     # Xem chi tiết thông số
│   │   │   ├── Agents.jsx          # Danh sách máy trạm
│   │   │   └── components/
│   │   │       ├── AgentUSB.jsx      
│   │   │       ├── AgentLogs.jsx     
│   │   │       └── AgentSoftware.jsx           
│   │   ├── AI/                 
│   │   │   └── AIChatAssistant.jsx # Trợ lý ảo AI
│   │   ├── Auth/                   # Xác thực
│   │   │   ├── Login.jsx
│   │   │   └── Register.jsx
│   │   └── Dashboard/              # Tổng quan
│   │       ├── Alerts.jsx          
│   │       └── Dashboard.jsx       
│   ├── styles/
│   ├── App.js                      # Định tuyến React Router, bảo vệ Route theo Role
│   └── index.js
├── tailwind.config.js          
└── .env
1. Luồng Xác thực và Phân quyền (RBAC Flow)
Level 1 (Staff): Thấy menu Máy trạm, AI Trợ lý và Dashboard. Không có quyền thay đổi luật.

Level 2 (Admin): Mở khóa toàn bộ tính năng bao gồm Quản lý User, Trung tâm Chính sách và Nạp Luật AI. Các chức năng SME đang được ẩn tạm thời để tập trung vào luồng SOC chính.

2. Luồng Giám sát Thiết bị (Agent Monitoring)
Danh sách (Agents): Theo dõi trạng thái Online/Offline theo thời gian thực (tính toán dựa trên last_seen). Sử dụng useAgents.js để fetch data.

Chi tiết (Agent Detail): Tách biệt dữ liệu Hardware, Software, Network.

3. Luồng Quản lý Tuân thủ Đa lớp (Policy Center)
Thao tác tại một Dashboard duy nhất (PolicyCenter.jsx).

Cơ chế áp dụng: Hỗ trợ áp dụng toàn hệ thống (GLOBAL) hoặc chọn các máy trạm cụ thể (SPECIFIC).

(Các file SoftwarePolicies.jsx và USBWhitelist.jsx cũ vẫn được giữ trong thư mục nhưng luồng chính đã chuyển qua PolicyCenter).

4. Luồng Tri thức AI (Docs)
Nạp dữ liệu (PolicyDocuments.jsx): Admin tải lên tệp định dạng PDF/Word chứa nội quy, tiêu chuẩn. Quản lý logic upload qua usePolicies.js.

Truy vấn: Nhân viên sử dụng AIChatAssistant.jsx để đặt câu hỏi.