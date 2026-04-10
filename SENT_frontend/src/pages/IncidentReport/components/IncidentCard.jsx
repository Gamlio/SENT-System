import { ShieldAlert, Clock, User, Activity, ArrowRight } from 'lucide-react';

const IncidentCard = ({ incident, onClick }) => {
    const getPriorityStyle = (p) => {
        const styles = {
            P1: "border-red-500/50 bg-red-500/5 text-red-500",
            P2: "border-orange-500/50 bg-orange-500/5 text-orange-500",
            P3: "border-yellow-500/50 bg-yellow-500/5 text-yellow-500"
        };
        return styles[p] || "border-slate-700 bg-slate-800/50 text-slate-400";
    };

    // Dùng incident.ID (của Postgres) nếu incident.id không có
    const displayId = incident.ID || incident.id;

    return (
        <div 
            onClick={() => onClick(displayId)}
            className="group relative bg-[#0A101D] border border-slate-800 rounded-lg p-4 hover:border-indigo-500/50 transition-all cursor-pointer"
        >
            <div className="flex justify-between items-start mb-4">
                <span className={`px-2 py-0.5 rounded text-[10px] font-black uppercase border tracking-widest ${getPriorityStyle(incident.priority)}`}>
                    {incident.priority || "UNK"} | {incident.severity || "UNK"}
                </span>
                <span className="text-[10px] font-mono text-slate-500">#{displayId}</span>
            </div>

            <h4 className="text-sm font-bold text-slate-200 mb-2 truncate">{incident.type || "Undefined Event"}</h4>
            <p className="text-xs text-slate-500 line-clamp-2 mb-4 h-8">{incident.description}</p>

            <div className="grid grid-cols-2 gap-3 pt-4 border-t border-slate-800/50">
                <div className="flex items-center gap-2">
                    <div className="p-1.5 bg-slate-800 rounded">
                        <User size={12} className="text-slate-400"/>
                    </div>
                    <div className="truncate">
                        <p className="text-[9px] text-slate-500 uppercase font-bold">Assignee</p>
                        <p className="text-[10px] text-slate-300 font-medium">
                            {incident.assignee?.full_name || incident.Assignee?.username || "Unassigned"}
                        </p>
                    </div>
                </div>
                <div className="flex items-center gap-2">
                    <div className="p-1.5 bg-slate-800 rounded">
                        <Activity size={12} className="text-indigo-400"/>
                    </div>
                    <div>
                        <p className="text-[9px] text-slate-500 uppercase font-bold">Status</p>
                        <p className="text-[10px] text-indigo-400 font-bold uppercase tracking-tight">{incident.status}</p>
                    </div>
                </div>
            </div>
        </div>
    );
};
export default IncidentCard;