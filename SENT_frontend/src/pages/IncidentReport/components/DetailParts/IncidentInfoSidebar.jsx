// src/pages/IncidentReport/components/DetailParts/IncidentInfoSidebar.jsx
import React from 'react';
import { Monitor, Cpu, Network, ShieldAlert, Clock } from 'lucide-react';
import IncidentPlaybook from './IncidentPlaybook';
const IncidentInfoSidebar = ({ incident, onUpdate }) => { // Nhận thêm prop onUpdate
    const agent = incident.Agent || {};

    return (
        <div className="w-[400px] border-r border-slate-800 bg-[#1e293b]/30 overflow-y-auto p-5 shrink-0">
            {/* --- 1. PLAYBOOK (VỊ TRÍ VIP NHẤT) --- */}
            <IncidentPlaybook incident={incident} onUpdate={onUpdate} />
            {/* THÔNG TIN THIẾT BỊ */}
            <div className="mb-8">
                <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-3 flex items-center gap-2">
                    <Monitor size={14} className="text-blue-400"/> Device Fingerprint
                </h3>
                <div className="bg-[#1e293b] p-4 rounded-xl border border-slate-700 space-y-3 shadow-lg">
                    <div>
                        <p className="text-[10px] text-slate-500 uppercase font-bold">Hostname</p>
                        <p className="text-sm text-white font-bold truncate" title={agent.hostname}>{agent.hostname || 'Unknown'}</p>
                    </div>
                    <div className="grid grid-cols-2 gap-2">
                        <div>
                            <p className="text-[10px] text-slate-500 uppercase font-bold">IP Address</p>
                            <p className="text-xs text-emerald-400 font-mono flex items-center gap-1">
                                <Network size={10}/> {agent.ip_address || 'N/A'}
                            </p>
                        </div>
                        <div>
                            <p className="text-[10px] text-slate-500 uppercase font-bold">HWID</p>
                            <p className="text-xs text-slate-400 font-mono truncate" title={incident.agent_hw_id}>
                                {incident.agent_hw_id?.substring(0, 8)}...
                            </p>
                        </div>
                    </div>
                    {/* Thêm thông tin cấu hình nếu có */}
                    {agent.Inventory && (
                        <div className="pt-2 border-t border-slate-700/50">
                            <p className="text-xs text-slate-400 flex items-center gap-1.5">
                                <Cpu size={12}/> {agent.Inventory.cpu_model}
                            </p>
                        </div>
                    )}
                </div>
            </div>

            {/* DANH SÁCH CẢNH BÁO GỐC */}
            <div>
                <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-3 flex items-center gap-2">
                    <ShieldAlert size={14} className="text-red-400"/> Triggered Alerts ({incident.Alerts?.length || 0})
                </h3>
                <div className="space-y-2.5">
                    {incident.Alerts?.map((alert, idx) => (
                        <div key={idx} className="relative pl-4 py-1">
                            {/* Đường nối timeline nhỏ */}
                            <div className="absolute left-0 top-2 w-1.5 h-1.5 rounded-full bg-red-500 ring-4 ring-red-500/10"></div>
                            <div className="absolute left-[2.5px] top-4 bottom-[-10px] w-px bg-slate-700 last:hidden"></div>
                            
                            <div className="bg-[#1e293b] p-3 rounded-lg border border-slate-700/50 hover:border-red-500/30 transition">
                                <div className="flex justify-between items-start mb-1">
                                    <span className="text-xs font-bold text-red-400">{alert.alert_type}</span>
                                    <span className="text-[9px] text-slate-500 font-mono">{new Date(alert.created_at).toLocaleTimeString()}</span>
                                </div>
                                <p className="text-[11px] text-slate-300 leading-snug">{alert.description}</p>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default IncidentInfoSidebar;