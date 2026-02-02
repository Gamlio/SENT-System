import bcrypt
from datetime import datetime, timedelta
from jose import jwt

# Giữ nguyên cấu hình của bạn
SECRET_KEY = "SENT_SUPER_SECRET_KEY_2026" 
ALGORITHM = "HS256"

def hash_password(password: str) -> str:
    # Chuyển password sang dạng bytes
    pwd_bytes = password.encode('utf-8')
    # Tạo salt và hash
    salt = bcrypt.gensalt()
    hashed = bcrypt.hashpw(pwd_bytes, salt)
    # Trả về dạng string để lưu vào Database
    return hashed.decode('utf-8')

def verify_password(plain_password: str, hashed_password: str) -> bool:
    # Chuyển cả hai sang dạng bytes để so sánh
    password_byte_enc = plain_password.encode('utf-8')
    hashed_password_byte_enc = hashed_password.encode('utf-8')
    return bcrypt.checkpw(password_byte_enc, hashed_password_byte_enc)

def create_access_token(data: dict):
    to_encode = data.copy()
    expire = datetime.utcnow() + timedelta(hours=24)
    to_encode.update({"exp": expire})
    return jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)