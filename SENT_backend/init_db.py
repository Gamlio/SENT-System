from database import SessionLocal
import models
from core import security

def init_system():
    db = SessionLocal()
    try:
        # 1. Tạo Organization mặc định cho hệ thống [cite: 6]
        org_name = "SENT Global System"
        existing_org = db.query(models.Organization).filter(models.Organization.name == org_name).first()
        
        if not existing_org:
            new_org = models.Organization(name=org_name)
            db.add(new_org)
            db.flush() 
            org_id = new_org.id
            print(f"✅ Đã tạo Org hệ thống: {org_name}")
        else:
            org_id = existing_org.id
            print(f"ℹ️ Org '{org_name}' đã tồn tại.")

        # 2. Tạo Super Admin (Level 1) 
        admin_user = "Admin" # Khớp với username bạn nhập trong log
        existing_user = db.query(models.User).filter(models.User.username == admin_user).first()
        
        if not existing_user:
            # Dùng hàm get_password_hash vừa thêm ở trên 
            hashed_pass = security.get_password_hash("Thanh@123")
            new_user = models.User(
                username=admin_user,
                hashed_password=hashed_pass,
                level=1,      # Cấp độ cao nhất 
                org_id=org_id, # Thuộc Org hệ thống [cite: 7]
                role_id=None   # Level 1 không cần Rules vì là Master
            )
            db.add(new_user)
            print(f"✅ Đã tạo Super Admin: {admin_user}")
        else:
            # Cập nhật lại mật khẩu nếu đã tồn tại để đảm bảo băm đúng bcrypt
            existing_user.hashed_password = security.get_password_hash("Thanh@123")
            print(f"ℹ️ Đã cập nhật mật khẩu chuẩn bcrypt cho: {admin_user}")

        db.commit()
    except Exception as e:
        db.rollback()
        print(f"❌ Lỗi khởi tạo: {e}")
    finally:
        db.close()

if __name__ == "__main__":
    init_system()