# 🛡️ SENT - AI-Powered Endpoint Detection & Response (EDR)

Hệ thống giám sát an ninh đầu cuối và quản trị chính sách tập trung, tích hợp **Trợ lý AI Copilot** hỗ trợ phân tích và phản ứng sự cố theo tiêu chuẩn SOC hiện đại.

##  Tổng quan hệ thống
SENT-SYSTEM (Sentinex)
Hệ thống Quản trị An ninh tập trung (SOC) & Tuân thủ rủi ro (GRC) tích hợp Trợ lý AI.

📌 Tổng quan
SENT-SYSTEM là một nền tảng giám sát an ninh mạng dành cho các doanh nghiệp SME, kết hợp giữa việc thu thập dữ liệu viễn trắc (Telemetry) từ máy trạm và khả năng tư vấn của Trí tuệ nhân tạo (AI Copilot). Hệ thống giúp đội ngũ IT HD/SOC phát hiện sớm các vi phạm chính sách, lệch chuẩn Baseline và nhận được hướng dẫn xử lý sự cố theo thời gian thực.

✨ Tính năng cốt lõi
Giám sát thiết bị (Asset Monitoring): Thu thập thông tin phần cứng, phần mềm, trạng thái cổng mạng, USB và lưu lượng I/O theo thời gian thực.

Quản lý tuân thủ (Baseline & Policy): Định nghĩa cấu hình chuẩn (Baseline) và các chính sách Whitelist/Blacklist để phát hiện sai lệch an ninh.

Trợ lý AI Copilot (RAG-based): Sử dụng Local LLM (Ollama) kết hợp với ngữ cảnh từ Chính sách và Playbook để tư vấn hướng xử lý sự cố cho IT HD.

Quản lý sự cố (Incident Management): Quy trình xử lý sự cố chặt chẽ với cơ chế xác minh tính toàn vẹn của bằng chứng (Hash Chain).

Quy trình phê duyệt (Maker-Checker): Mọi thay đổi nhạy cảm về hệ thống, tài liệu hoặc người dùng đều phải được Admin phê duyệt.

Phân tích rủi ro (Risk & Trust Scoring): Tự động tính toán điểm rủi ro tức thời và điểm uy tín dài hạn cho từng thiết bị.

🏗 Kiến trúc hệ thống
Hệ thống bao gồm hai thành phần chính:

SENT Backend (Golang): Trung tâm điều phối, xử lý dữ liệu viễn trắc, quản lý cơ sở dữ liệu (PostgreSQL & MongoDB) và giao tiếp với AI API.

Go-SENT  (Golang): Bộ thu thập dữ liệu siêu nhẹ (Ninja Thin-Client) chạy trên máy trạm Windows/Linux để đẩy dữ liệu về Server.

🛠 Công nghệ sử dụng
Ngôn ngữ: Golang (Gin Framework).

Cơ sở dữ liệu:

PostgreSQL: Quản lý cấu hình, người dùng, chính sách (Dữ liệu quan hệ).

MongoDB: Lưu trữ nhật ký sự cố, log viễn trắc, audit logs (Dữ liệu lớn).

AI: Ollama API (Model: Qwen/Phi3).

Bảo mật: JWT Authentication, HMAC Payload Signing, Hash Chaining.

Khác: WebSockets (Real-time updates), Redis (Planned for Caching)
```
SENT-System
├─ docker-compose.yml
├─ docs
│  ├─ BACKEND.md
│  ├─ DATABASE_GUIDE.md
│  ├─ DOCKER.md
│  ├─ FRONTEND.md
│  ├─ SCORING.md
│  ├─ SENT_FLOW.md
│  └─ UCSentquece
│     ├─ Hành vi.md
│     ├─ nhân sự cụ thể chịu trách nhiệm quản lý trực tiếp máy trạm.md
│     ├─ Nhóm.md
│     ├─ Phân loại tài sản.md
│     ├─ Quản lý Chính sách.md
│     ├─ Quản lý Cơ cấu Tổ chức.md
│     ├─ Quản lý Mã cấp phép gia nhập hạ tầng.md
│     ├─ Quản lý Nhân sự và Đặc quyền.md
│     ├─ Quản lý tài liệu.md
│     ├─ Quản lý và Điều tra Sự cố.md
│     ├─ Tra cứu và Xem Tài liệu.md
│     ├─ Trung tâm Kiểm soát và Phê duyệt thay đổi.md
│     ├─ Trợ lý Phân tích và Phản ứng An ninh.md
│     ├─ Yêu cầu Thu hồi và Loại biên Máy trạm.md
│     ├─ Đăng ký.md
│     ├─ Đăng nhập.md
│     └─ Đăng Xuất.md
├─ nginx.conf
├─ README.md
├─ SENT
│  ├─ build.sh
│  ├─ cmd
│  │  └─ main.go
│  ├─ go.mod
│  ├─ go.sum
│  └─ internal
│     ├─ collector
│     │  ├─ antivirus.go
│     │  ├─ antivirus_unix.go
│     │  ├─ antivirus_windows.go
│     │  ├─ data_transfer.go
│     │  ├─ firewall.go
│     │  ├─ firewall_unix.go
│     │  ├─ firewall_windows.go
│     │  ├─ localdb.go
│     │  ├─ network.go
│     │  ├─ registry.go
│     │  ├─ software.go
│     │  ├─ software_darwin.go
│     │  ├─ software_linux.go
│     │  ├─ software_windows.go
│     │  ├─ system.go
│     │  ├─ usb.go
│     │  ├─ usb_darwin.go
│     │  ├─ usb_linux.go
│     │  └─ usb_windows.go
│     ├─ config
│     │  └─ config.go
│     ├─ policy
│     │  └─ policy_manager.go
│     ├─ transport
│     │  └─ client.go
│     └─ utils
│        ├─ helpers.go
│        ├─ privilege_unix.go
│        ├─ privilege_windows.go
│        └─ utils.go
├─ SENT_backend
│  ├─ Dockerfile
│  ├─ go.mod
│  ├─ go.sum
│  └─ service
│     ├─ ai
│     │  ├─ cmd
│     │  │  └─ main.go
│     │  ├─ Dockerfile
│     │  └─ internal
│     │     ├─ handlers
│     │     │  └─ chat_handler.go
│     │     └─ service
│     │        ├─ chat_service.go
│     │        └─ vector_store.go
│     ├─ approvals
│     │  ├─ cmd
│     │  │  └─ main.go
│     │  ├─ Dockerfile
│     │  └─ internal
│     │     ├─ handlers
│     │     │  └─ approval_handlers.go
│     │     └─ service
│     │        ├─ approval_service.go
│     │        └─ strategies
│     │           ├─ asset_strategy.go
│     │           ├─ document_strategy.go
│     │           ├─ group_strategy.go
│     │           ├─ policy_strategy.go
│     │           ├─ strategy.go
│     │           └─ user_strategy.go
│     ├─ assets
│     │  ├─ cmd
│     │  │  └─ main.go
│     │  ├─ Dockerfile
│     │  └─ internal
│     │     ├─ handlers
│     │     │  ├─ active_handlers.go
│     │     │  ├─ manager_handler.go
│     │     │  ├─ receiver_handler.go
│     │     │  └─ type_handlers.go
│     │     └─ service
│     │        ├─ active_service.go
│     │        ├─ asset_service.go
│     │        ├─ data
│     │        │  ├─ antivirus.go
│     │        │  ├─ data_transfer.go
│     │        │  ├─ firewall.go
│     │        │  ├─ helpers.go
│     │        │  ├─ inventory.go
│     │        │  ├─ port.go
│     │        │  ├─ software.go
│     │        │  └─ usb.go
│     │        ├─ data_service.go
│     │        ├─ lifecycle_service.go
│     │        └─ type_service.go
│     ├─ auth
│     │  ├─ cmd
│     │  │  └─ main.go
│     │  ├─ Dockerfile
│     │  └─ internal
│     │     ├─ handlers
│     │     │  └─ auth_handlers.go
│     │     └─ service
│     │        ├─ auth_service.go
│     │        └─ security.go
│     ├─ behavior
│     │  ├─ cmd
│     │  │  └─ main.go
│     │  ├─ Dockerfile
│     │  └─ internal
│     │     ├─ handlers
│     │     │  └─ behavior_handlers.go
│     │     └─ service
│     │        └─ behavior_service.go
│     ├─ dashboard
│     │  ├─ cmd
│     │  │  └─ main.go
│     │  ├─ Dockerfile
│     │  └─ internal
│     │     ├─ handlers
│     │     │  └─ dashboard_handlers.go
│     │     └─ service
│     │        └─ dashboard_controller.go
│     ├─ documents
│     │  ├─ cmd
│     │  │  └─ main.go
│     │  ├─ Dockerfile
│     │  └─ internal
│     │     ├─ handlers
│     │     │  └─ doc_handlers.go
│     │     └─ service
│     │        └─ document_service.go
│     ├─ groups
│     │  ├─ cmd
│     │  │  └─ main.go
│     │  ├─ Dockerfile
│     │  └─ internal
│     │     ├─ handlers
│     │     │  └─ group_handlers.go
│     │     └─ service
│     │        └─ group_service.go
│     ├─ incidents
│     │  ├─ cmd
│     │  │  └─ main.go
│     │  ├─ Dockerfile
│     │  └─ internal
│     │     ├─ handlers
│     │     │  └─ incident_handlers.go
│     │     └─ service
│     │        ├─ event_engine.go
│     │        └─ incident_service.go
│     ├─ policies
│     │  ├─ cmd
│     │  │  └─ main.go
│     │  ├─ Dockerfile
│     │  └─ internal
│     │     ├─ handlers
│     │     │  └─ policy_handlers.go
│     │     └─ service
│     │        └─ policy_service.go
│     ├─ scoring
│     │  ├─ cmd
│     │  │  └─ main.go
│     │  ├─ Dockerfile
│     │  └─ internal
│     │     ├─ handlers
│     │     │  └─ scoring_handlers.go
│     │     └─ service
│     │        └─ score_service.go
│     ├─ software
│     │  ├─ cmd
│     │  │  └─ main.go
│     │  ├─ Dockerfile
│     │  └─ internal
│     │     ├─ handlers
│     │     │  └─ software_handler.go
│     │     └─ service
│     │        └─ software_service.go
│     └─ users
│        ├─ cmd
│        │  └─ main.go
│        ├─ Dockerfile
│        └─ internal
│           ├─ handlers
│           │  └─ user_handlers.go
│           └─ service
│              └─ user_service.go
├─ SENT_frontend
│  ├─ .dockerignore
│  ├─ Dockerfile
│  ├─ index.html
│  ├─ public
│  ├─ src
│  │  ├─ api
│  │  │  └─ axios.jsx
│  │  ├─ App.jsx
│  │  ├─ components
│  │  │  ├─ AppDialog.jsx
│  │  │  ├─ common
│  │  │  │  ├─ BulkDeleteModal.jsx
│  │  │  │  └─ Pagination.jsx
│  │  │  ├─ GlobalCopilotDrawer.jsx
│  │  │  ├─ Navbar.jsx
│  │  │  ├─ Profile.jsx
│  │  │  ├─ SearchableList.jsx
│  │  │  └─ Sidebar.jsx
│  │  ├─ context
│  │  │  ├─ AuthContext.jsx
│  │  │  ├─ useSocketSubscription.js
│  │  │  └─ WebSocketContext.jsx
│  │  ├─ index.jsx
│  │  ├─ pages
│  │  │  ├─ AIChat
│  │  │  │  ├─ AIChatPage.jsx
│  │  │  │  ├─ components
│  │  │  │  │  ├─ ChatSidebar.jsx
│  │  │  │  │  ├─ ChatWindow.jsx
│  │  │  │  │  ├─ MessageBubble.jsx
│  │  │  │  │  └─ ThinkingBlock.jsx
│  │  │  │  └─ hooks
│  │  │  │     └─ useChat.js
│  │  │  ├─ approvals
│  │  │  │  ├─ ApprovalCenter.jsx
│  │  │  │  ├─ components
│  │  │  │  │  ├─ ApprovalList.jsx
│  │  │  │  │  └─ ApprovalTicketModal.jsx
│  │  │  │  └─ hooks
│  │  │  │     └─ useApprovals.js
│  │  │  ├─ Assets
│  │  │  │  ├─ Assets.jsx
│  │  │  │  ├─ AssetsDetail.jsx
│  │  │  │  ├─ AssetTypeManagement.jsx
│  │  │  │  ├─ components
│  │  │  │  │  ├─ AssetsActions.jsx
│  │  │  │  │  ├─ AssetsBulkActions.jsx
│  │  │  │  │  ├─ AssetsLogs.jsx
│  │  │  │  │  ├─ AssetsPort.jsx
│  │  │  │  │  ├─ AssetsSoftware.jsx
│  │  │  │  │  ├─ AssetsUSB.jsx
│  │  │  │  │  ├─ GenerateTokenButton.jsx
│  │  │  │  │  └─ TypeFormModal.jsx
│  │  │  │  └─ hooks
│  │  │  │     ├─ useAssets.js
│  │  │  │     ├─ useAssetsBulkActions.js
│  │  │  │     └─ useAssetTypes.js
│  │  │  ├─ Auth
│  │  │  │  ├─ Login.jsx
│  │  │  │  ├─ Register.jsx
│  │  │  │  └─ ResetPassword.jsx
│  │  │  ├─ Behaviors
│  │  │  │  ├─ BehaviorManager.jsx
│  │  │  │  ├─ components
│  │  │  │  │  ├─ BehaviorCard.jsx
│  │  │  │  │  ├─ BehaviorDetail.jsx
│  │  │  │  │  └─ BehaviorList.jsx
│  │  │  │  └─ hooks
│  │  │  │     └─ useBehaviors.js
│  │  │  ├─ Dashboard
│  │  │  │  ├─ components
│  │  │  │  │  ├─ DashboardCharts.jsx
│  │  │  │  │  └─ DashboardTable.jsx
│  │  │  │  ├─ Dashboard.jsx
│  │  │  │  └─ UserStats.jsx
│  │  │  ├─ Documents
│  │  │  │  ├─ Documents.jsx
│  │  │  │  └─ hooks
│  │  │  │     └─ useDocuments.js
│  │  │  ├─ Incident
│  │  │  │  ├─ components
│  │  │  │  │  └─ AuditChat.jsx
│  │  │  │  ├─ hooks
│  │  │  │  │  └─ useIncidents.js
│  │  │  │  ├─ IncidentDetail.jsx
│  │  │  │  └─ IncidentManager.jsx
│  │  │  ├─ policies
│  │  │  │  ├─ components
│  │  │  │  │  ├─ PolicyForm.jsx
│  │  │  │  │  ├─ PolicyList.jsx
│  │  │  │  │  ├─ PolicyOverview.jsx
│  │  │  │  │  └─ PolicyShared.jsx
│  │  │  │  ├─ hooks
│  │  │  │  │  └─ usePolicies.js
│  │  │  │  └─ PolicyCenter.jsx
│  │  │  ├─ Software
│  │  │  │  ├─ components
│  │  │  │  │  └─ SoftwareCard.jsx
│  │  │  │  ├─ hooks
│  │  │  │  │  └─ useSoftware.js
│  │  │  │  └─ SoftwarePage.jsx
│  │  │  └─ User
│  │  │     ├─ components
│  │  │     │  ├─ GroupManagementModal.jsx
│  │  │     │  ├─ UserDeleteModal.jsx
│  │  │     │  ├─ UserFormModal.jsx
│  │  │     │  ├─ UserTable.jsx
│  │  │     │  └─ UserToolbar.jsx
│  │  │     ├─ hooks
│  │  │     │  ├─ useGroups.js
│  │  │     │  ├─ UserFormModal.js
│  │  │     │  └─ useUsers.js
│  │  │     └─ UserManagement.jsx
│  │  ├─ styles
│  │  │  ├─ animations.css
│  │  │  ├─ auth.css
│  │  │  └─ index.css
│  │  └─ utils
│  │     └─ fileParsers.js
│  └─ vite.config.js
└─ SETUP.md

```