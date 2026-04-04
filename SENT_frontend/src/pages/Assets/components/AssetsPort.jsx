import React, { useState, useEffect } from 'react';
import { Network, Activity } from 'lucide-react';

const assetPort = ({ portLogs }) => {
    const [search] = useState('');
    const [page, setPage] = useState(1);
    const itemsPerPage = 6;

    useEffect(() => setPage(1), [search]);

    const filtered = (portLogs || []).filter(p => 
        p.process_name?.toLowerCase().includes(search.toLowerCase()) || 
        p.port?.toString().includes(search)
    );

    const displayedItems = filtered.slice((page - 1) * itemsPerPage, page * itemsPerPage);

    return (
        <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-xl overflow-hidden flex flex-col h-full min-h-[400px]">
            <div className="p-5 border-b border-slate-800 bg-slate-800/20 flex justify-between items-center">
                <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 tracking-widest">
                    <Network size={16} className="text-amber-400"/> Giám sát Cổng Mạng ({filtered.length})
                </h3>
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
                        {displayedItems.map((p, i) => {
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
                        })}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

            
        
export default assetPort;