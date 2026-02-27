import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { ShieldAlert, Search, Filter, AlertTriangle, CheckCircle } from 'lucide-react';
import axiosInstance from '../../api/axios';

const IncidentList = () => {
    const [incidents, setIncidents] = useState([]);
    const [loading, setLoading] = useState(true);
    const navigate = useNavigate();

    useEffect(() => {
        const fetchIncidents = async () => {
            try {
                // Gọi API lấy danh sách tổng hợp
                const response = await axiosInstance.get('/api/v1/incidents');
                setIncidents(response.data.data || []);
            } catch (error) {
                console.error("Lỗi tải danh sách sự cố:", error);
            } finally {
                setLoading(false);
            }
        };
        fetchIncidents();
        
        // Tự động làm mới mỗi 30 giây
        const interval = setInterval(fetchIncidents, 30000);
        return () => clearInterval(interval);
    }, []);

    if (loading) return <div className="p-6 text-slate-400 font-bold animate-pulse">Đang tải dữ liệu sự cố...</div>;

    return (
        <div className="p-6 text-slate-200 h-full flex flex-col">
            {/* Header */}
            <div className="flex justify-between items-center mb-6">
                <div>
                    <h1 className="text-3xl font-black text-white flex items-center gap-3">
                        <ShieldAlert className="text-red-500" size={32} />
                        Điều tra Sự cố (Incident Response)
                    </h1>
                    <p className="text-slate-400 text-sm mt-1">Quản lý và xử lý các mối đe dọa an ninh từ máy trạm</p>
                </div>
                <div className="flex gap-3">
                    <button className="flex items-center gap-2 bg-slate-800 px-4 py-2 rounded-xl text-sm font-bold border border-slate-700 hover:bg-slate-700 transition">
                        <Filter size={16}/> Lọc trạng thái
                    </button>
                    <div className="relative">
                        <Search className="absolute left-3 top-2.5 text-slate-500" size={16} />
                        <input 
                            type="text" 
                            placeholder="Tìm kiếm HWID, Loại đe dọa..." 
                            className="bg-slate-900 border border-slate-700 rounded-xl pl-9 pr-4 py-2 text-sm text-white focus:outline-none focus:border-blue-500"
                        />
                    </div>
                </div>
            </div>

            {/* Bảng dữ liệu */}
            <div className="bg-slate-800 rounded-2xl border border-slate-700 shadow-xl overflow-hidden flex-1">
                <div className="overflow-x-auto">
                    <table className="w-full text-left border-collapse">
                        <thead>
                            <tr className="bg-slate-900/50 text-slate-400 text-xs uppercase tracking-wider border-b border-slate-700">
                                <th className="p-4 font-bold">Mã Case</th>
                                <th className="p-4 font-bold">Máy Trạm (HWID)</th>
                                <th className="p-4 font-bold">Loại Đe Dọa</th>
                                <th className="p-4 font-bold text-center">Mức Độ</th>
                                <th className="p-4 font-bold text-center">Trạng Thái</th>
                                <th className="p-4 font-bold">Thời Gian</th>
                                <th className="p-4 font-bold text-center">Thao tác</th>
                            </tr>
                        </thead>
                        <tbody className="text-sm divide-y divide-slate-700/50">
                            {incidents.length === 0 ? (
                                <tr>
                                    <td colSpan="7" className="p-8 text-center text-slate-500 font-medium">
                                        <CheckCircle className="mx-auto mb-2 text-emerald-500 opacity-50" size={32} />
                                        Hệ thống an toàn, không có sự cố nào đang mở.
                                    </td>
                                </tr>
                            ) : (
                                incidents.map((incident) => (
                                    <tr key={incident.id} className="hover:bg-slate-700/30 transition group">
                                        <td className="p-4 font-mono text-slate-400">#{incident.id}</td>
                                        <td className="p-4">
                                            <p className="font-bold text-blue-400">{incident.agent?.hostname || 'Unknown'}</p>
                                            <p className="text-xs font-mono text-slate-500 bg-slate-900 px-1 py-0.5 rounded w-fit mt-1">
                                                {incident.agent_hwid}
                                            </p>
                                        </td>
                                        <td className="p-4 font-semibold text-slate-300">{incident.type}</td>
                                        <td className="p-4 text-center">
                                            <span className={`px-2 py-1 rounded text-[10px] font-black uppercase tracking-wider ${
                                                incident.severity === 'Critical' ? 'bg-red-500/20 text-red-500 border border-red-500/50' :
                                                incident.severity === 'High' ? 'bg-orange-500/20 text-orange-400 border border-orange-500/50' :
                                                'bg-yellow-500/20 text-yellow-400 border border-yellow-500/50'
                                            }`}>
                                                {incident.severity}
                                            </span>
                                        </td>
                                        <td className="p-4 text-center">
                                            {incident.status === 'Resolved' ? (
                                                <span className="text-emerald-400 font-bold flex items-center justify-center gap-1 text-xs">
                                                    <CheckCircle size={14} /> Đã xử lý
                                                </span>
                                            ) : (
                                                <span className="text-red-400 font-bold flex items-center justify-center gap-1 text-xs animate-pulse">
                                                    <AlertTriangle size={14} /> Đang mở
                                                </span>
                                            )}
                                        </td>
                                        <td className="p-4 text-slate-400 font-mono text-xs">
                                            {new Date(incident.created_at).toLocaleString('vi-VN')}
                                        </td>
                                        <td className="p-4 text-center">
                                            <button 
                                                onClick={() => navigate(`/incidents/${incident.id}`)}
                                                className="bg-blue-600/20 text-blue-400 border border-blue-500/50 hover:bg-blue-500 hover:text-white px-3 py-1.5 rounded-lg text-xs font-bold transition shadow-lg"
                                            >
                                                Điều tra ngay
                                            </button>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
};

export default IncidentList;