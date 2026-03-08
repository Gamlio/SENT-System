import React, { useState } from 'react';
import { ShieldAlert, Search, ChevronLeft, ChevronRight } from 'lucide-react';

const AgentLogs = ({ logs }) => {
    const [search, setSearch] = useState('');
    const [page, setPage] = useState(1);
    const perPage = 5; // Chỉ hiện 5 dòng mỗi trang cho gọn

    const filtered = logs.filter(l => 
        l.description.toLowerCase().includes(search.toLowerCase()) || 
        l.alert_type.toLowerCase().includes(search.toLowerCase())
    );
    const getAlertColor = (type) => {
        switch(type) {
            case 'Firewall Disabled': 
                return 'text-red-400 border-red-500/50 bg-red-500/10';     // P1 - Critical
            case 'Malware/AV Alert': 
                return 'text-orange-400 border-orange-500/50 bg-orange-500/10'; // P2 - High
            case 'Unauthorized Port': 
                return 'text-amber-400 border-amber-500/50 bg-amber-500/10';   // P2 - High
            case 'Unpatched OS': 
                return 'text-yellow-400 border-yellow-500/50 bg-yellow-500/10'; // P3 - Medium
            case 'Software Violation': 
                return 'text-blue-400 border-blue-500/50 bg-blue-500/10';     // P3 - Medium
            case 'USB Violation': 
                return 'text-emerald-400 border-emerald-500/50 bg-emerald-500/10'; // P3 - Medium
            default: 
                return 'text-slate-300 border-slate-700 bg-slate-800';
        }
    };
    const totalPages = Math.ceil(filtered.length / perPage);
    const display = filtered.slice((page - 1) * perPage, page * perPage);

    return (
        <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-xl overflow-hidden flex flex-col">
            <div className="p-4 border-b border-slate-800 bg-slate-800/20 flex flex-col sm:flex-row justify-between items-center gap-2">
                <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 tracking-widest">
                    <ShieldAlert size={14} className="text-red-400"/> Nhật ký ({filtered.length})
                </h3>
                <div className="relative w-full sm:w-48">
                    <Search className="absolute left-2 top-1/2 -translate-y-1/2 text-slate-500" size={12}/>
                    <input 
                        type="text" placeholder="Lọc nhật ký..." 
                        value={search} onChange={e => { setSearch(e.target.value); setPage(1); }}
                        className="w-full pl-7 pr-2 py-1 bg-slate-900 border border-slate-700 rounded-lg text-[10px] text-white outline-none focus:border-red-500"
                    />
                </div>
            </div>

            <div className="overflow-x-auto">
                <table className="w-full text-left whitespace-nowrap">
                    <thead className="bg-slate-950/30 text-[9px] uppercase text-slate-500 font-bold">
                        <tr>
                            <th className="p-3">Thời gian</th>
                            <th className="p-3">Loại</th>
                            <th className="p-3">Chi tiết</th>
                            <th className="p-3 text-right">Trạng thái</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/50 text-[11px]">
                        {display.map((log) => (
                            <tr key={log.ID} className="hover:bg-slate-800/30">
                                <td className="p-3 font-mono text-slate-500">{new Date(log.CreatedAt).toLocaleString('vi-VN')}</td>
                                
                                {/* CẬP NHẬT MÀU SẮC Ở ĐÂY */}
                                <td className="p-3">
                                    <span className={`px-2 py-1 rounded-md text-[9px] font-black tracking-wider border uppercase ${getAlertColor(log.alert_type)}`}>
                                        {log.alert_type}
                                    </span>
                                </td>

                                <td className="p-3 text-slate-300 max-w-[200px] truncate" title={log.description}>{log.description}</td>
                                <td className="p-3 text-right">
                                    <span className={`text-[9px] font-black px-1.5 py-0.5 rounded uppercase ${log.is_resolved ? 'text-emerald-500 bg-emerald-500/10 border border-emerald-500/20' : 'text-red-500 bg-red-500/10 border border-red-500/20'}`}>
                                        {log.is_resolved ? 'FIXED' : 'OPEN'}
                                    </span>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
            
            {totalPages > 1 && (
                <div className="p-2 border-t border-slate-800 flex justify-between items-center bg-slate-900/30">
                    <span className="text-[9px] text-slate-500">Trang {page}/{totalPages}</span>
                    <div className="flex gap-1">
                        <button disabled={page===1} onClick={()=>setPage(p=>p-1)} className="p-1 rounded bg-slate-800 text-slate-400 hover:text-white disabled:opacity-20"><ChevronLeft size={12}/></button>
                        <button disabled={page===totalPages} onClick={()=>setPage(p=>p+1)} className="p-1 rounded bg-slate-800 text-slate-400 hover:text-white disabled:opacity-20"><ChevronRight size={12}/></button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default AgentLogs;