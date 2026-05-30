import React, { useState, useEffect, useCallback } from 'react';
import axiosInstance from '../../api/axios';
import { Activity, Search, ChevronLeft, ChevronRight } from 'lucide-react';
import BehaviorCard from "./components/BehaviorCard";

const SEVERITY_OPTIONS = [
    { key: 'all', label: 'Tất cả' },
    { key: 'Critical', label: 'Critical' },
    { key: 'High', label: 'High' },
    { key: 'Medium', label: 'Medium' },
];

const STATUS_OPTIONS = [
    { key: 'all', label: 'Tất cả' },
    { key: 'open', label: 'Chưa xử lý' },
    { key: 'resolved', label: 'Đã xử lý' },
];

const BehaviorManager = () => {
    const [behaviors, setBehaviors] = useState([]);
    const [loading, setLoading] = useState(true);
    const [page, setPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);
    const [totalAlerts, setTotalAlerts] = useState(0);
    const [searchQuery, setSearchQuery] = useState('');
    const [debouncedSearch, setDebouncedSearch] = useState('');
    const [severityFilter, setSeverityFilter] = useState('all');
    const [statusFilter, setStatusFilter] = useState('all');

   const fetchBehaviors = useCallback(async () => {
    setLoading(true);
    try {
        const params = new URLSearchParams();
        params.set('page', String(page));
        params.set('limit', '10');
        if (debouncedSearch) params.set('search', debouncedSearch);
        if (severityFilter !== 'all') params.set('severity', severityFilter);
        if (statusFilter !== 'all') params.set('status', statusFilter);

        const res = await axiosInstance.get(`/behaviors?${params.toString()}`);
        
        const items = res.data.items || [];
        const total = res.data.total || 0;
        
        setBehaviors(items);
        setTotalAlerts(total);
        setTotalPages(Math.max(1, Math.ceil(total / 10)));

    } catch (err) {
        console.error("Lỗi tải danh sách hành vi:", err);
        setBehaviors([]);
        setTotalAlerts(0);
        setTotalPages(1);
    } finally {
        setLoading(false);
    }
}, [page, debouncedSearch, severityFilter, statusFilter]);

    useEffect(() => { fetchBehaviors(); }, [fetchBehaviors]);

    useEffect(() => {
        const timer = setTimeout(() => setDebouncedSearch(searchQuery.trim()), 250);
        return () => clearTimeout(timer);
    }, [searchQuery]);

    useEffect(() => {
        setPage(1);
    }, [debouncedSearch, severityFilter, statusFilter]);

    const unresolvedCount = behaviors.filter((b) => !b.is_resolved).length;
    const resolvedCount = behaviors.filter((b) => b.is_resolved).length;
    const criticalCount = behaviors.filter((b) => b.severity === 'Critical').length;

    return (
        <div className="p-6 bg-[#050B14] min-h-[calc(100vh-60px)] text-slate-200">
            <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between mb-8">
                <div>
                    <h1 className="text-2xl font-black text-white flex items-center gap-3 tracking-tighter">
                        <Activity className="text-emerald-500" size={28}/> GIÁM SÁT HÀNH VI
                    </h1>
                    <p className="text-[10px] text-slate-500 mt-1 uppercase tracking-[0.2em] font-bold">
                        Dữ liệu sự kiện an ninh và cảnh báo hành vi từ MongoDB
                    </p>
                </div>

                <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
                    <div className="rounded-2xl border border-slate-800 bg-[#0A101D] p-4 text-sm">
                        <p className="text-[10px] uppercase tracking-[0.2em] text-slate-500">Tổng cảnh báo</p>
                        <p className="mt-2 text-2xl font-black text-white">{totalAlerts}</p>
                    </div>
                    <div className="rounded-2xl border border-slate-800 bg-[#0A101D] p-4 text-sm">
                        <p className="text-[10px] uppercase tracking-[0.2em] text-slate-500">Chưa xử lý</p>
                        <p className="mt-2 text-2xl font-black text-amber-400">{unresolvedCount}</p>
                    </div>
                    <div className="rounded-2xl border border-slate-800 bg-[#0A101D] p-4 text-sm">
                        <p className="text-[10px] uppercase tracking-[0.2em] text-slate-500">Đã lập sự cố</p>
                        <p className="mt-2 text-2xl font-black text-emerald-400">{resolvedCount}</p>
                    </div>
                    <div className="rounded-2xl border border-slate-800 bg-[#0A101D] p-4 text-sm">
                        <p className="text-[10px] uppercase tracking-[0.2em] text-slate-500">Critical</p>
                        <p className="mt-2 text-2xl font-black text-red-500">{criticalCount}</p>
                    </div>
                </div>
            </div>

            <div className="grid gap-4 lg:grid-cols-[1.8fr_1fr] mb-6">
                <div className="relative bg-[#0A101D] border border-slate-800 rounded-3xl p-4">
                    <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500" size={18} />
                    <input
                        type="text"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        placeholder="Tìm kiếm theo Asset HWID, loại cảnh báo, tiêu đề..."
                        className="w-full bg-transparent pl-12 pr-4 py-3 rounded-2xl border border-slate-800 text-sm text-slate-200 outline-none focus:border-indigo-500/70 transition-all"
                    />
                </div>

                <div className="grid gap-3 sm:grid-cols-2">
                    <div className="rounded-3xl border border-slate-800 bg-[#0A101D] p-4">
                        <p className="text-[10px] uppercase tracking-[0.2em] text-slate-500 mb-3">Mức độ</p>
                        <div className="flex flex-wrap gap-2">
                            {SEVERITY_OPTIONS.map((option) => (
                                <button
                                    key={option.key}
                                    onClick={() => setSeverityFilter(option.key)}
                                    className={`rounded-full px-3 py-1 text-xs font-semibold transition ${
                                        severityFilter === option.key
                                            ? 'bg-slate-700 text-white border border-indigo-500'
                                            : 'bg-slate-900 text-slate-400 hover:bg-slate-800'
                                    }`}
                                >
                                    {option.label}
                                </button>
                            ))}
                        </div>
                    </div>

                    <div className="rounded-3xl border border-slate-800 bg-[#0A101D] p-4">
                        <p className="text-[10px] uppercase tracking-[0.2em] text-slate-500 mb-3">Trạng thái</p>
                        <div className="flex flex-wrap gap-2">
                            {STATUS_OPTIONS.map((option) => (
                                <button
                                    key={option.key}
                                    onClick={() => setStatusFilter(option.key)}
                                    className={`rounded-full px-3 py-1 text-xs font-semibold transition ${
                                        statusFilter === option.key
                                            ? 'bg-slate-700 text-white border border-indigo-500'
                                            : 'bg-slate-900 text-slate-400 hover:bg-slate-800'
                                    }`}
                                >
                                    {option.label}
                                </button>
                            ))}
                        </div>
                    </div>
                </div>
            </div>

            <div className="hidden md:flex items-center gap-4 px-6 py-4 mb-3 bg-slate-900/20 backdrop-blur-sm border-y border-slate-800/50 rounded-3xl shadow-[0_0_18px_rgba(0,0,0,0.2)]">
                <div className="w-24 text-center text-[10px] font-black text-indigo-400 uppercase tracking-[0.25em]">
                    Mức độ
                </div>
                <div className="flex-1 text-[10px] font-black text-slate-400 uppercase tracking-[0.25em] flex items-center gap-2">
                    <span className="inline-flex h-2 w-2 rounded-full bg-indigo-500 animate-pulse"></span>
                    Thông tin chi tiết
                </div>
                <div className="w-48 text-right text-[10px] font-black text-slate-400 uppercase tracking-[0.25em]">
                    Thiết bị / Thời gian
                </div>
                <div className="w-20 text-right text-[10px] font-black text-slate-400 uppercase tracking-[0.25em]">
                    Thao tác
                </div>
            </div>

            {loading ? (
                <div className="flex flex-col gap-4">
                    {[1, 2, 3, 4].map((i) => (
                        <div key={i} className="h-44 rounded-3xl border border-slate-800 bg-slate-900/50 animate-pulse" />
                    ))}
                </div>
            ) : behaviors.length === 0 ? (
                <div className="py-24 text-center rounded-3xl border border-dashed border-slate-800 bg-[#0A101D]/30">
                    <p className="text-slate-500 uppercase tracking-[0.2em] font-bold">Không có hành vi phù hợp với bộ lọc</p>
                </div>
            ) : (
                <div className="flex flex-col gap-4">
                    {behaviors.map((item) => (
                        <BehaviorCard
                            key={item.id}
                            behavior={item}
                            onEscalated={fetchBehaviors}
                        />
                    ))}
                </div>
            )}

            <div className="mt-8 flex flex-wrap justify-between items-center gap-3">
                <div className="text-sm text-slate-400">
                    Hiển thị <span className="text-white">{behaviors.length}</span> trên <span className="text-white">{totalAlerts}</span> cảnh báo
                </div>
                <div className="flex items-center gap-2">
                    <button
                        disabled={page === 1}
                        onClick={() => setPage((p) => p - 1)}
                        className="p-2 rounded-2xl border border-slate-800 bg-[#0A101D] disabled:opacity-30"
                    >
                        <ChevronLeft size={18} />
                    </button>
                    {Array.from({ length: totalPages }, (_, i) => i + 1)
                        .filter((p) => p === 1 || p === totalPages || (p >= page - 1 && p <= page + 1))
                        .map((p, index, array) => (
                            <React.Fragment key={p}>
                                {index > 0 && array[index - 1] !== p - 1 && <span className="text-slate-600">...</span>}
                                <button
                                    onClick={() => setPage(p)}
                                    className={`px-3 py-1 rounded-2xl border text-xs font-mono transition ${
                                        page === p
                                            ? 'bg-indigo-600 border-indigo-500 text-white shadow-lg shadow-indigo-500/20'
                                            : 'bg-[#0A101D] border-slate-800 text-slate-400 hover:border-slate-600'
                                    }`}
                                >
                                    {p}
                                </button>
                            </React.Fragment>
                        ))}
                    <button
                        disabled={page === totalPages}
                        onClick={() => setPage((p) => p + 1)}
                        className="p-2 rounded-2xl border border-slate-800 bg-[#0A101D] disabled:opacity-30"
                    >
                        <ChevronRight size={18} />
                    </button>
                </div>
            </div>
        </div>
    );
};

export default BehaviorManager;
