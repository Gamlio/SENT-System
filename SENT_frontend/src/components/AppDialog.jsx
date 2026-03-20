import React from 'react';
import { AlertTriangle, CheckCircle, Info, X } from 'lucide-react';

const AppDialog = ({ 
    isOpen, onClose, onConfirm, title, message, type = 'danger',
    confirmText = 'Xác nhận', cancelText = 'Hủy bỏ', isAlertOnly = false,
    // --- [MỚI] THÊM PROPS CHO Ô NHẬP LIỆU ---
    showInput = false,
    inputValue = '',
    onInputChange = () => {},
    inputPlaceholder = 'Nhập lý do thực hiện thao tác này...'
}) => {
    if (!isOpen) return null;

    const config = {
        danger: { icon: <AlertTriangle size={24} className="text-red-500" />, bg: 'bg-red-500/10', border: 'border-red-500/20', btn: 'bg-red-500 hover:bg-red-600 focus:ring-red-500' },
        warning: { icon: <AlertTriangle size={24} className="text-amber-500" />, bg: 'bg-amber-500/10', border: 'border-amber-500/20', btn: 'bg-amber-500 hover:bg-amber-600 focus:ring-amber-500' },
        success: { icon: <CheckCircle size={24} className="text-emerald-500" />, bg: 'bg-emerald-500/10', border: 'border-emerald-500/20', btn: 'bg-emerald-500 hover:bg-emerald-600 focus:ring-emerald-500' },
        info: { icon: <Info size={24} className="text-blue-500" />, bg: 'bg-blue-500/10', border: 'border-blue-500/20', btn: 'bg-blue-500 hover:bg-blue-600 focus:ring-blue-500' }
    };
    const current = config[type];

    return (
        <div className="fixed inset-0 z-[99999] flex items-center justify-center p-4">
            <div className="absolute inset-0 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200" onClick={onClose}></div>

            <div className="relative bg-[#1e293b] w-full max-w-sm rounded-3xl border border-slate-700 shadow-2xl p-6 animate-in zoom-in-95 duration-200">
                <button onClick={onClose} className="absolute top-4 right-4 text-slate-500 hover:text-white transition"><X size={20} /></button>

                <div className="flex flex-col items-center text-center mt-2">
                    <div className={`p-4 rounded-full ${current.bg} ${current.border} border mb-4 shadow-inner`}>
                        {current.icon}
                    </div>
                    <h3 className="text-lg font-bold text-white mb-2">{title}</h3>
                    <p className="text-sm text-slate-400 mb-2">{message}</p>
                    
                    {/* --- [MỚI] Ô NHẬP LÝ DO XÓA --- */}
                    {showInput && (
                        <textarea
                            value={inputValue}
                            onChange={(e) => onInputChange(e.target.value)}
                            placeholder={inputPlaceholder}
                            className="w-full mt-3 p-3 bg-slate-900 border border-slate-700 rounded-xl text-white text-sm outline-none focus:border-red-500 transition-colors custom-scrollbar"
                            rows={3}
                        />
                    )}
                </div>

                <div className="flex gap-3 w-full mt-6">
                    {!isAlertOnly && (
                        <button onClick={onClose} className="flex-1 py-2.5 rounded-xl text-sm font-bold text-slate-300 bg-slate-800 hover:bg-slate-700 border border-slate-700 transition">{cancelText}</button>
                    )}
                    <button 
                        // Chặn bấm Xác nhận nếu có input mà chưa nhập lý do
                        disabled={showInput && inputValue.trim().length < 5}
                        onClick={() => { if (onConfirm) onConfirm(); if (isAlertOnly) onClose(); }}
                        className={`flex-1 py-2.5 rounded-xl text-sm font-bold text-white transition focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-[#1e293b] disabled:opacity-50 disabled:cursor-not-allowed ${current.btn}`}
                    >
                        {isAlertOnly ? 'Đã hiểu' : confirmText}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default AppDialog;