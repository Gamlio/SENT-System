import React from 'react';
import { MapPin, Plus, Copy, Key } from 'lucide-react';

const Regions = () => {
    return (
        <div className="text-slate-200">
            <header className="flex justify-between items-center mb-10">
                <div>
                    <h1 className="text-3xl font-bold text-white">Vùng quản lý</h1>
                    <p className="text-slate-400 text-sm">Thiết lập các chi nhánh và mã định danh Enrollment [cite: 67]</p>
                </div>
                <button className="bg-emerald-500 hover:bg-emerald-600 text-white px-6 py-3 rounded-2xl font-bold flex items-center gap-2 transition shadow-lg shadow-emerald-500/20">
                    <Plus size={20}/> Thêm chi nhánh mới
                </button>
            </header>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Mỗi chi nhánh có một Token riêng để Agent tự gán vùng  */}
                <div className="bg-[#1e293b] p-8 rounded-3xl border border-slate-800">
                    <div className="flex justify-between items-center mb-6">
                        <div className="flex items-center gap-3">
                            <div className="p-3 bg-blue-500/10 rounded-xl text-blue-400"><MapPin size={24}/></div>
                            <h3 className="text-xl font-bold text-white">Chi nhánh Bắc Giang</h3>
                        </div>
                        <span className="text-xs text-slate-500 font-bold">124 Máy trạm [cite: 74]</span>
                    </div>

                    <div className="bg-slate-900/50 p-4 rounded-2xl border border-slate-800">
                        <div className="flex justify-between items-center mb-2">
                            <span className="text-[10px] text-slate-500 font-bold uppercase tracking-widest flex items-center gap-1">
                                <Key size={10}/> Enrollment Token [cite: 67]
                            </span>
                            <button className="text-emerald-400 hover:text-emerald-300 transition"><Copy size={14}/></button>
                        </div>
                        <code className="text-emerald-400 font-mono text-sm break-all">SENT-BG-9981-XXXX-2026</code>
                    </div>

                    <p className="text-[10px] text-slate-500 mt-4 italic font-medium">* Sử dụng mã này khi cài đặt Agent để tự động phân vùng máy trạm[cite: 67].</p>
                </div>
            </div>
        </div>
    );
};

export default Regions;