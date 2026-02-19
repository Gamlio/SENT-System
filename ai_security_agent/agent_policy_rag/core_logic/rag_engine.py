import os
from dotenv import load_dotenv # <-- Thêm thư viện này
from langchain_community.tools import DuckDuckGoSearchRun
from langchain.tools.retriever import create_retriever_tool
from langchain_community.vectorstores import Chroma
from langchain_openai import OpenAIEmbeddings, ChatOpenAI
from langchain.agents import AgentExecutor, create_tool_calling_agent
from langchain_core.prompts import ChatPromptTemplate
from langchain.tools import Tool

# Load các biến môi trường từ file .env
load_dotenv() 

class PolicyRAGAgent:
    def __init__(self, vector_db_path="../vector_db"):
        # Lấy API key trực tiếp từ biến môi trường đã được load
        openai_api_key = os.getenv("OPENAI_API_KEY")
        if not openai_api_key:
            raise ValueError("[LỖI] Không tìm thấy OPENAI_API_KEY trong file .env")
        
        # 1. Khởi tạo LLM làm "bộ não" cho Agent
        self.llm = ChatOpenAI(model="gpt-3.5-turbo", temperature=0, api_key=openai_api_key)

        # 2. Tool 1: Tìm kiếm Web
        self.web_search = DuckDuckGoSearchRun()
        self.web_tool = Tool(
            name="WebSearch",
            description="Hữu ích khi cần tìm kiếm các chính sách, luật pháp an ninh mạng quốc gia hoặc thông tin public trên internet.",
            func=self.web_search.run
        )

        self.tools = [self.web_tool]

        # 3. Tool 2: Tìm kiếm tài liệu nội bộ (Policy của công ty)
        if os.path.exists(vector_db_path):
            try:
                embeddings = OpenAIEmbeddings(api_key=openai_api_key)
                self.vectorstore = Chroma(persist_directory=vector_db_path, embedding_function=embeddings)
                self.retriever = self.vectorstore.as_retriever(search_kwargs={"k": 3})
                self.doc_tool = create_retriever_tool(
                    self.retriever,
                    "LocalPolicySearch",
                    "Sử dụng công cụ này ĐẦU TIÊN để tìm kiếm các quy định, chính sách bảo mật nội bộ trong cơ sở dữ liệu."
                )
                self.tools.append(self.doc_tool)
                print("[INFO] Đã tải thành công Local Policy Tool.")
            except Exception as e:
                print(f"[WARN] Lỗi tải Vector DB: {e}. Hệ thống sẽ chỉ dùng Web Search.")
        else:
            print("[WARN] Chưa có Vector DB nội bộ. Hệ thống sẽ chỉ dùng Web Search.")

        # 4. Viết Prompt chỉ đạo Agent
        prompt = ChatPromptTemplate.from_messages([
            ("system", "Bạn là một AI Security Agent chuyên nghiệp. "
                       "Nhiệm vụ của bạn là phân tích quy định và chính sách an toàn thông tin. "
                       "Nếu người dùng hỏi về quy định nội bộ, hãy dùng LocalPolicySearch. "
                       "Nếu hỏi về luật pháp quốc gia hoặc không tìm thấy ở nội bộ, hãy dùng WebSearch."),
            ("human", "{input}"),
            ("placeholder", "{agent_scratchpad}"),
        ])

        # 5. Khởi tạo Agent và Executor
        self.agent = create_tool_calling_agent(self.llm, self.tools, prompt)
        self.agent_executor = AgentExecutor(agent=self.agent, tools=self.tools, verbose=True)

    def chat(self, user_query):
        """Hàm chính để giao tiếp với AI"""
        try:
            response = self.agent_executor.invoke({"input": user_query})
            return response["output"]
        except Exception as e:
            return f"Lỗi trong quá trình xử lý: {str(e)}"