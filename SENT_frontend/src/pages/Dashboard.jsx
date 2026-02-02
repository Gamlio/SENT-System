import React from 'react';
import { Activity, ShieldAlert, Cpu, Server } from 'lucide-react';

const Dashboard = () => {
    return (
        <div className="p-10 bg-[#0f172a] min-h-screen text-slate-200">
            {/* Header Dashboard */}
            <header className="mb-10">
                <h2 className="text-4xl font-bold text-white mb-2">Hệ thống giám sát</h2>
                <p className="text-slate-400">Chào mừng bạn trở lại, hệ thống đang ở trạng thái <span className="text-emerald-400 font-bold">BẢO VỆ</span></p>
            </header>

            {/* Grid Thống kê */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-12">
                <StatCard icon={<Server className="text-blue-400"/>} title="Tổng máy trạm" value="1,248" detail="+12 hôm nay" />
                <StatCard icon={<Activity className="text-emerald-400"/>} title="Đang Online" value="842" detail="68% hoạt động" />
                <StatCard icon={<ShieldAlert className="text-red-400"/>} title="Cảnh báo AI" value="24" detail="5 nghiêm trọng" />
                <StatCard icon={<Cpu className="text-amber-400"/>} title="Vùng quản lý" value="12" detail="Toàn hệ thống" />
            </div>

            {/* Bảng dữ liệu tập trung */}
            <div className="bg-[#1e293b] rounded-3xl border border-slate-800 overflow-hidden shadow-2xl">
                <div className="p-6 border-b border-slate-800 flex justify-between items-center bg-slate-800/50">
                    <h3 className="font-bold text-lg">Máy trạm mới gia nhập</h3>
                    <button className="text-xs bg-emerald-500 hover:bg-emerald-600 text-white px-4 py-2 rounded-full font-bold transition">Xem tất cả</button>
                </div>
                <div className="overflow-x-auto">
                    <table className="w-full text-left">
                        <thead className="text-[10px] uppercase tracking-widest text-slate-500 bg-slate-900/30">
                            <tr>
                                <th className="p-5 font-bold">Tên máy (Hostname)</th>
                                <th className="p-5 font-bold">Khu vực</th>
                                <th className="p-5 font-bold">Hệ điều hành</th>
                                <th className="p-5 font-bold">Trạng thái</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-800">
                            <AgentRow name="ICTU-LAB-01" region="Bắc Giang" os="Windows 11" status="Online" />
                            <AgentRow name="SEC-DESKTOP-44" region="Thái Nguyên" os="Ubuntu 22.04" status="Offline" />
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
};

const StatCard = ({ icon, title, value, detail }) => (
    <div className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 hover:border-emerald-500/50 transition-all duration-300 group shadow-lg">
        <div className="flex justify-between items-start mb-4">
            <div className="p-3 bg-slate-900 rounded-2xl group-hover:scale-110 transition-transform">{icon}</div>
        </div>
        <p className="text-slate-500 text-xs font-bold uppercase tracking-wider mb-1">{title}</p>
        <h3 className="text-3xl font-black text-white mb-1">{value}</h3>
        <p className="text-[10px] text-slate-400 font-medium">{detail}</p>
    </div>
);

const AgentRow = ({ name, region, os, status }) => (
    <tr className="hover:bg-slate-800/50 transition-colors">
        <td className="p-5 font-bold text-sm">{name}</td>
        <td className="p-5 text-sm"><span className="bg-slate-900 px-3 py-1 rounded-full text-slate-400 border border-slate-700">{region}</span></td>
        <td className="p-5 text-sm text-slate-400">{os}</td>
        <td className="p-5">
            <div className={`flex items-center gap-2 text-[10px] font-bold uppercase tracking-tighter ${status === 'Online' ? 'text-emerald-400' : 'text-red-400'}`}>
                <span className={`w-2 h-2 rounded-full ${status === 'Online' ? 'bg-emerald-400 animate-pulse' : 'bg-red-400'}`}></span>
                {status}
            </div>
        </td>
    </tr>
);

export default Dashboard;