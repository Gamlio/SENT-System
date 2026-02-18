// src/components/Navbar.jsx
import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { LayoutDashboard, Monitor, ShieldAlert, Settings, LogOut, Users } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

const Navbar = () => {
    const location = useLocation();
    const { logout, user } = useAuth();

    // Hàm kiểm tra link đang active
    const isActive = (path) => location.pathname === path 
        ? "bg-emerald-500/10 text-emerald-400" 
        : "text-slate-400 hover:text-emerald-400 hover:bg-slate-800/50";

    const NavItem = ({ to, icon: Icon, label }) => (
        <Link to={to} className={`flex items-center gap-2 px-4 py-2 rounded-xl transition-all font-medium text-sm ${isActive(to)}`}>
            <Icon size={18} />
            <span>{label}</span>
        </Link>
    );

    return (
        <div className="h-16 bg-[#1e293b] border-b border-slate-800 flex items-center justify-between px-6 sticky top-0 z-50 shadow-lg">
            {/* 1. Logo */}
            <div className="flex items-center gap-3">
                <div className="w-8 h-8 bg-emerald-500 rounded-lg flex items-center justify-center shadow-[0_0_15px_rgba(16,185,129,0.5)]">
                    <ShieldAlert className="text-white" size={20} />
                </div>
                <span className="text-xl font-black text-white tracking-wider">SENT<span className="text-emerald-400">.SOC</span></span>
            </div>

            {/* 2. Menu Ngang (Center) */}
            <nav className="hidden md:flex items-center gap-2">
                <NavItem to="/" icon={LayoutDashboard} label="Tổng quan" />
                <NavItem to="/agents" icon={Monitor} label="Máy trạm" />
                
                {/* Chỉ hiện menu Admin nếu là R1/R2 */}
                {user?.level <= 2 && (
                    <NavItem to="/admin/orgs" icon={Users} label="Đối tác SME" />
                )}

                {/* Menu Policies (Ví dụ) */}
                <NavItem to="/policies/software" icon={Settings} label="Chính sách" />
            </nav>

            {/* 3. User Info & Logout (Right) */}
            <div className="flex items-center gap-4">
                <div className="text-right hidden sm:block">
                    <p className="text-xs text-slate-400 font-bold uppercase">Admin</p>
                    <p className="text-xs text-emerald-400 font-mono">{user?.username || 'User'}</p>
                </div>
                <div className="h-8 w-[1px] bg-slate-700"></div>
                <button onClick={logout} className="p-2 text-slate-400 hover:text-red-400 transition-colors" title="Đăng xuất">
                    <LogOut size={20} />
                </button>
            </div>
        </div>
    );
};

export default Navbar;