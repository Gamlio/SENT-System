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
            // Gọi endpoint /behaviors với phân trang
            const res = await axiosInstance.get(`/behaviors?page=${page}&limit=12`);
            // Backend trả về mảng trực tiếp hoặc object phân trang tùy cấu trúc service
            // SỬA TẠI ĐÂY: Nếu res.data là null, dùng mảng rỗng []
            setBehaviors(res.data || []); 
        } catch (err) {
            console.error("Lỗi tải danh sách hành vi:", err);
            setBehaviors([]); // Đảm bảo luôn là mảng khi lỗi
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
                <button className="bg-[#0A101D] border border-slate-800 px-4 rounded-xl hover:bg-slate-800 transition-colors">
                    <Filter size={16} className="text-slate-400"/>
                </button>
            </div>

            {/* Danh sách hành vi */}
            {loading ? (
                <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
                    {[1,2,3,4,5,6].map(i => <div key={i} className="h-48 bg-slate-800/20 animate-pulse rounded-2xl border border-slate-800"></div>)}
                </div>
            ) : behaviors.length === 0 ? (
                <div className="py-20 text-center border border-dashed border-slate-800 rounded-2xl bg-[#0A101D]/30">
                    <p className="text-slate-600 font-mono text-sm uppercase tracking-widest">No suspicious behavior detected.</p>
                </div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
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
            <div className="mt-8 flex justify-center items-center gap-4">
                <button 
                    disabled={page === 1}
                    onClick={() => setPage(p => p - 1)}
                    className="p-2 bg-[#0A101D] border border-slate-800 rounded-lg disabled:opacity-30"
                >
                    <ChevronLeft size={20}/>
                </button>
                <span className="text-xs font-mono font-bold text-slate-500">PAGE {page}</span>
                <button 
                    onClick={() => setPage(p => p + 1)}
                    className="p-2 bg-[#0A101D] border border-slate-800 rounded-lg"
                >
                    <ChevronRight size={20}/>
                </button>
            </div>
        </div>
    );
};

export default BehaviorManager;