import os
from dotenv import load_dotenv
from langchain_community.tools import DuckDuckGoSearchRun
from langchain_chroma import Chroma 
from langchain_google_genai import GoogleGenerativeAIEmbeddings, ChatGoogleGenerativeAI
from langchain_core.tools import tool
from langchain_core.messages import HumanMessage # <--- Bổ sung định dạng tin nhắn chuẩn
from langgraph.prebuilt import create_react_agent

from ai_security_agent.agent_policy_rag.core_logic.web_search import web_search_tool 

load_dotenv()

class PolicyRAGAgent:
    def __init__(self, vector_db_path="../vector_db"):
        google_api_key = os.getenv("GOOGLE_API_KEY")
        if not google_api_key:
            raise ValueError("[LỖI] Không tìm thấy GOOGLE_API_KEY trong file .env")
        
        # 1. Khởi tạo LLM 
        self.llm = ChatGoogleGenerativeAI(model="gemini-2.5-flash", temperature=0, google_api_key=google_api_key)

        # 2. Tool 1: Tìm kiếm web (Luôn có sẵn)
        self.tools = [web_search_tool]

        # 3. Tool 2: Tìm kiếm tài liệu nội bộ
        if os.path.exists(vector_db_path):
            try:
                embeddings = GoogleGenerativeAIEmbeddings(model="models/embedding-001", google_api_key=google_api_key)
                self.vectorstore = Chroma(persist_directory=vector_db_path, embedding_function=embeddings)
                self.retriever = self.vectorstore.as_retriever(search_kwargs={"k": 3})
                
                @tool
                def local_policy_tool(query: str) -> str:
                    """Sử dụng công cụ này ĐẦU TIÊN để tìm kiếm các quy định, chính sách bảo mật nội bộ trong cơ sở dữ liệu."""
                    docs = self.retriever.invoke(query)
                    if not docs:
                        return "Không có dữ liệu trong database nội bộ."
                    return "\n\n".join([doc.page_content for doc in docs])

                self.tools.append(local_policy_tool)
            except Exception as e:
                pass

        # 4. Khởi tạo Agent với System Prompt điều khiển độ dài
        system_prompt = """Bạn là một Chuyên gia An toàn thông tin cấp cao.
        Khi trả lời người dùng, hãy tuân thủ quy tắc sau:
        - Luôn trả lời NGẮN GỌN, SÚC TÍCH, đi thẳng vào vấn đề.
        - Giới hạn câu trả lời tối đa trong vòng 3-4 câu hoặc 1 đoạn văn ngắn (trừ khi người dùng yêu cầu chi tiết).
        - Nếu có danh sách, chỉ liệt kê tối đa 3 ý quan trọng nhất.
        """
        
        self.agent_executor = create_react_agent(
            self.llm, 
            tools=self.tools, 
            state_modifier=system_prompt # <--- Nơi gắn luật cho AI
        )

    def chat(self, user_query):
        """Hàm chính để giao tiếp với AI"""
        try:
            # Gói câu hỏi vào HumanMessage để tránh lỗi API
            inputs = {"messages": [HumanMessage(content=user_query)]}
            response = self.agent_executor.invoke(inputs)
            return response["messages"][-1].content
        except Exception as e:
            return f"Lỗi trong quá trình xử lý: {str(e)}"