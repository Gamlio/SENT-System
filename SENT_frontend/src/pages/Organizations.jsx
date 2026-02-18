import React, { useState } from 'react';
import { Globe, ShieldCheck, ShieldAlert, Plus } from 'lucide-react';

const Organizations = () => {
    return (
        <div className="text-slate-200">
            <header className="flex justify-between items-center mb-8">
                <div>
                    <h1 className="text-3xl font-bold text-white font-sans">Quản lý Đối tác SME</h1>
                    <p className="text-slate-400 text-sm">Quản lý License và hạ tầng cho 20+ công ty</p>
                </div>
                <button className="bg-emerald-500 hover:bg-emerald-600 text-white px-6 py-3 rounded-2xl font-bold flex items-center gap-2 shadow-lg shadow-emerald-500/20 transition">
                    <Plus size={20}/> Đăng ký SME mới
                </button>
            </header>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {/* Thẻ công ty SME */}
                <div className="p-6 bg-[#1e293b] rounded-3xl border border-slate-800 hover:border-emerald-500/30 transition shadow-xl">
                    <div className="flex justify-between items-start mb-4">
                        <div className="p-3 bg-emerald-500/10 rounded-xl text-emerald-400"><Globe size={24}/></div>
                        <span className="bg-emerald-500/10 text-emerald-400 text-[10px] font-bold px-3 py-1 rounded-full uppercase">PRO PLAN</span>
                    </div>
                    <h3 className="text-xl font-bold text-white mb-1">Foxconn BN-01</h3>
                    <p className="text-xs text-slate-500 mb-4">Mã định danh: SENT-FOX-2026</p>
                    <div className="flex justify-between items-center pt-4 border-t border-slate-800">
                        <p className="text-xs font-bold text-slate-400">1,240 Agents</p>
                        <button className="text-emerald-400 text-xs font-bold hover:underline">Cấu hình Rules</button>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Organizations;