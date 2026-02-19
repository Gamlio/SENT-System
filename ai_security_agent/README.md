# AI Security Agent: Intelligent Log Analysis & Compliance System

Hệ thống AI Security Agent được thiết kế để tự động hóa quá trình giám sát an toàn thông tin (SOC). Hệ thống kết hợp khả năng phân tích log chuyên sâu và truy xuất các chính sách bảo mật để đưa ra quyết định xử lý sự cố có kèm theo căn cứ pháp lý.

## 🏗 Kiến trúc hệ thống (Multi-Agent Architecture)

Dự án được chia thành 3 phân hệ chính:

1. **Agent 1 (Log Analyzer):** - Nhiệm vụ: Đọc, phân tích log hệ thống (Firewall, Server, Network) và nhận diện các dấu hiệu bất thường (Anomalies/Attacks).
   - Công nghệ: Fine-tuning (QLoRA) các mô hình LLM mã nguồn mở dựa trên dataset log bảo mật.
2. **Agent 2 (Policy & Compliance RAG):** - Nhiệm vụ: Trả lời các câu hỏi về chính sách, quy định pháp lý an toàn thông tin để đối chiếu với hành vi vi phạm.
   - Công nghệ: Retrieval-Augmented Generation (RAG) với Vector Database.
3. **Orchestrator:**
   - Nhiệm vụ: Tiếp nhận dữ liệu đầu vào, điều phối luồng giao tiếp giữa Agent 1 và Agent 2, sau đó tổng hợp kết quả cuối cùng.
   - Công nghệ: LangChain / CrewAI.

## 🚀 Hướng dẫn cài đặt (Setup Instructions)

**Bước 1: Clone dự án và thiết lập môi trường**
Tạo môi trường ảo (Virtual Environment) để tránh xung đột thư viện:
`python -m venv venv`
`source venv/bin/activate`  # Trên Windows dùng: venv\Scripts\activate

**Bước 2: Cài đặt thư viện phụ thuộc**
Di chuyển vào thư mục `orchestrator` và cài đặt các packages cần thiết:
`pip install -r orchestrator/requirements.txt`

## 🛠 Hướng dẫn chạy dự án từng bước

**Phần 1: Khởi tạo Agent Policy RAG**
1. Đặt các file tài liệu chính sách (PDF, TXT) vào thư mục `agent_policy_rag/documents/`.
2. Chạy script để vector hóa dữ liệu:
   `python agent_policy_rag/ingest_data.py`

**Phần 2: Huấn luyện Agent Log Analyzer (Tùy chọn)**
1. Chuẩn bị dữ liệu JSONL trong `agent_log_analyzer/data/`.
2. Mở file `notebooks/finetune_log_model.ipynb` bằng Jupyter Notebook hoặc Google Colab để tiến hành train mô hình.

**Phần 3: Chạy hệ thống điều phối (Main Run)**
1. Mở file `orchestrator/main_agent.ipynb`.
2. Chạy các cell để mô phỏng luồng đưa một đoạn log bất thường vào, cho 2 Agent phối hợp phân tích và đưa ra kết luận.
ai_security_agent/
│
├── agent_log_analyzer/         # Thư mục chứa AI xử lý và phân tích Log
│   ├── data/                   # Chứa file log thô và file JSONL để train
│   ├── models/                 # Chứa model weights sau khi fine-tune (LoRA adapters)
│   ├── notebooks/              # Chứa file .ipynb để train (QLoRA)
│   └── inference_log.py        # Script chạy model để dự đoán log mới
│
├── agent_policy_rag/           # (ĐÃ CẬP NHẬT) AI đọc hiểu Policy & Tìm kiếm Web
│   ├── documents/              # Chứa các file PDF, Word chính sách bảo mật nội bộ
│   ├── vector_db/              # Thư mục lưu database vector (ChromaDB/FAISS)
│   ├── core_logic/             # Chứa logic Python (Dùng cho Web Backend sau này)
│   │   ├── document_loader.py  # Hàm đọc file, cắt chunk và đưa vào Vector DB
│   │   ├── web_search.py       # Hàm tích hợp công cụ tìm kiếm mạng (Tavily/DuckDuckGo)
│   │   └── rag_engine.py       # Chứa Agent cấu hình Tools (chọn Web hoặc Local DB)
│   └── test_chat_rag.ipynb     # File Notebook dùng để chat test trực tiếp với Agent RAG
│
├── orchestrator/               # Hệ thống điều phối (Multi-Agent) ghép nối 2 AI
│   ├── main_agent.ipynb        # File Notebook để chạy demo ghép nối toàn hệ thống
│   ├── app.py                  # (Tùy chọn) Giao diện web test nhanh (Streamlit/Gradio)
│   └── requirements.txt        # Danh sách thư viện cần cài đặt cho toàn dự án
│
└── README.md                   # File hướng dẫn tổng quan