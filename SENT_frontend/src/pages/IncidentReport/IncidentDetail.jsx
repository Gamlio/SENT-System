import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import axiosInstance from '../../api/axios';
import { 
    ShieldAlert, Monitor, Network, Clock, UserCheck, ArrowLeft, Bot, 
    TerminalSquare, AlertTriangle, Fingerprint, Lock, Terminal, X, Zap 
} from 'lucide-react';

import IncidentTimeline from './components/DetailParts/IncidentTimeline';
import IncidentActionBox from './components/DetailParts/IncidentActionBox';
import IncidentPlaybook from './components/DetailParts/IncidentPlaybook';

const extractIOCs = (text) => {
    if (!text) return [];
    const iocs = [];
    const ipRegex = /\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b/g;
    const exeRegex = /\b[\w-]+\.(exe|dll|bat|sh|ps1)\b/gi;
    const ips = text.match(ipRegex) || [];
    const exes = text.match(exeRegex) || [];
    ips.forEach(ip => iocs.push({ type: 'IP', value: ip }));
    exes.forEach(exe => iocs.push({ type: 'FILE', value: exe }));
    return iocs;
};

const IncidentDetail = () => {
    const { id } = useParams();
    const navigate = useNavigate();
    
    const [incident, setIncident] = useState(null);
    const [loading, setLoading] = useState(true);
    
    const [terminalOpen, setTerminalOpen] = useState(false);
    const [terminalLogs, setTerminalLogs] = useState([]);
    const [slaTime, setSlaTime] = useState('00:00:00');

    // REF ĐIỀU KHIỂN KHUNG CHAT
    const chatScrollRef = useRef(null);

    const fetchDetail = async () => {
        try {
            const res = await axiosInstance.get(`/incidents/${id}`);
            setIncident(res.data);
        } catch (err) { console.error(err); } finally { setLoading(false); }
    };

    useEffect(() => { fetchDetail(); }, [id]);

    // Tự động cuộn xuống đáy khi có log mới
    useEffect(() => {
        if (chatScrollRef.current) {
            setTimeout(() => {
                chatScrollRef.current.scrollTop = chatScrollRef.current.scrollHeight;
            }, 100);
        }
    }, [incident]);

    useEffect(() => {
        if (!incident || incident.status === 'Resolved') return;
        const interval = setInterval(() => {
            const createdTime = new Date(incident.CreatedAt || incident.created_at).getTime();
            const diff = new Date().getTime() - createdTime;
            const hours = Math.floor(diff / (1000 * 60 * 60)).toString().padStart(2, '0');
            const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60)).toString().padStart(2, '0');
            const seconds = Math.floor((diff % (1000 * 60)) / 1000).toString().padStart(2, '0');
            setSlaTime(`${hours}:${minutes}:${seconds}`);
        }, 1000);
        return () => clearInterval(interval);
    }, [incident]);

    const handleAssignToMe = async () => {
        try {
            await axiosInstance.put(`/incidents/${id}/assign`);
            fetchDetail(); 
        } catch (err) { alert("Lỗi khi nhận xử lý!"); }
    };

    const handleAction = async (type, content, files = [], resolutionSummary = '') => {
        if (type === 'RESOLVE') {
            try {
                const steps = JSON.parse(incident.playbook_progress || '{}').steps || [];
                if (steps.some(s => !s.done)) {
                    alert("⛔ KHÔNG THỂ ĐÓNG SỰ CỐ!\nBạn chưa hoàn thành hết các bước trong Playbook.");
                    return; 
                }
            } catch (e) {}
        }
        try {
            const formData = new FormData();
            formData.append('action_type', type);
            formData.append('content', content);
            if (resolutionSummary) formData.append('resolution_summary', resolutionSummary);
            files.forEach(f => formData.append('files', f));

            await axiosInstance.post(`/incidents/${id}/activity`, formData, { headers: { 'Content-Type': 'multipart/form-data' } });
            fetchDetail(); 
        } catch (err) { alert("Lỗi khi gửi dữ liệu!"); }
    };

    const handleIsolateNetwork = async () => {
        if(window.confirm("CẢNH BÁO: Cắt toàn bộ mạng máy trạm? Hành động này sẽ được ghi log Audit!")) {
            setTerminalOpen(true);
            setTerminalLogs(['[SYSTEM] Bắt đầu gọi API ExecuteLiveAction...']);
            try {
                await axiosInstance.post(`/incidents/${id}/execute`, { command: 'ISOLATE_NETWORK' });
                setTimeout(() => setTerminalLogs(prev => [...prev, `[INFO] Gửi payload tới Endpoint thành công.`]), 800);
                setTimeout(() => setTerminalLogs(prev => [...prev, '[AGENT] Đang ngắt card mạng ngoại vi...']), 1500);
                setTimeout(() => setTerminalLogs(prev => [...prev, '[SUCCESS] Máy trạm đã bị ngắt khỏi hệ thống mạng!']), 2500);
                setTimeout(() => fetchDetail(), 3000); 
            } catch (error) { setTerminalLogs(prev => [...prev, '[ERROR] Đứt kết nối tới Agent hoặc Server lỗi.']); }
        }
    };

    if (loading || !incident) return (
        <div className="h-screen bg-[#050B14] flex flex-col items-center justify-center gap-4">
            <Zap size={40} className="text-emerald-500 animate-pulse"/>
            <p className="text-emerald-500 font-mono text-sm tracking-widest animate-pulse">ESTABLISHING SECURE CONNECTION...</p>
        </div>
    );

    // FIX CHUẨN TÊN HOSTNAME (Lấy đúng chữ thường của Go)
    const agent = incident.agent || incident.Agent || {};
    const alerts = incident.alerts || incident.Alerts || [];

    return (
        <div className="h-[calc(100vh-80px)] bg-[#050B14] text-slate-200 flex flex-col font-sans overflow-hidden">
            
            {/* HEADER TỔNG */}
            <div className="h-16 border-b border-slate-800 bg-[#0A101D] px-6 flex justify-between items-center shrink-0 shadow-md z-10">
                <div className="flex items-center gap-4">
                    <button onClick={() => navigate('/incidents')} className="p-2 bg-slate-800 rounded-lg text-slate-400 hover:text-white transition"><ArrowLeft size={18}/></button>
                    <div className="h-8 w-px bg-slate-700"></div>
                    <span className={`px-3 py-1 rounded-md text-xs font-black uppercase tracking-widest border flex items-center gap-2 ${incident.priority === 'P1' ? 'bg-red-500/10 text-red-500 border-red-500/50 shadow-[0_0_15px_rgba(239,68,68,0.2)] animate-pulse' : 'bg-orange-500/10 text-orange-500 border-orange-500/50'}`}>
                        {incident.priority === 'P1' ? <AlertTriangle size={14}/> : <ShieldAlert size={14}/>}
                        {incident.priority} - {incident.severity}
                    </span>
                    <h1 className="text-xl font-black text-white tracking-tight">INCIDENT <span className="text-slate-500">#{incident.ID || incident.id}</span></h1>
                </div>
                
                <div className="flex items-center gap-6">
                    <div className="flex items-center gap-2 bg-slate-900 border border-slate-800 px-3 py-1.5 rounded-lg">
                        <Clock size={14} className={incident.status === 'Resolved' ? 'text-emerald-500' : 'text-red-500 animate-spin-slow'}/>
                        <span className={`font-mono text-sm font-bold ${incident.status === 'Resolved' ? 'text-emerald-500' : 'text-red-400'}`}>
                            {incident.status === 'Resolved' ? 'Đã đóng' : slaTime}
                        </span>
                    </div>
                    <div className={`px-4 py-1.5 rounded-full text-xs font-bold uppercase tracking-widest border ${incident.status === 'Open' ? 'text-red-400 border-red-500/30 bg-red-500/10' : incident.status === 'Investigating' ? 'text-blue-400 border-blue-500/30 bg-blue-500/10' : 'text-emerald-400 border-emerald-500/30 bg-emerald-500/10'}`}>
                        {incident.status}
                    </div>
                </div>
            </div>

            {/* BỐ CỤC 3 CỘT CHÍNH (ĐÃ FIX TRÀN LƯỚI BẰNG: min-h-0) */}
            <div className="flex-1 grid grid-cols-12 gap-0 overflow-hidden min-h-0">
                
                {/* COL 1: CONTEXT */}
                <div className="col-span-3 h-full overflow-y-auto custom-scrollbar border-r border-slate-800 bg-[#0A101D]/80 p-5 flex flex-col gap-6">
                    <div>
                        <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-3 flex items-center gap-2"><Monitor size={14}/> Nạn nhân (Victim Asset)</h3>
                        <div className="bg-slate-900 rounded-xl border border-slate-700/50 p-4 shadow-lg cursor-pointer hover:border-indigo-500/50 transition group" onClick={() => navigate(`/agents/${agent.hwid}`)}>
                            <div className="flex justify-between items-start mb-2">
                                {/* FIX HOSTNAME */}
                                <p className="text-lg font-bold text-white group-hover:text-indigo-400 transition truncate" title={agent.hostname || 'Unknown Device'}>
                                    {agent.hostname || 'Unknown Device'}
                                </p>
                                <span className={`shrink-0 w-2.5 h-2.5 rounded-full ${agent.status === 'online' ? 'bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.8)]' : 'bg-slate-600'}`}></span>
                            </div>
                            <p className="text-xs text-emerald-400 font-mono mb-3"><Network size={12} className="inline mr-1"/> {agent.ip_address || 'N/A'}</p>
                            
                            <div className="bg-slate-950 rounded-lg p-2 flex justify-between items-center border border-slate-800">
                                <span className="text-[10px] text-slate-500 uppercase font-bold">Risk Score</span>
                                <span className={`text-xs font-black px-2 py-0.5 rounded ${agent.risk_score >= 80 ? 'bg-red-500/20 text-red-400 border border-red-500/30' : 'bg-orange-500/20 text-orange-400 border border-orange-500/30'}`}>
                                    {agent.risk_score || 0}
                                </span>
                            </div>
                        </div>
                    </div>

                    <div>
                        <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-3 flex items-center gap-2"><Fingerprint size={14}/> Cảnh báo & IOCs</h3>
                        <div className="space-y-3">
                            {alerts.map((al, idx) => {
                                const iocs = extractIOCs(al.description);
                                return (
                                    <div key={idx} className="bg-[#111827] p-3 rounded-lg border-l-2 border-red-500 border-y border-r border-slate-800 shadow-md">
                                        <p className="text-xs font-bold text-red-400 mb-1">{al.alert_type}</p>
                                        <p className="text-[11px] text-slate-400 leading-relaxed mb-2 break-words">{al.description}</p>
                                        {iocs.length > 0 && (
                                            <div className="flex flex-wrap gap-1.5 mt-2 pt-2 border-t border-slate-800/50">
                                                {iocs.map((ioc, i) => (
                                                    <span key={i} className={`px-1.5 py-0.5 rounded text-[9px] font-mono font-bold border cursor-pointer hover:scale-105 transition-transform ${ioc.type === 'IP' ? 'bg-blue-500/10 text-blue-400 border-blue-500/30' : 'bg-yellow-500/10 text-yellow-400 border-yellow-500/30'}`}>
                                                        {ioc.type}: {ioc.value}
                                                    </span>
                                                ))}
                                            </div>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                </div>

                {/* COL 2: TIMELINE CHAT (ĐÃ FIX ÉP CỨNG KÍCH THƯỚC CHỐNG TRÀN) */}
                <div className="col-span-6 flex flex-col h-full bg-[#050B14] relative border-r border-slate-800 overflow-hidden min-h-0">
                    {/* Header Cột Giữa */}
                    <div className="p-4 border-b border-slate-800 bg-[#0A101D] shrink-0 z-10">
                        <h2 className="text-base font-bold text-slate-200">{incident.type || incident.description}</h2>
                    </div>
                    
                    {/* KHU VỰC CUỘN LOG (Sẽ tự động hiện scrollbar nếu quá dài) */}
                    <div ref={chatScrollRef} className="flex-1 overflow-y-auto custom-scrollbar p-4 scroll-smooth min-h-0">
                        <IncidentTimeline incident={incident} />
                    </div>
                    
                    {/* KHU VỰC NHẬP TEXT (Ghim chặt xuống đáy) */}
                    <div className="shrink-0 bg-[#0A101D] border-t border-slate-800 z-10">
                        <IncidentActionBox status={incident.status} onAction={handleAction} sending={false} />
                    </div>
                </div>

                {/* COL 3: PLAYBOOK */}
                <div className="col-span-3 h-full overflow-y-auto custom-scrollbar bg-[#0A101D]/80 p-5 flex flex-col gap-6">
                    <div className="bg-slate-900 rounded-xl border border-slate-700/50 p-4 text-center shadow-lg relative overflow-hidden">
                        <div className={`absolute top-0 left-0 w-full h-1 ${incident.assignee ? 'bg-gradient-to-r from-blue-500 to-emerald-500' : 'bg-slate-700'}`}></div>
                        <div className="w-10 h-10 bg-slate-800 rounded-full flex items-center justify-center mx-auto mb-3">
                            <UserCheck size={18} className={incident.assignee ? "text-emerald-400" : "text-slate-500"}/>
                        </div>
                        {incident.assignee ? (
                            <>
                                <p className="text-[9px] text-slate-500 uppercase tracking-widest font-black mb-1">Đang điều tra bởi</p>
                                <p className="text-sm font-bold text-emerald-400">{incident.assignee?.full_name || incident.assignee?.username}</p>
                            </>
                        ) : (
                            <button onClick={handleAssignToMe} className="w-full py-2 bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold rounded-lg transition shadow-lg shadow-blue-500/20 animate-pulse">
                                Nhận Triage Case Này
                            </button>
                        )}
                    </div>

                    <div>
                        <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-3 flex items-center gap-2"><TerminalSquare size={14}/> Live Response</h3>
                        <div className="grid grid-cols-2 gap-3">
                            <button 
                                onClick={handleIsolateNetwork}
                                disabled={incident.status === 'Resolved'}
                                className="flex flex-col items-center justify-center gap-2 p-4 bg-red-500/10 border border-red-500/30 rounded-xl text-red-400 hover:bg-red-500 hover:text-white transition group disabled:opacity-30 disabled:cursor-not-allowed hover:shadow-[0_0_20px_rgba(239,68,68,0.4)]"
                            >
                                <Lock size={20} className="group-hover:scale-110 transition-transform"/>
                                <span className="text-[10px] font-bold uppercase text-center leading-tight">Cô lập<br/>Mạng</span>
                            </button>
                            <button 
                                disabled={incident.status === 'Resolved'}
                                className="flex flex-col items-center justify-center gap-2 p-4 bg-purple-500/10 border border-purple-500/30 rounded-xl text-purple-400 hover:bg-purple-500 hover:text-white transition group disabled:opacity-30 disabled:cursor-not-allowed"
                            >
                                <Bot size={20} className="group-hover:scale-110 transition-transform"/>
                                <span className="text-[10px] font-bold uppercase text-center leading-tight">Hỏi AI<br/>Phân tích</span>
                            </button>
                        </div>
                    </div>

                    <div>
                        <IncidentPlaybook incident={incident} onUpdate={fetchDetail} />
                    </div>
                </div>
            </div>

            {/* --- LIVE TERMINAL MODAL --- */}
            {terminalOpen && (
                <div className="fixed inset-0 bg-black/90 flex items-center justify-center z-[9999] p-4 backdrop-blur-sm">
                    <div className="bg-black w-full max-w-3xl rounded-xl border border-slate-700 shadow-[0_0_50px_rgba(0,0,0,0.8)] overflow-hidden font-mono flex flex-col h-[500px]">
                        <div className="flex justify-between items-center bg-slate-900 px-4 py-2 border-b border-slate-700">
                            <div className="flex items-center gap-3">
                                <Terminal size={16} className="text-emerald-500"/>
                                <span className="text-xs text-slate-400 font-bold">SENT SOC // Live Response Terminal</span>
                            </div>
                            <button onClick={() => setTerminalOpen(false)} className="text-slate-500 hover:text-red-500 transition"><X size={18}/></button>
                        </div>
                        <div className="p-6 flex-1 overflow-y-auto text-[13px] leading-relaxed space-y-2 text-slate-300">
                            <p className="text-slate-500 italic mb-4">Kết nối tới socket <span className="text-blue-400">{agent.ip_address || 'N/A'}:8443</span>...</p>
                            {terminalLogs.map((log, i) => (
                                <p key={i} className={
                                    log.includes('[SUCCESS]') ? 'text-emerald-400' :
                                    log.includes('[AGENT]') ? 'text-yellow-400' :
                                    log.includes('[ERROR]') ? 'text-red-500' : 'text-slate-300'
                                }>
                                    <span className="opacity-50 mr-2">{new Date().toLocaleTimeString()}</span> {log}
                                </p>
                            ))}
                            {terminalLogs.length < 5 && !terminalLogs.some(l => l.includes('[ERROR]')) && <p className="animate-pulse">_</p>}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default IncidentDetail;