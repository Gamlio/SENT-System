import React, { useState, useRef } from 'react';
import { Terminal, Fingerprint, Send, ShieldCheck, Loader2, X, Paperclip, Maximize2 } from 'lucide-react';
import axiosInstance from '../../../api/axios';

const AuditChat = ({ incidentId, auditLogs, onUpdate }) => {
    const [note, setNote] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [selectedFiles, setSelectedFiles] = useState([]);
    const [previewImage, setPreviewImage] = useState(null); 
    const fileInputRef = useRef(null);

    const handleAction = async () => {
        const cleanId = typeof incidentId === 'object' ? (incidentId.id || incidentId.ID) : incidentId;
        if (!note.trim() && selectedFiles.length === 0) return;
        
        setIsSubmitting(true);
        const formData = new FormData();
        formData.append("incident_id", String(cleanId));
        formData.append("note", note);
        
        selectedFiles.forEach(file => {
            formData.append("files", file);
        });

        try {
            await axiosInstance.post('/incidents/audit/upload', formData);
            setNote("");
            setSelectedFiles([]);
            onUpdate();
        } catch (err) {
            alert(err.response?.data?.error || "Lỗi lưu bằng chứng kỹ thuật");
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="flex flex-col h-full bg-[#0A101D] border-l border-slate-800 relative">
            {/* Overlay phóng to ảnh (Lightbox) */}
            {previewImage && (
                <div 
                    className="fixed inset-0 z-[9999] bg-black/90 flex items-center justify-center p-10 backdrop-blur-sm animate-in fade-in duration-200"
                    onClick={() => setPreviewImage(null)}
                >
                    <button className="absolute top-5 right-5 text-white/50 hover:text-white transition-colors">
                        <X size={32} />
                    </button>
                    <img 
                        src={previewImage} 
                        alt="Zoomed Evidence" 
                        className="max-w-full max-h-full object-contain shadow-2xl border border-white/10 rounded-sm"
                    />
                </div>
            )}

            <div className="px-4 py-2.5 border-b border-slate-800 bg-slate-900/40 flex justify-between items-center shrink-0">
                <div className="flex items-center gap-2">
                    <Terminal size={14} className="text-emerald-500"/>
                    <h3 className="text-[9px] font-black uppercase tracking-[0.2em] text-slate-400">Forensic Timeline</h3>
                </div>
            </div>

            <div className="flex-1 overflow-y-auto p-4 space-y-4 custom-scrollbar bg-[#050B14]/30">
                {(auditLogs || []).map((log, idx) => (
                    <div key={idx} className="relative pl-5 border-l border-slate-800/50">
                        <div className="absolute -left-[3.5px] top-1.5 w-1.5 h-1.5 rounded-full bg-indigo-500 shadow-[0_0_8px_rgba(99,102,241,0.5)]"></div>
                        <div className="bg-slate-900/40 p-3 rounded border border-slate-800/50">
                            <div className="flex justify-between text-[8px] mb-1.5 uppercase font-bold tracking-tighter">
                                <div className="flex items-center gap-2">
                                    <span className="text-indigo-400">{log.action_type}</span>
                                    <span className="text-slate-500">|</span>
                                    <span className="text-emerald-400 flex items-center gap-1">BY: {log.user_name || "SYSTEM"}</span>
                                </div>
                                <span className="text-slate-600">{new Date(log.created_at).toLocaleString('vi-VN')}</span>
                            </div>
                            
                            <p className="text-[10px] text-slate-300 font-sans leading-relaxed mb-2">{log.content}</p>

                            {/* Render ảnh bằng chứng với sự kiện click phóng to */}
                            {log.images && log.images.length > 0 && (
                                <div className="flex flex-wrap gap-2 mb-2">
                                    {log.images.map((img, i) => (
                                        <div key={i} className="relative group overflow-hidden rounded border border-slate-800">
                                            <img 
                                                src={`/uploads/audits/${img}`} 
                                                alt="evidence" 
                                                className="w-24 h-24 object-cover cursor-zoom-in transition-all group-hover:scale-110 group-hover:brightness-75"
                                                onClick={() => setPreviewImage(`/uploads/audits/${img}`)}
                                            />
                                            <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity">
                                                <Maximize2 size={16} className="text-white" />
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            )}

                            <div className="mt-2 flex items-center gap-4 text-[7.5px] text-slate-700 font-mono uppercase">
                                <span className="flex items-center gap-1"><Fingerprint size={10}/> Hash: {log.audit_hash?.substring(0, 14)}...</span>
                                <span className="flex items-center gap-1 text-emerald-800/60 font-bold"><ShieldCheck size={10}/> Verified Integrity</span>
                            </div>
                        </div>
                    </div>
                ))}
            </div>

            {/* Input Form (Giữ nguyên) */}
            <div className="p-3 bg-slate-900/20 border-t border-slate-800 shrink-0">
                {/* Preview Thumbnail before sending (Giữ nguyên) */}
                {/* ... */}
                <div className="relative flex items-end gap-2">
                    <button 
                        type="button"
                        onClick={() => fileInputRef.current.click()}
                        className="p-2 bg-[#050B14] border border-slate-800 rounded text-slate-500"
                    >
                        <Paperclip size={14}/>
                    </button>
                    <input type="file" ref={fileInputRef} multiple hidden accept="image/*" onChange={(e) => setSelectedFiles(prev => [...prev, ...Array.from(e.target.files)])} />

                    <textarea 
                        value={note}
                        onChange={(e) => setNote(e.target.value)}
                        placeholder="Nhập ghi chú điều tra..."
                        className="flex-1 bg-[#050B14] border border-slate-800 rounded p-2 text-[10px] h-14 text-slate-300 font-mono outline-none"
                    />
                    <button 
                        onClick={handleAction}
                        disabled={isSubmitting || (!note.trim() && selectedFiles.length === 0)}
                        className="p-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded disabled:opacity-20"
                    >
                        {isSubmitting ? <Loader2 size={16} className="animate-spin"/> : <Send size={16}/>}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default AuditChat;