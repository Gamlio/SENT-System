import React, { useState } from 'react';
import { MessageSquare, Plus, Trash2, Edit2, Check, X } from 'lucide-react';

const ChatSidebar = ({ sessions, currentSessionId, onSelectSession, onNewSession, onRename, onDelete }) => {
    // State quản lý việc đang sửa tên dòng nào
    const [editingId, setEditingId] = useState(null);
    const [editTitle, setEditTitle] = useState("");

    const startEditing = (e, session) => {
        e.stopPropagation(); // Chặn click vào dòng (để không bị select)
        setEditingId(session.id);
        setEditTitle(session.title || "Cuộc trò chuyện mới");
    };

    const cancelEditing = (e) => {
        e.stopPropagation();
        setEditingId(null);
    };

    const saveTitle = (e, id) => {
        e.stopPropagation();
        if (editTitle.trim()) {
            onRename(id, editTitle);
        }
        setEditingId(null);
    };

    const handleDelete = (e, id) => {
        e.stopPropagation();
        onDelete(id);
    };

    return (
        <div className="flex flex-col h-full bg-[#1e293b]">
            <div className="p-3 border-b border-slate-800 shrink-0">
                <button 
                    onClick={onNewSession}
                    className="w-full flex items-center justify-center gap-2 bg-indigo-600 hover:bg-indigo-500 text-white py-2.5 rounded-xl transition font-bold text-xs uppercase tracking-wider shadow-lg shadow-indigo-500/20 active:scale-95"
                >
                    <Plus size={16} /> Cuộc trò chuyện mới
                </button>
            </div>

            <div className="flex-1 overflow-y-auto p-2 space-y-1 scrollbar-thin scrollbar-thumb-slate-700">
                {sessions.map(session => (
                    <div
                        key={session.id}
                        onClick={() => onSelectSession(session.id)}
                        className={`group w-full flex items-center gap-3 p-3 rounded-lg text-sm transition-all border border-transparent cursor-pointer relative ${
                            currentSessionId === session.id 
                            ? 'bg-slate-800 text-white font-bold border-slate-700 shadow-md' 
                            : 'text-slate-400 hover:bg-slate-800/50 hover:text-slate-200'
                        }`}
                    >
                        <MessageSquare size={16} className={currentSessionId === session.id ? "text-indigo-400" : "text-slate-600"} />
                        
                        {/* CHẾ ĐỘ HIỂN THỊ THƯỜNG */}
                        {editingId !== session.id ? (
                            <>
                                <div className="flex flex-col items-start overflow-hidden flex-1">
                                    <span className="truncate w-[140px] text-left">{session.title || "Phiên chưa đặt tên"}</span>
                                    <span className="text-[10px] text-slate-600 font-mono mt-0.5">
                                        {new Date(session.updated_at).toLocaleDateString('vi-VN')}
                                    </span>
                                </div>
                                
                                {/* Nút thao tác (Chỉ hiện khi Hover) */}
                                <div className="hidden group-hover:flex items-center gap-1 absolute right-2 bg-slate-800 rounded-md p-1 shadow-lg">
                                    <button onClick={(e) => startEditing(e, session)} className="p-1 hover:text-amber-400 transition"><Edit2 size={12}/></button>
                                    <button onClick={(e) => handleDelete(e, session.id)} className="p-1 hover:text-red-400 transition"><Trash2 size={12}/></button>
                                </div>
                            </>
                        ) : (
                            /* CHẾ ĐỘ CHỈNH SỬA */
                            <div className="flex items-center gap-1 flex-1 z-10" onClick={(e) => e.stopPropagation()}>
                                <input 
                                    autoFocus
                                    value={editTitle}
                                    onChange={(e) => setEditTitle(e.target.value)}
                                    className="w-full bg-slate-900 border border-slate-600 rounded px-1 py-0.5 text-xs text-white outline-none focus:border-indigo-500"
                                    onKeyDown={(e) => e.key === 'Enter' && saveTitle(e, session.id)}
                                />
                                <button onClick={(e) => saveTitle(e, session.id)} className="text-emerald-400"><Check size={14}/></button>
                                <button onClick={cancelEditing} className="text-red-400"><X size={14}/></button>
                            </div>
                        )}
                    </div>
                ))}
            </div>
        </div>
    );
};

export default ChatSidebar;