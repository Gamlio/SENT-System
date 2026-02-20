import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Search, Monitor, ArrowUpRight, Smartphone, User, Activity, ChevronLeft, ChevronRight, ShieldBan, ShieldCheck, X } from 'lucide-react'; 
import { PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { useAgents, getTimeAgo } from '../../hooks/useAgents'; 
import AgentActions from '../../components/AgentActions'; 
import axios from '../../api/axios';

const Agents = () => {
    const navigate = useNavigate();
    const {
        currentAgents, filteredAgents,
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages, indexOfFirstItem, indexOfLastItem,
        statusChartData, osChartData, onlineCount, offlineCount
    } = useAgents();

    // STATE CHO TÍNH NĂNG CHỌN HÀNG LOẠT
    const [selectedAgents, setSelectedAgents] = useState([]);
    const [showBulkModal, setShowBulkModal] = useState(false);
    const [bulkSoftwareName, setBulkSoftwareName] = useState('');

    // Logic Checkbox
    const handleSelectAll = (e) => {
        if (e.target.checked) setSelectedAgents(currentAgents.map(a => a.hwid));
        else setSelectedAgents([]);
    };

    const handleSelectOne = (hwid) => {
        if (selectedAgents.includes(hwid)) setSelectedAgents(selectedAgents.filter(id => id !== hwid));
        else setSelectedAgents([...selectedAgents, hwid]);
    };

    // Gọi API cấp phép hàng loạt
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
            setSelectedAgents([]); // Reset chọn sau khi thành công
            alert(`Đã cấp phép ${bulkSoftwareName} cho ${selectedAgents.length} thiết bị!`);
        } catch (err) { alert("Lỗi cấp phép hàng loạt!"); }
    };

    return (
        <div className="text-slate-200 pb-20 relative">
            {/* Header */}
            <header className="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-end gap-4">
                <div>
                    <h1 className="text-3xl font-bold text-white tracking-tight">Quản lý Máy trạm</h1>
                    <p className="text-slate-400 text-sm mt-1">Định danh, giám sát và kiểm soát thiết bị đầu cuối toàn hệ thống</p>
                </div>
                <div className="flex items-center gap-3">
                    <button onClick={() => navigate('/admin/software-policies')} className="flex items-center gap-2 bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white px-5 py-2.5 rounded-xl border border-slate-700 transition font-bold text-sm shadow-lg">
                        <ShieldBan size={18} className="text-red-400"/> Chính sách phần mềm (Blacklist chung)
                    </button>
                </div>
            </header>

            {/* --- KHU VỰC 1: BIỂU ĐỒ (Giữ nguyên) --- */}
            {/* ... (Đoạn code Biểu đồ PieChart và BarChart vẫn giữ nguyên như cũ, mình lược bớt ở đây để bạn dễ copy, bạn cứ dán chèn vào file của bạn) ... */}

            {/* --- KHU VỰC 2: TÌM KIẾM --- */}
            <div className="mb-6 relative">
                <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500" size={20}/>
                <input type="text" placeholder="Tìm kiếm theo Hostname, IP, Tên nhân viên hoặc Số điện thoại..." value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} className="w-full pl-12 pr-4 py-4 bg-[#1e293b] border border-slate-800 rounded-2xl text-white outline-none focus:border-emerald-500 transition shadow-lg" />
            </div>

            {/* --- KHU VỰC 3: BẢNG DỮ LIỆU CÓ CHECKBOX --- */}
            <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-xl pb-10">
                <div className="w-full">
                    <table className="w-full text-left whitespace-nowrap">
                        <thead className="text-[10px] uppercase tracking-widest text-slate-500 bg-slate-900/50">
                            <tr>
                                {/* CỘT CHECKBOX CHỌN TẤT CẢ */}
                                <th className="p-5 w-10">
                                    <input 
                                        type="checkbox" 
                                        checked={selectedAgents.length === currentAgents.length && currentAgents.length > 0} 
                                        onChange={handleSelectAll} 
                                        className="w-4 h-4 rounded bg-slate-800 border-slate-700 accent-emerald-500 cursor-pointer"
                                    />
                                </th>
                                <th className="p-5 font-bold">Máy trạm (Hostname)</th>
                                <th className="p-5 font-bold">Người sử dụng</th>
                                <th className="p-5 font-bold">Mạng (IP)</th>
                                <th className="p-5 font-bold">Trạng thái</th>
                                <th className="p-5 font-bold text-right">Chi tiết</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-800">
                            {currentAgents.length === 0 ? (
                                <tr><td colSpan="6" className="p-10 text-center text-slate-500 italic">Không tìm thấy máy trạm nào. Hãy chạy Agent ở máy Client.</td></tr>
                            ) : (
                                currentAgents.map(agent => (
                                    <tr key={agent.hwid} className={`hover:bg-slate-800/50 transition-colors group ${selectedAgents.includes(agent.hwid) ? 'bg-emerald-500/5' : ''}`}>
                                        
                                        {/* CỘT CHECKBOX TỪNG DÒNG */}
                                        <td className="p-5">
                                            <input 
                                                type="checkbox" 
                                                checked={selectedAgents.includes(agent.hwid)} 
                                                onChange={() => handleSelectOne(agent.hwid)} 
                                                className="w-4 h-4 rounded bg-slate-800 border-slate-700 accent-emerald-500 cursor-pointer"
                                            />
                                        </td>
                                        
                                        <td className="p-5">
                                            <div className="flex items-center gap-4">
                                                <div className="p-3 bg-slate-900 rounded-xl text-emerald-400 group-hover:bg-emerald-500/10 transition"><Monitor size={20}/></div>
                                                <div>
                                                    <h3 className="font-bold text-white text-sm">{agent.hostname || "Unknown"}</h3>
                                                    <p className="text-[10px] text-slate-500 font-mono mt-1">HWID: {agent.hwid?.substring(0, 15)}...</p>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="p-5">
                                            <div className="flex flex-col gap-1">
                                                <div className="flex items-center gap-2 text-sm text-slate-300"><User size={14} className="text-blue-400"/> <span className="font-bold">{agent.user_name || "Chưa định danh"}</span></div>
                                                <div className="flex items-center gap-2 text-xs text-slate-500"><Smartphone size={12} className="text-slate-400"/> {agent.user_phone || "---"}</div>
                                            </div>
                                        </td>
                                        <td className="p-5 text-sm text-slate-400 font-mono">{agent.ip_address || "N/A"}</td>
                                        <td className="p-5">
                                            <div className="flex flex-col">
                                                <span className={`flex w-fit items-center gap-1.5 text-[10px] font-bold px-2.5 py-1 rounded-full uppercase ${agent.status === 'online' ? 'text-emerald-400 bg-emerald-500/10' : 'text-red-400 bg-red-500/10'}`}>
                                                    <span className={`w-1.5 h-1.5 rounded-full ${agent.status === 'online' ? 'bg-emerald-400 animate-pulse' : 'bg-red-400'}`}></span> {agent.status}
                                                </span>
                                                <span className="text-[10px] text-slate-500 mt-1.5 italic font-medium">{getTimeAgo(agent.last_seen, agent.status)}</span>
                                            </div>
                                        </td>
                                        <td className="p-5">
                                            <div className="flex justify-end items-center gap-2 opacity-100 lg:opacity-0 lg:group-hover:opacity-100 transition-opacity duration-300">
                                                <button onClick={() => navigate(`/agents/${agent.hwid}`)} className="p-2 bg-slate-800 text-slate-400 hover:text-blue-400 hover:bg-blue-500/10 rounded-lg transition" title="Xem chi tiết"><ArrowUpRight size={18}/></button>
                                                <AgentActions agent={agent} />
                                            </div>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* --- THANH CÔNG CỤ NỔI KHI CÓ MÁY ĐƯỢC CHỌN (FLOATING ACTION BAR) --- */}
            {selectedAgents.length > 0 && (
                <div className="fixed bottom-10 left-1/2 -translate-x-1/2 bg-slate-800 text-white px-6 py-4 rounded-2xl shadow-2xl border border-emerald-500/50 flex items-center gap-6 z-50 animate-bounce-in">
                    <span className="text-sm font-bold flex items-center gap-2">
                        <span className="bg-emerald-500 text-white px-2.5 py-1 rounded-lg">{selectedAgents.length}</span> máy trạm đã chọn
                    </span>
                    <div className="w-px h-6 bg-slate-700"></div>
                    <button 
                        onClick={() => setShowBulkModal(true)}
                        className="flex items-center gap-2 bg-emerald-500 hover:bg-emerald-600 text-white px-4 py-2 rounded-xl transition font-bold text-sm shadow-lg shadow-emerald-500/20"
                    >
                        <ShieldCheck size={18}/> Cấp phép hàng loạt
                    </button>
                    <button onClick={() => setSelectedAgents([])} className="p-2 text-slate-400 hover:text-red-400 transition" title="Hủy chọn">
                        <X size={20}/>
                    </button>
                </div>
            )}

            {/* --- MODAL NHẬP TÊN PHẦN MỀM --- */}
            {showBulkModal && (
                <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-[9999] backdrop-blur-sm">
                    <div className="bg-[#1e293b] p-8 rounded-3xl border border-slate-700 shadow-2xl w-[400px]">
                        <h3 className="text-xl font-bold text-white mb-2 flex items-center gap-2">
                            <ShieldCheck className="text-emerald-400" size={24}/> Cấp phép Zero Trust
                        </h3>
                        <p className="text-sm text-slate-400 mb-6">Thêm phần mềm hợp lệ cho {selectedAgents.length} máy trạm đang chọn.</p>
                        
                        <form onSubmit={handleBulkWhitelist}>
                            <input 
                                type="text" 
                                value={bulkSoftwareName} 
                                onChange={(e) => setBulkSoftwareName(e.target.value)} 
                                placeholder="VD: zalo.exe, unikey..."
                                required
                                autoFocus
                                className="w-full p-4 bg-slate-900 border border-slate-700 rounded-xl outline-none focus:border-emerald-500 text-white mb-6" 
                            />
                            <div className="flex gap-3">
                                <button type="button" onClick={() => setShowBulkModal(false)} className="flex-1 py-3 text-slate-400 hover:text-white bg-slate-800 hover:bg-slate-700 rounded-xl font-bold transition">Hủy</button>
                                <button type="submit" className="flex-1 py-3 bg-emerald-500 hover:bg-emerald-600 text-white rounded-xl font-bold transition shadow-lg shadow-emerald-500/20">Xác nhận</button>
                            </div>
                        </form>
                    </div>
                </div>
            )}

        </div>
    );
};

export default Agents;