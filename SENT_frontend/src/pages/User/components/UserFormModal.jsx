import React, { useState, useEffect } from 'react';
import { X, User, Mail, Phone, Lock, Shield, CheckSquare, Square, Loader2, Badge } from 'lucide-react';

const UserFormModal = ({ isOpen, onClose, onSubmit, initialData, isLoading }) => {
    const isEditMode = !!initialData;

    const [formData, setFormData] = useState({
        employee_id: '', // MÃ NHÂN VIÊN
        username: '',
        full_name: '',
        email: '',
        phone: '',
        password: '',
        role: 'USER',
        can_manage_agents: false,
        can_manage_policies: false,
        can_manage_incidents: false,
        can_manage_users: false,
    });

    useEffect(() => {
        if (initialData) {
            setFormData({
                ...initialData,
                employee_id: initialData.employee_id || '',
                password: '', 
            });
        }
    }, [initialData]);

    if (!isOpen) return null;

    const handleSubmit = (e) => {
        e.preventDefault();
        onSubmit(formData);
    };

    return (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-[9999] backdrop-blur-sm p-4 animate-in fade-in">
            <div className="bg-[#1e293b] w-full max-w-2xl rounded-3xl border border-slate-700 shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
                
                <div className="p-6 border-b border-slate-800 flex justify-between items-center bg-slate-800/50 shrink-0">
                    <div>
                        <h3 className="text-lg font-bold text-white flex items-center gap-2">
                            {isEditMode ? <Shield className="text-blue-400"/> : <User className="text-emerald-400"/>}
                            {isEditMode ? 'Cập nhật Tài khoản' : 'Thêm Nhân sự Mới'}
                        </h3>
                    </div>
                    <button onClick={onClose} className="text-slate-500 hover:text-white transition"><X size={20}/></button>
                </div>

                <div className="p-6 overflow-y-auto custom-scrollbar">
                    <form id="userForm" onSubmit={handleSubmit} className="space-y-6">
                        
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            {/* THÊM TRƯỜNG MÃ NHÂN VIÊN Ở ĐÂY */}
                            <InputRow label="Mã Nhân Viên (Tùy chọn)" icon={Badge}>
                                <input type="text" value={formData.employee_id} onChange={e => setFormData({...formData, employee_id: e.target.value})} className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 pl-10 text-sm text-white focus:border-emerald-500 outline-none uppercase" placeholder="VD: SEC-001" />
                            </InputRow>
                            
                            <InputRow label="Tên Đăng nhập" icon={User} required disabled={isEditMode}>
                                <input type="text" required={!isEditMode} disabled={isEditMode} value={formData.username} onChange={e => setFormData({...formData, username: e.target.value})} className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 pl-10 text-sm text-white focus:border-emerald-500 outline-none disabled:opacity-50" placeholder="vd: tung_nguyen" />
                            </InputRow>
                            
                            <InputRow label="Họ và Tên" icon={User} required>
                                <input type="text" required value={formData.full_name} onChange={e => setFormData({...formData, full_name: e.target.value})} className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 pl-10 text-sm text-white focus:border-emerald-500 outline-none" placeholder="Nguyễn Văn Tùng" />
                            </InputRow>
                            
                            <InputRow label="Email" icon={Mail} required>
                                <input type="email" required value={formData.email} onChange={e => setFormData({...formData, email: e.target.value})} className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 pl-10 text-sm text-white focus:border-emerald-500 outline-none" placeholder="tung@company.com" />
                            </InputRow>

                            <InputRow label="Số điện thoại" icon={Phone}>
                                <input type="text" value={formData.phone} onChange={e => setFormData({...formData, phone: e.target.value})} className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 pl-10 text-sm text-white focus:border-emerald-500 outline-none" placeholder="0987..." />
                            </InputRow>
                        </div>

                        <div className="bg-slate-800/30 p-4 rounded-xl border border-slate-700/50">
                            <InputRow label={isEditMode ? "Mật khẩu mới (Để trống nếu không đổi)" : "Mật khẩu"} icon={Lock} required={!isEditMode}>
                                <input type="password" required={!isEditMode} value={formData.password} onChange={e => setFormData({...formData, password: e.target.value})} className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 pl-10 text-sm text-white focus:border-emerald-500 outline-none" placeholder="••••••••" />
                            </InputRow>
                        </div>

                        <div>
                            <h4 className="text-xs font-black text-slate-500 uppercase tracking-widest mb-3 border-b border-slate-800 pb-2">Phân quyền Hệ thống</h4>
                            <div className="space-y-4">
                                <div className="flex items-center gap-4">
                                    <span className="text-sm font-bold text-slate-300 w-1/4">Vai trò (Role)</span>
                                    <select value={formData.role} onChange={e => setFormData({...formData, role: e.target.value})} className="flex-1 bg-slate-900 border border-slate-700 rounded-lg p-2.5 text-sm text-white focus:border-emerald-500 outline-none cursor-pointer">
                                        <option value="USER">Nhân viên (USER)</option>
                                        <option value="ADMIN">Quản trị viên (ADMIN)</option>
                                    </select>
                                </div>

                                {formData.role === 'USER' && (
                                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2">
                                        <Checkbox field="can_manage_agents" label="Quản lý Máy trạm" data={formData} setData={setFormData}/>
                                        <Checkbox field="can_manage_policies" label="Tạo & Sửa Luật bảo mật" data={formData} setData={setFormData}/>
                                        <Checkbox field="can_manage_incidents" label="Xử lý Sự cố (Incidents)" data={formData} setData={setFormData}/>
                                        <Checkbox field="can_manage_users" label="Thêm/Xóa Nhân sự" data={formData} setData={setFormData}/>
                                    </div>
                                )}
                            </div>
                        </div>

                    </form>
                </div>

                <div className="p-4 border-t border-slate-800 bg-slate-800/50 flex justify-end gap-3 shrink-0">
                    <button type="button" onClick={onClose} className="px-5 py-2.5 rounded-xl text-sm font-bold text-slate-400 hover:text-white transition">Hủy bỏ</button>
                    <button type="submit" form="userForm" disabled={isLoading} className="px-6 py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded-xl text-sm font-bold shadow-lg transition flex items-center gap-2 disabled:opacity-50">
                        {isLoading ? <Loader2 size={16} className="animate-spin"/> : null}
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

const Checkbox = ({ field, label, data, setData }) => (
    <div 
        onClick={() => setData({...data, [field]: !data[field]})}
        className={`flex items-center gap-3 p-3 rounded-xl border cursor-pointer transition select-none ${data[field] ? 'bg-emerald-500/10 border-emerald-500/30' : 'bg-slate-900 border-slate-700 hover:border-slate-500'}`}
    >
        {data[field] ? <CheckSquare size={18} className="text-emerald-500"/> : <Square size={18} className="text-slate-600"/>}
        <span className={`text-sm font-bold ${data[field] ? 'text-emerald-400' : 'text-slate-400'}`}>{label}</span>
    </div>
);

export default UserFormModal;