import React, { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { 
    LayoutDashboard, Monitor, MessageSquare, FileText, 
    ShieldCheck, ChevronLeft, ChevronRight, 
    User, ClipboardCheck, AlertTriangle, Cpu, Package
} from 'lucide-react';
import { useAuth } from '../context/AuthContext';

const Sidebar = () => {
    const { user } = useAuth();
    const location = useLocation();
    const [isCollapsed, setIsCollapsed] = useState(false);

    if (!user) return null;

    const isActive = (path) => (path === '/' ? location.pathname === '/' : location.pathname.startsWith(path));

    const menuItems = [
        { path: '/', icon: <LayoutDashboard size={18}/>, label: 'Tổng quan', show: true },
        { path: '/assets', icon: <Monitor size={18}/>, label: 'Máy trạm', show: user.permissions?.asset_view },
        { path: '/policy-center', icon: <ShieldCheck size={18}/>, label: 'Chính sách', show: user.permissions?.policy_view },
        { path: '/docs', icon: <FileText size={18}/>, label: 'Tài liệu', show: user.permissions?.doc_view },
        { path: '/users', icon: <User size={18}/>, label: 'Nhân sự', show: user.permissions?.user_manage },
        { path: '/incidents', icon: <AlertTriangle size={18}/>, label: 'Sự cố', show: user.permissions?.incident_view },
        { path: '/behaviors', icon: <AlertTriangle size={18}/>, label: 'Hành vi', show: user.permissions?.incident_view },
        { path: '/approvals', icon: <ClipboardCheck size={18}/>, label: 'Phê duyệt', show: user.permissions?.approval_manage },
        { path: '/versions', icon: <Package size={18}/>, label: 'Phần mềm', show: true }
    ];

    return (
        <div 
            className={`${isCollapsed ? 'w-20' : 'w-64'} bg-[#050B14] h-screen flex flex-col border-r border-slate-800 shadow-[10px_0_30px_rgba(0,0,0,0.5)] relative z-20 transition-all duration-300 ease-in-out font-sans`}
        >
            {/* TOGGLE BUTTON - Đổi sang màu Indigo/Cyber */}
            <button 
                onClick={() => setIsCollapsed(!isCollapsed)}
                className="absolute -right-3 top-10 bg-[#111827] text-indigo-400 p-1.5 rounded-md border border-slate-700 hover:border-indigo-500 transition-all z-50 shadow-xl"
            >
                {isCollapsed ? <ChevronRight size={14}/> : <ChevronLeft size={14}/>}
            </button>

            {/* LOGO SECTION - Typography giống assets */}
            <div className={`p-6 flex items-center ${isCollapsed ? 'justify-center' : 'gap-3'}`}>
                <div className="p-2 bg-indigo-500/10 rounded border border-indigo-500/30 text-indigo-500 shadow-[0_0_15px_rgba(99,102,241,0.2)]">
                    <Cpu size={24} />
                </div>
                {!isCollapsed && (
                    <div className="animate-in fade-in duration-500">
                        <h1 className="text-lg font-black text-white tracking-tighter leading-none">
                            SENT <span className="text-indigo-500">SYSTEM</span>
                        </h1>
                        <p className="text-[9px] text-slate-500 font-black uppercase tracking-[0.2em] mt-1">
                            Security Operations
                        </p>
                    </div>
                )}
            </div>
            
            {/* MENU LIST */}
            <nav className="flex-1 px-3 space-y-1.5 overflow-y-auto mt-4 custom-scrollbar">
                {menuItems.map((item) => (
                    item.show && (
                        <Link 
                            key={item.path} 
                            to={item.path}
                            className={`flex items-center gap-3 px-3 py-2.5 rounded transition-all duration-200 group relative
                                ${isActive(item.path)
                                    ? 'bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 shadow-[0_0_15px_rgba(99,102,241,0.1)]' 
                                    : 'text-slate-500 hover:text-slate-200 hover:bg-slate-800/40'
                                }
                                ${isCollapsed ? 'justify-center' : ''}
                            `}
                        >
                            <span className={`${isActive(item.path) ? 'text-indigo-400' : 'text-slate-500 group-hover:text-indigo-400'}`}>
                                {item.icon}
                            </span>
                            
                            {!isCollapsed && (
                                <span className="text-[11px] font-black uppercase tracking-widest">
                                    {item.label}
                                </span>
                            )}

                            {/* Active Indicator Line */}
                            {isActive(item.path) && (
                                <div className="absolute left-0 w-1 h-4 bg-indigo-500 rounded-r-full shadow-[0_0_8px_rgba(99,102,241,0.8)]"></div>
                            )}

                            {isCollapsed && (
                                <div className="absolute left-16 bg-[#0A101D] text-white text-[10px] font-black uppercase tracking-widest px-3 py-2 rounded border border-slate-700 opacity-0 group-hover:opacity-100 pointer-events-none transition-all z-50 whitespace-nowrap shadow-2xl">
                                    {item.label}
                                </div>
                            )}
                        </Link>
                    )
                ))}
            </nav>

            {/* AI ASSISTANT - Giao diện Button Utility */}
            <div className="p-4 border-t border-slate-800/50 bg-[#0A101D]/50">
                <Link 
                    to="/chat-ai" 
                    className={`flex items-center gap-3 p-2.5 w-full bg-[#111827] hover:bg-indigo-500/10 text-slate-400 hover:text-indigo-400 rounded border border-slate-800 hover:border-indigo-500/30 transition-all group ${isCollapsed ? 'justify-center' : ''}`}
                >
                    <MessageSquare size={18} className="shrink-0"/> 
                    {!isCollapsed && (
                        <span className="text-[10px] font-black uppercase tracking-widest">
                            AI Terminal
                        </span>
                    )}
                </Link>
            </div>
        </div>
    );
};

export default Sidebar;