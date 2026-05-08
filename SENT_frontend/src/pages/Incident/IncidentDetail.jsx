import React, { useEffect, useState } from 'react';
import { useIncidents } from './hooks/useIncidents';
import AuditChat from './components/AuditChat';
import { ChevronLeft, Monitor, Terminal, ShieldAlert, CheckCircle2, Lock, AlertCircle } from 'lucide-react';
import axiosInstance from '../../../api/axios';
import { useAuth } from '../../context/AuthContext';

const IncidentDetail = ({ incidentId, onBack }) => {
    const { user } = useAuth();
    const { detail, loading, fetchDetail } = useIncidents();
    const [isClosing, setIsClosing] = useState(false);

    useEffect(() => { 
        fetchDetail(incidentId); 
    }, [incidentId, fetchDetail]);

    if (loading || !detail) return (
        <div className="h-full flex items-center justify-center text-indigo-500 font-mono text-[10px] animate-pulse">
            LOADING_FORENSIC_DATABASE...
        </div>
    );

    const { incident, audit_logs } = detail;

    // Hàm xử lý đóng sự cố
    const handleCloseCase = async () => {
        const note = prompt("Nhập báo cáo tổng kết điều tra (Tối thiểu 15 ký tự):");
        
        if (note === null) return; // Người dùng bấm cancel
        if (note.length < 15) {
            alert("Báo cáo quá ngắn! Vui lòng nhập chi tiết hơn (tối thiểu 15 ký tự).");
            return;
        }

        setIsClosing(true);
        try {
            await axiosInstance.post('/incidents/close', {
                incident_id: incident.id || incident.ID,
                note: note,
                // Gửi snapshot trạng thái máy hiện tại làm bằng chứng Baseline
                evidence_data: JSON.stringify({
                    asset_state: incident.asset,
                    closed_at: new Date().toISOString(),
                    verdict: "RESOLVED_BY_HUMAN_OPERATOR"
                })
            });
            
            alert("Hồ sơ đã được đóng và niêm phong bảo mật!");
            fetchDetail(incidentId); // Tải lại để cập nhật trạng thái
        } catch (err) {
            alert(err.response?.data?.error || "Không thể đóng hồ sơ. Kiểm tra lại quyền hạn hoặc bằng chứng P1.");
        } finally {
            setIsClosing(false);
        }
    };

    return (
        <div className="h-[calc(100vh-60px)] bg-[#050B14] flex flex-col overflow-hidden text-slate-300">
            {/* Top Bar - Tích hợp nút Đóng Case */}
            <div className="px-4 py-2 bg-[#0A101D] border-b border-slate-800 flex justify-between items-center shrink-0">
                <div className="flex items-center gap-4">
                    <button onClick={onBack} className="flex items-center gap-2 text-slate-500 hover:text-white transition-colors text-[9px] font-black uppercase">
                        <ChevronLeft size={14}/> Quay lại
                    </button>
                    <div className="h-4 w-px bg-slate-800"></div>
                    <span className="text-[10px] font-black text-red-500 uppercase tracking-widest">Case ID: {incident.id || incident.ID}</span>
                </div>

                <div className="flex items-center gap-3">
                    {/* Hiển thị nút đóng case nếu chưa resolved và có quyền */}
                    {incident.status !== 'Resolved' ? (
                        user?.permissions?.incident_action && (
                            <button 
                                onClick={handleCloseCase}
                                disabled={isClosing}
                                className="bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white text-[9px] px-3 py-1.5 rounded flex items-center gap-2 font-black uppercase transition-all shadow-lg shadow-emerald-500/10"
                            >
                                {isClosing ? <div className="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin"></div> : <CheckCircle2 size={12}/>}
                                Kết thúc điều tra
                            </button>
                        )
                    ) : (
                        <div className="flex items-center gap-2 px-3 py-1.5 bg-slate-900 border border-slate-800 rounded">
                            <Lock size={12} className="text-emerald-500"/>
                            <span className="text-[9px] font-mono text-emerald-500 font-bold uppercase">Case Sealed</span>
                        </div>
                    )}
                </div>
            </div>

            {/* Main Workspace */}
            <div className="flex-1 flex min-h-0">
                {/* CỘT TRÁI: Thông tin chi tiết & Cảnh báo */}
                <div className="w-1/2 border-r border-slate-800 overflow-y-auto p-4 custom-scrollbar space-y-4 bg-black/20">
                    {/* Trạng thái hiện tại */}
                    <div className={`p-3 rounded-lg border flex items-center justify-between ${incident.status === 'Resolved' ? 'bg-emerald-500/5 border-emerald-500/20' : 'bg-orange-500/5 border-orange-500/20'}`}>
                        <div className="flex items-center gap-3">
                            <AlertCircle size={18} className={incident.status === 'Resolved' ? 'text-emerald-500' : 'text-orange-500'}/>
                            <div>
                                <p className="text-[8px] uppercase font-bold text-slate-500">Current Status</p>
                                <p className={`text-xs font-black uppercase ${incident.status === 'Resolved' ? 'text-emerald-500' : 'text-orange-500'}`}>
                                    {incident.status}
                                </p>
                            </div>
                        </div>
                        {incident.last_audit_hash && (
                            <div className="text-right">
                                <p className="text-[8px] uppercase font-bold text-slate-600 font-mono">Seal Hash</p>
                                <p className="text-[9px] font-mono text-slate-500 italic">{incident.last_audit_hash.substring(0, 12)}...</p>
                            </div>
                        )}
                    </div>

                    <div className="bg-[#0A101D] border border-slate-800 rounded-lg p-4 shadow-xl">
                        <h2 className="text-xs font-black text-white uppercase mb-2 tracking-widest flex items-center gap-2">
                            <ShieldAlert size={14} className="text-red-500"/> {incident.type}
                        </h2>
                        <p className="text-[10px] text-slate-400 bg-black/40 p-3 rounded border border-slate-800 leading-relaxed font-sans">
                            {incident.description}
                        </p>
                    </div>

                    <div className="bg-[#0A101D] border border-slate-800 rounded-lg p-4">
                        <span className="text-[9px] font-black text-slate-500 uppercase tracking-widest mb-3 flex items-center gap-2 border-b border-slate-800 pb-2">
                            <Monitor size={12} className="text-emerald-500"/> Victim Fingerprint
                        </span>
                        <div className="space-y-2">
                            <div className="bg-[#050B14] p-2 rounded border border-slate-800">
                                <span className="text-[8px] text-slate-600 uppercase font-bold block">Asset HWID</span>
                                <code className="text-[9px] text-indigo-400 font-mono break-all">{incident.asset_hwid}</code>
                            </div>
                            <div className="grid grid-cols-2 gap-2 text-[9px]">
                                <div className="bg-[#050B14] p-2 rounded border border-slate-800">
                                    <span className="text-slate-600 uppercase block">IP Address</span>
                                    <span className="text-slate-200 font-mono">{incident.asset?.ip_address || "10.0.x.x"}</span>
                                </div>
                                <div className="bg-[#050B14] p-2 rounded border border-slate-800 text-right">
                                    <span className="text-slate-600 uppercase block">Risk Score</span>
                                    <span className="text-red-500 font-black">{incident.asset?.risk_score || 0}/100</span>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Hiển thị tóm tắt giải quyết nếu đã đóng case */}
                    {incident.resolution_summary && (
                        <div className="bg-emerald-500/5 border border-emerald-500/20 rounded-lg p-4">
                            <span className="text-[9px] font-black text-emerald-500 uppercase tracking-widest mb-2 block border-b border-emerald-500/10 pb-2">
                                Resolution Summary
                            </span>
                            <p className="text-[10px] text-slate-300 italic leading-relaxed">
                                "{incident.resolution_summary}"
                            </p>
                        </div>
                    )}

                    <div className="bg-[#0A101D] border border-slate-800 rounded-lg p-4">
                        <span className="text-[9px] font-black text-slate-500 uppercase tracking-widest mb-3 flex items-center gap-2 border-b border-slate-800 pb-2">
                            <Terminal size={12} className="text-red-500"/> Raw Evidence Logs
                        </span>
                        <div className="space-y-1.5">
                            {(incident.alerts || []).map((alert, i) => (
                                <div key={i} className="bg-black/20 p-2 rounded border-l-2 border-red-500/30 border border-slate-800/40">
                                    <div className="flex justify-between text-[8px] font-mono text-slate-500 mb-1">
                                        <span className="text-red-400/80">{alert.alert_type}</span>
                                        <span>{new Date(alert.created_at).toLocaleTimeString()}</span>
                                    </div>
                                    <p className="text-[9px] text-slate-400 font-mono leading-tight">{alert.description}</p>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>

                {/* CỘT PHẢI: Khung Chat Audit */}
                <div className="w-1/2 flex flex-col bg-[#050B14]">
                    <AuditChat 
                        incidentId={incidentId} 
                        auditLogs={audit_logs} 
                        onUpdate={() => fetchDetail(incidentId)} 
                    />
                </div>
            </div>
        </div>
    );
};

export default IncidentDetail;