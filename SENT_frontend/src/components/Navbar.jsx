import React from 'react';
import { useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { LogOut, Bell, Search } from 'lucide-react';

const Navbar = () => {
    const { user, logout } = useAuth();
    const location = useLocation();

    if (!user) return null;

    // Tạo Breadcrumb động dựa trên URL
    const getPageTitle = () => {
        const path = location.pathname;
        if (path === '/') return 'Tổng quan hệ thống';
        if (path.startsWith('/agents')) return 'Quản lý Máy trạm';
        if (path.startsWith('/admin/policy-center')) return 'Trung tâm Chính sách';
        if (path.startsWith('/admin/docs')) return 'Kho tri thức AI';
        if (path.startsWith('/admin/users')) return 'Quản lý Người dùng';
        if (path.startsWith('/chat-ai')) return 'Trợ lý AI Copilot';
        return 'Bảng điều khiển';
    };

    return (
        <header className="bg-[#1e293b]/80 backdrop-blur-md border-b border-slate-800 px-8 py-4 sticky top-0 z-10 flex justify-between items-center">
            
            {/* Trái: Page Title */}
            <div>
                <h2 className="text-lg font-bold text-white">{getPageTitle()}</h2>
                <p className="text-xs text-slate-500">Real-time Security Operations Center</p>
            </div>

            {/* Phải: Tools & Profile */}
            <div className="flex items-center gap-6">
                
                {/* Search & Alerts (Trang trí cho giống thật) */}
                <div className="flex items-center gap-4 text-slate-400">
                    <button className="hover:text-white transition"><Search size={20}/></button>
                    <button className="hover:text-white transition relative">
                        <Bell size={20}/>
                        <span className="absolute -top-1 -right-1 w-2.5 h-2.5 bg-red-500 rounded-full"></span>
                    </button>
                </div>

                <div className="w-px h-6 bg-slate-700"></div>

                {/* User Profile */}
                <div className="flex items-center gap-4">
                    
                    {/* HIỂN THỊ MÃ CÔNG TY ĐỂ NHÂN VIÊN COPY */}
                    <div className="hidden md:flex items-center gap-2 px-3 py-1.5 bg-slate-900 rounded-xl border border-slate-700 shadow-inner">
                        <span className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Mã CTY:</span>
                        <span className="text-sm font-mono font-black text-emerald-400 select-all cursor-pointer" title="Click đúp để bôi đen copy">
                            {user.company_code || 'N/A'}
                        </span>
                    </div>

                    <div className="text-right hidden md:block ml-2">
                        <p className="text-sm font-bold text-white leading-tight">{user.username}</p>
                        <p className="text-[10px] font-black text-emerald-400 uppercase tracking-widest">
                            Level {user.level}
                        </p>
                    </div>
                    
                    <button 
                        onClick={logout}
                        className="p-2 bg-slate-800 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-xl transition border border-slate-700"
                        title="Đăng xuất"
                    >
                        <LogOut size={18} />
                    </button>
                </div>
            </div>
        </header>
    );
};

export default Navbar;