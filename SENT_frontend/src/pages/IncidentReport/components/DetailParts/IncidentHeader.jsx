// src/pages/IncidentReport/components/DetailParts/IncidentHeader.jsx
import React from 'react';
import { X, AlertTriangle, ShieldAlert } from 'lucide-react';

const IncidentHeader = ({ incident, onClose }) => {
    // Helper màu sắc
    const getBadgeColor = (p) => {
        if (p === 'P1') return 'bg-red-500/10 text-red-500 border-red-500/50 shadow-[0_0_10px_rgba(239,68,68,0.3)]';
        if (p === 'P2') return 'bg-orange-500/10 text-orange-500 border-orange-500/50';
        return 'bg-yellow-500/10 text-yellow-500 border-yellow-500/50';
    };

    return (
        <div className="p-5 border-b border-slate-800 bg-[#0f172a] flex justify-between items-start shrink-0">
            <div>
                <div className="flex items-center gap-3 mb-2">
                    <span className={`px-2.5 py-1 rounded-lg text-xs font-black uppercase border tracking-wider flex items-center gap-1.5 ${getBadgeColor(incident.priority)}`}>
                        {incident.priority === 'P1' ? <ShieldAlert size={14}/> : <AlertTriangle size={14}/>}
                        {incident.priority} - {incident.severity}
                    </span>
                    <h2 className="text-xl font-black text-white tracking-tight">CASE #{incident.ID}</h2>
                </div>
                <p className="text-sm text-slate-400 font-medium">{incident.description}</p>
            </div>
            <button onClick={onClose} className="p-2 hover:bg-slate-800 rounded-full text-slate-500 hover:text-white transition">
                <X size={24} />
            </button>
        </div>
    );
};

export default IncidentHeader;