import React from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { LayoutDashboard, Monitor, MessageSquare, FileText, LogOut, ShieldCheck, Users } from 'lucide-react';

const Navbar = () => {
    const { user, logout } = useAuth();
    const location = useLocation();
    const navigate = useNavigate();

    if (!user) return null;

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    const isActive = (path) => location.pathname === path;

    return (
        <nav className="bg-[#1e293b] border-b border-slate-800 px-8 py-4 sticky top-0 z-50 shadow-xl">
            <div className="max-w-7xl mx-auto flex justify-between items-center">
                
                {/* LOGO & BRAND */}
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-emerald-500/10 rounded-xl text-emerald-400">
                        <ShieldCheck size={28} />
                    </div>
                    <div>
                        <h1 className="text-xl font-black text-white tracking-wider">SENT <span className="text-emerald-400">SOC</span></h1>
                        <p className="text-[10px] text-slate-500 font-bold uppercase tracking-widest -mt-1">Global System</p>
                    </div>
                </div>

                {/* MAIN MENU */}
                <div className="flex items-center gap-2">
                    <NavItem to="/" icon={<LayoutDashboard size={18} />} label="Tổng quan" active={isActive('/')} />
                    <NavItem to="/agents" icon={<Monitor size={18} />} label="Máy trạm" active={location.pathname.startsWith('/agents')} />
                    <NavItem to="/admin/policy-center" icon={<ShieldCheck size={18} />}  label="Trung tâm Chính sách"   active={location.pathname.startsWith('/admin/policy-center')}    />
                    <NavItem to="/chat-ai" icon={<MessageSquare size={18} />} label="AI Trợ lý" active={isActive('/chat-ai')} />

                    {/* Các menu CHỈ ADMIN mới thấy */}
                    {user.level >= 2 && (
                        <>
                            <div className="w-px h-6 bg-slate-700 mx-2"></div> 
                            <NavItem to="/admin/docs" icon={<FileText size={18} />} label="Tài liệu" active={isActive('/admin/docs')} isSpecial />
                            <NavItem to="/admin/users" icon={<Users size={18} />} label="Người dùng" active={isActive('/admin/users')} />
                        </>
                    )}
                </div>

                {/* USER PROFILE & LOGOUT */}
                <div className="flex items-center gap-4 pl-6 border-l border-slate-700">
                    <div className="text-right hidden md:block">
                        <p className="text-sm font-bold text-white">{user.username}</p>
                        <p className="text-[10px] font-black text-emerald-400 uppercase tracking-widest">
                            {user.level >= 2 ? 'Quản trị viên' : 'Nhân viên'}
                        </p>
                    </div>
                    <button 
                        onClick={handleLogout}
                        className="p-2 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-xl transition"
                        title="Đăng xuất"
                    >
                        <LogOut size={20} />
                    </button>
                </div>

            </div>
        </nav>
    );
};

// Component con 1: Nút bấm đơn thường
const NavItem = ({ to, icon, label, active, isSpecial }) => {
    return (
        <Link 
            to={to} 
            className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-bold transition-all duration-200
                ${active 
                    ? (isSpecial ? 'bg-emerald-500 text-white shadow-lg shadow-emerald-500/20' : 'bg-slate-800 text-white') 
                    : 'text-slate-400 hover:text-white hover:bg-slate-800/50'
                }
            `}
        >
            {icon} {label}
        </Link>
    );
};

export default Navbar;