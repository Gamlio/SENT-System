import React, { useState, useEffect } from 'react';
import { Activity, ShieldAlert, Server, Wifi, WifiOff, PlusCircle } from 'lucide-react';
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';
import axios from '../../api/axios';
import DashboardTable from '../../components/DashboardTable'; // Import component vừa tạo

const Dashboard = () => {
    const [stats, setStats] = useState({ total: 0, online: 0, alerts: 0, regions: 0 });
    const [allAgents, setAllAgents] = useState([]);

    const fetchData = async () => {
        try {
            const statsRes = await axios.get('/agents/stats');
            setStats(statsRes.data);
            const agentsRes = await axios.get('/agents');
            setAllAgents(agentsRes.data || []);
        } catch (err) { console.error("Lỗi data:", err); }
    };

    useEffect(() => {
        fetchData();
        const interval = setInterval(fetchData, 30000);
        return () => clearInterval(interval);
    }, []);

    // --- PHÂN LOẠI DỮ LIỆU ---
    const onlineAgents = allAgents.filter(a => a.status === 'online');
    const offlineAgents = allAgents.filter(a => a.status === 'offline');
    // Giả lập "Mới gia nhập": Lấy 10 máy có created_at mới nhất (hoặc tạm lấy cuối danh sách)
    const newAgents = [...allAgents].reverse().slice(0, 10); 

    // Dữ liệu biểu đồ (Mockup chuẩn hóa UI)
    const chartData = [
        { time: '00:00', online: 45, offline: 2 }, { time: '04:00', online: 42, offline: 5 },
        { time: '08:00', online: 80, offline: 10 }, { time: '12:00', online: 95, offline: 8 },
        { time: '16:00', online: 90, offline: 4 }, { time: '20:00', online: 60, offline: 3 },
        { time: '23:59', online: 50, offline: 2 },
    ];

    return (
       <div className="text-slate-200 pb-10">
            <header className="mb-8">
                <h1 className="text-3xl font-bold text-white tracking-tight">Trung tâm Điều hành (SOC)</h1>
                <p className="text-slate-400 text-sm mt-1">Giám sát an ninh và trạng thái hoạt động thời gian thực</p>
            </header>

            {/* CARD THỐNG KÊ */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5 mb-8">
                <StatCard icon={<Server className="text-blue-400"/>} title="Tổng thiết bị" value={stats.total} color="blue" />
                <StatCard icon={<Activity className="text-emerald-400"/>} title="Hoạt động" value={stats.online} color="emerald" />
                <StatCard icon={<WifiOff className="text-slate-400"/>} title="Mất kết nối" value={offlineAgents.length} color="slate" />
                <StatCard icon={<ShieldAlert className="text-red-400"/>} title="Cảnh báo rủi ro" value={stats.alerts} color="red" isAlert />
            </div>

            {/* BIỂU ĐỒ CHÍNH (Full Width) */}
            <div className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl mb-8">
                <div className="flex justify-between items-center mb-6">
                    <h3 className="text-sm font-bold text-slate-400 uppercase tracking-widest">Xu hướng hoạt động (24h)</h3>
                    <span className="text-xs bg-emerald-500/10 text-emerald-400 px-3 py-1 rounded-full font-bold">Live Update</span>
                </div>
                <div className="h-[250px] w-full">
                    <ResponsiveContainer width="100%" height="100%">
                        <AreaChart data={chartData}>
                            <defs>
                                <linearGradient id="colorOnline" x1="0" y1="0" x2="0" y2="1"><stop offset="5%" stopColor="#10b981" stopOpacity={0.3}/><stop offset="95%" stopColor="#10b981" stopOpacity={0}/></linearGradient>
                            </defs>
                            <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                            <XAxis dataKey="time" stroke="#64748b" fontSize={12} tickLine={false} axisLine={false} />
                            <YAxis stroke="#64748b" fontSize={12} tickLine={false} axisLine={false} />
                            <Tooltip contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '12px', color: '#fff' }} />
                            <Area type="monotone" dataKey="online" stroke="#10b981" strokeWidth={3} fillOpacity={1} fill="url(#colorOnline)" />
                        </AreaChart>
                    </ResponsiveContainer>
                </div>
            </div>
            
            {/* KHU VỰC 3 BẢNG DỮ LIỆU */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* 1. MÁY OFFLINE */}
                <DashboardTable 
                    title="Mất kết nối" 
                    icon={WifiOff} 
                    data={offlineAgents} 
                    colorClass="border-red-500/30 shadow-red-900/10" 
                    statusType="offline"
                />

                {/* 2. MÁY ONLINE */}
                <DashboardTable 
                    title="Đang Online" 
                    icon={Wifi} 
                    data={onlineAgents} 
                    colorClass="border-slate-800" 
                    statusType="online"
                />

                {/* 3. MÁY MỚI */}
                <DashboardTable 
                    title="Mới gia nhập" 
                    icon={PlusCircle} 
                    data={newAgents} 
                    colorClass="border-blue-500/30 shadow-blue-900/10" 
                    statusType="online"
                />
            </div>
        </div>
    );
};

// Card thống kê nhỏ gọn
const StatCard = ({ icon, title, value, color, isAlert }) => (
    <div className={`bg-[#1e293b] p-6 rounded-3xl border ${isAlert ? 'border-red-500/40 animate-pulse' : 'border-slate-800'} shadow-lg flex items-center gap-5`}>
        <div className={`p-4 rounded-2xl bg-${color}-500/10 text-${color}-400`}>{icon}</div>
        <div>
            <p className="text-slate-500 text-[10px] font-bold uppercase tracking-widest">{title}</p>
            <h3 className="text-3xl font-black text-white">{value}</h3>
        </div>
    </div>
);

export default Dashboard;