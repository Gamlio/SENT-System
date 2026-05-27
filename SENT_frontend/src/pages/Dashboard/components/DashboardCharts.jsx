import React, { useState, useEffect, useMemo } from 'react';
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, RadarChart, PolarGrid, PolarAngleAxis, Radar } from 'recharts';
import { Activity, ShieldAlert, Crosshair } from 'lucide-react';
import { useSocketSubscription } from '../../../context/useSocketSubscription';

const DashboardCharts = ({ stats }) => {
    const [liveTrendData, setLiveTrendData] = useState([]);
    const [liveTotalAlerts, setLiveTotalAlerts] = useState(0);

    useEffect(() => {
        if (stats && stats.summary) {
            setLiveTrendData(stats.summary.threat_trends || []);
            setLiveTotalAlerts(stats.summary.total_alerts || 0);
        }
    }, [stats]);

    useSocketSubscription('NEW_ALERT_TICK', (payload) => {
        console.log("🔥 Live Alert Nhận Được từ Socket:", payload);
        
        setLiveTotalAlerts(prev => prev + 1);

        setLiveTrendData(prevData => {
            const newData = [...prevData];
            if (newData.length > 0) {
                const lastIndex = newData.length - 1;
                newData[lastIndex] = {
                    ...newData[lastIndex],
                    alerts: (newData[lastIndex].alerts || 0) + 1
                };
            }
            return newData;
        });
    });

    // 3. DỮ LIỆU CHO PIE CHART (Giữ nguyên)
    const severityData = useMemo(() => {
        const raw = stats.summary.alerts_by_severity || {};
        const mapColor = { 'Critical': '#ef4444', 'High': '#f97316', 'Medium': '#eab308', 'Low': '#3b82f6' };
        return Object.keys(raw).map(key => ({
            name: key,
            value: raw[key],
            color: mapColor[key] || '#64748b'
        })).filter(d => d.value > 0);
    }, [stats]);

    // 4. RADAR DATA: So sánh các Vector bảo mật (Giữ nguyên)
    const radarData = useMemo(() => [
        { subject: 'Zero Trust', A: stats.posture.zero_trust_health, fullMark: 100 },
        { subject: 'Trust Score', A: stats.summary.average_trust_score, fullMark: 100 },
        { subject: 'Online Ratio', A: (stats.summary.online_assets / (stats.summary.total_assets || 1)) * 100, fullMark: 100 },
        { subject: 'Risk Level', A: (stats.summary.high_risk_assets / (stats.summary.total_assets || 1)) * 100, fullMark: 100 },
        { subject: 'Resolve Rate', A: (stats.summary.resolved_incidents / ((stats.summary.open_incidents + stats.summary.resolved_incidents) || 1)) * 100, fullMark: 100 }
    ], [stats]);

    const CustomTooltip = ({ active, payload, label }) => {
        if (active && payload && payload.length) {
            return (
                <div className="bg-[#050B14] border border-slate-700 p-3 rounded-lg shadow-xl">
                    <p className="text-[10px] font-black uppercase text-slate-400 mb-2 tracking-widest">{label}</p>
                    {payload.map((p, idx) => (
                        <p key={idx} className="text-[11px] font-mono font-bold" style={{ color: p.color }}>
                            {p.name.toUpperCase()}: {p.value}
                        </p>
                    ))}
                </div>
            );
        }
        return null;
    };

    return (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 h-72">
            {/* AREA CHART: DETECTION TRENDS (BÂY GIỜ ĐÃ LÀ SỐ LIỆU THẬT) */}
            <div className="bg-[#0A101D] border border-slate-800 rounded-xl shadow-lg p-4 flex flex-col relative overflow-hidden">
                <div key={liveTotalAlerts} className="absolute inset-0 bg-red-500/10 pointer-events-none animate-fade-out opacity-0"></div>
                
                <div className="flex items-center gap-2 mb-4 shrink-0">
                    <Activity size={14} className="text-blue-500"/>
                    <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest flex items-center gap-2">
                        Live Threat Velocity 
                        <span className="w-1.5 h-1.5 rounded-full bg-red-500 animate-pulse"></span>
                    </h3>
                </div>
                <div className="flex-1 min-h-0">
                    <ResponsiveContainer width="100%" height="100%">
                        <AreaChart data={liveTrendData} margin={{ top: 0, right: 0, left: -20, bottom: 0 }}>
                            <defs>
                                <linearGradient id="colorAlerts" x1="0" y1="0" x2="0" y2="1">
                                    <stop offset="5%" stopColor="#f97316" stopOpacity={0.4}/>
                                    <stop offset="95%" stopColor="#f97316" stopOpacity={0}/>
                                </linearGradient>
                            </defs>
                            <XAxis dataKey="time" stroke="#334155" fontSize={9} tickLine={false} axisLine={false} />
                            <YAxis stroke="#334155" fontSize={9} tickLine={false} axisLine={false} />
                            <Tooltip content={<CustomTooltip />} />
                            <Area isAnimationActive={true} type="monotone" dataKey="alerts" stroke="#f97316" strokeWidth={2} fill="url(#colorAlerts)" />
                        </AreaChart>
                    </ResponsiveContainer>
                </div>
            </div>

            {/* DONUT CHART: SEVERITY BREAKDOWN (Giữ nguyên) */}
            <div className="bg-[#0A101D] border border-slate-800 rounded-xl shadow-lg p-4 flex flex-col items-center">
                <div className="flex items-center gap-2 mb-2 w-full shrink-0">
                    <ShieldAlert size={14} className="text-orange-500"/>
                    <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest">Alert Breakdown</h3>
                </div>
                <div className="flex-1 w-full min-h-0 relative">
                    <ResponsiveContainer width="100%" height="100%">
                        <PieChart>
                            <Pie data={severityData} cx="50%" cy="50%" innerRadius={50} outerRadius={75} paddingAngle={3} dataKey="value" stroke="none">
                                {severityData.map((entry, index) => <Cell key={`cell-${index}`} fill={entry.color} />)}
                            </Pie>
                            <Tooltip content={<CustomTooltip />} />
                        </PieChart>
                    </ResponsiveContainer>
                    <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
                        <span className="text-2xl font-black font-mono text-white">{liveTotalAlerts}</span>
                        <span className="text-[8px] uppercase tracking-widest text-slate-500 font-bold">Total</span>
                    </div>
                </div>
            </div>

            {/* RADAR CHART: SECURITY POSTURE (Giữ nguyên) */}
            <div className="bg-[#0A101D] border border-slate-800 rounded-xl shadow-lg p-4 flex flex-col">
                <div className="flex items-center gap-2 mb-2 shrink-0">
                    <Crosshair size={14} className="text-emerald-500"/>
                    <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest">Security Posture Vector</h3>
                </div>
                <div className="flex-1 min-h-0">
                    <ResponsiveContainer width="100%" height="100%">
                        <RadarChart cx="50%" cy="50%" outerRadius={60} data={radarData}>
                            <PolarGrid stroke="#334155" />
                            <PolarAngleAxis dataKey="subject" tick={{ fill: '#64748b', fontSize: 8, fontWeight: 'bold' }} />
                            <Radar name="Org State" dataKey="A" stroke="#10b981" strokeWidth={2} fill="#10b981" fillOpacity={0.2} />
                            <Tooltip content={<CustomTooltip />} />
                        </RadarChart>
                    </ResponsiveContainer>
                </div>
            </div>
        </div>
    );
};

export default DashboardCharts;