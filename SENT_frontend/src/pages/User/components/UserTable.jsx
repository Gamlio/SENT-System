import React from 'react';
import { Shield, Phone, Mail, Edit, Trash2 } from 'lucide-react';

const UserTable = ({ users, onEdit, onDelete, canManageUsers }) => {
    return (
        <div className="bg-[#1e293b] rounded-2xl border border-slate-800 shadow-2xl overflow-hidden">
            <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse">
                    <thead>
                        <tr className="bg-slate-900/50 border-b border-slate-800 text-[10px] uppercase tracking-wider text-slate-400">
                            <th className="p-4 font-black">Nhân viên</th>
                            <th className="p-4 font-black">Liên hệ</th>
                            <th className="p-4 font-black">Vai trò</th>
                            <th className="p-4 font-black">Đặc quyền (Permissions)</th>
                            {canManageUsers && <th className="p-4 font-black text-right">Thao tác</th>}
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/50">
                        {users.map((u) => (
                            <tr key={u.id} className="hover:bg-slate-800/30 transition group">
                               <td className="p-4">
                                    <div className="flex items-center gap-3">
                                        <div className="w-10 h-10 rounded-full bg-gradient-to-br from-emerald-500/20 to-blue-500/20 border border-slate-700 flex items-center justify-center text-emerald-400 font-bold shrink-0">
                                            {u.full_name?.charAt(0) || 'U'}
                                        </div>
                                        <div>
                                            <p className="font-bold text-slate-200">{u.full_name}</p>
                                            <p className="text-xs text-slate-500 font-mono">
                                                @{u.username} {u.employee_id && <span className="text-emerald-400 ml-2">| ID: {u.employee_id}</span>}
                                            </p>
                                        </div>
                                    </div>
                                </td>
                                <td className="p-4">
                                    <div className="flex flex-col gap-1.5 text-xs text-slate-400">
                                        <div className="flex items-center gap-2"><Phone size={12} className="text-emerald-500"/> {u.phone || 'N/A'}</div>
                                        <div className="flex items-center gap-2"><Mail size={12} className="text-blue-500"/> {u.email || 'N/A'}</div>
                                    </div>
                                </td>
                                <td className="p-4">
                                    <span className={`px-2.5 py-1 rounded-lg text-[10px] font-black tracking-wider uppercase border ${
                                        u.role === 'ADMIN' ? 'bg-purple-500/10 text-purple-400 border-purple-500/30' : 'bg-blue-500/10 text-blue-400 border-blue-500/30'
                                    }`}>
                                        {u.role}
                                    </span>
                                </td>
                                <td className="p-4">
                                    <div className="flex flex-wrap gap-1.5">
                                        {u.can_manage_agents && <Badge text="AGENTS" color="emerald"/>}
                                        {u.can_manage_policies && <Badge text="POLICIES" color="orange"/>}
                                        {u.can_manage_incidents && <Badge text="INCIDENTS" color="red"/>}
                                        {(!u.can_manage_agents && !u.can_manage_policies && !u.can_manage_incidents) && <span className="text-xs text-slate-600">Quyền cơ bản</span>}
                                    </div>
                                </td>
                                {canManageUsers && (
                                    <td className="p-4 text-right">
                                        <div className="flex justify-end gap-2 opacity-0 group-hover:opacity-100 transition">
                                            <button onClick={() => onEdit(u)} className="p-2 bg-slate-800 text-blue-400 hover:text-white hover:bg-blue-600 rounded-lg transition"><Edit size={16}/></button>
                                            <button onClick={() => onDelete(u)} className="p-2 bg-slate-800 text-red-400 hover:text-white hover:bg-red-600 rounded-lg transition"><Trash2 size={16}/></button>
                                        </div>
                                    </td>
                                )}
                            </tr>
                        ))}
                    </tbody>
                </table>
                {users.length === 0 && <div className="p-10 text-center text-slate-500 italic">Không tìm thấy nhân sự nào.</div>}
            </div>
        </div>
    );
};

// Component phụ trợ
const Badge = ({ text, color }) => {
    const colorClasses = {
        emerald: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
        orange: "bg-orange-500/10 text-orange-400 border-orange-500/20",
        red: "bg-red-500/10 text-red-400 border-red-500/20",
    };
    return <span className={`px-2 py-0.5 rounded text-[9px] font-bold border ${colorClasses[color]} flex items-center gap-1`}><Shield size={10}/> {text}</span>;
};

export default UserTable;