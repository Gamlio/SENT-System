import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { LayoutDashboard, ShieldAlert, Monitor, Globe, Settings, LogOut, Usb, PackageSearch } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

const Sidebar = () => {
    const { user, logout } = useAuth();
    const location = useLocation();

    // Định nghĩa Menu theo Role Level
    const menuItems = [
        { name: 'Dashboard', path: '/', icon: <LayoutDashboard size={20}/>, roles: [1, 2, 3, 4] },
        
        // R1, R2 Quản lý SME
        ...(user?.level <= 2 ? [{ name: 'Quản lý SME', path: '/admin/orgs', icon: <Globe size={20}/> }] : []),
        
        // Tất cả đều thấy Máy trạm nhưng nội dung hiển thị do Backend lọc
        { name: 'Máy trạm', path: '/agents', icon: <Monitor size={20}/>, roles: [1, 2, 3, 4] },
        
        // R3 mới có quyền quản lý chính sách (Whitelist/Software)
        ...(user?.level === 3 ? [
            { name: 'Chính sách USB', path: '/policies/usb', icon: <Usb size={20}/> },
            { name: 'Phần mềm cấm', path: '/policies/software', icon: <PackageSearch size={20}/> }
        ] : []),
        
        { name: 'Cảnh báo', path: '/alerts', icon: <ShieldAlert size={20}/>, roles: [3, 4] }
    ];

    return (
        <div className="w-72 bg-[#1e293b] h-screen flex flex-col border-r border-slate-800 shadow-2xl sticky top-0">
            {/* Logo Section */}
            <div className="p-8">
                <h1 className="text-3xl font-black text-emerald-400 tracking-tighter">SENT<span className="text-slate-500">.</span></h1>
            </div>
            
            <nav className="flex-1 px-4 space-y-2">
                {menuItems.map((item) => (
                    <Link 
                        key={item.path} 
                        to={item.path}
                        className={`flex items-center gap-4 px-4 py-3.5 rounded-xl font-medium transition-all duration-300 ${
                            location.pathname === item.path 
                            ? 'bg-emerald-500 text-white shadow-lg' 
                            : 'text-slate-400 hover:bg-slate-800 hover:text-white'
                        }`}
                    >
                        {item.icon} <span>{item.name}</span>
                    </Link>
                ))}
            </nav>

            {/* User Profile Info */}
            <div className="p-4 mt-auto border-t border-slate-800">
                <div className="bg-slate-900/50 p-4 rounded-2xl mb-4 border border-slate-800">
                    <p className="text-[10px] text-emerald-500 font-bold uppercase">LEVEL {user?.level}</p>
                    <p className="text-sm font-bold text-white truncate">{user?.username}</p>
                </div>
                <button onClick={logout} className="w-full flex items-center justify-center gap-2 p-3 text-red-400 hover:bg-red-500/10 rounded-xl">
                    <LogOut size={18}/> <span className="font-bold text-sm">Đăng xuất</span>
                </button>
            </div>
        </div>
    );
};

export default Sidebar;