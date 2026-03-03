import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Monitor, Cpu, ArrowLeft, User, Smartphone, Mail, ShieldAlert } from 'lucide-react';
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

    return (
        <div className="text-slate-200 p-6 pb-20 max-w-[1600px] mx-auto">
            {/* Header + Back Button */}
            <div className="flex items-center gap-4 mb-6">
                <button onClick={() => navigate('/agents')} className="p-2 bg-slate-800 rounded-xl hover:bg-slate-700 transition text-slate-400 hover:text-white">
                    <ArrowLeft size={20}/>
                </button>
                <div>
                    <h1 className="text-2xl font-black text-white">{agent.hostname}</h1>
                    <p className="text-xs text-slate-500 font-mono mt-0.5">{agent.hwid}</p>
                </div>
                <div className="ml-auto flex items-center gap-2">
                     <div className={`px-3 py-1 rounded-lg border text-[10px] font-bold uppercase tracking-wider ${agent.status === 'online' ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400' : 'bg-red-500/10 border-red-500/30 text-red-400'}`}>
                        {agent.status}
                     </div>
                </div>
            </div>

            {/* GRID CHÍNH: 3 CỘT (Thông tin - Phần mềm - USB) */}
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

                    {/* Card Người quản lý (Nằm ngay dưới Cấu hình) */}
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

            {/* PHẦN DƯỚI: NHẬT KÝ CẢNH BÁO (Full Width) */}
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