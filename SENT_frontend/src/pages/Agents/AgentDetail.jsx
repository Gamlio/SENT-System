import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Cpu, ArrowLeft, User, Smartphone, Mail, ShieldAlert } from 'lucide-react';
import axios from '../../api/axios';

import AgentUSB from './components/AgentUSB';
import AgentSoftware from './components/AgentSoftware';
import AgentLogs from './components/AgentLogs';

const AgentDetail = () => {
    const { hwid } = useParams();
    const navigate = useNavigate();
    const [agent, setAgent] = useState(null);
    const [logs, setLogs] = useState([]);   

    useEffect(() => {
        const fetchDetail = async () => {
            try {
                const res = await axios.get(`/agents/${hwid}`);
                setAgent(res.data);
                const logRes = await axios.get(`/agents/${hwid}/logs`);
                setLogs(logRes.data);
            } catch (err) { console.error(err); }
        };
        fetchDetail();
    }, [hwid]);

    if (!agent) return <div className="p-10 text-slate-400 font-bold text-center">Đang tải dữ liệu...</div>;
    
    const inv = agent.inventory || {}; 
    const isOnline = agent.is_online === true;

    return (
        <div className="text-slate-200 p-6 pb-20 max-w-[1600px] mx-auto">
            {/* --- HEADER --- */}
            <div className="flex items-center gap-4 mb-6">
                <button onClick={() => navigate('/agents')} className="p-2 bg-slate-800 rounded-xl hover:bg-slate-700 transition text-slate-400 hover:text-white shrink-0">
                    <ArrowLeft size={20}/>
                </button>
                
                <div>
                    <h1 className="text-2xl font-black text-white">{agent.hostname}</h1>
                    <p className="text-xs text-slate-500 font-mono mt-0.5">{agent.hwid}</p>
                </div>
                
                {/* THANH TRẠNG THÁI (CĂN LỀ PHẢI) */}
                <div className="ml-auto flex items-center gap-3 flex-wrap justify-end">
                    
                    {/* 1. BADGE PHÂN LOẠI TÀI SẢN (ASSET TIER) */}
                    {(() => {
                        let tierLabel = 'VĂN PHÒNG (TIER 3)';
                        let tierStyle = 'bg-slate-500/10 text-slate-400 border-slate-500/20';
                        
                        if (agent.device_type === 'SERVER') {
                            tierLabel = 'MÁY CHỦ (TIER 1)';
                            tierStyle = 'bg-purple-500/10 text-purple-400 border-purple-500/30 shadow-[0_0_10px_rgba(168,85,247,0.2)]';
                        } else if (agent.device_type === 'IT_ADMIN') {
                            tierLabel = 'IT ADMIN (TIER 2)';
                            tierStyle = 'bg-blue-500/10 text-blue-400 border-blue-500/30';
                        } else if (agent.device_type === 'GUEST') {
                            tierLabel = 'MÁY KHÁCH (TIER 4)';
                            tierStyle = 'bg-stone-500/10 text-stone-400 border-stone-500/20';
                        }

                        return (
                            <div className={`px-3 py-1.5 rounded-full text-[10px] font-black uppercase tracking-widest border flex items-center gap-1.5 ${tierStyle}`}>
                                <Cpu size={12}/> {tierLabel}
                            </div>
                        );
                    })()}

                    {/* 2. BADGE ĐIỂM RỦI RO (RISK SCORE) */}
                    <div className={`px-3 py-1.5 rounded-full border text-[10px] font-black tracking-wider flex items-center gap-1.5 ${
                        agent.risk_score >= 80 ? 'bg-red-500/10 border-red-500/50 text-red-500 shadow-[0_0_15px_rgba(239,68,68,0.2)]' :
                        agent.risk_score >= 50 ? 'bg-orange-500/10 border-orange-500/50 text-orange-400' :
                        agent.risk_score >= 20 ? 'bg-yellow-500/10 border-yellow-500/50 text-yellow-400' :
                        'bg-emerald-500/10 border-emerald-500/30 text-emerald-400'
                    }`}>
                        <ShieldAlert size={14} /> 
                        RISK SCORE: {agent.risk_score || 0}
                    </div>

                    {/* 3. BADGE PHÊ DUYỆT ZERO TRUST */}
                    <div className={`px-3 py-1.5 rounded-full text-[10px] font-black uppercase tracking-widest border ${
                        agent.status === 'ACTIVE' ? 'bg-blue-500/10 text-blue-400 border-blue-500/20' :
                        agent.status === 'PENDING' ? 'bg-amber-500/10 text-amber-400 border-amber-500/20' :
                        'bg-red-500/10 text-red-400 border-red-500/20'
                    }`}>
                        {agent.status === 'ACTIVE' ? 'Hợp lệ' : agent.status === 'PENDING' ? 'Chờ duyệt' : 'Bị cấm'}
                    </div>

                    {/* 4. BADGE TRỰC TUYẾN / NGOẠI TUYẾN */}
                    <div className={`px-3 py-1.5 rounded-full text-[10px] font-bold uppercase tracking-widest border flex items-center gap-2 ${
                        isOnline ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-slate-800 text-slate-400 border-slate-700'
                    }`}>
                        <div className={`w-1.5 h-1.5 rounded-full ${isOnline ? 'bg-emerald-400 animate-pulse' : 'bg-slate-500'}`}></div>
                        {isOnline ? 'Trực tuyến' : 'Ngoại tuyến'}
                    </div>
                </div>
            </div>

            {/* --- GRID CHÍNH: 3 CỘT (Thông tin - Phần mềm - USB) --- */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start mb-8">
                
                {/* CỘT 1: CẤU HÌNH & QUẢN LÝ */}
                <div className="space-y-6">
                    
                    {/* Card Cấu hình */}
                    <div className="bg-[#1e293b] p-5 rounded-3xl border border-slate-800 shadow-xl">
                        <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 mb-4 tracking-widest">
                            <Cpu size={14}/> Cấu hình phần cứng
                        </h3>
                        <div className="space-y-3">
                            <InfoRow label="IP Address" value={agent.ip_address} />
                            <InfoRow label="CPU Model" value={inv.cpu_model} />
                            <InfoRow label="RAM Total" value={`${inv.ram_total_gb} GB`} />
                            <InfoRow label="OS System" value={inv.os_info} />
                        </div>
                    </div>

                    {/* Card Người quản lý */}
                    <div className="bg-[#1e293b] p-5 rounded-3xl border border-slate-800 shadow-xl relative overflow-hidden">
                        <div className="absolute top-0 right-0 p-3 opacity-10 text-emerald-500"><User size={80}/></div>
                        <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 mb-4 tracking-widest z-10 relative">
                            <User size={14}/> Người chịu trách nhiệm
                        </h3>
                        
                        {agent.manager ? (
                            <div className="relative z-10">
                                <p className="text-lg font-bold text-white mb-1">{agent.manager.full_name}</p>
                                <div className="space-y-1.5 mt-3">
                                    <div className="flex items-center gap-2 text-xs text-slate-400 bg-slate-900/50 p-2 rounded-lg border border-slate-700/50">
                                        <Smartphone size={12} className="text-blue-400"/> {agent.manager.phone}
                                    </div>
                                    <div className="flex items-center gap-2 text-xs text-slate-400 bg-slate-900/50 p-2 rounded-lg border border-slate-700/50">
                                        <Mail size={12} className="text-emerald-400"/> {agent.manager.email || agent.manager.username}
                                    </div>
                                </div>
                            </div>
                        ) : (
                            <div className="py-4 text-center border-2 border-dashed border-slate-700 rounded-xl bg-slate-900/30">
                                <p className="text-xs text-slate-500 font-bold">Chưa bàn giao</p>
                            </div>
                        )}
                    </div>
                </div>

                {/* CỘT 2: DANH SÁCH PHẦN MỀM */}
                <div>
                    <AgentSoftware software={agent.software || []} />
                </div>

                {/* CỘT 3: DANH SÁCH USB */}
                <div>
                    <AgentUSB usbLogs={agent.usb_logs || []} />
                </div>
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
        <span className="text-white font-mono truncate max-w-[60%]">{value}</span>
    </div>
);

export default AgentDetail;