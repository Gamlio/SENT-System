import React, { useState } from 'react';
import { ShieldAlert, Search, ChevronLeft, ChevronRight, AlertTriangle } from 'lucide-react';

const assetLogs = ({ logs }) => {
    const [search, setSearch] = useState('');
    const [page, setPage] = useState(1);
    const perPage = 5; 

    // Lọc theo mô tả hoặc loại lỗi
    const filtered = (logs || []).filter(l => 
        l.description?.toLowerCase().includes(search.toLowerCase()) || 
        l.alert_type?.toLowerCase().includes(search.toLowerCase())
    );

    // Đồng bộ 100% với Backend (alerts.go & software.go)
    const getAlertStyle = (type) => {
        switch(type) {
            case 'Firewall Disabled': 
                return { color: 'text-red-400', bg: 'bg-red-500/10', border: 'border-red-500/30', badge: 'P1' };
            case 'Malware Detected': 
                return { color: 'text-rose-400', bg: 'bg-rose-500/10', border: 'border-rose-500/30', badge: 'P1' };
            case 'Defense Evasion': // Ghost Registry
                return { color: 'text-orange-400', bg: 'bg-orange-500/10', border: 'border-orange-500/30', badge: 'P2' };
            case 'Unauthorized Port': 
                return { color: 'text-amber-400', bg: 'bg-amber-500/10', border: 'border-amber-500/30', badge: 'P2' };
            case 'Software Violation': 
                return { color: 'text-blue-400', bg: 'bg-blue-500/10', border: 'border-blue-500/30', badge: 'P3' };
            case 'USB Violation': 
                return { color: 'text-emerald-400', bg: 'bg-emerald-500/10', border: 'border-emerald-500/30', badge: 'P3' };
            default:
                return { color: 'text-slate-400', bg: 'bg-slate-500/10', border: 'border-slate-500/30', badge: 'INFO' };
        }
    };

    const totalPages = Math.ceil(filtered.length / perPage);
    const displayedLogs = filtered.slice((page - 1) * perPage, page * perPage);

    return (
        <div className="bg-slate-900/50 rounded-2xl border border-slate-800 overflow-hidden shadow-inner">
            <div className="p-3 border-b border-slate-800 flex justify-between items-center bg-slate-800/20">
                <div className="relative w-48">
                    <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                    <input 
                        type="text" placeholder="Lọc cảnh báo..." 
                        value={search} onChange={(e) => {setSearch(e.target.value); setPage(1);}}
                        className="w-full bg-slate-950/50 border border-slate-700 text-xs text-white rounded-lg pl-8 pr-3 py-1.5 outline-none focus:border-red-500/50 transition-colors"
                    />
                </div>
                <span className="text-[10px] text-slate-500 font-bold bg-slate-800 px-2 py-1 rounded-md">
                    Tổng: {filtered.length}
                </span>
            </div>

            <div className="overflow-x-auto min-h-[250px]">
                <table className="w-full text-left border-collapse">
                    <thead>
                        <tr className="border-b border-slate-800 text-[10px] uppercase tracking-wider text-slate-500 bg-slate-900">
                            <th className="p-3 font-bold">Thời gian</th>
                            <th className="p-3 font-bold">Phân loại</th>
                            <th className="p-3 font-bold">Chi tiết sự kiện</th>
                            <th className="p-3 font-bold text-right">Trạng thái</th>
                        </tr>
                    </thead>
                    <tbody className="text-xs divide-y divide-slate-800/50">
                        {displayedLogs.length === 0 ? (
                            <tr>
                                <td colSpan="4" className="p-8 text-center text-slate-500">
                                    <ShieldAlert size={32} className="mx-auto mb-2 opacity-20"/>
                                    <span className="font-bold">Không có dữ liệu cảnh báo</span>
                                </td>
                            </tr>
                        ) : (
                            displayedLogs.map((log) => {
                                const style = getAlertStyle(log.alert_type);
                                return (
                                    <tr key={log.ID} className="hover:bg-slate-800/30 transition-colors group">
                                        <td className="p-3 text-slate-400 font-mono text-[10px] whitespace-nowrap">
                                            {new Date(log.created_at).toLocaleString('vi-VN')}
                                        </td>
                                        <td className="p-3">
                                            <div className="flex items-center gap-2">
                                                <span className={`text-[9px] font-black px-1.5 py-0.5 rounded border ${style.bg} ${style.color} ${style.border}`}>
                                                    {style.badge}
                                                </span>
                                                <span className={`font-bold ${style.color}`}>{log.alert_type}</span>
                                            </div>
                                        </td>
                                        <td className="p-3 text-slate-300 max-w-md">
                                            <p className="font-semibold text-white mb-0.5">{log.title}</p>
                                            <p className="text-[11px] text-slate-500 line-clamp-2" title={log.description}>{log.description}</p>
                                        </td>
                                        <td className="p-3 text-right">
                                            <span className={`flex items-center justify-end gap-1 text-[9px] font-black px-2 py-1 rounded uppercase ${log.is_resolved ? 'text-emerald-500 bg-emerald-500/10' : 'text-red-400 bg-red-500/10 border border-red-500/20'}`}>
                                                {!log.is_resolved && <AlertTriangle size={10} className="animate-pulse"/>}
                                                {log.is_resolved ? 'Đã xử lý' : 'Đang mở'}
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
                <div className="p-2 border-t border-slate-800 flex justify-between items-center bg-slate-900/80">
                    <span className="text-[10px] text-slate-500 font-bold uppercase ml-2">Trang <span className="text-white">{page}</span> / {totalPages}</span>
                    <div className="flex gap-1.5 pr-1">
                        <button disabled={page===1} onClick={()=>setPage(p=>p-1)} className="p-1.5 rounded bg-slate-800 text-slate-400 hover:text-white disabled:opacity-30 transition-colors"><ChevronLeft size={14}/></button>
                        <button disabled={page===totalPages} onClick={()=>setPage(p=>p+1)} className="p-1.5 rounded bg-slate-800 text-slate-400 hover:text-white disabled:opacity-30 transition-colors"><ChevronRight size={14}/></button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default assetLogs;