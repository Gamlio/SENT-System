import React, { useState } from 'react';
import { Check, X, Laptop, Shield, FileText, UserPlus, Trash2, Info, Edit } from 'lucide-react';
import ApprovalTicketModal from './ApprovalTicketModal';

const ApprovalList = ({ tickets, onReview, loading }) => {
    const [selectedTicket, setSelectedTicket] = useState(null);

    const getModuleConfig = (type) => {
        switch (type) {
            case 'ASSET_ENROLL': return { label: 'Máy mới', icon: <Laptop size={16} className="text-blue-400"/>, requester: 'Thiết bị (asset)' };
            case 'ASSET_DELETE': return { label: 'Gỡ máy', icon: <Trash2 size={16} className="text-red-400"/>, requester: 'Quản trị viên' };
            case 'ASSET_BULK_DELETE': return { label: 'Gỡ nhiều máy', icon: <Trash2 size={16} className="text-red-400"/>, requester: 'Quản trị viên' };
            case 'POLICY_CREATE': return { label: 'Chính sách', icon: <Shield size={16} className="text-emerald-400"/>, requester: 'SOC Admin' };
            case 'DOCUMENT_UPLOAD': return { label: 'Tài liệu', icon: <FileText size={16} className="text-purple-400"/>, requester: 'Nhân viên' };
            case 'USER_CREATE': return { label: 'Nhân sự', icon: <UserPlus size={16} className="text-orange-400"/>, requester: 'CISO' };
            case 'USER_UPDATE': return { label: 'Đổi quyền', icon: <Edit size={16} className="text-amber-400"/>, requester: 'Admin' };
            case 'USER_DELETE': return { label: 'Xóa Nhân sự', icon: <Trash2 size={16} className="text-red-400"/>, requester: 'Admin' };
            default: return { label: 'Khác', icon: <Info size={16} className="text-slate-400"/>, requester: 'Hệ thống' };
        }
    };

    const renderContent = (ticket) => {
        try {
            const data = JSON.parse(ticket.snapshot_data || '{}');
            if (ticket.module_type === 'USER_CREATE') return `Cấp tài khoản: ${data.full_name || data.FullName} (@${data.username || data.Username})`;
            if (ticket.module_type === 'USER_UPDATE') return `Đổi quyền cho: ${data.full_name || data.FullName} (@${data.username || data.Username})`;
            if (ticket.module_type === 'USER_DELETE') return `Yêu cầu Xóa tài khoản nhân sự`;
            if (ticket.module_type === 'ASSET_ENROLL') return `Yêu cầu gia nhập: ${data.hostname || 'N/A'} (${data.ip_address || 'N/A'})`;
            if (ticket.module_type === 'ASSET_DELETE' || ticket.module_type === 'ASSET_BULK_DELETE') return `Xóa tài sản: ${ticket.target_name}`;
            return ticket.target_name;
        } catch { return ticket.target_name; }
    };

    if (loading) return <div className="p-10 text-center text-slate-500 animate-pulse">Đang truy vấn dữ liệu...</div>;
    if (tickets.length === 0) return <div className="p-20 text-center text-slate-600 italic">Không có yêu cầu nào cần xử lý.</div>;

    return (
        <div className="overflow-x-auto custom-scrollbar flex-1 min-h-[450px] relative pb-20">
            <table className="w-full text-left border-collapse whitespace-nowrap">
                <thead className="sticky top-0 z-10 bg-[#111827]">
                    <tr className="border-b border-slate-800 text-[10px] uppercase tracking-widest text-slate-500">
                        <th className="p-3 font-black">Phân loại</th>
                        <th className="p-3 font-black">Chi tiết yêu cầu</th>
                        <th className="p-3 font-black">Nguồn yêu cầu</th>
                        <th className="p-3 text-center font-black">Trạng thái</th>
                        <th className="p-3 text-right font-black">Thao tác</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/50">
                    {tickets.map((t) => {
                        const config = getModuleConfig(t.module_type);
                        return (
                            <tr key={t.id} onClick={() => setSelectedTicket(t)} className="hover:bg-slate-800/30 transition-colors cursor-pointer group">
                                <td className="p-3">
                                    <div className="flex items-center gap-2">
                                        {config.icon}
                                        <span className="text-xs font-bold text-slate-300">{config.label}</span>
                                    </div>
                                </td>
                                <td className="p-3">
                                    <div className="text-xs text-white font-medium">{renderContent(t)}</div>
                                    <div className="text-[10px] text-slate-500 mt-1 uppercase tracking-tighter">ID: {t.id} | {new Date(t.created_at).toLocaleString('vi-VN')}</div>
                                </td>
                                <td className="p-3">
                                    <div className="flex items-center gap-2">
                                        <div className="w-1.5 h-1.5 rounded-full bg-slate-600"></div>
                                        <span className="text-xs text-slate-400 font-mono">{t.requested_by || config.requester}</span>
                                    </div>
                                </td>
                                <td className="p-3 text-center">
                                    <span className={`px-2 py-1 rounded-lg text-[9px] font-black uppercase ${
                                        t.status === 'APPROVED' ? 'bg-emerald-500/10 text-emerald-500' :
                                        t.status === 'REJECTED' ? 'bg-red-500/10 text-red-500' : 'bg-amber-500/10 text-amber-500'
                                    }`}>
                                        {t.status}
                                    </span>
                                </td>
                                <td className="p-3 text-right">
                                    {t.status === 'PENDING' ? (
                                        <div className="flex justify-end gap-2 opacity-0 group-hover:opacity-100 transition">
                                            <button onClick={(e) => { e.stopPropagation(); setSelectedTicket(t); }} className="px-3 py-1.5 bg-indigo-600/20 text-indigo-400 hover:bg-indigo-600 hover:text-white rounded-lg transition text-xs font-bold">Xem chi tiết & Duyệt</button>
                                        </div>
                                    ) : (
                                        <span className="text-[10px] text-slate-500 italic">Xử lý bởi: {t.reviewed_by}</span>
                                    )}
                                </td>
                            </tr>
                        );
                    })}
                </tbody>
            </table>
            
            {/* Modal */}
            <ApprovalTicketModal 
                ticket={selectedTicket} 
                onClose={() => setSelectedTicket(null)} 
                onReview={onReview} 
            />
        </div>
    );
};

export default ApprovalList;