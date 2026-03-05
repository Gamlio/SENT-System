import React, { useState } from 'react';
import { X, Clock, Monitor, Cpu, Terminal, CheckCircle, BookOpen, AlertOctagon, Edit3 } from 'lucide-react';

const getPriorityColor = (priority) => {
    switch (priority) {
        case 'P1': return 'bg-purple-500 text-white border-purple-400 shadow-[0_0_15px_rgba(168,85,247,0.5)]';
        case 'P2': return 'bg-red-500 text-white border-red-400';
        case 'P3': return 'bg-orange-500 text-white border-orange-400';
        default: return 'bg-slate-500 text-white border-slate-400';
    }
};

const IncidentDetailPanel = ({ incident, onClose, onResolve }) => {
    const [resolveNote, setResolveNote] = useState('');
    const [isResolving, setIsResolving] = useState(false);

    if (!incident) return null;

    const handleConfirmResolve = () => {
        if (!resolveNote.trim()) {
            alert("Vui lòng nhập ghi chú hoặc hành động đã thực hiện để đóng Case.");
            return;
        }
        onResolve(incident.id, resolveNote);
    };

    return (
        <div className="absolute inset-0 z-50 flex justify-end bg-black/60 backdrop-blur-md animate-in fade-in duration-200">
            <div className="w-full max-w-2xl h-full bg-[#0f172a] border-l border-slate-700 shadow-2xl flex flex-col animate-in slide-in-from-right duration-300">
                
                {/* --- HEADER --- */}
                <div className="p-6 border-b border-slate-800 flex justify-between items-start bg-slate-900 relative overflow-hidden">
                    <div className="absolute top-0 right-0 w-64 h-64 bg-indigo-500/5 rounded-full blur-3xl -translate-y-1/2 translate-x-1/3"></div>
                    <div className="relative z-10 w-full">
                        <div className="flex justify-between items-start mb-2">
                            <div className="flex items-center gap-3">
                                <span className={`px-3 py-1 rounded-md text-xs font-black uppercase border ${getPriorityColor(incident.priority)}`}>
                                    {incident.priority || 'P?'} - {incident.severity}
                                </span>
                                <h2 className="text-2xl font-black text-white tracking-tight">
                                    CASE #{incident.id}
                                </h2>
                            </div>
                            <button onClick={onClose} className="p-2 bg-slate-800 rounded-full text-slate-400 hover:text-white hover:bg-red-500 transition">
                                <X size={20} />
                            </button>
                        </div>
                        
                        <h3 className="text-lg text-slate-200 font-bold mb-3">{incident.type}</h3>
                        
                        <div className="flex flex-wrap items-center gap-4 text-xs font-medium text-slate-400 bg-slate-950/50 p-3 rounded-xl border border-slate-800">
                            <span className="flex items-center gap-1.5"><Monitor size={14} className="text-indigo-400"/> {incident.agent?.hostname || 'Unknown'}</span>
                            <span className="text-slate-600">|</span> 
                            <span className="flex items-center gap-1.5"><Clock size={14} className="text-emerald-400"/> {new Date(incident.created_at).toLocaleString('vi-VN')}</span>
                            <span className="text-slate-600">|</span> 
                            <span className="flex items-center gap-1.5 text-blue-400 font-bold"><BookOpen size={14}/> {incident.playbook_name || 'Chưa phân loại Playbook'}</span>
                        </div>
                    </div>
                </div>

                {/* --- BODY (SCROLLABLE) --- */}
                <div className="flex-1 overflow-y-auto p-6 space-y-6 scrollbar-thin scrollbar-thumb-slate-700">
                    
                    {/* Phần 1: AI Analysis */}
                    <div className="bg-indigo-900/10 border border-indigo-500/20 rounded-2xl p-5 relative overflow-hidden">
                        <div className="absolute top-0 left-0 w-1 h-full bg-indigo-500"></div>
                        <h3 className="text-indigo-400 font-black text-xs uppercase tracking-widest mb-3 flex items-center gap-2">
                            <Cpu size={16}/> Phân tích tự động (AI SOC)
                        </h3>
                        <p className="text-sm text-slate-300 leading-relaxed font-mono whitespace-pre-wrap">
                            {incident.ai_analysis || "Hệ thống tự động phát hiện vi phạm chính sách bảo mật dựa trên hành vi của thiết bị. Hãy tham khảo Playbook để có hướng xử lý."}
                        </p>
                    </div>

                    {/* Phần 2: Timeline Alerts */}
                    <div>
                        <h3 className="text-slate-500 font-black text-xs uppercase tracking-widest mb-4 flex items-center gap-2">
                            <AlertOctagon size={16} className="text-red-400"/> Nhật ký Cảnh báo (Alerts)
                        </h3>
                        <div className="space-y-4 pl-3 border-l-2 border-slate-800 ml-2">
                            {incident.alerts?.map((alert, idx) => (
                                <div key={idx} className="relative pl-6">
                                    <div className="absolute -left-[27px] top-2 p-1.5 rounded-full bg-slate-900 border-2 border-red-500/50">
                                        <div className="w-2 h-2 bg-red-500 rounded-full animate-pulse"></div>
                                    </div>
                                    <div className="bg-[#1e293b] p-4 rounded-xl border border-slate-700/50 hover:border-slate-600 transition shadow-lg">
                                        <div className="flex justify-between items-start mb-2">
                                            <span className="text-red-400 font-bold text-sm">{alert.title || alert.alert_type}</span>
                                            <span className="text-[10px] text-slate-500 font-mono bg-slate-900 px-2 py-1 rounded">{new Date(alert.created_at).toLocaleTimeString('vi-VN')}</span>
                                        </div>
                                        <p className="text-sm text-slate-300">{alert.description}</p>
                                    </div>
                                </div>
                            ))}
                            {(!incident.alerts || incident.alerts.length === 0) && (
                                <p className="pl-6 text-sm text-slate-500 italic">Không có chi tiết cảnh báo mạng.</p>
                            )}
                        </div>
                    </div>

                    {/* Phần 3: Lịch sử xử lý (Nếu đã Resolved) */}
                    {incident.status === 'Resolved' && (
                        <div className="bg-emerald-900/10 border border-emerald-500/20 rounded-2xl p-5 relative overflow-hidden">
                            <div className="absolute top-0 left-0 w-1 h-full bg-emerald-500"></div>
                            <h3 className="text-emerald-400 font-black text-xs uppercase tracking-widest mb-3 flex items-center gap-2">
                                <CheckCircle size={16}/> Thông tin đóng hồ sơ
                            </h3>
                            <p className="text-sm text-slate-300 leading-relaxed">
                                <strong>Ghi chú xử lý:</strong><br/>
                                {incident.resolution || "Không có ghi chú."}
                            </p>
                        </div>
                    )}
                </div>

                {/* --- FOOTER ACTIONS --- */}
                <div className="p-4 border-t border-slate-800 bg-slate-900">
                    {incident.status !== 'Resolved' ? (
                        isResolving ? (
                            <div className="animate-in slide-in-from-bottom-2">
                                <label className="flex items-center gap-2 text-xs font-bold text-emerald-400 mb-2 uppercase tracking-wide"><Edit3 size={14}/> Nhập hành động đã xử lý (Bắt buộc)</label>
                                <textarea 
                                    autoFocus
                                    className="w-full bg-[#0f172a] text-sm text-white border border-emerald-500/50 rounded-xl p-3 outline-none focus:border-emerald-400 min-h-[80px] placeholder:text-slate-600 mb-3"
                                    placeholder="VD: Đã liên hệ User, gỡ phần mềm trái phép và scan lại hệ thống..."
                                    value={resolveNote}
                                    onChange={(e) => setResolveNote(e.target.value)}
                                />
                                <div className="flex justify-end gap-3">
                                    <button onClick={() => setIsResolving(false)} className="px-4 py-2 rounded-xl text-sm font-bold text-slate-400 hover:text-white transition">Hủy</button>
                                    <button onClick={handleConfirmResolve} className="px-6 py-2 rounded-xl text-sm font-bold bg-emerald-600 hover:bg-emerald-500 text-white shadow-lg shadow-emerald-600/20 transition flex items-center gap-2">
                                        <CheckCircle size={16}/> Xác nhận đóng
                                    </button>
                                </div>
                            </div>
                        ) : (
                            <div className="flex justify-end gap-3">
                                <button onClick={onClose} className="px-6 py-2.5 rounded-xl text-sm font-bold text-slate-400 hover:bg-slate-800 transition">Đóng</button>
                                <button 
                                    onClick={() => setIsResolving(true)}
                                    className="px-6 py-2.5 rounded-xl text-sm font-black bg-indigo-600 hover:bg-indigo-500 text-white shadow-lg shadow-indigo-600/20 transition tracking-wide"
                                >
                                    TIẾN HÀNH XỬ LÝ
                                </button>
                            </div>
                        )
                    ) : (
                        <div className="flex justify-end">
                            <button onClick={onClose} className="px-6 py-2.5 rounded-xl text-sm font-bold bg-slate-800 text-white hover:bg-slate-700 transition">Đóng Panel</button>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default IncidentDetailPanel;