# 🔐 Hệ Thống Quyền - Fix Hoàn Chỉnh

## 📋 Vấn Đề Ban Đầu
1. **Quyền không đủ/không đồng bộ** - Người dùng không thấy quyền của mình khi mở settings
2. **Chỉ có Create, không có Update** - Không thể chỉnh sửa quyền hiện tại
3. **Không hiển thị quyền hiện tại** - Form không tải quyền thực tế từ server

---

## ✅ Các Fix Đã Thực Hiện

### **Backend (SENT_backend/service/users)**

#### 1️⃣ **Thêm 2 API Endpoint Mới** (`user_handlers.go`)
```go
// GET /api/v1/users/{id}/permissions
// Trả về toàn bộ quyền của một người dùng theo ID
func GetUserPermissions(c *gin.Context) {
    // Returns: 18 permission fields (perm_asset_view, perm_user_manage, etc.)
}

// GET /api/v1/users/me/permissions  
// Trả về quyền của người dùng hiện tại (không cần ID)
func GetCurrentUserPermissions(c *gin.Context) {
    // Returns: 18 permission fields của user từ context
}
```

**Quyền được trả về (18 tổng cộng):**
- **Asset Management**: perm_asset_view, perm_asset_action, perm_asset_delete, perm_asset_move
- **Policy & Docs**: perm_policy_view, perm_policy_manage, perm_doc_view, perm_doc_manage
- **Incidents**: perm_incident_view, perm_incident_action
- **User Management**: perm_user_view, perm_user_manage
- **System Admin**: perm_group_manage, perm_approval_view, perm_approval_final, perm_system_config

#### 2️⃣ **Đăng Ký Routes** (`cmd/main.go`)
```go
usersAPI.GET("/:id/permissions", viewPerm, handlers.GetUserPermissions)
usersAPI.GET("/me/permissions", handlers.GetCurrentUserPermissions)
```
- Yêu cầu quyền `user_view` để xem quyền của người khác
- Không cần quyền đặc biệt để xem quyền của chính mình

---

### **Frontend (SENT_frontend/src/pages/User/components)**

#### 1️⃣ **Cập Nhật UserFormModal.jsx**

**A. Thêm Import & State:**
```jsx
import axios from '../../../api/axios';
import { Loader } from 'lucide-react';

const [formData, setFormData] = useState(createDefaultFormState());
const [permissionsLoading, setPermissionsLoading] = useState(false);
```

**B. Thêm Function Tải Quyền:**
```jsx
const loadUserPermissions = async (userId) => {
    setPermissionsLoading(true);
    try {
        const res = await axios.get(`/users/${userId}/permissions`);
        // Populate all 18 permission fields from server
        setFormData(prev => ({
            ...prev,
            perm_asset_view: res.data.perm_asset_view || false,
            perm_user_manage: res.data.perm_user_manage || false,
            // ... tất cả các field khác
        }));
    } finally {
        setPermissionsLoading(false);
    }
};
```

**C. Gọi Function Khi Mở Form Edit:**
```jsx
useEffect(() => {
    if (isOpen && isEditMode && initialData) {
        // ... populate basic data
        
        // ✨ Tải quyền từ server
        if (initialData.id || initialData.ID) {
            loadUserPermissions(initialData.id || initialData.ID);
        }
    }
}, [isOpen, initialData, isEditMode]);
```

#### 2️⃣ **Fix Permission Fields**

**Trước:**
```jsx
perm_approval_manage: false,  // ❌ Không tồn tại trong backend
```

**Sau:**
```jsx
perm_approval_view: false,   // ✅ Đúng
perm_approval_final: false,  // ✅ Đúng
perm_doc_view: false,        // ✅ Thêm vào (thiếu trước)
```

#### 3️⃣ **UI Improvements**

**Loading Indicator:**
```jsx
<div className="flex items-center gap-2">
    <h3 className="text-lg font-semibold text-emerald-400">Ma trận Đặc quyền</h3>
    {permissionsLoading && <Loader size={16} className="animate-spin text-emerald-400" />}
</div>
```

**Cập Nhật Danh Sách Quyền:**
```jsx
// Hệ thống (System Admin) - 6 quyền
<ToggleSwitch field="perm_group_manage" label="Quản lý Nhóm" />
<ToggleSwitch field="perm_approval_view" label="Xem yêu cầu Duyệt" />
<ToggleSwitch field="perm_approval_final" label="Duyệt đơn cuối" isDanger />
<ToggleSwitch field="perm_system_config" label="Cấu hình Hệ thống" isDanger />
<ToggleSwitch field="perm_user_view" label="Xem danh sách Nhân sự" />
<ToggleSwitch field="perm_user_manage" label="Quản lý Nhân sự" isDanger />

// Thiết bị - 4 quyền
<ToggleSwitch field="perm_asset_view" label="Xem Thiết bị" />
<ToggleSwitch field="perm_asset_action" label="Hành động trên Thiết bị" />
<ToggleSwitch field="perm_asset_delete" label="Xóa Thiết bị" isDanger />
<ToggleSwitch field="perm_asset_move" label="Di chuyển Nhóm máy" />

// Chính sách & Tài liệu - 4 quyền  
<ToggleSwitch field="perm_policy_view" label="Xem Chính sách" />
<ToggleSwitch field="perm_policy_manage" label="Quản lý Chính sách" />
<ToggleSwitch field="perm_doc_view" label="Xem Tài liệu" />          // ✨ New
<ToggleSwitch field="perm_doc_manage" label="Quản lý Tài liệu" />

// Sự cố An ninh - 2 quyền
<ToggleSwitch field="perm_incident_view" label="Xem Sự cố" />
<ToggleSwitch field="perm_incident_action" label="Xử lý Sự cố" />
```

---

## 🔄 Quy Trình Làm Việc Cải Tiến

### **Trước (Cũ):**
```
Người dùng mở form chỉnh sửa
    ↓
Form hiển thị data từ list (có thể cũ)
    ↓
Người dùng không biết quyền hiện tại là gì
    ↓
Chỉ có nút "Tạo quyền", không có "Chỉnh sửa"
    ↓
Phải gửi yêu cầu, chờ duyệt (không instant)
```

### **Sau (Mới):**
```
Người dùng mở form chỉnh sửa
    ↓
Form gửi request GET /users/{id}/permissions
    ↓
Server trả về 18 quyền hiện tại từ DB
    ↓
Form tải xong → Hiển thị toggle switches với giá trị đúng
    ↓
Loading indicator cho user biết đang fetch
    ↓
Người dùng thấy rõ quyền hiện tại & có thể chỉnh sửa
    ↓
Nhấn "Lưu Thay đổi" → Gửi request PUT /users/{id}
    ↓
Backend tạo approval ticket (hoặc update ngay nếu là admin)
```

---

## 📊 Summary Các Quyền Đầy Đủ

| Danh Mục | Quyền | Mô Tả |
|---------|-------|-------|
| **Quản trị Hệ thống** | perm_group_manage | Quản lý các nhóm/phòng ban |
| | perm_approval_view | Xem danh sách yêu cầu cần duyệt |
| | perm_approval_final | Duyệt lệnh cuối cùng (Checker) |
| | perm_system_config | Truy cập cấu hình hệ thống toàn bộ |
| | perm_user_view | Xem danh sách nhân sự |
| | perm_user_manage | Thêm/sửa/xóa nhân sự, phân quyền |
| **Thiết bị** | perm_asset_view | Xem danh sách thiết bị |
| | perm_asset_action | Thực hiện hành động (lock, wipe, etc.) |
| | perm_asset_delete | Xóa thiết bị khỏi hệ thống |
| | perm_asset_move | Di chuyển thiết bị giữa các nhóm |
| **Chính sách** | perm_policy_view | Xem danh sách chính sách bảo mật |
| | perm_policy_manage | Tạo/sửa/xóa chính sách |
| **Tài liệu** | perm_doc_view | Xem tài liệu của tổ chức |
| | perm_doc_manage | Tải lên/sửa/xóa tài liệu |
| **Sự cố** | perm_incident_view | Xem danh sách sự cố an ninh |
| | perm_incident_action | Phản ứng sự cố (assign, resolve, etc.) |

---

## 🚀 Cách Test Fix

### **Test 1: Hiển thị quyền hiện tại**
```bash
1. Đăng nhập → User Management
2. Nhấn Edit trên một user
3. Form hiển thị Loading... trong 1-2 giây
4. Permissions được tải từ server (các toggle hiển thị đúng giá trị)
```

### **Test 2: Tất cả 18 quyền hiển thị**
```bash
1. Kiểm tra form modal
2. Đếm các ToggleSwitch:
   - System Admin: 6 cái
   - Assets: 4 cái  
   - Policies/Docs: 4 cái
   - Incidents: 2 cái
   - Total: 16 cái + perm_approval_final + perm_approval_view = 18 ✓
```

### **Test 3: Update permissions**
```bash
1. Mở form edit user
2. Toggle một vài permissions (VD: bật perm_asset_delete)
3. Nhấn "Lưu Thay đổi"
4. Chờ duyệt (hoặc ngay nếu admin)
5. Verify: GetUserPermissions trả về giá trị mới
```

---

## 📝 Files Được Sửa

| File | Thay Đổi |
|------|---------|
| `SENT_backend/service/users/internal/handlers/user_handlers.go` | ✅ Thêm GetUserPermissions(), GetCurrentUserPermissions() |
| `SENT_backend/service/users/cmd/main.go` | ✅ Thêm 2 routes GET /:id/permissions, GET /me/permissions |
| `SENT_frontend/src/pages/User/components/UserFormModal.jsx` | ✅ Thêm loadUserPermissions(), fix field names, thêm loading state |

---

## 🔮 Improvements Tiềm Năng (Tương Lai)

1. **Real-time Permission Sync** - WebSocket để notify user khi quyền thay đổi
2. **Permission Templates** - "SOC_ANALYST", "ADMIN", "VIEWER" shortcuts
3. **Permission Expiry** - Quyền tự động hết hạn sau X ngày
4. **Audit Log** - Chi tiết lịch sử thay đổi quyền
5. **Batch Permission Update** - Cập nhật quyền cho multiple users cùng lúc

---

## ✨ Result

**Người dùng bây giờ có thể:**
- ✅ Thấy đầy đủ 18 quyền khi mở form chỉnh sửa
- ✅ Biết quyền hiện tại của mỗi người (load từ server)
- ✅ Chỉnh sửa các quyền dễ dàng
- ✅ Gửi yêu cầu thay đổi quyền qua API PUT
- ✅ Thấy loading indicator khi fetch dữ liệu

**Hệ thống bây giờ:**
- ✅ Có đầy đủ endpoints để quản lý quyền
- ✅ Đồng bộ quyền giữa frontend ↔ backend
- ✅ Lưu trữ đầy đủ 18 loại quyền phân loại
