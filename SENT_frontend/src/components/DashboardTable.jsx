import React, { useState, useEffect } from 'react';
import { Search, ChevronLeft, ChevronRight, Monitor, Clock } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

const DashboardTable = ({ title, icon: Icon, data, colorClass, statusType }) => {
    const navigate = useNavigate();
    const [searchTerm, setSearchTerm] = useState("");
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 5;

    useEffect(() => setCurrentPage(1), [searchTerm]);

    const filteredData = data.filter(item => 
        item.hostname?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        item.ip_address?.includes(searchTerm)
    );

    const totalPages = Math.ceil(filteredData.length / itemsPerPage);
    const displayedData = filteredData.slice(
        (currentPage - 1) * itemsPerPage, 
        currentPage * itemsPerPage
    );

    const getTimeAgo = (dateString) => {
        if (!dateString) return "N/A";
        const diff = Math.floor((new Date() - new Date(dateString)) / 1000);
        if (diff < 60) return `${diff}s trước`;
        if (diff < 3600) return `${Math.floor(diff/60)}p trước`;
        if (diff < 86400) return `${Math.floor(diff/3600)}h trước`;
        return `${Math.floor(diff/86400)} ngày`;
    };

    return (
        <div className={`bg-[#1e293b] rounded-3xl border ${colorClass} overflow-hidden shadow-xl flex flex-col h-full`}>
            {/* Header + Tìm kiếm */}
            <div className={`p-5 border-b ${colorClass.replace('border-', 'border-opacity-30 border-')} flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 bg-slate-800/20`}>
                <h3 className="font-bold text-lg text-white flex items-center gap-2 whitespace-nowrap">
                    <Icon size={20} className={statusType === 'offline' ? 'text-red-400' : 'text-emerald-400'}/> 
                    {title} <span className="text-xs bg-slate-800 px-2 py-0.5 rounded-md text-slate-400 border border-slate-700">{filteredData.length}</span>
                </h3>
                
                <div className="relative w-full">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" size={14}/>
                    <input 
                        type="text" 
                        placeholder="Tìm kiếm..." 
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        className="w-full pl-9 pr-3 py-1.5 bg-slate-900 border border-slate-700 rounded-xl text-xs text-white outline-none focus:border-emerald-500 transition"
                    />
                </div>
            </div>

            {/* Bảng dữ liệu - ĐÃ SỬA CSS TẠI ĐÂY */}
            <div className="flex-1 overflow-hidden relative">
                <table className="w-full text-left table-fixed"> {/* table-fixed: Ép bảng không bị phình to */}
                    <thead className="text-[10px] uppercase tracking-widest text-slate-500 bg-slate-900/50">
                        <tr>
                            <th className="p-4 pl-6 w-[50%]">Hostname</th> {/* Cấp 50% chiều rộng */}
                            <th className="p-4 w-[25%]">IP</th>       {/* Cấp 25% chiều rộng */}
                            <th className="p-4 w-[25%] text-right pr-6">Time</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/50 text-sm">
                        {displayedData.length === 0 ? (
                            <tr><td colSpan="3" className="p-8 text-center text-slate-500 italic text-xs">Không có máy nào Online.</td></tr>
                        ) : (
                            displayedData.map((agent, idx) => (
                                <tr 
                                    key={agent.hwid || idx} 
                                    onClick={() => navigate(`/agents/${agent.hwid}`)}
                                    className="hover:bg-slate-800/60 cursor-pointer transition-colors group"
                                >
                                    <td className="p-4 pl-6 overflow-hidden">
                                        <div className="flex items-center gap-3">
                                            <div className={`p-2 rounded-lg shrink-0 ${statusType === 'offline' ? 'bg-red-500/10 text-red-400' : 'bg-emerald-500/10 text-emerald-400'}`}>
                                                <Monitor size={16}/>
                                            </div>
                                            <div className="min-w-0"> {/* min-w-0 giúp truncate hoạt động trong flex */}
                                                <div className="font-bold text-slate-200 group-hover:text-white truncate" title={agent.hostname}>
                                                    {agent.hostname}
                                                </div>
                                                <div className="text-[10px] text-slate-500 font-mono truncate">
                                                    {agent.hwid}
                                                </div>
                                            </div>
                                        </div>
                                    </td>
                                    <td className="p-4 font-mono text-slate-400 text-xs truncate" title={agent.ip_address}>
                                        {agent.ip_address || "---"}
                                    </td>
                                    <td className="p-4 text-xs text-slate-500 text-right pr-6 whitespace-nowrap">
                                        <div className="flex items-center justify-end gap-1.5">
                                            <Clock size={12}/> {getTimeAgo(agent.last_seen)}
                                        </div>
                                    </td>
                                </tr>
                            ))
                        )}
                    </tbody>
                </table>
            </div>

            {/* Phân trang */}
            {totalPages > 1 && (
                <div className="p-3 bg-slate-900/40 border-t border-slate-800 flex justify-between items-center mt-auto">
                    <span className="text-[10px] text-slate-500 font-bold uppercase">Trang {currentPage}/{totalPages}</span>
                    <div className="flex gap-1">
                        <button disabled={currentPage === 1} onClick={() => setCurrentPage(p => p - 1)} className="p-1.5 rounded bg-slate-800 text-slate-400 hover:text-white disabled:opacity-20"><ChevronLeft size={14}/></button>
                        <button disabled={currentPage === totalPages} onClick={() => setCurrentPage(p => p + 1)} className="p-1.5 rounded bg-slate-800 text-slate-400 hover:text-white disabled:opacity-20"><ChevronRight size={14}/></button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default DashboardTable;