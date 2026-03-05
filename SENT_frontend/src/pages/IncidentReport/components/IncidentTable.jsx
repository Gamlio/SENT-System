import React, { useState, useMemo, useEffect } from 'react';
import { ShieldAlert, Search, AlertTriangle, CheckCircle, BookOpen } from 'lucide-react';
import Pagination from '../../../components/common/Pagination';

// Hàm helper để tô màu Badge theo Priority
const getPriorityColor = (priority) => {
    switch (priority) {
        case 'P1': return 'bg-purple-500/20 text-purple-400 border-purple-500/50 shadow-[0_0_10px_rgba(168,85,247,0.4)]';
        case 'P2': return 'bg-red-500/20 text-red-400 border-red-500/50';
        case 'P3': return 'bg-orange-500/20 text-orange-400 border-orange-500/50';
        default: return 'bg-slate-500/20 text-slate-400 border-slate-500/50';
    }
};

const IncidentTable = ({ rawData, onViewDetail }) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [filterStatus, setFilterStatus] = useState('ALL'); 
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 10;

    useEffect(() => setCurrentPage(1), [searchTerm, filterStatus]);

    const filteredData = useMemo(() => {
        return rawData.filter(item => {
            if (filterStatus !== 'ALL' && item.status !== filterStatus) return false;
            if (searchTerm) {
                const term = searchTerm.toLowerCase();
                return item.agent?.hostname?.toLowerCase().includes(term) || 
                       item.agent_hwid?.toLowerCase().includes(term) ||
                       item.type?.toLowerCase().includes(term) ||
                       item.playbook_name?.toLowerCase().includes(term);
            }
            return true;
        });
    }, [rawData, searchTerm, filterStatus]);

    const totalPages = Math.ceil(filteredData.length / itemsPerPage);
    const currentData = filteredData.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);

    return (
        <div className="flex flex-col h-full">
            <div className="flex flex-col md:flex-row justify-between items-end mb-6 gap-4">
                <div>
                    <h1 className="text-3xl font-black text-white flex items-center gap-3 tracking-tight">
                        <ShieldAlert className="text-red-500 drop-shadow-[0_0_8px_rgba(239,68,68,0.8)]" size={32} />
                        Trung tâm Ứng phó Sự cố (SOC)
                    </h1>
                    <p className="text-slate-400 text-sm mt-1">Giám sát và điều phối xử lý các mối đe dọa an ninh theo Playbook.</p>
                </div>
                <div className="flex gap-2">
                    <div className="flex bg-[#1e293b] rounded-xl p-1 border border-slate-700 shadow-lg">
                        {['ALL', 'Open', 'Resolved'].map(status => (
                            <button 
                                key={status}
                                onClick={() => setFilterStatus(status)}
                                className={`px-4 py-1.5 rounded-lg text-xs font-bold transition ${filterStatus === status ? 'bg-indigo-600 text-white shadow-md' : 'text-slate-400 hover:text-white'}`}
                            >
                                {status === 'ALL' ? 'Tất cả' : status === 'Open' ? 'Đang mở' : 'Đã xong'}
                            </button>
                        ))}
                    </div>
                    <div className="relative shadow-lg">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" size={16} />
                        <input 
                            type="text" 
                            onChange={(e) => setSearchTerm(e.target.value)}
                            placeholder="Tìm máy, lỗi, playbook..." 
                            className="w-64 bg-[#1e293b] border border-slate-700 rounded-xl pl-10 pr-4 py-2.5 text-sm text-white focus:border-indigo-500 outline-none transition"
                        />
                    </div>
                </div>
            </div>

            <div className="bg-[#1e293b] rounded-2xl border border-slate-700 shadow-2xl flex-1 flex flex-col overflow-hidden">
                <div className="overflow-auto flex-1">
                    <table className="w-full text-left border-collapse whitespace-nowrap">
                        <thead className="bg-slate-900/90 sticky top-0 z-10 backdrop-blur-md">
                            <tr className="text-slate-400 text-[10px] uppercase tracking-widest border-b border-slate-700">
                                <th className="p-4 w-20">Case ID</th>
                                <th className="p-4">Mức độ</th>
                                <th className="p-4">Sự cố & Playbook</th>
                                <th className="p-4">Máy Trạm (Endpoint)</th>
                                <th className="p-4 text-center">Trạng thái</th>
                                <th className="p-4">Thời gian</th>
                                <th className="p-4 text-center">Thao tác</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-800 text-sm">
                            {currentData.length === 0 ? (
                                <tr><td colSpan="7" className="p-8 text-center text-slate-500 italic">Không có sự cố nào.</td></tr>
                            ) : currentData.map((inc) => (
                                <tr key={inc.id} className="hover:bg-slate-800/60 transition cursor-pointer group" onClick={() => onViewDetail(inc.id)}>
                                    <td className="p-4 font-mono text-slate-500 font-bold">#{inc.id}</td>
                                    
                                    {/* CỘT MỨC ĐỘ (P1, P2, P3) */}
                                    <td className="p-4">
                                        <div className="flex flex-col items-start gap-1">
                                            <span className={`px-2.5 py-1 rounded text-[11px] font-black uppercase border ${getPriorityColor(inc.priority || 'P3')}`}>
                                                {inc.priority || 'P?'} - {inc.severity}
                                            </span>
                                        </div>
                                    </td>

                                    {/* CỘT PLAYBOOK & LOẠI SỰ CỐ */}
                                    <td className="p-4">
                                        <div className="font-bold text-white group-hover:text-indigo-400 transition mb-1">{inc.type}</div>
                                        <div className="flex items-center gap-1.5 text-xs text-indigo-400 bg-indigo-500/10 w-fit px-2 py-0.5 rounded border border-indigo-500/20 font-medium">
                                            <BookOpen size={12}/> {inc.playbook_name || 'General Response'}
                                        </div>
                                    </td>

                                    <td className="p-4">
                                        <div className="font-bold text-slate-200">{inc.agent?.hostname || 'Unknown Host'}</div>
                                        <div className="text-[10px] font-mono text-slate-500 mt-0.5">{inc.agent_hwid?.substring(0,20)}...</div>
                                    </td>
                                    
                                    <td className="p-4 text-center">
                                        {inc.status === 'Open' ? 
                                            <span className="text-red-400 font-bold text-xs flex justify-center items-center gap-1"><AlertTriangle size={14}/> OPEN</span> : 
                                            <span className="text-emerald-400 font-bold text-xs flex justify-center items-center gap-1"><CheckCircle size={14}/> RESOLVED</span>
                                        }
                                    </td>
                                    <td className="p-4 text-xs font-mono text-slate-400">
                                        {new Date(inc.created_at).toLocaleString('vi-VN')}
                                    </td>
                                    <td className="p-4 text-center">
                                        <button className="text-indigo-400 hover:text-white text-xs font-bold border border-indigo-500/30 px-4 py-1.5 rounded-lg bg-indigo-500/10 hover:bg-indigo-600 transition shadow-lg">
                                            Chi tiết
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
                <div className="p-3 border-t border-slate-700 bg-slate-900/50">
                    <Pagination currentPage={currentPage} totalPages={totalPages} onPageChange={setCurrentPage} />
                </div>
            </div>
        </div>
    );
};

export default IncidentTable;