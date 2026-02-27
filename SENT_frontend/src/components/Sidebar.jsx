import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { LayoutDashboard, Monitor, MessageSquare, FileText, ShieldCheck, Users } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

const Sidebar = () => {
    const { user } = useAuth();
    const location = useLocation();

    if (!user) return null;

    const isActive = (path) => location.pathname.startsWith(path);

    // Gom menu lại thành mảng để dễ render
    const menuItems = [
       { path: '/', icon: <LayoutDashboard size={20}/>, label: 'Tổng quan', show: true },
        { path: '/agents', icon: <Monitor size={20}/>, label: 'Máy trạm', show: user.permissions?.view_agents },
        { path: '/admin/policy-center', icon: <ShieldCheck size={20}/>, label: 'Trung tâm Chính sách', show: user.permissions?.manage_policies },
        { path: '/admin/docs', icon: <FileText size={20}/>, label: 'Tài liệu AI (Docs)', show: user.permissions?.view_docs },
        { path: '/admin/users', icon: <Users size={20}/>, label: 'Người dùng', show: user.permissions?.manage_users },
        { path: '/incidents', icon: <ShieldCheck size={20}/>, label: 'Điều tra Sự cố', show: user.permissions?.manage_incidents }    
    ];

    return (
        <div className="w-64 bg-[#1e293b] h-screen flex flex-col border-r border-slate-800 shadow-2xl relative z-20">
            {/* Logo Section */}
            <div className="p-6 flex items-center gap-3">
                <div className="p-2 bg-emerald-500/10 rounded-xl text-emerald-400">
                    <ShieldCheck size={28} />
                </div>
                <div>
                    <h1 className="text-xl font-black text-white tracking-wider">SENT <span className="text-emerald-400">SOC</span></h1>
                    <p className="text-[10px] text-slate-500 font-bold uppercase tracking-widest -mt-1">Global System</p>
                </div>
            </div>
            
            {/* Menu List */}
            <nav className="flex-1 px-4 space-y-1 overflow-y-auto mt-4">
                {menuItems.map((item) => (
                    item.show && (
                        <Link 
                            key={item.path} 
                            to={item.path}
                            className={`flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-bold transition-all duration-200 ${
                                (item.path === '/' ? location.pathname === '/' : isActive(item.path))
                                ? 'bg-emerald-500 text-white shadow-lg shadow-emerald-500/20' 
                                : 'text-slate-400 hover:bg-slate-800/50 hover:text-white'
                            }`}
                        >
                            {item.icon} {item.label}
                        </Link>
                    )
                ))}
            </nav>

            {/* Float Button cho AI Chat */}
            <div className="p-4">
                <Link to="/chat-ai" className="flex items-center justify-center gap-2 p-3 w-full bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-xl border border-slate-700 transition font-bold text-sm shadow-xl">
                    <MessageSquare size={18} className="text-blue-400"/> Hỏi AI Copilot
                </Link>
            </div>
        </div>
    );
};

export default Sidebar;