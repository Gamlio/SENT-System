import React, { useState, useEffect, useRef } from 'react';
import { MoreVertical, ShieldCheck, Terminal, Lock, Usb } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

const AgentActions = ({ agent }) => {
    const navigate = useNavigate();
    const [isOpen, setIsOpen] = useState(false);
    const menuRef = useRef(null);

    // Xử lý click ra ngoài để đóng menu
    useEffect(() => {
        const handleClickOutside = (event) => {
            if (menuRef.current && !menuRef.current.contains(event.target)) {
                setIsOpen(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);
    return (
        <div className="relative flex items-center" ref={menuRef}>
            <button 
                onClick={(e) => {
                    e.stopPropagation(); 
                    setIsOpen(!isOpen);
                }}
                className={`p-2 rounded-lg transition ${isOpen ? 'bg-slate-700 text-emerald-400' : 'text-slate-400 hover:text-emerald-400 hover:bg-slate-800'}`}
            >
                <MoreVertical size={20} />
            </button>

            {isOpen && (
                // SỬA VỊ TRÍ TẠI ĐÂY:
                // 1. right-0: Căn thẳng lề phải với nút
                // 2. top-full: Nằm ngay bên dưới nút
                // 3. mt-2: Cách nút một chút cho thoáng
                // 4. z-50: Đảm bảo nổi lên trên tất cả các dòng khác
                <div className="absolute right-0 top-full mt-2 w-56 bg-[#1e293b] border border-slate-700 rounded-xl shadow-2xl z-50 overflow-hidden animate-in fade-in zoom-in duration-200">
                    <div className="px-4 py-2 border-b border-slate-700/50 bg-slate-800/50">
                        <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Tiện ích mở rộng</p>
                    </div>
                    
                    <button
                        onClick={(e) => { 
                            e.stopPropagation(); 
                            setIsOpen(false);
                            navigate(`/agents/${agent.hwid}/software`); 
                        }}
                        className="w-full text-left px-4 py-3 text-sm text-slate-300 hover:bg-slate-700 hover:text-emerald-400 flex items-center gap-3 transition"
                    >
                        <ShieldCheck size={16} /> Cấu hình Phần mềm
                    </button>

                    <button className="w-full text-left px-4 py-3 text-sm text-slate-600 flex items-center gap-3 cursor-not-allowed" title="Sắp ra mắt">
                        <Usb size={16} /> Quản lý USB (Sắp có)
                    </button>
                    <button className="w-full text-left px-4 py-3 text-sm text-slate-600 flex items-center gap-3 cursor-not-allowed" title="Sắp ra mắt">
                        <Terminal size={16} /> Remote Terminal
                    </button>
                    <button className="w-full text-left px-4 py-3 text-sm text-red-500/50 flex items-center gap-3 cursor-not-allowed" title="Sắp ra mắt">
                        <Lock size={16} /> Khóa máy từ xa
                    </button>
                </div>
            )}
        </div>
    );
};

export default AgentActions;