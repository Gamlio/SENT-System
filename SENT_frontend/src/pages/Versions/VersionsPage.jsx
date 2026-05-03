import React from 'react';
import { Cpu, RefreshCcw, Plus, BookOpen, Terminal, Box } from 'lucide-react';
import { useVersions } from './hooks/useVersions';
import VersionCard from './components/VersionCard';

const VersionsPage = () => {
    const { versions, loading, refresh } = useVersions();

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
                    <p className="text-[11px] text-slate-500 mt-2 uppercase tracking-[0.2em] font-bold flex items-center gap-2 text-decoration-none">
                        <span className="w-2 h-2 bg-emerald-500 rounded-full animate-pulse"></span>
                        Hệ thống phân phối Agent bảo mật trực tuyến
                    </p>
                </div>
                
                <div className="flex gap-3">
                    <button onClick={refresh} className="p-2.5 bg-slate-900 border border-slate-800 rounded-xl text-slate-400 hover:text-white transition-all shadow-xl">
                        <RefreshCcw size={18} className={loading ? 'animate-spin' : ''} />
                    </button>
                    <button className="flex items-center gap-2 px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-xl text-xs font-black uppercase tracking-widest transition-all shadow-lg shadow-indigo-500/20">
                        <Plus size={16}/> New Release
                    </button>
                </div>
            </div>

            <div className="flex flex-col lg:flex-row gap-10 flex-1 overflow-hidden relative z-10">
                {/* CỘT TRÁI: DOCUMENTATION */}
                <div className="lg:w-[380px] flex flex-col gap-6 overflow-y-auto pr-4 custom-scrollbar shrink-0">
                    <section className="bg-slate-900/40 backdrop-blur-md border border-slate-800/60 rounded-2xl p-6">
                        <h2 className="text-xs font-black text-indigo-400 flex items-center gap-2 mb-5 uppercase tracking-widest">
                            <BookOpen size={14}/> System Overview
                        </h2>
                        <p className="text-[13px] text-slate-400 leading-relaxed font-medium">
                            SENT-Agent đóng vai trò là "mắt xích" đầu cuối giúp thu thập telemetry và phản ứng với các mối đe dọa.
                        </p>
                    </section>

                    <section className="bg-slate-900/40 backdrop-blur-md border border-slate-800/60 rounded-2xl p-6 relative overflow-hidden group">
                        <h2 className="text-xs font-black text-emerald-400 flex items-center gap-2 mb-6 uppercase tracking-widest">
                            <Terminal size={14}/> Quick Installation
                        </h2>
                        <div className="space-y-6">
                            {[
                                { step: "01", title: "Select Build", desc: "Tải gói cài đặt tương ứng với OS." },
                                { step: "02", title: "Grant Permissions", desc: "Sử dụng chmod +x trên Unix-like." },
                                { step: "03", title: "Connect", desc: "Sử dụng API Key để xác thực Agent." }
                            ].map((item, idx) => (
                                <div key={idx} className="flex gap-4">
                                    <span className="text-[10px] font-black text-slate-600 mt-1 italic">{item.step}</span>
                                    <div>
                                        <h3 className="text-[11px] font-bold text-slate-200 uppercase tracking-wide">{item.title}</h3>
                                        <p className="text-[11px] text-slate-500 mt-1">{item.desc}</p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </section>
                </div>

                {/* CỘT PHẢI: VERSIONS LIST */}
                <div className="flex-1 flex flex-col overflow-y-auto custom-scrollbar pr-2">
                    {loading ? (
                        <div className="h-full flex items-center justify-center">
                            <div className="w-10 h-10 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin"></div>
                        </div>
                    ) : versions.length === 0 ? (
                        <div className="flex-1 flex flex-col items-center justify-center bg-slate-900/20 border-2 border-dashed border-slate-800/60 rounded-2xl min-h-[400px]">
                            <Box size={32} className="text-slate-600 mb-4" />
                            <p className="text-slate-400 text-sm font-black uppercase tracking-widest">No builds found</p>
                            <p className="text-[10px] text-slate-500 mt-2 uppercase">Chưa có phiên bản nào được deploy từ hệ thống.</p>
                        </div>
                    ) : (
                        <div className="grid grid-cols-1 xl:grid-cols-2 gap-6 pb-10">
                            {versions.map((v) => (
                                <VersionCard key={v.tag} version={v} />
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default VersionsPage;