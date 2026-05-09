import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { 
    Shield, Monitor, AlertTriangle, Activity, 
    Flame, Lock, ShieldAlert, Crosshair, Radar 
} from 'lucide-react';
import axiosInstance from '../../api/axios';
import { useSocketSubscription } from '../../context/useSocketSubscription';

import DashboardCharts from './components/DashboardCharts';
import DashboardTable from './components/DashboardTable';

const Dashboard = () => {
    const navigate = useNavigate();
    const [stats, setStats] = useState(null);
    const [loading, setLoading] = useState(true);

    const fetchDashboardData = useCallback(async () => {
        try {
            const res = await axiosInstance.get('/dashboard/stats');
            setStats(res.data);
        } catch (error) {
            console.error("Lỗi tải Dashboard:", error);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => { fetchDashboardData(); }, [fetchDashboardData]);

    useSocketSubscription(['NEW_INCIDENT', 'asset_STATUS_CHANGED', 'INCIDENT_RESOLVED'], () => {
        fetchDashboardData();
    });

    if (loading || !stats) return (
        <div className="h-[calc(100vh-60px)] bg-[#050B14] flex flex-col items-center justify-center gap-4">
            <Radar size={40} className="text-indigo-500 animate-spin-slow"/>
            <p className="text-indigo-500 font-mono text-sm tracking-widest animate-pulse uppercase">Syncing Telemetry Data...</p>
        </div>
    );

    // Chuẩn bị dữ liệu cho Bảng Top Risk
    const topRiskColumns = [
        { key: 'hostname', label: 'Asset Name', render: (row) => <span className="font-bold text-white">{row.hostname}</span> },
        { key: 'ip_address', label: 'IP', render: (row) => <span className="font-mono text-slate-400">{row.ip_address}</span> },
        { key: 'department_tag', label: 'Dept', render: (row) => <span className="bg-[#050B14] border border-slate-700 px-1.5 py-0.5 rounded text-[9px] text-slate-300">{row.department_tag}</span> },
        { key: 'risk_score', label: 'Risk', render: (row) => <span className={`font-black font-mono ${row.risk_score > 70 ? 'text-red-500' : 'text-orange-500'}`}>{row.risk_score}</span> }
    ];

    // Chuẩn bị dữ liệu cho Bảng Live Alerts
    const recentAlertColumns = [
        { key: 'alert_type', label: 'Detection', render: (row) => <span className="font-bold text-slate-200">{row.alert_type}</span> },
        { key: 'severity', label: 'Level', render: (row) => <span className={`px-1.5 py-0.5 rounded text-[9px] font-black uppercase border ${row.severity === 'Critical' ? 'bg-red-500/10 text-red-500 border-red-500/30' : 'bg-orange-500/10 text-orange-400 border-orange-500/30'}`}>{row.severity}</span> },
        { key: 'asset_name', label: 'Target', render: (row) => <span className="font-mono text-slate-400">{row.asset_name.substring(0,8)}...</span> },
        { key: 'created_at', label: 'Time', render: (row) => <span className="font-mono text-slate-500">{new Date(row.created_at).toLocaleTimeString('vi-VN')}</span> }
    ];

    return (
        <div className="h-[calc(100vh-60px)] bg-[#050B14] text-slate-200 p-5 font-sans overflow-y-auto custom-scrollbar">
            
            <div className="mb-4 flex items-end justify-between">
                <div>
                    <h1 className="text-xl font-black text-white flex items-center gap-2 uppercase tracking-tight">
                        <Crosshair className="text-indigo-500"/> SOC COMMAND CENTER
                    </h1>
                    <p className="text-[10px] text-slate-500 mt-1 uppercase tracking-widest font-bold">Real-time Telemetry & Threat Intelligence</p>
                </div>
                <div className="flex items-center gap-2 bg-[#0A101D] border border-slate-800 px-3 py-1.5 rounded-lg">
                    <div className={`w-2 h-2 rounded-full animate-pulse ${stats.posture.threat_level === 'Critical' ? 'bg-red-500' : 'bg-emerald-500'}`}></div>
                    <span className={`text-[10px] font-mono font-bold tracking-widest ${stats.posture.threat_level === 'Critical' ? 'text-red-400' : 'text-emerald-400'}`}>SYSTEM {stats.posture.threat_level.toUpperCase()}</span>
                </div>
            </div>

            {/* TIER 1: KPI CỐT LÕI (MẬT ĐỘ DÀY) */}
            <div className="grid grid-cols-2 md:grid-cols-6 gap-4 mb-4">
                <KpiCard icon={<Monitor size={18}/>} title="Total Assets" value={stats.summary.total_assets} subValue={`${stats.summary.online_assets} Online`} color="blue" />
                <KpiCard icon={<Flame size={18}/>} title="High Risk" value={stats.summary.high_risk_assets} subValue={`${stats.summary.total_assets > 0 ? ((stats.summary.high_risk_assets/stats.summary.total_assets)*100).toFixed(1) : 0}% Fleet`} color="red" isAlert={stats.summary.high_risk_assets > 0}/>
                <KpiCard icon={<ShieldAlert size={18}/>} title="Raw Alerts" value={stats.summary.total_alerts} subValue="Last 24h" color="orange" />
                <KpiCard icon={<AlertTriangle size={18}/>} title="Open Cases" value={stats.summary.open_incidents} subValue={`${stats.summary.resolved_incidents} Resolved`} color="purple" />
                <KpiCard icon={<Lock size={18}/>} title="Health Score" value={`${stats.posture.overall_score}%`} subValue={`Threat: ${stats.posture.threat_level}`} color="emerald" isAlert={stats.posture.threat_level === 'Critical'} />
                <KpiCard icon={<Shield size={18}/>} title="Avg Trust" value={stats.summary.average_trust_score.toFixed(1)} subValue="System Health" color="indigo" />
            </div>

            {/* TIER 2: BIỂU ĐỒ TRỰC QUAN (CHART) */}
            <div className="mb-4">
                <DashboardCharts stats={stats} />
            </div>

            {/* TIER 3: BẢNG DỮ LIỆU ACTIONABLE */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <DashboardTable 
                    title="Top Risk assets" 
                    icon={Flame} 
                    data={stats.summary.top_risk_assets} 
                    columns={topRiskColumns} 
                    colorClass="text-red-400"
                    onRowClick={(row) => navigate(`/assets/${row.hwid}`)}
                />
                <DashboardTable 
                    title="Live Threat Feed" 
                    icon={Activity} 
                    data={stats.summary.recent_alerts} 
                    columns={recentAlertColumns} 
                    colorClass="text-orange-400"
                />
            </div>

        </div>
    );
};

const KpiCard = ({ icon, title, value, subValue, color, isAlert }) => {
    const colorMap = {
        red: 'text-red-500 bg-red-500/10 border-red-500/30',
        emerald: 'text-emerald-500 bg-emerald-500/10 border-emerald-500/30',
        orange: 'text-orange-500 bg-orange-500/10 border-orange-500/30',
        blue: 'text-blue-500 bg-blue-500/10 border-blue-500/30',
        purple: 'text-purple-500 bg-purple-500/10 border-purple-500/30',
        indigo: 'text-indigo-500 bg-indigo-500/10 border-indigo-500/30',
    };
    return (
        <div className={`bg-[#0A101D] p-4 rounded-xl border border-slate-800 shadow-lg relative overflow-hidden group ${isAlert ? 'border-red-500/50 shadow-[0_0_15px_rgba(239,68,68,0.1)]' : ''}`}>
            {isAlert && <div className="absolute top-0 right-0 w-12 h-12 bg-red-500/10 rounded-bl-full animate-pulse"></div>}
            <div className="flex items-start justify-between mb-3">
                <div className={`p-2 rounded-lg ${colorMap[color]}`}>{icon}</div>
                <span className="text-[9px] font-black uppercase text-slate-500 tracking-widest">{title}</span>
            </div>
            <div>
                <p className={`text-2xl font-black font-mono tracking-tighter ${colorMap[color].split(' ')[0]}`}>{value}</p>
                <p className="text-[10px] text-slate-500 font-mono mt-0.5">{subValue}</p>
            </div>
        </div>
    );
};

export default Dashboard;