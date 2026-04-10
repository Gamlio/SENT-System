import React from 'react';
import { X, AlertTriangle, ShieldAlert } from 'lucide-react';

const IncidentHeader = ({ incident, onClose }) => {
    const getBadgeColor = (p) => {
        if (p === 'P1') return 'bg-red-500/10 text-red-500 border-red-500/50 shadow-[0_0_10px_rgba(239,68,68,0.2)]';
        if (p === 'P2') return 'bg-orange-500/10 text-orange-500 border-orange-500/50';
        return 'bg-yellow-500/10 text-yellow-500 border-yellow-500/50';
    };

    return (
        <div className="h-14 px-5 border-b border-slate-800 bg-[#0A101D] flex justify-between items-center shrink-0">
            <div className="flex items-center gap-4">
                <span className={`px-2.5 py-0.5 rounded text-[10px] font-black uppercase border tracking-widest flex items-center gap-1.5 ${getBadgeColor(incident.priority)}`}>
                    {incident.priority === 'P1' ? <ShieldAlert size={12}/> : <AlertTriangle size={12}/>}
                    {incident.priority} | {incident.severity}
                </span>
                <div className="h-5 w-px bg-slate-800"></div>
                <div>
                    <h2 className="text-xs font-black text-white tracking-widest uppercase flex items-center gap-2">
                        CASE #{incident.ID} <span className="text-slate-600 font-normal">|</span> <span className="text-slate-300 font-bold">{incident.type || 'N/A'}</span>
                    </h2>
                </div>
            </div>
            {onClose && (
                <button onClick={onClose} className="p-1.5 hover:bg-red-500/10 hover:text-red-500 rounded text-slate-500 transition">
                    <X size={16} />
                </button>
            )}
        </div>
    );
};

export default IncidentHeader;