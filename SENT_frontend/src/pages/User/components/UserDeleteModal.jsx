import React, { useState } from 'react';
import { AlertTriangle, Trash2, X, Loader2 } from 'lucide-react';

const UserDeleteModal = ({ isOpen, onClose, onConfirm, user, isLoading }) => {
    const [confirmText, setConfirmText] = useState('');

    if (!isOpen || !user) return null;

    const isMatch = confirmText === user.username;

    const handleConfirm = () => {
        if (isMatch) onConfirm(user.id);
    };

    return (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-[9999] backdrop-blur-sm p-4 animate-in fade-in">
            <div className="bg-[#1e293b] w-full max-w-md rounded-3xl border border-red-900/50 shadow-[0_0_50px_rgba(239,68,68,0.1)] overflow-hidden">
                
                {/* Header DANGER */}
                <div className="p-6 border-b border-red-900/30 bg-red-500/10 flex justify-between items-start">
                    <div className="flex gap-4 items-start">
                        <div className="p-3 bg-red-500/20 text-red-500 rounded-2xl shrink-0">
                            <AlertTriangle size={28}/>
                        </div>
                        <div>
                            <h3 className="text-lg font-black text-red-400">Xóa Tài Khoản</h3>
                            <p className="text-sm text-red-300/70 mt-1">Hành động này không thể hoàn tác!</p>
                        </div>
                    </div>
                    <button onClick={onClose} className="text-slate-500 hover:text-white transition"><X size={20}/></button>
                </div>

                {/* Body */}
                <div className="p-6 space-y-4">
                    <p className="text-sm text-slate-300">
                        Bạn đang chuẩn bị xóa quyền truy cập của nhân viên <span className="font-bold text-white">"{user.full_name}"</span>. 
                        Toàn bộ dữ liệu Lịch trực và Sự cố mà người này đang thụ lý sẽ bị mồ côi hoặc xóa bỏ.
                    </p>

                    <div className="bg-slate-900 p-4 rounded-xl border border-slate-700">
                        <label className="block text-xs font-bold text-slate-400 mb-2">
                            Để xác nhận, vui lòng nhập lại username: <span className="text-emerald-400 font-mono select-all">{user.username}</span>
                        </label>
                        <input 
                            type="text" 
                            value={confirmText}
                            onChange={(e) => setConfirmText(e.target.value)}
                            placeholder="Nhập username tại đây..."
                            className="w-full bg-slate-800 border border-slate-700 rounded-lg p-2.5 text-sm text-white focus:border-red-500 outline-none"
                        />
                    </div>
                </div>

                {/* Footer */}
                <div className="p-4 border-t border-slate-800 bg-slate-800/50 flex justify-end gap-3">
                    <button onClick={onClose} className="px-5 py-2.5 rounded-xl text-sm font-bold text-slate-400 hover:text-white transition">Hủy bỏ</button>
                    <button 
                        onClick={handleConfirm}
                        disabled={!isMatch || isLoading}
                        className="px-6 py-2.5 bg-red-600 hover:bg-red-500 text-white rounded-xl text-sm font-bold shadow-lg shadow-red-500/20 transition flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {isLoading ? <Loader2 size={16} className="animate-spin"/> : <Trash2 size={16}/>}
                        Xác nhận Xóa
                    </button>
                </div>
            </div>
        </div>
    );
};

export default UserDeleteModal;