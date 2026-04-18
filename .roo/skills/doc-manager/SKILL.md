---
name: doc-manager
"description": "Skill chuyên biệt để cập nhật tài liệu và biểu đồ cho dự án SENT-System",
  "instructions": "1. Quét cấu trúc folder thực tế bằng lệnh ls hoặc docker exec. 2. Cập nhật các file .md (BACKEND, FRONTEND, DATABASE_GUIDE) dựa trên sự thay đổi của mã nguồn. 3. Sử dụng Mermaid.js để vẽ lại biểu đồ Sequence khi luồng API thay đổi. 4. Đảm bảo tuân thủ mô hình Hybrid Database (Postgres cho Core, Mongo cho Telemetry) khi mô tả dữ liệu.",
  "tools": [
    {
      "name": "update_structure",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-everything"]
    },
    {
      "name": "render_mermaid",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-mermaid"]
    }
  ]
}
---

# Doc Manager

## Instructions

Sử dụng skill này để quét sơ đồ thư mục và giải thích chức năng của từng thư mục/file trong SENT-System.

Cấu trúc Backend (Go): Phải giải thích rõ vai trò của internal/agent (thu thập log) và cmd/server (xử lý trung tâm).

Cấu trúc Frontend (React): Mô tả cách các components trong src/components kết nối với API backend qua cổng 8000.

Tự động cập nhật: Mỗi khi có file mới được tạo (ví dụ: module AI SENT mới), AI phải tự động cập nhật vào file README.md hoặc ARCHITECT.md., BACKEND_flow.md, FRONTEND_flow.md để đảm bảo tài liệu luôn đồng bộ với mã nguồn.

Luôn ưu tiên quét cấu trúc folder thực tế trước khi cập nhật tài liệu. Sử dụng ls -R hoặc Docker skill để kiểm tra.

Đối với Backend (Go): Phải phân tách rõ các lớp Handler (API v1), Service (Business Logic) và Models (SQL vs NoSQL).

Đối với Frontend (React): Phải liệt kê đúng các component trong src/pages và mapping chính xác với các endpoint API tại cổng 8000.

Đối với Agent (Ninja Sensor): Phải mô tả đúng các module collector, network, và diff (thuật toán Differential Reporting).

Khi thấy file .go mới trong internal/service/assets/data/, hãy tự động cập nhật mục "Chi Tiết Các Loại Dữ Liệu Giám Sát" trong BACKEND.md.

cấu trúc thư mục phải được giải thích rõ ràng để người mới có thể hiểu ngay chức năng của từng phần mà không cần đọc code chi tiết:
SENT-System/
├── SENT_backend/ (Go + Gin)
│   ├── cmd/server/main.go          # Điểm khởi chạy, cấu hình DB & Router
│   ├── internal/
│   │   ├── api/v1/                 # HANDLERS (Tiếp nhận request)
│   │   │   ├── assets/             # Enrollment, Receiver, Dashboard handlers
│   │   │   ├── incidents/          # Xử lý sự cố và AI Analysis
│   │   │   └── policies/           # Quản lý luật an ninh và tài liệu AI
│   │   ├── service/                # SERVICE LAYER (Logic nghiệp vụ)
│   │   │   ├── asset_data/         # Xử lý Software, USB, Network, Port
│   │   │   ├── security/           # Event Engine, Alerts, Compliance
│   │   │   └── scoring/            # Risk Scoring Logic (0-100)
│   │   ├── models/                 # DATA MODELS
│   │   │   ├── models.go           # PostgreSQL (Core/Relational)
│   │   │   └── mongo_models.go     # MongoDB (Telemetry/Big Data)
│   │   └── database/db.go          # Khởi tạo Hybrid DB (Postgres + Mongo)
│   └── Dockerfile
├── SENT_frontend/ (React + Vite)
│   ├── src/
│   │   ├── api/axios.jsx           # Cấu hình kết nối API port 8000
│   │   ├── context/                # AuthContext & WebSocketContext
│   │   ├── pages/                  # Modules: Assets, Incidents, Policies, AI Chat
│   │   └── components/             # Sidebar, Navbar, GlobalCopilotDrawer
│   └── .env (VITE_API_URL)
├── SENT_agent/ (Ninja Sensor - Go)
│   ├── internal/
│   │   ├── collector/              # Module: Software, USB, System, Telemetry
│   │   ├── network/                # Module: Client & Differential Reporting(Diff)
│   │   └── config/                 # Quản lý HWID và Company Code
│   └── main.go
└── docker-compose.yml (Hạ tầng DB & Containers)