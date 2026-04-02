import React, { useState, useEffect, useCallback } from 'react';
import { ClipboardCheck, Shield, Laptop, FileText, Layers, UserPlus, Trash2, Clock, CheckCircle2, XCircle,Edit } from 'lucide-react';
import { useApprovals } from './hooks/useApprovals';
import ApprovalList from './components/ApprovalList';
import { useSocketSubscription } from '../../context/useSocketSubscription';

const ApprovalCenter = () => {
    const { tickets, loading, fetchTickets, reviewTicket } = useApprovals();
    const [activeTab, setActiveTab] = useState(''); 
    const [activeStatus, setActiveStatus] = useState('PENDING');

    // Lắng nghe sự kiện từ WebSocket và tải lại dữ liệu
    const handleTicketUpdate = useCallback(() => {
        console.log('[WS] Nhận vé chờ duyệt mới hoặc thay đổi trạng thái! Đang tải lại...');
        fetchTickets(activeTab, activeStatus);
    }, [fetchTickets, activeTab, activeStatus]);
    useSocketSubscription(['NEW_APPROVAL_TICKET', 'TICKET_STATUS_UPDATED'], handleTicketUpdate);

    useEffect(() => {
        fetchTickets(activeTab, activeStatus);
    }, [fetchTickets, activeTab, activeStatus]);

    const handleReview = async (id, status) => {
        const res = await reviewTicket(id, status);
        if (res.success) fetchTickets(activeTab, activeStatus);
        else alert(res.error);
    };

    const modules = [
        { key: '', label: 'Tất cả', icon: <Layers size={16}/> },
        { key: 'AGENT_ENROLL', label: 'Máy mới', icon: <Laptop size={16}/> },
        { key: 'AGENT_DELETE', label: 'Gỡ bỏ máy', icon: <Trash2 size={16}/> },
        { key: 'POLICY_CREATE', label: 'Chính sách', icon: <Shield size={16}/> },
        { key: 'DOCUMENT_UPLOAD', label: 'Tài liệu', icon: <FileText size={16}/> },
        { key: 'USER_CREATE', label: 'Nhân sự', icon: <UserPlus size={16}/> },
        { key: 'USER_UPDATE', label: 'Đổi quyền', icon: <Edit size={16}/> },
        { key: 'USER_DELETE', label: 'Xóa Nhân sự', icon: <Trash2 size={16}/> },
    ];

    const statusFilters = [
        { key: 'PENDING', label: 'Chờ duyệt', icon: <Clock size={14}/>, color: 'bg-amber-500' },
        { key: 'APPROVED', label: 'Đã duyệt', icon: <CheckCircle2 size={14}/>, color: 'bg-emerald-500' },
        { key: 'REJECTED', label: 'Từ chối', icon: <XCircle size={14}/>, color: 'bg-red-500' },
        { key: 'ALL', label: 'Lịch sử', icon: <Layers size={14}/>, color: 'bg-slate-600' },
    ];

    return (
        <div className="p-6 text-slate-200 h-[calc(100vh-60px)] flex flex-col bg-[#050B14] font-sans">
            
            <div className="flex justify-between items-end mb-4 shrink-0">
                <div>
                    <h1 className="text-xl font-black text-white flex items-center gap-2 uppercase tracking-tight">
                        <ClipboardCheck className="text-indigo-500" size={28}/> TRUNG TÂM PHÊ DUYỆT
                    </h1>
                    <p className="text-[11px] text-slate-500 mt-1 uppercase tracking-widest font-bold">Quản lý và phê duyệt các yêu cầu hệ thống.</p>
                </div>
                
                {/* Bộ lọc trạng thái */}
                <div className="flex bg-slate-900/50 p-1 rounded-xl border border-slate-800">
                    {statusFilters.map(s => (
                        <button key={s.key} onClick={() => setActiveStatus(s.key)}
                            className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-[10px] font-bold transition-all ${
                                activeStatus === s.key ? `${s.color} text-white` : 'text-slate-500 hover:text-white'
                            }`}>
                            {s.icon} {s.label}
                        </button>
                    ))}
                </div>
            </div>

            {/* Bộ lọc Module */}
            <div className="flex flex-wrap gap-2 mb-4">
                {modules.map(tab => (
                    <button key={tab.key} onClick={() => setActiveTab(tab.key)}
                        className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold transition-all border ${
                            activeTab === tab.key ? 'bg-blue-500/10 border-blue-500 text-blue-400' : 'bg-[#1e293b] border-slate-800 text-slate-500'
                        }`}>
                        {tab.icon} {tab.label}
                    </button>
                ))}
            </div>

            <div className="flex-1 bg-[#0A101D] rounded-lg border border-slate-800 shadow-2xl flex flex-col overflow-hidden">
                <ApprovalList tickets={tickets} onReview={handleReview} loading={loading} />
            </div>
        </div>
    );
};

export default ApprovalCenter;