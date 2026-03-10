import React from 'react';
import { Check, X, ShieldAlert, FileText, Laptop, Shield } from 'lucide-react';

const ApprovalList = ({ tickets, onReview, loading }) => {
    
    // Icon và Tên đẹp cho từng Module
    const getModuleInfo = (moduleType) => {
        switch (moduleType) {
            case 'AGENT_ENROLL': return { label: 'Máy trạm mới', icon: <Laptop size={16} className="text-blue-400"/>, bg: 'bg-blue-500/10' };
            case 'POLICY_CREATE': return { label: 'Tạo Chính sách', icon: <Shield size={16} className="text-emerald-400"/>, bg: 'bg-emerald-500/10' };
            case 'DOCUMENT_UPLOAD': return { label: 'Tài liệu AI', icon: <FileText size={16} className="text-purple-400"/>, bg: 'bg-purple-500/10' };
            default: return { label: 'Hệ thống', icon: <ShieldAlert size={16} className="text-amber-400"/>, bg: 'bg-amber-500/10' };
        }
    };

    if (loading) return <div className="p-10 text-center text-slate-400">Đang tải dữ liệu...</div>;
    if (tickets.length === 0) return <div className="p-10 text-center text-slate-500 italic">Hiện không có đơn nào cần phê duyệt.</div>;

    return (
        <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
                <thead className="bg-slate-900/50 text-[10px] uppercase text-slate-500 font-bold sticky top-0">
                    <tr>
                        <th className="p-4">Loại Yêu Cầu</th>
                        <th className="p-4">Đối tượng</th>
                        <th className="p-4">Người Yêu Cầu</th>
                        <th className="p-4">Thời gian</th>
                        <th className="p-4 text-right">Tác vụ</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-slate-800 text-sm text-slate-200">
                    {tickets.map(ticket => {
                        const modInfo = getModuleInfo(ticket.module_type);
                        return (
                            <tr key={ticket.id} className="hover:bg-slate-800/40 transition">
                                <td className="p-4">
                                    <span className={`inline-flex items-center gap-2 px-2.5 py-1 rounded-lg text-xs font-bold ${modInfo.bg}`}>
                                        {modInfo.icon} {modInfo.label}
                                    </span>
                                </td>
                                <td className="p-4">
                                    <p className="font-bold text-white">{ticket.target_name}</p>
                                    <code className="text-[10px] text-slate-400 bg-slate-900 px-1 rounded mt-1 line-clamp-1">{ticket.snapshot_data}</code>
                                </td>
                                <td className="p-4 text-xs">{ticket.requested_by}</td>
                                <td className="p-4 text-xs text-slate-400">{new Date(ticket.created_at).toLocaleString()}</td>
                                <td className="p-4 text-right">
                                    <div className="flex justify-end gap-2">
                                        <button 
                                            onClick={() => onReview(ticket.id, 'APPROVED')} 
                                            className="flex items-center gap-1 px-3 py-1.5 bg-emerald-500/10 text-emerald-500 hover:bg-emerald-500 hover:text-white rounded-lg text-xs font-bold transition border border-emerald-500/30"
                                        >
                                            <Check size={14}/> Duyệt
                                        </button>
                                        <button 
                                            onClick={() => onReview(ticket.id, 'REJECTED')} 
                                            className="flex items-center gap-1 px-3 py-1.5 bg-red-500/10 text-red-500 hover:bg-red-500 hover:text-white rounded-lg text-xs font-bold transition border border-red-500/30"
                                        >
                                            <X size={14}/> Từ chối
                                        </button>
                                    </div>
                                </td>
                            </tr>
                        );
                    })}
                </tbody>
            </table>
        </div>
    );
};

export default ApprovalList;