import React from 'react';
import { Cpu, RefreshCcw, Plus } from 'lucide-react';
import { useVersions } from './hooks/useVersions';
import VersionCard from './components/VersionCard';

const VersionsPage = () => {
    const { versions, loading, refresh } = useVersions();

    return (
        <div className="p-6 h-[calc(100vh-60px)] flex flex-col text-slate-200 bg-[#050B14] font-sans">
            {/* Header chuyên nghiệp */}
            <div className="flex justify-between items-end mb-8 shrink-0">
                <div>
                    <h1 className="text-xl font-black text-white flex items-center gap-2 uppercase tracking-tight">
                        <Cpu className="text-indigo-500"/> Software Management
                    </h1>
                    <p className="text-[11px] text-slate-500 mt-1 uppercase tracking-widest font-bold">
                        Build & Release Central | Total Releases: {versions.length}
                    </p>
                </div>
                
                <div className="flex gap-2">
                    <button 
                        onClick={refresh}
                        className="p-2 bg-[#0A101D] border border-slate-800 rounded text-slate-400 hover:text-white transition-all"
                    >
                        <RefreshCcw size={16} className={loading ? 'animate-spin' : ''} />
                    </button>
                    <button className="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded text-[10px] font-black uppercase tracking-widest transition-all shadow-[0_0_15px_rgba(79,70,229,0.3)]">
                        <Plus size={14}/> Create New Release
                    </button>
                </div>
            </div>

            {/* Danh sách phiên bản (Grid) */}
            {loading ? (
                <div className="flex-1 flex items-center justify-center">
                    <div className="w-8 h-8 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin"></div>
                </div>
            ) : (
                <div className="flex-1 overflow-y-auto custom-scrollbar pr-2">
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {versions.map((v) => (
                            <VersionCard key={v.tag} version={v} />
                        ))}
                    </div>
                    
                    {versions.length === 0 && (
                        <div className="h-64 flex flex-col items-center justify-center border-2 border-dashed border-slate-800 rounded-xl">
                            <p className="text-slate-500 text-xs font-black uppercase tracking-widest">No versions found</p>
                            <p className="text-[10px] text-slate-600 mt-1">Push a git tag to trigger auto-build</p>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
};

export default VersionsPage;