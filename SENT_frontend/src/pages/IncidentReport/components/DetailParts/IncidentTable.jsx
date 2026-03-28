import React, { useState, useEffect } from 'react';
import { ShieldAlert, Search, AlertTriangle, CheckCircle, Flame, Lock, Usb, Activity, Clock } from 'lucide-react';
import Pagination from '../../../../components/common/Pagination';

const getPriorityColor = (priority) => {
    if (priority === 'P1') return 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.8)]';
    if (priority === 'P2') return 'bg-orange-500';
    return 'bg-yellow-500';
};

const getTypeBadge = (type = '') => {
    if (type.includes('Malware')) return { icon: Flame, color: 'text-rose-400' };
    if (type.includes('Firewall')) return { icon: ShieldAlert, color: 'text-red-400' };
    if (type.includes('Defense') || type.includes('Evasion')) return { icon: Lock, color: 'text-orange-400' };
    if (type.includes('USB')) return { icon: Usb, color: 'text-emerald-400' };
    return { icon: Activity, color: 'text-blue-400' };
};

const IncidentTable = ({ rawData, onViewDetail }) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [filterStatus, setFilterStatus] = useState('ALL'); 
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 10;

    useEffect(() => setCurrentPage(1), [searchTerm, filterStatus]);

    const filteredData = (rawData || []).filter(inc => {
        const matchesSearch = inc.Agent?.hostname?.toLowerCase().includes(searchTerm.toLowerCase()) || 
                              inc.type?.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = filterStatus === 'ALL' || inc.status === filterStatus;
        return matchesSearch && matchesStatus;
    });

    const totalPages = Math.ceil(filteredData.length / itemsPerPage);
    const currentData = filteredData.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);

    return (
        <div className="bg-[#0A101D] rounded-lg border border-slate-800 shadow-2xl flex flex-col overflow-hidden font-sans mt-4">
            {/* TOOLBAR */}
            <div className="p-3 border-b border-slate-800 bg-[#111827] flex flex-wrap gap-4 justify-between items-center">
                <div className="flex gap-1 p-1 bg-[#050B14] rounded-md border border-slate-800">
                    {['ALL', 'Open', 'Investigating', 'Resolved'].map(status => (
                        <button key={status} onClick={() => setFilterStatus(status)}
                            className={`px-3 py-1.5 rounded text-[10px] font-black uppercase tracking-widest transition-all ${filterStatus === status ? 'bg-indigo-600 text-white shadow-md' : 'text-slate-500 hover:text-white hover:bg-slate-800'}`}>
                            {status === 'ALL' ? 'Tất cả' : status}
                        </button>
                    ))}
                </div>
                <div className="relative w-72">
                    <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                    <input type="text" placeholder="Search IPs, Hostnames, Rules..." value={searchTerm} onChange={(e) => setSearchTerm(e.target.value)}
                        className="w-full bg-[#050B14] border border-slate-800 text-xs text-white rounded pl-9 pr-4 py-2 outline-none focus:border-indigo-500 transition-colors font-mono" />
                </div>
            </div>

            {/* BẢNG DỮ LIỆU */}
            <div className="overflow-x-auto custom-scrollbar">
                <table className="w-full text-left border-collapse whitespace-nowrap">
                    <thead>
                        <tr className="bg-[#111827] text-[10px] uppercase tracking-widest text-slate-500 border-b border-slate-800">
                            <th className="p-3 font-black">Level</th>
                            <th className="p-3 font-black">Detection Type</th>
                            <th className="p-3 font-black">Target Asset</th>
                            <th className="p-3 font-black">Status</th>
                            <th className="p-3 font-black text-right">Timestamp</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/50">
                        {currentData.length === 0 ? (
                            <tr><td colSpan="5" className="p-8 text-center text-slate-600 text-[11px] font-bold uppercase tracking-widest">NO RECORDS FOUND</td></tr>
                        ) : currentData.map((inc) => {
                            const badge = getTypeBadge(inc.type);
                            const hoursOpen = Math.floor((new Date() - new Date(inc.CreatedAt || inc.created_at)) / (1000 * 60 * 60));
                            
                            return (
                                <tr key={inc.ID} onClick={() => onViewDetail(inc.ID)} className="hover:bg-slate-800/40 transition-colors cursor-pointer group">
                                    <td className="p-3">
                                        <div className="flex items-center gap-2">
                                            <span className={`w-1.5 h-6 rounded-full ${getPriorityColor(inc.priority)}`}></span>
                                            <div>
                                                <p className="text-[10px] font-mono text-slate-500">#{inc.ID}</p>
                                                <p className={`text-[11px] font-black ${inc.priority === 'P1' ? 'text-red-400' : 'text-orange-400'}`}>{inc.priority}</p>
                                            </div>
                                        </div>
                                    </td>
                                    <td className="p-3 min-w-[200px]">
                                        <div className="flex items-center gap-1.5 mb-0.5">
                                            <badge.icon size={12} className={badge.color}/>
                                            <span className="text-xs font-bold text-white group-hover:text-indigo-300 transition">{inc.type}</span>
                                        </div>
                                        <p className="text-[10px] text-slate-500 font-mono truncate max-w-[250px]">{inc.description}</p>
                                    </td>
                                    <td className="p-3">
                                        <p className="text-xs font-bold text-slate-200 mb-0.5 group-hover:text-white">{inc.Agent?.hostname || 'Unknown'}</p>
                                        <p className="text-[10px] text-slate-500 font-mono bg-[#050B14] inline-block px-1 rounded border border-slate-800">{inc.Agent?.ip_address || '0.0.0.0'}</p>
                                    </td>
                                    <td className="p-3">
                                        <span className={`px-2 py-0.5 rounded text-[9px] font-black uppercase tracking-widest border ${
                                            inc.status === 'Open' ? 'text-red-400 border-red-500/30 bg-red-500/10' : 
                                            inc.status === 'Investigating' ? 'text-blue-400 border-blue-500/30 bg-blue-500/10' : 
                                            'text-emerald-400 border-emerald-500/30 bg-emerald-500/10'
                                        }`}>
                                            {inc.status}
                                        </span>
                                    </td>
                                    <td className="p-3 text-right">
                                        <p className="text-[11px] font-mono text-slate-400">{new Date(inc.CreatedAt || inc.created_at).toLocaleString('vi-VN')}</p>
                                        {inc.status !== 'Resolved' && (
                                            <p className={`text-[9px] font-bold mt-1 flex items-center justify-end gap-1 uppercase tracking-widest ${hoursOpen > 24 ? 'text-red-500 animate-pulse' : 'text-slate-500'}`}>
                                                <Clock size={10}/> {hoursOpen}h Pending
                                            </p>
                                        )}
                                    </td>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
            </div>
            {totalPages > 1 && (
                <div className="p-2 border-t border-slate-800 bg-[#111827]">
                    <Pagination currentPage={currentPage} totalPages={totalPages} onPageChange={setCurrentPage} />
                </div>
            )}
        </div>
    );
};

export default IncidentTable;