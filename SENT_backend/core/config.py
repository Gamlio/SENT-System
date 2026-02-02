import os
from dotenv import load_dotenv

load_dotenv()

class Settings:
    def __init__(self):
        self.PROJECT_NAME = "SENT SaaS"
        # Đảm bảo DATABASE_URL khớp với .env 
        self.DATABASE_URL = os.getenv("DATABASE_URL")
        # ĐỔI TÊN Ở ĐÂY để khớp với .env
        self.SECRET_KEY = os.getenv("SECRET_KEY", "SENT_SUPER_SECRET_KEY_2026")
        self.ALGORITHM = os.getenv("ALGORITHM", "HS256")
        
        raw_origins = os.getenv("ALLOWED_ORIGINS", "http://localhost:3000")
        self.ALLOWED_ORIGINS = [o.strip() for o in raw_origins.split(",") if o.strip()]

settings = Settings()