import React, { useMemo } from 'react';
import { PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { PieChart as PieIcon, Activity } from 'lucide-react';

const UserStats = ({ users }) => {
    
    // --- 1. TÍNH TOÁN SỐ LIỆU (Logic chuyển từ file cũ sang) ---
    
    // Biểu đồ tròn: Tỷ lệ Admin/User
    const roleStats = useMemo(() => [
        { name: 'Admin', value: users.filter(u => u.role === 'ADMIN').length, color: '#f59e0b' },
        { name: 'User', value: users.filter(u => u.role === 'USER').length, color: '#3b82f6' }
    ], [users]);

    // Biểu đồ cột: Tăng trưởng 6 tháng
    const growthData = useMemo(() => {
        const dataPoints = [];
        const today = new Date();

        for (let i = 5; i >= 0; i--) {
            const d = new Date(today.getFullYear(), today.getMonth() - i, 1);
            const monthLabel = `T${d.getMonth() + 1}`; 
            
            const count = users.filter(user => {
                if (!user.created_at) return false;
                const userDate = new Date(user.created_at);
                return userDate.getMonth() === d.getMonth() && userDate.getFullYear() === d.getFullYear();
            }).length;

            dataPoints.push({ name: monthLabel, users: count });
        }
        return dataPoints;
    }, [users]);

    // --- 2. RENDER GIAO DIỆN ---
    return (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-8">
            {/* Biểu đồ Donut */}
            <div className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl flex flex-col h-64">
                <h3 className="text-sm font-bold text-slate-400 uppercase tracking-widest mb-2 flex items-center gap-2">
                    <PieIcon size={16} className="text-emerald-400"/> Tỷ lệ Phân quyền
                </h3>
                <div className="flex-1">
                    <ResponsiveContainer width="100%" height="100%">
                        <PieChart>
                            <Pie data={roleStats} cx="50%" cy="50%" innerRadius={50} outerRadius={70} paddingAngle={5} dataKey="value">
                                {roleStats.map((entry, index) => <Cell key={`cell-${index}`} fill={entry.color} />)}
                            </Pie>
                            <Tooltip contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '12px', color: '#fff' }} itemStyle={{ color: '#fff' }} />
                        </PieChart>
                    </ResponsiveContainer>
                </div>
                <div className="flex justify-center gap-6 mt-2">
                    {roleStats.map(stat => (
                        <div key={stat.name} className="flex items-center gap-2 text-xs font-bold">
                            <span className="w-3 h-3 rounded-full" style={{ backgroundColor: stat.color }}></span>
                            {stat.name}: <span className="text-white">{stat.value}</span>
                        </div>
                    ))}
                </div>
            </div>

            {/* Biểu đồ Cột */}
            <div className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl flex flex-col h-64">
                <h3 className="text-sm font-bold text-slate-400 uppercase tracking-widest mb-2 flex items-center gap-2">
                    <Activity size={16} className="text-emerald-400"/> Tăng trưởng nhân sự
                </h3>
                <div className="flex-1">
                    <ResponsiveContainer width="100%" height="100%">
                        <BarChart data={growthData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                            <XAxis dataKey="name" stroke="#475569" fontSize={12} tickLine={false} axisLine={false} />
                            <YAxis stroke="#475569" fontSize={12} tickLine={false} axisLine={false} />
                            <Tooltip cursor={{ fill: '#334155', opacity: 0.4 }} contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '12px' }} />
                            <Bar dataKey="users" fill="#10b981" radius={[4, 4, 0, 0]} />
                        </BarChart>
                    </ResponsiveContainer>
                </div>
            </div>
        </div>
    );
};

export default UserStats;