
import { useChat } from './hooks/useChat';
import ChatWindow from './components/ChatWindow';
import ChatSidebar from './components/ChatSidebar';
import { Sparkles, ShieldCheck, Zap } from 'lucide-react';

const AIChatPage = () => {
    const { 
        messages, isLoading, sendMessage, messagesEndRef,
        sessions, currentSession, createNewSession, selectSession,
        renameSession, deleteSession 
    } = useChat();

    return (
        <div className="flex flex-col h-screen bg-[#050B14] text-slate-300 font-sans overflow-hidden animate-in fade-in duration-700">
            <div className="flex justify-between items-end px-8 pt-6 mb-6 shrink-0">
                <div>
                    <h2 className="text-2xl font-black text-white tracking-tight flex items-center gap-3">
                        <div className="p-2 bg-indigo-500/10 rounded-xl text-indigo-400">
                            <ShieldCheck size={24}/>
                        </div>
                        AI SECURITY HUB
                    </h2>
                    <p className="text-[11px] text-slate-500 font-bold uppercase tracking-[0.2em] mt-1 ml-1">
                        Hệ thống phân tích & Truy vấn an ninh thời gian thực
                    </p>
                </div>
                <div className="flex items-center gap-2 bg-emerald-500/10 border border-emerald-500/20 px-4 py-2 rounded-2xl">
                    <Zap size={14} className="text-emerald-400 animate-pulse" />
                    <span className="text-[10px] font-black text-emerald-400 uppercase tracking-widest">Neural Network Active</span>
                </div>
            </div>

            {/* MAIN CONTENT AREA */}
            <div className="flex-1 bg-[#0A101D] rounded-[32px] border border-slate-800 shadow-2xl overflow-hidden flex">
                <div className="w-80 hidden lg:flex flex-col border-r border-slate-800/50 bg-[#0A101D]/50">
                    <ChatSidebar 
                        sessions={sessions} 
                        currentSessionId={currentSession?.id} 
                        onSelectSession={selectSession}
                        onNewSession={createNewSession}
                        onRename={renameSession}
                        onDelete={deleteSession}
                    />
                </div>

                <div className="flex-1 flex flex-col bg-[#0f172a]/30 relative">
                    {/* Active Session Header */}
                    <div className="px-8 py-5 border-b border-slate-800/50 flex justify-between items-center bg-[#0A101D]/40 backdrop-blur-xl z-10">
                        <div className="flex items-center gap-4">
                            <div className="w-2 h-2 rounded-full bg-indigo-500 shadow-[0_0_10px_rgba(99,102,241,0.5)] animate-pulse"></div>
                            <h3 className="font-black text-white text-sm uppercase tracking-wider">
                                {currentSession ? currentSession.title : "New Security Protocol"}
                            </h3>
                        </div>
                        <div className="text-[9px] font-black text-slate-600 uppercase tracking-[0.2em]">
                            Encrypted Channel
                        </div>
                    </div>

                    <div className="flex-1 relative overflow-hidden">
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