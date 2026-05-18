```mermaid
%%{init: {
  'theme': 'base',
  'themeVariables': {
    'background': '#FFFFFF',
    'actorBkg': '#F8FAFC',
    'actorBorder': '#475569',
    'actorTextColor': '#1E293B',
    'actorLineColor': '#94A3B8',
    'signalColor': '#334155',
    'signalLineColor': '#64748B',
    'labelBoxBkgColor': '#F1F5F9',
    'labelBoxBorderColor': '#CBD5E1',
    'labelTextColor': '#1E293B',
    'activationBkgColor': '#4F46E5',
    'activationBorderColor': '#4F46E5'
  }
}}%%
sequenceDiagram
    actor Analyst as 👤 Chuyên viên SOC
    participant FE as 🖥️ Giao diện (Frontend)
    participant BE as ⚙️ Hệ thống trung tâm (Backend)
    participant AI as 🤖 Trợ lý AI (Local LLM)

    Analyst->>FE: Mở khung tương tác Trợ lý AI trên giao diện
    activate FE
    FE-->>Analyst: Trình diện giao diện chat sẵn sàng nhận truy vấn
    deactivate FE
    
    Analyst->>FE: Nhập yêu cầu phân tích sự cố / cách xử lý vi phạm & nhấn gửi
    activate FE
    FE->>BE: Gửi nội dung câu hỏi truy vấn của người dùng
    activate BE
    
    BE->>BE: 1. Thu thập dữ liệu bối cảnh thực tế của sự cố hiện tại
    BE->>BE: 2. Tra cứu quy trình xử lý chuẩn (Playbook) & chính sách liên quan
    
    BE->>AI: Nạp yêu cầu kèm theo dữ liệu bối cảnh và quy trình nghiệp vụ đã trích xuất
    activate AI
    
    AI->>AI: Thực hiện phân tích logic từng bước dựa trên tài liệu thực tế (Thought)
    
    AI-->>BE: Trả về nội dung phản hồi (Tóm tắt vụ việc - Đánh giá rủi ro - Đề xuất hướng xử lý)
    deactivate AI
    
    BE-->>FE: Phản hồi luồng suy luận và kết quả tư vấn xử lý sự cố
    deactivate BE
    
    FE-->>Analyst: Hiển thị nội dung hướng dẫn trực quan theo cấu trúc kịch bản phản ứng chuẩn
    deactivate FE