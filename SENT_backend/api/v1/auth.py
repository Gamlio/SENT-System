from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from core import security
import models, schemas, database

router = APIRouter()

@router.post("/register-sme")
def register_sme(org_name: str, username: str, password: str, db: Session = Depends(database.get_db)):
    # 1. Tạo công ty mới
    new_org = models.Organization(name=org_name)
    db.add(new_org)
    db.commit()
    db.refresh(new_org)

    # 2. Tạo User Admin cho công ty đó (Level 2)
    new_user = models.User(
        username=username,
        hashed_password=security.hash_password(password),
        level=2,
        org_id=new_org.id
    )
    db.add(new_user)
    db.commit()
    return {"message": "Đăng ký công ty thành công!"}

@router.post("/login")
def login(data: schemas.UserLogin, db: Session = Depends(database.get_db)):
    user = db.query(models.User).filter(models.User.username == data.username).first()
    if not user or not security.verify_password(data.password, user.hashed_password):
        raise HTTPException(status_code=400, detail="Sai tài khoản hoặc mật khẩu")
    if data.otp != "000000":
        raise HTTPException(status_code=400, detail="Mã OTP không chính xác")
    # Ở đây bạn có thể thêm logic kiểm tra 2FA và Captcha thực tế
    token = security.create_access_token(data={
        "sub": user.username, 
        "level": user.level, 
        "org_id": user.org_id
    })
    
    return {
        "access_token": token, 
        "token_type": "bearer", 
        "level": user.level,
        "username": user.username
    }