import React, { useState } from 'react';
import { ShieldAlert, Zap, Monitor, Clock, Terminal, Search, Loader2, ArrowUpRight } from 'lucide-react';
import axiosInstance from '../../../api/axios';
import BehaviorDetailModal from './BehaviorDetail';

const BehaviorCard = ({ behavior, onEscalated }) => {
    const [isEscalating, setIsEscalating] = useState(false);
    const [isModalOpen, setIsModalOpen] = useState(false);

    const getPriorityStyle = (p) => {
        const styles = {
            P1: "text-red-500 bg-red-500/10 border-red-500/20",
            P2: "text-orange-500 bg-orange-500/10 border-orange-500/20",
            P3: "text-yellow-500 bg-yellow-500/10 border-yellow-500/20",
        };
        return styles[p] || "text-slate-400 bg-slate-800";
    };

    return (
                <>
                <div className={`group flex flex-col md:flex-row items-center gap-4 bg-[#0A101D] border border-slate-800 p-4 rounded-xl hover:border-indigo-500/40 transition-all ${behavior.is_resolved ? 'opacity-60' : ''}`}>
                    
            <div className={`w-full md:w-24 text-center py-1 rounded text-[10px] font-black border uppercase tracking-tighter ${getPriorityStyle(behavior.priority)}`}>
                {behavior.priority} | {behavior.severity}
            </div>

            <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                    <h4 className="text-sm font-bold text-slate-200 truncate group-hover:text-white">
                        {behavior.title}
                    </h4>
                    <span className="text-[9px] text-indigo-400 font-mono px-2 py-0.5 bg-indigo-500/5 border border-indigo-500/10 rounded uppercase">
                        {behavior.alert_type}
                    </span>
                </div>
                <p className="text-[11px] text-slate-500 truncate italic">
                    {behavior.description}
                </p>
            </div>

            <div className="flex flex-row md:flex-col items-center md:items-end gap-4 md:gap-1 text-[11px] font-mono whitespace-nowrap md:w-48">
                <div className="flex items-center gap-1.5 text-indigo-300">
                    <Monitor size={12}/>
                    <span>{behavior.asset_hwid.substring(0, 16)}</span>
                </div>
                <div className="flex items-center gap-1.5 text-slate-500">
                    <Clock size={12}/>
                    <span>{new Date(behavior.created_at).toLocaleString('vi-VN')}</span>
                </div>
            </div>

            {/* 4. Nhóm nút thao tác - Khớp w-20 */}
            <div className="flex items-center justify-end gap-2 pl-4 border-l border-slate-800 md:w-20">
                <button 
                    onClick={() => setIsModalOpen(true)}
                    className="p-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-lg transition-colors border border-slate-700"
                    title="Xem chi tiết"
                >
                    <Search size={14} />
                </button>

                {behavior.is_resolved && (
                    <div className="p-2 bg-emerald-500/10 text-emerald-400 rounded-lg" title="Đã lập hồ sơ">
                        <ShieldAlert size={16} />
                    </div>
                )}
            </div>
        </div>

        {isModalOpen && (
            <BehaviorDetailModal 
                behaviorId={behavior.id} 
                onClose={() => setIsModalOpen(false)} 
            />
        )}
        </>
    );
};

export default BehaviorCard;