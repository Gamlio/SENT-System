import React, { useState, useRef, useEffect } from 'react';
import { Send, Bot, User, ShieldCheck, Sparkles, FileText, AlertTriangle, Command } from 'lucide-react';

const AIChatAssistant = () => {
    const [messages, setMessages] = useState([
        { 
            sender: 'ai', 
            text: 'Xin chào! Tôi là AI Security Agent của hệ thống SENT. Tôi đã được nạp các chính sách bảo mật và tiêu chuẩn ISO 27001. Bạn cần tôi hỗ trợ phân tích log hệ thống hay giải đáp nội quy hôm nay?' 
        }
    ]);
    const [input, setInput] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const messagesEndRef = useRef(null);

    // Tự động cuộn xuống tin nhắn mới nhất
    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    };
    useEffect(() => { scrollToBottom(); }, [messages, isLoading]);

    // Các câu hỏi gợi ý nhanh (Quick Prompts)
    const suggestions = [
        { icon: <FileText size={16}/>, text: "Tóm tắt quy định sử dụng USB" },
        { icon: <AlertTriangle size={16}/>, text: "Phân tích cảnh báo mức độ High" },
        { icon: <ShieldCheck size={16}/>, text: "Tiêu chuẩn ISO 27001 về truy cập" }
    ];

    const handleSend = async (textToSend) => {
        const query = textToSend || input;
        if (!query.trim()) return;

        // Thêm tin nhắn user
        setMessages(prev => [...prev, { sender: 'user', text: query }]);
        setInput('');
        setIsLoading(true);

        // TODO: Mở comment đoạn này khi nối với API RAG Python
        /*
        try {
            const res = await axios.post('/api/v1/ai/chat', { query: query });
            setMessages(prev => [...prev, { sender: 'ai', text: res.data.answer }]);
        } catch (error) {
            setMessages(prev => [...prev, { sender: 'ai', text: 'Lỗi kết nối đến Server AI. Vui lòng thử lại!' }]);
        }
        */

        // Giả lập AI đang suy nghĩ (Xóa phần này khi có API thật)
        setTimeout(() => {
            setMessages(prev => [...prev, { 
                sender: 'ai', 
                text: 'Đây là dữ liệu mô phỏng. Khi hệ thống Backend Python (RAG) được kích hoạt, tôi sẽ trích xuất trực tiếp từ các file tài liệu mà Admin đã tải lên cơ sở dữ liệu PostgreSQL để trả lời bạn một cách chính xác nhất.' 
            }]);
            setIsLoading(false);
        }, 1500);
    };

    return (
        <div className="flex flex-col h-[85vh] text-slate-200">
            {/* Header: Chuyên nghiệp, chuẩn phong cách SOC */}
            <header className="mb-6 flex items-center justify-between bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl">
                <div className="flex items-center gap-4">
                    <div className="relative">
                        <div className="p-3 bg-gradient-to-br from-emerald-500 to-teal-600 rounded-2xl text-white shadow-lg shadow-emerald-500/30">
                            <Bot size={28} />
                        </div>
                        <span className="absolute -bottom-1 -right-1 flex h-4 w-4">
                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                            <span className="relative inline-flex rounded-full h-4 w-4 bg-emerald-500 border-2 border-[#1e293b]"></span>
                        </span>
                    </div>
                    <div>
                        <h1 className="text-2xl font-black text-white flex items-center gap-2">
                            SENT Copilot <Sparkles size={18} className="text-amber-400"/>
                        </h1>
                        <p className="text-slate-400 text-sm font-medium">Powered by Gemini 2.5 & RAG Architecture</p>
                    </div>
                </div>
                <div className="hidden md:flex items-center gap-2 bg-slate-900/50 px-4 py-2 rounded-xl border border-slate-700">
                    <Command size={16} className="text-slate-500"/>
                    <span className="text-xs text-slate-400 font-mono">Trạng thái: Sẵn sàng</span>
                </div>
            </header>

            {/* Main Chat Area */}
            <div className="flex-1 bg-[#1e293b] rounded-3xl border border-slate-800 shadow-2xl overflow-hidden flex flex-col relative">
                
                {/* Vùng hiển thị tin nhắn */}
                <div className="flex-1 overflow-y-auto p-6 space-y-6 scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-transparent">
                    {messages.map((msg, i) => (
                        <div key={i} className={`flex gap-4 ${msg.sender === 'user' ? 'flex-row-reverse' : ''} animate-in fade-in slide-in-from-bottom-4 duration-500`}>
                            {/* Avatar */}
                            <div className={`p-3 rounded-2xl flex-shrink-0 h-fit ${msg.sender === 'user' ? 'bg-slate-800 text-white' : 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'}`}>
                                {msg.sender === 'user' ? <User size={20}/> : <ShieldCheck size={20}/>}
                            </div>
                            
                            {/* Bong bóng Chat */}
                            <div className={`max-w-[75%] p-5 rounded-3xl text-sm leading-relaxed ${
                                msg.sender === 'user' 
                                ? 'bg-gradient-to-br from-slate-700 to-slate-800 text-white rounded-tr-sm border border-slate-600 shadow-lg' 
                                : 'bg-slate-900 border border-slate-700/50 text-slate-300 rounded-tl-sm shadow-md'
                            }`}>
                                {msg.text}
                            </div>
                        </div>
                    ))}
                    
                    {/* Hiệu ứng AI đang gõ */}
                    {isLoading && (
                        <div className="flex gap-4 animate-in fade-in">
                            <div className="p-3 rounded-2xl bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 h-fit"><ShieldCheck size={20}/></div>
                            <div className="max-w-[70%] p-5 rounded-3xl rounded-tl-sm bg-slate-900 border border-slate-700 text-slate-400 flex items-center gap-2">
                                <span className="w-2 h-2 bg-emerald-500 rounded-full animate-bounce"></span>
                                <span className="w-2 h-2 bg-emerald-500 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></span>
                                <span className="w-2 h-2 bg-emerald-500 rounded-full animate-bounce" style={{ animationDelay: '0.4s' }}></span>
                            </div>
                        </div>
                    )}
                    <div ref={messagesEndRef} />
                </div>

                {/* Vùng Gợi ý & Input (Dính dưới đáy) */}
                <div className="p-4 bg-slate-800/80 backdrop-blur-md border-t border-slate-700">
                    
                    {/* Gợi ý câu hỏi (Chỉ hiện khi chưa có nhiều tin nhắn) */}
                    {messages.length <= 2 && (
                        <div className="flex flex-wrap gap-2 mb-4">
                            {suggestions.map((sug, idx) => (
                                <button 
                                    key={idx}
                                    onClick={() => handleSend(sug.text)}
                                    className="flex items-center gap-2 text-xs bg-slate-900 hover:bg-emerald-500/10 hover:text-emerald-400 hover:border-emerald-500/30 text-slate-400 px-4 py-2 rounded-full border border-slate-700 transition-all"
                                >
                                    {sug.icon} {sug.text}
                                </button>
                            ))}
                        </div>
                    )}

                    {/* Khung nhập liệu */}
                    <form 
                        onSubmit={(e) => { e.preventDefault(); handleSend(input); }} 
                        className="relative flex items-center group"
                    >
                        <input 
                            type="text" 
                            value={input} 
                            onChange={(e) => setInput(e.target.value)}
                            placeholder="Hỏi AI về chính sách bảo mật, luật lệ hoặc phân tích mã độc..." 
                            className="w-full bg-slate-900 border border-slate-700 focus:border-emerald-500 rounded-2xl py-4 pl-6 pr-16 text-white outline-none transition-all shadow-inner"
                        />
                        <button 
                            type="submit" 
                            disabled={isLoading || !input.trim()} 
                            className="absolute right-2 p-3 bg-emerald-500 hover:bg-emerald-600 disabled:bg-slate-700 disabled:text-slate-500 text-white rounded-xl transition-all shadow-lg shadow-emerald-500/20 disabled:shadow-none"
                        >
                            <Send size={18} className={input.trim() && !isLoading ? "translate-x-0.5 -translate-y-0.5 transition-transform" : ""} />
                        </button>
                    </form>
                    <p className="text-center text-[10px] text-slate-500 mt-3 font-medium">
                        AI có thể tạo ra thông tin không chính xác. Hãy luôn kiểm chứng với chính sách gốc của hệ thống.
                    </p>
                </div>
            </div>
        </div>
    );
};

export default AIChatAssistant;