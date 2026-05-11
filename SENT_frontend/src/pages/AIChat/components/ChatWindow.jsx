import React, { useState, useRef, useEffect } from 'react';
import { Send, Loader2, Cpu, ShieldAlert } from 'lucide-react';
import MessageBubble from './MessageBubble';

const ChatWindow = ({ messages = [], isLoading, onSend, messagesEndRef }) => {
    const [input, setInput] = useState('');
    const textareaRef = useRef(null);

    const handleSubmit = (e) => {
        e.preventDefault();
        if (!input.trim() || isLoading) return;
        onSend(input);
        setInput('');
    };

    return (
        <div className="flex flex-col h-full">
            {/* MESSAGES LIST */}
            <div className="flex-1 overflow-y-auto p-8 space-y-8 pb-40 custom-scrollbar">
                {messages.length === 0 ? (
                    <div className="h-full flex flex-col items-center justify-center opacity-20">
                        <Cpu size={64} className="text-indigo-500 mb-4" />
                        <p className="text-[10px] font-black uppercase tracking-[0.4em]">Awaiting Command...</p>
                    </div>
                ) : (
                    messages.map((msg, idx) => <MessageBubble key={idx} message={msg} />)
                )}
                
                {isLoading && (
                    <div className="flex gap-4 animate-pulse">
                        <div className="w-10 h-10 rounded-2xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center">
                            <Loader2 size={18} className="animate-spin text-indigo-400"/>
                        </div>
                        <div className="space-y-2 mt-2">
                            <div className="h-2 w-48 bg-slate-800 rounded-full"></div>
                            <div className="h-2 w-32 bg-slate-800/50 rounded-full"></div>
                        </div>
                    </div>
                )}
                <div ref={messagesEndRef} />
            </div>

            {/* INPUT AREA - Styled like the Search Bar in Documents */}
            <div className="absolute bottom-0 left-0 right-0 p-8 bg-gradient-to-t from-[#0A101D] via-[#0A101D]/90 to-transparent">
                <div className="max-w-5xl mx-auto relative">
                    <form onSubmit={handleSubmit} className="relative group">
                        <textarea 
                            ref={textareaRef}
                            rows={1}
                            value={input}
                            onChange={(e) => setInput(e.target.value)}
                            onKeyDown={(e) => e.key === 'Enter' && !e.shiftKey && handleSubmit(e)}
                            placeholder="Nhập yêu cầu phân tích hệ thống..."
                            className="w-full bg-[#050B14] border border-slate-800 group-focus-within:border-indigo-500/50 rounded-[24px] py-5 pl-6 pr-16 text-sm outline-none transition-all shadow-2xl placeholder:text-slate-700 min-h-[64px] max-h-[200px] resize-none"
                        />
                        <button 
                            type="submit" 
                            disabled={!input.trim() || isLoading}
                            className="absolute right-3 top-3 p-3 bg-indigo-600 hover:bg-indigo-500 disabled:bg-slate-800 text-white rounded-2xl transition-all active:scale-95 shadow-lg shadow-indigo-600/20"
                        >
                            {isLoading ? <Loader2 size={20} className="animate-spin" /> : <Send size={20} />}
                        </button>
                    </form>
                    <div className="flex justify-center mt-3 gap-6">
                        <div className="flex items-center gap-1.5 opacity-40">
                            <ShieldAlert size={10} className="text-orange-500"/>
                            <span className="text-[9px] font-black uppercase tracking-widest">Security Verified</span>
                        </div>
                        <div className="flex items-center gap-1.5 opacity-40">
                            <span className="text-[9px] font-black uppercase tracking-widest text-slate-500">SENT Copilot AI v3.0</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default ChatWindow;