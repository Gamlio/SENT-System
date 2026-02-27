import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Search, Monitor, ArrowUpRight, Smartphone, User, ChevronLeft, ChevronRight, ShieldBan, ShieldCheck, X, UserCheck } from 'lucide-react'; 
import { useAgents, getTimeAgo } from '../../hooks/useAgents'; 
import { useUsers } from '../../hooks/useUsers'; // IMPORT THÊM ĐỂ LẤY DANH SÁCH NHÂN VIÊN
import AgentActions from '../../components/AgentActions'; 
import axios from '../../api/axios';

const Agents = () => {
    const navigate = useNavigate();
    const {
        currentAgents, filteredAgents,
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages, indexOfFirstItem, indexOfLastItem,
        fetchAgents // Lấy thêm hàm fetch để load lại dữ liệu sau khi gán
    } = useAgents();

    const { users } = useUsers(); // Danh sách nhân viên từ hệ thống

    // --- STATE QUẢN LÝ ---
    const [selectedAgents, setSelectedAgents] = useState([]);
    const [showBulkModal, setShowBulkModal] = useState(false);
    const [bulkSoftwareName, setBulkSoftwareName] = useState('');

    // State cho Modal gán người quản lý
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

    // Hàm thực hiện gán người quản lý xuống Backend
    const handleAssignManager = async (userId) => {
        try {
            await axios.put(`/agents/${targetAgent.hwid}/assign`, { user_id: userId });
            setAssignModalOpen(false);
            fetchAgents(); // Tải lại danh sách để cập nhật tên người quản lý mới
            alert("Đã phân bổ người chịu trách nhiệm thành công!");
        } catch (err) {
            alert("Lỗi khi phân bổ người quản lý!");
        }
    };

    const handleBulkWhitelist = async (e) => {
        e.preventDefault();
        if (!bulkSoftwareName.trim()) return;
        try {
            await axios.post('/agents/bulk-whitelist', {
                hwids: selectedAgents,
                software_name: bulkSoftwareName.trim()
            });
            setShowBulkModal(false);
            setBulkSoftwareName('');
            setSelectedAgents([]);
            alert(`Đã cấp phép ${bulkSoftwareName} thành công!`);
        } catch (err) { alert("Lỗi cấp phép hàng loạt!"); }
    };

    return (
        <div className="text-slate-200 pb-20 relative">
            {/* Header */}
            <header className="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-end gap-4">
                <div>
                    <h1 className="text-3xl font-bold text-white tracking-tight">Quản lý Máy trạm</h1>
                    <p className="text-slate-400 text-sm mt-1">Giao người chịu trách nhiệm và giám sát thiết bị toàn công ty</p>
                </div>
                <div className="flex items-center gap-3">
                    <button onClick={() => navigate('/admin/software-policies')} className="flex items-center gap-2 bg-slate-800 hover:bg-slate-700 text-slate-300 px-5 py-2.5 rounded-xl border border-slate-700 transition font-bold text-sm shadow-lg">
                        <ShieldBan size={18} className="text-red-400"/> Chính sách Blacklist
                    </button>
                </div>
            </header>

            {/* Tìm kiếm */}
            <div className="mb-6 relative">
                <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500" size={20}/>
                <input type="text" placeholder="Tìm theo Hostname, IP, Người quản lý..." value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} className="w-full pl-12 pr-4 py-4 bg-[#1e293b] border border-slate-800 rounded-2xl text-white outline-none focus:border-emerald-500 transition shadow-lg" />
            </div>

            {/* Bảng dữ liệu */}
            <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-xl pb-10">
                <table className="w-full text-left whitespace-nowrap">
                    <thead className="text-[10px] uppercase tracking-widest text-slate-500 bg-slate-900/50">
                        <tr>
                            <th className="p-5 w-10">
                                <input type="checkbox" checked={selectedAgents.length === currentAgents.length && currentAgents.length > 0} onChange={handleSelectAll} className="w-4 h-4 rounded bg-slate-800 border-slate-700 accent-emerald-500 cursor-pointer"/>
                            </th>
                            <th className="p-5 font-bold">Máy trạm (Hostname)</th>
                            <th className="p-5 font-bold">Người chịu trách nhiệm</th>
                            <th className="p-5 font-bold">Mạng (IP)</th>
                            <th className="p-5 font-bold">Trạng thái</th>
                            <th className="p-5 font-bold text-right">Thao tác</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800">
                        {currentAgents.map(agent => (
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
                                            <span className="font-bold">{agent.manager?.full_name || "Chưa bàn giao"}</span>
                                        </div>
                                        <div className="flex items-center gap-2 text-xs text-slate-500">
                                            <Smartphone size={12}/> {agent.manager?.phone || "---"}
                                        </div>
                                        {/* Nút mở Modal gán nhanh */}
                                        <button 
                                            onClick={() => { setTargetAgent(agent); setAssignModalOpen(true); }}
                                            className="text-[10px] text-emerald-500 hover:text-emerald-400 font-bold mt-1 opacity-0 group-hover:opacity-100 transition-opacity"
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
                                <td className="p-5">
                                    <div className="flex justify-end items-center gap-2">
                                        <button onClick={() => navigate(`/agents/${agent.hwid}`)} className="p-2 bg-slate-800 text-slate-400 hover:text-blue-400 rounded-lg transition"><ArrowUpRight size={18}/></button>
                                        <AgentActions agent={agent} />
                                    </div>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            {/* --- MODAL CHỌN NHÂN VIÊN CHỊU TRÁCH NHIỆM --- */}
            {assignModalOpen && (
                <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-[9999] backdrop-blur-md p-4">
                    <div className="bg-[#1e293b] w-full max-w-md rounded-3xl border border-slate-700 shadow-2xl overflow-hidden animate-in zoom-in duration-200">
                        <div className="p-6 border-b border-slate-800 flex justify-between items-center bg-slate-800/50">
                            <div>
                                <h3 className="font-bold text-white flex items-center gap-2"><UserCheck className="text-emerald-400"/> Bàn giao máy trạm</h3>
                                <p className="text-[10px] text-slate-500 uppercase font-black mt-1">Máy: {targetAgent?.hostname}</p>
                            </div>
                            <button onClick={() => setAssignModalOpen(false)} className="text-slate-500 hover:text-white"><X size={20}/></button>
                        </div>
                        
                        <div className="max-h-80 overflow-y-auto p-2">
                            {users.length === 0 ? (
                                <p className="text-center p-10 text-slate-500 italic">Chưa có danh sách nhân sự.</p>
                            ) : (
                                users.map(u => (
                                    <button 
                                        key={u.id}
                                        onClick={() => handleAssignManager(u.id)}
                                        className="w-full flex items-center justify-between p-4 hover:bg-emerald-500/10 rounded-2xl group transition"
                                    >
                                        <div className="flex items-center gap-3">
                                            <div className="p-2 bg-slate-900 rounded-lg text-slate-500 group-hover:text-emerald-400"><User size={16}/></div>
                                            <div className="text-left">
                                                <p className="text-sm font-bold text-white group-hover:text-emerald-400">{u.full_name}</p>
                                                <p className="text-[10px] text-slate-500">@{u.username}</p>
                                            </div>
                                        </div>
                                        <ChevronRight size={16} className="text-slate-700 group-hover:text-emerald-500"/>
                                    </button>
                                ))
                            )}
                        </div>
                        <div className="p-4 bg-slate-900/50 text-center">
                            <button onClick={() => setAssignModalOpen(false)} className="text-xs font-bold text-slate-500 hover:text-slate-300">Hủy bỏ</button>
                        </div>
                    </div>
                </div>
            )}

            {/* Floating Action Bar (Giữ nguyên) */}
            {selectedAgents.length > 0 && (
                <div className="fixed bottom-10 left-1/2 -translate-x-1/2 bg-slate-800 text-white px-6 py-4 rounded-2xl shadow-2xl border border-emerald-500/50 flex items-center gap-6 z-50">
                    <span className="text-sm font-bold"><span className="bg-emerald-500 px-2.5 py-1 rounded-lg mr-2">{selectedAgents.length}</span> máy đã chọn</span>
                    <button onClick={() => setShowBulkModal(true)} className="flex items-center gap-2 bg-emerald-500 hover:bg-emerald-600 px-4 py-2 rounded-xl transition font-bold text-sm shadow-lg"><ShieldCheck size={18}/> Cấp phép hàng loạt</button>
                    <button onClick={() => setSelectedAgents([])} className="p-2 text-slate-400 hover:text-red-400"><X size={20}/></button>
                </div>
            )}
        </div>
    );
};

export default Agents;