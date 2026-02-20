import React, { useState, useEffect } from 'react';
import { ShieldAlert, CheckCircle, Clock, Filter } from 'lucide-react';
import axios from '../../api/axios'; // Sử dụng instance axios đã cấu hình

const Alerts = () => {
    const [alerts, setAlerts] = useState([]);

    useEffect(() => {
        // fetchAlerts(); // Gọi API GET /api/v1/alerts
    }, []);

    return (
        <div className="text-slate-200">
            <header className="flex justify-between items-center mb-8">
                <div>
                    <h1 className="text-3xl font-bold text-white">Cảnh báo an ninh</h1>
                    <p className="text-slate-400 text-sm">Quản lý và xử lý các hành vi vi phạm chính sách</p>
                </div>
                <button className="flex items-center gap-2 bg-slate-800 px-4 py-2 rounded-xl text-sm font-bold border border-slate-700 hover:border-emerald-500 transition">
                    <Filter size={16}/> Bộ lọc
                </button>
            </header>

            <div className="space-y-4">
                {/* Mỗi Alert là một Ticket để truy cứu */}
                <div className="bg-[#1e293b] p-6 rounded-2xl border-l-4 border-red-500 shadow-lg">
                    <div className="flex justify-between items-start">
                        <div className="flex gap-4">
                            <div className="p-3 bg-red-500/10 rounded-xl text-red-500">
                                <ShieldAlert size={24} />
                            </div>
                            <div>
                                <h3 className="text-lg font-bold text-white">Phát hiện USB lạ (Unauthorized USB)</h3>
                                <p className="text-sm text-slate-400 mt-1">Thiết bị: <span className="text-red-400 font-mono">USB\VID_0951&PID_1666</span></p>
                                <p className="text-sm text-slate-400">Máy trạm: <span className="text-white font-bold">ICTU-LAB-01</span></p>
                            </div>
                        </div>
                        <div className="text-right">
                            <span className="px-3 py-1 bg-red-500 text-white text-[10px] font-black rounded-full uppercase">Critical</span>
                            <p className="text-[10px] text-slate-500 mt-2 flex items-center justify-end gap-1"><Clock size={10}/> 10 phút trước</p>
                        </div>
                    </div>
                    <div className="mt-4 pt-4 border-t border-slate-800 flex justify-between items-center">
                        <p className="text-xs text-slate-400 italic font-medium">Bằng chứng kỹ thuật đã được lưu vào nhật ký hệ thống</p>
                        <button className="bg-emerald-500/10 text-emerald-400 px-4 py-2 rounded-lg text-xs font-bold hover:bg-emerald-500 hover:text-white transition flex items-center gap-2">
                            <CheckCircle size={14}/> Đánh dấu đã xử lý
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Alerts;