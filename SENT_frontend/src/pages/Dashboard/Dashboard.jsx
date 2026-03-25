// src/pages/Dashboard/Dashboard.jsx (Mẫu cấu trúc đề xuất)
import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Shield, Monitor, AlertTriangle, Activity, Flame, ArrowUpRight } from 'lucide-react';
import axiosInstance from '../../api/axios';
import { useSocketSubscription } from '../../context/useSocketSubscription'; // Dùng hook bạn đã có

const Dashboard = () => {
    const navigate = useNavigate();
    const [stats, setStats] = useState(null);
    const [loading, setLoading] = useState(true);

    // Lấy dữ liệu tổng quan từ Backend
    const fetchDashboardData = useCallback(async () => {
        try {
            const res = await axiosInstance.get('/dashboard/summary');
            setStats(res.data);
        } catch (error) {
            console.error("Lỗi tải Dashboard:", error);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchDashboardData();
    }, [fetchDashboardData]);

    // Lắng nghe Real-time: Khi có sự cố mới hoặc thiết bị thay đổi trạng thái -> Load lại Dashboard
    useSocketSubscription(['NEW_INCIDENT', 'AGENT_STATUS_CHANGED', 'INCIDENT_RESOLVED'], () => {
        console.log('[WS] Có thay đổi hệ thống, cập nhật lại Dashboard...');
        fetchDashboardData();
    });

    if (loading || !stats) return <div className="p-10 text-emerald-500 animate-pulse">Đang tải trung tâm chỉ huy...</div>;

    return (
        <div className="p-6 h-full overflow-y-auto bg-[#050B14] text-slate-200">
            <h1 className="text-2xl font-black text-white mb-6 flex items-center gap-3">
                <Activity className="text-blue-500"/> Trung Tâm Chỉ Huy SOC
            </h1>

            {/* HÀNG 1: KPI CARDS (Thống kê nhanh) */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
                <KpiCard icon={<AlertTriangle/>} title="Sự cố đang mở (OPEN)" value={stats.open_incidents} color="red" />
                <KpiCard icon={<Monitor/>} title="Máy trạm Online" value={`${stats.online_agents}/${stats.total_agents}`} color="emerald" />
                <KpiCard icon={<Flame/>} title="Máy có Risk Score > 70" value={stats.high_risk_agents} color="orange" />
                <KpiCard icon={<Shield/>} title="Tỷ lệ phủ Zero Trust" value={`${stats.zero_trust_coverage}%`} color="blue" />
            </div>

            {/* HÀNG 2: BIỂU ĐỒ & DANH SÁCH RỦI RO */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-6">
                {/* Khu vực Biểu đồ (Chiếm 2 cột) - Gợi ý dùng Recharts hoặc Chart.js */}
                <div className="lg:col-span-2 bg-[#0A101D] p-5 rounded-2xl border border-slate-800 shadow-lg">
                    <h3 className="font-bold text-slate-400 mb-4 uppercase text-xs tracking-widest">Xu hướng cảnh báo 7 ngày qua</h3>
                    {/* Thêm Component LineChart của Recharts vào đây */}
                    <div className="h-64 border-2 border-dashed border-slate-800 rounded-xl flex items-center justify-center text-slate-600">Khu vực nhúng Biểu đồ</div>
                </div>

                {/* Danh sách máy trạm nguy hiểm nhất (Top Risk Assets) */}
                <div className="bg-[#0A101D] p-5 rounded-2xl border border-slate-800 shadow-lg flex flex-col">
                    <h3 className="font-bold text-red-400 mb-4 uppercase text-xs tracking-widest flex items-center gap-2">
                        <Flame size={14}/> Top Tài sản rủi ro (Risk Score)
                    </h3>
                    <div className="flex-1 space-y-3">
                        {stats.top_risk_agents.map((agent, idx) => (
                            <div key={idx} onClick={() => navigate(`/agents/${agent.hwid}`)} className="flex items-center justify-between p-3 bg-slate-900/50 rounded-xl border border-slate-800 hover:border-red-500/50 cursor-pointer transition group">
                                <div>
                                    <p className="font-bold text-sm text-slate-200 group-hover:text-white">{agent.hostname}</p>
                                    <p className="text-[10px] text-slate-500">{agent.ip_address}</p>
                                </div>
                                <div className="flex items-center gap-3">
                                    <span className="text-red-500 font-black text-lg">{agent.risk_score}</span>
                                    <ArrowUpRight size={16} className="text-slate-600 group-hover:text-red-400"/>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
};

const KpiCard = ({ icon, title, value, color }) => {
    const colorClasses = {
        red: 'text-red-500 bg-red-500/10 border-red-500/20',
        emerald: 'text-emerald-500 bg-emerald-500/10 border-emerald-500/20',
        orange: 'text-orange-500 bg-orange-500/10 border-orange-500/20',
        blue: 'text-blue-500 bg-blue-500/10 border-blue-500/20',
    };
    return (
        <div className="bg-[#0A101D] p-5 rounded-2xl border border-slate-800 shadow-lg flex items-center gap-4 relative overflow-hidden">
            <div className={`p-4 rounded-xl ${colorClasses[color]} z-10`}>{icon}</div>
            <div className="z-10">
                <p className="text-xs font-bold text-slate-500 uppercase tracking-wider mb-1">{title}</p>
                <p className="text-3xl font-black text-white tracking-tighter">{value}</p>
            </div>
            {/* Background Glow */}
            <div className={`absolute -right-4 -bottom-4 w-24 h-24 blur-3xl opacity-20 ${colorClasses[color].split(' ')[0].replace('text-', 'bg-')}`}></div>
        </div>
    );
};

export default Dashboard;
