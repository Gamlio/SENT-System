import React, { useState, useEffect, useCallback } from 'react';
import { Network, Activity, ChevronLeft, ChevronRight } from 'lucide-react';
import axios from '../../../api/axios';

const assetPort = ({ hwid }) => {
    const [search, setSearch] = useState('');
    const [page, setPage] = useState(1);
    const [data, setData] = useState([]);
    const [totalItems, setTotalItems] = useState(0);
    const itemsPerPage = 6;

    useEffect(() => setPage(1), [search]);

    const fetchData = useCallback(async () => {
        if (!hwid) return;
        try {
            const res = await axios.get(`/assets/${hwid}/ports`, {
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
        <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-xl overflow-hidden flex flex-col h-full min-h-[400px]">
            <div className="p-5 border-b border-slate-800 bg-slate-800/20 flex justify-between items-center">
                <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 tracking-widest">
                    <Network size={16} className="text-amber-400"/> Giám sát Cổng Mạng ({totalItems})
                </h3>
                <div className="relative w-40">
                    <input 
                        type="text" placeholder="Tìm tiến trình, cổng..." 
                        value={search} onChange={(e) => setSearch(e.target.value)}
                        className="w-full bg-slate-900/50 border border-slate-700 text-xs text-white rounded-lg px-3 py-1.5 outline-none focus:border-amber-500 transition-colors"
                    />
                </div>
            </div>

            <div className="flex-1 overflow-auto custom-scrollbar p-2">
                <table className="w-full text-left border-separate border-spacing-y-2">
                    <thead>
                        <tr className="text-[10px] font-black text-slate-500 uppercase tracking-tighter">
                            <th className="px-4 py-2">Tiến trình</th>
                            <th className="px-4 py-2">Cổng</th>
                            <th className="px-4 py-2">Giao thức</th>
                            <th className="px-4 py-2">Trạng thái</th>
                        </tr>
                    </thead>
                    <tbody>
                        {data.length === 0 ? (
                            <tr>
                                <td colSpan="4" className="text-center text-slate-500 p-4 text-xs font-bold">
                                    Chưa có dữ liệu cổng kết nối
                                </td>
                            </tr>
                        ) : (
                        data.map((p, i) => {
                            const isRisky = [22, 3389, 445, 135].includes(p.port);
                            return (
                                <tr key={i} className="bg-slate-900/40 hover:bg-slate-800/50 transition-colors group">
                                    <td className="px-4 py-3 rounded-l-xl">
                                        <div className="flex items-center gap-3">
                                            <div className={`p-1.5 rounded-lg ${isRisky ? 'bg-red-500/10 text-red-400' : 'bg-slate-800 text-slate-400'}`}>
                                                <Activity size={14}/>
                                            </div>
                                            <span className="text-xs font-bold text-slate-200">{p.process_name}</span>
                                        </div>
                                    </td>
                                    <td className="px-4 py-3">
                                        <span className={`font-mono text-xs font-black ${isRisky ? 'text-red-400' : 'text-emerald-400'}`}>
                                            :{p.port}
                                        </span>
                                    </td>
                                    <td className="px-4 py-3 text-[10px] font-bold text-slate-500 uppercase">TCP</td>
                                    <td className="px-4 py-3 rounded-r-xl">
                                        <span className="px-2 py-0.5 rounded text-[9px] font-black uppercase bg-emerald-500/10 text-emerald-500 border border-emerald-500/20">
                                            LISTENING
                                        </span>
                                    </td>
                                </tr>
                            );
                        })
                        )}
                    </tbody>
                </table>
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

export default assetPort;