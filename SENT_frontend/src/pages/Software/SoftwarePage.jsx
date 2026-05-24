import React from 'react';
import { Cpu, RefreshCcw, Box } from 'lucide-react';
import VersionCard from './components/SoftwareCard';

const VersionsPage = () => {
    const loading = false;
    const refresh = () => { window.location.reload(); };

    const softwares = [
        {
            tag: "v4.2.5",
            is_latest: true,
            release_note: "Bản build ổn định hệ thống phân phối nội bộ SENT Agent.",
            release_date: "2026-05-23",
            links: {
                windows: "/sof_builds/SENT_v4.2.5_windows.exe",
                linux: "/sof_builds/SENT_v4.2.5_linux",
                mac: "/sof_builds/SENT_v4.2.5_mac"
            }
        }
    ];

    return (
        <div className="p-8 h-[calc(100vh-64px)] flex flex-col text-slate-200 bg-[#020617] font-sans relative overflow-hidden">
            {/* Background Decor */}
            <div className="absolute top-0 right-0 w-[500px] h-[500px] bg-indigo-500/5 blur-[120px] rounded-full -mr-64 -mt-64 pointer-events-none"></div>
            
            {/* Header */}
            <div className="flex justify-between items-center mb-10 shrink-0 relative z-10">
                <div>
                    <h1 className="text-2xl font-black text-white flex items-center gap-3 uppercase tracking-tighter">
                        <div className="p-2 bg-indigo-500/10 rounded-lg border border-indigo-500/20">
                            <Cpu className="text-indigo-400" size={24}/>
                        </div>
                        Agent Deployment Central
                    </h1>
                    <p className="text-[11px] text-slate-500 mt-2 uppercase tracking-[0.2em] font-bold flex items-center gap-2">
                        <span className="w-2 h-2 bg-emerald-500 rounded-full animate-pulse"></span>
                        Hệ thống phân phối Agent bảo mật trực tuyến (Static Mode)
                    </p>
                </div>
                
                <div className="flex gap-3">
                    <button onClick={refresh} className="p-2.5 bg-slate-900 border border-slate-800 rounded-xl text-slate-400 hover:text-white transition-all shadow-xl">
                        <RefreshCcw size={18} />
                    </button>
                </div>
            </div>

            {/* VERSIONS LIST */}
            <div className="flex-1 overflow-y-auto custom-scrollbar pr-2 relative z-10">
                {softwares.length === 0 ? (
                    <div className="flex-1 flex flex-col items-center justify-center bg-slate-900/20 border-2 border-dashed border-slate-800/60 rounded-2xl min-h-[400px]">
                        <Box size={32} className="text-slate-600 mb-4" />
                        <p className="text-slate-400 text-sm font-black uppercase tracking-widest">No builds found</p>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 gap-6 pb-10">
                        {softwares.map((v) => (
                            <VersionCard key={v.tag} software={v} />
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
};

export default VersionsPage;