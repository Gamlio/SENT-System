import React, { useState, useEffect } from 'react';
import { X, User, Mail, Phone,  Shield, Loader2, Monitor, FileText, AlertTriangle, Settings } from 'lucide-react';

const UserFormModal = ({ isOpen, onClose, onSubmit, initialData, isLoading }) => {
    const isEditMode = !!initialData;

    const defaultState = {
        employee_id: '', username: '', full_name: '', email: '', phone: '', password: '',
        perm_agent_view: false, perm_agent_action: false, perm_agent_delete: false,
        perm_policy_view: false, perm_policy_action: false,
        perm_incident_view: false, perm_incident_action: false,
        perm_doc_view: false, perm_doc_manage: false,
        perm_user_manage: false, perm_approval_manage: false,
    };

    const [formData, setFormData] = useState(defaultState);

    useEffect(() => {
        if (isOpen) {
            if (isEditMode && initialData) {
                setFormData({
                    ...initialData,
                    password: '', // Không bao giờ hiện pass cũ
                    phone: initialData.phone || '', 
                    perm_agent_view: !!initialData.perm_agent_view,
                    perm_agent_action: !!initialData.perm_agent_action,
                    perm_agent_delete: !!initialData.perm_agent_delete,
                    perm_policy_view: !!initialData.perm_policy_view,
                    perm_policy_action: !!initialData.perm_policy_action,
                    perm_incident_view: !!initialData.perm_incident_view,
                    perm_incident_action: !!initialData.perm_incident_action,
                    perm_doc_view: !!initialData.perm_doc_view,
                    perm_doc_manage: !!initialData.perm_doc_manage,
                    perm_user_manage: !!initialData.perm_user_manage,
                    perm_approval_manage: !!initialData.perm_approval_manage,
                });
            } else {
                setFormData(defaultState);
            }
        }
    }, [initialData, isEditMode, isOpen]);

    const handleToggle = (field) => {
        setFormData(prev => {
            const val = !prev[field];
            const updated = { ...prev, [field]: val };
            // Logic liên đới
            if (val) {
                if (['perm_agent_action', 'perm_agent_delete'].includes(field)) updated.perm_agent_view = true;
                if (field === 'perm_policy_action') updated.perm_policy_view = true;
                if (field === 'perm_incident_action') updated.perm_incident_view = true;
                if (field === 'perm_doc_manage') updated.perm_doc_view = true;
            }
            return updated;
        });
    };

    const handleSubmit = (e) => {
        e.preventDefault();
        // LÀM SẠCH DỮ LIỆU trước khi gửi
        const payload = { ...formData };
        delete payload.id;
        delete payload.CreatedAt;
        delete payload.UpdatedAt;
        delete payload.DeletedAt;
        
        onSubmit(payload);
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-[9999] backdrop-blur-sm p-4 animate-in fade-in">
            <div className="bg-[#1e293b] w-full max-w-4xl rounded-3xl border border-slate-700 shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
                <div className="p-6 border-b border-slate-800 flex justify-between items-center bg-slate-800/50 shrink-0">
                    <h3 className="text-lg font-bold text-white flex items-center gap-2">
                        {isEditMode ? <Shield className="text-blue-400"/> : <User className="text-emerald-400"/>}
                        {isEditMode ? 'Cập nhật Quyền & Tài khoản' : 'Thêm Nhân sự Mới'}
                    </h3>
                    <button onClick={onClose} className="text-slate-500 hover:text-white transition"><X size={20}/></button>
                </div>

                <div className="p-6 overflow-y-auto custom-scrollbar flex-1">
                    <form id="userForm" onSubmit={handleSubmit} className="space-y-8">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <InputRow label="Tên đăng nhập" icon={User} required disabled={isEditMode}>
                                <input type="text" disabled={isEditMode} value={formData.username} onChange={e => setFormData({...formData, username: e.target.value})} className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 pl-10 text-sm text-white focus:border-emerald-500 outline-none disabled:opacity-50" />
                            </InputRow>
                            <InputRow label="Họ và Tên" icon={User} required>
                                <input type="text" required value={formData.full_name} onChange={e => setFormData({...formData, full_name: e.target.value})} className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 pl-10 text-sm text-white focus:border-emerald-500 outline-none" />
                            </InputRow>
                            <InputRow label="Số điện thoại" icon={Phone} required>
                                 <input type="text" required value={formData.phone} onChange={e => setFormData({...formData, phone: e.target.value})} className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 pl-10 text-sm text-white focus:border-emerald-500 outline-none" />
                            </InputRow>
                            <InputRow label="Email" icon={Mail} required>
                                <input type="email" required value={formData.email} onChange={e => setFormData({...formData, email: e.target.value})} className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 pl-10 text-sm text-white focus:border-emerald-500 outline-none" />
                            </InputRow>
                        </div>

                        <div>
                            <h4 className="text-xs font-black text-slate-500 uppercase tracking-widest mb-4 border-b border-slate-800 pb-2">Ma trận Đặc quyền (Permissions)</h4>
                            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                                <PermissionBlock title="Máy trạm" icon={<Monitor size={16}/>} color="text-emerald-400">
                                    <ToggleSwitch field="perm_agent_view" label="Xem danh sách" data={formData} onToggle={handleToggle}/>
                                    <ToggleSwitch field="perm_agent_action" label="Chỉnh sửa máy" data={formData} onToggle={handleToggle}/>
                                    <ToggleSwitch field="perm_agent_delete" label="Xóa máy" data={formData} onToggle={handleToggle} isDanger/>
                                </PermissionBlock>
                                <PermissionBlock title="Sự cố" icon={<AlertTriangle size={16}/>} color="text-red-400">
                                    <ToggleSwitch field="perm_incident_view" label="Xem sự cố" data={formData} onToggle={handleToggle}/>
                                    <ToggleSwitch field="perm_incident_action" label="Xử lý sự cố" data={formData} onToggle={handleToggle}/>
                                </PermissionBlock>
                                <PermissionBlock title="Tài liệu & Chính sách" icon={<FileText size={16}/>} color="text-blue-400">
                                    <ToggleSwitch field="perm_policy_view" label="Xem Chính sách" data={formData} onToggle={handleToggle}/>
                                    <ToggleSwitch field="perm_policy_action" label="Đổi Chính sách" data={formData} onToggle={handleToggle}/>
                                    <ToggleSwitch field="perm_doc_view" label="Xem Tài liệu " data={formData} onToggle={handleToggle}/>
                                    <ToggleSwitch field="perm_doc_manage" label="Quản lý Tài liệu " data={formData} onToggle={handleToggle}/>
                                </PermissionBlock>
                                <PermissionBlock title="Quản trị Hệ thống" icon={<Settings size={16}/>} color="text-purple-400">
                                    <ToggleSwitch field="perm_approval_manage" label="Duyệt yêu cầu (Maker-Checker)" data={formData} onToggle={handleToggle}/>
                                    <ToggleSwitch field="perm_user_manage" label="Quản lý Nhân sự (Thêm/Xóa)" data={formData} onToggle={handleToggle} isDanger/>
                                </PermissionBlock>
                            </div>
                        </div>
                    </form>
                </div>

                <div className="p-4 border-t border-slate-800 bg-slate-800/50 flex justify-end gap-3 shrink-0">
                    <button type="button" onClick={onClose} className="px-5 py-2.5 rounded-xl text-sm font-bold text-slate-400 hover:text-white transition">Hủy bỏ</button>
                    <button type="submit" form="userForm" disabled={isLoading} className="px-6 py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded-xl text-sm font-bold shadow-lg transition flex items-center gap-2">
                        {isLoading && <Loader2 size={16} className="animate-spin"/>}
                        {isEditMode ? 'Lưu Cập Nhật' : 'Tạo Tài Khoản'}
                    </button>
                </div>
            </div>
        </div>
    );
};

const InputRow = ({ label, icon: Icon, required, disabled, children }) => (
    <div className="relative">
        <label className="block text-xs font-bold text-slate-400 mb-1.5 pl-1">{label} {required && <span className="text-red-500">*</span>}</label>
        <div className="relative">
            <Icon size={16} className={`absolute left-3 top-1/2 -translate-y-1/2 ${disabled ? 'text-slate-600' : 'text-slate-400'}`}/>
            {children}
        </div>
    </div>
);

const PermissionBlock = ({ title, icon, color, children }) => (
    <div className="bg-slate-900/50 border border-slate-700 p-4 rounded-xl">
        <div className={`flex items-center gap-2 mb-3 font-bold ${color}`}>{icon} {title}</div>
        <div className="space-y-2">{children}</div>
    </div>
);

// Component Gạt (Toggle Switch) trực quan
const ToggleSwitch = ({ field, label, data, onToggle, isDanger }) => {
    const isChecked = !!data[field];
    const activeColor = isDanger ? 'bg-red-500' : 'bg-emerald-500';
    const bgContainer = isChecked 
        ? (isDanger ? 'bg-red-500/10 border-red-500/30' : 'bg-emerald-500/10 border-emerald-500/30') 
        : 'bg-[#1e293b] border-slate-700 hover:border-slate-500';

    return (
        <div onClick={() => onToggle(field)} className={`flex items-center justify-between p-3 rounded-xl border cursor-pointer transition-all select-none ${bgContainer}`}>
            <span className={`text-sm font-bold transition-colors ${isChecked ? (isDanger ? 'text-red-400' : 'text-emerald-400') : 'text-slate-400'}`}>
                {label}
            </span>
            <div className={`relative w-10 h-5 rounded-full transition-colors duration-300 ease-in-out ${isChecked ? activeColor : 'bg-slate-700'}`}>
                <div className={`absolute top-1 left-1 bg-white w-3 h-3 rounded-full transition-transform duration-300 ease-in-out ${isChecked ? 'translate-x-5' : 'translate-x-0 shadow-sm'}`}></div>
            </div>
        </div>
    );
};

export default UserFormModal;