import os
import psycopg2
from dotenv import load_dotenv
from langchain_community.document_loaders import PyPDFLoader
from langchain_text_splitters import RecursiveCharacterTextSplitter
from langchain_google_genai import GoogleGenerativeAIEmbeddings
from langchain_postgres import PGVector
from langchain_core.documents import Document

load_dotenv()

class RAGDocumentLoader:
    def __init__(self, db_url=None):
        self.google_api_key = os.getenv("GOOGLE_API_KEY")
        self.db_url = db_url or os.getenv("DB_URL") # Chuỗi kết nối Postgres của bạn
        
        # Cấu hình nhúng Vector bằng Gemini
        self.embeddings = GoogleGenerativeAIEmbeddings(
            model="models/embedding-001", 
            google_api_key=self.google_api_key
        )
        
        # Kết nối tới bảng Vector (Tự tạo nếu chưa có)
        self.vector_store = PGVector(
            embeddings=self.embeddings,
            collection_name="policy_vectors",
            connection=self.db_url,
            use_jsonb=True,
        )

    def process_new_documents(self):
        """Quét DB tìm tài liệu mới, đọc PDF và chuyển thành Vector"""
        print("[*] Đang kiểm tra tài liệu mới từ Admin Website...")
        
        # 1. Kết nối trực tiếp vào Postgres bằng psycopg2 để lấy file vật lý
        try:
            conn = psycopg2.connect(self.db_url)
            cur = conn.cursor()
            
            # Chỉ lấy những file chưa được AI xử lý
            cur.execute("SELECT id, title, file_path, category FROM policy_documents WHERE is_processed = false")
            new_docs = cur.fetchall()
            
            if not new_docs:
                print("[-] Không có tài liệu nào mới cần học.")
                cur.close()
                conn.close()
                return

            print(f"[+] Tìm thấy {len(new_docs)} tài liệu mới! Bắt đầu nạp não...")
            
            text_splitter = RecursiveCharacterTextSplitter(chunk_size=1000, chunk_overlap=200)

            for doc_id, title, file_path, category in new_docs:
                # Chỉnh lại đường dẫn file nếu Go và Python chạy ở 2 thư mục khác nhau
                # Giả sử thư mục uploads nằm ở chung thư mục gốc (Tùy cấu trúc máy bạn)
                real_file_path = f"../SENT_backend/{file_path}" 
                
                print(f"  -> Đang đọc: {title} ({real_file_path})")
                
                if not os.path.exists(real_file_path):
                    print(f"  [LỖI] Không tìm thấy file vật lý: {real_file_path}")
                    continue

                # 2. Đọc file PDF
                loader = PyPDFLoader(real_file_path)
                pages = loader.load()
                
                # Cắt nhỏ văn bản
                chunks = text_splitter.split_documents(pages)
                
                # Ép thêm Metadata để AI dễ tìm sau này
                for chunk in chunks:
                    chunk.metadata["title"] = title
                    chunk.metadata["category"] = category
                    chunk.metadata["doc_id"] = doc_id

                # 3. Đẩy lên Vector Database
                self.vector_store.add_documents(chunks)
                
                # 4. Đánh dấu tài liệu này ĐÃ ĐƯỢC HỌC
                cur.execute("UPDATE policy_documents SET is_processed = true WHERE id = %s", (doc_id,))
                conn.commit()
                print(f"  ✅ Đã học xong: {title}")

            cur.close()
            conn.close()
            print("[*] Hoàn tất quá trình nạp tri thức!")
            
        except Exception as e:
            print(f"[LỖI CRITICAL] Lỗi kết nối hoặc xử lý: {e}")

# Chạy file này độc lập để test
if __name__ == "__main__":
    # Đảm bảo file .env của bạn có: DB_URL=postgresql://user:password@localhost:5432/dbname
    loader = RAGDocumentLoader()
    loader.process_new_documents()