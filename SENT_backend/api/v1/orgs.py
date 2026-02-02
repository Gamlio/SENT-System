from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from database import get_db
import models, schemas
from core.security import get_current_user # Bạn sẽ viết hàm này để check Level

router = APIRouter()

# 1. Lấy toàn bộ danh sách SME
@router.get("/all")
def get_all_orgs(db: Session = Depends(get_db), current_user: models.User = Depends(get_current_user)):
    if current_user.level != 1:
        raise HTTPException(status_code=403, detail="Chỉ Level 1 mới có quyền này")
    return db.query(models.Organization).all()
@router.get("/all", response_model=list[schemas.OrgResponse])
def get_all_organizations(
    db: Session = Depends(get_db), 
    current_user: models.User = Depends(get_current_user)
):
    # RÀO CHẮN BẢO MẬT: Chỉ Level 1 mới được thấy tất cả SME
    if current_user.level != 1:
        raise HTTPException(status_code=403, detail="Bạn không có quyền truy cập dữ liệu này")
    
    return db.query(models.Organization).all()

# 2. Tạo mới một Organization (SME)
@router.post("/create")
def create_org(org_data: schemas.OrgCreate, db: Session = Depends(get_db), current_user: models.User = Depends(get_current_user)):
    if current_user.level != 1:
        raise HTTPException(status_code=403, detail="Bạn không phải Admin hệ thống")
    
    new_org = models.Organization(name=org_data.name)
    db.add(new_org)
    db.commit()
    db.refresh(new_org)
    return {"message": "Tạo SME thành công", "org": new_org}

# 3. Quản lý Rules cho Level 2
@router.post("/roles/level2")
def create_system_role(role_data: schemas.RoleCreate, db: Session = Depends(get_db), current_user: models.User = Depends(get_current_user)):
    if current_user.level != 1:
        raise HTTPException(status_code=403, detail="Quyền hạn không đủ")
    # Tạo role cho nhân viên hệ thống (Level 2)
    new_role = models.Role(name=role_data.name, permissions=role_data.permissions, org_id=None)
    db.add(new_role)
    db.commit()
    return new_role