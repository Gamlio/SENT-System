import React, { useState, useEffect } from 'react';
import { Usb, Search, ChevronLeft, ChevronRight, HardDrive } from 'lucide-react';

const AgentUSB = ({ usbLogs }) => {
    const [search, setSearch] = useState('');
    const [page, setPage] = useState(1);
    const itemsPerPage = 6; // Số lượng ít hơn do card USB to hơn

    useEffect(() => setPage(1), [search]);

    const filtered = (usbLogs || []).filter(u => 
        u.device_name?.toLowerCase().includes(search.toLowerCase()) || 
        u.device_id?.toLowerCase().includes(search.toLowerCase())
    );

    const totalPages = Math.ceil(filtered.length / itemsPerPage);
    const displayedItems = filtered.slice((page - 1) * itemsPerPage, page * itemsPerPage);

    return (
        <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-xl overflow-hidden flex flex-col h-[520px]">
            {/* Header */}
            <div className="p-4 border-b border-slate-800 bg-slate-800/20 flex justify-between items-center">
                <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 tracking-widest">
                    <Usb size={14} className="text-emerald-400"/> Thiết bị USB ({filtered.length})
                </h3>
                <div className="relative w-28">
                   <input 
                        type="text" placeholder="Tìm ID/Tên..." 
                        value={search} onChange={e => setSearch(e.target.value)}
                        className="w-full px-3 py-1.5 bg-slate-900 border border-slate-700 rounded-lg text-[10px] text-white outline-none focus:border-emerald-500 transition"
                    />
                    <Search className="absolute right-2 top-1/2 -translate-y-1/2 text-slate-600 pointer-events-none" size={10}/>
                </div>
            </div>
            
            {/* List */}
            <div className="flex-1 overflow-y-auto p-3 space-y-2 custom-scrollbar">
                {displayedItems.length === 0 ? (
                    <div className="h-full flex flex-col items-center justify-center text-slate-500 opacity-50">
                        <Usb size={40} strokeWidth={1}/>
                        <span className="text-xs mt-2">Chưa ghi nhận USB nào</span>
                    </div>
                ) : (
                    displayedItems.map((usb, i) => (
                        <div key={i} className="p-3 bg-slate-900/40 rounded-xl border border-slate-800/50 flex flex-col hover:bg-slate-800 hover:border-emerald-500/30 transition group">
                            <div className="flex items-center gap-2 mb-1">
                                <HardDrive size={12} className="text-emerald-500"/>
                                <span className="text-[11px] font-bold text-slate-300 truncate w-full group-hover:text-emerald-400 transition-colors" title={usb.device_name}>
                                    {usb.device_name}
                                </span>
                            </div>
                            <div className="bg-black/20 p-1.5 rounded-lg border border-white/5">
                                <span className="text-[9px] text-slate-500 font-mono break-all line-clamp-1" title={usb.device_id}>
                                    ID: {usb.device_id}
                                </span>
                            </div>
                        </div>
                    ))
                )}
            </div>

            {/* Pagination Bar */}
            {totalPages > 1 && (
                <div className="p-3 border-t border-slate-800 bg-slate-900/50 flex justify-between items-center backdrop-blur-sm">
                    <span className="text-[9px] text-slate-500 font-bold uppercase tracking-wider">
                        Trang <span className="text-white">{page}</span> / {totalPages}
                    </span>
                    <div className="flex gap-1.5">
                        <button disabled={page===1} onClick={()=>setPage(p=>p-1)} className="p-1.5 rounded-lg bg-slate-800 text-slate-400 border border-slate-700 hover:bg-emerald-600 hover:text-white hover:border-emerald-500 disabled:opacity-30 transition-all"><ChevronLeft size={14}/></button>
                        <button disabled={page===totalPages} onClick={()=>setPage(p=>p+1)} className="p-1.5 rounded-lg bg-slate-800 text-slate-400 border border-slate-700 hover:bg-emerald-600 hover:text-white hover:border-emerald-500 disabled:opacity-30 transition-all"><ChevronRight size={14}/></button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default AgentUSB;