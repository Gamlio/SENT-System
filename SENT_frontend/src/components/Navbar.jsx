import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useNavigate, useLocation, Link } from 'react-router-dom';
import { 
    LayoutDashboard, Monitor, ShieldAlert, LogOut, Search, Command, 
    Bell, ChevronDown, User, Settings, Lock, X, Activity,
    Zap, Cpu, AlertTriangle, Bot, CheckCircle,
} from 'lucide-react';

import { useSocketSubscription } from '../context/useSocketSubscription';

const notificationsData = [
    { id: 1, type: 'INCIDENT', status: 'UNREAD', title: 'Máy trạm #A098 cắm USB lạ', severity: 'High', time: '1p trước' },
    { id: 2, type: 'ALERT', status: 'UNREAD', title: 'Quá tải CPU trên Server #PROD-01', severity: 'Critical', time: '5p trước' },
    { id: 3, type: 'SYSTEM', status: 'READ', title: 'Baseline completed on #WIN-HANA', severity: 'Info', time: '12p trước' },
    { id: 4, type: 'POLICY', status: 'UNREAD', title: 'Cập nhật Policy Zero-Trust cho Sales', severity: 'Medium', time: '1h trước' },
];

const Navbar = ({ onOpenCopilot }) => {
    const navigate = useNavigate();
    const location = useLocation();
    
    const [showNotifPanel, setShowNotifPanel] = useState(false);
    const [notifications, setNotifications] = useState(notificationsData);
    const [unreadCount, setUnreadCount] = useState(0);
    const notifRef = useRef(null);

    const [showUserDropdown, setShowUserDropdown] = useState(false);
    const userDropdownRef = useRef(null);

    // [FIX-CRASH] An toàn khi parse dữ liệu user từ localStorage
    const user = (() => {
        try {
            const userString = localStorage.getItem('user');
            return userString ? JSON.parse(userString) : { full_name: 'SOC Analyst', role: 'admin', department_tag: 'SOC_L1' };
        } catch (error) {
            console.error("Lỗi parse dữ liệu user từ localStorage:", error);
            return { full_name: 'SOC Analyst', role: 'admin', department_tag: 'SOC_L1' }; // Fallback
        }
    })();

    // CẬP NHẬT SỐ THÔNG BÁO CHƯA ĐỌC
    useEffect(() => {
        setUnreadCount(notifications.filter(n => n.status === 'UNREAD').length);
    }, [notifications]);

    // ĐÓNG PANEL KHI CLICK RA NGOÀI
    useEffect(() => {
        const handleClickOutside = (event) => {
            if (notifRef.current && !notifRef.current.contains(event.target)) { setShowNotifPanel(false); }
            if (userDropdownRef.current && !userDropdownRef.current.contains(event.target)) { setShowUserDropdown(false); }
        };
        document.addEventListener("mousedown", handleClickOutside);
        return () => document.removeEventListener("mousedown", handleClickOutside);
    }, []);

    // -----------------------------------------------------------------
    // LẮNG NGHE SỰ KIỆN LIVE QUA SOCKET (TRUE REAL-TIME)
    // -----------------------------------------------------------------
    const handleNewLiveEvent = useCallback((payload) => {
        // console.log("🔥 Bắt được sóng Live Event trong Navbar:", payload);
        
        const newNotif = {
            id: Date.now(),
            type: payload.type || 'ALERT',
            status: 'UNREAD',
            title: payload.content || 'Cảnh báo vi phạm mới nhận được',
            severity: payload.priority || 'Critical', // P1 -> Critical
            time: 'Vừa xong'
        };

        // Bơm thông báo mới lên đầu danh sách
        setNotifications(prev => [newNotif, ...prev.slice(0, 19)]); // Giữ tối đa 20 cái
    }, []);

    useSocketSubscription(['NEW_INCIDENT', 'asset_STATUS_CHANGED'], handleNewLiveEvent);
    // -----------------------------------------------------------------

    const handleLogout = () => {
        localStorage.clear();
        navigate('/login');
    };

    const markAllAsRead = () => {
        setNotifications(prev => prev.map(n => ({ ...n, status: 'READ' })));
    };

    return (
        <nav className="h-[60px] bg-[#0A101D] border-b border-slate-800 flex items-center justify-between px-6 shrink-0 relative z-50 shadow-md font-sans">
            
            <div className="flex items-center gap-3 cursor-pointer" onClick={() => navigate('/') }>
                <div className="w-10 h-10 rounded-xl overflow-hidden bg-[#111827] border border-slate-700 shadow-[0_0_15px_rgba(99,102,241,0.3)]">
                    <img src="/sent.png" alt="SENT SOC" className="w-full h-full object-cover" />
                </div>
                <div>
                    <h1 className="text-xl font-extrabold text-white tracking-tighter">SENT</h1>
                </div>
            </div>

           
            <div className="flex items-center gap-3">
                
                <button
                    onClick={onOpenCopilot}
                    className="p-2.5 rounded-lg border border-slate-700 hover:border-indigo-500 hover:bg-slate-800 transition group"
                    aria-label="Open Copilot"
                >
                    <Command size={16} className="text-slate-500 group-hover:text-indigo-400" />
                </button>

                {/* NOTIFICATION SECTION (LIVE) */}
                <div className="relative" ref={notifRef}>
                    <button 
                        onClick={() => setShowNotifPanel(!showNotifPanel)}
                        className={`p-2.5 rounded-lg border border-slate-700 hover:border-indigo-500 hover:bg-slate-800 transition group ${showNotifPanel ? 'bg-slate-800 border-indigo-500' : ''}`}
                    >
                        <Bell size={16} className={`text-slate-500 group-hover:text-indigo-400 ${showNotifPanel ? 'text-indigo-400' : ''}`} />
                        {unreadCount > 0 && (
                            <span className="absolute -top-1.5 -right-1.5 min-w-[18px] h-[18px] px-1 bg-red-600 rounded-full text-white text-[9px] font-black font-mono flex items-center justify-center border-2 border-[#0A101D] animate-pulse">
                                {unreadCount > 9 ? '9+' : unreadCount}
                            </span>
                        )}
                    </button>

                    {/* LIVE NOTIFICATION PANEL */}
                    {showNotifPanel && (
                        <div className="absolute right-0 top-full mt-2 w-[350px] bg-[#0A101D] rounded-xl border border-slate-700 shadow-[0_10px_40px_rgba(0,0,0,0.8)] overflow-hidden animate-in fade-in zoom-in-95 duration-100 flex flex-col">
                            {/* Panel Header */}
                            <div className="px-4 py-3 border-b border-slate-800 bg-[#111827] flex justify-between items-center">
                                <h3 className="text-[11px] font-black uppercase tracking-widest text-white flex items-center gap-2">
                                    <Bot size={14} className="text-indigo-400"/> Live Operations Feed
                                </h3>
                                <div className="flex items-center gap-2">
                                    <button onClick={markAllAsRead} className="text-[10px] text-slate-500 hover:text-white transition uppercase font-black">Đánh dấu đã đọc</button>
                                    <button onClick={() => setShowNotifPanel(false)} className="text-slate-500 hover:text-white transition"><X size={16}/></button>
                                </div>
                            </div>

                            {/* Panel Body (Scrollable) */}
                            <div className="flex-1 max-h-[350px] overflow-y-auto custom-scrollbar bg-[#050B14]">
                                {notifications.length === 0 ? (
                                    <div className="p-8 text-center text-slate-600 text-[11px] font-mono italic">
                                        SYSTEMS SECURE. NO RECENT INCIDENTS.
                                    </div>
                                ) : notifications.map(n => (
                                    <div key={n.id} className={`p-3 border-b border-slate-800 flex items-start gap-3 hover:bg-slate-800/40 cursor-pointer ${n.status === 'UNREAD' ? 'bg-[#111827]/50' : ''}`}>
                                        <div className={`shrink-0 p-1.5 rounded border ${n.severity === 'Critical' ? 'bg-red-500/10 text-red-500 border-red-500/30 shadow-[0_0_8px_rgba(239,68,68,0.2)]' : n.severity === 'High' ? 'bg-orange-500/10 text-orange-400 border-orange-500/30' : 'bg-blue-500/10 text-blue-400 border-blue-500/30'}`}>
                                            {n.severity === 'Critical' ? <AlertTriangle size={14}/> : <Zap size={14}/>}
                                        </div>
                                        <div className="flex-1">
                                            <p className="text-[11px] text-slate-100 leading-relaxed"><span className={`font-black font-mono mr-1.5 ${n.severity === 'Critical' ? 'text-red-400' : 'text-orange-400'}`}>{n.severity}</span>{n.title}</p>
                                            <div className="flex justify-between items-center mt-1">
                                                <p className="text-[9px] text-slate-500 font-black uppercase tracking-widest">{n.type}</p>
                                                <p className="text-[9px] text-slate-600 font-mono italic">{n.time}</p>
                                            </div>
                                        </div>
                                        {n.status === 'UNREAD' && <div className="w-2 h-2 rounded-full bg-blue-500 shrink-0 mt-1.5 shadow-[0_0_8px_rgba(59,130,246,0.6)] animate-pulse"></div>}
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}
                </div>

                {/* USER DROPDOWN SECTION */}
                <div className="relative" ref={userDropdownRef}>
                    <button 
                        onClick={() => setShowUserDropdown(!showUserDropdown)}
                        className={`pl-3 pr-2 py-1.5 rounded-lg border transition-colors flex items-center gap-2 group ${showUserDropdown ? 'bg-slate-800 border-indigo-500' : 'bg-[#111827] border-slate-700 hover:border-indigo-500'}`}
                    >
                        <div className="w-7 h-7 rounded bg-indigo-500/10 border border-indigo-500/30 text-indigo-400 flex items-center justify-center font-black text-xs">
                            {user.full_name?.substring(0,2).toUpperCase()}
                        </div>
                        <div className="text-left leading-tight hidden md:block">
                            <p className="text-xs font-black text-white">{user.full_name}</p>
                            <p className="text-[9px] text-slate-500 font-mono tracking-widest uppercase">@{user.department_tag || 'OFFICE'}</p>
                        </div>
                        <ChevronDown size={14} className={`text-slate-600 group-hover:text-indigo-400 transition-transform ${showUserDropdown ? 'rotate-180 text-indigo-400' : ''}`}/>
                    </button>

                    {/* DROPDOWN MENU */}
                    {showUserDropdown && (
                        <div className="absolute right-0 top-full mt-2 w-48 bg-[#0A101D] rounded-xl border border-slate-700 shadow-[0_10px_40px_rgba(0,0,0,0.8)] overflow-hidden animate-in fade-in zoom-in-95 duration-100">
                            
                            {/* Menu Header */}
                            <div className="px-4 py-3 border-b border-slate-800 bg-[#111827]">
                                <p className="text-[9px] font-black text-slate-500 uppercase tracking-widest">Account & Ops</p>
                                <p className="text-xs font-bold text-white mt-1 truncate">{user.full_name}</p>
                            </div>
                            
                            {/* Menu Items */}
                            <div className="p-1.5 space-y-0.5">
                                {/* ADD LINK TO PROFILE */}
                                <Link to="/profile" onClick={() => setShowUserDropdown(false)} className="w-full flex items-center gap-2.5 px-3 py-2 text-[11px] text-slate-300 hover:bg-slate-800 hover:text-white rounded-lg transition-colors font-bold">
                                    <User size={14} className="text-slate-500"/> Personal Profile
                                </Link>
                                <button className="w-full flex items-center gap-2.5 px-3 py-2 text-[11px] text-slate-300 hover:bg-slate-800 hover:text-white rounded-lg transition-colors font-bold">
                                    <Settings size={14} className="text-slate-500"/> Preferences
                                </button>
                                <button className="w-full flex items-center gap-2.5 px-3 py-2 text-[11px] text-slate-300 hover:bg-slate-800 hover:text-white rounded-lg transition-colors font-bold">
                                    <Lock size={14} className="text-slate-500"/> Session & Access
                                </button>
                            </div>

                            {/* Menu Footer */}
                            <div className="border-t border-slate-800 p-1.5 bg-[#050B14]">
                                <button 
                                    onClick={handleLogout}
                                    className="w-full flex items-center gap-2.5 px-3 py-2 text-[11px] text-red-400 hover:bg-red-500/10 rounded-lg transition-colors font-bold"
                                >
                                    <LogOut size={14}/> TERMINATE SESSION
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </nav>
    );
};

export default Navbar;