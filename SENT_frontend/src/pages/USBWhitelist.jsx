import React, { useState } from 'react';
import { Usb, Plus, Trash2, UserCheck } from 'lucide-react';

const USBWhitelist = () => {
    return (
        <div className="text-slate-200">
            <header className="flex justify-between items-center mb-8">
                <div>
                    <h1 className="text-3xl font-bold text-white">Danh sách trắng USB</h1>
                    <p className="text-slate-400 text-sm">Chỉ các thiết bị trong danh sách này mới được coi là hợp lệ</p>
                </div>
                <button className="bg-emerald-500 hover:bg-emerald-600 text-white px-6 py-3 rounded-2xl font-bold flex items-center gap-2 shadow-lg shadow-emerald-500/20 transition">
                    <Plus size={20}/> Cấp phép USB mới
                </button>
            </header>

            <div className="bg-[#1e293b] rounded-3xl border border-slate-800 overflow-hidden">
                <table className="w-full text-left">
                    <thead className="bg-slate-900/30 text-[10px] uppercase tracking-widest text-slate-500">
                        <tr>
                            <th className="p-5">Thiết bị / ID</th>
                            <th className="p-5">Người sở hữu</th>
                            <th className="p-5">Ngày cấp</th>
                            <th className="p-5 text-right">Thao tác</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800">
                        <tr className="hover:bg-slate-800/50 transition">
                            <td className="p-5">
                                <div className="flex items-center gap-3">
                                    <Usb size={18} className="text-emerald-400"/>
                                    <div>
                                        <p className="font-bold text-sm">Kingston IT Office 01</p>
                                        <p className="text-[10px] text-slate-500 font-mono">USB\VID_0951&PID_1666</p>
                                    </div>
                                </div>
                            </td>
                            <td className="p-5">
                                <div className="flex items-center gap-2 text-sm">
                                    <UserCheck size={14} className="text-blue-400"/> Nguyen Van A
                                </div>
                            </td>
                            <td className="p-5 text-xs text-slate-400">09/02/2026</td>
                            <td className="p-5 text-right">
                                <button className="text-slate-500 hover:text-red-400 transition"><Trash2 size={18}/></button>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default USBWhitelist;