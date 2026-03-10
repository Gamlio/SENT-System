import React, { useState, useEffect } from 'react';
import { ClipboardCheck, Shield, Laptop, FileText, Layers } from 'lucide-react';
import { useApprovals } from './hooks/useApprovals';
import ApprovalList from './components/ApprovalList';

const ApprovalCenter = () => {
    const { tickets, loading, fetchTickets, reviewTicket } = useApprovals();
    const [activeTab, setActiveTab] = useState(''); // '' = Tất cả

    useEffect(() => {
        fetchTickets(activeTab);
    }, [fetchTickets, activeTab]);

    const handleReview = async (id, status) => {
        const confirmMsg = status === 'APPROVED' ? 'Bạn xác nhận DUYỆT yêu cầu này?' : 'Bạn muốn TỪ CHỐI yêu cầu này?';
        if (!window.confirm(confirmMsg)) return;
        
        await reviewTicket(id, status);
    };

    const tabs = [
        { key: '', label: 'Tất cả chờ duyệt', icon: <Layers size={16}/> },
        { key: 'POLICY_CREATE', label: 'Chính sách', icon: <Shield size={16}/> },
        { key: 'AGENT_ENROLL', label: 'Máy trạm', icon: <Laptop size={16}/> },
        { key: 'DOCUMENT_UPLOAD', label: 'Tài liệu', icon: <FileText size={16}/> },
    ];

    return (
        <div className="p-6 h-full flex flex-col">
            <div className="mb-6">
                <h1 className="text-2xl font-black text-white flex items-center gap-3">
                    <ClipboardCheck className="text-amber-500" size={28}/> 
                    Trung tâm Phê duyệt (Approval Center)
                </h1>
                <p className="text-sm text-slate-400 mt-1">Nơi Quản trị viên (SOC Manager) phê duyệt các thay đổi bảo mật của hệ thống.</p>
            </div>

            {/* Filter Tabs */}
            <div className="flex gap-2 mb-6">
                {tabs.map(tab => (
                    <button
                        key={tab.key}
                        onClick={() => setActiveTab(tab.key)}
                        className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold transition-all ${
                            activeTab === tab.key 
                            ? 'bg-amber-500 text-white shadow-lg shadow-amber-500/20' 
                            : 'bg-[#1e293b] text-slate-400 hover:text-white hover:bg-slate-800 border border-slate-800'
                        }`}
                    >
                        {tab.icon} {tab.label}
                    </button>
                ))}
            </div>

            {/* Bảng dữ liệu */}
            <div className="flex-1 bg-[#1e293b] rounded-3xl border border-slate-800 shadow-xl overflow-hidden">
                <ApprovalList tickets={tickets} onReview={handleReview} loading={loading} />
            </div>
        </div>
    );
};

export default ApprovalCenter;