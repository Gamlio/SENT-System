from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from core import security
import models, schemas, database
import os
from dotenv import load_dotenv

load_dotenv()

class Settings:
    PROJECT_NAME: str = "SENT SaaS"
    DATABASE_URL: str = os.getenv("DATABASE_URL")
    SECRET_KEY: str = os.getenv("SECRET_KEY")
    ALGORITHM: str = os.getenv("ALGORITHM")

settings = Settings()
router = APIRouter()

@router.post("/login")
def login(form_data: schemas.UserLogin, db: Session = Depends(database.get_db)):
    user = db.query(models.User).filter(models.User.username == form_data.username).first()
    if not user or not security.verify_password(form_data.password, user.hashed_password):
        raise HTTPException(status_code=400, detail="Sai tài khoản hoặc mật khẩu")
    
    # Tạo token chứa thông tin định danh
    token = security.create_access_token(data={
        "sub": user.username, 
        "level": user.level, 
        "org_id": user.org_id
    })
    return {"access_token": token, "token_type": "bearer", "level": user.level}