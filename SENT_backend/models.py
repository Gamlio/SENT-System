import uuid
import datetime
from sqlalchemy import Column, Integer, String, ForeignKey, JSON, DateTime, Boolean
from database import Base

# models.py cập nhật
class Role(Base):
    __tablename__ = "roles"
    id = Column(Integer, primary_key=True)
    name = Column(String) # Ví dụ: "Kỹ thuật viên", "Người xem log"
    # Quyền hạn lưu dạng JSON: {"can_approve_agent": true, "view_all_regions": false}
    permissions = Column(JSON) 
    level_context = Column(Integer) # Role này dành cho Level 2 hay Level 4
    org_id = Column(Integer, ForeignKey("organizations.id"), nullable=True) # [cite: 6]
# 1. Công ty khách hàng (SME)
class Organization(Base):
    __tablename__ = 'organizations'
    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, unique=True)
    created_at = Column(DateTime, default=datetime.datetime.utcnow)

# 2. Vùng quản lý trong công ty
class Region(Base):
    __tablename__ = 'regions'
    id = Column(Integer, primary_key=True, index=True)
    org_id = Column(Integer, ForeignKey('organizations.id'))
    name = Column(String)
    # Token duy nhất để nhân viên nhập vào Agent khi cài máy
    enroll_token = Column(String, unique=True, default=lambda: str(uuid.uuid4())[:8].upper())

# 3. Người dùng (Level 1-4)
class User(Base):
    __tablename__ = 'users'
    id = Column(Integer, primary_key=True, index=True)
    username = Column(String, unique=True, index=True)
    hashed_password = Column(String)
    level = Column(Integer) # 1: SENT Admin, 2: SME Admin, 3: Operator, 4: Viewer
    org_id = Column(Integer, ForeignKey('organizations.id'), nullable=True)
    two_fa_secret = Column(String, nullable=True) # Lưu mã 2FA
    role_id = Column(Integer, ForeignKey("roles.id"), nullable=True)

# 4. Máy trạm (Agent)
class Agent(Base):
    __tablename__ = 'agents'
    id = Column(Integer, primary_key=True, index=True)
    hwid = Column(String, unique=True, index=True) # Mã phần cứng từ Go
    hostname = Column(String)
    org_id = Column(Integer, ForeignKey('organizations.id'))
    region_id = Column(Integer, ForeignKey('regions.id'))
    status = Column(String, default='pending') # pending, approved, blocked
    last_seen = Column(DateTime, default=datetime.datetime.utcnow)