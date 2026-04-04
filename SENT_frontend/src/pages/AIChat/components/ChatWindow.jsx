import React, { useState, useRef, useEffect } from 'react';
import { Send, Loader2 } from 'lucide-react';
import MessageBubble from './MessageBubble';

const ChatWindow = ({ messages = [], isLoading, onSend, messagesEndRef }) => {
    const [input, setInput] = useState('');
    const textareaRef = useRef(null);

    // Hàm tự động chỉnh độ cao khi gõ
    const adjustHeight = () => {
        if (textareaRef.current) {
            textareaRef.current.style.height = 'auto'; 
            textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 150)}px`;
        }
    };

    useEffect(() => {
        adjustHeight();
    }, [input]);

    const handleKeyDown = (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault(); 
            handleSubmit(e);
        }
    };

    const handleSubmit = (e) => {
        e.preventDefault();
        if (!input.trim() || isLoading) return;
        onSend(input);
        setInput('');
        if (textareaRef.current) textareaRef.current.style.height = 'auto';
    };

    return (
        <div className="flex flex-col h-full relative bg-[#0f172a]">
            
            {/* PHẦN 1: DANH SÁCH TIN NHẮN */}
            {/* pb-32 cực kỳ quan trọng để đẩy danh sách lên trên khung input */}
            <div className="flex-1 overflow-y-auto p-4 md:p-6 scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-transparent space-y-2 pb-32"> 
                
                {messages.length === 0 ? (
                    <div className="flex-1 flex flex-col items-center justify-center h-full text-slate-500 opacity-60 mt-20">
                        <div className="w-16 h-20 bg-slate-800 rounded-2xl flex items-center justify-center mb-4">
                            <Send size={24} />
                        </div>
                        <p className="text-sm font-medium">Bắt đầu trò chuyện với AI Security asset</p>
                    </div>
                ) : (
                    messages.map((msg, idx) => (
                        <MessageBubble key={idx} message={msg} />
                    ))
                )}
                
                {/* Loading Indicator */}
                {isLoading && (
                    <div className="flex gap-4 mb-4 ml-2 animate-in fade-in zoom-in duration-300">
                        <div className="w-8 h-8 rounded-full bg-indigo-600/20 flex items-center justify-center border border-indigo-500/30">
                            <Loader2 size={16} className="animate-spin text-indigo-400"/>
                        </div>
                        <div className="flex items-center gap-2 text-xs text-slate-400 italic">
                            AI đang suy luận hệ thống...
                        </div>
                    </div>
                )}
                
                {/* Dummy div để cuộn xuống đáy */}
                <div ref={messagesEndRef} className="h-4" />
            </div>

           {/* PHẦN 2: THANH NHẬP LIỆU */}
            <div className="absolute bottom-0 left-0 right-0 p-4 bg-gradient-to-t from-[#0f172a] via-[#0f172a] to-transparent z-10">
                <div className="max-w-4xl mx-auto">
                    <form onSubmit={handleSubmit} className="relative flex items-end gap-2 bg-[#1e293b] p-2 rounded-2xl border border-slate-700 focus-within:border-indigo-500 transition-all shadow-2xl">
                        
                        <textarea 
                            ref={textareaRef}
                            value={input}
                            onChange={(e) => setInput(e.target.value)}
                            onKeyDown={handleKeyDown}
                            placeholder="Hỏi về sự cố, kiểm tra mã độc hoặc phân tích log..."
                            rows={1}
                            className="w-full bg-transparent text-slate-200 text-sm py-2.5 px-4 outline-none resize-none scrollbar-thin scrollbar-thumb-slate-600"
                            style={{ minHeight: '40px', maxHeight: '120px' }}
                        />
                        
                        <button 
                            type="submit" 
                            disabled={!input.trim() || isLoading}
                            className="mb-0.5 p-2.5 bg-indigo-600 hover:bg-indigo-500 disabled:bg-slate-800 disabled:text-slate-500 disabled:border-slate-700 text-white rounded-xl transition-all active:scale-95 flex-shrink-0 border border-transparent"
                        >
                            {isLoading ? <Loader2 size={18} className="animate-spin" /> : <Send size={18} className="translate-x-[1px]" />}
                        </button>
                    </form>
                    <div className="text-center mt-2 pb-1">
                        <p className="text-[10px] text-slate-500 font-medium">SENT Copilot AI có thể cung cấp kết quả chưa chính xác. Vui lòng đối chiếu với Playbook.</p>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default ChatWindow;