import React, { useState } from 'react';
import { MessageSquare, Plus, Trash2, Edit2, Check, X, Clock } from 'lucide-react';

const ChatSidebar = ({ sessions, currentSessionId, onSelectSession, onNewSession, onRename, onDelete }) => {
    const [editingId, setEditingId] = useState(null);
    const [editTitle, setEditTitle] = useState("");

    return (
        <div className="flex flex-col h-full overflow-hidden">
            <div className="p-6 shrink-0">
                <button 
                    onClick={onNewSession}
                    className="w-full flex items-center justify-center gap-2 bg-indigo-600 hover:bg-indigo-500 text-white py-4 rounded-2xl font-black text-[11px] uppercase tracking-widest transition-all shadow-lg shadow-indigo-600/20 active:scale-95"
                >
                    <Plus size={16} /> Phiên mới
                </button>
            </div>

            <div className="flex-1 overflow-y-auto px-4 pb-6 space-y-3 custom-scrollbar">
                {sessions.map(session => (
                    <div
                        key={session.id}
                        onClick={() => onSelectSession(session.id)}
                        className={`group relative p-4 rounded-2xl border transition-all cursor-pointer ${
                            currentSessionId === session.id 
                            ? 'bg-indigo-600/10 border-indigo-500/50 shadow-inner' 
                            : 'bg-[#050B14]/50 border-slate-800 hover:border-slate-700'
                        }`}
                    >
                        {editingId !== session.id ? (
                            <div className="flex flex-col gap-2">
                                <div className="flex items-center gap-3">
                                    <MessageSquare size={14} className={currentSessionId === session.id ? "text-indigo-400" : "text-slate-600"} />
                                    <span className={`text-xs font-bold truncate flex-1 ${currentSessionId === session.id ? 'text-white' : 'text-slate-400'}`}>
                                        {session.title || "Chưa đặt tên"}
                                    </span>
                                </div>
                                <div className="flex items-center justify-between">
                                    <span className="text-[9px] font-black text-slate-600 uppercase tracking-widest flex items-center gap-1">
                                        <Clock size={10}/> {new Date(session.updated_at).toLocaleDateString('vi-VN')}
                                    </span>
                                    <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                        <button onClick={(e) => { e.stopPropagation(); setEditingId(session.id); setEditTitle(session.title); }} className="p-1.5 hover:text-indigo-400 text-slate-600 transition-colors">
                                            <Edit2 size={12}/>
                                        </button>
                                        <button onClick={(e) => { e.stopPropagation(); onDelete(session.id); }} className="p-1.5 hover:text-red-400 text-slate-600 transition-colors">
                                            <Trash2 size={12}/>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        ) : (
                            <div className="flex items-center gap-1" onClick={e => e.stopPropagation()}>
                                <input 
                                    autoFocus
                                    value={editTitle}
                                    onChange={e => setEditTitle(e.target.value)}
                                    className="w-full bg-slate-900 border border-indigo-500/50 rounded-lg px-2 py-1 text-[11px] text-white outline-none"
                                    onKeyDown={e => e.key === 'Enter' && onRename(session.id, editTitle)}
                                />
                                <button onClick={() => {onRename(session.id, editTitle); setEditingId(null)}} className="text-emerald-400"><Check size={14}/></button>
                            </div>
                        )}
                    </div>
                ))}
            </div>
        </div>
    );
};

export default ChatSidebar;