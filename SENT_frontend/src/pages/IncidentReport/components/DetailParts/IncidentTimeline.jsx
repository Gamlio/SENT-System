import React, { useMemo } from 'react';
import { User, Activity, AlertOctagon, FileText} from 'lucide-react';

// [QUAN TRỌNG] Cấu hình đường dẫn tới Backend để load ảnh
// Nếu Backend chạy port khác, hãy sửa số 8080 lại cho đúng
const API_URL = process.env.REACT_APP_API_URL ;

const IncidentTimeline = ({ incident }) => {
    
    const timelineEvents = useMemo(() => {
        if (!incident) return [];

        const events = [
            {
                id: 'init_event',
                type: 'CREATED',
                created_at: incident.CreatedAt,
                content: `Hệ thống SENT phát hiện sự cố: ${incident.type || 'N/A'}`,
                user: { username: 'SENT System', is_system: true }
            },
            // Map dữ liệu từ Backend
            ...(incident.Activities || []).map(act => ({
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

        // Sắp xếp thời gian tăng dần
        return events.sort((a, b) => new Date(a.created_at) - new Date(b.created_at));
    }, [incident]);

    return (
        <div className="flex-1 overflow-y-auto p-6 bg-[#0f172a] space-y-6">
            {timelineEvents.map((event, idx) => {
                const isSystem = event.user?.is_system || !event.user;
                const userName = isSystem ? 'Hệ thống' : (event.user?.full_name || event.user?.username || 'Unknown');
                
                return (
                    <div key={idx} className="flex gap-4 animate-in fade-in slide-in-from-bottom-2">
                        {/* Avatar */}
                        <div className="flex flex-col items-center">
                            <div className={`w-10 h-10 rounded-full flex items-center justify-center border-2 border-[#0f172a] shadow-lg shrink-0 z-10 ${
                                event.type === 'RESOLVE' ? 'bg-emerald-600 text-white' :
                                event.type === 'INVESTIGATE' ? 'bg-blue-600 text-white' :
                                isSystem ? 'bg-red-600 text-white' : 'bg-slate-700 text-slate-300'
                            }`}>
                                {isSystem ? <AlertOctagon size={18}/> : <User size={18}/>}
                            </div>
                            {idx !== timelineEvents.length - 1 && <div className="w-0.5 flex-1 bg-slate-800 my-1"></div>}
                        </div>

                        {/* Nội dung Chat & Ảnh */}
                        <div className="flex-1 pb-4">
                            <div className="flex items-center gap-2 mb-1">
                                <span className="text-xs font-bold text-slate-300">{userName}</span>
                                <span className="text-[10px] text-slate-500 font-mono">
                                    {new Date(event.created_at).toLocaleString('vi-VN')}
                                </span>
                            </div>

                            <div className="bg-[#1e293b] p-4 rounded-xl border border-slate-700 shadow-sm text-sm text-slate-300 whitespace-pre-wrap">
                                {event.content}

                                {/* [HIỂN THỊ ẢNH] */}
                                {event.images && event.images.length > 0 && (
                                    <div className="mt-3 pt-3 border-t border-slate-700">
                                        <p className="text-[10px] font-bold text-slate-500 uppercase mb-2 flex items-center gap-1">
                                            <FileText size={12}/> Bằng chứng đính kèm:
                                        </p>
                                        <div className="flex flex-wrap gap-2">
                                            {event.images.map((img, i) => (
                                                <a 
                                                    key={i} 
                                                    // Ghép API_URL vào đường dẫn ảnh
                                                    href={`${API_URL}${img}`} 
                                                    target="_blank" 
                                                    rel="noreferrer"
                                                    className="block w-32 h-24 rounded-lg overflow-hidden border border-slate-600 hover:border-indigo-500 transition relative group"
                                                >
                                                    <img 
                                                        src={`${API_URL}${img}`} 
                                                        alt="evidence" 
                                                        className="w-full h-full object-cover"
                                                        onError={(e) => {e.target.src = 'https://placehold.co/100x100?text=Error'}} 
                                                    />
                                                    <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 flex items-center justify-center transition">
                                                        <span className="text-xs font-bold text-white">Xem</span>
                                                    </div>
                                                </a>
                                            ))}
                                        </div>
                                    </div>
                                )}
                                
                                {/* Log đổi trạng thái */}
                                {event.old_status !== event.new_status && event.old_status && (
                                    <div className="mt-2 pt-2 border-t border-slate-700/50 flex items-center gap-2 text-xs font-mono text-indigo-400">
                                        <Activity size={12}/>
                                        Đổi trạng thái: {event.old_status} ➜ {event.new_status}
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