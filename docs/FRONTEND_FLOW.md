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
│   │   │    └── Pagination.jsx   
│   │   ├── GlobalCopilotDrawer.jsx
│   │   ├── Navbar.jsx              # Thanh điều hướng chính (Chứa menu Agents, Docs, Policy Center)
│   │   ├── SearchableList.jsx     
│   │   └── Sidebar.jsx             
│   ├── context/
│   │   └── AuthContext.js          # Quản lý trạng thái Login, phân quyền (Level 1, Level 2)
│   ├── pages/
│   │   ├── Admin/                  # Module Cấu hình (Chỉ Admin Level 2)
│   │   │   └── System/             # Hệ thống
│   │   │        ├── AdminSME.jsx
│   │   │        ├── Organizations.jsx
│   │   │        ├── Regions.jsx
│   │   │        ├── UserManagement.jsx
│   │   │        └── hooks/
│   │   │            └── useUsers.js
│   │   ├── Agents/                 # Quản lý Thiết bị
│   │   │   ├── AgentDetail.jsx     # Xem chi tiết thông số
│   │   │   ├── Agents.jsx          # Danh sách máy trạm
│   │   │   │── hook/
│   │   │   │    └── useAgents.js    
│   │   │   └── components/
│   │   │       ├── AgentActions.jsx  
│   │   │       ├── AgentUSB.jsx      
│   │   │       ├── AgentLogs.jsx     
│   │   │       └── AgentSoftware.jsx     
│   │   ├── approvals/   
│   │   │   ├── components/
│   │   │   │   └── ApprovalList.jsx
│   │   │   ├── hooks/
│   │   │   │   └── useApprovals.js
│   │   │   └── ApprovalCenter.jsx  
│   │   ├── KnowledgeBase/                 
│   │   │   ├── Documents.jsx     
│   │   │   └── hook/
│   │   │       └── useDocuments.js   
│   │   ├── IncidentReport/                 
│   │   │   ├── components/
│   │   │   │   ├── IncidentTable.jsx
│   │   │   │   └── DetailParts/
│   │   │   │        └── IncidentActionBox.jsx
│   │   │   │        └── IncidentHeader.jsx
│   │   │   │        └── IncidentInfoSidebar.jsx
│   │   │   │        └── IncidentPlaybook.jsx
│   │   │   │        └── IncidentTimeline.jsx
│   │   │   ├── IncidentManager.jsx 
│   │   │   └── IncidentDetail.jsx
│   │   ├── policies/          # Quản trị Tuân thủ & Tri thức
│   │   │   ├── components/
│   │   │   │   ├── PolicyForm.jsx
│   │   │   │   ├── PolicyList.jsx
│   │   │   │   ├── PolicyOverview.jsx
│   │   │   │   └── PolicyShared.jsx
│   │   │   ├── hooks/
│   │   │   │   ├── usePolicies.js
│   │   │   └── PolicyCenter.jsx     # Giao diện tổng hợp luật    
│   │   ├── AIChat/                 
│   │   │   ├── components/
│   │   │   │   ├── ChatSidebar.jsx
│   │   │   │   ├── ChatWindow.jsx
│   │   │   │   ├── MessageBubble.jsx
│   │   │   │   └── ThinkingBlock.jsx
│   │   │   ├── hooks/
│   │   │   │   └── useChat.js
│   │   │   └── AIChatPage.jsx
│   │   ├── Auth/                   # Xác thực
│   │   │   ├── Login.jsx
│   │   │   └── Register.jsx
│   │   └── Dashboard/              # Tổng quan
│   │       ├── components/
│   │       │   └── DashboardTable.jsx
│   │       ├── UserStats.jsx       
│   │       └── Dashboard.jsx       
│   ├── styles/                  
│   │    ├── animations.css
│   │    │── auth.css
│   │    ├── components.css
│   │    ├── index.css
│   │    └── dashboard.css
│   ├── App.js                      # Định tuyến React Router, bảo vệ Route theo Role
│   └── index.js
├── Dockerfile    
├── tailwind.config.js     
├── postcss.config.js    
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

# SENT Frontend Architecture (v3.5 - React Ecosystem)

Tài liệu mô tả luồng giao diện người dùng, tập trung vào trải nghiệm giám sát thời gian thực và quản lý danh sách lớn.

## 1. Kiến trúc Component & Hooks

### A. Custom Hooks (Logic Layer)
Tách biệt hoàn toàn logic gọi API ra khỏi giao diện để dễ tái sử dụng.
* **`useAgents.js`**: Hook quan trọng nhất.
    * *Real-time:* Tự động polling API `/agents` mỗi 30 giây.
    * *Sorting:* Xử lý sắp xếp phía Client (IP Address sorting, Hostname A-Z, Last Seen).
    * *Pagination:* Cắt dữ liệu thành các trang (Default: 8 item/trang).
    * *Filtering:* Tìm kiếm theo tên máy, IP hoặc người quản lý.
* **`useUsers.js`**: Quản lý danh sách nhân sự để phân quyền quản lý máy trạm.

### B. Core Components (UI Layer)
* **`Agents.jsx` (Danh sách máy trạm):**
    * Toolbar: Bao gồm ô tìm kiếm và **Dropdown Sắp xếp** (Mới).
    * Table: Hiển thị danh sách máy, trạng thái Online/Offline, IP Address.
    * Pagination Control: Thanh điều hướng trang ở dưới cùng.
* **`AgentDetail.jsx` (Chi tiết máy trạm):**
    * Bố cục Grid 3 cột: Thông tin phần cứng - Danh sách phần mềm - Lịch sử USB.
    * Logs Panel: Hiển thị lịch sử cảnh báo an ninh bên dưới.
* **`DashboardTable.jsx`**: Widget hiển thị tóm tắt trên trang chủ.

## 2. Các luồng nghiệp vụ chính (User Flows)

### Flow 1: Giám sát Trạng thái (Monitoring)
1.  User truy cập Menu **"Máy trạm"**.
2.  Hệ thống tải danh sách, tự động tính toán trạng thái:
    * Nếu `last_seen` < 2 phút trước -> **ONLINE** (Xanh lá).
    * Nếu `last_seen` > 2 phút trước -> **OFFLINE** (Xám).
3.  User có thể lọc nhanh bằng thanh tìm kiếm hoặc sắp xếp theo IP để gom nhóm các máy cùng phòng ban.

### Flow 2: Điều tra chi tiết (Investigation)
1.  User bấm vào tên máy trạm.
2.  Trang Detail hiện ra. User kiểm tra:
    * **Inventory:** Máy này cài Win hay Linux? Cấu hình mạnh hay yếu?
    * **Software:** Có cài phần mềm cấm (VD: uTorrent, Game) không?
    * **USB:** Có ai cắm USB lạ vào máy này gần đây không?

### Flow 3: Phân quyền quản lý (Assignment)
1.  Trên danh sách máy trạm, User bấm nút **"Thay đổi người quản lý"**.
2.  Modal hiện ra danh sách nhân viên (lấy từ `useUsers`).
3.  Chọn nhân viên -> API `PUT /agents/:hwid/assign` được gọi để cập nhật người chịu trách nhiệm.