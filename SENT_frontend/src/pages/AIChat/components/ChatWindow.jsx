import React, { useState, useRef, useEffect } from 'react'; // Thêm useRef, useEffect
import { Send, Loader2 } from 'lucide-react';
import MessageBubble from './MessageBubble';

const ChatWindow = ({ messages, isLoading, onSend, messagesEndRef }) => {
    const [input, setInput] = useState('');
    const textareaRef = useRef(null); // Ref để chỉnh độ cao

    // Hàm tự động chỉnh độ cao khi gõ
    const adjustHeight = () => {
        if (textareaRef.current) {
            textareaRef.current.style.height = 'auto'; // Reset để tính toán lại
            textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 150)}px`; // Max cao 150px
        }
    };

    useEffect(() => {
        adjustHeight();
    }, [input]);

    const handleKeyDown = (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault(); // Chặn xuống dòng khi Enter
            handleSubmit(e);
        }
    };

    const handleSubmit = (e) => {
        e.preventDefault();
        onSend(input);
        setInput('');
        if (textareaRef.current) textareaRef.current.style.height = 'auto'; // Reset chiều cao
    };

    return (
        <div className="flex flex-col h-full relative bg-[#0f172a]">
            
            {/* PHẦN 1: DANH SÁCH TIN NHẮN (Cuộn độc lập) */}
            <div className="flex-1 overflow-y-auto p-6 scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-transparent space-y-2 pb-24"> 
                {/* pb-24: Để tin nhắn cuối không bị che bởi thanh Input */}
                
                {messages.length === 0 ? (
                    <div className="flex-1 overflow-y-auto p-4 space-y-6 pb-32">
                            <div className="w-16 h-16 bg-slate-800 rounded-2xl flex items-center justify-center mb-4">
                            <Send size={32} />
                        </div>
                        <p className="text-sm font-medium">Bắt đầu trò chuyện với AI Security Agent</p>
                    </div>
                ) : (
                    messages.map((msg, idx) => (
                        <MessageBubble key={idx} message={msg} />
                    ))
                )}
                
                {/* Loading Indicator */}
                {isLoading && (
                    <div className="flex gap-4 mb-4 ml-2 animate-pulse">
                        <div className="w-8 h-8 rounded-full bg-indigo-600/20 flex items-center justify-center border border-indigo-500/30">
                            <Loader2 size={16} className="animate-spin text-indigo-400"/>
                        </div>
                        <div className="flex items-center gap-2 text-xs text-slate-400 italic mt-2">
                            AI đang phân tích & suy luận...
                        </div>
                    </div>
                )}
                
                {/* Dummy div để cuộn xuống đáy */}
                <div ref={messagesEndRef} />
            </div>

           {/* PHẦN 2: THANH NHẬP LIỆU MỚI (TEXTAREA) */}
            <div className="absolute bottom-0 left-0 right-0 p-4 bg-[#0f172a] border-t border-slate-800 z-10">
                <form onSubmit={handleSubmit} className="relative flex items-end gap-2 bg-[#1e293b] p-2 rounded-2xl border border-slate-700 focus-within:border-indigo-500 transition-all shadow-lg">
                    
                    <textarea 
                        ref={textareaRef}
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        onKeyDown={handleKeyDown}
                        placeholder="Hỏi về sự cố, kiểm tra mã độc hoặc chính sách..."
                        rows={1}
                        className="w-full bg-transparent text-slate-200 text-sm py-3 px-4 outline-none resize-none scrollbar-thin scrollbar-thumb-slate-600 max-h-[150px]"
                        style={{ minHeight: '44px' }}
                    />
                    
                    <button 
                        type="submit" 
                        disabled={!input.trim() || isLoading}
                        className="mb-1 p-2.5 bg-indigo-600 hover:bg-indigo-500 disabled:bg-slate-700 disabled:opacity-50 text-white rounded-xl transition-all shadow-lg active:scale-95 flex-shrink-0"
                    >
                        {isLoading ? <Loader2 size={18} className="animate-spin" /> : <Send size={18} />}
                    </button>
                </form>
                <div className="text-center mt-2">
                     <p className="text-[10px] text-slate-600 font-medium">AI có thể mắc lỗi. Hãy kiểm chứng thông tin quan trọng.</p>
                </div>
            </div>
        </div>
    );
};

export default ChatWindow;