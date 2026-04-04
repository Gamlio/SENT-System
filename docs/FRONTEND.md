# Luồng Hoạt Động Chi Tiết - SENT Frontend

Tài liệu này đã được cập nhật và làm mới để phản ánh đúng mã nguồn hiện tại của `SENT_frontend` (React + Vite).

## TOÀN BỘ NỘI DUNG MỚI

### 1. Kiến trúc thư mục

SENT_frontend/
- public/: static assets (index.html, image, favicon).
- src/index.jsx: mount React app vào DOM.
- src/App.jsx: route definitions (public + protected), layout chung (Sidebar, Navbar, GlobalCopilotDrawer).
- src/api/axios.jsx: cấu hình axios với `baseURL` từ `VITE_API_URL`, `Content-Type`, interceptor `Authorization`.
- src/context/AuthContext.jsx: `login`, `logout`, localStorage `sent_token`, `sent_user`.
- src/context/WebSocketContext.jsx: quản lý WebSocket kết nối với `/ws`, dispatch event toàn cục.
- src/pages/: module funcional (assets, IncidentReport, Dashboard, policies, approvals, AIChat, KnowledgeBase, User, Auth).
- src/components/: chung UI component (Sidebar, Navbar, AppDialog, SearchableList, GlobalCopilotDrawer).
- src/styles/: CSS/Tailwind.

### 2. Entry point + Routing

- `src/index.jsx`: render App bên trong `AuthProvider`.
- `App.jsx`:
  - public: `/login`, `/login/:companyCode`, `/register`.
  - protected: `/`, `/profile`, `/assets`, `/assets/:hwid`, `/incidents`, `/incidents/:id`, `/chat-ai`, `/approvals`, `/policy-center`, `/docs`, `/users`.
  - `ProtectedRoute` kiểm tra `useAuth().user` và `requiredPermission`.

### 3. Auth flow

- `AuthContext.login` gọi `POST /api/v1/auth/login`.
- lưu token vào localStorage key `sent_token`, thông tin user vào `sent_user`.
- Axios interceptor tự thêm `Authorization: Bearer <token>`.
- `logout` xóa localStorage.

### 4. WebSocket

- endpoint backend: `/ws`.
- message loại: `asset_STATUS_CHANGED`, `REFRESH_asset_LIST`, `BASELINE_COMPLETED`, `INCIDENT_UPDATED`.
- component dùng `useSocketSubscription` để listen & cập nhật.

### 5. API mapping module

assets:
- GET /api/v1/assets
- GET /api/v1/assets/:hwid
- GET /api/v1/assets/:hwid/logs
- GET /api/v1/assets/stats
- GET /api/v1/assets/active-token
- POST /api/v1/assets/generate-token
- PUT /api/v1/assets/:hwid/assign
- POST /api/v1/assets/bulk-assign
- PUT /api/v1/assets/:hwid/device-type
- PUT /api/v1/assets/:hwid/department
- POST /api/v1/assets/:hwid/request-delete
- POST /api/v1/assets/bulk-request-delete
- POST /api/v1/assets/:hwid/trigger-baseline

Incidents:
- GET /api/v1/incidents
- GET /api/v1/incidents/:id
- POST /api/v1/incidents/:id/activity
- PUT /api/v1/incidents/:id/playbook
- PUT /api/v1/incidents/:id/assign
- POST /api/v1/incidents/:id/execute
- POST /api/v1/incidents/:id/ai-analyze
- GET /api/v1/files/incidents/:filename

Approvals:
- GET /api/v1/approvals
- PUT /api/v1/approvals/:id/review

Policies & Docs:
- GET /api/v1/policies
- POST /api/v1/policies
- POST /api/v1/policies/bulk
- DELETE /api/v1/policies/:id
- POST /api/v1/policies/bulk-delete
- GET /api/v1/docs
- POST /api/v1/docs/upload
- PUT /api/v1/docs/:id
- DELETE /api/v1/docs/:id

AI Chat:
- POST /api/v1/ai/chat
- POST /api/v1/ai/chat/:session_id
- POST /api/v1/ai/sessions
- GET /api/v1/ai/sessions
- PUT /api/v1/ai/sessions/:id
- DELETE /api/v1/ai/sessions/:id
- GET /api/v1/ai/chat/:session_id

Users:
- GET /api/v1/users
- POST /api/v1/users
- PUT /api/v1/users/:id
- DELETE /api/v1/users/:id

Dashboard:
- GET /api/v1/dashboard/stats

### 6. Performance

- useMemo, useCallback, tránh re-render khi hiển thị bảng lớn.
- Polling 60s + WebSocket.

### 7. Chạy và test

- npm install
- npm run dev
- npm run build
- Env: VITE_API_URL=http://localhost:8000/api/v1

### 8. Lưu ý so với cũ

- permissions logic `requiredPermission: asset_view/incident_view/policy_view/approval_manage/user_manage`.
- login response token property là `token`, không phải `access_token`.
- app có GlobalCopilotDrawer và profile route.

---

## 9. Nội dung cũ (giữ nguyên để tham khảo)


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

Ví dụ: Hiển thị danh sách assets tại trang `src/pages/assets/assets.jsx`.

1.  **Component Tải Dữ Liệu:**
    *   Khi component `assets.jsx` được mount (hiển thị lần đầu), nó sẽ gọi một custom hook, ví dụ `useassets()` (định nghĩa trong `src/pages/assets/hooks/`).

2.  **Custom Hook (`src/pages/assets/hooks/useassets.js`):**
    *   Hook này chịu trách nhiệm cho toàn bộ logic liên quan đến assets: state (loading, error, data), và các hàm để fetch/thêm/xóa.
    *   Nó sử dụng `useEffect` để gọi hàm `fetchassets` khi component được mount.
    *   Hàm `fetchassets` sẽ dùng `axios` để gửi request `GET` đến `/api/v1/assets`.
    *   Trong khi chờ dữ liệu, hook sẽ trả về `isLoading: true`. Khi có dữ liệu, `isLoading: false` và `data: [...]`. Nếu lỗi, `error: ...`.

3.  **Hiển Thị Giao Diện:**
    *   Component `assets.jsx` nhận về state (`isLoading`, `data`, `error`) từ hook `useassets`.
    *   Nó hiển thị một spinner nếu `isLoading` là `true`.
    *   Hiển thị thông báo lỗi nếu `error` tồn tại.
    *   Nếu có `data`, nó sẽ truyền dữ liệu này xuống các component con như `SearchableList.jsx` hoặc một table để render ra danh sách.

---

## 4. Luồng Cập Nhật Real-time (WebSocket Flow)

Ví dụ: Một asset chuyển từ `online` sang `offline`.

1.  **Backend Phát Sự Kiện:** Backend phát hiện asset `offline` và gửi một thông điệp qua WebSocket đến tất cả các client đang kết nối. Ví dụ: `{ event: 'asset_STATUS_CHANGED', payload: { assetId: 'xyz', status: 'offline' } }`.

2.  **Frontend Nhận Sự Kiện (`src/context/WebSocketContext.jsx`):**
    *   Provider này lắng nghe tất cả các message từ WebSocket server.
    *   Khi nhận được một message, nó sẽ phát một `CustomEvent` trên đối tượng `window` hoặc gọi một callback function đã được đăng ký.

3.  **Component Lắng Nghe (`src/pages/assets/assets.jsx`):**
    *   Trong component `assets.jsx`, một hook `useSocketSubscription('asset_STATUS_CHANGED', callback)` được sử dụng.
    *   Hook này đăng ký `callback` function với `WebSocketContext`.
    *   Khi `WebSocketContext` nhận được sự kiện `asset_STATUS_CHANGED`, `callback` function này sẽ được gọi với `payload` của sự kiện.

4.  **Cập Nhật UI:**
    *   `callback` function sẽ tìm asset có `assetId: 'xyz'` trong state dữ liệu hiện tại và cập nhật trạng thái của nó thành `offline`.
    *   Vì state thay đổi, React sẽ tự động render lại component `assets.jsx`, và người dùng sẽ thấy trạng thái của asset thay đổi ngay lập tức mà không cần làm gì cả.
SENT_frontend/
├── public/
├── src/
│   ├── api/
│   │   └── axios.js                # Cấu hình Axios, đính kèm JWT Token tự động
│   ├── components/
│   │   ├── common/                 # Component UI tái sử dụng
│   │   │    └── Pagination.jsx   
│   │   ├── GlobalCopilotDrawer.jsx
│   │   ├── Navbar.jsx              # Thanh điều hướng chính (Chứa menu assets, Docs, Policy Center)
│   │   ├── SearchableList.jsx     
│   │   ├── AppDialog.jsx     
│   │   └── Sidebar.jsx             
│   ├── context/
│   │   └── AuthContext.js          # Quản lý trạng thái Login, phân quyền (Level 1, Level 2)
│   ├── hooks/
│   │   └── useWebSocket.js         # Hook quản lý kết nối WebSocket Real-time tập trung
│   ├── pages/
│   │   ├── assets/                 # Quản lý Thiết bị
│   │   │   ├── assetDetail.jsx     # Xem chi tiết thông số
│   │   │   ├── assets.jsx          # Danh sách máy trạm
│   │   │   │── hook/
│   │   │   │    ├── useassetBulkActions.js    
│   │   │   │    ├── useassets.js    
│   │   │   │    └── useWebSocket.js    
│   │   │   └── components/
│   │   │       ├── assetBulkActions.jsx  
│   │   │       ├── GenerateTokenButton.jsx      
│   │   │       ├── assetPort.jsx 
│   │   │       ├── assetActions.jsx  
│   │   │       ├── assetUSB.jsx      
│   │   │       ├── assetLogs.jsx     
│   │   │       └── assetSoftware.jsx     
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