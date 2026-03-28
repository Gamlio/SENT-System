import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { ShieldAlert, AlertTriangle, Clock, Monitor, Search, Flame, ShieldCheck, Activity, UserCheck } from 'lucide-react';
import axiosInstance from '../../api/axios';
import { useSocketSubscription } from '../../context/useSocketSubscription';

const IncidentManager = () => {
    const navigate = useNavigate();
    const [incidents, setIncidents] = useState([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');

    const fetchIncidents = useCallback(async () => {
        try {
            const response = await axiosInstance.get('/incidents');
            const data = Array.isArray(response.data) ? response.data : (response.data.data || []);
            setIncidents(data);
        } catch (error) {
            console.error("Lỗi tải danh sách sự cố:", error);
        } finally {
            setLoading(false);
        }
    }, []);

    useSocketSubscription(['NEW_INCIDENT', 'INCIDENT_SEVERITY_ESCALATED'], () => {
        fetchIncidents();
    });

    useEffect(() => { fetchIncidents(); }, [fetchIncidents]);

    const filteredIncidents = incidents.filter(inc => 
        inc.description?.toLowerCase().includes(searchQuery.toLowerCase()) || 
        inc.Agent?.hostname?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        inc.type?.toLowerCase().includes(searchQuery.toLowerCase())
    );

    if (loading) return <div className="p-10 text-emerald-500 font-bold flex flex-col items-center justify-center h-full animate-pulse"><Flame size={40} className="mb-4"/> Đang tải dữ liệu SOC...</div>;

    return (
        <div className="p-6 text-slate-200 h-[calc(100vh-60px)] flex flex-col bg-[#050B14] font-sans">
            
            {/* TOOLBAR */}
            <div className="mb-4 flex justify-between items-end shrink-0">
                <div>
                    <h1 className="text-xl font-black text-white flex items-center gap-2 tracking-tight">
                        <Activity className="text-indigo-500"/> THỐNG KÊ SỰ CỐ AN NINH
                    </h1>
                    <p className="text-[11px] text-slate-500 mt-1 uppercase tracking-widest font-bold">Danh sách liệt kê toàn bộ các cảnh báo từ Agent</p>
                </div>
                <div className="relative w-80">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" size={14}/>
                    <input 
                        type="text" 
                        placeholder="Tìm theo ID, IP, Hostname, Mã lỗi..." 
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-full bg-[#0A101D] text-xs text-white pl-9 pr-4 py-2 rounded border border-slate-800 focus:border-indigo-500 outline-none transition-all shadow-inner font-mono"
                    />
                </div>
            </div>

            {/* DATA GRID TRÀN VIỀN */}
            <div className="flex-1 bg-[#0A101D] border border-slate-800 rounded-lg overflow-hidden flex flex-col shadow-2xl">
                <div className="overflow-x-auto flex-1 custom-scrollbar">
                    <table className="w-full text-left border-collapse whitespace-nowrap">
                        <thead className="bg-[#111827] sticky top-0 z-10">
                            <tr className="text-[10px] uppercase tracking-widest text-slate-500 border-b border-slate-800">
                                <th className="p-3 font-black">ID & Mức Độ</th>
                                <th className="p-3 font-black">Mã Lỗi & Chi Tiết</th>
                                <th className="p-3 font-black">Thiết Bị Nạn Nhân</th>
                                <th className="p-3 font-black">Trạng Thái</th>
                                <th className="p-3 font-black">Điều Tra Viên</th>
                                <th className="p-3 font-black text-right">Thời Gian (SLA)</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-800/50">
                            {filteredIncidents.length === 0 ? (
                                <tr><td colSpan="6" className="p-10 text-center text-slate-500 text-xs font-bold uppercase">Không có dữ liệu</td></tr>
                            ) : filteredIncidents.map(inc => {
                                const hoursOpen = Math.floor((new Date() - new Date(inc.CreatedAt || inc.created_at)) / (1000 * 60 * 60));
                                return (
                                    <tr 
                                        key={inc.ID} 
                                        onClick={() => navigate(`/incidents/${inc.ID}`)}
                                        className="hover:bg-slate-800/30 transition cursor-pointer group"
                                    >
                                        <td className="p-3">
                                            <div className="flex items-center gap-2">
                                                <span className={`w-1.5 h-6 rounded-full ${inc.priority === 'P1' ? 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.8)]' : inc.priority === 'P2' ? 'bg-orange-500' : 'bg-yellow-500'}`}></span>
                                                <div>
                                                    <p className="text-[11px] font-mono text-slate-400">#{inc.ID}</p>
                                                    <p className={`text-xs font-black ${inc.priority === 'P1' ? 'text-red-400' : 'text-orange-400'}`}>{inc.priority}</p>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="p-3 min-w-[250px] max-w-[400px]">
                                            <p className="text-xs font-bold text-white mb-0.5 truncate flex items-center gap-1.5">
                                                <ShieldAlert size={12} className={inc.priority === 'P1' ? 'text-red-500' : 'text-slate-500'}/> {inc.type}
                                            </p>
                                            <p className="text-[11px] text-slate-500 truncate">{inc.description}</p>
                                        </td>
                                        <td className="p-3">
                                            <div className="flex items-center gap-2">
                                                <Monitor size={14} className="text-slate-600 group-hover:text-indigo-400 transition"/>
                                                <div>
                                                    <p className="text-xs font-bold text-slate-300 group-hover:text-white transition">{inc.Agent?.hostname || 'Unknown'}</p>
                                                    <p className="text-[10px] text-slate-500 font-mono">{inc.Agent?.ip_address || 'N/A'}</p>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="p-3">
                                            <span className={`px-2 py-1 rounded text-[10px] font-black uppercase tracking-widest border ${
                                                inc.status === 'Open' ? 'text-red-400 border-red-500/30 bg-red-500/10' : 
                                                inc.status === 'Investigating' ? 'text-blue-400 border-blue-500/30 bg-blue-500/10' : 
                                                'text-emerald-400 border-emerald-500/30 bg-emerald-500/10'
                                            }`}>
                                                {inc.status}
                                            </span>
                                        </td>
                                        <td className="p-3">
                                            {inc.assignee ? (
                                                <div className="flex items-center gap-1.5 text-[11px] text-emerald-400 font-bold bg-emerald-500/10 border border-emerald-500/20 px-2 py-1 rounded w-max">
                                                    <UserCheck size={12}/> {inc.assignee.username}
                                                </div>
                                            ) : (
                                                <span className="text-[10px] text-slate-600 font-bold italic border border-slate-700 border-dashed px-2 py-1 rounded w-max block">Chưa phân công</span>
                                            )}
                                        </td>
                                        <td className="p-3 text-right">
                                            <p className="text-[11px] font-mono text-slate-400">{new Date(inc.CreatedAt || inc.created_at).toLocaleString('vi-VN')}</p>
                                            {inc.status !== 'Resolved' && (
                                                <p className={`text-[10px] font-bold mt-0.5 flex items-center justify-end gap-1 ${hoursOpen > 24 ? 'text-red-500 animate-pulse' : 'text-slate-500'}`}>
                                                    <Clock size={10}/> Tồn đọng {hoursOpen}h
                                                </p>
                                            )}
                                        </td>
                                    </tr>
                                );
                            })}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
};

export default IncidentManager;