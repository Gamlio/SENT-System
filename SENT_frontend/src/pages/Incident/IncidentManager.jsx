import React, { useState, useEffect } from 'react';
import { useIncidents } from './hooks/useIncidents';
import IncidentDetail from './IncidentDetail';
import { ShieldAlert, Search, ChevronLeft, ChevronRight, Activity, User, Clock } from 'lucide-react';

const IncidentManager = () => {
    const [page, setPage] = useState(1);
    const [selectedId, setSelectedId] = useState(null);
    const { incidents, total, loading, fetchList } = useIncidents();

    useEffect(() => { fetchList(page); }, [page, fetchList]);

    if (selectedId) return <IncidentDetail incidentId={selectedId} onBack={() => setSelectedId(null)} />;

    const totalPages = Math.ceil(total / 12);

    return (
        <div className="p-6 bg-[#050B14] min-h-[calc(100vh-60px)] text-slate-200">
            {/* Header chuyên nghiệp */}
            <div className="flex justify-between items-end mb-8">
                <div>
                    <h1 className="text-2xl font-black text-white flex items-center gap-3 tracking-tighter uppercase">
                        <ShieldAlert className="text-indigo-500" size={28}/> Incident Workbench
                    </h1>
                    <p className="text-[10px] text-slate-500 mt-1 uppercase tracking-[0.2em] font-bold">
                        Centralized Threat Investigation & Case Management
                    </p>
                </div>
                <div className="flex items-center gap-2 bg-[#0A101D] border border-slate-800 px-4 py-2 rounded-xl">
                    <div className="w-2 text-indigo-500 h-2 rounded-full bg-indigo-500 animate-pulse"></div>
                    <span className="text-xs font-mono text-indigo-400 font-bold uppercase tracking-widest">Active Cases: {total}</span>
                </div>
            </div>

            {/* Table Header Refined - Glassmorphism */}
            <div className="hidden md:flex items-center gap-4 px-6 py-4 mb-3 bg-slate-900/20 backdrop-blur-sm border-y border-slate-800/50 rounded-lg shadow-[0_0_15px_rgba(0,0,0,0.1)]">
                <div className="w-24 text-center">
                    <span className="text-[10px] font-black text-indigo-400 uppercase tracking-[0.25em]">Mức độ</span>
                </div>
                <div className="flex-1">
                    <span className="text-[10px] font-black text-slate-400 uppercase tracking-[0.25em] flex items-center gap-2">
                        <div className="w-1 h-1 bg-indigo-500 rounded-full animate-pulse"></div>
                        Chi tiết hồ sơ sự cố
                    </span>
                </div>
                <div className="w-48 text-right">
                    <span className="text-[10px] font-black text-slate-400 uppercase tracking-[0.25em]">Phân công / Trạng thái</span>
                </div>
            </div>

            {/* Danh sách sự cố */}
            <div className="flex flex-col gap-3">
                {loading ? (
                    [...Array(6)].map((_, i) => <div key={i} className="h-20 bg-slate-800/10 animate-pulse rounded-xl border border-slate-800/50"></div>)
                ) : incidents.map(inc => (
                    <div 
                        key={inc.id || inc.ID} 
                        onClick={() => setSelectedId(inc.id || inc.ID)}
                        className="group flex flex-col md:flex-row items-center gap-4 bg-[#0A101D] border border-slate-800 p-4 rounded-xl hover:border-indigo-500/40 transition-all cursor-pointer"
                    >
                        <div className="w-full md:w-24 text-center py-1 rounded text-[10px] font-black border border-slate-700 uppercase bg-slate-800/30 text-slate-400 group-hover:border-indigo-500/30">
                            {inc.priority} | {inc.severity}
                        </div>
                        <div className="flex-1 min-w-0">
                            <h4 className="text-sm font-bold text-slate-200 truncate group-hover:text-white uppercase tracking-tight">{inc.type}</h4>
                            <p className="text-[11px] text-slate-500 truncate italic mt-0.5">{inc.description}</p>
                        </div>
                        <div className="flex flex-row md:flex-col items-center md:items-end gap-4 md:gap-1 text-[11px] font-mono md:w-48 text-right">
                            <div className="flex items-center gap-1.5 text-indigo-300 uppercase font-bold">
                                <User size={12}/> {inc.assignee?.username || 'Unassigned'}
                            </div>
                            <div className="flex items-center gap-1.5 text-slate-500 italic">
                                <Clock size={12}/> {new Date(inc.CreatedAt || inc.created_at).toLocaleDateString()}
                            </div>
                        </div>
                    </div>
                ))}
            </div>

            {/* Pagination[cite: 66] */}
            <div className="mt-8 flex justify-center items-center gap-2">
                <button disabled={page === 1} onClick={() => setPage(p => p - 1)} className="p-2 bg-[#0A101D] border border-slate-800 rounded-lg disabled:opacity-20"><ChevronLeft size={18}/></button>
                <button disabled={page >= totalPages} onClick={() => setPage(p => p + 1)} className="p-2 bg-[#0A101D] border border-slate-800 rounded-lg disabled:opacity-20"><ChevronRight size={18}/></button>
            </div>
        </div>
    );
};

export default IncidentManager;