import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Cpu, ArrowLeft, User, Smartphone, Mail, ShieldAlert, ShieldCheck, Fingerprint, Tag } from 'lucide-react';
import axios from '../../api/axios';

import AgentUSB from './components/AgentUSB';
import AgentSoftware from './components/AgentSoftware';
import AgentLogs from './components/AgentLogs';
import AgentPort from './components/AgentPort';
import { useSocketSubscription } from '../../context/useSocketSubscription';


const AgentDetail = () => {
    const { hwid } = useParams();
    const navigate = useNavigate();
    const [agent, setAgent] = useState(null);
    const [logs, setLogs] = useState([]);
    const lastFetchTime = useRef(0);

    const fetchDetail = useCallback(async (force = false) => {
        const now = Date.now();
        // Chốt chặn 2 giây
        if (!force && now - lastFetchTime.current < 2000) return;

        try {
            const res = await axios.get(`/agents/${hwid}`);
            setAgent(res.data);
            const logRes = await axios.get(`/agents/${hwid}/logs`);
            setLogs(logRes.data);
            lastFetchTime.current = Date.now();
        } catch (err) { console.error(err); }
    }, [hwid]);

    useEffect(() => {
        fetchDetail(true);
    }, [fetchDetail]);

    useSocketSubscription(['REFRESH_DATA', 'AGENT_UPDATE'], (data) => { if (data?.hwid === hwid) fetchDetail(); });

    // [MỚI]: Hàm gọi cập nhật Phòng ban
    const handleUpdateDepartment = async (tag) => {
        try {
            await axios.put(`/agents/${hwid}/department`, { department_tag: tag });
            fetchDetail(true); // Tải lại để thấy thay đổi
        } catch (err) { alert("Lỗi cập nhật phòng ban!"); }
    };

    if (!agent) return <div className="p-10 text-slate-400 font-bold text-center">Đang tải dữ liệu...</div>;
    
    const inv = agent.inventory || {}; 
    const isOnline = agent.status === 'online';

    return (
        <div className="text-slate-200 p-6 pb-20 max-w-[1600px] mx-auto">
            {/* --- HEADER --- */}
            <div className="flex flex-col md:flex-row md:items-center gap-4 mb-6">
                <div className="flex items-center gap-4">
                    <button onClick={() => navigate('/agents')} className="p-2 bg-slate-800 rounded-xl hover:bg-slate-700 transition text-slate-400 hover:text-white shrink-0">
                        <ArrowLeft size={20}/>
                    </button>
                    <div>
                        <h1 className="text-2xl font-black text-white">{agent.hostname}</h1>
                        <p className="text-xs text-slate-500 font-mono mt-0.5">{agent.hwid}</p>
                    </div>
                </div>
                
                {/* THANH TRẠNG THÁI */}
                <div className="md:ml-auto flex items-center gap-3 flex-wrap justify-start md:justify-end">
                    
                    {/* Ngữ Cảnh Phòng Ban (Drop-down đổi Tag trực tiếp) */}
                    <div className="flex items-center bg-slate-800 rounded-full border border-slate-700 overflow-hidden">
                        <div className="px-3 py-1.5 bg-slate-900 flex items-center gap-1.5 text-[10px] font-black uppercase text-slate-400 border-r border-slate-700">
                            <Tag size={12}/> Phòng Ban
                        </div>
                        <select 
                            value={agent.department_tag || 'OFFICE'} 
                            onChange={(e) => handleUpdateDepartment(e.target.value)}
                            className="bg-transparent text-[10px] font-bold uppercase text-amber-400 px-3 py-1.5 outline-none cursor-pointer hover:bg-slate-700 transition-colors"
                        >
                            <option value="OFFICE" className="bg-slate-800">Văn Phòng (Mặc định)</option>
                            <option value="DEV" className="bg-slate-800 text-blue-400">DEV (Lập trình/IT)</option>
                            <option value="FINANCE" className="bg-slate-800 text-emerald-400">FINANCE (Kế Toán)</option>
                            <option value="PROD" className="bg-slate-800 text-purple-400">PROD (Sản xuất)</option>
                        </select>
                    </div>

                    {/* Trạng thái Baseline */}
                    <div className={`px-3 py-1.5 rounded-full text-[10px] font-black uppercase tracking-widest border flex items-center gap-1.5 ${
                        agent.baseline_status === 'COMPLETED' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' :
                        agent.baseline_status === 'SCANNING' ? 'bg-amber-500/10 text-amber-400 border-amber-500/20 animate-pulse' :
                        'bg-slate-800 text-slate-500 border-slate-700'
                    }`}>
                        <Fingerprint size={12}/> BASELINE: {agent.baseline_status || 'NONE'}
                    </div>

                    <div className={`px-3 py-1.5 rounded-full text-[10px] font-bold uppercase tracking-widest border flex items-center gap-2 ${
                        isOnline ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-slate-800 text-slate-400 border-slate-700'
                    }`}>
                        <div className={`w-1.5 h-1.5 rounded-full ${isOnline ? 'bg-emerald-400 animate-pulse' : 'bg-slate-500'}`}></div>
                        {isOnline ? 'Trực tuyến' : 'Ngoại tuyến'}
                    </div>
                </div>
            </div>
           

            {/* --- GRID CHÍNH: 3 CỘT (Thông tin - Uy tín - Người quản lý) --- */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 items-stretch mb-8">
                
                {/* 1. ĐIỂM RỦI RO (TỨC THỜI) & CẤU HÌNH */}
                <div className="space-y-6 flex flex-col">
                    <div className="p-5 bg-slate-900/80 rounded-3xl border border-slate-700/50 shadow-xl relative overflow-hidden flex-shrink-0">
                        <div className="flex justify-between items-end mb-4 relative z-10">
                            <div>
                                <p className="text-[10px] text-slate-400 font-bold uppercase tracking-wider mb-1">Rủi ro Tức thời</p>
                                <div className="flex items-baseline gap-1">
                                    <span className={`text-4xl font-black tracking-tighter ${agent.risk_score > 70 ? 'text-red-500' : agent.risk_score > 30 ? 'text-yellow-500' : 'text-emerald-500'}`}>{agent.risk_score || 0}</span>
                                    <span className="text-xs text-slate-500 font-bold">/ 100</span>
                                </div>
                            </div>
                        </div>
                        
                        <div className="w-full bg-slate-800 rounded-full h-2.5 overflow-hidden">
                            <div className={`h-2.5 rounded-full ${agent.risk_score > 70 ? 'bg-red-500' : agent.risk_score > 30 ? 'bg-yellow-500' : 'bg-emerald-500'}`} style={{ width: `${agent.risk_score || 0}%` }}></div>
                        </div>
                    </div>

                    <div className="bg-[#1e293b] p-5 rounded-3xl border border-slate-800 shadow-xl flex-1">
                        <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 mb-4 tracking-widest">
                            <Cpu size={14}/> Cấu hình
                        </h3>
                        <div className="space-y-3">
                            <InfoRow label="IP Address" value={agent.ip_address} />
                            <InfoRow label="CPU Model" value={inv.cpu_model} />
                            <InfoRow label="RAM Total" value={`${inv.ram_total_gb} GB`} />
                            <InfoRow label="OS System" value={inv.os_info} />
                        </div>
                    </div>
                </div>

                {/* 2. CHỈ SỐ UY TÍN (TRUST SCORE) - Đánh giá 1 năm */}
                <div className="bg-slate-900/80 p-6 rounded-3xl border border-slate-700/50 shadow-xl flex flex-col justify-center items-center text-center relative overflow-hidden">
                    <ShieldCheck size={120} className="absolute -right-10 -bottom-10 text-emerald-500/5 rotate-12" />
                    
                    <h3 className="text-[11px] font-black text-slate-400 uppercase tracking-widest mb-6">Lý lịch An ninh (Trust Score)</h3>
                    
                    {/* Vòng tròn điểm */}
                    <div className="relative w-40 h-40 flex items-center justify-center mb-4">
                        <svg className="w-full h-full rotate-[-90deg]" viewBox="0 0 36 36">
                            <path className="text-slate-800" strokeWidth="3" stroke="currentColor" fill="none" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" />
                            <path 
                                className={`${(agent.trust_score ?? 100) >= 80 ? 'text-emerald-500' : (agent.trust_score ?? 100) >= 50 ? 'text-yellow-500' : 'text-red-500'} transition-all duration-1000`} 
                                strokeDasharray={`${agent.trust_score ?? 100}, 100`} 
                                strokeWidth="3" strokeLinecap="round" stroke="currentColor" fill="none" d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831" 
                            />
                        </svg>
                        <div className="absolute flex flex-col items-center">
                            <span className="text-4xl font-black text-white tracking-tighter">{agent.trust_score ?? 100}</span>
                            <span className="text-[9px] text-slate-500 uppercase font-bold tracking-widest">Điểm 1 Năm</span>
                        </div>
                    </div>
                    
                    <p className="text-[11px] text-slate-400 mt-2 px-4 leading-relaxed">
                        Mỗi sự cố <span className="text-red-400 font-bold">P1</span> sẽ trừ 15 điểm. Hoạt động an toàn 7 ngày liên tiếp cộng 2 điểm.
                    </p>
                </div>

                {/* 3. NGƯỜI CHỊU TRÁCH NHIỆM */}
                <div className="bg-[#1e293b] p-5 rounded-3xl border border-slate-800 shadow-xl relative overflow-hidden">
                    <div className="absolute top-0 right-0 p-3 opacity-10 text-emerald-500"><User size={80}/></div>
                    <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 mb-4 tracking-widest z-10 relative">
                        <User size={14}/> Người chịu trách nhiệm
                    </h3>
                    {agent.manager ? (
                        <div className="relative z-10">
                            <p className="text-lg font-bold text-white mb-1">{agent.manager.full_name}</p>
                            <p className="text-xs text-slate-500 font-mono mb-4">@{agent.manager.username}</p>
                            <div className="space-y-2 mt-3">
                                <div className="flex items-center gap-2 text-xs text-slate-300 bg-slate-900/50 p-2.5 rounded-xl border border-slate-700/50">
                                    <Smartphone size={14} className="text-blue-400"/> {agent.manager.phone || 'Chưa cập nhật SDT'}
                                </div>
                                <div className="flex items-center gap-2 text-xs text-slate-300 bg-slate-900/50 p-2.5 rounded-xl border border-slate-700/50">
                                    <Mail size={14} className="text-emerald-400"/> {agent.manager.email || 'Chưa cập nhật Email'}
                                </div>
                            </div>
                        </div>
                    ) : (
                        <div className="py-6 text-center border-2 border-dashed border-slate-700 rounded-xl bg-slate-900/30 mt-4">
                            <p className="text-sm text-slate-500 font-bold">Chưa phân công nhân sự</p>
                        </div>
                    )}
                </div>
            </div>

            {/* --- PHẦN 2: DANH SÁCH CHI TIẾT (XẾP DỌC, MỞ RỘNG NGANG FULL) --- */}
            <div className="space-y-6 mb-8 w-full">
                <AgentSoftware software={agent.software || []} />
                <AgentUSB usbLogs={agent.usb_logs || []} />
                <AgentPort portLogs={agent.open_ports || []} />
            </div>

            {/* --- PHẦN DƯỚI: NHẬT KÝ CẢNH BÁO (Full Width) --- */}
            <div className="space-y-4">
                 <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 tracking-widest pl-2">
                    <ShieldAlert size={14} className="text-red-400"/> Lịch sử Cảnh báo & Vi phạm
                </h3>
                <AgentLogs logs={logs} />
            </div>
        </div> 
    );
};

const InfoRow = ({ label, value }) => (
    <div className="flex justify-between text-xs border-b border-slate-800/50 pb-2 last:border-0">
        <span className="text-slate-500 font-bold">{label}</span>
        <span className="text-white font-mono truncate max-w-[60%]">{value || 'N/A'}</span>
    </div>
);

export default AgentDetail;