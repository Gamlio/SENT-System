import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { 
    Search, Monitor, ArrowUpRight, Smartphone, User, 
    ChevronLeft, ChevronRight, ShieldBan, ShieldCheck, 
    X, UserCheck, LayoutList, ArrowUpDown // Thêm icon ArrowUpDown
} from 'lucide-react'; 
import { useAgents, getTimeAgo } from './hooks/useAgents'; 
import { useUsers } from '../User/hooks/useUsers'; 
import AgentActions from './components/AgentActions'; 
import axios from '../../api/axios';

const Agents = () => {
    const navigate = useNavigate();
    
    // Lấy sortConfig và setSortConfig từ Hook mới
    const {
        currentAgents,
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages,
        fetchAgents,
        sortConfig, setSortConfig 
    } = useAgents();

    const { users } = useUsers(); 

    const [selectedAgents, setSelectedAgents] = useState([]);
    const [showBulkModal, setShowBulkModal] = useState(false);
    const [assignModalOpen, setAssignModalOpen] = useState(false);
    const [targetAgent, setTargetAgent] = useState(null);

    // --- LOGIC XỬ LÝ ---
    const handleSelectAll = (e) => {
        if (e.target.checked) setSelectedAgents(currentAgents.map(a => a.hwid));
        else setSelectedAgents([]);
    };

    const handleSelectOne = (hwid) => {
        if (selectedAgents.includes(hwid)) setSelectedAgents(selectedAgents.filter(id => id !== hwid));
        else setSelectedAgents([...selectedAgents, hwid]);
    };

    const handleAssignManager = async (userId) => {
        try {
            await axios.put(`/agents/${targetAgent.hwid}/assign`, { user_id: userId });
            setAssignModalOpen(false);
            fetchAgents();
            alert("Đã phân bổ người chịu trách nhiệm thành công!");
        } catch (err) {
            alert("Lỗi khi phân bổ người quản lý!");
        }
    };

    // Hàm xử lý khi chọn Dropdown sắp xếp
    const handleSortChange = (e) => {
        const value = e.target.value;
        const [key, direction] = value.split('-');
        setSortConfig({ key, direction });
    };

    return (
        <div className="text-slate-200 pb-20 relative">
            <header className="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-end gap-4">
                <div>
                    <h1 className="text-3xl font-bold text-white tracking-tight">Quản lý Máy trạm</h1>
                    <p className="text-slate-400 text-sm mt-1">Giao người chịu trách nhiệm và giám sát thiết bị toàn công ty</p>
                </div>
            </header>

            {/* --- TOOLBAR: TÌM KIẾM & SẮP XẾP --- */}
            <div className="mb-6 flex flex-col md:flex-row gap-4">
                {/* 1. Ô Tìm kiếm */}
                <div className="relative flex-1">
                    <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500" size={20}/>
                    <input 
                        type="text" 
                        placeholder="Tìm theo Hostname, IP, Người quản lý..." 
                        value={searchQuery} 
                        onChange={(e) => setSearchQuery(e.target.value)} 
                        className="w-full pl-12 pr-4 py-3 bg-[#1e293b] border border-slate-800 rounded-2xl text-white outline-none focus:border-emerald-500 transition shadow-lg" 
                    />
                </div>

                {/* 2. Dropdown Sắp xếp (Mới thêm) */}
                <div className="relative min-w-[220px]">
                    <div className="absolute left-4 top-1/2 -translate-y-1/2 text-emerald-500 pointer-events-none">
                        <ArrowUpDown size={18} />
                    </div>
                    <select 
                        value={`${sortConfig.key}-${sortConfig.direction}`}
                        onChange={handleSortChange}
                        className="w-full pl-11 pr-8 py-3 bg-[#1e293b] border border-slate-800 rounded-2xl text-white outline-none focus:border-emerald-500 appearance-none cursor-pointer font-bold text-sm shadow-lg hover:bg-slate-800 transition"
                    >
                        <option value="last_seen-desc">🕒 Mới cập nhật (Mặc định)</option>
                        <option value="last_seen-asc">🕒 Cũ nhất trước</option>
                        <option value="hostname-asc">🔤 Tên máy (A-Z)</option>
                        <option value="hostname-desc">🔤 Tên máy (Z-A)</option>
                        <option value="ip_address-asc">🌐 IP (Tăng dần)</option>
                        <option value="ip_address-desc">🌐 IP (Giảm dần)</option>
                        <option value="manager-asc">👤 Người quản lý (A-Z)</option>
                        <option value="status-asc">🟢 Trạng thái</option>
                    </select>
                    {/* Mũi tên custom cho select */}
                    <div className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-500 pointer-events-none">
                        <svg width="10" height="6" viewBox="0 0 10 6" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M1 1L5 5L9 1" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
                        </svg>
                    </div>
                </div>
            </div>

            {/* Bảng dữ liệu */}
            <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-xl overflow-hidden flex flex-col min-h-[600px]">
                <div className="flex-1 overflow-x-auto">
                    <table className="w-full text-left whitespace-nowrap">
                        <thead className="text-[10px] uppercase tracking-widest text-slate-500 bg-slate-900/50">
                            <tr>
                                <th className="p-5 w-10">
                                    <input type="checkbox" checked={selectedAgents.length === currentAgents.length && currentAgents.length > 0} onChange={handleSelectAll} className="w-4 h-4 rounded bg-slate-800 border-slate-700 accent-emerald-500 cursor-pointer"/>
                                </th>
                                {/* Click vào Header để sort nhanh (Optional UX) */}
                                <th className="p-5 font-bold cursor-pointer hover:text-emerald-400 transition" onClick={() => setSortConfig({ key: 'hostname', direction: sortConfig.direction === 'asc' ? 'desc' : 'asc' })}>
                                    Máy trạm (Hostname)
                                </th>
                                <th className="p-5 font-bold">Người chịu trách nhiệm</th>
                                <th className="p-5 font-bold cursor-pointer hover:text-emerald-400 transition" onClick={() => setSortConfig({ key: 'ip_address', direction: sortConfig.direction === 'asc' ? 'desc' : 'asc' })}>
                                    Mạng (IP)
                                </th>
                                <th className="p-5 font-bold">Trạng thái</th>
                                <th className="p-5 font-bold text-right">Thao tác</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-800">
                            {currentAgents.length === 0 ? (
                                <tr>
                                    <td colSpan="6" className="p-10 text-center text-slate-500 italic">
                                        Không tìm thấy máy trạm nào phù hợp.
                                    </td>
                                </tr>
                            ) : (
                                currentAgents.map(agent => (
                                    <tr key={agent.hwid} className={`hover:bg-slate-800/50 transition-colors group ${selectedAgents.includes(agent.hwid) ? 'bg-emerald-500/5' : ''}`}>
                                        <td className="p-5">
                                            <input type="checkbox" checked={selectedAgents.includes(agent.hwid)} onChange={() => handleSelectOne(agent.hwid)} className="w-4 h-4 rounded bg-slate-800 border-slate-700 accent-emerald-500 cursor-pointer"/>
                                        </td>
                                        <td className="p-5">
                                            <div className="flex items-center gap-4">
                                                <div className="p-3 bg-slate-900 rounded-xl text-emerald-400"><Monitor size={20}/></div>
                                                <div>
                                                    <h3 className="font-bold text-white text-sm">{agent.hostname || "Unknown"}</h3>
                                                    <p className="text-[10px] text-slate-500 font-mono mt-1">HWID: {agent.hwid?.substring(0, 15)}...</p>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="p-5">
                                            <div className="flex flex-col group/manager relative">
                                                <div className="flex items-center gap-2 text-sm text-slate-300">
                                                    <User size={14} className="text-blue-400"/> 
                                                {agent.manager?.full_name || "Chưa bàn giao"}
                                                </div>
                                                <div className="flex items-center gap-2 text-xs text-slate-500">
                                                    <Smartphone size={12}/> {agent.manager?.phone || "---"}
                                                </div>
                                                <button 
                                                    onClick={() => { setTargetAgent(agent); setAssignModalOpen(true); }}
                                                    className="text-[10px] text-emerald-500 hover:text-emerald-400 font-bold mt-1 opacity-0 group-hover:opacity-100 transition-opacity text-left"
                                                >
                                                    + Thay đổi người quản lý
                                                </button>
                                            </div>
                                        </td>
                                        <td className="p-5 text-sm text-slate-400 font-mono">{agent.ip_address}</td>
                                        <td className="p-5">
                                            <div className="flex flex-col">
                                                <span className={`flex w-fit items-center gap-1.5 text-[10px] font-bold px-2.5 py-1 rounded-full uppercase ${agent.status === 'online' ? 'text-emerald-400 bg-emerald-500/10' : 'text-red-400 bg-red-500/10'}`}>
                                                    <span className={`w-1.5 h-1.5 rounded-full ${agent.status === 'online' ? 'bg-emerald-400 animate-pulse' : 'bg-red-400'}`}></span> {agent.status}
                                                </span>
                                                <span className="text-[10px] text-slate-500 mt-1.5 italic">{getTimeAgo(agent.last_seen, agent.status)}</span>
                                            </div>
                                        </td>
                                        <td className="p-5 text-right">
                                            <div className="flex justify-end items-center gap-2">
                                                <button onClick={() => navigate(`/agents/${agent.hwid}`)} className="p-2 bg-slate-800 text-slate-400 hover:text-blue-400 rounded-lg transition"><ArrowUpRight size={18}/></button>
                                                <AgentActions agent={agent} />
                                            </div>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>

                {/* --- THANH PHÂN TRANG --- */}
                {totalPages > 1 && (
                    <div className="p-4 bg-slate-900/80 border-t border-slate-800 flex flex-col sm:flex-row justify-between items-center gap-4">
                        <div className="flex items-center gap-2 text-xs text-slate-500 font-bold uppercase tracking-wider">
                            <LayoutList size={14}/>
                            Hiển thị trang <span className="text-white">{currentPage}</span> / {totalPages}
                        </div>

                        <div className="flex items-center gap-2">
                            <button 
                                disabled={currentPage === 1}
                                onClick={() => setCurrentPage(prev => prev - 1)}
                                className="flex items-center gap-1 px-3 py-2 rounded-xl bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700 disabled:opacity-30 disabled:cursor-not-allowed transition font-bold text-xs"
                            >
                                <ChevronLeft size={14}/> Trước
                            </button>

                            {/* Dãy số trang: 1 2 3 ... */}
                            <div className="flex gap-1">
                                {(() => {
                                    const pages = [];
                                    const maxVisible = 5;
                                    let startPage = Math.max(1, currentPage - Math.floor(maxVisible / 2));
                                    let endPage = Math.min(totalPages, startPage + maxVisible - 1);
                                    
                                    if (endPage - startPage < maxVisible - 1) {
                                        startPage = Math.max(1, endPage - maxVisible + 1);
                                    }
                                    
                                    if (startPage > 1) {
                                        pages.push(
                                            <button key="1" onClick={() => setCurrentPage(1)} className="w-8 h-8 rounded-xl text-xs font-bold bg-slate-800 text-slate-400 hover:bg-slate-700 hover:text-white transition">1</button>
                                        );
                                        if (startPage > 2) pages.push(<span key="ellipsis-start" className="px-1 text-slate-500">...</span>);
                                    }
                                    
                                    for (let page = startPage; page <= endPage; page++) {
                                        pages.push(
                                            <button
                                                key={page}
                                                onClick={() => setCurrentPage(page)}
                                                className={`w-8 h-8 rounded-xl text-xs font-bold transition ${
                                                    currentPage === page 
                                                    ? 'bg-emerald-500 text-white shadow-lg shadow-emerald-500/20' 
                                                    : 'bg-slate-800 text-slate-400 hover:bg-slate-700 hover:text-white'
                                                }`}
                                            >
                                                {page}
                                            </button>
                                        );
                                    }
                                    
                                    if (endPage < totalPages) {
                                        if (endPage < totalPages - 1) pages.push(<span key="ellipsis-end" className="px-1 text-slate-500">...</span>);
                                        pages.push(
                                            <button key={totalPages} onClick={() => setCurrentPage(totalPages)} className="w-8 h-8 rounded-xl text-xs font-bold bg-slate-800 text-slate-400 hover:bg-slate-700 hover:text-white transition">{totalPages}</button>
                                        );
                                    }
                                    
                                    return pages;
                                })()}
                            </div>

                            <button 
                                disabled={currentPage === totalPages}
                                onClick={() => setCurrentPage(prev => prev + 1)}
                                className="flex items-center gap-1 px-3 py-2 rounded-xl bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700 disabled:opacity-30 disabled:cursor-not-allowed transition font-bold text-xs"
                            >
                                Sau <ChevronRight size={14}/>
                            </button>
                        </div>
                    </div>
                )}
            </div>

            {/* Modal & Floating Action Bar (Giữ nguyên) */}
            {assignModalOpen && (
                <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-[9999] backdrop-blur-md p-4 animate-in fade-in duration-200">
                    <div className="bg-[#1e293b] w-full max-w-md rounded-3xl border border-slate-700 shadow-2xl overflow-hidden">
                        <div className="p-6 border-b border-slate-800 flex justify-between items-center bg-slate-800/50">
                            <div>
                                <h3 className="font-bold text-white flex items-center gap-2"><UserCheck className="text-emerald-400"/> Bàn giao máy trạm</h3>
                                <p className="text-[10px] text-slate-500 uppercase font-black mt-1">Máy: {targetAgent?.hostname}</p>
                            </div>
                            <button onClick={() => setAssignModalOpen(false)} className="text-slate-500 hover:text-white"><X size={20}/></button>
                        </div>
                        
                        <div className="max-h-80 overflow-y-auto p-2 scrollbar-hide">
                            {users.length === 0 ? (
                                <p className="text-center p-10 text-slate-500 italic text-sm">Chưa có danh sách nhân sự.</p>
                            ) : (
                                users.map(u => (
                                    <button 
                                        key={u.id}
                                        onClick={() => handleAssignManager(u.id)}
                                        className="w-full flex items-center justify-between p-4 hover:bg-emerald-500/10 rounded-2xl group transition border border-transparent hover:border-emerald-500/30 mb-1"
                                    >
                                        <div className="flex items-center gap-3">
                                            <div className="p-2.5 bg-slate-900 rounded-xl text-slate-500 group-hover:text-emerald-400 transition-colors"><User size={18}/></div>
                                            <div className="text-left">
                                                <p className="text-sm font-bold text-white group-hover:text-emerald-400">{u.full_name}</p>
                                                <p className="text-[10px] text-slate-500 font-mono">@{u.username}</p>
                                            </div>
                                        </div>
                                        <ChevronRight size={16} className="text-slate-700 group-hover:text-emerald-500 transition-transform group-hover:translate-x-1"/>
                                    </button>
                                ))
                            )}
                        </div>
                    </div>
                </div>
            )}

            {selectedAgents.length > 0 && (
                <div className="fixed bottom-10 left-1/2 -translate-x-1/2 bg-slate-800 text-white px-6 py-4 rounded-2xl shadow-2xl border border-emerald-500/50 flex items-center gap-6 z-50 animate-in slide-in-from-bottom-5 duration-300">
                    <span className="text-sm font-bold"><span className="bg-emerald-500 px-2.5 py-1 rounded-lg mr-2 text-xs font-black">{selectedAgents.length}</span> máy đã chọn</span>
                    <button onClick={() => setShowBulkModal(true)} className="flex items-center gap-2 bg-emerald-500 hover:bg-emerald-600 px-5 py-2.5 rounded-xl transition font-bold text-xs uppercase tracking-wider shadow-lg"><ShieldCheck size={16}/> Cấp phép hàng loạt</button>
                    <button onClick={() => setSelectedAgents([])} className="p-2 text-slate-400 hover:text-red-400 transition-colors"><X size={20}/></button>
                </div>
            )}
        </div>
    );
};

export default Agents;