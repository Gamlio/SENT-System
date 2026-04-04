import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { 
    User, Mail, Briefcase, Phone, Zap, Fingerprint, Activity, Clock,
    Monitor, ShieldCheck, FileText, Server, Shield, Lock, AlertTriangle
} from 'lucide-react';

const Profile = () => {
    const navigate = useNavigate();
    
    // 1. STATE AN TOÀN TRÁNH LỖI UNDEFINED
    const [user, setUser] = useState({
        full_name: 'SOC Analyst', 
        username: 'analyst_01',
        email: 'N/A', 
        phone: 'N/A',
        role: 'analyst', 
        department_tag: 'OFFICE', 
        id: 'UNKNOWN',
        created_at: new Date().toISOString()
    });

    // 2. PARSE LOCALSTORAGE AN TOÀN TRONG USEEFFECT
    useEffect(() => {
        try {
            const userString = localStorage.getItem('user');
            if (userString) {
                const parsedUser = JSON.parse(userString);
                setUser(prev => ({ ...prev, ...parsedUser }));
            }
        } catch (error) {
            console.error("Lỗi parse dữ liệu User:", error);
        }
    }, []);

    const getRoleBadge = (role) => {
        if (role === 'admin') return 'text-purple-400 border-purple-500/30 bg-purple-500/10 shadow-[0_0_10px_rgba(168,85,247,0.2)]';
        if (role === 'manager') return 'text-indigo-400 border-indigo-500/30 bg-indigo-500/10';
        return 'text-blue-400 border-blue-500/30 bg-blue-500/10';
    };

    // 3. HÀM TẠO AVATAR CHỐNG CRASH TOUPPERCASE()
    const safeInitials = () => {
        const name = user?.full_name || user?.username || 'SC';
        return name.substring(0, 2).toUpperCase();
    };

    // 4. HÀM RENDER NGÀY THÁNG AN TOÀN
    const safeDate = () => {
        try {
            return user.created_at ? new Date(user.created_at).toLocaleDateString('vi-VN') : 'Unknown';
        } catch (e) {
            return 'Invalid Date';
        }
    };

    return (
        <div className="h-[calc(100vh-60px)] bg-[#050B14] text-slate-200 p-6 flex flex-col font-sans overflow-hidden">
            
            {/* Header section */}
            <div className="shrink-0 mb-6 pb-6 border-b border-slate-800/80 flex justify-between items-end">
                <div>
                    <h1 className="text-2xl font-black text-white tracking-tighter uppercase flex items-center gap-3">
                        <Fingerprint className="text-indigo-500" size={24}/> Identity Profile
                    </h1>
                    <p className="text-[11px] text-slate-500 mt-1 uppercase tracking-widest font-bold">Authentication & Operations Center</p>
                </div>
                <div className="flex gap-2">
                    <button className="bg-[#111827] border border-slate-700 hover:border-slate-500 text-slate-300 text-[10px] font-black uppercase tracking-widest px-4 py-1.5 rounded-lg transition flex items-center gap-1.5">
                        <Server size={12}/> View Full Audit Trail
                    </button>
                    <button className="bg-indigo-600 hover:bg-indigo-500 text-white text-[10px] font-black uppercase tracking-widest px-4 py-1.5 rounded-lg transition flex items-center gap-1.5 shadow-lg shadow-indigo-500/20">
                        <Shield size={12}/> Revoke Sessions
                    </button>
                </div>
            </div>

            {/* Main content - 50/50 Split */}
            <div className="flex-1 flex gap-6 min-h-0">
                
                {/* 50% Trái: THÔNG TIN CÁ NHÂN */}
                <div className="w-1/2 flex flex-col gap-6 overflow-y-auto custom-scrollbar pr-2">
                    
                    {/* AVATAR HERO CARD */}
                    <div className="bg-[#0A101D] border border-slate-800 rounded-xl p-5 flex items-center gap-5 shadow-2xl relative overflow-hidden group">
                        <div className="absolute top-0 left-0 right-0 h-[3px] bg-gradient-to-r from-indigo-500 via-purple-500 to-indigo-500 opacity-50"></div>
                        <div className="w-16 h-16 rounded bg-indigo-500 flex items-center justify-center font-black text-2xl text-white shadow-xl">
                            {safeInitials()}
                        </div>
                        <div>
                            <span className={`px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-widest border ${getRoleBadge(user.role)}`}>
                                {user.role || 'Analyst'}
                            </span>
                            <h2 className="text-lg font-black text-white mt-1 leading-tight tracking-tight">{user.full_name || user.username || 'System User'}</h2>
                            <p className="text-[11px] text-emerald-400 font-mono flex items-center gap-1.5 mt-0.5 bg-emerald-950/20 w-max px-2 py-0.5 rounded-full">
                                <ShieldCheck size={12} className="animate-pulse"/> Active Duty <span className="text-slate-700">•</span> @{user.department_tag || 'OFFICE'}
                            </p>
                        </div>
                        <Fingerprint size={80} className="absolute -right-5 -bottom-5 text-indigo-500/5 rotate-12 group-hover:scale-110 transition-transform duration-500"/>
                    </div>

                    {/* CONTACT DETAILS */}
                    <div className="bg-[#0A101D] border border-slate-800 rounded-xl shadow-lg p-5">
                        <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-4 flex items-center gap-2 border-b border-slate-800 pb-2.5">
                            <Activity size={12} className="text-indigo-400"/> Operational Context
                        </h3>
                        <div className="grid grid-cols-2 gap-4">
                            <InfoCard icon={<Lock/>} label="User ID" value={`#${user.id}`} mono />
                            <InfoCard icon={<Briefcase/>} label="Dept/Context" value={user.department_tag || 'OFFICE'} />
                            <InfoCard icon={<User/>} label="Username" value={`@${user.username}`} mono />
                            <InfoCard icon={<Mail/>} label="Email" value={user.email} mono />
                            <InfoCard icon={<Phone/>} label="Phone" value={user.phone} mono />
                            <InfoCard icon={<Clock/>} label="Joined Ops" value={safeDate()} />
                        </div>
                    </div>

                </div>

                {/* 50% Phải: HIỆU SUẤT LÀM DÀY */}
                <div className="w-1/2 flex flex-col gap-6 overflow-y-auto custom-scrollbar pr-2">
                    
                    {/* SOC Performance Grid */}
                    <div className="bg-[#0A101D] border border-slate-800 rounded-xl p-5 shadow-2xl">
                        <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-4 flex items-center gap-2 border-b border-slate-800 pb-2.5">
                            <Activity size={12} className="text-orange-400"/> L1/L2 OPS Metrics
                        </h3>
                        <div className="grid grid-cols-3 gap-3">
                            <StatBox icon={<AlertTriangle/>} title="Incidents Handled" value="128" sub="Closed cases" color="red" />
                            <StatBox icon={<Monitor/>} title="assets Assigned" value="45" sub="Managed endpoints" color="blue" />
                            <StatBox icon={<Zap/>} title="Avg. Triage Time" value="7.2m" sub="Response SLA" color="amber" />
                            <StatBox icon={<Lock/>} title="Policy Approvals" value="19" sub="Makers process" color="purple" />
                            <StatBox icon={<FileText/>} title="Open Tickets" value="3" sub="Awaiting response" color="emerald" />
                            <StatBox icon={<ShieldCheck/>} title="Risk Score Baseline" value="Medium" sub="Current profile" color="slate" />
                        </div>
                    </div>

                    {/* LIVE AUDIT FEED */}
                    <div className="bg-[#0A101D] border border-slate-800 rounded-xl p-5 shadow-inner">
                        <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-3.5 flex items-center gap-2 border-b border-slate-800 pb-2.5">
                            <Activity size={12} className="text-slate-600"/> Latest Session Audits
                        </h3>
                        <div className="space-y-2">
                            {auditLogs.map((log, idx) => (
                                <div key={idx} className="bg-[#111827] border border-slate-800/80 p-2.5 rounded-lg flex justify-between items-center text-[10px]">
                                    <div className="flex items-center gap-2.5">
                                        <div className={`shrink-0 p-1 rounded border ${log.type === 'ACCESS' ? 'text-emerald-400 border-emerald-500/30' : 'text-blue-400 border-blue-500/30'}`}>
                                            {log.icon}
                                        </div>
                                        <div>
                                            <p className="text-slate-300 font-black">{log.action}</p>
                                            <p className="text-[9px] text-slate-500 font-mono">@{log.ip}</p>
                                        </div>
                                    </div>
                                    <span className="text-slate-600 font-mono italic">{log.time}</span>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>

            </div>
        </div>
    );
};

// COMPONENT HỖ TRỢ TRÁNH CRASH REACT.CLONEELEMENT
const InfoCard = ({ icon, label, value, mono }) => {
    // Bọc cẩn thận để tránh truyền undefined icon vào CloneElement
    const renderIcon = () => {
        if (!icon) return <Lock size={12} className="group-hover:text-indigo-400" />;
        try {
            return React.cloneElement(icon, { size: 12, className: "group-hover:text-indigo-400" });
        } catch (e) {
            return <Lock size={12} className="group-hover:text-indigo-400" />;
        }
    };

    return (
        <div className="bg-[#111827] border border-slate-800 rounded-lg p-3 flex flex-col gap-0.5 shadow-sm group hover:border-indigo-500 transition duration-300">
            <p className="text-[9px] font-black uppercase text-slate-600 tracking-widest flex items-center gap-1.5 group-hover:text-indigo-400 transition-colors">
                {renderIcon()}
                {label}
            </p>
            <p className={`text-[11px] font-bold text-slate-300 truncate ${mono ? 'font-mono' : ''}`} title={value}>
                {value || 'N/A'}
            </p>
        </div>
    );
};

const StatBox = ({ icon, title, value, sub, color }) => {
    const colorMap = {
        red: 'text-red-400 border-red-500/30 bg-red-500/10',
        blue: 'text-blue-400 border-blue-500/30 bg-blue-500/10',
        amber: 'text-amber-400 border-amber-500/30 bg-amber-500/10',
        purple: 'text-purple-400 border-purple-500/30 bg-purple-500/10',
        emerald: 'text-emerald-400 border-emerald-500/30 bg-emerald-500/10',
        slate: 'text-slate-400 border-slate-500/30 bg-slate-500/10',
    };
    return (
        <div className="bg-[#111827] border border-slate-800 rounded-lg p-3 group flex items-start gap-2.5">
            <div className={`p-1.5 rounded border ${colorMap[color]}`}>{icon || <Activity size={16}/>}</div>
            <div>
                <p className="text-2xl font-black text-white font-mono tracking-tight leading-none mb-0.5">{value}</p>
                <p className="text-[9px] text-slate-300 font-black uppercase group-hover:text-indigo-400 transition-colors leading-tight">{title}</p>
                <p className="text-[8px] text-slate-600 font-mono italic leading-tight">{sub}</p>
            </div>
        </div>
    );
};

const auditLogs = [
    { type: 'ACTION', icon: <FileText size={12}/>, action: 'Closed Incident #1098', ip: '10.0.98.12', time: '12m ago' },
    { type: 'ACCESS', icon: <Fingerprint size={12}/>, action: 'User Session Login', ip: '10.0.98.12', time: '1h 10m ago' },
    { type: 'ACTION', icon: <Activity size={12}/>, action: 'Trigger Baseline on Win-Prod', ip: '10.0.98.12', time: '2h 15m ago' },
];

export default Profile;