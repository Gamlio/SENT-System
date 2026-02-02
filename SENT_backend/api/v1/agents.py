from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
import models, schemas, database

router = APIRouter()

@router.post("/enroll")
async def enroll_agent(request: schemas.AgentEnroll, db: Session = Depends(database.get_db)):
    # 1. Tìm Vùng dựa trên Token người dùng nhập vào Agent
    region = db.query(models.Region).filter(models.Region.enroll_token == request.token).first()
    if not region:
        raise HTTPException(status_code=403, detail="Mã vùng không hợp lệ")

    # 2. Kiểm tra xem máy đã tồn tại chưa (qua HWID)
    agent = db.query(models.Agent).filter(models.Agent.hwid == request.hwid).first()
    if not agent:
        # Tạo bản ghi mới ở trạng thái chờ duyệt
        agent = models.Agent(
            hwid=request.hwid,
            hostname=request.hostname,
            region_id=region.id,
            org_id=region.org_id,
            status='pending' 
        )
        db.add(agent)
        db.commit()
    
    return {"status": agent.status, "message": "Yêu cầu đã được gửi, vui lòng chờ Admin phê duyệt."}