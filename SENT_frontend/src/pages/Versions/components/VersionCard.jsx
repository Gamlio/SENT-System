import React from 'react';
import { Download, Monitor, Terminal, Apple, Calendar, Hash } from 'lucide-react';

const VersionCard = ({ version }) => {
    const platforms = [
        { name: 'Windows', key: 'windows', icon: <Monitor size={14}/>, color: 'text-blue-400' },
        { name: 'Linux', key: 'linux', icon: <Terminal size={14}/>, color: 'text-orange-400' },
        { name: 'MacOS', key: 'mac', icon: <Apple size={14}/>, color: 'text-slate-300' }
    ];

    return (
        <div className="bg-[#0A101D] border border-slate-800 rounded-xl p-5 hover:border-indigo-500/50 transition-all group">
            <div className="flex justify-between items-start mb-4">
                <div>
                    <div className="flex items-center gap-2">
                        <span className="text-white font-black text-lg tracking-tight">VERSION {version.tag}</span>
                        {version.is_latest && (
                            <span className="bg-emerald-500/10 text-emerald-400 text-[9px] font-black px-2 py-0.5 rounded border border-emerald-500/30 uppercase animate-pulse">
                                Latest Release
                            </span>
                        )}
                    </div>
                    <div className="flex items-center gap-3 mt-1 text-slate-500 text-[10px] font-mono uppercase">
                        <span className="flex items-center gap-1"><Calendar size={10}/> {version.release_date}</span>
                        <span className="flex items-center gap-1"><Hash size={10}/> {version.checksum_short}</span>
                    </div>
                </div>
            </div>

            <div className="grid grid-cols-1 gap-2">
                {platforms.map(p => (
                    <a key={p.key} href={version.links[p.key]} 
                       className="flex items-center justify-between p-2.5 bg-[#050B14] border border-slate-800 rounded-lg hover:bg-slate-800 transition-colors group/item">
                        <div className="flex items-center gap-3">
                            <span className={p.color}>{p.icon}</span>
                            <span className="text-xs font-bold text-slate-300">{p.name} Agent</span>
                        </div>
                        <Download size={14} className="text-slate-600 group-hover/item:text-indigo-400 transition-colors"/>
                    </a>
                ))}
            </div>
        </div>
    );
};

export default VersionCard;