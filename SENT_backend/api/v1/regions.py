from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
import models, schemas, database

router = APIRouter()

@router.post("/", response_model=schemas.RegionResponse)
def create_new_region(region: schemas.RegionCreate, db: Session = Depends(database.get_db)):
    # Bước 1: Kiểm tra xem tên vùng đã tồn tại chưa
    db_region = db.query(models.Region).filter(models.Region.name == region.name).first()
    if db_region:
        raise HTTPException(status_code=400, detail="Vùng này đã tồn tại")
    
    # Bước 2: Tạo vùng mới (API Key sẽ tự động sinh bởi Model)
    new_region = models.Region(
        name=region.name, 
        description=region.description
    )
    db.add(new_region)
    db.commit()
    db.refresh(new_region)
    
    # Bước 3: Trả về API Key để người dùng dán vào file .yml của Agent Go
    return new_region