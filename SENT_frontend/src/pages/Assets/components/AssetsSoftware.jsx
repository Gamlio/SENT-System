import React, { useState, useEffect, useCallback } from 'react';
import { Package, Search, ChevronLeft, ChevronRight, Activity, AlertTriangle, Fingerprint } from 'lucide-react';
import axios from '../../../api/axios';

const assetSoftware = ({ hwid }) => {
    const [search, setSearch] = useState('');
    const [page, setPage] = useState(1);
    const [data, setData] = useState([]);
    const [totalItems, setTotalItems] = useState(0);
    const itemsPerPage = 6;

    useEffect(() => setPage(1), [search]);

    const fetchData = useCallback(async () => {
        if (!hwid) return;
        try {
            const res = await axios.get(`/assets/${hwid}/software`, {
                params: { page, limit: itemsPerPage, search }
            });
            setData(res.data.items || []);
            setTotalItems(res.data.total || 0);
        } catch (err) {
            console.error(err);
        }
    }, [hwid, page, search]);

    // Sử dụng debounce cho ô tìm kiếm
    useEffect(() => {
        const timer = setTimeout(() => fetchData(), 300);
        return () => clearTimeout(timer);
    }, [fetchData]);

    const totalPages = Math.ceil(totalItems / itemsPerPage) || 1;

    return (
        <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-xl overflow-hidden flex flex-col h-[520px]">
            <div className="p-4 border-b border-slate-800 bg-slate-800/20 flex justify-between items-center">
                <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 tracking-widest">
                    <Package size={14} className="text-blue-400"/> Phần mềm ({totalItems})
                </h3>
                <div className="relative w-40">
                    <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                    <input 
                        type="text" placeholder="Tìm tên, hãng..." 
                        value={search} onChange={(e) => setSearch(e.target.value)}
                        className="w-full bg-slate-900/50 border border-slate-700 text-xs text-white rounded-lg pl-8 pr-3 py-1.5 outline-none focus:border-blue-500 transition-colors"
                    />
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-4 space-y-3 custom-scrollbar">
                {data.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-full text-slate-500">
                        <Package size={32} className="mb-2 opacity-20"/>
                        <p className="text-xs font-bold">Không tìm thấy phần mềm</p>
                    </div>
                ) : (
                    data.map((s, idx) => (
                        <div key={idx} className={`p-3 rounded-xl border flex flex-col gap-2 transition group ${s.status === 'GHOST_REGISTRY' ? 'bg-red-500/5 border-red-500/20' : 'bg-slate-900/50 border-slate-800 hover:border-slate-600'}`}>
                            <div className="flex justify-between items-start">
                                <div className="flex items-center gap-2">
                                    <div className={`p-1.5 rounded-lg ${s.is_running ? 'bg-emerald-500/20 text-emerald-400' : 'bg-slate-800 text-slate-500'}`}>
                                        <Activity size={14} className={s.is_running ? 'animate-pulse' : ''}/>
                                    </div>
                                    <div>
                                        <p className="text-sm font-bold text-slate-200 leading-tight">{s.software_name || 'Unknown'}</p>
                                        <p className="text-[10px] text-slate-500">{s.publisher || 'Unknown Publisher'} • v{s.version}</p>
                                    </div>
                                </div>
                                {s.status === 'GHOST_REGISTRY' && (
                                    <span className="flex items-center gap-1 text-[9px] font-black text-red-400 bg-red-500/10 px-2 py-1 rounded border border-red-500/20">
                                        <AlertTriangle size={10}/> GHOST REGISTRY
                                    </span>
                                )}
                            </div>
                            
                            {s.file_hash && (
                                <div className="flex items-center gap-1.5 mt-1 text-[10px] text-slate-500 bg-black/20 p-1.5 rounded border border-slate-800/50">
                                    <Fingerprint size={12} className="text-indigo-400 shrink-0"/>
                                    <span className="font-mono truncate" title={s.file_hash}>SHA256: {s.file_hash}</span>
                                </div>
                            )}
                        </div>
                    ))
                )}
            </div>

            {totalPages > 1 && (
                <div className="p-3 border-t border-slate-800 bg-slate-900/50 flex justify-between items-center">
                    <span className="text-[9px] text-slate-500 font-bold uppercase">Trang <span className="text-white">{page}</span> / {totalPages}</span>
                    <div className="flex gap-1.5">
                        <button disabled={page === 1} onClick={() => setPage(p => p - 1)} className="p-1.5 rounded-lg bg-slate-800 text-slate-400 hover:text-white disabled:opacity-30"><ChevronLeft size={14}/></button>
                        <button disabled={page === totalPages} onClick={() => setPage(p => p + 1)} className="p-1.5 rounded-lg bg-slate-800 text-slate-400 hover:text-white disabled:opacity-30"><ChevronRight size={14}/></button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default assetSoftware;