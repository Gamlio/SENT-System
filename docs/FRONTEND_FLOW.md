# Luồng Hoạt Động Chi Tiết - SENT Frontend

Tài liệu này mô tả chi tiết kiến trúc và luồng hoạt động của ứng dụng frontend được xây dựng bằng React.

---

## 1. Luồng Khởi Động và Cấu Trúc (Application Bootstrap)

1.  **Entry Point (`src/index.js`):**
    *   Đây là file đầu tiên được thực thi.
    *   Nó import các file CSS toàn cục (`src/styles/index.css`).
    *   Sử dụng `ReactDOM.createRoot()` để render component chính là `App` vào thẻ `<div id="root">` trong `public/index.html`.

2.  **Component Gốc (`src/App.js`):**
    *   Đây là component cha của toàn bộ ứng dụng.
    *   **Bao bọc các Context Provider:** Toàn bộ ứng dụng được bao bọc bởi các Context Provider quan trọng từ `src/context/`, giúp chia sẻ state và logic trên toàn ứng dụng.
        *   `AuthProvider`: Cung cấp thông tin người dùng đang đăng nhập, token, và các hàm `login`, `logout`.
        *   `WebSocketProvider`: Khởi tạo và duy trì kết nối WebSocket với backend.
    *   **Thiết Lập Routing:** Sử dụng thư viện `react-router-dom` để định nghĩa các routes của ứng dụng. Các routes được bảo vệ (protected routes) sẽ kiểm tra trạng thái đăng nhập từ `AuthContext` trước khi cho phép truy cập.

---

## 2. Luồng Đăng Nhập (Authentication Flow)

1.  **Giao Diện (`src/pages/Auth/Login.jsx`):**
    *   Hiển thị form đăng nhập với username và password.
    *   Khi người dùng nhấn "Login", component này gọi hàm `login` được cung cấp bởi `AuthContext`.

2.  **Context (`src/context/AuthContext.js`):**
    *   Hàm `login(username, password)` được gọi.
    *   Bên trong hàm này, nó sử dụng `axios` (từ `src/api/axios.js`) để gửi một request `POST` đến API endpoint `/api/v1/auth/login` của backend với `username` và `password`.

3.  **API Layer (`src/api/axios.js`):**
    *   Đây là một file cấu hình một instance của `axios`.
    *   Nó có một `interceptor` tự động thêm header `Authorization: Bearer <token>` vào tất cả các request gửi đi nếu token đã tồn tại (sau khi đăng nhập thành công).
    *   Nó cũng có thể có `interceptor` để xử lý các lỗi chung, ví dụ khi nhận lỗi `401 Unauthorized` (token hết hạn), nó sẽ tự động gọi hàm `logout` và điều hướng người dùng về trang đăng nhập.

4.  **Xử Lý Kết Quả:**
    *   Nếu request thành công, backend trả về thông tin user và `accessToken`.
    *   `AuthContext` lưu các thông tin này vào state của nó và vào `localStorage` (để giữ trạng thái đăng nhập khi người dùng refresh trang).
    *   Component `Login` (hoặc `App.js`) sẽ điều hướng người dùng đến trang `Dashboard`.

---

## 3. Luồng Hiển Thị Dữ Liệu (Data Fetching & Display)

Ví dụ: Hiển thị danh sách Agents tại trang `src/pages/Agents/Agents.jsx`.

1.  **Component Tải Dữ Liệu:**
    *   Khi component `Agents.jsx` được mount (hiển thị lần đầu), nó sẽ gọi một custom hook, ví dụ `useAgents()` (định nghĩa trong `src/pages/Agents/hooks/`).

2.  **Custom Hook (`src/pages/Agents/hooks/useAgents.js`):**
    *   Hook này chịu trách nhiệm cho toàn bộ logic liên quan đến agents: state (loading, error, data), và các hàm để fetch/thêm/xóa.
    *   Nó sử dụng `useEffect` để gọi hàm `fetchAgents` khi component được mount.
    *   Hàm `fetchAgents` sẽ dùng `axios` để gửi request `GET` đến `/api/v1/agents`.
    *   Trong khi chờ dữ liệu, hook sẽ trả về `isLoading: true`. Khi có dữ liệu, `isLoading: false` và `data: [...]`. Nếu lỗi, `error: ...`.

3.  **Hiển Thị Giao Diện:**
    *   Component `Agents.jsx` nhận về state (`isLoading`, `data`, `error`) từ hook `useAgents`.
    *   Nó hiển thị một spinner nếu `isLoading` là `true`.
    *   Hiển thị thông báo lỗi nếu `error` tồn tại.
    *   Nếu có `data`, nó sẽ truyền dữ liệu này xuống các component con như `SearchableList.jsx` hoặc một table để render ra danh sách.

---

## 4. Luồng Cập Nhật Real-time (WebSocket Flow)

Ví dụ: Một agent chuyển từ `online` sang `offline`.

1.  **Backend Phát Sự Kiện:** Backend phát hiện agent `offline` và gửi một thông điệp qua WebSocket đến tất cả các client đang kết nối. Ví dụ: `{ event: 'AGENT_STATUS_CHANGED', payload: { agentId: 'xyz', status: 'offline' } }`.

2.  **Frontend Nhận Sự Kiện (`src/context/WebSocketContext.jsx`):**
    *   Provider này lắng nghe tất cả các message từ WebSocket server.
    *   Khi nhận được một message, nó sẽ phát một `CustomEvent` trên đối tượng `window` hoặc gọi một callback function đã được đăng ký.

3.  **Component Lắng Nghe (`src/pages/Agents/Agents.jsx`):**
    *   Trong component `Agents.jsx`, một hook `useSocketSubscription('AGENT_STATUS_CHANGED', callback)` được sử dụng.
    *   Hook này đăng ký `callback` function với `WebSocketContext`.
    *   Khi `WebSocketContext` nhận được sự kiện `AGENT_STATUS_CHANGED`, `callback` function này sẽ được gọi với `payload` của sự kiện.

4.  **Cập Nhật UI:**
    *   `callback` function sẽ tìm agent có `agentId: 'xyz'` trong state dữ liệu hiện tại và cập nhật trạng thái của nó thành `offline`.
    *   Vì state thay đổi, React sẽ tự động render lại component `Agents.jsx`, và người dùng sẽ thấy trạng thái của agent thay đổi ngay lập tức mà không cần làm gì cả.
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
│   ├── hooks/
│   │   └── useWebSocket.js         # Hook quản lý kết nối WebSocket Real-time tập trung
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