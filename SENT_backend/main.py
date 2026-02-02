from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import models, database
from core.config import settings
from api.v1 import auth,orgs, agents, regions
import models, database
models.Base.metadata.create_all(bind=database.engine)
# Tạo bảng trong DB (Nếu chưa có)
models.Base.metadata.create_all(bind=database.engine)

app = FastAPI()

# 1. Cấu hình CORS PHẢI đặt ở đây (trước các router) [cite: 13, 23]
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.ALLOWED_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 2. Sau đó mới đến Router [cite: 31]
app.include_router(auth.router, prefix="/api/v1/auth", tags=["Authentication"])
app.include_router(orgs.router, prefix="/api/v1/orgs", tags=["Organizations"])
app.include_router(agents.router, prefix="/api/v1/agents", tags=["Agents"])
app.include_router(regions.router, prefix="/api/v1/regions", tags=["Regions"])

@app.get("/")
def root():
    return {"message": "SENT API is running in SaaS mode"}