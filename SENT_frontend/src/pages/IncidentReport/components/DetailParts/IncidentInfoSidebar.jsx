import React from 'react';
import { Monitor, Cpu, Network, ShieldAlert } from 'lucide-react';
import IncidentPlaybook from './IncidentPlaybook';

const IncidentInfoSidebar = ({ incident, onUpdate }) => {
    const agent = incident.agent || incident.Agent || {};
    const alerts = incident.alerts || incident.Alerts || [];

    return (
        <div className="w-[450px] border-r border-slate-800 bg-[#0A101D]/50 overflow-y-auto custom-scrollbar p-5 shrink-0 flex flex-col gap-6">
            
            {/* THÔNG TIN THIẾT BỊ NẠN NHÂN */}
            <div>
                <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-3 flex items-center gap-2 border-b border-slate-800 pb-2">
                    <Monitor size={12} className="text-emerald-400"/> Asset Fingerprint
                </h3>
                <div className="bg-[#111827] p-4 rounded-lg border border-slate-800/80 grid grid-cols-2 gap-4">
                    <div className="col-span-2 flex justify-between items-start">
                        <div>
                            <p className="text-[9px] text-slate-500 uppercase font-bold mb-0.5">Hostname</p>
                            <p className="text-sm text-white font-bold truncate flex items-center gap-2" title={agent.hostname}>
                                {agent.hostname || 'Unknown'}
                                <span className={`w-1.5 h-1.5 rounded-full ${agent.status === 'online' ? 'bg-emerald-500 shadow-[0_0_5px_#10b981]' : 'bg-slate-600'}`}></span>
                            </p>
                        </div>
                        <div className="text-right">
                            <p className="text-[9px] text-slate-500 uppercase font-bold mb-0.5">Risk Score</p>
                            <p className={`text-xs font-black px-1.5 py-0.5 rounded border ${agent.risk_score > 70 ? 'bg-red-500/10 text-red-500 border-red-500/30' : 'bg-orange-500/10 text-orange-400 border-orange-500/30'}`}>
                                {agent.risk_score || 0}
                            </p>
                        </div>
                    </div>
                    
                    <div>
                        <p className="text-[9px] text-slate-500 uppercase font-bold mb-0.5">IP Address</p>
                        <p className="text-[11px] text-emerald-400 font-mono flex items-center gap-1 bg-[#050B14] px-1.5 py-0.5 rounded border border-slate-800 w-max">
                            <Network size={10}/> {agent.ip_address || 'N/A'}
                        </p>
                    </div>
                    <div>
                        <p className="text-[9px] text-slate-500 uppercase font-bold mb-0.5">HWID</p>
                        <p className="text-[10px] text-slate-400 font-mono truncate bg-[#050B14] px-1.5 py-0.5 rounded border border-slate-800 w-max" title={incident.agent_hw_id}>
                            {incident.agent_hw_id?.substring(0, 12)}...
                        </p>
                    </div>
                    
                    {agent.Inventory && (
                        <div className="col-span-2 pt-3 border-t border-slate-800/50 flex flex-col gap-1">
                            <p className="text-[10px] text-slate-500 uppercase font-bold">Hardware Info</p>
                            <p className="text-[11px] text-slate-400 flex items-center gap-1.5 font-mono">
                                <Cpu size={12} className="text-slate-500"/> {agent.Inventory.cpu_model}
                            </p>
                        </div>
                    )}
                </div>
            </div>

            {/* AI PLAYBOOK */}
            <IncidentPlaybook incident={incident} onUpdate={onUpdate} />

            {/* DANH SÁCH CẢNH BÁO GỐC (RAW ALERTS) */}
            <div className="flex-1">
                <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-3 flex items-center gap-2 border-b border-slate-800 pb-2">
                    <ShieldAlert size={12} className="text-red-400"/> Triggered Alerts ({alerts.length})
                </h3>
                <div className="space-y-2">
                    {alerts.length === 0 ? (
                        <p className="text-[10px] font-mono text-slate-600 italic">No raw alerts found.</p>
                    ) : alerts.map((alert, idx) => (
                        <div key={idx} className="bg-[#111827] p-3 rounded border-l-2 border-l-red-500 border-y border-r border-slate-800/50 shadow-sm relative group hover:border-r-red-500/30 hover:border-y-red-500/30 transition">
                            <div className="flex justify-between items-start mb-1.5">
                                <span className="text-[11px] font-bold text-red-400 group-hover:text-red-300 transition">{alert.alert_type}</span>
                                <span className="text-[9px] text-slate-500 font-mono">{new Date(alert.created_at || alert.CreatedAt).toLocaleTimeString()}</span>
                            </div>
                            <p className="text-[11px] text-slate-400 leading-relaxed font-mono whitespace-pre-wrap break-words">{alert.description}</p>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default IncidentInfoSidebar;