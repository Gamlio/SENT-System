import React from 'react';
import { PackageSearch, ShieldX, AlertTriangle } from 'lucide-react';

const SoftwarePolicies = () => {
    return (
        <div className="text-slate-200">
            <h1 className="text-3xl font-bold text-white mb-8">Chính sách Phần mềm</h1>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                {/* Panel thêm rule mới */}
                <div className="bg-[#1e293b] p-8 rounded-3xl border border-slate-800">
                    <h3 className="text-xl font-bold mb-6 flex items-center gap-2">
                        <PackageSearch className="text-emerald-400"/> Thêm phần mềm vào "Danh sách đen"
                    </h3>
                    <div className="space-y-4">
                        <input type="text" placeholder="Tên phần mềm (VD: Terraria.exe)" 
                            className="w-full bg-slate-900 border border-slate-700 p-4 rounded-xl text-sm outline-none focus:border-red-500 transition"/>
                        <button className="w-full bg-red-500/10 text-red-500 hover:bg-red-500 hover:text-white p-4 rounded-xl font-bold transition">
                            Cấm phần mềm này trên toàn hệ thống
                        </button>
                    </div>
                </div>

                {/* Danh sách đang cấm */}
                <div className="space-y-4">
                    <div className="bg-red-500/5 border border-red-500/20 p-6 rounded-2xl flex justify-between items-center">
                        <div className="flex items-center gap-4">
                            <ShieldX className="text-red-500" size={32}/>
                            <div>
                                <p className="font-bold">Terraria.exe</p>
                                <p className="text-xs text-slate-500">Tự động báo động khi phát hiện</p>
                            </div>
                        </div>
                        <AlertTriangle className="text-red-500 animate-pulse" size={20}/>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default SoftwarePolicies;