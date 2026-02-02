from pydantic import BaseModel
from typing import Optional

class UserLogin(BaseModel):
    username: str
    password: str
    captcha: str
    otp: Optional[str] = None

class AgentEnroll(BaseModel):
    token: str      # Enroll Token của vùng
    hwid: str       # Hardware ID
    hostname: str   # Tên máy
    # --- SCHEMAS CHO REGION ---
class RegionBase(BaseModel):
    name: str

class RegionCreate(RegionBase):
    pass

# Đây là Class mà hệ thống của bạn đang báo thiếu
class RegionResponse(RegionBase):
    id: int
    org_id: int
    enroll_token: str

    class Config:
        from_attributes = True # Giúp Pydantic hiểu được dữ liệu từ SQLAlchemy

# --- SCHEMAS CHO USER / AUTH ---
class UserLogin(BaseModel):
    username: str
    password: str
    captcha: str
    otp: Optional[str] = None

# --- SCHEMAS CHO AGENT ---
class AgentEnroll(BaseModel):
    token: str      # Enroll Token của vùng
    hwid: str       # Hardware ID từ máy trạm
    hostname: str   # Tên máy