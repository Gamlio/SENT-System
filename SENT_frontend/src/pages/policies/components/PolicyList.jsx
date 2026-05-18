import React, { useState, useMemo } from 'react';
import { Users, Globe, Trash2, X, Search, ShieldCheck, ChevronLeft, ChevronRight, ArrowUpDown, Tag } from 'lucide-react';
import { PolicyStatusTag, PolicyTypeTag } from './PolicyShared';
import BulkDeleteModal from '../../../components/common/BulkDeleteModal';

const PolicyList = ({ policies, onDelete, onBulkDelete, groups }) => {
    const [selectedIds, setSelectedIds] = useState([]);
    const [isBulkDeleteModalOpen, setIsBulkDeleteModalOpen] = useState(false);
    const [search, setSearch] = useState('');
    const [sortConfig, setSortConfig] = useState({ key: 'CreatedAt', direction: 'desc' });
    const [page, setPage] = useState(1);
    const perPage = 10;

    const getGroupName = (groupId) => {
        if (!groupId) return "GLOBAL POLICY";
        const group = groups?.find(g => (g.id || g.ID) === groupId);
        return group ? (group.name || group.Name).toUpperCase() : `NHÓM #${groupId}`;
    };

    const filteredPolicies = useMemo(() => {
        return (policies || []).filter(p =>
            (p.title || '').toLowerCase().includes(search.toLowerCase()) ||
            (p.value || '').toLowerCase().includes(search.toLowerCase())
        );
    }, [policies, search]);

    const displayedPolicies = filteredPolicies.slice((page - 1) * perPage, page * perPage);
    const totalPages = Math.ceil(filteredPolicies.length / perPage);

    return (
        <div className="flex-1 flex flex-col min-h-0 bg-[#0A101D] rounded-lg border border-slate-800 shadow-2xl overflow-hidden relative">
            <div className="p-3 border-b border-slate-800 bg-[#111827] flex justify-between items-center shrink-0">
                <div className="relative w-72">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" size={14} />
                    <input 
                        type="text" placeholder="Search rules, values..." 
                        value={search} onChange={(e) => {setSearch(e.target.value); setPage(1);}}
                        className="w-full bg-[#050B14] border border-slate-800 text-xs text-white rounded pl-9 pr-4 py-1.5 outline-none focus:border-indigo-500 font-mono transition-colors"
                    />
                </div>
                <div className="flex items-center gap-2">
                    <Tag className="text-indigo-500" size={12}/> 
                    <span className="text-[10px] font-mono font-bold text-slate-400 uppercase">Tổng số: {filteredPolicies.length} luật</span>
                </div>
            </div>

            <div className="overflow-x-auto custom-scrollbar flex-1 pb-20">
                <table className="w-full text-left border-collapse whitespace-nowrap">
                    <thead className="sticky top-0 z-10 bg-[#111827]">
                        <tr className="border-b border-slate-800 text-[10px] uppercase tracking-widest text-slate-500 font-black">
                            <th className="p-3 w-10 text-center">
                                <input type="checkbox" className="w-3.5 h-3.5 rounded border-slate-700 bg-[#050B14] accent-indigo-500 cursor-pointer" 
                                    onChange={(e) => setSelectedIds(e.target.checked ? displayedPolicies.map(p => p.ID) : [])}
                                    checked={displayedPolicies.length > 0 && selectedIds.length === displayedPolicies.length}
                                />
                            </th>
                            <th className="p-3">Quy tắc / Giá trị</th>
                            <th className="p-3">Phạm vi áp dụng</th>
                            <th className="p-3">Phân loại</th>
                            <th className="p-3">Trạng thái</th>
                            <th className="p-3 text-right">Hành động</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/50">
                        {displayedPolicies.length === 0 ? (
                            <tr><td colSpan="6" className="p-20 text-center text-slate-600 font-bold uppercase tracking-widest opacity-20">No policies found</td></tr>
                        ) : (
                            displayedPolicies.map((p) => (
                                <tr key={p.ID} className="hover:bg-slate-800/30 transition-colors group cursor-pointer">
                                <td className="p-3 text-center">
                                        <input type="checkbox" checked={selectedIds.includes(p.ID)} onChange={() => setSelectedIds(prev => prev.includes(p.ID) ? prev.filter(i => i !== p.ID) : [...prev, p.ID])}
                                            className="w-3.5 h-3.5 rounded border-slate-700 bg-[#050B14] accent-indigo-500" />
                                </td>
                                <td className="p-3">
                                    <div className="flex flex-col">
                                            <span className="font-black text-white text-xs group-hover:text-indigo-400 transition-colors uppercase tracking-tight">{p.title}</span>
                                        <code className="text-[10px] text-slate-500 font-mono mt-0.5">{p.value}</code>
                                    </div>
                                </td>
                                <td className="p-3">
                                        <div className="text-[9px] text-indigo-400 uppercase font-black tracking-widest border border-indigo-500/30 bg-indigo-500/10 w-max px-2 py-0.5 rounded flex items-center gap-1.5">
                                            {p.group_id ? <Users size={10}/> : <Globe size={10}/>} {getGroupName(p.group_id)}
                                    </div>
                                </td>
                                    <td className="p-3"><PolicyTypeTag type={p.policy_type} /></td>
                                    <td className="p-3"><PolicyStatusTag status={p.approval_status} /></td>
                                <td className="p-3 text-right">
                                        <button onClick={() => onDelete(p.ID)} className="p-2 text-slate-500 hover:text-red-400 transition-colors rounded-lg hover:bg-red-500/10 shadow-sm">
                                        <Trash2 size={14}/>
                                    </button>
                                </td>
                            </tr>
                            ))
                        )}
                    </tbody>
                </table>
            </div>

            {totalPages > 1 && (
                <div className="p-2 border-t border-slate-800 flex justify-between items-center bg-[#111827] shrink-0">
                    <span className="text-[10px] text-slate-500 font-black uppercase ml-2 tracking-widest">Page <span className="text-white">{page}</span> / {totalPages}</span>
                    <div className="flex gap-1.5 pr-1">
                        <button disabled={page===1} onClick={()=>setPage(p=>p-1)} className="p-1.5 rounded bg-slate-800 border border-slate-700 text-slate-400 hover:text-white disabled:opacity-30 transition-colors"><ChevronLeft size={14}/></button>
                        <button disabled={page===totalPages} onClick={()=>setPage(p=>p+1)} className="p-1.5 rounded bg-slate-800 border border-slate-700 text-slate-400 hover:text-white disabled:opacity-30 transition-colors"><ChevronRight size={14}/></button>
                    </div>
                </div>
            )}

            {selectedIds.length > 0 && (
                <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 bg-[#0A101D] border border-slate-700 px-6 py-3 rounded-2xl shadow-[0_-10px_30px_rgba(0,0,0,0.5)] flex items-center gap-6 animate-in slide-in-from-bottom-4">
                    <span className="text-xs font-black text-white uppercase tracking-tight"><span className="text-indigo-400">{selectedIds.length}</span> Rules Selected</span>
                    <button onClick={() => setIsBulkDeleteModalOpen(true)} className="bg-red-600 hover:bg-red-700 text-white text-[10px] font-black px-4 py-2 rounded-xl uppercase tracking-widest transition-all shadow-lg shadow-red-900/20">Remove Bulk</button>
                    <button onClick={() => setSelectedIds([])} className="text-slate-500 hover:text-white transition-colors"><X size={18}/></button>
                </div>
            )}
            
            <BulkDeleteModal isOpen={isBulkDeleteModalOpen} onClose={() => setIsBulkDeleteModalOpen(false)} onConfirm={onBulkDelete} items={policies.filter(p => selectedIds.includes(p.ID))} type="POLICY" />
        </div>
    );
};

export default PolicyList;