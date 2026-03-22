import React from 'react';
import { Shield, Phone, Mail, Edit, Trash2, Settings, AlertTriangle, FileText, Monitor } from 'lucide-react';

const UserTable = ({ users, onEdit, onDelete, canManageUsers }) => {
    return (
        <div className="bg-[#1e293b] rounded-2xl border border-slate-800 shadow-2xl overflow-hidden">
            <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse">
                    <thead>
                        <tr className="bg-slate-900/50 border-b border-slate-800 text-[10px] uppercase tracking-wider text-slate-400">
                            <th className="p-4 font-black">Nhân viên</th>
                            <th className="p-4 font-black">Liên hệ</th>
                            <th className="p-4 font-black">Máy trạm & Sự cố</th>
                            <th className="p-4 font-black">Chính sách & Tài liệu</th>
                            <th className="p-4 font-black">Quản trị Hệ thống</th>
                            {canManageUsers && <th className="p-4 font-black text-right">Thao tác</th>}
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/50">
                        {users.map((u) => (
                            <tr key={u.id} className="hover:bg-slate-800/30 transition group">
                               {/* Cột 1: Thông tin nhân viên */}
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
                                
                                {/* Cột 2: Liên hệ */}
                                <td className="p-4">
                                    <div className="flex flex-col gap-1.5 text-xs text-slate-400">
                                        <div className="flex items-center gap-2"><Phone size={12} className="text-emerald-500"/> {u.phone || 'N/A'}</div>
                                        <div className="flex items-center gap-2"><Mail size={12} className="text-blue-500"/> {u.email || 'N/A'}</div>
                                    </div>
                                </td>

                                {/* Cột 3: Máy trạm & Sự cố */}
                                <td className="p-4">
                                    <div className="flex flex-col gap-2">
                                        <div className="flex flex-wrap gap-1.5 items-center">
                                            <Monitor size={12} className="text-slate-500 mr-1"/>
                                            {!!u.perm_agent_view && <Badge text="XEM MÁY" color="emerald"/>}
                                            {!!u.perm_agent_action && <Badge text="SỬA MÁY" color="emerald"/>}
                                            {!!u.perm_agent_delete && <Badge text="XÓA MÁY" color="red"/>}
                                            {!(u.perm_agent_view || u.perm_agent_action || u.perm_agent_delete) && <span className="text-[10px] text-slate-600 italic">Không có quyền</span>}
                                        </div>
                                        <div className="flex flex-wrap gap-1.5 items-center">
                                            <AlertTriangle size={12} className="text-slate-500 mr-1"/>
                                            {!!u.perm_incident_view && <Badge text="XEM SỰ CỐ" color="orange"/>}
                                            {!!u.perm_incident_action && <Badge text="XỬ LÝ" color="orange"/>}
                                            {!(u.perm_incident_view || u.perm_incident_action) && <span className="text-[10px] text-slate-600 italic">Không có quyền</span>}
                                        </div>
                                    </div>
                                </td>

                                {/* Cột 4: Chính sách & Tài liệu */}
                                <td className="p-4">
                                    <div className="flex flex-col gap-2">
                                        {/* Nhóm Chính sách */}
                                        <div className="flex flex-wrap gap-1.5 items-center">
                                            <Shield size={12} className="text-slate-500 mr-1"/>
                                            {!!u.perm_policy_view && <Badge text="XEM LUẬT" color="blue"/>}
                                            {!!u.perm_policy_action && <Badge text="ĐỔI LUẬT" color="blue"/>}
                                            {!(u.perm_policy_view || u.perm_policy_action) && <span className="text-[10px] text-slate-600 italic">Không có quyền luật</span>}
                                        </div>
                                        {/* Nhóm Tài liệu */}
                                        <div className="flex flex-wrap gap-1.5 items-center">
                                            <FileText size={12} className="text-slate-500 mr-1"/>
                                            {!!u.perm_doc_view && <Badge text="XEM TÀI LIỆU" color="blue"/>}
                                            {!!u.perm_doc_manage && <Badge text="UP TÀI LIỆU" color="blue"/>}
                                            {!(u.perm_doc_view || u.perm_doc_manage) && <span className="text-[10px] text-slate-600 italic">Không có quyền tài liệu</span>}
                                        </div>
                                    </div>
                                </td>

                                {/* Cột 5: Quản trị Hệ thống */}
                                <td className="p-4">
                                    <div className="flex flex-col gap-2">
                                        <div className="flex flex-wrap gap-1.5 items-center">
                                            <Settings size={12} className="text-slate-500 mr-1"/>
                                            {!!u.perm_user_manage && <Badge text="NHÂN SỰ" color="purple"/>}
                                            {!!u.perm_approval_manage && <Badge text="DUYỆT ĐƠN" color="purple"/>}
                                            {!(u.perm_user_manage || u.perm_approval_manage) && <span className="text-[10px] text-slate-600 italic">Không có quyền</span>}
                                        </div>
                                    </div>
                                </td>

                                {/* Cột 6: Thao tác */}
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

const Badge = ({ text, color }) => {
    const colorClasses = {
        emerald: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
        orange: "bg-orange-500/10 text-orange-400 border-orange-500/20",
        red: "bg-red-500/10 text-red-400 border-red-500/20",
        blue: "bg-blue-500/10 text-blue-400 border-blue-500/20",
        purple: "bg-purple-500/10 text-purple-400 border-purple-500/20",
    };
    return <span className={`px-2 py-0.5 rounded text-[9px] font-bold border ${colorClasses[color]} flex items-center gap-1`}><Shield size={10}/> {text}</span>;
};

export default UserTable;