import React from 'react';
import { Download, Monitor, Terminal, Apple, Calendar, Info } from 'lucide-react';

const VersionCard = ({ software }) => {
    const platforms = [
        { name: 'Windows', key: 'windows', icon: <Monitor size={16}/>, color: 'text-blue-400', bg: 'bg-blue-400/5' },
        { name: 'Linux', key: 'linux', icon: <Terminal size={16}/>, color: 'text-orange-400', bg: 'bg-orange-400/5' },
        { name: 'macOS', key: 'mac', icon: <Apple size={16}/>, color: 'text-slate-200', bg: 'bg-slate-200/5' }
    ];

    return (
        <div className={`group relative bg-[#0B1224]/80 backdrop-blur-md border ${software.is_latest ? 'border-indigo-500/30 shadow-[0_0_30px_rgba(79,70,229,0.1)]' : 'border-slate-800'} rounded-2xl p-6 hover:border-indigo-500/60 transition-all duration-500`}>
            {software.is_latest && (
                <div className="absolute -top-px left-10 right-10 h-[2px] bg-gradient-to-r from-transparent via-indigo-500 to-transparent shadow-[0_0_10px_#6366f1]"></div>
            )}

            <div className="flex justify-between items-start mb-6">
                <div>
                    <div className="flex items-center gap-3">
                        <span className="text-white font-black text-lg tracking-tight uppercase">SENT System Agent</span>
                        {software.is_latest && (
                            <span className="bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-[8px] font-black px-2 py-0.5 rounded-md uppercase tracking-tighter">
                                LATEST
                            </span>
                        )}
                    </div>
                    <div className="flex items-center gap-4 mt-1.5 text-slate-500 text-[9px] font-bold uppercase tracking-[0.15em]">
                        <span className="flex items-center gap-1.5"><Calendar size={12}/> {software.release_date}</span>
                        <span className="text-indigo-500/60">●</span>
                        <span className="text-indigo-400/80 italic lowercase">v{software.tag}</span>
                    </div>
                </div>
            </div>

            {software.release_note && (
                <div className="mb-6 p-4 bg-indigo-500/5 border border-indigo-500/10 rounded-xl">
                    <div className="flex items-center gap-2 text-indigo-400 mb-2">
                        <Info size={12}/>
                        <span className="text-[10px] font-black uppercase tracking-widest">Release Intelligence</span>
                    </div>
                    <p className="text-[11px] text-slate-400 leading-relaxed font-medium line-clamp-2 group-hover:line-clamp-none transition-all duration-300">
                        {software.release_note}
                    </p>
                </div>
            )}

            {/* Platform Downloads */}
            <div className="grid grid-cols-1 gap-2.5">
                {platforms.map(p => {
                    const fileUrl = software.links?.[p.key] || "#";
                    const fileName = fileUrl.split('/').pop();

                    return (
                        <a 
                            key={p.key} 
                            href={fileUrl} 
                            download={fileName}
                            className={`flex items-center justify-between p-3.5 ${p.bg} border border-slate-800/40 rounded-xl hover:border-indigo-500/50 hover:bg-indigo-500/10 transition-all duration-300 group/btn shadow-inner`}
                        >
                            <div className="flex items-center gap-3">
                                <div className={`p-2 rounded-lg bg-[#020617] border border-slate-800 group-hover/btn:border-indigo-500/40 ${p.color} transition-all`}>
                                    {p.icon}
                                </div>
                                <span className="text-[11px] font-black text-slate-400 group-hover/btn:text-white uppercase tracking-wider transition-colors">{p.name}</span>
                            </div>
                            <Download size={14} className="text-slate-600 group-hover/btn:text-indigo-400 transform group-hover/btn:scale-110 transition-all"/>
                        </a>
                    );
                })}
            </div>
        </div>
    );
};

export default VersionCard;