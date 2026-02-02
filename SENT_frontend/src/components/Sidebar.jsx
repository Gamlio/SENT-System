import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { LayoutDashboard, ShieldAlert, Monitor, Globe, Settings, LogOut } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

const Sidebar = () => {
    const { user, logout } = useAuth();
    const location = useLocation();

    const menuItems = [
       ...(user.level === 1 ? [{ name: 'Quản lý SME', path: '/admin/orgs' }] : []),
  
  // Level 3 & 4: Chỉ thấy Agent thuộc Org mình [cite: 7, 50, 52]
  ...(user.level >= 3 ? [{ name: 'Máy trạm', path: '/agents' }] : []),
  
  // Logic dựa trên Rules (Ví dụ: quyền quản lý Role)
  ...(user.permissions?.can_manage_roles ? [{ name: 'Thiết lập Rules', path: '/roles' }] : [])
];
    return (
        <div className="w-72 bg-[#1e293b] h-screen flex flex-col border-r border-slate-800 shadow-2xl sticky top-0">
            <div className="p-8">
                <h1 className="text-3xl font-black text-emerald-400 tracking-tighter">SENT<span className="text-slate-500">.</span></h1>
                <p className="text-[10px] text-slate-500 font-bold uppercase tracking-widest mt-1">Security Intelligence</p>
            </div>
            
            <nav className="flex-1 px-4 space-y-2">
                {menuItems.map((item) => (
                    <Link 
                        key={item.path} 
                        to={item.path}
                        className={`flex items-center gap-4 px-4 py-3.5 rounded-xl font-medium transition-all duration-300 ${
                            location.pathname === item.path 
                            ? 'bg-emerald-500 text-white shadow-lg shadow-emerald-500/20' 
                            : 'text-slate-400 hover:bg-slate-800 hover:text-white'
                        }`}
                    >
                        {item.icon} <span>{item.name}</span>
                    </Link>
                ))}
            </nav>

            <div className="p-4 mt-auto border-t border-slate-800">
                <div className="bg-slate-900/50 p-4 rounded-2xl mb-4 border border-slate-800">
                    <p className="text-xs text-slate-500 uppercase font-bold tracking-wider mb-1">Đang đăng nhập</p>
                    <p className="text-sm font-bold text-white truncate">{user?.username}</p>
                    <p className="text-[10px] text-emerald-500 font-bold">LEVEL {user?.level}</p>
                </div>
                <button onClick={logout} className="w-full flex items-center justify-center gap-2 p-3 text-red-400 hover:bg-red-500/10 rounded-xl transition-all">
                    <LogOut size={18}/> <span className="font-bold text-sm">Đăng xuất</span>
                </button>
            </div>
        </div>
    );
};

export default Sidebar;