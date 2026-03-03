import React from 'react';
import { ShieldCheck, CheckCircle2, Laptop, Trash2 } from 'lucide-react'; // Đã xóa List, Usb, Activity

export const StatCard = ({ icon, label, value, color, bg }) => (
    <div className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl flex items-center gap-5">
        <div className={`p-4 rounded-2xl ${bg} ${color}`}>{icon}</div>
        <div>
            <p className="text-slate-400 text-xs font-bold uppercase">{label}</p>
            <h3 className="text-3xl font-black text-white">{value}</h3>
        </div>
    </div>
);

export const PolicyItem = ({ policy, onDelete, agents }) => {
    // Helper tìm tên máy từ ID
    const getTargetNames = () => {
        if (!policy.target_hwids) return "Unknown";
        const ids = typeof policy.target_hwids === 'string' ? JSON.parse(policy.target_hwids) : policy.target_hwids;
        if (!ids || ids.length === 0) return "Không có máy";
        
        const names = ids.map(id => agents.find(a => a.hwid === id)?.hostname || id.substring(0,6));
        return names.join(', ');
    };

    return (
        <div className="p-4 flex items-center justify-between group hover:bg-slate-800/50 transition">
            <div className="flex items-center gap-4 min-w-0">
                <div className={`p-2.5 rounded-xl shrink-0 ${policy.policy_type === 'BLACKLIST' ? 'bg-red-500/10 text-red-400' : 'bg-emerald-500/10 text-emerald-400'}`}>
                    {policy.policy_type === 'BLACKLIST' ? <ShieldCheck size={20}/> : <CheckCircle2 size={20}/>}
                </div>
                <div className="min-w-0">
                    <div className="flex items-center gap-2">
                        <h4 className="font-bold text-white text-sm truncate">{policy.title}</h4>
                        <span className={`text-[9px] font-black px-1.5 py-0.5 rounded uppercase ${policy.policy_type === 'BLACKLIST' ? 'text-red-400 bg-red-400/10' : 'text-emerald-400 bg-emerald-400/10'}`}>
                            {policy.policy_type}
                        </span>
                    </div>
                    <div className="flex items-center gap-2 mt-1">
                        <code className="text-[10px] bg-slate-900 px-1.5 py-0.5 rounded text-slate-300 font-mono border border-slate-700">{policy.value}</code>
                        {policy.target_type === 'SPECIFIC' && (
                            <span className="text-[10px] text-purple-400 flex items-center gap-1 truncate max-w-[200px]" title={getTargetNames()}>
                                <Laptop size={10}/> {getTargetNames()}
                            </span>
                        )}
                    </div>
                </div>
            </div>
            <button onClick={() => onDelete(policy.ID)} className="p-2 text-slate-600 hover:text-red-400 hover:bg-red-500/10 rounded-xl transition opacity-0 group-hover:opacity-100">
                <Trash2 size={18}/>
            </button>
        </div>
    );
};