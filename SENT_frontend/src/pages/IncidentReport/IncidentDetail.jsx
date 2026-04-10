import React, { useState, useEffect } from 'react';
import axiosInstance from '../../api/axios';
import AuditTrailItem from "./components/IncidentAuditTrail"; 
import IncidentActionBox from './components/IncidentActionBox';
import { Radar } from 'lucide-react';

const IncidentDetail = ({ incidentId, onBack }) => {
    const [data, setData] = useState(null);
    const [loading, setLoading] = useState(true);

    const fetchDetail = async () => {
        try {
            const res = await axiosInstance.get(`/incidents/${incidentId}`);
            setData(res.data);
        } catch (err) {
            console.error("Không thể tải chi tiết case:", err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => { fetchDetail(); }, [incidentId]);

    // Màn hình Loading đồng bộ
    if (loading || !data) return (
        <div className="h-[calc(100vh-60px)] bg-[#050B14] flex flex-col items-center justify-center gap-4">
            <Radar size={40} className="text-indigo-500 animate-spin-slow"/>
            <p className="text-indigo-500 font-mono text-sm tracking-widest animate-pulse uppercase">Fetching Case Evidence...</p>
        </div>
    );

    const { incident, audit_logs } = data;

    return (
        // ĐỒNG BỘ: Wrapper bọc toàn màn hình
        <div className="h-[calc(100vh-60px)] bg-[#050B14] text-slate-200 p-5 font-sans overflow-y-auto custom-scrollbar">
            
            <div className="grid grid-cols-1 lg:grid-cols-[1fr,350px] gap-6">
                <div className="space-y-4">
                    {/* Header Card */}
                    <div className="bg-[#0A101D] border border-slate-800 shadow-lg rounded-xl p-6 relative overflow-hidden">
                        <button onClick={onBack} className="text-slate-500 hover:text-indigo-400 text-[10px] font-black uppercase tracking-widest mb-4 transition-colors">← Back to Workbench</button>
                        <div className="flex justify-between items-start">
                            <h2 className="text-2xl font-black text-white tracking-tighter uppercase">{incident.type}</h2>
                            <span className="px-3 py-1 bg-[#050B14] border border-slate-700 text-indigo-400 text-xs font-black rounded uppercase">{incident.status}</span>
                        </div>
                        <p className="text-sm text-slate-400 mt-4 leading-relaxed bg-[#050B14] p-4 rounded-lg border border-slate-800">{incident.description}</p>
                    </div>

                    {/* Audit Trail Card */}
                    <div className="bg-[#0A101D] border border-slate-800 shadow-lg rounded-xl p-6">
                        <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-6 border-b border-slate-800 pb-3">Action Timeline & Technical Proof</h3>
                        <div className="space-y-0">
                            {audit_logs.map(log => (
                                <AuditTrailItem key={log.id} audit={log} />
                            ))}
                        </div>
                    </div>
                </div>

                {/* Right Column: Action Box */}
                <div className="space-y-4">
                    <IncidentActionBox 
                        incident={incident} 
                        onSuccess={fetchDetail} 
                    />
                </div>
            </div>
        </div>
    );
};

export default IncidentDetail;