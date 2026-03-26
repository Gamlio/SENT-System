import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { 
    Search, Monitor, ShieldAlert, Smartphone, User, 
    ChevronRight, X, UserCheck, LayoutList, ArrowUpDown, Trash2,
    ShieldCheck, Laptop, Activity,Tag
} from 'lucide-react'; 
import { useAgents, getTimeAgo } from './hooks/useAgents'; 
import { useUsers } from '../User/hooks/useUsers'; 
import AgentActions from './components/AgentActions'; 
import GenerateTokenButton from './components/GenerateTokenButton';
import AgentBulkActions from './components/AgentBulkActions';
import AppDialog from '../../components/AppDialog';
import { useSocketSubscription } from '../../context/useSocketSubscription';
import axios from '../../api/axios';

// --- Static Data ---
const riskScoringLevels = [
    { label: 'Rủi Ro Thấp', value: 'Low', color: 'cyan', icon: <ShieldCheck />, bg: '#e6fffb' },
    { label: 'Rủi Ro Trung Bình', value: 'Medium', color: 'orange', icon: <Monitor />, bg: '#fff7e6' },
    { label: 'Rủi Ro Cao', value: 'High', color: 'red', icon: <ShieldAlert />, bg: '#fff1f0' },
    { label: 'Nguy Hiểm', value: 'Critical', color: 'magenta', icon: <ShieldAlert />, bg: '#fff0f6' },
];

const AgentStatusTag = React.memo(({ status }) => {
    const isOnline = status === 'online';
    return (
        <div className={`px-2 py-0.5 rounded-md text-[10px] font-black uppercase tracking-tighter border flex items-center gap-1 w-fit ${
            isOnline ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-slate-800 text-slate-500 border-slate-700'
        }`}>
            <div className={`w-1 h-1 rounded-full ${isOnline ? 'bg-emerald-400 animate-pulse' : 'bg-slate-500'}`}></div>
            {isOnline ? 'Online' : 'Offline'}
        </div>
    );
});

const RiskScoreDisplay = React.memo(({ score }) => {
    const level = riskScoringLevels.find(l => {
        if (l.value === 'Low' && score <= 30) return true;
        if (l.value === 'Medium' && score > 30 && score <= 70) return true;
        if (l.value === 'High' && score > 70 && score <= 90) return true;
        if (l.value === 'Critical' && score > 90) return true;
        return false;
    });

    if (!level) return <span className="text-slate-500 text-sm font-mono">{score}</span>;

    return (
        <div className="flex items-center gap-2">
            <span className="text-xl font-black text-white w-8 text-right font-mono">{score}</span>
            <div style={{ color: level.color }} className="flex items-center gap-1 text-xs font-bold px-2 py-1 rounded-lg bg-slate-800 border border-slate-700">
                {React.cloneElement(level.icon, { size: 12 })}
                {level.label}
            </div>
        </div>
    );
});

const Agents = () => {
    const navigate = useNavigate();
    const {
        currentAgents = [], searchQuery, setSearchQuery, 
        currentPage, setCurrentPage, totalPages, fetchAgents,
        sortConfig, setSortConfig, agents = [], loading
    } = useAgents();

    const { users = [] } = useUsers();
    
    // --- State Management ---
    const [selectedAgents, setSelectedAgents] = useState([]);
    const [showAssignModal, setShowAssignModal] = useState(false);
    const [targetAgentHwid, setTargetAgentHwid] = useState(null);
    const [deleteReason, setDeleteReason] = useState('');
    const [dialogConfig, setDialogConfig] = useState({ isOpen: false, type: '', hwids: [] });

    // --- Real-time Optimization (Debounce) ---
    const debounceRef = useRef(null);
    const handleAgentUpdate = useCallback(() => {
        if (debounceRef.current) clearTimeout(debounceRef.current);
        debounceRef.current = setTimeout(() => {
            console.log("🔄 [WS] Refreshing Agent List (Debounced)");
            fetchAgents();
        }, 500);
    }, [fetchAgents]);

    useSocketSubscription(['AGENT_STATUS_CHANGED', 'REFRESH_AGENT_LIST', 'BASELINE_COMPLETED'], handleAgentUpdate);

    // --- Action Handlers ---
    const closeDialog = () => {
        setDialogConfig(prev => ({ ...prev, isOpen: false }));
        setDeleteReason('');
    };

    const handleAssignManager = async (userId) => {
        try {
            if (targetAgentHwid) {
                await axios.post(`/agents/${targetAgentHwid}/assign`, { user_id: userId });
            } else {
                await axios.post('/agents/bulk-assign', { hwids: selectedAgents, user_id: userId });
            }
            setShowAssignModal(false);
            setSelectedAgents([]);
            fetchAgents();
        } catch (err) {
            alert("Lỗi phân công: " + (err.response?.data?.error || "Server error"));
        }
    };

    // --- Table UI Structure (TỐI ƯU KÍCH THƯỚC VỊ TRÍ) ---
   const tableColumns = useMemo(() => [
            {
                key: 'hostname',
                label: 'Thiết bị & Trạng thái',
                sortable: true,
                className: 'w-[35%]', // Cố định tỷ lệ
                render: (a) => (
                    <div className="flex items-center gap-3 max-w-full overflow-hidden">
                        <div className={`p-2.5 rounded-xl shrink-0 transition-colors ${a.status === 'online' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-slate-800 text-slate-500'}`}>
                            <Monitor size={18} />
                        </div>
                        <div className="min-w-0 flex-1"> {/* min-w-0 là then chốt để truncate hoạt động trong flex */}
                            <h4 className="font-bold text-white text-sm truncate group-hover:text-emerald-400 transition-colors" title={a.hostname}>
                                {a.hostname}
                            </h4>
                            <div className="flex items-center gap-2 mt-0.5">
                                <span className="text-[10px] font-mono text-slate-500 shrink-0">{a.ip_address}</span>
                                <span className="text-slate-700">•</span>
                                <AgentStatusTag status={a.status} />
                            </div>
                        </div>
                    </div>
                )
            },
            {
                key: 'manager',
                label: 'Người phụ trách',
                className: 'w-[25%]',
                render: (a) => (
                    <div className="flex items-center gap-3 overflow-hidden">
                        <div className="w-8 h-8 rounded-full bg-slate-800 shrink-0 flex items-center justify-center text-slate-500 border border-slate-700">
                            <User size={14} />
                        </div>
                        <div className="min-w-0">
                            <p className="text-xs font-bold text-slate-300 truncate">{a.manager?.full_name || 'Chưa bàn giao'}</p>
                            <p className="text-[10px] text-slate-500 uppercase font-black tracking-tighter truncate">
                                {a.device_type || 'OFFICE'}
                            </p>
                        </div>
                    </div>
                )
            },
            {
                key: 'risk',
                label: 'Rủi ro',
                sortable: true,
                className: 'w-[20%]',
                render: (a) => <div className="flex justify-start"><RiskScoreDisplay score={a.risk_score} /></div>
            },
            {
                key: 'last_seen',
                label: 'Cập nhật',
                sortable: true,
                className: 'w-[20%]',
                render: (a) => (
                    <div className="text-slate-400 text-[11px] font-medium whitespace-nowrap overflow-hidden text-ellipsis">
                        {getTimeAgo(a.last_seen, a.status)}
                    </div>
                )
            }
        ], []);
    // --- Render ---
    return (
        //pb-40 để nhường chỗ cho thanh Bulk Actions ở đáy
        <div className="p-6 max-w-[1600px] mx-auto text-slate-200 pb-40 relative">
            
            {/* 1. Header Section: Sửa lỗi vị trí và tỷ lệ */}
            {/* Trên mobile flex-col gap-4, trên md+ flex-row items-center justify-between mb-10 */}
            <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 mb-10">
                <div>
                    <h1 className="text-3xl font-black text-white tracking-tight flex items-center gap-3">
                        <Monitor className="text-emerald-500" size={32}/> 
                        Hệ thống Máy trạm
                    </h1>
                    <p className="text-slate-500 font-medium mt-1">
                        Quản lý và giám sát an ninh thiết bị đầu cuối ({agents.length} máy).
                    </p>
                </div>
                {/* Nút bấm tự động nằm sang bên phải trên máy tính */}
                <div className="w-full md:w-auto">
                    <GenerateTokenButton />
                </div>
            </div>

            {/* 2. Main Content Table */}
            <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-2xl overflow-hidden relative">
                
                {/* Thanh tìm kiếm: Tối ưu vị trí */}
                <div className="p-5 border-b border-slate-800 bg-slate-800/20 flex flex-col sm:flex-row gap-4 sm:items-center justify-between">
                    <div className="relative w-full sm:max-w-md">
                        <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500" size={18} />
                        <input 
                            type="text" 
                            placeholder="Tìm kiếm theo tên máy, IP, HWID..."
                            className="w-full bg-slate-900 border border-slate-700 rounded-2xl py-3 pl-12 pr-4 text-sm text-white placeholder:text-slate-600 focus:border-emerald-500/50 outline-none transition-all shadow-inner"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                        />
                    </div>
                    <div className="flex items-center gap-2 self-end sm:self-auto">
                        <Tag color="green" className="font-mono text-xs px-3 py-1 rounded-lg">Online: {agents.filter(a=>a.status === 'online').length}</Tag>
                        <button onClick={() => fetchAgents()} className="p-3 bg-slate-900 hover:bg-slate-800 rounded-xl transition text-slate-500 hover:text-white border border-slate-700">
                            <Activity size={18} />
                        </button>
                    </div>
                </div>

                {/* Bảng dữ liệu co giãn thông minh */}
                <div className="overflow-x-auto">
                    <table className="w-full text-left border-collapse table-fixed min-w-[1000px] ">
                        <thead>
                            <tr className="border-b border-slate-800/50 bg-slate-900/30">
                                <th className="p-4 w-12 text-center">
                                    <input 
                                        type="checkbox" 
                                        className="w-4 h-4 rounded border-slate-700 bg-slate-800 checked:bg-emerald-500 transition-all cursor-pointer accent-emerald-500"
                                        onChange={(e) => {
                                            if (e.target.checked) setSelectedAgents(currentAgents.map(a => a.hwid));
                                            else setSelectedAgents([]);
                                        }}
                                    />
                                </th>
                                {tableColumns.map(col => (
                                    <th key={col.key} className={`p-4 text-[10px] font-black uppercase tracking-widest text-slate-500 ${col.className}`}>
                                        <div className="flex items-center gap-2 cursor-pointer hover:text-slate-300 transition-colors" 
                                             onClick={() => col.sortable && setSortConfig({ key: col.key, direction: sortConfig.direction === 'asc' ? 'desc' : 'asc' })}>
                                            {col.label} {col.sortable && <ArrowUpDown size={10}/>}
                                        </div>
                                    </th>
                                ))}
                                <th className="p-4 text-[10px] font-black uppercase tracking-widest text-slate-500 w-24 text-right">Thao tác</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-800/40">
                            {currentAgents.map((agent) => (
                                <tr key={agent.hwid} className="group hover:bg-slate-800/30 transition-all cursor-pointer items-center" onClick={() => navigate(`/agents/${agent.hwid}`)}>
                                    <td className="p-4 text-center" onClick={(e) => e.stopPropagation()}>
                                        <input 
                                            type="checkbox" 
                                            checked={selectedAgents.includes(agent.hwid)}
                                            onChange={() => {
                                                setSelectedAgents(prev => prev.includes(agent.hwid) ? prev.filter(id => id !== agent.hwid) : [...prev, agent.hwid]);
                                            }}
                                            className="w-4 h-4 rounded border-slate-700 bg-slate-800 checked:bg-emerald-500 transition-all cursor-pointer accent-emerald-500"
                                        />
                                    </td>
                                    {tableColumns.map(col => (
                                        <td key={col.key} className={`p-4 align-middle ${col.className}`}>
                                            {col.render(agent)}
                                        </td>
                                    ))}
                                    <td className="p-4 text-right" onClick={(e) => e.stopPropagation()}>
                                        <AgentActions agent={agent} onRefresh={fetchAgents} />
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* 3. Bulk Actions Bar: Sửa vị trí thành cố định ở đáy */}
            {selectedAgents.length > 0 && (
                <div className="fixed bottom-0 left-0 right-0 z-50 p-4 bg-slate-900/80 backdrop-blur-sm border-t border-slate-800 shadow-lg animate-in slide-in-from-bottom duration-300">
                    <div className="max-w-[1600px] mx-auto">
                        <AgentBulkActions 
                            selectedAgents={selectedAgents} 
                            clearSelection={() => setSelectedAgents([])} 
                            onRefresh={fetchAgents}
                            onOpenAssignModal={() => { setTargetAgentHwid(null); setShowAssignModal(true); }}
                            onOpenDeleteModal={(hwids) => setDialogConfig({ isOpen: true, type: 'BULK_DELETE', hwids })}
                        />
                    </div>
                </div>
            )}

            {/* 4. Modals giữ nguyên logic */}
            {showAssignModal && (
                <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-[100] flex items-center justify-center p-4">
                    <div className="bg-[#1e293b] w-full max-w-lg rounded-3xl border border-slate-700 shadow-2xl overflow-hidden animate-in zoom-in duration-200">
                        <div className="p-6 border-b border-slate-800 flex justify-between items-center bg-slate-800/50">
                            <h3 className="text-xl font-bold text-white flex items-center gap-3"><UserCheck className="text-emerald-400"/> Phân công quản lý</h3>
                            <button onClick={() => setShowAssignModal(false)} className="text-slate-500 hover:text-white transition"><X size={24}/></button>
                        </div>
                        <div className="p-6 max-h-[400px] overflow-y-auto space-y-2">
                            {users.map(u => (
                                <button key={u.id} onClick={() => handleAssignManager(u.id)} className="w-full flex items-center justify-between p-4 hover:bg-emerald-500/10 rounded-2xl border border-transparent hover:border-emerald-500/30 transition-all group">
                                    <div className="flex items-center gap-3">
                                        <div className="p-2 bg-slate-900 rounded-lg text-slate-500 group-hover:text-emerald-400 transition-colors"><User size={18}/></div>
                                        <div className="text-left">
                                            <p className="text-sm font-bold text-white">{u.full_name}</p>
                                            <p className="text-[10px] text-slate-500 font-mono">@{u.username}</p>
                                        </div>
                                    </div>
                                    <ChevronRight size={16} className="text-slate-700 group-hover:text-emerald-400 transition-transform group-hover:translate-x-1"/>
                                </button>
                            ))}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Agents;