import React, { useState, useMemo } from 'react';
import { Users, Globe, Trash2, X, Search, ShieldCheck, ChevronLeft, ChevronRight } from 'lucide-react';
import BulkDeleteModal from '../../../components/common/BulkDeleteModal';

// Component con hiển thị Tag Trạng thái (Đồng bộ style Assets)
const PolicyStatusTag = ({ status }) => {
    const isApproved = status === 'APPROVED';
    return (
        <div className={`px-2 py-0.5 rounded text-[9px] font-black uppercase tracking-widest border flex items-center gap-1.5 w-max ${
            isApproved ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30' : 'bg-amber-500/10 text-amber-500 border-amber-500/30'
        }`}>
            <div className={`w-1.5 h-1.5 rounded-full ${isApproved ? 'bg-emerald-400' : 'bg-amber-400 animate-pulse'}`}></div>
            {isApproved ? 'Hoạt động' : 'Chờ duyệt'}
        </div>
    );
};

const PolicyList = ({ policies, onDelete, onBulkDelete, groups }) => {
    const [selectedIds, setSelectedIds] = useState([]);
    const [isBulkDeleteModalOpen, setIsBulkDeleteModalOpen] = useState(false);
    const [search, setSearch] = useState('');
    const [page, setPage] = useState(1);
    const perPage = 10;

    const handleSelect = (id) => {
        setSelectedIds(prev => prev.includes(id) ? prev.filter(i => i !== id) : [...prev, id]);
    };

    const getGroupName = (groupId) => {
        if (!groupId) return "Toàn cục";
        const group = groups?.find(g => (g.id || g.ID) === groupId);
        return group ? group.name || group.Name : `Nhóm #${groupId}`;
    };

    // Lọc và Phân trang
    const filteredPolicies = useMemo(() => {
        return (policies || []).filter(p =>
            (p.title || '').toLowerCase().includes(search.toLowerCase()) ||
            (p.value || '').toLowerCase().includes(search.toLowerCase())
        );
    }, [policies, search]);

    const totalPages = Math.ceil(filteredPolicies.length / perPage);
    const displayedPolicies = filteredPolicies.slice((page - 1) * perPage, page * perPage);

    return (
        <div className="flex-1 flex flex-col min-h-0 bg-slate-900/50 rounded-2xl border border-slate-800 overflow-hidden shadow-inner">
            {/* Thanh công cụ tìm kiếm */}
            <div className="p-3 border-b border-slate-800 flex justify-between items-center bg-slate-800/20 shrink-0">
                <div className="relative w-64">
                    <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                    <input 
                        type="text" placeholder="Tìm kiếm chính sách, giá trị..." 
                        value={search} onChange={(e) => {setSearch(e.target.value); setPage(1);}}
                        className="w-full bg-slate-950/50 border border-slate-700 text-xs text-white rounded-lg pl-8 pr-3 py-2 outline-none focus:border-indigo-500/50 transition-colors"
                    />
                </div>
                <span className="text-[10px] text-slate-500 font-bold bg-slate-800 px-3 py-1.5 rounded-lg border border-slate-700">
                    Tổng số: <span className="text-white">{filteredPolicies.length}</span> luật
                </span>
            </div>

            {/* Bảng dữ liệu */}
            <div className="overflow-x-auto custom-scrollbar flex-1 relative">
                <table className="w-full text-left border-collapse whitespace-nowrap">
                    <thead className="sticky top-0 z-10 bg-slate-900">
                        <tr className="border-b border-slate-800 text-[10px] uppercase tracking-widest text-slate-500 font-black">
                            <th className="p-3 w-10 text-center">
                                <input type="checkbox" className="w-3.5 h-3.5 rounded border-slate-700 bg-[#050B14] accent-indigo-500 cursor-pointer" 
                                    onChange={(e) => setSelectedIds(e.target.checked ? displayedPolicies.map(p => p.ID) : [])}
                                    checked={displayedPolicies.length > 0 && selectedIds.length === displayedPolicies.length}
                                />
                            </th>
                            <th className="p-3">Quy tắc / Giá trị</th>
                            <th className="p-3">Phạm vi</th>
                            <th className="p-3">Loại</th>
                            <th className="p-3">Trạng thái</th>
                            <th className="p-3 text-right">Hành động</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/50 text-xs">
                        {displayedPolicies.length === 0 ? (
                            <tr>
                                <td colSpan="6" className="p-12 text-center text-slate-500">
                                    <ShieldCheck size={40} className="mx-auto mb-3 opacity-20"/>
                                    <span className="font-bold">Không tìm thấy chính sách nào</span>
                                </td>
                            </tr>
                        ) : (
                            displayedPolicies.map((p) => (
                            <tr key={p.ID} className="hover:bg-slate-800/30 transition-colors group">
                                <td className="p-3 text-center">
                                    <input type="checkbox" checked={selectedIds.includes(p.ID)} onChange={() => handleSelect(p.ID)}
                                        className="w-3.5 h-3.5 rounded border-slate-700 bg-[#050B14] accent-indigo-500 cursor-pointer" />
                                </td>
                                <td className="p-3">
                                    <div className="flex flex-col">
                                        <span className="font-bold text-white group-hover:text-indigo-400 transition-colors">{p.title}</span>
                                        <code className="text-[10px] text-slate-500 font-mono mt-0.5">{p.value}</code>
                                    </div>
                                </td>
                                <td className="p-3">
                                    <div className="flex items-center gap-1.5 text-[10px] font-bold">
                                        {p.group_id ? <Users size={12} className="text-purple-400"/> : <Globe size={12} className="text-blue-400"/>}
                                        <span className={p.group_id ? "text-purple-400" : "text-blue-400"}>
                                            {getGroupName(p.group_id)}
                                        </span>
                                    </div>
                                </td>
                                <td className="p-3">
                                    <span className={`text-[9px] font-black px-1.5 py-0.5 rounded border ${
                                        p.policy_type === 'BLACKLIST' ? 'text-red-400 border-red-500/20 bg-red-500/5' : 'text-emerald-400 border-emerald-500/20 bg-emerald-500/5'
                                    }`}>
                                        {p.policy_type}
                                    </span>
                                </td>
                                <td className="p-3">
                                    <PolicyStatusTag status={p.approval_status} />
                                </td>
                                <td className="p-3 text-right">
                                    <button onClick={() => onDelete(p.ID)} className="p-2 text-slate-500 hover:text-red-400 transition-colors rounded-lg hover:bg-red-500/10">
                                        <Trash2 size={14}/>
                                    </button>
                                </td>
                            </tr>
                            ))
                        )}
                    </tbody>
                </table>
            </div>

            {/* Phân trang */}
            {totalPages > 1 && (
                <div className="p-2 border-t border-slate-800 flex justify-between items-center bg-slate-900/80 shrink-0">
                    <span className="text-[10px] text-slate-500 font-bold uppercase ml-2">Trang <span className="text-white">{page}</span> / {totalPages}</span>
                    <div className="flex gap-1.5 pr-1">
                        <button disabled={page===1} onClick={()=>setPage(p=>p-1)} className="p-1.5 rounded bg-slate-800 border border-slate-700 text-slate-400 hover:text-white disabled:opacity-30 transition-colors shadow-sm"><ChevronLeft size={14}/></button>
                        <button disabled={page===totalPages} onClick={()=>setPage(p=>p+1)} className="p-1.5 rounded bg-slate-800 border border-slate-700 text-slate-400 hover:text-white disabled:opacity-30 transition-colors shadow-sm"><ChevronRight size={14}/></button>
                    </div>
                </div>
            )}

            {/* Bulk Actions (Thanh công cụ chọn nhiều) */}
            {selectedIds.length > 0 && (
                <div className="absolute bottom-16 left-1/2 -translate-x-1/2 z-50 bg-[#0A101D] border border-slate-700 px-6 py-3 rounded-2xl shadow-2xl flex items-center gap-6 animate-in slide-in-from-bottom-4">
                    <span className="text-xs font-bold text-white"><span className="text-indigo-400">{selectedIds.length}</span> luật đang chọn</span>
                    <button onClick={() => setIsBulkDeleteModalOpen(true)} className="bg-red-500 hover:bg-red-600 text-white text-[10px] font-black px-4 py-2 rounded-xl uppercase tracking-widest transition-all">Gỡ bỏ hàng loạt</button>
                    <button onClick={() => setSelectedIds([])} className="text-slate-500 hover:text-white"><X size={18}/></button>
                </div>
            )}
            
            <BulkDeleteModal isOpen={isBulkDeleteModalOpen} onClose={() => setIsBulkDeleteModalOpen(false)} onConfirm={onBulkDelete} items={policies.filter(p => selectedIds.includes(p.ID))} type="POLICY" />
        </div>
    );
};

export default PolicyList;