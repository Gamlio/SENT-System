import React from 'react';
import { X, Sparkles } from 'lucide-react';
import ChatWindow from '../pages/AIChat/components/ChatWindow';
import { useChat } from '../pages/AIChat/hooks/useChat';

const GlobalCopilotDrawer = ({ isOpen, onClose }) => {
    // Tận dụng lại chính Hook bạn đã viết!
    const { messages, isLoading, sendMessage, messagesEndRef } = useChat();

    return (
        <>
            {/* Lớp nền mờ khi mở AI */}
            {isOpen && (
                <div 
                    className="fixed inset-0 bg-black/20 backdrop-blur-[2px] z-[90] transition-opacity"
                    onClick={onClose}
                />
            )}

            {/* Khung Chat Trượt từ phải sang */}
            <div className={`fixed inset-y-0 right-0 w-[400px] bg-[#0f172a] border-l border-slate-700 shadow-2xl z-[100] transform transition-transform duration-300 flex flex-col ${isOpen ? 'translate-x-0' : 'translate-x-full'}`}>
                
                {/* Header của Copilot */}
                <div className="flex items-center justify-between p-4 border-b border-slate-800 bg-[#1e293b]">
                    <div className="flex items-center gap-2">
                        <div className="p-1.5 bg-indigo-500/20 rounded-lg">
                            <Sparkles size={16} className="text-indigo-400" />
                        </div>
                        <h2 className="font-bold text-white text-sm">SENT Copilot</h2>
                    </div>
                    <button onClick={onClose} className="p-1.5 text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition">
                        <X size={18} />
                    </button>
                </div>

                {/* Tái sử dụng Component ChatWindow của bạn */}
                <div className="flex-1 overflow-hidden relative">
                    <ChatWindow 
                        messages={messages} 
                        isLoading={isLoading} 
                        onSend={sendMessage}
                        messagesEndRef={messagesEndRef}
                    />
                </div>
            </div>
        </>
    );
};

export default GlobalCopilotDrawer;