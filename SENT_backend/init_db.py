from database import SessionLocal
import models
from core import security

def init_system():
    db = SessionLocal()
    try:
        # 1. Kiểm tra xem Organization đã tồn tại chưa
        org_name = "SENT System Admin"
        existing_org = db.query(models.Organization).filter(models.Organization.name == org_name).first()
        
        if not existing_org:
            new_org = models.Organization(name=org_name)
            db.add(new_org)
            db.flush() # Lấy ID của org mới mà chưa commit
            org_id = new_org.id
            print(f"Đã tạo Organization: {org_name}")
        else:
            org_id = existing_org.id
            print(f"Organization '{org_name}' đã tồn tại, bỏ qua bước tạo.")

        # 2. Kiểm tra xem User Admin đã tồn tại chưa
        admin_user = "admin"
        existing_user = db.query(models.User).filter(models.User.username == admin_user).first()
        
        if not existing_user:
            hashed_pass = security.hash_password("Thanh@123")
            new_user = models.User(
                username=admin_user,
                hashed_password=hashed_pass,
                level=1,
                org_id=org_id
            )
            db.add(new_user)
            print(f"Đã tạo User: {admin_user}")
        else:
            print(f"User '{admin_user}' đã tồn tại, không tạo mới.")

        db.commit()
    except Exception as e:
        db.rollback()
        print(f"Lỗi khởi tạo: {e}")
    finally:
        db.close()

if __name__ == "__main__":
    init_system()