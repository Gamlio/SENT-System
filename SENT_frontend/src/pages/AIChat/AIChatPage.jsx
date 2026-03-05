import React from 'react';
import { useChat } from './hooks/useChat';
import ChatWindow from './components/ChatWindow';
import ChatSidebar from './components/ChatSidebar';
import { Sparkles } from 'lucide-react';

const AIChatPage = () => {
    const { 
        messages, isLoading, sendMessage, messagesEndRef,
        sessions, currentSession, createNewSession, selectSession,
        renameSession, deleteSession 
    } = useChat();

    return (
        
        <div className="relative flex flex-col h-[calc(100vh-100px)] w-full overflow-hidden bg-[#0f172a] rounded-2xl shadow-2xl border border-slate-800">
            
            {/* Header của khung Chat */}
            <div className="flex items-center justify-between px-6 py-3 shrink-0 bg-[#0f172a] border-b border-slate-800/50">
                <h1 className="text-lg font-bold text-white flex items-center gap-2">
                    <Sparkles className="text-indigo-400" size={20} /> AI Security Hub
                </h1>
                <div className="text-[10px] bg-indigo-500/10 text-indigo-400 px-2 py-1 rounded border border-indigo-500/20">
                    Pro Mode
                </div>
            </div>

            {/* Body chứa Sidebar và Chat */}
            <div className="flex-1 grid grid-cols-1 lg:grid-cols-12 gap-0 min-h-0">
                
                {/* Sidebar (Cột trái) */}
                <div className="hidden lg:flex lg:col-span-3 h-full flex-col bg-[#1e293b] border-r border-slate-800 overflow-hidden">
                    <ChatSidebar 
                        sessions={sessions} 
                        currentSessionId={currentSession?.id} 
                        onSelectSession={selectSession}
                        onNewSession={createNewSession}
                        onRename={renameSession}
                        onDelete={deleteSession}
                    />
                </div>

                {/* Chat Window (Cột phải) */}
                <div className="col-span-1 lg:col-span-9 h-full flex flex-col bg-[#0f172a] overflow-hidden relative">
                    {/* Header Phiên Chat */}
                    <div className="px-5 py-3 border-b border-slate-800 flex justify-between items-center shrink-0 z-10 bg-[#0f172a]/80 backdrop-blur">
                        <div className="flex items-center gap-3">
                            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
                            <span className="font-bold text-white text-sm truncate max-w-[300px]">
                                {currentSession ? currentSession.title : "Phiên làm việc mới"}
                            </span>
                        </div>
                    </div>

                    {/* Chat Content */}
                    <div className="flex-1 relative min-h-0">
                        <ChatWindow 
                            messages={messages} 
                            isLoading={isLoading} 
                            onSend={sendMessage}
                            messagesEndRef={messagesEndRef}
                        />
                    </div>
                </div>
            </div>
        </div>
    );
};

export default AIChatPage;