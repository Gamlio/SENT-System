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
        <div className="p-6 text-slate-200 h-[calc(100vh-60px)] w-full overflow-hidden bg-[#050B14] font-sans">
            <div className="mb-4 flex justify-between items-end shrink-0">
                <div>
                    <h1 className="text-xl font-black text-white flex items-center gap-2 uppercase tracking-tight">
                        <Sparkles className="text-indigo-400" size={24}/> AI SECURITY HUB
                    </h1>
                    <p className="text-[11px] text-slate-500 mt-1 uppercase tracking-widest font-bold">Giao diện hội thoại AI thời gian thực và quản lý phiên.</p>
                </div>
                <div className="text-[10px] bg-indigo-500/10 text-indigo-400 px-2 py-1 rounded border border-indigo-500/20">Pro Mode</div>
            </div>

            <div className="flex-1 bg-[#0A101D] rounded-lg border border-slate-800 shadow-2xl overflow-hidden">
                <div className="h-full grid grid-cols-1 lg:grid-cols-12 gap-0 min-h-0">
                    <div className="hidden lg:flex lg:col-span-3 flex-col bg-[#1e293b] border-r border-slate-800 overflow-hidden">
                        <ChatSidebar 
                            sessions={sessions} 
                            currentSessionId={currentSession?.id} 
                            onSelectSession={selectSession}
                            onNewSession={createNewSession}
                            onRename={renameSession}
                            onDelete={deleteSession}
                        />
                    </div>

                    <div className="col-span-1 lg:col-span-9 h-full flex flex-col bg-[#0f172a] overflow-hidden relative">
                        <div className="px-5 py-3 border-b border-slate-800 flex justify-between items-center shrink-0 z-10 bg-[#0f172a]/90 backdrop-blur">
                            <div className="flex items-center gap-3">
                                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
                                <span className="font-bold text-white text-sm truncate max-w-[300px]">
                                    {currentSession ? currentSession.title : "Phiên làm việc mới"}
                                </span>
                            </div>
                        </div>

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
        </div>
    );
};

export default AIChatPage;