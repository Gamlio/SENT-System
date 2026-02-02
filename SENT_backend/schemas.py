from pydantic import BaseModel
from typing import Optional, Dict, Any
from datetime import datetime

# --- SCHEMAS CHO ROLE / RULES (Dành cho Level 2 & 4) ---
class RoleBase(BaseModel):
    name: str
    # Cho phép lưu trữ Rules linh hoạt: {"can_delete": false, "can_approve": true}
    permissions: Dict[str, Any] 
    level_context: int # Xác định Role này áp dụng cho Level 2 hay Level 4

class RoleCreate(RoleBase):
    # org_id có thể để trống (None) nếu là Role hệ thống của Level 1 tạo cho Level 2
    org_id: Optional[int] = None 

class RoleResponse(RoleBase):
    id: int
    org_id: Optional[int]

    class Config:
        from_attributes = True

# --- SCHEMAS CHO ORGANIZATION (Admin Level 1) --- [cite: 33]
class OrgBase(BaseModel):
    name: str

class OrgCreate(OrgBase):
    pass

class OrgResponse(OrgBase):
    id: int
    is_active: bool
    created_at: datetime

    class Config:
        from_attributes = True

# --- SCHEMAS CHO REGION ---
class RegionBase(BaseModel):
    name: str

class RegionCreate(RegionBase):
    pass

class RegionResponse(RegionBase):
    id: int
    org_id: int
    enroll_token: str

    class Config:
        from_attributes = True

# --- SCHEMAS CHO USER / AUTH --- [cite: 32]
class UserLogin(BaseModel):
    username: str
    password: str
    captcha: str
    otp: Optional[str] = None

# --- SCHEMAS CHO AGENT --- [cite: 34]
class AgentEnroll(BaseModel):
    token: str      
    hwid: str       
    hostname: str
    
class OrgCreateRequest(BaseModel):
    company_name: str
    username: str
    password: str