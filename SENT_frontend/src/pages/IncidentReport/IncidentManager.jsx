import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { ShieldAlert, AlertTriangle, Clock, Monitor, UserCheck, Search, Flame } from 'lucide-react';
import axiosInstance from '../../api/axios';

const IncidentManager = () => {
    const navigate = useNavigate();
    const [incidents, setIncidents] = useState([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');

    const fetchIncidents = async () => {
        try {
            const response = await axiosInstance.get('/incidents');
            const data = Array.isArray(response.data) ? response.data : (response.data.data || []);
            setIncidents(data);
        } catch (error) {
            console.error("Lỗi tải danh sách sự cố:", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchIncidents();
        const interval = setInterval(fetchIncidents, 10000); // Polling mỗi 10s cho giống Real-time
        return () => clearInterval(interval);
    }, []);

    // Nhóm dữ liệu vào 3 cột
    const filteredIncidents = incidents.filter(inc => 
        inc.description?.toLowerCase().includes(searchQuery.toLowerCase()) || 
        inc.Agent?.hostname?.toLowerCase().includes(searchQuery.toLowerCase())
    );

    const openIncidents = filteredIncidents.filter(i => i.status === 'Open');
    const investigatingIncidents = filteredIncidents.filter(i => i.status === 'Investigating');
    const resolvedIncidents = filteredIncidents.filter(i => i.status === 'Resolved');

    if (loading) return <div className="p-10 text-emerald-500 font-bold flex flex-col items-center justify-center h-full animate-pulse"><Flame size={40} className="mb-4"/> Khởi động Triage Board...</div>;

    return (
        <div className="p-6 text-slate-200 h-[calc(100vh-60px)] flex flex-col relative overflow-hidden bg-[#050B14]">
            
            {/* --- HEADER & TOOLBAR --- */}
            <div className="mb-6 flex flex-col md:flex-row justify-between items-start md:items-end gap-4 shrink-0">
                <div>
                    <h1 className="text-2xl font-black text-white flex items-center gap-3">
                        <Flame className="text-red-500 animate-pulse"/> Alert Triage Board
                    </h1>
                    <p className="text-sm text-slate-500 mt-1">Hệ thống phân loại và điều phối sự cố an ninh mạng thời gian thực.</p>
                </div>
                <div className="relative w-full md:w-80 group">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 group-focus-within:text-indigo-400 transition-colors" size={18}/>
                    <input 
                        type="text" 
                        placeholder="Tìm IP, Hostname, Loại sự cố..." 
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-full bg-[#0A101D] text-white pl-10 pr-4 py-2.5 rounded-xl border border-slate-800 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 outline-none transition-all shadow-inner"
                    />
                </div>
            </div>

            {/* --- KANBAN BOARD (3 CỘT) --- */}
            <div className="flex-1 grid grid-cols-1 md:grid-cols-3 gap-6 overflow-hidden pb-4">
                
                {/* CỘT 1: OPEN (BÁO ĐỘNG ĐỎ) */}
                <KanbanColumn 
                    title="Mới phát hiện (OPEN)" 
                    count={openIncidents.length} 
                    color="red"
                    icon={AlertTriangle}
                    incidents={openIncidents}
                    onClickCard={(id) => navigate(`/incidents/${id}`)}
                />

                {/* CỘT 2: INVESTIGATING (ĐANG XỬ LÝ) */}
                <KanbanColumn 
                    title="Đang điều tra (INVESTIGATING)" 
                    count={investigatingIncidents.length} 
                    color="blue"
                    icon={Search}
                    incidents={investigatingIncidents}
                    onClickCard={(id) => navigate(`/incidents/${id}`)}
                />

                {/* CỘT 3: RESOLVED (LƯU TRỮ) */}
                <KanbanColumn 
                    title="Đã đóng (RESOLVED)" 
                    count={resolvedIncidents.length} 
                    color="emerald"
                    icon={ShieldAlert}
                    incidents={resolvedIncidents}
                    onClickCard={(id) => navigate(`/incidents/${id}`)}
                />

            </div>
        </div>
    );
};

// --- COMPONENT CỘT KANBAN ---
const KanbanColumn = ({ title, count, color, icon: Icon, incidents, onClickCard }) => {
    const colorStyles = {
        red: "border-t-red-500 text-red-500 bg-red-500/10",
        blue: "border-t-blue-500 text-blue-500 bg-blue-500/10",
        emerald: "border-t-emerald-500 text-emerald-500 bg-emerald-500/10"
    };

    return (
        <div className="flex flex-col bg-[#0A101D] rounded-2xl border border-slate-800 overflow-hidden">
            {/* Header Cột */}
            <div className={`p-4 border-t-4 border-b border-b-slate-800 flex justify-between items-center ${colorStyles[color]}`}>
                <h2 className="font-black text-sm uppercase tracking-widest flex items-center gap-2"><Icon size={16}/> {title}</h2>
                <span className="bg-black/50 px-2.5 py-1 rounded-lg text-xs font-bold">{count}</span>
            </div>
            
            {/* Danh sách Thẻ Sự cố */}
            <div className="flex-1 overflow-y-auto custom-scrollbar p-3 space-y-3">
                {incidents.length === 0 ? (
                    <div className="h-32 flex items-center justify-center text-slate-600 text-xs font-bold uppercase border-2 border-dashed border-slate-800 rounded-xl">Không có sự cố</div>
                ) : (
                    incidents.map(inc => <IncidentCard key={inc.ID} incident={inc} onClick={() => onClickCard(inc.ID)} color={color} />)
                )}
            </div>
        </div>
    );
};

// --- COMPONENT THẺ SỰ CỐ (CARD) ---
const IncidentCard = ({ incident, onClick, color }) => {
    const agent = incident.Agent || {};
    
    // Tính toán thời gian trôi qua (SLA)
    const hoursOpen = Math.floor((new Date() - new Date(incident.CreatedAt)) / (1000 * 60 * 60));

    return (
        <div 
            onClick={onClick}
            className={`bg-[#111827] p-4 rounded-xl border border-slate-800 cursor-pointer hover:-translate-y-1 hover:shadow-xl transition-all duration-200 group ${color === 'red' ? 'hover:border-red-500/50 hover:shadow-[0_5px_20px_rgba(239,68,68,0.1)]' : color === 'blue' ? 'hover:border-blue-500/50' : 'hover:border-emerald-500/50 opacity-60 hover:opacity-100'}`}
        >
            <div className="flex justify-between items-start mb-3">
                <span className={`px-2 py-0.5 rounded text-[9px] font-black uppercase tracking-wider border ${incident.priority === 'P1' ? 'bg-red-500/20 text-red-400 border-red-500/50' : 'bg-orange-500/20 text-orange-400 border-orange-500/50'}`}>
                    {incident.priority} | {incident.severity}
                </span>
                <span className="text-[10px] text-slate-500 font-mono">#{incident.ID}</span>
            </div>
            
            <h3 className="text-sm font-bold text-slate-200 mb-2 group-hover:text-white transition line-clamp-2">{incident.description || incident.type}</h3>
            
            <div className="flex items-center gap-2 mb-4 text-xs text-slate-400 bg-slate-900/50 p-2 rounded-lg border border-slate-800">
                <Monitor size={12} className="text-indigo-400"/>
                <span className="font-bold truncate">{agent.hostname || 'Unknown Device'}</span>
            </div>

            <div className="flex justify-between items-center pt-3 border-t border-slate-800/80">
                {incident.assignee ? (
                    <div className="flex items-center gap-1.5 text-[10px] text-emerald-400 font-bold">
                        <div className="w-5 h-5 rounded-full bg-emerald-500/20 flex items-center justify-center"><UserCheck size={10}/></div>
                        {incident.assignee.username}
                    </div>
                ) : (
                    <span className="text-[10px] text-slate-500 font-bold italic">Chưa phân công</span>
                )}

                {color !== 'emerald' && (
                    <div className={`flex items-center gap-1 text-[10px] font-mono font-bold ${hoursOpen > 24 ? 'text-red-500 animate-pulse' : 'text-slate-400'}`}>
                        <Clock size={12}/> {hoursOpen}h trôi qua
                    </div>
                )}
            </div>
        </div>
    );
};

export default IncidentManager;