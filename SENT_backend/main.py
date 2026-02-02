from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
import models, database
from api.v1 import auth, agents, regions
import models, database
models.Base.metadata.create_all(bind=database.engine)
# Tạo bảng trong DB (Nếu chưa có)
models.Base.metadata.create_all(bind=database.engine)

app = FastAPI(title="SENT Backend SaaS")

# QUAN TRỌNG: Cho phép Frontend React kết nối
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(auth.router, prefix="/api/v1/auth", tags=["Auth"])
app.include_router(agents.router, prefix="/api/v1/agents", tags=["Agents"])
app.include_router(regions.router, prefix="/api/v1/regions", tags=["Regions"])

@app.get("/")
def root():
    return {"message": "SENT API is running in SaaS mode"}