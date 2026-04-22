import React, { useState } from 'react';
import { ShieldAlert, Zap, Monitor, Clock, Terminal, ArrowUpRight, Loader2 } from 'lucide-react';
import axiosInstance from '../../../api/axios';

const BehaviorCard = ({ behavior, onEscalated }) => {
    const [isEscalating, setIsEscalating] = useState(false);

    const getPriorityStyle = (p) => {
        const styles = {
            P1: "bg-red-500/10 text-red-500 border-red-500/50",
            P2: "bg-orange-500/10 text-orange-500 border-orange-500/50",
            P3: "bg-yellow-500/10 text-yellow-500 border-yellow-500/50",
            P4: "bg-blue-500/10 text-blue-500 border-blue-500/50"
        };
        return styles[p] || "bg-slate-800 text-slate-400 border-slate-700";
    };

    const handleEscalate = async () => {
        if (!window.confirm("Xác nhận nâng cấp hành vi này thành Sự cố chính thức?")) return;
        
        setIsEscalating(true);
        try {
            // Gọi API nâng cấp
            await axiosInstance.post('/incidents/escalate', {
                alert_id: behavior.id || behavior._id
            });
            alert("Đã nâng cấp thành công!");
            onEscalated(); // Tải lại danh sách
        } catch (err) {
            alert(err.response?.data?.error || "Lỗi nâng cấp");
        } finally {
            setIsEscalating(false);
        }
    };

    return (
        <div className={`relative bg-[#0A101D] border border-slate-800 rounded-2xl p-5 hover:border-indigo-500/40 transition-all group overflow-hidden ${behavior.is_resolved ? 'opacity-75 bg-slate-900/30' : ''}`}>
            
            {/* Tag trạng thái/mức độ */}
            <div className="flex justify-between items-start mb-4">
                <span className={`px-2 py-0.5 rounded text-[10px] font-black uppercase border tracking-widest ${getPriorityStyle(behavior.priority)}`}>
                    {behavior.priority} | {behavior.severity}
                </span>
                <span className="text-[10px] font-mono text-slate-600 italic">
                    <Clock size={10} className="inline mr-1"/>
                    {new Date(behavior.created_at).toLocaleTimeString()}
                </span>
            </div>

            {/* Nội dung hành vi */}
            <h4 className="text-sm font-black text-slate-200 mb-2 truncate group-hover:text-white transition-colors">
                {behavior.title}
            </h4>
            <div className="flex items-center gap-2 text-[10px] text-indigo-400 font-bold mb-3">
                <Terminal size={12}/> {behavior.alert_type}
            </div>
            
            <p className="text-[11px] text-slate-500 line-clamp-2 h-8 leading-relaxed mb-4">
                {behavior.description}
            </p>

            {/* Thông tin thiết bị */}
            <div className="pt-4 border-t border-slate-800/50 flex items-center justify-between">
                <div className="flex items-center gap-2">
                    <div className="p-1.5 bg-slate-800/50 rounded-lg">
                        <Monitor size={12} className="text-slate-400"/>
                    </div>
                    <code className="text-[10px] text-slate-400 font-mono">
                        {behavior.asset_hwid.substring(0, 12)}...
                    </code>
                </div>

                {/* Nút Escalate - Trái tim của giao diện mới */}
                {!behavior.is_resolved ? (
                    <button 
                        onClick={handleEscalate}
                        disabled={isEscalating}
                        className="flex items-center gap-1.5 bg-indigo-600/10 hover:bg-indigo-600 text-indigo-400 hover:text-white border border-indigo-500/30 px-3 py-1.5 rounded-lg transition-all text-[10px] font-black uppercase"
                    >
                        {isEscalating ? <Loader2 size={12} className="animate-spin"/> : <Zap size={12}/>}
                        Escalate
                    </button>
                ) : (
                    <div className="flex items-center gap-2">
                        <span className="flex items-center gap-1.5 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-3 py-1.5 rounded-lg text-[10px] font-black uppercase shadow-[0_0_15px_rgba(16,185,129,0.15)]">
                            <ShieldAlert size={12} className="animate-pulse" />
                            Đã thành Sự cố
                        </span>
                        {behavior.incident_id && (
                            <a href={`/incidents/${behavior.incident_id}`} className="p-1.5 bg-slate-800/80 hover:bg-slate-700 text-slate-400 hover:text-emerald-400 rounded-lg transition-all border border-slate-700 hover:border-emerald-500/40" title="Xem chi tiết Sự cố">
                                <ArrowUpRight size={14} />
                            </a>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
};

export default BehaviorCard;