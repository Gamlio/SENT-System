import React, { useState, useEffect } from 'react';
import { Package, Search, ChevronLeft, ChevronRight, Layers } from 'lucide-react';

const AgentSoftware = ({ software }) => {
    const [search, setSearch] = useState('');
    const [page, setPage] = useState(1);
    const itemsPerPage = 8; // Số lượng hiển thị mỗi trang

    // Reset về trang 1 khi tìm kiếm
    useEffect(() => setPage(1), [search]);

    const filtered = (software || []).filter(s => 
        s.software_name?.toLowerCase().includes(search.toLowerCase())
    );

    const totalPages = Math.ceil(filtered.length / itemsPerPage);
    const displayedItems = filtered.slice((page - 1) * itemsPerPage, page * itemsPerPage);

    return (
        <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-xl overflow-hidden flex flex-col h-[520px]">
            {/* Header */}
            <div className="p-4 border-b border-slate-800 bg-slate-800/20 flex justify-between items-center">
                <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 tracking-widest">
                    <Package size={14} className="text-blue-400"/> Phần mềm ({filtered.length})
                </h3>
                <div className="relative w-32">
                    <Search className="absolute left-2 top-1/2 -translate-y-1/2 text-slate-500" size={12}/>
                    <input 
                        type="text" placeholder="Tìm kiếm..." 
                        value={search} onChange={e => setSearch(e.target.value)}
                        className="w-full pl-7 pr-2 py-1.5 bg-slate-900 border border-slate-700 rounded-lg text-[10px] text-white outline-none focus:border-blue-500 transition"
                    />
                </div>
            </div>
            
            {/* List */}
            <div className="flex-1 overflow-y-auto p-3 space-y-2 custom-scrollbar">
                {displayedItems.length === 0 ? (
                    <div className="h-full flex flex-col items-center justify-center text-slate-500 opacity-50">
                        <Package size={40} strokeWidth={1}/>
                        <span className="text-xs mt-2">Không tìm thấy phần mềm</span>
                    </div>
                ) : (
                    displayedItems.map((sw, i) => (
                        <div key={i} className="p-3 rounded-xl border border-slate-800/50 bg-slate-900/30 flex justify-between items-center hover:bg-slate-800 hover:border-blue-500/30 transition group">
                            <div className="flex flex-col overflow-hidden w-full">
                                <span className="text-[11px] font-bold text-slate-300 truncate group-hover:text-blue-400 transition-colors" title={sw.software_name}>
                                    {sw.software_name}
                                </span>
                                <span className="text-[9px] text-slate-600 font-mono flex items-center gap-1">
                                    <Layers size={8}/> v{sw.version || 'Unknown'}
                                </span>
                            </div>
                        </div>
                    ))
                )}
            </div>

            {/* Pagination Bar (Giao diện đẹp) */}
            {totalPages > 1 && (
                <div className="p-3 border-t border-slate-800 bg-slate-900/50 flex justify-between items-center backdrop-blur-sm">
                    <span className="text-[9px] text-slate-500 font-bold uppercase tracking-wider">
                        Trang <span className="text-white">{page}</span> / {totalPages}
                    </span>
                    <div className="flex gap-1.5">
                        <button 
                            disabled={page === 1} 
                            onClick={() => setPage(p => p - 1)} 
                            className="p-1.5 rounded-lg bg-slate-800 text-slate-400 border border-slate-700 hover:bg-blue-600 hover:text-white hover:border-blue-500 disabled:opacity-30 disabled:cursor-not-allowed transition-all"
                        >
                            <ChevronLeft size={14}/>
                        </button>
                        <button 
                            disabled={page === totalPages} 
                            onClick={() => setPage(p => p + 1)} 
                            className="p-1.5 rounded-lg bg-slate-800 text-slate-400 border border-slate-700 hover:bg-blue-600 hover:text-white hover:border-blue-500 disabled:opacity-30 disabled:cursor-not-allowed transition-all"
                        >
                            <ChevronRight size={14}/>
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default AgentSoftware;