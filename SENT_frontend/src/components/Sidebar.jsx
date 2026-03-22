import React, { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { 
    LayoutDashboard, Monitor, MessageSquare, FileText, 
    ShieldCheck, ChevronLeft, ChevronRight, 
    User,ClipboardCheck,AlertTriangle
} from 'lucide-react';
import { useAuth } from '../context/AuthContext';

const Sidebar = () => {
    const { user } = useAuth();
    const location = useLocation();
    
    // State quản lý đóng/mở (Mặc định là mở - false)
    const [isCollapsed, setIsCollapsed] = useState(false);

    if (!user) return null;

    const isActive = (path) => location.pathname.startsWith(path);

    const menuItems = [
    { path: '/', icon: <LayoutDashboard size={20}/>, label: 'Tổng quan', show: true },
    { path: '/agents', icon: <Monitor size={20}/>, label: 'Máy trạm', show: user.permissions?.agent_view },
    { path: '/policy-center', icon: <ShieldCheck size={20}/>, label: 'Chính sách', show: user.permissions?.policy_view },
    { path: '/docs', icon: <FileText size={20}/>, label: 'Tài liệu', show: user.permissions?.doc_view },
    { path: '/users', icon: <User size={20}/>, label: 'Nhân sự', show: user.permissions?.user_manage },
    { path: '/incidents', icon: <AlertTriangle size={20}/>, label: 'Sự cố', show: user.permissions?.incident_view },
    { path: '/approvals', icon: <ClipboardCheck size={20}/>, label: 'Phê duyệt', show: user.permissions?.approval_manage }
];
console.log("DỮ LIỆU USER ĐANG CÓ:", user);
    return (
        <div 
            className={`${isCollapsed ? 'w-25' : 'w-60'} bg-[#1e293b] h-screen flex flex-col border-r border-slate-800 shadow-2xl relative z-20 transition-all duration-300 ease-in-out`}
        >
            {/* --- NÚT TOGGLE ĐÓNG/MỞ --- */}
            <button 
                onClick={() => setIsCollapsed(!isCollapsed)}
                className="absolute -right-0.5 top-9 bg-emerald-500 text-white p-1 rounded-full shadow-lg border-2 border-[#1e293b] hover:bg-emerald-600 transition-colors z-50"
            >
                {isCollapsed ? <ChevronRight size={14}/> : <ChevronLeft size={14}/>}
            </button>

            {/* --- LOGO SECTION --- */}
            <div className={`p-6 flex items-center ${isCollapsed ? 'justify-center' : 'gap-3'} transition-all`}>
                <div className="p-2 bg-emerald-500/10 rounded-xl text-emerald-400 shrink-0">
                    <ShieldCheck size={28} />
                </div>
                
                {/* Chỉ hiện chữ khi Sidebar mở */}
                <div className={`overflow-hidden transition-all duration-300 ${isCollapsed ? 'w-0 opacity-0' : 'w-auto opacity-100'}`}>
                    <h1 className="text-xl font-black text-white tracking-wider whitespace-nowrap">
                        SENT <span className="text-emerald-400">SOC</span>
                    </h1>
                    <p className="text-[10px] text-slate-500 font-bold uppercase tracking-widest -mt-1 whitespace-nowrap">
                        Global System
                    </p>
                </div>
            </div>
            
            {/* --- MENU LIST --- */}
            <nav className="flex-1 px-3 space-y-2 overflow-y-auto mt-4 overflow-x-hidden">
                {menuItems.map((item) => (
                    item.show && (
                        <Link 
                            key={item.path} 
                            to={item.path}
                            title={isCollapsed ? item.label : ""} // Hiện tooltip khi hover nếu đang đóng
                            className={`flex items-center gap-3 px-3 py-3 rounded-xl text-sm font-bold transition-all duration-200 group
                                ${
                                    (item.path === '/' ? location.pathname === '/' : isActive(item.path))
                                    ? 'bg-emerald-500 text-white shadow-lg shadow-emerald-500/20' 
                                    : 'text-slate-400 hover:bg-slate-800/50 hover:text-white'
                                }
                                ${isCollapsed ? 'justify-center' : ''}
                            `}
                        >
                            {/* Icon giữ nguyên kích thước */}
                            <span className="shrink-0">{item.icon}</span>
                            
                            {/* Label trượt ẩn đi */}
                            <span className={`whitespace-nowrap overflow-hidden transition-all duration-300 ${isCollapsed ? 'w-0 opacity-0' : 'w-auto opacity-100'}`}>
                                {item.label}
                            </span>

                            {/* Tooltip giả lập khi hover lúc đóng sidebar */}
                            {isCollapsed && (
                                <div className="absolute left-16 bg-slate-900 text-white text-xs px-2 py-1 rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap border border-slate-700">
                                    {item.label}
                                </div>
                            )}
                        </Link>
                    )
                ))}
            </nav>

            {/* --- AI CHAT BUTTON --- */}
               <div className="p-4 border-t border-slate-800 relative">
                <Link 
                    to="/chat-ai" 
                    className={`flex items-center gap-2 p-3 w-full bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-xl border border-slate-700 transition font-bold text-sm shadow-xl group ${isCollapsed ? 'justify-center' : ''}`}
                    title={isCollapsed ? "Hỏi trợ lý AI" : ""}
                >
                    <MessageSquare size={18} className="text-blue-400 shrink-0"/> 
                    <span className={`whitespace-nowrap overflow-hidden transition-all duration-300 ${isCollapsed ? 'w-0 opacity-0' : 'w-auto opacity-100'}`}>
                            Hỏi trợ lý AI
                    </span>
                    
                    {/* Tooltip when collapsed */}
                    {isCollapsed && (
                        <div className="absolute left-16 bg-slate-900 text-white text-xs px-2 py-1 rounded opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity z-50 whitespace-nowrap border border-slate-700">
                            Hỏi trợ lý AI
                        </div>
                    )}
                </Link>
            </div>
        </div>
    );
};

export default Sidebar;