@router.get("/")
def get_alerts(current_user = Depends(get_current_user), db: Session = Depends(database.get_db)):
    query = db.query(models.Alert)
    # Phân quyền: Level 1 thấy hết, Level 2 thấy của Org mình
    if current_user.level > 1:
        query = query.filter(models.Alert.org_id == current_user.org_id)
    return query.order_by(models.Alert.created_at.desc()).all()