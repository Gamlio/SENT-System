import React, { useMemo } from 'react';
import { User, Activity, FileText, CheckCircle, Search, AlertTriangle } from 'lucide-react';

const API_URL = import.meta.env.VITE_API_URL;
const token = localStorage.getItem('token');

const IncidentTimeline = ({ incident }) => {
    
    const timelineEvents = useMemo(() => {
        if (!incident) return [];

        const events = [
            {
                id: 'init_event',
                type: 'CREATED',
                created_at: incident.CreatedAt || incident.created_at,
                content: `Hệ thống tự động ghi nhận cảnh báo vi phạm.`,
                user: { username: 'SENT Agent Sensor', is_system: true }
            },
            ...(incident.activities || incident.Activities || []).map(act => {
                let parsedImages = [];
                try {
                    if (act.images && typeof act.images === 'string') parsedImages = JSON.parse(act.images);
                    else if (Array.isArray(act.images)) parsedImages = act.images;
                } catch (error) { parsedImages = []; }

                return {
                    id: act.id,
                    type: act.action_type,
                    created_at: act.created_at || act.CreatedAt,
                    content: act.content,
                    user: act.user,
                    old_status: act.old_status,
                    new_status: act.new_status,
                    images: parsedImages
                };
            })
        ];

        return events.sort((a, b) => new Date(a.created_at) - new Date(b.created_at));
    }, [incident]);

    return (
        <div className="space-y-0 relative">
            {/* Trục đường ray giữa */}
            <div className="absolute left-[15px] top-2 bottom-0 w-px bg-slate-800"></div>

            {timelineEvents.map((event, idx) => {
                const isSystem = event.user?.is_system || !event.user;
                const userName = isSystem ? 'System Event' : (event.user?.username || 'Analyst');
                
                return (
                    <div key={idx} className="flex gap-4 relative pb-5 group">
                        
                        {/* ICON (Nút trên đường ray) */}
                        <div className="relative z-10 w-8 h-8 rounded border-2 border-[#050B14] shadow-md flex items-center justify-center shrink-0 mt-0.5 bg-[#0A101D] text-slate-500">
                            {event.type === 'RESOLVE' ? <CheckCircle size={14} className="text-emerald-500"/> :
                             event.type === 'INVESTIGATE' ? <Search size={14} className="text-blue-500"/> :
                             isSystem ? <AlertTriangle size={14} className="text-orange-500"/> : 
                             <User size={14} className="text-slate-400"/>}
                        </div>

                        {/* NỘI DUNG LOG */}
                        <div className="flex-1 min-w-0 bg-[#0A101D] border border-slate-800/80 p-3 rounded-lg shadow-sm group-hover:border-slate-700 transition">
                            <div className="flex justify-between items-center mb-1">
                                <span className={`text-[10px] font-black uppercase tracking-widest ${isSystem ? 'text-orange-400' : 'text-blue-400'}`}>{userName}</span>
                                <span className="text-[9px] text-slate-500 font-mono">{new Date(event.created_at).toLocaleString('vi-VN')}</span>
                            </div>

                            <div className="text-[11px] text-slate-300 whitespace-pre-wrap break-words leading-relaxed font-sans">
                                {event.content}

                                {/* HIỂN THỊ BẰNG CHỨNG (ẢNH/FILE) */}
                                {event.images && event.images.length > 0 && (
                                    <div className="mt-3 flex flex-wrap gap-2">
                                        {event.images.map((img, i) => (
                                            <a key={i} href={`${API_URL}/files/incidents/${img}?token=${token}`} target="_blank" rel="noreferrer" className="block w-20 h-14 rounded border border-slate-700 overflow-hidden hover:border-indigo-500 transition relative">
                                                <img src={`${API_URL}/files/incidents/${img}?token=${token}`} alt="evidence" className="w-full h-full object-cover opacity-80 hover:opacity-100"/>
                                            </a>
                                        ))}
                                    </div>
                                )}
                                
                                {/* LOG CHUYỂN STATE */}
                                {event.old_status !== event.new_status && event.old_status && (
                                    <div className="mt-2 pt-2 border-t border-slate-800 flex items-center gap-1.5 text-[9px] font-mono font-bold text-slate-500">
                                        <Activity size={10} className="text-slate-600"/>
                                        State changed: <span className="line-through opacity-50">{event.old_status}</span> ➜ <span className={event.new_status === 'Resolved' ? 'text-emerald-500' : 'text-blue-400'}>{event.new_status}</span>
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>
                );
            })}
        </div>
    );
};

export default IncidentTimeline;