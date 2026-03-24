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
│   │   ├── AppDialog.jsx     
│   │   └── Sidebar.jsx             
│   ├── context/
│   │   └── AuthContext.js          # Quản lý trạng thái Login, phân quyền (Level 1, Level 2)
│   ├── pages/
│   │   ├── Agents/                 # Quản lý Thiết bị
│   │   │   ├── AgentDetail.jsx     # Xem chi tiết thông số
│   │   │   ├── Agents.jsx          # Danh sách máy trạm
│   │   │   │── hook/
│   │   │   │    ├── useAgentBulkActions.js    
│   │   │   │    ├── useAgents.js    
│   │   │   │    └── useWebSocket.js    
│   │   │   └── components/
│   │   │       ├── AgentBulkActions.jsx  
│   │   │       ├── GenerateTokenButton.jsx      
│   │   │       ├── AgentPort.jsx 
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
│   │   │   │   └── DetailParts/
│   │   │   │        ├── IncidentActionBox.jsx
│   │   │   │        ├── IncidentHeader.jsx
│   │   │   │        ├── IncidentInfoSidebar.jsx
│   │   │   │        ├── IncidentPlaybook.jsx
│   │   │   │        ├── IncidentTable.jsx
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
│   │   ├── User/                 
│   │   │   ├── components/
│   │   │   │   ├── UserDeleteModal.jsx
│   │   │   │   ├── UserFormModal.jsx
│   │   │   │   ├── UserTable.jsx
│   │   │   │   └── UserToolbar.jsx
│   │   │   ├── hooks/
│   │   │   │   └── useChat.js
│   │   │   └── UserManagement.jsx
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
# SENT Frontend Flow (v4.5 - Real-time SOC Interface)

## 1. Cơ chế Đồng bộ Dữ liệu (State Management)
- **Hybrid Sync:** * Sử dụng Polling (30-60s) làm phương án dự phòng (Fallback).
    * Sử dụng **WebSocket (useWebSocket hook)** làm phương thức chính để cập nhật UI tức thì.
- **Event Listeners:** UI lắng nghe các sự kiện `REFRESH_DATA`, `AGENT_STATUS_CHANGED` để tự động gọi hàm `fetch()` mà người dùng không cần F5.

## 2. Luồng nghiệp vụ Zero-Trust (User Flow)
- **Bước 1:** Truy cập Agent Detail -> Nhấn "Thiết lập Baseline".
- **Bước 2:** UI hiển thị trạng thái `SCANNING` (Amber Badge - Pulse animation).
- **Bước 3:** Khi nhận được tín hiệu WebSocket `BASELINE_COMPLETED`, UI tự động hiển thị danh sách phần mềm đã vào Whitelist và bật Badge `ZERO TRUST ACTIVE`.

## 3. Kiểm soát Trạng thái Chờ (Maker-Checker UX)
- **Visual Feedback:** Các bản ghi (Agent/User/Policy) đang trong trạng thái `PENDING_DELETE` hoặc `PENDING_APPROVAL` sẽ bị làm mờ (Opacity-50), áp dụng bộ lọc Grayscale và khóa mọi tương tác Write.
- **Bulk Actions:** Thanh công cụ nổi (Floating Toolbar) xuất hiện khi có >= 1 item được chọn, cho phép gửi lệnh hàng loạt qua tầng Service Brain.

## 4. Cấu trúc Thư mục Nâng cao
- `src/hooks/useWebSocket.js`: Quản lý kết nối tập trung, tự động reconnect và đính kèm token vào URL.
- `src/pages/Agents/hooks/useAgentBulkActions.js`: Tách rời logic chọn/xóa hàng loạt khỏi UI Component.
