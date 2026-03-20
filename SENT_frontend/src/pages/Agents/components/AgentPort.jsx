import React, { useState, useEffect } from 'react';
import { Network, Search, ChevronLeft, ChevronRight, Activity } from 'lucide-react';

const AgentPort = ({ portLogs }) => {
    const [search, setSearch] = useState('');
    const [page, setPage] = useState(1);
    const itemsPerPage = 6;

    useEffect(() => setPage(1), [search]);

    const filtered = (portLogs || []).filter(p => 
        p.process_name?.toLowerCase().includes(search.toLowerCase()) || 
        p.port?.toString().includes(search)
    );

    const totalPages = Math.ceil(filtered.length / itemsPerPage);
    const displayedItems = filtered.slice((page - 1) * itemsPerPage, page * itemsPerPage);

    return (
        <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-xl overflow-hidden flex flex-col h-[520px]">
            <div className="p-4 border-b border-slate-800 bg-slate-800/20 flex justify-between items-center">
                <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 tracking-widest">
                    <Network size={14} className="text-amber-400"/> Cổng Mạng ({filtered.length})
                </h3>
                <div className="relative w-36">
                    <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                    <input 
                        type="text" placeholder="Tìm process, port..." 
                        value={search} onChange={(e) => setSearch(e.target.value)}
                        className="w-full bg-slate-900/50 border border-slate-700 text-xs text-white rounded-lg pl-8 pr-3 py-1.5 outline-none focus:border-amber-500 transition-colors"
                    />
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-4 space-y-3 custom-scrollbar">
                {displayedItems.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-full text-slate-500">
                        <Network size={32} className="mb-2 opacity-20"/>
                        <p className="text-xs font-bold">Không có cổng nào đang mở</p>
                    </div>
                ) : (
                    displayedItems.map((p, idx) => {
                        // Highlight màu đỏ nếu là port nhạy cảm (3389 RDP, 22 SSH)
                        const isRisky = p.port === 3389 || p.port === 22 || p.port === 445;
                        return (
                            <div key={idx} className={`p-3 rounded-xl border flex justify-between items-center transition group ${isRisky ? 'bg-red-500/10 border-red-500/30' : 'bg-slate-900/50 border-slate-800 hover:border-slate-600'}`}>
                                <div className="flex items-center gap-3">
                                    <div className={`p-1.5 rounded-lg ${isRisky ? 'bg-red-500/20 text-red-400' : 'bg-amber-500/20 text-amber-400'}`}>
                                        <Activity size={14} className="animate-pulse" />
                                    </div>
                                    <div>
                                        <p className="text-sm font-bold text-slate-200">{p.process_name || 'System'}</p>
                                        <p className="text-[10px] text-slate-500">Listening Port</p>
                                    </div>
                                </div>
                                <div className={`font-mono text-sm font-bold px-2.5 py-1 rounded border ${isRisky ? 'text-red-400 border-red-500/50 bg-red-500/20' : 'text-slate-300 border-slate-700 bg-slate-800'}`}>
                                    :{p.port}
                                </div>
                            </div>
                        );
                    })
                )}
            </div>

            {totalPages > 1 && (
                <div className="p-3 border-t border-slate-800 bg-slate-900/50 flex justify-between items-center">
                    <span className="text-[9px] text-slate-500 font-bold uppercase">Trang <span className="text-white">{page}</span> / {totalPages}</span>
                    <div className="flex gap-1.5">
                        <button disabled={page===1} onClick={()=>setPage(p=>p-1)} className="p-1.5 rounded-lg bg-slate-800 text-slate-400 hover:text-white disabled:opacity-30"><ChevronLeft size={14}/></button>
                        <button disabled={page===totalPages} onClick={()=>setPage(p=>p+1)} className="p-1.5 rounded-lg bg-slate-800 text-slate-400 hover:text-white disabled:opacity-30"><ChevronRight size={14}/></button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default AgentPort;