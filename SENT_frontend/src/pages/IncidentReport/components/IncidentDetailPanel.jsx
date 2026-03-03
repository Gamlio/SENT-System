import React from 'react';
import { X, Clock, Monitor, Cpu, Terminal, CheckCircle } from 'lucide-react';

const IncidentDetailPanel = ({ incident, onClose, onResolve }) => {
    if (!incident) return null;

    return (
        <div className="absolute inset-0 z-50 flex justify-end bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
            <div className="w-full max-w-2xl h-full bg-[#1e293b] border-l border-slate-700 shadow-2xl flex flex-col animate-in slide-in-from-right duration-300">
                
                {/* --- HEADER --- */}
                <div className="p-6 border-b border-slate-700 flex justify-between items-start bg-slate-900">
                    <div>
                        <h2 className="text-xl font-bold text-white flex items-center gap-2">
                            CASE #{incident.id}: {incident.type}
                        </h2>
                        <p className="text-sm text-slate-400 mt-1 flex items-center gap-2">
                            <Monitor size={14}/> {incident.agent?.hostname} 
                            <span className="text-slate-600">|</span> 
                            <Clock size={14}/> {new Date(incident.created_at).toLocaleString()}
                        </p>
                    </div>
                    <button onClick={onClose} className="p-2 hover:bg-slate-800 rounded-full text-slate-400 hover:text-white transition">
                        <X size={24} />
                    </button>
                </div>

                {/* --- BODY (SCROLLABLE) --- */}
                <div className="flex-1 overflow-y-auto p-6 space-y-6 custom-scrollbar">
                    
                    {/* Phần 1: AI Analysis */}
                    <div className="bg-blue-900/20 border border-blue-500/30 rounded-xl p-4">
                        <h3 className="text-blue-400 font-bold text-sm mb-2 flex items-center gap-2">
                            <Cpu size={16}/> Phân tích tự động (AI Analysis)
                        </h3>
                        <p className="text-sm text-blue-100 leading-relaxed">
                            {incident.ai_analysis || "Hệ thống đã tự động gom nhóm các cảnh báo liên quan đến máy trạm này."}
                        </p>
                    </div>

                    {/* Phần 2: Timeline Alerts */}
                    <div>
                        <h3 className="text-white font-bold text-sm mb-4 flex items-center gap-2">
                            <Terminal size={16}/> Nhật ký Cảnh báo ({incident.alerts?.length || 0})
                        </h3>
                        <div className="space-y-3 pl-2 border-l-2 border-slate-700 ml-2">
                            {incident.alerts?.map((alert, idx) => (
                                <div key={idx} className="relative pl-6 pb-2">
                                    <div className="absolute -left-[9px] top-1 w-4 h-4 rounded-full bg-slate-900 border-2 border-red-500"></div>
                                    <div className="bg-slate-800/50 p-3 rounded-lg border border-slate-700 hover:bg-slate-800 transition">
                                        <div className="flex justify-between items-start">
                                            <span className="text-red-400 font-bold text-xs">{alert.alert_type}</span>
                                            <span className="text-[10px] text-slate-500 font-mono">{new Date(alert.created_at).toLocaleTimeString()}</span>
                                        </div>
                                        <p className="text-sm text-slate-300 mt-1">{alert.description}</p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>

                {/* --- FOOTER ACTIONS --- */}
                <div className="p-4 border-t border-slate-700 bg-slate-900 flex justify-end gap-3">
                    <button onClick={onClose} className="px-4 py-2 rounded-lg text-sm font-bold text-slate-400 hover:bg-slate-800 transition">
                        Đóng
                    </button>
                    {incident.status !== 'Resolved' && (
                        <button 
                            onClick={() => onResolve(incident.id)}
                            className="px-4 py-2 rounded-lg text-sm font-bold bg-emerald-600 hover:bg-emerald-500 text-white shadow-lg transition flex items-center gap-2"
                        >
                            <CheckCircle size={16}/> Đánh dấu Đã xử lý
                        </button>
                    )}
                </div>
            </div>
        </div>
    );
};

export default IncidentDetailPanel;