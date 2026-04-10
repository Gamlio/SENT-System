// components/DetailParts/IncidentAuditTrail.jsx
import { Fingerprint, ShieldCheck, ShieldAlert, Terminal, Search } from 'lucide-react';
import axiosInstance from '../../../api/axios';

const AuditTrailItem = ({ audit }) => {
    const isSystem = !audit.user_id; // Check con trỏ UserID
    
    const verifyHash = async (auditId) => {
        const res = await axiosInstance.get(`/incidents/audit/${auditId}/verify`);
        if (res.data.is_tampered) alert("CẢNH BÁO: Log này đã bị can thiệp!");
        else alert("Xác thực thành công: Dữ liệu nguyên bản.");
    };

    return (
        <div className="relative pl-6 pb-6 border-l border-slate-800 last:pb-0">
            {/* Dot Icon */}
            <div className="absolute -left-[5px] top-0 w-2 h-2 rounded-full bg-slate-700 border border-[#050B14]"></div>
            
            <div className="bg-[#0A101D] border border-slate-800/50 rounded-md p-3">
                <div className="flex justify-between items-center mb-2">
                    <div className="flex items-center gap-2">
                        <span className="text-[10px] font-black text-indigo-400 uppercase tracking-widest">{audit.action_type}</span>
                        <span className="text-slate-600">|</span>
                        <span className="text-[10px] text-slate-500 font-mono">
                            {isSystem ? "SYSTEM_AI" : audit.user?.username}
                        </span>
                    </div>
                    <span className="text-[9px] text-slate-600 font-mono italic">
                        {new Date(audit.created_at).toLocaleString()}
                    </span>
                </div>

                <p className="text-xs text-slate-300 leading-relaxed mb-3">{audit.content}</p>

                {/* Technical Evidence Block */}
                {audit.evidence_data && (
                    <div className="bg-black/40 border border-slate-800 rounded p-2 mb-2 font-mono text-[10px]">
                        <div className="flex justify-between items-center text-emerald-500 mb-1">
                            <span className="flex items-center gap-1"><ShieldCheck size={10}/> Technical Evidence</span>
                            <button onClick={() => verifyHash(audit.id)} className="text-slate-500 hover:text-indigo-400 underline uppercase text-[8px]">Verify Hash</button>
                        </div>
                        <pre className="text-slate-400 overflow-x-auto">{JSON.stringify(JSON.parse(audit.evidence_data), null, 2)}</pre>
                    </div>
                )}

                <div className="flex items-center gap-3 text-[9px] text-slate-600 font-mono">
                    <span className="flex items-center gap-1"><Terminal size={10}/> IP: {audit.ip_address}</span>
                    <span className="flex items-center gap-1"><Fingerprint size={10}/> HASH: {audit.audit_hash?.substring(0, 12)}...</span>
                </div>
            </div>
        </div>
    );
};
export default AuditTrailItem;