import React, { useState, useMemo, useEffect } from 'react';
import { ShieldAlert, Search, AlertTriangle, CheckCircle } from 'lucide-react';
import Pagination from '../../../components/common/Pagination';

const IncidentTable = ({ rawData, onViewDetail }) => {
    // State nội bộ cho UI (Search, Filter, Page)
    const [searchTerm, setSearchTerm] = useState('');
    const [filterStatus, setFilterStatus] = useState('ALL'); 
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 10;

    // Reset về trang 1 khi filter thay đổi
    useEffect(() => setCurrentPage(1), [searchTerm, filterStatus]);

    // Logic lọc dữ liệu (Client-side)
    const filteredData = useMemo(() => {
        return rawData.filter(item => {
            if (filterStatus !== 'ALL' && item.status !== filterStatus) return false;
            if (searchTerm) {
                const term = searchTerm.toLowerCase();
                return item.agent?.hostname?.toLowerCase().includes(term) || 
                       item.agent_hwid?.toLowerCase().includes(term) ||
                       item.type?.toLowerCase().includes(term);
            }
            return true;
        });
    }, [rawData, searchTerm, filterStatus]);

    // Logic cắt trang
    const totalPages = Math.ceil(filteredData.length / itemsPerPage);
    const currentData = filteredData.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);

    return (
        <div className="flex flex-col h-full">
            {/* 1. TOOLBAR */}
            <div className="flex flex-col md:flex-row justify-between items-end mb-6 gap-4">
                <div>
                    <h1 className="text-3xl font-black text-white flex items-center gap-3">
                        <ShieldAlert className="text-red-500" size={32} />
                        Trung tâm Ứng phó Sự cố (SOC)
                    </h1>
                    <p className="text-slate-400 text-sm mt-1">Quản lý và xử lý các mối đe dọa an ninh từ máy trạm.</p>
                </div>
                <div className="flex gap-2">
                    {/* Bộ lọc */}
                    <div className="flex bg-slate-900 rounded-xl p-1 border border-slate-700">
                        {['ALL', 'Open', 'Resolved'].map(status => (
                            <button 
                                key={status}
                                onClick={() => setFilterStatus(status)}
                                className={`px-3 py-1.5 rounded-lg text-xs font-bold transition ${filterStatus === status ? 'bg-blue-600 text-white shadow' : 'text-slate-400 hover:text-white'}`}
                            >
                                {status === 'ALL' ? 'Tất cả' : status === 'Open' ? 'Đang mở' : 'Đã xong'}
                            </button>
                        ))}
                    </div>
                    {/* Tìm kiếm */}
                    <div className="relative">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" size={16} />
                        <input 
                            type="text" 
                            onChange={(e) => setSearchTerm(e.target.value)}
                            placeholder="Tìm máy, lỗi..." 
                            className="w-64 bg-slate-900 border border-slate-700 rounded-xl pl-10 pr-4 py-2 text-sm text-white focus:border-blue-500 outline-none"
                        />
                    </div>
                </div>
            </div>

            {/* 2. TABLE */}
            <div className="bg-slate-800 rounded-2xl border border-slate-700 shadow-xl flex-1 flex flex-col overflow-hidden">
                <div className="overflow-auto flex-1">
                    <table className="w-full text-left border-collapse">
                        <thead className="bg-slate-900/80 sticky top-0 z-10 backdrop-blur-sm">
                            <tr className="text-slate-400 text-xs uppercase border-b border-slate-700">
                                <th className="p-4 w-20">Case ID</th>
                                <th className="p-4">Máy Trạm</th>
                                <th className="p-4">Loại Sự Cố (Gom nhóm)</th>
                                <th className="p-4 text-center">Mức độ</th>
                                <th className="p-4 text-center">Trạng thái</th>
                                <th className="p-4">Cập nhật cuối</th>
                                <th className="p-4 text-center">Hành động</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-700/50 text-sm">
                            {currentData.map((inc) => (
                                <tr key={inc.id} className="hover:bg-slate-700/40 transition cursor-pointer group" onClick={() => onViewDetail(inc.id)}>
                                    <td className="p-4 font-mono text-slate-500">#{inc.id}</td>
                                    <td className="p-4">
                                        <div className="font-bold text-white group-hover:text-blue-400 transition">{inc.agent?.hostname || 'N/A'}</div>
                                        <div className="text-[10px] font-mono text-slate-500">{inc.agent_hwid}</div>
                                    </td>
                                    <td className="p-4">
                                        <div className="font-semibold text-slate-200">{inc.type}</div>
                                        <div className="text-xs text-slate-500 truncate max-w-[200px]">{inc.description}</div>
                                    </td>
                                    <td className="p-4 text-center">
                                        <span className={`px-2 py-1 rounded text-[10px] font-black uppercase border ${inc.severity === 'High' ? 'bg-red-500/10 text-red-500 border-red-500/20' : 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20'}`}>
                                            {inc.severity}
                                        </span>
                                    </td>
                                    <td className="p-4 text-center">
                                        {inc.status === 'Open' ? 
                                            <span className="text-red-400 font-bold text-xs flex justify-center items-center gap-1"><AlertTriangle size={12}/> Mở</span> : 
                                            <span className="text-emerald-400 font-bold text-xs flex justify-center items-center gap-1"><CheckCircle size={12}/> Xong</span>
                                        }
                                    </td>
                                    <td className="p-4 text-xs font-mono text-slate-400">
                                        {new Date(inc.updated_at).toLocaleTimeString()} {new Date(inc.updated_at).toLocaleDateString()}
                                    </td>
                                    <td className="p-4 text-center">
                                        <button className="text-blue-400 hover:text-white text-xs font-bold border border-blue-500/30 px-3 py-1 rounded bg-blue-500/10 hover:bg-blue-600 transition">
                                            Xem
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
                {/* 3. PAGINATION */}
                <div className="p-2 border-t border-slate-700 bg-slate-900/50">
                    <Pagination currentPage={currentPage} totalPages={totalPages} onPageChange={setCurrentPage} />
                </div>
            </div>
        </div>
    );
};

export default IncidentTable;