import React, { useState, useEffect } from 'react';
import { ShieldAlert, Search, AlertTriangle, CheckCircle, Flame, Lock, Usb, Activity } from 'lucide-react';
import Pagination from '../../../../components/common/Pagination';

// Giữ nguyên getPriorityColor
const getPriorityColor = (priority) => {
    switch (priority) {
        case 'P1': return 'bg-red-500/20 text-red-500 border-red-500/50 shadow-[0_0_15px_rgba(239,68,68,0.5)]'; 
        case 'P2': return 'bg-orange-500/20 text-orange-400 border-orange-500/50'; 
        case 'P3': return 'bg-yellow-500/20 text-yellow-400 border-yellow-500/50'; 
        default: return 'bg-slate-500/20 text-slate-400 border-slate-500/50';
    }
};

// [MỚI] Icon và Màu cho Tên Case
const getTypeBadge = (type) => {
    if (type.includes('Malware')) return { icon: Flame, color: 'text-rose-400 bg-rose-500/10 border-rose-500/20' };
    if (type.includes('Firewall')) return { icon: ShieldAlert, color: 'text-red-400 bg-red-500/10 border-red-500/20' };
    if (type.includes('Defense') || type.includes('Evasion')) return { icon: Lock, color: 'text-orange-400 bg-orange-500/10 border-orange-500/20' };
    if (type.includes('USB')) return { icon: Usb, color: 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20' };
    return { icon: Activity, color: 'text-blue-400 bg-blue-500/10 border-blue-500/20' };
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
        <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-2xl overflow-hidden mt-6">
            {/* Header / Filter giữ nguyên */}
            <div className="p-5 border-b border-slate-800 bg-slate-800/20 flex flex-wrap gap-4 justify-between items-center">
                <div className="flex gap-2">
                    {['ALL', 'Open', 'Investigating', 'Resolved'].map(status => (
                        <button key={status} onClick={() => setFilterStatus(status)}
                            className={`px-4 py-2 rounded-xl text-xs font-bold transition-all ${filterStatus === status ? 'bg-indigo-600 text-white shadow-lg shadow-indigo-500/20' : 'bg-slate-900/50 text-slate-400 hover:text-white hover:bg-slate-800'}`}>
                            {status === 'ALL' ? 'Tất cả' : status === 'Open' ? 'Đang mở' : status === 'Investigating' ? 'Đang điều tra' : 'Đã đóng'}
                        </button>
                    ))}
                </div>
                <div className="relative w-64">
                    <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                    <input type="text" placeholder="Tìm theo máy trạm hoặc lỗi..." value={searchTerm} onChange={(e) => setSearchTerm(e.target.value)}
                        className="w-full bg-slate-900 border border-slate-700 text-sm text-white rounded-xl pl-9 pr-4 py-2 outline-none focus:border-indigo-500 transition-colors" />
                </div>
            </div>

            {/* Bảng Dữ liệu */}
            <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse">
                    <thead>
                        <tr className="bg-slate-900/80 border-b border-slate-800 text-[10px] uppercase tracking-widest text-slate-500">
                            <th className="p-4 font-bold">Mức độ</th>
                            <th className="p-4 font-bold">Hồ sơ sự cố</th>
                            <th className="p-4 font-bold">Nạn nhân (Asset)</th>
                            <th className="p-4 font-bold">Trạng thái</th>
                            <th className="p-4 font-bold">Thời gian tạo</th>
                            <th className="p-4 font-bold text-center">Thao tác</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/50">
                        {currentData.length === 0 ? (
                            <tr><td colSpan="6" className="p-10 text-center text-slate-500 font-bold">Không tìm thấy sự cố nào</td></tr>
                        ) : currentData.map((inc) => {
                            const badge = getTypeBadge(inc.type);
                            return (
                                <tr key={inc.ID} onClick={() => onViewDetail(inc.ID)} className="hover:bg-slate-800/30 transition-colors cursor-pointer group">
                                    <td className="p-4">
                                        <div className={`inline-flex flex-col items-center justify-center w-12 h-12 rounded-xl border ${getPriorityColor(inc.priority)}`}>
                                            <span className="text-sm font-black">{inc.priority}</span>
                                        </div>
                                    </td>
                                    <td className="p-4">
                                        <div className="flex items-center gap-2 mb-1">
                                            <span className={`px-2 py-0.5 rounded border text-[10px] font-black uppercase flex items-center gap-1 ${badge.color}`}>
                                                <badge.icon size={10}/> {inc.type}
                                            </span>
                                        </div>
                                        <p className="text-xs text-slate-400 font-medium line-clamp-1 max-w-md">Bao gồm: {inc.Alerts?.length || 1} cảnh báo chi tiết</p>
                                    </td>
                                    <td className="p-4">
                                        <p className="text-sm font-bold text-white mb-0.5">{inc.Agent?.hostname || 'Unknown'}</p>
                                        <p className="text-[10px] text-slate-500 font-mono bg-slate-900 inline-block px-1.5 rounded">{inc.Agent?.ip_address || '0.0.0.0'}</p>
                                    </td>
                                    <td className="p-4">
                                        {inc.status === 'Resolved' ? 
                                            <span className="flex items-center gap-1.5 text-xs font-bold text-emerald-500"><CheckCircle size={14}/> Đã xử lý</span> :
                                            inc.status === 'Investigating' ?
                                            <span className="flex items-center gap-1.5 text-xs font-bold text-blue-400"><Search size={14}/> Đang điều tra</span> :
                                            <span className="flex items-center gap-1.5 text-xs font-bold text-red-400 animate-pulse"><AlertTriangle size={14}/> Mới</span>
                                        }
                                    </td>
                                    <td className="p-4 text-xs font-mono text-slate-400">{new Date(inc.CreatedAt || inc.created_at).toLocaleString('vi-VN')}</td>
                                    <td className="p-4 text-center">
                                        <button onClick={(e) => { e.stopPropagation(); onViewDetail(inc.ID); }}
                                            className="text-indigo-400 hover:text-white text-xs font-bold border border-indigo-500/30 px-4 py-1.5 rounded-lg bg-indigo-500/10 hover:bg-indigo-600 transition shadow-lg" >
                                            Chi tiết
                                        </button>
                                    </td>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
            </div>
            {totalPages > 1 && (
                <div className="p-3 border-t border-slate-700 bg-slate-900/50">
                    <Pagination currentPage={currentPage} totalPages={totalPages} onPageChange={setCurrentPage} />
                </div>
            )}
        </div>
    );
};

export default IncidentTable;