import React, { useMemo, useEffect, useRef } from 'react';
import { User, Activity, AlertOctagon, FileText} from 'lucide-react';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8000';

const IncidentTimeline = ({ incident }) => {
    // TẠO MỎ NEO ĐỂ TỰ ĐỘNG CUỘN XUỐNG DƯỚI
    const messagesEndRef = useRef(null);

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    };

    const timelineEvents = useMemo(() => {
        if (!incident) return [];

        const events = [
            {
                id: 'init_event',
                type: 'CREATED',
                created_at: incident.CreatedAt || incident.created_at,
                content: `Hệ thống SENT phát hiện sự cố: ${incident.type || 'N/A'}`,
                user: { username: 'SENT System', is_system: true }
            },
            ...(incident.activities || incident.Activities || []).map(act => ({
                id: act.id,
                type: act.action_type,
                created_at: act.created_at,
                content: act.content,
                user: act.user,
                old_status: act.old_status,
                new_status: act.new_status,
                images: act.images // Mảng link ảnh
            }))
        ];

        return events.sort((a, b) => new Date(a.created_at) - new Date(b.created_at));
    }, [incident]);

    // GỌI HÀM CUỘN MỖI KHI TIMELINE CÓ SỰ THAY ĐỔI
    useEffect(() => {
        scrollToBottom();
    }, [timelineEvents]);

    return (
        // Đã xóa overflow-y-auto và flex-1 ở đây để nhường quyền cuộn cho component cha
        <div className="space-y-2 pb-4">
            {timelineEvents.map((event, idx) => {
                const isSystem = event.user?.is_system || !event.user;
                const userName = isSystem ? 'Hệ thống' : (event.user?.full_name || event.user?.username || 'Unknown');
                
                return (
                    <div key={idx} className="flex gap-4 animate-in fade-in slide-in-from-bottom-2 relative">
                        {/* 1. CỘT AVATAR */}
                        <div className="flex flex-col items-center">
                            <div className={`w-10 h-10 rounded-full flex items-center justify-center border-2 border-[#0f172a] shadow-lg shrink-0 z-10 ${
                                event.type === 'RESOLVE' ? 'bg-emerald-600 text-white' :
                                event.type === 'INVESTIGATE' ? 'bg-blue-600 text-white' :
                                isSystem ? 'bg-red-600 text-white' : 'bg-slate-700 text-slate-300'
                            }`}>
                                {isSystem ? <AlertOctagon size={18}/> : <User size={18}/>}
                            </div>
                            {idx !== timelineEvents.length - 1 && (
                                <div className="w-0.5 flex-1 bg-slate-800 my-1 rounded-full"></div>
                            )}
                        </div>

                        {/* 2. NỘI DUNG CHAT */}
                        <div className="flex-1 min-w-0 pb-6">
                            <div className="flex items-center gap-2 mb-1.5 pl-1">
                                <span className="text-xs font-bold text-slate-300">{userName}</span>
                                <span className="text-[10px] text-slate-500 font-mono">
                                    {new Date(event.created_at).toLocaleString('vi-VN')}
                                </span>
                            </div>

                            <div className="bg-[#1e293b] p-4 rounded-xl border border-slate-700 shadow-sm text-sm text-slate-300 whitespace-pre-wrap break-words leading-relaxed">
                                {event.content}

                                {event.images && event.images.length > 0 && (
                                    <div className="mt-4 pt-3 border-t border-slate-700/80">
                                        <p className="text-[10px] font-bold text-slate-400 uppercase mb-3 flex items-center gap-1.5">
                                            <FileText size={12}/> Bằng chứng đính kèm:
                                        </p>
                                        <div className="flex flex-wrap gap-3">
                                            {event.images.map((img, i) => (
                                                <a key={i} href={`${API_URL}${img}`} target="_blank" rel="noreferrer" className="block w-36 h-24 rounded-lg overflow-hidden border border-slate-600 hover:border-indigo-500 hover:shadow-lg transition relative group bg-black">
                                                    <img src={`${API_URL}${img}`} alt="evidence" className="w-full h-full object-contain opacity-90 group-hover:opacity-100 transition"/>
                                                </a>
                                            ))}
                                        </div>
                                    </div>
                                )}
                                
                                {event.old_status !== event.new_status && event.old_status && (
                                    <div className="mt-3 pt-3 border-t border-slate-700/50 flex items-center gap-2 text-xs font-mono font-bold text-indigo-400">
                                        <Activity size={12}/> Đổi trạng thái: <span className="text-slate-400 line-through">{event.old_status}</span> ➜ {event.new_status}
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