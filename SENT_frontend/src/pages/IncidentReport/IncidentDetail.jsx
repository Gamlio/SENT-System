import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import axiosInstance from '../../api/axios';
import { ShieldAlert, Monitor, Network, Clock, ArrowLeft, AlertTriangle, Fingerprint, Activity, Terminal } from 'lucide-react';
import IncidentTimeline from './components/DetailParts/IncidentTimeline';
import IncidentActionBox from './components/DetailParts/IncidentActionBox';
import IncidentPlaybook from './components/DetailParts/IncidentPlaybook';

const IncidentDetail = () => {
    const { id } = useParams();
    const navigate = useNavigate();
    
    const [incident, setIncident] = useState(null);
    const [loading, setLoading] = useState(true);
    const chatScrollRef = useRef(null);

    const fetchDetail = async () => {
        try {
            const res = await axiosInstance.get(`/incidents/${id}`);
            setIncident(res.data);
        } catch (err) { console.error(err); } finally { setLoading(false); }
    };

    useEffect(() => { fetchDetail(); }, [id]);

    useEffect(() => {
        if (chatScrollRef.current) {
            setTimeout(() => { chatScrollRef.current.scrollTop = chatScrollRef.current.scrollHeight; }, 100);
        }
    }, [incident]);

    const handleAction = async (type, text, files, summary) => {
        try {
            const formData = new FormData();
            formData.append('action_type', type);
            formData.append('content', text || summary || '');

            if (files && files.length > 0) {
                files.forEach(file => formData.append('images', file));
            }

            await axiosInstance.post(`/incidents/${id}/activity`, formData, {
                headers: { 'Content-Type': 'multipart/form-data' }
            });
            fetchDetail(); 
        } catch (error) {
            console.error(error);
            alert("Lỗi cập nhật Audit Log!");
        }
    };

    if (loading || !incident) return <div className="h-screen bg-[#050B14] flex items-center justify-center text-emerald-500 font-mono text-sm animate-pulse">ESTABLISHING SECURE CONNECTION...</div>;

    const asset = incident.asset || incident.asset || {};
    const alerts = incident.alerts || incident.Alerts || [];

    return (
        <div className="h-[calc(100vh-60px)] bg-[#050B14] text-slate-200 flex flex-col overflow-hidden font-sans">
            
            {/* HEADER 100% WIDE */}
            <div className="h-14 border-b border-slate-800 bg-[#0A101D] px-6 flex justify-between items-center shrink-0 z-10 shadow-md">
                <div className="flex items-center gap-4">
                    <button onClick={() => navigate('/incidents')} className="text-slate-500 hover:text-white transition"><ArrowLeft size={16}/></button>
                    <div className="h-6 w-px bg-slate-800"></div>
                    <span className={`px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-widest border flex items-center gap-1.5 ${incident.priority === 'P1' ? 'bg-red-500/10 text-red-500 border-red-500/50 shadow-[0_0_10px_rgba(239,68,68,0.3)] animate-pulse' : 'bg-orange-500/10 text-orange-500 border-orange-500/50'}`}>
                        {incident.priority === 'P1' ? <AlertTriangle size={12}/> : <ShieldAlert size={12}/>}
                        {incident.priority}
                    </span>
                    <h1 className="text-sm font-black text-white tracking-widest uppercase">CASE #{incident.ID} <span className="text-slate-600 mx-2">|</span> {incident.type}</h1>
                </div>
                
                <div className="flex items-center gap-3">
                    <div className={`px-3 py-1 rounded text-[10px] font-black uppercase tracking-widest border ${incident.status === 'Open' ? 'text-red-400 border-red-500/30 bg-red-500/10' : incident.status === 'Investigating' ? 'text-blue-400 border-blue-500/30 bg-blue-500/10' : 'text-emerald-400 border-emerald-500/30 bg-emerald-500/10'}`}>
                        Trạng thái: {incident.status}
                    </div>
                </div>
            </div>

            {/* BODY CỦA CHI TIẾT SỰ CỐ: 50/50 SPLIT */}
            <div className="flex-1 flex overflow-hidden">
                
                {/* NỬA TRÁI (50%): THÔNG TIN SỰ CỐ CHI TIẾT */}
                <div className="w-1/2 h-full overflow-y-auto custom-scrollbar border-r border-slate-800 bg-[#0A101D]/50 p-6 space-y-6">
                    
                    {/* KHỐI 1: THIẾT BỊ NẠN NHÂN */}
                    <div>
                        <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-2 flex items-center gap-2 border-b border-slate-800 pb-2"><Monitor size={12}/> Asset Fingerprint</h3>
                        <div className="grid grid-cols-2 gap-3 bg-slate-900/50 p-4 rounded-lg border border-slate-800/80">
                            <div>
                                <p className="text-[9px] text-slate-500 uppercase font-bold mb-0.5">Hostname</p>
                                <p className="text-xs font-bold text-white flex items-center gap-2">
                                    {asset.hostname || 'Unknown Device'}
                                    <span className={`w-1.5 h-1.5 rounded-full ${asset.status === 'online' ? 'bg-emerald-500' : 'bg-slate-600'}`}></span>
                                </p>
                            </div>
                            <div>
                                <p className="text-[9px] text-slate-500 uppercase font-bold mb-0.5">IP Address</p>
                                <p className="text-[11px] text-emerald-400 font-mono flex items-center gap-1"><Network size={10}/> {asset.ip_address || 'N/A'}</p>
                            </div>
                            <div className="col-span-2">
                                <p className="text-[9px] text-slate-500 uppercase font-bold mb-0.5">HWID / asset ID</p>
                                <p className="text-[10px] text-slate-400 font-mono">{asset.hwid || incident.asset_hw_id}</p>
                            </div>
                        </div>
                    </div>

                    {/* KHỐI 2: PLAYBOOK AI */}
                    <div>
                        <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-2 flex items-center gap-2 border-b border-slate-800 pb-2"><Activity size={12}/> Điều Tra / Khắc Phục (Playbook)</h3>
                        <IncidentPlaybook incident={incident} onUpdate={fetchDetail} />
                    </div>

                    {/* KHỐI 3: RAW ALERTS (LOG GỐC) */}
                    <div>
                        <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-2 flex items-center gap-2 border-b border-slate-800 pb-2"><Fingerprint size={12}/> Logs Sự kiện Gốc ({alerts.length})</h3>
                        <div className="space-y-2">
                            {alerts.map((al, idx) => (
                                <div key={idx} className="bg-[#111827] p-3 rounded border border-slate-800/80 border-l-2 border-l-red-500 shadow-sm">
                                    <div className="flex justify-between items-start mb-1">
                                        <p className="text-[11px] font-bold text-red-400">{al.alert_type}</p>
                                        <p className="text-[9px] font-mono text-slate-500">{new Date(al.created_at).toLocaleTimeString()}</p>
                                    </div>
                                    <p className="text-[11px] text-slate-400 leading-relaxed font-mono whitespace-pre-wrap break-words">{al.description}</p>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>

                {/* NỬA PHẢI (50%): AUDIT TIMELINE & THAO TÁC */}
                <div className="w-1/2 h-full flex flex-col bg-[#050B14] relative">
                    {/* Header Nhỏ Nửa Phải */}
                    <div className="p-3 border-b border-slate-800 bg-[#0A101D] shrink-0 text-[10px] font-black text-slate-500 uppercase tracking-widest flex items-center justify-between">
                        <span>Audit Log & Evidence Tracker</span>
                        <Terminal size={12} className="text-slate-600"/>
                    </div>
                    
                    {/* KHUNG HIỂN THỊ LOG ĐIỀU TRA */}
                    <div ref={chatScrollRef} className="flex-1 overflow-y-auto custom-scrollbar p-5 scroll-smooth bg-[radial-gradient(ellipse_at_top_right,_var(--tw-gradient-stops))] from-slate-900/20 via-[#050B14] to-[#050B14]">
                        <IncidentTimeline incident={incident} />
                    </div>
                    
                    {/* KHUNG NHẬP LIỆU (Action Box) */}
                    <div className="shrink-0 bg-[#0A101D] border-t border-slate-800">
                        <IncidentActionBox status={incident.status} onAction={handleAction} assignee={incident.assignee} />
                    </div>
                </div>

            </div>
        </div>
    );
};

export default IncidentDetail;