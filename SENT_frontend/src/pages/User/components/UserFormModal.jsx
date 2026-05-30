import React, { useState, useEffect } from 'react';
import { User, Lock, Mail, Shield, Settings, X, Save, Key, Briefcase, Users, FileText, AlertTriangle, Loader } from 'lucide-react';
import axios from '../../../api/axios';

// --- Placeholder Components (Giả lập để code dễ đọc) ---
const Modal = ({ isOpen, onClose, children, title }) => {
    if (!isOpen) return null;
    return (
        <div className="fixed inset-0 bg-black bg-opacity-70 z-50 flex justify-center items-center animate-fade-in">
            <div className="bg-[#161d2b] rounded-2xl shadow-2xl w-full max-w-4xl border border-slate-700 transform transition-all duration-300 scale-95 animate-modal-pop-in">
                <div className="flex justify-between items-center p-5 border-b border-slate-800">
                    <h3 className="text-xl font-bold text-white">{title}</h3>
                    <button onClick={onClose} className="text-slate-400 hover:text-white transition-colors">
                        <X size={24} />
                    </button>
                </div>
                {children}
            </div>
        </div>
    );
};

const InputRow = ({ label, icon: Icon, children }) => (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-2 items-center">
        <label className="text-slate-400 text-sm font-medium flex items-center gap-3">
            {Icon && <Icon size={16} className="text-slate-500" />}
            {label}
        </label>
        <div className="md:col-span-2 relative">
             {Icon && <div className="absolute left-3 top-1/2 -translate-y-1/2 hidden">{/* Placeholder for potential icon inside input */}</div>}
            {children}
        </div>
    </div>
);

const PermissionBlock = ({ title, icon, color, children }) => (
    <div>
        <h4 className={`flex items-center gap-2 font-bold mb-3 ${color}`}>
            {icon}
            {title}
        </h4>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 pl-6">
            {children}
        </div>
    </div>
);

const ToggleSwitch = ({ field, label, data, onToggle, isDanger = false }) => (
    <div className="flex items-center justify-between bg-slate-800/50 p-3 rounded-lg border border-slate-700 hover:border-slate-600 transition-colors">
        <span className={`text-sm font-medium ${isDanger ? 'text-red-400' : 'text-slate-300'}`}>{label}</span>
        <label className="relative inline-flex items-center cursor-pointer">
            <input
                type="checkbox"
                checked={!!data[field]}
                onChange={() => onToggle(field, !data[field])}
                className="sr-only peer"
            />
            <div className="w-11 h-6 bg-slate-700 rounded-full peer peer-focus:ring-2 peer-focus:ring-emerald-500/50 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-emerald-600"></div>
        </label>
    </div>
);
// --- End Placeholder Components ---


const UserFormModal = ({ isOpen, onClose, onSubmit, initialData, isLoading, groups = [] }) => {
    const isEditMode = !!initialData?.ID;

    const createDefaultFormState = () => ({
        username: '',
        full_name: '',
        employee_id: '',
        email: '',
        password: '',
        group_id: null,
        perm_asset_view: false,
        perm_asset_action: false,
        perm_asset_delete: false,
        perm_asset_move: false,
        perm_policy_view: false,
        perm_policy_manage: false,
        perm_incident_view: false,
        perm_incident_action: false,
        perm_doc_view: false,
        perm_doc_manage: false,
        perm_user_view: false,
        perm_user_manage: false,
        perm_group_manage: false,
        perm_approval_view: false,
        perm_approval_final: false,
        perm_system_config: false,
    });

    const [formData, setFormData] = useState(createDefaultFormState());
    const [permissionsLoading, setPermissionsLoading] = useState(false);

    // Tải quyền hiện tại của người dùng
    const loadUserPermissions = async (userId) => {
        setPermissionsLoading(true);
        try {
            const res = await axios.get(`/users/${userId}/permissions`);
            const permData = res.data;
            setFormData(prev => ({
                ...prev,
                perm_asset_view: permData.perm_asset_view || false,
                perm_asset_action: permData.perm_asset_action || false,
                perm_asset_delete: permData.perm_asset_delete || false,
                perm_asset_move: permData.perm_asset_move || false,
                perm_policy_view: permData.perm_policy_view || false,
                perm_policy_manage: permData.perm_policy_manage || false,
                perm_incident_view: permData.perm_incident_view || false,
                perm_incident_action: permData.perm_incident_action || false,
                perm_doc_view: permData.perm_doc_view || false,
                perm_doc_manage: permData.perm_doc_manage || false,
                perm_user_view: permData.perm_user_view || false,
                perm_user_manage: permData.perm_user_manage || false,
                perm_group_manage: permData.perm_group_manage || false,
                perm_approval_view: permData.perm_approval_view || false,
                perm_approval_final: permData.perm_approval_final || false,
                perm_system_config: permData.perm_system_config || false,
            }));
        } catch (err) {
            console.error("Lỗi tải quyền:", err);
        } finally {
            setPermissionsLoading(false);
        }
    };

    useEffect(() => {
        if (isOpen) {
            if (isEditMode && initialData) {
                const populatedData = { ...createDefaultFormState(), ...initialData };
                populatedData.password = ''; // Không hiển thị password cũ
                setFormData(populatedData);
                
                // Tải quyền từ server để đảm bảo cập nhật nhất
                if (initialData.id || initialData.ID) {
                    loadUserPermissions(initialData.id || initialData.ID);
                }
            } else {
                setFormData(createDefaultFormState());
            }
        }
    }, [isOpen, initialData, isEditMode]);

    const handleInputChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    const handleToggle = (field, value) => {
        setFormData(prev => ({ ...prev, [field]: value }));
    };

    const handleSubmit = (e) => {
        e.preventDefault();
        onSubmit(formData);
    };

    return (
        <Modal isOpen={isOpen} onClose={onClose} title={isEditMode ? "Chỉnh sửa Thông tin Nhân sự" : "Thêm Nhân sự Mới"}>
            <form onSubmit={handleSubmit}>
                <div className="p-6 space-y-8 max-h-[80vh] overflow-y-auto">
                    {/* Phần 1: Thông tin cơ bản */}
                    <div className="p-5 bg-slate-900/50 border border-slate-800 rounded-xl space-y-4">
                        <h3 className="text-lg font-semibold text-emerald-400 mb-4">Thông tin Cơ bản</h3>
                        <InputRow label="Tên đăng nhập" icon={User}>
                            <input type="text" name="username" value={formData.username} onChange={handleInputChange} required disabled={isEditMode} className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 pl-4 text-sm text-white focus:border-emerald-500 outline-none disabled:bg-slate-800 disabled:text-slate-500" />
                        </InputRow>
                        <InputRow label="Họ và Tên" icon={User}>
                            <input type="text" name="full_name" value={formData.full_name || ''} onChange={handleInputChange} required className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 pl-4 text-sm text-white focus:border-emerald-500 outline-none" />
                        </InputRow>
                        <InputRow label="Mã nhân viên" icon={Briefcase}>
                            <input type="text" name="employee_id" value={formData.employee_id || ''} onChange={handleInputChange} className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 pl-4 text-sm text-white focus:border-emerald-500 outline-none" />
                        </InputRow>
                        <InputRow label="Email" icon={Mail}>
                            <input type="email" name="email" value={formData.email || ''} onChange={handleInputChange} required className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 pl-4 text-sm text-white focus:border-emerald-500 outline-none" />
                        </InputRow>
                        <InputRow label="Mật khẩu" icon={Lock}>
                            <input type="password" name="password" value={formData.password} onChange={handleInputChange} placeholder={isEditMode ? "Để trống nếu không đổi" : "Tối thiểu 8 ký tự"} required={!isEditMode} className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 pl-4 text-sm text-white focus:border-emerald-500 outline-none" />
                        </InputRow>
                        
                        {/* === [MỚI] Thêm trường Phòng ban === */}
                        <InputRow label="Phòng ban / Nhóm" icon={Shield}>
                            <select
                                value={formData.group_id || ''}
                                onChange={e => setFormData({ ...formData, group_id: e.target.value ? Number(e.target.value) : null })}
                                className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 text-sm text-white focus:border-emerald-500 outline-none appearance-none"
                            >
                                <option value="">-- Chọn phòng ban --</option>
                                {groups.map(g => (
                                    <option key={g.ID} value={g.ID}>{g.name}</option>
                                ))}
                            </select>
                        </InputRow>
                    </div>

                    {/* Phần 2: Ma trận đặc quyền */}
                    <div className="p-5 bg-slate-900/50 border border-slate-800 rounded-xl space-y-6">
                        <div className="flex items-center gap-2">
                            <h3 className="text-lg font-semibold text-emerald-400">Ma trận Đặc quyền</h3>
                            {permissionsLoading && <Loader size={16} className="animate-spin text-emerald-400" />}
                            {isEditMode && <span className="text-xs text-slate-500">({isEditMode ? 'Quyền hiện tại' : 'Quyền mới'})</span>}
                        </div>
                        
                        <PermissionBlock title="Quản trị Hệ thống" icon={<Settings size={16} />} color="text-purple-400">
                            <ToggleSwitch field="perm_group_manage" label="Quản lý Nhóm (Groups)" data={formData} onToggle={handleToggle} />
                            <ToggleSwitch field="perm_approval_view" label="Xem yêu cầu Duyệt" data={formData} onToggle={handleToggle} />
                            <ToggleSwitch field="perm_system_config" label="Cấu hình Hệ thống" data={formData} onToggle={handleToggle} isDanger />
                            <ToggleSwitch field="perm_approval_final" label="Duyệt đơn cuối (Checker)" data={formData} onToggle={handleToggle} isDanger />
                            <ToggleSwitch field="perm_user_view" label="Xem danh sách Nhân sự" data={formData} onToggle={handleToggle} />                                                                   
                            <ToggleSwitch field="perm_user_manage" label="Quản lý Nhân sự" data={formData} onToggle={handleToggle} isDanger />
                        </PermissionBlock>

                        <PermissionBlock title="Quản lý Thiết bị" icon={<Key size={16} />} color="text-sky-400">
                            <ToggleSwitch field="perm_asset_view" label="Xem Thiết bị" data={formData} onToggle={handleToggle} />
                            <ToggleSwitch field="perm_asset_action" label="Hành động trên Thiết bị" data={formData} onToggle={handleToggle} />
                            <ToggleSwitch field="perm_asset_delete" label="Xóa Thiết bị" data={formData} onToggle={handleToggle} isDanger />
                            <ToggleSwitch field="perm_asset_move" label="Di chuyển Nhóm máy" data={formData} onToggle={handleToggle} />
                        </PermissionBlock>

                        <PermissionBlock title="Chính sách & Tài liệu" icon={<FileText size={16} />} color="text-amber-400">
                            <ToggleSwitch field="perm_policy_view" label="Xem Chính sách" data={formData} onToggle={handleToggle} />
                            <ToggleSwitch field="perm_policy_manage" label="Quản lý Chính sách" data={formData} onToggle={handleToggle} />
                            <ToggleSwitch field="perm_doc_view" label="Xem Tài liệu" data={formData} onToggle={handleToggle} />
                            <ToggleSwitch field="perm_doc_manage" label="Quản lý Tài liệu" data={formData} onToggle={handleToggle} />
                        </PermissionBlock>

                        <PermissionBlock title="Sự cố An ninh" icon={<AlertTriangle size={16} />} color="text-rose-400">
                            <ToggleSwitch field="perm_incident_view" label="Xem Sự cố" data={formData} onToggle={handleToggle} />
                            <ToggleSwitch field="perm_incident_action" label="Xử lý Sự cố" data={formData} onToggle={handleToggle} />
                        </PermissionBlock>
                    </div>
                </div>

                {/* Phần 3: Nút bấm */}
                <div className="flex justify-end items-center p-5 bg-slate-900/50 border-t border-slate-800 rounded-b-2xl">
                    <button type="button" onClick={onClose} className="px-4 py-2.5 text-sm font-bold text-slate-300 hover:text-white rounded-lg mr-3">
                        Hủy
                    </button>
                    <button type="submit" disabled={isLoading} className="flex items-center gap-2 bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2.5 rounded-xl text-sm font-bold shadow-lg transition disabled:bg-slate-500 disabled:cursor-not-allowed">
                        <Save size={18} />
                        {isLoading ? 'Đang lưu...' : (isEditMode ? 'Lưu Thay đổi' : 'Tạo Tài khoản')}
                    </button>
                </div>
            </form>
        </Modal>
    );
};

export default UserFormModal;