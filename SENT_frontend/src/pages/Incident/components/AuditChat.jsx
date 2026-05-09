import React, { useState, useRef } from 'react';
import { Terminal, Fingerprint, Send, ShieldCheck, Loader2, X, Paperclip, Maximize2, Image as ImageIcon } from 'lucide-react';
import axiosInstance from '../../../api/axios';

const AuditChat = ({ incidentId, auditLogs, onUpdate }) => {
    const [note, setNote] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [selectedFiles, setSelectedFiles] = useState([]);
    const [previewImage, setPreviewImage] = useState(null); 
    const fileInputRef = useRef(null);

    // Xử lý xóa ảnh trong danh sách chờ
    const removeFile = (index) => {
        setSelectedFiles(prev => prev.filter((_, i) => i !== index));
    };

    const handleAction = async () => {
        // Đảm bảo lấy ID từ props hoặc object incident một cách chắc chắn
        const cleanId = typeof incidentId === 'object' ? (incidentId.id || incidentId.ID) : incidentId;
        
        if (!cleanId) {
            alert("Lỗi: Không tìm thấy ID hồ sơ.");
            return;
        }

        if (!note.trim() && selectedFiles.length === 0) return;
        
        setIsSubmitting(true);
        const formData = new FormData();
        
        // [QUAN TRỌNG] Append các trường text trước để Backend parse dễ hơn
        formData.append("incident_id", String(cleanId));
        formData.append("note", note.trim());
        
        // Append danh sách file
        selectedFiles.forEach(file => {
            formData.append("files", file); // Key 'files' phải khớp với backend form.File["files"]
        });

        try {
            // [FIX] Để Axios tự quản lý Content-Type cho FormData
            await axiosInstance.post('/incidents/audit/upload', formData, {
                headers: {
                    'Content-Type': 'multipart/form-data'
                }
            });
            
            setNote("");
            setSelectedFiles([]);
            if (onUpdate) onUpdate(); // Tải lại Forensic Timeline
        } catch (err) {
            console.error("Upload error:", err.response?.data);
            alert(err.response?.data?.error || "Lỗi lưu bằng chứng: Kiểm tra kích thước ảnh hoặc kết nối.");
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="flex flex-col h-full bg-[#0A101D] border-l border-slate-800 relative">
            {/* Overlay phóng to ảnh */}
            {previewImage && (
                <div className="fixed inset-0 z-[9999] bg-black/90 flex items-center justify-center p-10 backdrop-blur-sm" onClick={() => setPreviewImage(null)}>
                    <img src={previewImage} alt="Evidence" className="max-w-full max-h-full object-contain" />
                </div>
            )}

            <div className="px-4 py-2.5 border-b border-slate-800 bg-slate-900/40 flex justify-between items-center shrink-0">
                <div className="flex items-center gap-2">
                    <Terminal size={14} className="text-emerald-500"/>
                    <h3 className="text-[9px] font-black uppercase tracking-[0.2em] text-slate-400">Forensic Timeline</h3>
                </div>
            </div>

            {/* Danh sách log */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4 custom-scrollbar bg-[#050B14]/30">
                {auditLogs?.map((log, idx) => (
                <div key={idx} className="relative pl-5 border-l border-slate-800/50">
                    {/* Điểm mốc Timeline */}
                    <div className="absolute -left-[3.5px] top-1.5 w-1.5 h-1.5 rounded-full bg-indigo-500 shadow-[0_0_8px_rgba(99,102,241,0.5)]"></div>
                    
                    <div className="bg-slate-900/40 p-3 rounded border border-slate-800/50 hover:border-slate-700 transition-colors">
                        {/* Header: Tên người gửi và Thời gian */}
                        <div className="flex justify-between items-center mb-2">
                            <div className="flex items-center gap-2">
                                {/* Icon đại diện nhỏ */}
                                <div className="w-5 h-5 rounded bg-indigo-500/20 flex items-center justify-center border border-indigo-500/30">
                                    <span className="text-[8px] font-black text-indigo-400">
                                        {log.user_name?.substring(0, 2).toUpperCase() || "SY"}
                                    </span>
                                </div>
                                <span className="text-[10px] font-black text-indigo-400 tracking-tight">
                                    {log.user_name || "Hệ thống"}
                                </span>
                                <span className="text-[8px] px-1.5 py-0.5 rounded bg-slate-800 text-slate-500 font-mono">
                                    {log.action_type}
                                </span>
                            </div>
                            <span className="text-[8px] text-slate-600 font-mono">
                                {new Date(log.created_at).toLocaleString()}
                            </span>
                        </div>

                        {/* Nội dung tin nhắn */}
                        <p className="text-[10px] text-slate-300 leading-relaxed font-sans mb-2">
                            {log.content}
                        </p>
                        
                        {/* Hiển thị ảnh bằng chứng (nếu có) */}
                        {log.images && log.images.length > 0 && (
                            <div className="flex flex-wrap gap-2 mt-2 pt-2 border-t border-slate-800/50">
                                {log.images.map((img, i) => (
                                    <div key={i} className="group relative">
                                        <img 
                                            src={`/uploads/audits/${img}`} 
                                            className="w-16 h-16 object-cover rounded border border-slate-800 cursor-zoom-in hover:border-indigo-500/50 transition-all"
                                            onClick={() => setPreviewImage(`/uploads/audits/${img}`)}
                                            alt="forensic evidence"
                                        />
                                        <div className="absolute inset-0 bg-indigo-500/10 opacity-0 group-hover:opacity-100 pointer-events-none rounded transition-opacity"></div>
                                    </div>
                                ))}
                            </div>
                        )}
                        
                        {/* Hiển thị mã Hash niêm phong để kiểm tra tính toàn vẹn */}
                        {log.audit_hash && (
                            <div className="mt-2 flex items-center gap-1 opacity-30 hover:opacity-100 transition-opacity">
                                <Fingerprint size={8} className="text-emerald-500"/>
                                <span className="text-[7px] font-mono text-slate-500 truncate max-w-[200px]">
                                    SEAL_HASH: {log.audit_hash}
                                </span>
                            </div>
                        )}
                    </div>
                </div>
            ))}
            </div>

            {/* KHU VỰC NHẬP LIỆU - NƠI SỬA LỖI CHÍNH */}
            <div className="p-3 bg-slate-900/20 border-t border-slate-800 shrink-0">
                
                {/* [MỚI] Hiển thị danh sách ảnh đang chờ gửi */}
                {selectedFiles.length > 0 && (
                    <div className="flex flex-wrap gap-2 mb-3 p-2 bg-black/40 rounded border border-slate-800/50">
                        {selectedFiles.map((file, i) => (
                            <div key={i} className="relative w-12 h-12 group">
                                <img 
                                    src={URL.createObjectURL(file)} 
                                    className="w-full h-full object-cover rounded border border-indigo-500/50" 
                                    alt="preview"
                                />
                                <button 
                                    onClick={() => removeFile(i)}
                                    className="absolute -top-1 -right-1 bg-red-500 rounded-full text-white p-0.5 hover:bg-red-400"
                                >
                                    <X size={8} />
                                </button>
                            </div>
                        ))}
                    </div>
                )}

                <div className="relative flex items-end gap-2">
                    <button 
                        type="button"
                        onClick={() => fileInputRef.current.click()}
                        className={`p-2 border rounded transition-colors ${selectedFiles.length > 0 ? 'bg-indigo-500/10 border-indigo-500 text-indigo-400' : 'bg-[#050B14] border-slate-800 text-slate-500'}`}
                    >
                        <Paperclip size={14}/>
                    </button>
                    
                    {/* [FIX] onChange sử dụng Array.from để xử lý FileList */}
                    <input 
                        type="file" 
                        ref={fileInputRef} 
                        multiple 
                        hidden 
                        accept="image/*" 
                        onChange={(e) => {
                            if (e.target.files) {
                                setSelectedFiles(prev => [...prev, ...Array.from(e.target.files)]);
                                e.target.value = null; // Reset để có thể chọn lại cùng 1 file
                            }
                        }} 
                    />

                    <textarea 
                        value={note}
                        onChange={(e) => setNote(e.target.value)}
                        placeholder="Nhập ghi chú hoặc đính kèm bằng chứng hình ảnh..."
                        className="flex-1 bg-[#050B14] border border-slate-800 rounded p-2 text-[10px] h-14 text-slate-300 font-mono outline-none focus:border-indigo-500/50"
                    />
                    
                    <button 
                        onClick={handleAction}
                        disabled={isSubmitting || (!note.trim() && selectedFiles.length === 0)}
                        className="p-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded disabled:opacity-20 transition-all"
                    >
                        {isSubmitting ? <Loader2 size={16} className="animate-spin"/> : <Send size={16}/>}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default AuditChat;