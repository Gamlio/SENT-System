import os
from dotenv import load_dotenv
from langchain_google_genai import GoogleGenerativeAIEmbeddings, ChatGoogleGenerativeAI
from langchain_postgres import PGVector 
from langchain_core.tools import tool
from langchain_core.messages import HumanMessage
from langgraph.prebuilt import create_react_agent 
from core_logic.web_search import web_search_tool

load_dotenv()

class PolicyRAGAgent:
    def __init__(self, db_url=None):
        self.google_api_key = os.getenv("GOOGLE_API_KEY")
        self.connection_string = db_url or os.getenv("DB_URL")
        
        # 1. Khởi tạo LLM & Embeddings
        self.llm = ChatGoogleGenerativeAI(model="gemini-2.5-flash", google_api_key=self.google_api_key)
        self.embeddings = GoogleGenerativeAIEmbeddings(model="models/embedding-001", google_api_key=self.google_api_key)

        # 2. Cấu hình Vector Store từ PostgreSQL
        self.vector_store = PGVector(
            embeddings=self.embeddings,
            collection_name="policy_vectors",
            connection=self.connection_string,
            use_jsonb=True, 
        )
        self.retriever = self.vector_store.as_retriever(search_kwargs={"k": 3})

        # 3. Khai báo Tools
        @tool
        def local_policy_tool(query: str) -> str:
            """Sử dụng công cụ này ĐẦU TIÊN để tìm kiếm các quy định nội bộ trong PostgreSQL."""
            try:
                docs = self.retriever.invoke(query)
                return "\n\n".join([doc.page_content for doc in docs]) if docs else "Không tìm thấy dữ liệu nội bộ."
            except Exception:
                return "Lỗi kết nối cơ sở dữ liệu PostgreSQL."

        self.tools = [web_search_tool, local_policy_tool]
        
        # LƯU Ý: Không khởi tạo self.agent_executor ở đây nữa, 
        # vì chúng ta cần nó linh hoạt tạo mới theo từng User/Admin ở hàm chat.

    def chat(self, user_query: str, user_level: int = 1, org_id: int = 1):
        """Hàm chính để giao tiếp với AI, có tích hợp phân quyền Role-Based"""
        try:
            # 1. Tùy biến System Prompt theo RoleLevel của người dùng
            if user_level >= 2:
                system_msg = f"""Bạn là Trợ lý An ninh mạng cấp cao dành cho ADMIN (Tổ chức ID: {org_id}). 
                Bạn có quyền truy cập toàn bộ luật lệ, cấu hình hệ thống. 
                Hãy tư vấn chuyên sâu, phân tích rủi ro kỹ thuật và có thể đề xuất sửa đổi chính sách."""
            else:
                system_msg = f"""Bạn là Trợ lý hỗ trợ IT cho Nhân viên (Tổ chức ID: {org_id}). 
                Chỉ trả lời dựa trên nội quy công ty và hướng dẫn an toàn cơ bản. 
                Luôn trả lời NGẮN GỌN, súc tích, dễ hiểu. 
                TUYỆT ĐỐI KHÔNG tiết lộ thông tin cấu hình bảo mật nhạy cảm hoặc kỹ thuật chuyên sâu."""

            # 2. Khởi tạo Agent động "bọc" đúng Prompt của người hỏi
            agent_executor = create_react_agent(
                self.llm, 
                tools=self.tools, 
                state_modifier=system_msg
            )

            # 3. Gửi câu hỏi cho Agent xử lý
            inputs = {"messages": [HumanMessage(content=user_query)]}
            response = agent_executor.invoke(inputs)
            return response["messages"][-1].content
            
        except Exception as e:
            return f"Lỗi trong quá trình xử lý: {str(e)}"