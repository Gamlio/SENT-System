import React, { useState, useEffect, useCallback } from 'react';
import { Usb, Search, ChevronLeft, ChevronRight, Fingerprint, Tag } from 'lucide-react';
import axios from '../../../api/axios';

const assetUSB = ({ hwid }) => {
    const [search, setSearch] = useState('');
    const [page, setPage] = useState(1);
    const [data, setData] = useState([]);
    const [totalItems, setTotalItems] = useState(0);
    const itemsPerPage = 6; 

    useEffect(() => setPage(1), [search]);

    const fetchData = useCallback(async () => {
        if (!hwid) return;
        try {
            const res = await axios.get(`/assets/${hwid}/usb`, {
                params: { page, limit: itemsPerPage, search }
            });
            setData(res.data.items || []);
            setTotalItems(res.data.total || 0);
        } catch (err) {
            console.error(err);
        }
    }, [hwid, page, search]);

    useEffect(() => {
        const timer = setTimeout(() => fetchData(), 300);
        return () => clearTimeout(timer);
    }, [fetchData]);

    const totalPages = Math.ceil(totalItems / itemsPerPage) || 1;

    return (
        <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-xl overflow-hidden flex flex-col h-[520px]">
            <div className="p-4 border-b border-slate-800 bg-slate-800/20 flex justify-between items-center">
                <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 tracking-widest">
                    <Usb size={14} className="text-emerald-400"/> Lịch sử USB ({totalItems})
                </h3>
                <div className="relative w-36">
                    <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                    <input 
                        type="text" placeholder="Tìm tên, VID..." 
                        value={search} onChange={(e) => setSearch(e.target.value)}
                        className="w-full bg-slate-900/50 border border-slate-700 text-xs text-white rounded-lg pl-8 pr-3 py-1.5 outline-none focus:border-emerald-500 transition-colors"
                    />
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-4 space-y-3 custom-scrollbar">
                {data.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-full text-slate-500">
                        <Usb size={32} className="mb-2 opacity-20"/>
                        <p className="text-xs font-bold">Chưa ghi nhận USB nào</p>
                    </div>
                ) : (
                    data.map((usb, idx) => (
                        <div key={idx} className="p-3 bg-slate-900/50 rounded-xl border border-slate-800 hover:border-emerald-500/50 transition-colors group">
                            <div className="flex gap-3 items-center mb-2">
                                <div className="p-2 bg-emerald-500/10 text-emerald-400 rounded-lg shrink-0">
                                    <Usb size={16} />
                                </div>
                                <div className="flex-1 min-w-0">
                                    <h4 className="text-sm font-bold text-slate-200 truncate" title={usb.device_name}>{usb.device_name || 'Unknown USB Device'}</h4>
                                </div>
                            </div>
                            
                            <div className="grid grid-cols-2 gap-2 mt-2 pt-2 border-t border-slate-800/50">
                                <div className="flex items-center gap-1.5 text-[10px] text-slate-400">
                                    <Tag size={12} className="text-slate-500"/>
                                    <span className="font-mono bg-black/30 px-1.5 py-0.5 rounded border border-slate-700">VID: {usb.vid || 'N/A'}</span>
                                    <span className="font-mono bg-black/30 px-1.5 py-0.5 rounded border border-slate-700">PID: {usb.pid || 'N/A'}</span>
                                </div>
                                <div className="flex items-center gap-1.5 text-[10px] text-slate-400">
                                    <Fingerprint size={12} className="text-slate-500"/>
                                    <span className="font-mono truncate" title={usb.device_hash}>{usb.device_hash?.substring(0,10)}...</span>
                                </div>
                            </div>
                        </div>
                    ))
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

export default assetUSB;