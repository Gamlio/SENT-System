import React, { useState, useEffect, useMemo } from 'react';
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer, PieChart, Pie, Cell, RadarChart, PolarGrid, PolarAngleAxis, Radar } from 'recharts';
import { Activity, ShieldAlert, Crosshair } from 'lucide-react';
import { useSocketSubscription } from '../../../context/useSocketSubscription';

const DashboardCharts = ({ stats }) => {
    
    // 1. TẠO STATE RIÊNG CHO BIỂU ĐỒ ĐỂ NÓ CÓ THỂ "SỐNG"
    const [liveTrendData, setLiveTrendData] = useState([]);
    const [liveTotalAlerts, setLiveTotalAlerts] = useState(0);

    // Khởi tạo data ban đầu từ API (chỉ chạy 1 lần khi load)
    useEffect(() => {
        if (stats) {
            // Giả lập data 7 ngày như cũ
            const data = [];
            const baseAlerts = Math.floor(stats.total_alerts / 7) || 10;
            for (let i = 6; i >= 0; i--) {
                const d = new Date();
                d.setDate(d.getDate() - i);
                data.push({
                    time: `${d.getDate()}/${d.getMonth()+1}`,
                    alerts: Math.abs(baseAlerts + Math.floor(Math.random() * 20) - 10),
                    incidents: Math.floor(Math.random() * 5)
                });
            }
            setLiveTrendData(data);
            setLiveTotalAlerts(stats.total_alerts);
        }
    }, [stats]);

    // 2. LẮNG NGHE SOCKET THỜI GIAN THỰC (TRUE REAL-TIME)
    useSocketSubscription('NEW_ALERT_TICK', (payload) => {
        console.log("🔥 Live Alert Nhận Được:", payload);
        
        // Cập nhật số tổng ngay lập tức
        setLiveTotalAlerts(prev => prev + 1);

        // Bơm data mới vào biểu đồ (Làm cột ngày hôm nay tăng lên 1)
        setLiveTrendData(prevData => {
            const newData = [...prevData];
            if (newData.length > 0) {
                // Lấy cột cuối cùng (Hôm nay) và +1 vào chỉ số alerts
                const lastIndex = newData.length - 1;
                newData[lastIndex] = {
                    ...newData[lastIndex],
                    alerts: newData[lastIndex].alerts + 1
                };
            }
            return newData;
        });
    });

    // 1. DỮ LIỆU CHO PIE CHART (Alerts By Severity)
    const severityData = useMemo(() => {
        const raw = stats.alerts_by_severity || {};
        const mapColor = { 'Critical': '#ef4444', 'High': '#f97316', 'Medium': '#eab308', 'Low': '#3b82f6' };
        return Object.keys(raw).map(key => ({
            name: key,
            value: raw[key],
            color: mapColor[key] || '#64748b'
        })).filter(d => d.value > 0);
    }, [stats]);

    // 2. DỮ LIỆU GIẢ LẬP TREND (Do API hiện chỉ trả Snapshot, ta tạo Data mẫu minh họa sự biến thiên)
    const trendData = useMemo(() => {
        const data = [];
        const baseAlerts = Math.floor(stats.total_alerts / 7) || 10;
        for (let i = 6; i >= 0; i--) {
            const d = new Date();
            d.setDate(d.getDate() - i);
            data.push({
                time: `${d.getDate()}/${d.getMonth()+1}`,
                alerts: Math.abs(baseAlerts + Math.floor(Math.random() * 20) - 10),
                incidents: Math.floor(Math.random() * 5)
            });
        }
        return data;
    }, [stats]);

    // 3. RADAR DATA: So sánh các Vector bảo mật
    const radarData = useMemo(() => [
        { subject: 'Zero Trust', A: stats.zero_trust_coverage, fullMark: 100 },
        { subject: 'Trust Score', A: stats.average_trust_score, fullMark: 100 },
        { subject: 'Online Ratio', A: (stats.online_agents / (stats.total_agents || 1)) * 100, fullMark: 100 },
        { subject: 'Risk Level', A: (stats.high_risk_agents / (stats.total_agents || 1)) * 100, fullMark: 100 },
        { subject: 'Resolve Rate', A: (stats.resolved_incidents / ((stats.open_incidents + stats.resolved_incidents) || 1)) * 100, fullMark: 100 }
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
            
            {/* AREA CHART: DETECTION TRENDS (BÂY GIỜ ĐÃ LÀ LIVE) */}
            <div className="bg-[#0A101D] border border-slate-800 rounded-xl shadow-lg p-4 flex flex-col relative overflow-hidden">
                {/* Hiệu ứng chớp nháy khi có data mới */}
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
                            
                            {/* Chú ý: Đổi isAnimationActive thành true để nó trượt mượt mà */}
                            <Area isAnimationActive={true} type="monotone" dataKey="alerts" stroke="#f97316" strokeWidth={2} fill="url(#colorAlerts)" />
                        </AreaChart>
                    </ResponsiveContainer>
                </div>
            </div>

            {/* DONUT CHART: SEVERITY BREAKDOWN */}
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
                        <span className="text-2xl font-black font-mono text-white">{stats.total_alerts}</span>
                        <span className="text-[8px] uppercase tracking-widest text-slate-500 font-bold">Total</span>
                    </div>
                </div>
            </div>

            {/* RADAR CHART: SECURITY POSTURE */}
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