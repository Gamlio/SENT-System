import React, { useState } from 'react';
import { BrainCircuit, ChevronDown, ChevronRight } from 'lucide-react';

const ThinkingBlock = ({ thought }) => {
    const [isOpen, setIsOpen] = useState(false); // Mặc định đóng cho gọn

    if (!thought) return null;

    return (
        <div className="mb-2 max-w-[85%]">
            <button 
                onClick={() => setIsOpen(!isOpen)}
                className="flex items-center gap-2 text-xs font-bold text-slate-500 hover:text-indigo-400 transition-colors mb-1"
            >
                <BrainCircuit size={14} />
                Quá trình suy luận
                {isOpen ? <ChevronDown size={12}/> : <ChevronRight size={12}/>}
            </button>
            
            {isOpen && (
                <div className="bg-slate-900/50 border-l-2 border-indigo-500/30 p-3 rounded-r-lg text-xs text-slate-400 italic font-mono whitespace-pre-wrap animate-in fade-in slide-in-from-top-1">
                    {thought}
                </div>
            )}
        </div>
    );
};

export default ThinkingBlock;