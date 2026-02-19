from langchain_community.tools import DuckDuckGoSearchRun
from langchain_core.tools import tool

# Khởi tạo công cụ tìm kiếm
ddg_search = DuckDuckGoSearchRun()

@tool
def web_search_tool(query: str) -> str:
    """Hữu ích khi cần tìm kiếm các chính sách, luật pháp an ninh mạng quốc gia hoặc thông tin public trên internet."""
    try:
        # Ép kiểu và kiểm tra nếu kết quả rỗng
        result = ddg_search.run(query)
        if not result or str(result).isspace():
            return "Không tìm thấy thông tin trên web, hãy trả lời theo kiến thức của bạn."
        return str(result)
    except Exception as e:
        # Nếu DuckDuckGo lỗi, trả về text để tránh sập hệ thống API
        return f"Lỗi truy xuất web: {str(e)}"