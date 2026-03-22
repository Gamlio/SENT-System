import React from 'react';
import { ShieldCheck, CheckCircle2, Laptop, Trash2, Clock, User, AlertTriangle } from 'lucide-react';

export const PolicyItem = ({ policy, onDelete, agents }) => {
    const getTargetNames = () => {
        if (!policy.target_hwids) return "Global";
        const ids = typeof policy.target_hwids === 'string' ? JSON.parse(policy.target_hwids) : policy.target_hwids;
        if (!ids || ids.length === 0) return "Global";
        return ids.map(id => agents.find(a => a.hwid === id)?.hostname || id.substring(0, 6)).join(', ');
    };

    // Xác định màu sắc dựa trên trạng thái phê duyệt[cite: 45, 60]
    const statusConfig = {
        PENDING: "border-amber-500/30 bg-amber-500/5 text-amber-500",
        APPROVED: "border-emerald-500/30 bg-emerald-500/5 text-emerald-400",
        REJECTED: "border-red-500/30 bg-red-500/5 text-red-400"
    };

    return (
        <div className="p-4 flex items-center justify-between group hover:bg-slate-800/40 transition-all border-b border-slate-800/50 last:border-0">
            <div className="flex items-center gap-5 min-w-0">
                {/* Icon loại hành động */}
                <div className={`p-3 rounded-2xl shrink-0 shadow-lg ${policy.policy_type === 'BLACKLIST' ? 'bg-red-500/20 text-red-500' : 'bg-emerald-500/20 text-emerald-500'}`}>
                    {policy.policy_type === 'BLACKLIST' ? <AlertTriangle size={22}/> : <CheckCircle2 size={22}/>}
                </div>

                <div className="min-w-0 space-y-1">
                    <div className="flex items-center gap-3">
                        <h4 className="font-black text-white text-sm tracking-tight truncate">{policy.title}</h4>
                        {/* Tag trạng thái phê duyệt[cite: 45, 58] */}
                        <span className={`text-[8px] font-black px-2 py-0.5 rounded-full border ${statusConfig[policy.approval_status] || "text-slate-400 border-slate-700"}`}>
                            {policy.approval_status}
                        </span>
                    </div>

                    <div className="flex flex-wrap items-center gap-x-4 gap-y-1">
                        <code className="text-[10px] bg-slate-900/80 px-2 py-0.5 rounded text-emerald-400 font-mono border border-slate-700">
                            {policy.value}
                        </code>
                        <span className="text-[10px] text-slate-500 flex items-center gap-1">
                            <User size={12}/> {policy.created_by || "System"}
                        </span>
                        <span className="text-[10px] text-slate-500 flex items-center gap-1">
                            <Clock size={12}/> {new Date(policy.CreatedAt).toLocaleDateString()}
                        </span>
                        {policy.target_type === 'SPECIFIC' && (
                            <span className="text-[10px] text-purple-400 flex items-center gap-1 font-bold">
                                <Laptop size={12}/> {getTargetNames()}
                            </span>
                        )}
                    </div>
                </div>
            </div>

            {/* Nút hành động nổi bật khi hover[cite: 60] */}
            <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-all transform translate-x-2 group-hover:translate-x-0">
                <button onClick={() => onDelete(policy.ID)} className="p-2.5 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-xl transition-colors shadow-sm">
                    <Trash2 size={18}/>
                </button>
            </div>
        </div>
    );
};