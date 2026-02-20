import React, { useState, useEffect } from 'react';
import { Activity, ShieldAlert, Cpu, Server, WifiOff } from 'lucide-react';
import { PieChart, Pie, Cell, AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';
import axios from '../../api/axios';

const Dashboard = () => {
    const [stats, setStats] = useState({ total: 0, online: 0, alerts: 0, regions: 0 });
    const [allAgents, setAllAgents] = useState([]);

    const fetchData = async () => {
        try {
            const statsRes = await axios.get('/agents/stats');
            setStats(statsRes.data);

            // Chỉ gọi API 1 lần lấy toàn bộ máy, tự phân loại trên Frontend
            const agentsRes = await axios.get('/agents');
            setAllAgents(agentsRes.data || []);
        } catch (err) {
            console.error("Lỗi cập nhật dữ liệu:", err);
        }
    };

    useEffect(() => {
        fetchData();
        const interval = setInterval(fetchData, 30000); // 30s reload 1 lần
        return () => clearInterval(interval);
    }, []);

    // --- LỌC DỮ LIỆU ĐỘNG ---
    const offlineAgents = allAgents.filter(a => a.status === 'offline');
    // Lấy 5 máy mới nhất dựa vào danh sách chung
    const recentAgents = [...allAgents].slice(0, 5); 

    // Biểu đồ
    const agentStatusData = [
        { name: 'Online', value: stats.online, color: '#10b981' }, 
        { name: 'Offline', value: stats.total > stats.online ? stats.total - stats.online : 0, color: '#64748b' }, 
        { name: 'Cảnh báo', value: stats.alerts, color: '#ef4444' } 
    ];

    const growthTrendData = [
        { time: 'Thứ 2', agents: 2, users: 1 }, { time: 'Thứ 3', agents: 5, users: 2 },
        { time: 'Thứ 4', agents: 8, users: 4 }, { time: 'Thứ 5', agents: 12, users: 5 },
        { time: 'Thứ 6', agents: 18, users: 8 }, { time: 'T7', agents: 22, users: 10 },
        { time: 'CN', agents: 25, users: 12 },
    ];

    return (
       <div className="text-slate-200">
            <header className="mb-8">
                <h1 className="text-3xl font-bold text-white tracking-tight">Tổng quan Hệ thống</h1>
                <p className="text-slate-400 text-sm mt-1">Giám sát trạng thái máy trạm và cảnh báo bảo mật thời gian thực</p>
            </header>

            {/* THỐNG KÊ */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
                <StatCard icon={<Server className="text-blue-400"/>} title="Tổng máy trạm" value={stats.total} detail="Đã đăng ký" />
                <StatCard icon={<Activity className="text-emerald-400"/>} title="Đang Online" value={stats.online} detail="Hoạt động" />
                <StatCard icon={<Activity className="text-red-400"/>} title="Đang Offline" value={offlineAgents.length} detail="Không hoạt động" />
                <StatCard icon={<ShieldAlert className="text-red-400"/>} title="Cảnh báo" value={stats.alerts} detail="Cần xử lý ngay" />
                <StatCard icon={<Cpu className="text-amber-400"/>} title="Vùng / Chi nhánh" value={stats.regions} detail="Toàn hệ thống" />
            </div>

            {/* BIỂU ĐỒ */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
                <div className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl flex flex-col">
                    <h3 className="text-sm font-bold text-slate-400 uppercase tracking-widest mb-6">Trạng thái Agent</h3>
                    <div className="flex-1 min-h-[220px]">
                        <ResponsiveContainer width="100%" height="100%">
                            <PieChart>
                                <Pie data={agentStatusData} cx="50%" cy="50%" innerRadius={70} outerRadius={90} paddingAngle={5} dataKey="value">
                                    {agentStatusData.map((entry, index) => (<Cell key={`cell-${index}`} fill={entry.color} />))}
                                </Pie>
                                <Tooltip contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '12px' }} itemStyle={{ color: '#fff' }} />
                            </PieChart>
                        </ResponsiveContainer>
                    </div>
                    <div className="flex justify-center gap-6 mt-4">
                        {agentStatusData.map(stat => (
                            <div key={stat.name} className="flex items-center gap-2 text-xs font-bold">
                                <span className="w-3 h-3 rounded-full shadow-lg" style={{ backgroundColor: stat.color }}></span>
                                <span className="text-slate-400">{stat.name}:</span> <span className="text-white">{stat.value}</span>
                            </div>
                        ))}
                    </div>
                </div>

                <div className="lg:col-span-2 bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl flex flex-col">
                    <h3 className="text-sm font-bold text-slate-400 uppercase tracking-widest mb-6">Tăng trưởng Người dùng & Máy trạm (7 ngày qua)</h3>
                    <div className="flex-1 min-h-[220px]">
                        <ResponsiveContainer width="100%" height="100%">
                            <AreaChart data={growthTrendData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                                <defs>
                                    <linearGradient id="colorAgents" x1="0" y1="0" x2="0" y2="1"><stop offset="5%" stopColor="#10b981" stopOpacity={0.3}/><stop offset="95%" stopColor="#10b981" stopOpacity={0}/></linearGradient>
                                    <linearGradient id="colorUsers" x1="0" y1="0" x2="0" y2="1"><stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3}/><stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/></linearGradient>
                                </defs>
                                <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                                <XAxis dataKey="time" stroke="#64748b" fontSize={12} tickLine={false} axisLine={false} />
                                <YAxis stroke="#64748b" fontSize={12} tickLine={false} axisLine={false} />
                                <Tooltip contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '12px' }} itemStyle={{ color: '#fff' }} />
                                <Area type="monotone" dataKey="agents" name="Máy trạm" stroke="#10b981" strokeWidth={3} fillOpacity={1} fill="url(#colorAgents)" />
                                <Area type="monotone" dataKey="users" name="Người dùng" stroke="#3b82f6" strokeWidth={3} fillOpacity={1} fill="url(#colorUsers)" />
                            </AreaChart>
                        </ResponsiveContainer>
                    </div>
                </div>
            </div>
            
            {/* KHU VỰC BẢNG DỮ LIỆU: CHIA LÀM 2 CỘT */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                
                {/* BẢNG 1: MÁY TRẠM MỚI */}
                <div className="bg-[#1e293b] rounded-3xl border border-slate-800 overflow-hidden shadow-xl">
                    <div className="p-6 border-b border-slate-800 flex justify-between items-center bg-slate-800/30">
                        <h3 className="font-bold text-lg text-white">Mới gia nhập</h3>
                    </div>
                    <div className="overflow-x-auto">
                        <table className="w-full text-left">
                            <thead className="text-[10px] uppercase tracking-widest text-slate-500 bg-slate-900/50">
                                <tr>
                                    <th className="p-5 font-bold">Hostname</th>
                                    <th className="p-5 font-bold">Trạng thái</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-slate-800">
                                {recentAgents.length > 0 ? (
                                    recentAgents.map(agent => (
                                        <AgentRow key={agent.hwid} name={agent.hostname || "Unknown"} status={agent.status} lastSeen={agent.last_seen} />
                                    ))
                                ) : (
                                    <tr><td colSpan="2" className="p-8 text-center text-slate-500 italic">Chưa có dữ liệu.</td></tr>
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>

                {/* BẢNG 2: MÁY TRẠM ĐANG OFFLINE (CẢNH BÁO) */}
                <div className="bg-[#1e293b] rounded-3xl border border-red-900/30 overflow-hidden shadow-xl">
                    <div className="p-6 border-b border-red-900/50 flex justify-between items-center bg-red-500/5">
                        <h3 className="font-bold text-lg text-red-400 flex items-center gap-2"><WifiOff size={20}/> Danh sách Offline</h3>
                    </div>
                    <div className="overflow-x-auto">
                        <table className="w-full text-left">
                            <thead className="text-[10px] uppercase tracking-widest text-red-400/50 bg-red-900/10">
                                <tr>
                                    <th className="p-5 font-bold">Hostname</th>
                                    <th className="p-5 font-bold">Thời gian ngắt kết nối</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-red-900/20">
                                {offlineAgents.length > 0 ? (
                                    offlineAgents.map(agent => (
                                        <AgentRow key={agent.hwid} name={agent.hostname || "Unknown"} status={agent.status} lastSeen={agent.last_seen} isOfflineList />
                                    ))
                                ) : (
                                    <tr><td colSpan="2" className="p-8 text-center text-emerald-500 italic font-medium">Toàn hệ thống đang hoạt động ổn định.</td></tr>
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>

            </div>
        </div>
    );
};

// Component hỗ trợ
const StatCard = ({ icon, title, value, detail }) => (
    <div className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl hover:border-slate-700 transition-colors">
        <div className="mb-4"><div className="p-3 bg-slate-900/80 rounded-2xl inline-block shadow-inner">{icon}</div></div>
        <p className="text-slate-500 text-xs font-bold uppercase tracking-wider mb-1">{title}</p>
        <h3 className="text-4xl font-black text-white mb-2">{value}</h3>
        <p className="text-[10px] text-slate-400 font-medium">{detail}</p>
    </div>
);

const getTimeAgo = (dateString, status) => {
    if (!dateString) return "Không rõ";
    const diffInSeconds = Math.floor((new Date() - new Date(dateString)) / 1000);
    if (status === 'online') {
        if (diffInSeconds < 60) return `Nhịp đập: vài giây trước`;
        return `Nhịp đập: ${Math.floor(diffInSeconds / 60)} phút trước`;
    }
    if (diffInSeconds < 60) return `Mất kết nối vài giây trước`;
    const mins = Math.floor(diffInSeconds / 60);
    if (mins < 60) return `Mất kết nối ${mins} phút trước`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `Offline ${hours} giờ trước`;
    return `Offline ${Math.floor(hours / 24)} ngày trước`;
};

// isOfflineList để tạo viền đỏ nhẹ nếu nằm trong danh sách Offline
const AgentRow = ({ name, status, lastSeen, isOfflineList }) => (
    <tr className={`hover:bg-slate-800/50 transition-colors ${isOfflineList ? 'hover:bg-red-500/5' : ''}`}>
        <td className={`p-5 font-bold text-sm ${isOfflineList ? 'text-red-300' : 'text-white'}`}>{name}</td>
        <td className="p-5">
            <div className="flex flex-col">
                <div className={`flex items-center gap-2 text-[10px] font-bold uppercase ${status === 'online' ? 'text-emerald-400' : 'text-red-400'}`}>
                    <span className={`w-2 h-2 rounded-full ${status === 'online' ? 'bg-emerald-400 animate-pulse' : 'bg-red-500'}`}></span>
                    {status}
                </div>
                <span className={`text-[10px] mt-1 font-medium italic ${isOfflineList ? 'text-red-400/70' : 'text-slate-400'}`}>
                    {getTimeAgo(lastSeen, status)}
                </span>
            </div>
        </td>
    </tr>
);

export default Dashboard;