import React, { useState } from 'react';
import { BrainCircuit, ChevronDown, ChevronRight, Activity } from 'lucide-react';

const ThinkingBlock = ({ thought }) => {
    const [isOpen, setIsOpen] = useState(false);
    if (!thought) return null;

    return (
        <div className="mb-3">
            <button 
                onClick={() => setIsOpen(!isOpen)}
                className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-indigo-500/10 border border-indigo-500/20 hover:bg-indigo-500/20 transition-all shadow-sm group"
            >
                <Activity size={12} className="text-indigo-400 animate-pulse" />
                <span className="text-[10px] font-black uppercase tracking-widest text-indigo-400">
                    Phân tích logic
                </span>
                {isOpen ? <ChevronDown size={12} className="text-indigo-400"/> : <ChevronRight size={12} className="text-indigo-400"/>}
            </button>
            
            {isOpen && (
                <div className="mt-2 bg-[#050B14] border-l-2 border-indigo-500/50 p-4 rounded-r-2xl text-[11px] text-slate-500 italic font-mono whitespace-pre-wrap animate-in slide-in-from-top-1 duration-300">
                    {thought}
                </div>
            )}
        </div>
    );
};

export default ThinkingBlock;