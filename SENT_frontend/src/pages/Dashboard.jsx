import React, { useState, useEffect } from 'react';
import { Activity, ShieldAlert, Cpu, Server } from 'lucide-react';
import axios from '../api/axios'; // Đảm bảo file axios.js đã cấu hình cổng 8000

const Dashboard = () => {
    const [stats, setStats] = useState({ total: 0, online: 0, alerts: 0, regions: 0 });
    const [recentAgents, setRecentAgents] = useState([]);

    useEffect(() => {
        const fetchData = async () => {
            try {
                // 1. Lấy con số thống kê thực
                const statsRes = await axios.get('/agents/stats');
                setStats(statsRes.data);

                // 2. Lấy danh sách máy mới nhất
                const agentsRes = await axios.get('/agents');
                setRecentAgents(agentsRes.data);
            } catch (err) {
                console.error("Lỗi lấy dữ liệu thực:", err);
            }
        };
        fetchData();
    }, []);

    return (
       <div className="p-10 bg-[#0f172a] min-h-screen text-slate-200">
            {/* Thay giá trị cứng bằng biến stats */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-12">
                <StatCard icon={<Server className="text-blue-400"/>} title="Tổng máy trạm" value={stats.total} detail="Realtime" />
                <StatCard icon={<Activity className="text-emerald-400"/>} title="Đang Online" value={stats.online} detail="Hoạt động" />
                <StatCard icon={<ShieldAlert className="text-red-400"/>} title="Cảnh báo" value={stats.alerts} detail="Cần xử lý" />
                <StatCard icon={<Cpu className="text-amber-400"/>} title="Vùng" value={stats.regions} detail="Toàn hệ thống" />
            </div>
            
            <div className="bg-[#1e293b] rounded-3xl border border-slate-800 overflow-hidden shadow-2xl">
                <div className="p-6 border-b border-slate-800 flex justify-between items-center bg-slate-800/50">
                    <h3 className="font-bold text-lg">Máy trạm mới gia nhập</h3>
                </div>
                <div className="overflow-x-auto">
                    <table className="w-full text-left">
                        <thead className="text-[10px] uppercase tracking-widest text-slate-500 bg-slate-900/30">
                            <tr>
                                <th className="p-5 font-bold">Tên máy (Hostname)</th>
                                <th className="p-5 font-bold">Trạng thái</th>
                                <th className="p-5 font-bold">HWID</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-800">
                            {/* Chuyển dữ liệu thực vào từng dòng bảng */}
                            {recentAgents.map(agent => (
                                <AgentRow 
                                    key={agent.hwid} 
                                    name={agent.hostname || "Unknown"} 
                                    hwid={agent.hwid.substring(0, 15) + "..."} 
                                    status={agent.status} 
                                />
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
};

// Các Component hỗ trợ giữ nguyên cấu trúc CSS của Thanh
const StatCard = ({ icon, title, value, detail }) => (
    <div className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-lg">
        <div className="mb-4"><div className="p-3 bg-slate-900 rounded-2xl inline-block">{icon}</div></div>
        <p className="text-slate-500 text-xs font-bold uppercase mb-1">{title}</p>
        <h3 className="text-3xl font-black text-white mb-1">{value}</h3>
        <p className="text-[10px] text-slate-400 font-medium">{detail}</p>
    </div>
);

const AgentRow = ({ name, hwid, status }) => (
    <tr className="hover:bg-slate-800/50 transition-colors">
        <td className="p-5 font-bold text-sm text-white">{name}</td>
        <td className="p-5">
            <div className={`flex items-center gap-2 text-[10px] font-bold uppercase ${status === 'online' ? 'text-emerald-400' : 'text-red-400'}`}>
                <span className={`w-2 h-2 rounded-full ${status === 'online' ? 'bg-emerald-400 animate-pulse' : 'bg-red-400'}`}></span>
                {status}
            </div>
        </td>
        <td className="p-5 text-sm text-slate-500 font-mono">{hwid}</td>
    </tr>
);

export default Dashboard;