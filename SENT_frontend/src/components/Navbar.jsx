import React, { useState, useRef, useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { 
    LogOut, Bell, Search, Sparkles, User, 
    Settings, ShieldAlert, Palette, UserCog, CheckCircle2 
} from 'lucide-react';

const Navbar = ({ onOpenCopilot }) => {
    const { user, logout } = useAuth();
    const location = useLocation();

    // --- STATE QUẢN LÝ DROPDOWN ---
    const [isNotifOpen, setIsNotifOpen] = useState(false);
    const [isProfileOpen, setIsProfileMenuOpen] = useState(false);

    // Refs để bắt sự kiện click ra ngoài
    const notifRef = useRef(null);
    const profileRef = useRef(null);

    // Xử lý đóng menu khi click ra ngoài
    useEffect(() => {
        const handleClickOutside = (event) => {
            if (notifRef.current && !notifRef.current.contains(event.target)) setIsNotifOpen(false);
            if (profileRef.current && !profileRef.current.contains(event.target)) setIsProfileMenuOpen(false);
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    if (!user) return null;

    // Breadcrumb động
    const getPageTitle = () => {
        const path = location.pathname;
        if (path === '/') return 'Tổng quan hệ thống';
        if (path.startsWith('/agents')) return 'Quản lý Máy trạm';
        if (path.startsWith('/incidents')) return 'Trung tâm Sự cố';
        if (path.startsWith('/admin/policy-center')) return 'Trung tâm Chính sách';
        if (path.startsWith('/admin/docs')) return 'Kho tri thức AI';
        if (path.startsWith('/admin/users')) return 'Quản lý Người dùng';
        if (path.startsWith('/chat-ai')) return 'Trợ lý AI Copilot';
        return 'Bảng điều khiển';
    };

    // Dữ liệu thông báo giả lập (Sau này gọi API lấy từ DB)
    const mockNotifications = [
        { id: 1, type: 'P1', text: 'LAPTOP-ATGELDMK vừa tắt Tường lửa', time: 'Vài giây trước' },
        { id: 2, type: 'P2', text: 'Phát hiện mã độc trên máy Kế toán 01', time: '5 phút trước' },
        { id: 3, type: 'INFO', text: 'Hệ thống vừa cập nhật Playbook mới', time: '1 giờ trước' },
    ];

    return (
        <header className="bg-[#1e293b]/90 backdrop-blur-md border-b border-slate-800 px-8 py-4 sticky top-0 z-40 flex justify-between items-center">
            
            {/* Trái: Page Title */}
            <div>
                <h2 className="text-lg font-bold text-white">{getPageTitle()}</h2>
                <p className="text-xs text-slate-500">Real-time Security Operations Center</p>
            </div>

            {/* Phải: Tools & Profile */}
            <div className="flex items-center gap-6">
                
                {/* 1. NÚT GỌI AI */}
                <button 
                    onClick={onOpenCopilot} 
                    className="flex items-center gap-2 bg-indigo-600/10 hover:bg-indigo-600/20 border border-indigo-500/30 text-indigo-400 px-4 py-1.5 rounded-full transition-all hover:scale-105"
                >
                    <Sparkles size={16} />
                    <span className="text-xs font-bold tracking-wide">Ask AI</span>
                </button>
                
                <div className="flex items-center gap-4 text-slate-400 relative">
                    <button className="hover:text-white transition"><Search size={20}/></button>
                    
                    {/* 2. THANH THÔNG BÁO (NOTIFICATION) */}
                    <div ref={notifRef} className="relative">
                        <button 
                            onClick={() => { setIsNotifOpen(!isNotifOpen); setIsProfileMenuOpen(false); }}
                            className={`transition relative p-1.5 rounded-lg ${isNotifOpen ? 'bg-slate-800 text-white' : 'hover:text-white hover:bg-slate-800/50'}`}
                        >
                            <Bell size={20}/>
                            <span className="absolute top-1 right-1.5 w-2.5 h-2.5 bg-red-500 rounded-full border-2 border-[#1e293b] animate-pulse"></span>
                        </button>

                        {/* Dropdown Thông báo */}
                        {isNotifOpen && (
                            <div className="absolute right-0 top-full mt-3 w-80 bg-[#0f172a] border border-slate-700 rounded-2xl shadow-[0_10px_40px_-10px_rgba(0,0,0,0.5)] overflow-hidden animate-in fade-in slide-in-from-top-2">
                                <div className="p-4 border-b border-slate-800 flex justify-between items-center bg-slate-900/50">
                                    <h3 className="font-bold text-white text-sm">Thông báo hệ thống</h3>
                                    <span className="text-[10px] text-indigo-400 font-bold bg-indigo-500/10 px-2 py-0.5 rounded-full">3 Mới</span>
                                </div>
                                <div className="max-h-[300px] overflow-y-auto scrollbar-thin scrollbar-thumb-slate-700">
                                    {mockNotifications.map(notif => (
                                        <div key={notif.id} className="p-4 border-b border-slate-800/50 hover:bg-slate-800/50 transition cursor-pointer flex gap-3">
                                            <div className="mt-0.5 shrink-0">
                                                {notif.type === 'P1' ? <ShieldAlert size={16} className="text-red-500"/> :
                                                 notif.type === 'P2' ? <ShieldAlert size={16} className="text-orange-500"/> :
                                                 <CheckCircle2 size={16} className="text-emerald-500"/>}
                                            </div>
                                            <div>
                                                <p className="text-sm text-slate-200 mb-1 leading-snug">{notif.text}</p>
                                                <p className="text-[10px] text-slate-500 font-mono">{notif.time}</p>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                                <div className="p-3 text-center bg-slate-900/50 hover:bg-slate-800 transition cursor-pointer">
                                    <span className="text-xs font-bold text-indigo-400">Xem tất cả</span>
                                </div>
                            </div>
                        )}
                    </div>
                </div>

                <div className="w-px h-6 bg-slate-700"></div>

                {/* 3. MENU TÀI KHOẢN (PROFILE) */}
                <div className="flex items-center gap-4 relative" ref={profileRef}>
                    
                    {/* Mã CTY */}
                    <div className="hidden md:flex items-center gap-2 px-3 py-1.5 bg-slate-900 rounded-xl border border-slate-700 shadow-inner">
                        <span className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Mã CTY:</span>
                        <span className="text-sm font-mono font-black text-emerald-400 select-all cursor-pointer" title="Click đúp để bôi đen copy">
                            {user.company_code || 'N/A'}
                        </span>
                    </div>

                    {/* Nút bấm mở Menu Tài khoản */}
                    <button 
                        onClick={() => { setIsProfileMenuOpen(!isProfileOpen); setIsNotifOpen(false); }}
                        className={`flex items-center gap-3 text-left pl-2 pr-1 py-1 rounded-xl transition ${isProfileOpen ? 'bg-slate-800' : 'hover:bg-slate-800/50'}`}
                    >
                        <div className="hidden md:block">
                            <p className="text-sm font-bold text-white leading-tight">{user.username}</p>
                            <p className="text-[10px] font-black text-emerald-400 uppercase tracking-widest">{user.role}</p>
                        </div>
                        <div className="w-9 h-9 rounded-lg bg-indigo-600 flex items-center justify-center text-white shadow-md border border-indigo-400">
                            <User size={18}/>
                        </div>
                    </button>

                    {/* Dropdown Menu Tài khoản */}
                    {isProfileOpen && (
                        <div className="absolute right-0 top-full mt-3 w-56 bg-[#0f172a] border border-slate-700 rounded-2xl shadow-[0_10px_40px_-10px_rgba(0,0,0,0.5)] overflow-hidden animate-in fade-in slide-in-from-top-2">
                            <div className="p-4 border-b border-slate-800 bg-slate-900/50">
                                <p className="text-sm font-bold text-white truncate">{user.full_name || user.username}</p>
                                <p className="text-xs text-slate-500 truncate mt-0.5">{user.email || 'Chưa cập nhật email'}</p>
                            </div>
                            
                            <div className="p-2 space-y-1">
                                <button className="w-full flex items-center gap-3 px-3 py-2.5 text-sm text-slate-300 hover:text-white hover:bg-slate-800 rounded-xl transition">
                                    <UserCog size={16} className="text-blue-400"/>
                                    Cập nhật thông tin
                                </button>
                                <button className="w-full flex items-center gap-3 px-3 py-2.5 text-sm text-slate-300 hover:text-white hover:bg-slate-800 rounded-xl transition">
                                    <Palette size={16} className="text-purple-400"/>
                                    Cài đặt giao diện
                                </button>
                                <button className="w-full flex items-center gap-3 px-3 py-2.5 text-sm text-slate-300 hover:text-white hover:bg-slate-800 rounded-xl transition">
                                    <Settings size={16} className="text-slate-400"/>
                                    Cài đặt hệ thống
                                </button>
                            </div>

                            <div className="p-2 border-t border-slate-800">
                                <button 
                                    onClick={logout}
                                    className="w-full flex items-center justify-center gap-2 px-3 py-2.5 text-sm font-bold text-red-400 hover:text-white bg-red-500/10 hover:bg-red-500 rounded-xl transition"
                                >
                                    <LogOut size={16} /> Đăng xuất
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </header>
    );
};

export default Navbar;