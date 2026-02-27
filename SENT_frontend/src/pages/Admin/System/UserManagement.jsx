import React, { useState } from 'react';
import { UserPlus, Users, Trash2, Shield, User, Search, ChevronLeft, ChevronRight, Phone, Mail, CheckSquare, Edit, AlertTriangle } from 'lucide-react';
import { useUsers } from '../../../hooks/useUsers';
import { useAuth } from '../../../context/AuthContext'; // 1. IMPORT AUTH CONTEXT
import UserStats from '../../Dashboard/UserStats';

const UserManagement = () => {
    const {
        users, currentUsers, filteredUsers,
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages, indexOfFirstItem, indexOfLastItem,
        isLoading, createUser, deleteUser, updateUser 
    } = useUsers();

    // 2. LẤY USER HIỆN TẠI ĐỂ KIỂM TRA QUYỀN
    const { user: currentUser } = useAuth();

    // --- STATE QUẢN LÝ MODAL ---
    const [isEditMode, setIsEditMode] = useState(false);
    const [editingUserId, setEditingUserId] = useState(null);
    const [deleteModalOpen, setDeleteModalOpen] = useState(false);
    const [userToDelete, setUserToDelete] = useState(null);
    const [deleteConfirmText, setDeleteConfirmText] = useState("");

    // Form data
    const [formData, setFormData] = useState({ 
        username: '', password: '', full_name: '', phone: '', email: '', role: 'USER', 
        can_view_agents: true, can_manage_agents: false,
        can_view_docs: true, can_manage_docs: false,
        can_manage_policies: false, can_manage_incidents: false, can_manage_users: false
    });

    const resetForm = () => {
        setFormData({ 
            username: '', password: '', full_name: '', phone: '', email: '', role: 'USER',
            can_view_agents: true, can_manage_agents: false, can_view_docs: true, 
            can_manage_docs: false, can_manage_policies: false, can_manage_incidents: false, can_manage_users: false
        });
        setIsEditMode(false);
        setEditingUserId(null);
    };

    const handleEditClick = (user) => {
        setFormData({
            username: user.username, password: '', full_name: user.full_name, phone: user.phone, email: user.email, role: user.role,
            can_view_agents: user.can_view_agents, can_manage_agents: user.can_manage_agents,
            can_view_docs: user.can_view_docs, can_manage_docs: user.can_manage_docs,
            can_manage_policies: user.can_manage_policies, can_manage_incidents: user.can_manage_incidents,
            can_manage_users: user.can_manage_users,
        });
        setEditingUserId(user.id);
        setIsEditMode(true);
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (isEditMode) {
            await updateUser(editingUserId, formData);
            resetForm(); 
        } else {
            await createUser(formData);
            resetForm();
        }
    };

    const handleCheckboxChange = (field) => {
        setFormData(prev => ({ ...prev, [field]: !prev[field] }));
    };

    const openDeleteModal = (user) => {
        setUserToDelete(user);
        setDeleteConfirmText("");
        setDeleteModalOpen(true);
    };

    const handleConfirmDelete = async () => {
        if (deleteConfirmText === "DELETE" && userToDelete) {
            await deleteUser(userToDelete.id);
            setDeleteModalOpen(false);
            setUserToDelete(null);
        }
    };

    return (
        <div className="text-slate-200 relative">
            <header className="mb-8">
                <h1 className="text-3xl font-bold text-white">Quản lý & Phân bổ Người dùng</h1>
                <p className="text-slate-400 text-sm">Kiểm soát truy cập và theo dõi sự phát triển tài khoản trong hệ thống</p>
            </header>

            {/* --- MODAL XÓA AN TOÀN --- */}
            {deleteModalOpen && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-4">
                    <div className="bg-[#1e293b] border border-red-500/30 rounded-2xl p-6 max-w-md w-full shadow-2xl transform scale-100 transition-all">
                        <div className="flex items-center gap-3 text-red-500 mb-4">
                            <AlertTriangle size={32} />
                            <h2 className="text-xl font-bold">Xác nhận xóa tài khoản?</h2>
                        </div>
                        <p className="text-slate-300 mb-4">
                            Bạn đang chuẩn bị xóa nhân viên <span className="font-bold text-white">{userToDelete?.username}</span>. Hành động này không thể hoàn tác.
                        </p>
                        <div className="mb-6">
                            <label className="text-xs text-slate-500 uppercase font-bold block mb-2">Nhập "DELETE" để xác nhận:</label>
                            <input type="text" value={deleteConfirmText} onChange={(e) => setDeleteConfirmText(e.target.value)} className="w-full p-3 bg-black/30 border border-slate-600 rounded-xl text-white outline-none focus:border-red-500 font-mono" placeholder="DELETE"/>
                        </div>
                        <div className="flex justify-end gap-3">
                            <button onClick={() => setDeleteModalOpen(false)} className="px-4 py-2 text-slate-400 hover:text-white font-bold transition">Hủy bỏ</button>
                            <button onClick={handleConfirmDelete} disabled={deleteConfirmText !== "DELETE"} className="px-6 py-2 bg-red-600 hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold rounded-xl transition shadow-lg shadow-red-600/20">Xóa vĩnh viễn</button>
                        </div>
                    </div>
                </div>
            )}

            {/* --- COMPONENT THỐNG KÊ --- */}
            <UserStats users={users} />

            <div className="grid grid-cols-1 xl:grid-cols-3 gap-8">
                
                {/* FORM TẠO/SỬA */}
                <div className={`bg-[#1e293b] p-6 rounded-3xl border ${isEditMode ? 'border-amber-500/50 shadow-amber-500/10' : 'border-slate-800'} shadow-xl h-fit transition-all duration-300`}>
                    <div className="flex justify-between items-center mb-6">
                        <h3 className={`text-lg font-bold flex items-center gap-2 ${isEditMode ? 'text-amber-400' : 'text-white'}`}>
                            {isEditMode ? <><Edit size={20}/> Cập nhật nhân viên</> : <><UserPlus className="text-emerald-400" size={20}/> Cấp tài khoản mới</>}
                        </h3>
                        {isEditMode && <button onClick={resetForm} className="text-xs bg-slate-800 hover:bg-slate-700 px-3 py-1 rounded-lg text-slate-300 transition">Hủy sửa</button>}
                    </div>
                    
                    <form onSubmit={handleSubmit} className="space-y-4">
                        {/* INPUT HỌ TÊN */}
                        <div>
                            <label className="text-xs font-bold text-slate-500 uppercase">Họ và tên</label>
                            <input type="text" value={formData.full_name} required placeholder="VD: Nguyễn Văn A" onChange={(e) => setFormData({...formData, full_name: e.target.value})} className="w-full mt-2 p-3 bg-slate-900 border border-slate-700 rounded-xl outline-none focus:border-emerald-500 text-white text-sm" />
                        </div>
                        {/* INPUT USER/PASS */}
                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="text-xs font-bold text-slate-500 uppercase">Tên đăng nhập</label>
                                <input type="text" disabled={isEditMode} value={formData.username} required placeholder="VD: nguyenva" onChange={(e) => setFormData({...formData, username: e.target.value})} className={`w-full mt-2 p-3 bg-slate-900 border border-slate-700 rounded-xl outline-none text-white text-sm ${isEditMode ? 'opacity-50 cursor-not-allowed' : 'focus:border-emerald-500'}`} />
                            </div>
                            <div>
                                <label className="text-xs font-bold text-slate-500 uppercase">Mật khẩu {isEditMode && "(Bỏ trống nếu không đổi)"}</label>
                                <input type="password" value={formData.password} required={!isEditMode} minLength="6" placeholder="******" onChange={(e) => setFormData({...formData, password: e.target.value})} className="w-full mt-2 p-3 bg-slate-900 border border-slate-700 rounded-xl outline-none focus:border-emerald-500 text-white text-sm" />
                            </div>
                        </div>
                        {/* INPUT LIÊN HỆ */}
                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="text-xs font-bold text-slate-500 uppercase">Số điện thoại</label>
                                <input type="tel" value={formData.phone} placeholder="VD: 0912345678" onChange={(e) => setFormData({...formData, phone: e.target.value})} className="w-full mt-2 p-3 bg-slate-900 border border-slate-700 rounded-xl outline-none focus:border-emerald-500 text-white text-sm" />
                            </div>
                            <div>
                                <label className="text-xs font-bold text-slate-500 uppercase">Email</label>
                                <input type="email" value={formData.email} placeholder="VD: email@congty.com" onChange={(e) => setFormData({...formData, email: e.target.value})} className="w-full mt-2 p-3 bg-slate-900 border border-slate-700 rounded-xl outline-none focus:border-emerald-500 text-white text-sm" />
                            </div>
                        </div>
                        {/* CHECKBOXES */}
                        <div className="pt-4 border-t border-slate-800 mt-4">
                            <label className="text-xs font-bold text-slate-400 uppercase flex items-center gap-2 mb-3"><CheckSquare size={16}/> Phân quyền truy cập</label>
                            <div className="space-y-3">
                                <label className="flex items-center gap-3 cursor-pointer group"><input type="checkbox" checked={formData.can_view_agents} onChange={() => handleCheckboxChange('can_view_agents')} className="w-4 h-4 accent-emerald-500 bg-slate-800 border-slate-600 rounded" /><span className="text-sm text-slate-300 group-hover:text-white transition">Máy trạm (Xem danh sách & Chi tiết)</span></label>
                                <label className="flex items-center gap-3 cursor-pointer group"><input type="checkbox" checked={formData.can_manage_policies} onChange={() => handleCheckboxChange('can_manage_policies')} className="w-4 h-4 accent-emerald-500 bg-slate-800 border-slate-600 rounded" /><span className="text-sm text-slate-300 group-hover:text-white transition">Trung tâm Chính sách (Quản lý Luật)</span></label>
                                <label className="flex items-center gap-3 cursor-pointer group"><input type="checkbox" checked={formData.can_view_docs} onChange={() => handleCheckboxChange('can_view_docs')} className="w-4 h-4 accent-emerald-500 bg-slate-800 border-slate-600 rounded" /><span className="text-sm text-slate-300 group-hover:text-white transition">Tài liệu AI (Xem & Hỏi Copilot)</span></label>
                                <label className="flex items-center gap-3 cursor-pointer group pl-7"><input type="checkbox" checked={formData.can_manage_docs} onChange={() => handleCheckboxChange('can_manage_docs')} className="w-4 h-4 accent-emerald-500 bg-slate-800 border-slate-600 rounded" /><span className="text-xs text-slate-400 group-hover:text-white transition">↳ Cho phép tải lên / xóa tài liệu</span></label>
                                <label className="flex items-center gap-3 cursor-pointer group"><input type="checkbox" checked={formData.can_manage_users} onChange={() => handleCheckboxChange('can_manage_users')} className="w-4 h-4 accent-emerald-500 bg-slate-800 border-slate-600 rounded" /><span className="text-sm text-slate-300 group-hover:text-white transition">Người dùng (Cấp tài khoản & Hệ thống)</span></label>
                                <label className="flex items-center gap-3 cursor-pointer group"><input type="checkbox" checked={formData.can_manage_incidents} onChange={() => handleCheckboxChange('can_manage_incidents')} className="w-4 h-4 accent-emerald-500 bg-slate-800 border-slate-600 rounded" /><span className="text-sm text-slate-300 group-hover:text-white transition">Điều tra Sự cố (SOC & Playbook)</span></label>
                            </div>
                        </div>
                        <button type="submit" disabled={isLoading} className={`w-full mt-6 text-white font-bold py-3 rounded-xl transition shadow-lg disabled:opacity-50 ${isEditMode ? 'bg-amber-500 hover:bg-amber-600 shadow-amber-500/20' : 'bg-emerald-500 hover:bg-emerald-600 shadow-emerald-500/20'}`}>{isLoading ? "Đang xử lý..." : (isEditMode ? "Lưu thay đổi" : "Tạo tài khoản")}</button>
                    </form>
                </div>

                {/* DANH SÁCH USER */}
                <div className="xl:col-span-2 flex flex-col h-full">
                    <div className="mb-4 relative">
                        <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500" size={20}/>
                        <input type="text" placeholder="Tìm kiếm..." value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} className="w-full pl-12 pr-4 py-3 bg-[#1e293b] border border-slate-800 rounded-2xl text-white outline-none focus:border-emerald-500 transition shadow-lg"/>
                    </div>
                    <div className="bg-[#1e293b] rounded-3xl border border-slate-800 overflow-hidden shadow-xl flex-1 flex flex-col">
                        <div className="p-5 border-b border-slate-800 bg-slate-800/30 flex items-center gap-3">
                            <Users className="text-emerald-400" size={20}/>
                            <h3 className="font-bold text-white">Danh sách nhân sự</h3>
                        </div>
                        <div className="overflow-x-auto flex-1">
                            <table className="w-full text-left">
                                <thead className="text-[10px] uppercase tracking-widest text-slate-500 bg-slate-900/50">
                                    <tr><th className="p-5 font-bold">Người dùng</th><th className="p-5 font-bold">Liên hệ</th><th className="p-5 font-bold">Quyền</th><th className="p-5 font-bold text-right">Thao tác</th></tr>
                                </thead>
                                <tbody className="divide-y divide-slate-800">
                                    {currentUsers.map((u) => (
                                        <tr key={u.id} className={`transition-colors group ${editingUserId === u.id ? 'bg-amber-500/10' : 'hover:bg-slate-800/50'}`}>
                                            <td className="p-5 flex items-center gap-3">
                                                <div className="p-2 bg-slate-900 rounded-lg text-slate-400 group-hover:text-emerald-400 transition"><User size={16}/></div>
                                                <div><p className="font-bold text-white text-sm">{u.full_name || u.username}</p><p className="text-[10px] text-slate-500 font-mono">@{u.username}</p></div>
                                            </td>
                                            <td className="p-5"><div className="text-xs text-slate-400 space-y-1"><div className="flex items-center gap-1.5"><Phone size={12}/> {u.phone || '—'}</div><div className="flex items-center gap-1.5"><Mail size={12}/> {u.email || '—'}</div></div></td>
                                            <td className="p-5">{u.role === 'ADMIN' ? <span className="flex items-center gap-1.5 w-fit px-3 py-1 bg-amber-500/10 text-amber-400 border border-amber-500/20 rounded-full text-[10px] font-black uppercase"><Shield size={12}/> Admin</span> : <span className="flex items-center gap-1.5 w-fit px-3 py-1 bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded-full text-[10px] font-black uppercase"><User size={12}/> Nhân viên</span>}</td>
                                            
                                            {/* 3. CỘT THAO TÁC ĐÃ SỬA: CHECK QUYỀN TRƯỚC KHI HIỂN THỊ */}
                                            <td className="p-5 text-right">
                                                {/* Logic ẩn nút: Nếu mình KHÔNG phải Admin VÀ dòng này LÀ Admin -> Thì ẩn đi */}
                                                {currentUser.role !== 'ADMIN' && u.role === 'ADMIN' ? (
                                                    <span className="text-xs text-slate-500 italic">Không được phép</span>
                                                ) : (
                                                    <div className="flex justify-end gap-2">
                                                        <button onClick={() => handleEditClick(u)} className="p-2 text-slate-500 hover:text-amber-400 hover:bg-amber-500/10 rounded-lg transition" title="Sửa quyền"><Edit size={16}/></button>
                                                        
                                                        {/* Không cho xóa chính mình */}
                                                        {currentUser.username !== u.username && (
                                                            <button onClick={() => openDeleteModal(u)} className="p-2 text-slate-500 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition" title="Xóa"><Trash2 size={16}/></button>
                                                        )}
                                                    </div>
                                                )}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                            {filteredUsers.length === 0 && <div className="p-10 text-center text-slate-500 italic">Không tìm thấy tài khoản nào.</div>}
                        </div>
                        {totalPages > 1 && (
                            <div className="p-4 border-t border-slate-800 flex justify-between items-center bg-slate-900/30">
                                <span className="text-xs text-slate-500 font-medium">Hiển thị {indexOfFirstItem + 1}-{Math.min(indexOfLastItem, filteredUsers.length)} / {filteredUsers.length}</span>
                                <div className="flex items-center gap-2">
                                    <button onClick={() => setCurrentPage(p => Math.max(p - 1, 1))} disabled={currentPage === 1} className="p-1.5 rounded bg-slate-800 text-slate-400 hover:text-white disabled:opacity-50"><ChevronLeft size={16}/></button>
                                    <span className="text-xs font-bold text-emerald-400 px-2">{currentPage} / {totalPages}</span>
                                    <button onClick={() => setCurrentPage(p => Math.min(p + 1, totalPages))} disabled={currentPage === totalPages} className="p-1.5 rounded bg-slate-800 text-slate-400 hover:text-white disabled:opacity-50"><ChevronRight size={16}/></button>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default UserManagement;