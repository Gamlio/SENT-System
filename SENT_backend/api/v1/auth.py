from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
import schemas
import models, schemas, database 
from core import security
router = APIRouter()

@router.post("/login")
def login(form_data: schemas.UserLogin, db: Session = Depends(database.get_db)):
    # Truy vấn người dùng từ database sent_db [cite: 24, 25]
    user = db.query(models.User).filter(models.User.username == form_data.username).first()
    
    if not user or not security.verify_password(form_data.password, user.hashed_password):
        raise HTTPException(status_code=400, detail="Sai tài khoản hoặc mật khẩu")
    
    # Tạo token chứa thông tin định danh cho 4 cấp độ quyền hạn [cite: 8]
    token = security.create_access_token(data={
        "sub": user.username, 
        "level": user.level, 
        "org_id": user.org_id # Cô lập dữ liệu theo công ty [cite: 6, 7]
    })
    
    return {
        "access_token": token, 
        "token_type": "bearer", 
        "level": user.level,
        "org_id": user.org_id
    }
@router.post("/register")
def register_sme(form_data: schemas.OrgCreateRequest, db: Session = Depends(database.get_db)):
    # 1. Kiểm tra xem username đã tồn tại chưa (Security check)
    existing_user = db.query(models.User).filter(models.User.username == form_data.username).first()
    if existing_user:
        raise HTTPException(status_code=400, detail="Tài khoản đã tồn tại trong hệ thống SENT")

    # 2. Tạo Organization mới (SME)
    new_org = models.Organization(name=form_data.company_name)
    db.add(new_org)
    db.flush() # Lấy ID của Org mới tạo

    # 3. Tạo Admin cho SME đó (Level 3)
    hashed_pass = security.get_password_hash(form_data.password)
    new_user = models.User(
        username=form_data.username,
        hashed_password=hashed_pass,
        level=3, # Mặc định là Admin của SME
        org_id=new_org.id
    )
    db.add(new_user)
    db.commit()
    
    return {"message": "Đăng ký thành công", "org_id": new_org.id}