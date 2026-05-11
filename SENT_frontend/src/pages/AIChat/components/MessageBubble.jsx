import React from 'react';
import ThinkingBlock from './ThinkingBlock';
import { Bot, User } from 'lucide-react';

const MessageBubble = ({ message }) => {
    const isAI = message.sender === 'ai';

    return (
        <div className={`flex flex-col ${isAI ? 'items-start' : 'items-end'} group animate-in fade-in slide-in-from-bottom-2 duration-500`}>
            <div className={`flex gap-4 max-w-[85%] ${isAI ? 'flex-row' : 'flex-row-reverse'}`}>
                <div className={`shrink-0 w-10 h-10 rounded-2xl border flex items-center justify-center shadow-lg ${
                    isAI ? 'bg-slate-900 border-slate-800 text-indigo-400' : 'bg-indigo-600 border-indigo-500 text-white'
                }`}>
                    {isAI ? <Bot size={20}/> : <User size={20}/>}
                </div>

                <div className="flex flex-col min-w-0">
                    <div className={`flex items-center mb-1 gap-2 ${!isAI && 'justify-end'}`}>
                        <span className="text-[10px] font-black uppercase tracking-widest text-slate-500">
                            {isAI ? 'AI Security Assistant' : 'Security Analyst'}
                        </span>
                    </div>

                    {isAI && message.thought && <ThinkingBlock thought={message.thought} />}

                    <div className={`px-6 py-4 rounded-[24px] text-sm leading-relaxed shadow-xl break-words whitespace-pre-wrap ${
                        isAI 
                        ? 'bg-[#0A101D] text-slate-200 border border-slate-800' 
                        : 'bg-indigo-600 text-white font-medium'
                    }`}>
                        {message.text}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default MessageBubble;