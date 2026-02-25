import React, { useState } from 'react';
import { UserPlus, Users, Trash2, Shield, User, Activity, PieChart as PieIcon, Search, ChevronLeft, ChevronRight, Phone, Mail } from 'lucide-react';
import { PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { useUsers } from '../../../hooks/useUsers'; // Import Hook

const UserManagement = () => {
    // Kéo toàn bộ logic từ Hook ra
    const {
        users, currentUsers, filteredUsers,
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages, indexOfFirstItem, indexOfLastItem,
        isLoading, createUser, deleteUser
    } = useUsers();

    // 1. CẬP NHẬT STATE: Thêm full_name, phone, email
    const [formData, setFormData] = useState({ 
        username: '', password: '', role_level: 1,
        full_name: '', phone: '', email: '' 
    });

    const handleCreateSubmit = async (e) => {
        e.preventDefault();
        await createUser(formData);
        // Reset toàn bộ form sau khi tạo
        setFormData({ username: '', password: '', role_level: 1, full_name: '', phone: '', email: '' });
    };

    // --- BIỂU ĐỒ ĐỘNG: Tự đếm dựa trên số lượng User thật ---
    const roleStats = [
        { name: 'Admin (Level 2)', value: users.filter(u => u.role_level === 2).length, color: '#f59e0b' },
        { name: 'User (Level 1)', value: users.filter(u => u.role_level === 1).length, color: '#3b82f6' }
    ];

    const growthData = [
        { name: 'T10', users: 2 }, { name: 'T11', users: 4 }, { name: 'T12', users: 5 }, { name: 'Hiện tại', users: users.length }
    ];

    return (
        <div className="text-slate-200">
            <header className="mb-8">
                <h1 className="text-3xl font-bold text-white">Quản lý & Phân bổ Người dùng</h1>
                <p className="text-slate-400 text-sm">Kiểm soát truy cập và theo dõi sự phát triển tài khoản trong hệ thống</p>
            </header>

            {/* --- KHU VỰC BIỂU ĐỒ (VISUALIZATION) --- */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-8">
                {/* Biểu đồ Donut */}
                <div className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl flex flex-col h-64">
                    <h3 className="text-sm font-bold text-slate-400 uppercase tracking-widest mb-2 flex items-center gap-2">
                        <PieIcon size={16} className="text-emerald-400"/> Tỷ lệ Phân quyền
                    </h3>
                    <div className="flex-1">
                        <ResponsiveContainer width="100%" height="100%">
                            <PieChart>
                                <Pie data={roleStats} cx="50%" cy="50%" innerRadius={50} outerRadius={70} paddingAngle={5} dataKey="value">
                                    {roleStats.map((entry, index) => <Cell key={`cell-${index}`} fill={entry.color} />)}
                                </Pie>
                                <Tooltip contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '12px', color: '#fff' }} itemStyle={{ color: '#fff' }} />
                            </PieChart>
                        </ResponsiveContainer>
                    </div>
                    <div className="flex justify-center gap-6 mt-2">
                        {roleStats.map(stat => (
                            <div key={stat.name} className="flex items-center gap-2 text-xs font-bold">
                                <span className="w-3 h-3 rounded-full" style={{ backgroundColor: stat.color }}></span>
                                {stat.name}: <span className="text-white">{stat.value}</span>
                            </div>
                        ))}
                    </div>
                </div>

                {/* Biểu đồ Cột */}
                <div className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl flex flex-col h-64">
                    <h3 className="text-sm font-bold text-slate-400 uppercase tracking-widest mb-2 flex items-center gap-2">
                        <Activity size={16} className="text-emerald-400"/> Tăng trưởng nhân sự
                    </h3>
                    <div className="flex-1">
                        <ResponsiveContainer width="100%" height="100%">
                            <BarChart data={growthData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                                <XAxis dataKey="name" stroke="#475569" fontSize={12} tickLine={false} axisLine={false} />
                                <YAxis stroke="#475569" fontSize={12} tickLine={false} axisLine={false} />
                                <Tooltip cursor={{ fill: '#334155', opacity: 0.4 }} contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '12px' }} />
                                <Bar dataKey="users" fill="#10b981" radius={[4, 4, 0, 0]} />
                            </BarChart>
                        </ResponsiveContainer>
                    </div>
                </div>
            </div>

            {/* --- KHU VỰC THÊM & DANH SÁCH --- */}
            <div className="grid grid-cols-1 xl:grid-cols-3 gap-8">
                {/* Form Tạo User */}
                <div className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl h-fit">
                    <h3 className="text-lg font-bold mb-6 flex items-center gap-2">
                        <UserPlus className="text-emerald-400" size={20}/> Cấp tài khoản mới
                    </h3>
                    <form onSubmit={handleCreateSubmit} className="space-y-4">
                        {/* 2. CÁC TRƯỜNG THÔNG TIN MỚI */}
                        <div>
                            <label className="text-xs font-bold text-slate-500 uppercase">Họ và tên</label>
                            <input type="text" value={formData.full_name} required placeholder="VD: Nguyễn Văn A" onChange={(e) => setFormData({...formData, full_name: e.target.value})} className="w-full mt-2 p-3 bg-slate-900 border border-slate-700 rounded-xl outline-none focus:border-emerald-500 text-white" />
                        </div>
                        
                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="text-xs font-bold text-slate-500 uppercase">Tên đăng nhập</label>
                                <input type="text" value={formData.username} required placeholder="VD: nguyenva" onChange={(e) => setFormData({...formData, username: e.target.value})} className="w-full mt-2 p-3 bg-slate-900 border border-slate-700 rounded-xl outline-none focus:border-emerald-500 text-white" />
                            </div>
                            <div>
                                <label className="text-xs font-bold text-slate-500 uppercase">Mật khẩu</label>
                                <input type="password" value={formData.password} required minLength="6" placeholder="******" onChange={(e) => setFormData({...formData, password: e.target.value})} className="w-full mt-2 p-3 bg-slate-900 border border-slate-700 rounded-xl outline-none focus:border-emerald-500 text-white" />
                            </div>
                        </div>

                        <div>
                            <label className="text-xs font-bold text-slate-500 uppercase">Số điện thoại</label>
                            <input type="tel" value={formData.phone} placeholder="VD: 0912345678" onChange={(e) => setFormData({...formData, phone: e.target.value})} className="w-full mt-2 p-3 bg-slate-900 border border-slate-700 rounded-xl outline-none focus:border-emerald-500 text-white" />
                        </div>

                        <div>
                            <label className="text-xs font-bold text-slate-500 uppercase">Email</label>
                            <input type="email" value={formData.email} placeholder="VD: email@congty.com" onChange={(e) => setFormData({...formData, email: e.target.value})} className="w-full mt-2 p-3 bg-slate-900 border border-slate-700 rounded-xl outline-none focus:border-emerald-500 text-white" />
                        </div>

                        <div>
                            <label className="text-xs font-bold text-slate-500 uppercase">Phân quyền</label>
                            <select value={formData.role_level} onChange={(e) => setFormData({...formData, role_level: parseInt(e.target.value)})} className="w-full mt-2 p-3 bg-slate-900 border border-slate-700 rounded-xl outline-none focus:border-emerald-500 text-white">
                                <option value={1}>Level 1 - Nhân viên</option>
                                <option value={2}>Level 2 - Admin</option>
                            </select>
                        </div>
                        <button type="submit" disabled={isLoading} className="w-full mt-4 bg-emerald-500 hover:bg-emerald-600 text-white font-bold py-3 rounded-xl transition shadow-lg shadow-emerald-500/20 disabled:opacity-50">
                            {isLoading ? "Đang xử lý..." : "Tạo tài khoản"}
                        </button>
                    </form>
                </div>

                {/* Danh sách & Phân trang */}
                <div className="xl:col-span-2 flex flex-col h-full">
                    <div className="mb-4 relative">
                        <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500" size={20}/>
                        <input 
                            type="text" 
                            placeholder="Tìm kiếm theo tên đăng nhập hoặc họ tên..." 
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="w-full pl-12 pr-4 py-3 bg-[#1e293b] border border-slate-800 rounded-2xl text-white outline-none focus:border-emerald-500 transition shadow-lg"
                        />
                    </div>

                    <div className="bg-[#1e293b] rounded-3xl border border-slate-800 overflow-hidden shadow-xl flex-1 flex flex-col">
                        <div className="p-5 border-b border-slate-800 bg-slate-800/30 flex items-center gap-3">
                            <Users className="text-emerald-400" size={20}/>
                            <h3 className="font-bold text-white">Danh sách nhân sự</h3>
                        </div>
                        
                        <div className="overflow-x-auto flex-1">
                            <table className="w-full text-left">
                                <thead className="text-[10px] uppercase tracking-widest text-slate-500 bg-slate-900/50">
                                    <tr>
                                        <th className="p-5 font-bold">Người dùng</th>
                                        <th className="p-5 font-bold">Liên hệ</th>
                                        <th className="p-5 font-bold">Quyền</th>
                                        <th className="p-5 font-bold text-right">Thao tác</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-slate-800">
                                    {currentUsers.map((u) => (
                                        <tr key={u.id} className="hover:bg-slate-800/50 transition-colors group">
                                            {/* 3. CẬP NHẬT CỘT NGƯỜI DÙNG: Hiển thị Full Name to, Username nhỏ */}
                                            <td className="p-5 flex items-center gap-3">
                                                <div className="p-2 bg-slate-900 rounded-lg text-slate-400 group-hover:text-emerald-400 transition"><User size={16}/></div>
                                                <div>
                                                    <p className="font-bold text-white text-sm">{u.full_name || u.username}</p>
                                                    <p className="text-[10px] text-slate-500 font-mono">@{u.username}</p>
                                                </div>
                                            </td>
                                            {/* CỘT MỚI: Liên hệ */}
                                            <td className="p-5">
                                                <div className="text-xs text-slate-400 space-y-1">
                                                    <div className="flex items-center gap-1.5"><Phone size={12}/> {u.phone || '—'}</div>
                                                    <div className="flex items-center gap-1.5"><Mail size={12}/> {u.email || '—'}</div>
                                                </div>
                                            </td>
                                            <td className="p-5">
                                                {u.role_level === 2 ? (
                                                    <span className="flex items-center gap-1.5 w-fit px-3 py-1 bg-amber-500/10 text-amber-400 border border-amber-500/20 rounded-full text-[10px] font-black uppercase"><Shield size={12}/> Admin</span>
                                                ) : (
                                                    <span className="flex items-center gap-1.5 w-fit px-3 py-1 bg-blue-500/10 text-blue-400 border border-blue-500/20 rounded-full text-[10px] font-black uppercase"><User size={12}/> Nhân viên</span>
                                                )}
                                            </td>
                                            <td className="p-5 text-right">
                                                <button onClick={() => deleteUser(u.id)} className="p-2 text-slate-500 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition" title="Xóa"><Trash2 size={16}/></button>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                            {filteredUsers.length === 0 && (
                                <div className="p-10 text-center text-slate-500 italic">Không tìm thấy tài khoản nào.</div>
                            )}
                        </div>

                        {/* Phân trang */}
                        {totalPages > 1 && (
                            <div className="p-4 border-t border-slate-800 flex justify-between items-center bg-slate-900/30">
                                <span className="text-xs text-slate-500 font-medium">
                                    Hiển thị {indexOfFirstItem + 1}-{Math.min(indexOfLastItem, filteredUsers.length)} / {filteredUsers.length}
                                </span>
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