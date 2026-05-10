import React from 'react';
import { Download, Monitor, Terminal, Apple, Calendar, Info, Hash } from 'lucide-react';

const VersionCard = ({ software }) => {
    const platforms = [
        { name: 'Windows', key: 'windows', icon: <Monitor size={16}/>, color: 'text-blue-400', bg: 'bg-blue-400/5' },
        { name: 'Linux', key: 'linux', icon: <Terminal size={16}/>, color: 'text-orange-400', bg: 'bg-orange-400/5' },
        { name: 'MacOS', key: 'mac', icon: <Apple size={16}/>, color: 'text-slate-200', bg: 'bg-slate-200/5' }
    ];

    return (
        <div className={`group relative bg-slate-900/50 backdrop-blur-sm border ${software.is_latest ? 'border-indigo-500/40' : 'border-slate-800'} rounded-2xl p-6 hover:border-indigo-500/60 transition-all duration-300`}>
            {/* Hiệu ứng tia sáng khi bản mới nhất */}
            {software.is_latest && (
                <div className="absolute -top-px left-10 right-10 h-px bg-gradient-to-r from-transparent via-indigo-500 to-transparent"></div>
            )}

            <div className="flex justify-between items-start mb-6">
                <div>
                    <div className="flex items-center gap-3">
                        <span className="text-white font-black text-xl tracking-tighter uppercase">Build {software.tag}</span>
                        {software.is_latest && (
                            <span className="bg-indigo-500 text-[9px] font-black px-2.5 py-0.5 rounded-full text-white uppercase tracking-tighter">
                                Stable
                            </span>
                        )}
                    </div>
                    <div className="flex items-center gap-4 mt-2 text-slate-500 text-[10px] font-bold uppercase tracking-widest">
                        <span className="flex items-center gap-1.5"><Calendar size={12}/> {software.release_date}</span>
                    </div>
                </div>
            </div>

            {/* Release Note Section */}
            {software.release_note && (
                <div className="mb-6 p-4 bg-black/20 border border-slate-800/50 rounded-xl">
                    <div className="flex items-center gap-2 text-indigo-400 mb-2">
                        <Info size={12}/>
                        <span className="text-[10px] font-black uppercase tracking-widest">Release Intelligence</span>
                    </div>
                    <p className="text-[12px] text-slate-400 leading-relaxed italic line-clamp-2 group-hover:line-clamp-none transition-all">
                        "{software.release_note}"[cite: 8]
                    </p>
                </div>
            )}

            {/* Platform Downloads */}
            <div className="grid grid-cols-1 gap-2.5">
                {platforms.map(p => (
                    <a 
                        key={p.key} 
                        href={software.links?.[p.key] || "#"} 
                        className={`flex items-center justify-between p-3 ${p.bg} border border-slate-800/50 rounded-xl hover:border-indigo-500/50 hover:bg-slate-800/80 transition-all group/btn`}
                    >
                        <div className="flex items-center gap-3">
                            <div className={`p-2 rounded-lg bg-slate-900 border border-slate-800 ${p.color}`}>
                                {p.icon}
                            </div>
                            <span className="text-xs font-black text-slate-300 group-hover/btn:text-white transition-colors">{p.name} Build</span>
                        </div>
                        <Download size={16} className="text-slate-600 group-hover/btn:text-indigo-400 transform group-hover/btn:translate-y-0.5 transition-all"/>
                    </a>
                ))}
            </div>
        </div>
    );
};

export default VersionCard;