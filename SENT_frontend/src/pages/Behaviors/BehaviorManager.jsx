import React, { useState, useEffect, useCallback } from 'react';
import axiosInstance from '../../api/axios';
import { Activity, ShieldAlert, Zap, Monitor, Filter, Search, ChevronLeft, ChevronRight } from 'lucide-react';
import BehaviorCard from "./components/BehaviorCard";

const BehaviorManager = () => {
    const [behaviors, setBehaviors] = useState([]);
    const [loading, setLoading] = useState(true);
    const [page, setPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);

   const fetchBehaviors = useCallback(async () => {
    setLoading(true);
    try {
        const res = await axiosInstance.get(`/behaviors?page=${page}&limit=10`);
        
        const items = res.data.items || [];
        const total = res.data.total || 0;
        
        setBehaviors(items); 
        setTotalPages(Math.ceil(total / 10)); 

    } catch (err) {
        console.error("Lỗi tải danh sách hành vi:", err);
        setBehaviors([]);
    } finally {
        setLoading(false);
    }
}, [page]);

    useEffect(() => { fetchBehaviors(); }, [fetchBehaviors]);

    return (
        <div className="p-6 bg-[#050B14] min-h-[calc(100vh-60px)] text-slate-200">
            {/* Header chuyên nghiệp mới */}
            <div className="flex justify-between items-end mb-8">
                <div>
                    <h1 className="text-2xl font-black text-white flex items-center gap-3 tracking-tighter">
                        <Activity className="text-emerald-500" size={28}/> BEHAVIOR ANALYTICS
                    </h1>
                    <p className="text-[10px] text-slate-500 mt-1 uppercase tracking-[0.2em] font-bold">
                        Raw Security Events & Policy Violations (MongoDB)
                    </p>
                </div>
                
                <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2 bg-[#0A101D] border border-slate-800 px-4 py-2 rounded-xl">
                        <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></div>
                        <span className="text-xs font-mono text-emerald-400 font-bold uppercase">Live Feed</span>
                    </div>
                </div>
            </div>

            {/* Toolbar: Search & Filter */}
            <div className="mb-6 flex gap-4">
                <div className="relative flex-1">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" size={16}/>
                    <input 
                        type="text" 
                        placeholder="Search by Asset HWID or Event Type..."
                        className="w-full bg-[#0A101D] border border-slate-800 rounded-xl py-2.5 pl-10 pr-4 text-xs focus:border-indigo-500/50 outline-none transition-all"
                    />
                </div>
                {/* <button className="bg-[#0A101D] border border-slate-800 px-4 rounded-xl hover:bg-slate-800 transition-colors">
                    <Filter size={16} className="text-slate-400"/>
                </button> */}
            </div>
               <div className="hidden md:flex items-center gap-4 px-6 py-4 mb-3 bg-slate-900/20 backdrop-blur-sm border-y border-slate-800/50 rounded-lg shadow-[0_0_15px_rgba(0,0,0,0.2)]">
                    <div className="w-24 text-center">
                        <span className="text-[10px] font-black text-indigo-400 uppercase tracking-[0.25em] drop-shadow-[0_0_8px_rgba(129,140,248,0.3)]">
                            Mức độ
                        </span>
                    </div>
                    
                    <div className="flex-1">
                        <span className="text-[10px] font-black text-slate-400 uppercase tracking-[0.25em] flex items-center gap-2">
                            <div className="w-1 h-1 bg-indigo-500 rounded-full animate-pulse"></div>
                            Thông tin
                        </span>
                    </div>
                    
                    <div className="w-48 text-right">
                        <span className="text-[10px] font-black text-slate-400 uppercase tracking-[0.25em]">
                            Thiết bị / Thời gian
                        </span>
                    </div>
                    
                    <div className="w-20 text-right">
                        <span className="text-[10px] font-black text-slate-400 uppercase tracking-[0.25em]">
                            Thao tác
                        </span>
                    </div>
                </div>
            {/* Danh sách hành vi */}
            {loading ? (
                <div className="flex flex-col gap-3">
                    {[1,2,3,4,5,6].map(i => <div key={i} className="h-48 bg-slate-800/20 animate-pulse rounded-2xl border border-slate-800"></div>)}
                </div>
            ) : behaviors.length === 0 ? (
                <div className="py-20 text-center border border-dashed border-slate-800 rounded-2xl bg-[#0A101D]/30">
                    <p className="text-slate-600 font-mono text-sm uppercase tracking-widest">No suspicious behavior detected.</p>
                </div>
            ) : (
                <div className="flex flex-col gap-3">
                    {behaviors.map(item => (
                        <BehaviorCard 
                            key={item.id} 
                            behavior={item} 
                            onEscalated={fetchBehaviors} 
                        />
                    ))}
                </div>
            )}

            {/* Phân trang */}
            <div className="mt-8 flex justify-center items-center gap-2">
                <button 
                    disabled={page === 1}
                    onClick={() => setPage(p => p - 1)}
                    className="p-2 bg-[#0A101D] border border-slate-800 rounded-lg disabled:opacity-30"
                >
                    <ChevronLeft size={18}/>
                </button>

                {/* Hiển thị danh sách số trang thông minh */}
                {Array.from({ length: totalPages }, (_, i) => i + 1)
                    .filter(p => p === 1 || p === totalPages || (p >= page - 1 && p <= page + 1))
                    .map((p, index, array) => (
                        <React.Fragment key={p}>
                            {index > 0 && array[index - 1] !== p - 1 && <span className="text-slate-600">...</span>}
                            <button
                                onClick={() => setPage(p)}
                                className={`px-3 py-1 rounded-lg border font-mono text-xs transition-all ${
                                    page === p 
                                    ? 'bg-indigo-600 border-indigo-500 text-white shadow-lg shadow-indigo-500/20' 
                                    : 'bg-[#0A101D] border-slate-800 text-slate-400 hover:border-slate-600'
                                }`}
                            >
                                {p}
                            </button>
                        </React.Fragment>
                    ))
                }

                <button 
                    disabled={page === totalPages}
                    onClick={() => setPage(p => p + 1)}
                    className="p-2 bg-[#0A101D] border border-slate-800 rounded-lg disabled:opacity-30"
                >
                    <ChevronRight size={18}/>
                </button>
            </div>
        </div>
    );
};

export default BehaviorManager;