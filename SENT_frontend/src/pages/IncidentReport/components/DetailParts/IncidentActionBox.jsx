import React, { useState, useRef } from 'react';
import { Send, CheckCircle, Loader2, Paperclip, X, ShieldCheck, ImagePlus } from 'lucide-react';

const IncidentActionBox = ({ status, onAction, sending, assignee }) => {
    const [note, setNote] = useState('');
    const [files, setFiles] = useState([]);
    const [isResolving, setIsResolving] = useState(false);
    const [resolutionSummary, setResolutionSummary] = useState('');
    const fileInputRef = useRef(null);

    // XỬ LÝ QUYỀN HẠN
    // Giả sử thông tin User đã được lưu vào LocalStorage lúc đăng nhập
    const userString = localStorage.getItem('user');
    const user = userString ? JSON.parse(userString) : {};
    
    // Quyền Đóng Case: Dành cho Admin, người có quyền perm_incident_action, HOẶC chính là người được Assign
    const hasPermissionToResolve = user.role === 'admin' || user.perm_incident_action === true || (assignee && assignee.id === user.id);

    const handleFileSelect = (e) => {
        if (e.target.files) setFiles([...files, ...Array.from(e.target.files)]);
    };
    const removeFile = (index) => setFiles(files.filter((_, i) => i !== index));

    const handleSubmit = (type) => {
        onAction(type, type === 'RESOLVE' ? `[BÁO CÁO KẾT LUẬN]\n${resolutionSummary}` : note, files, resolutionSummary);
        if (type === 'COMMENT') { setNote(''); setFiles([]); }
        setIsResolving(false);
    };

    if (isResolving) {
        return (
            <div className="p-4 bg-emerald-950/20 border-t border-emerald-900/50">
                <div className="flex items-center gap-2 mb-2 text-emerald-400 font-bold text-xs">
                    <ShieldCheck size={14}/> Báo cáo Tổng kết & Đóng sự cố
                </div>
                <textarea 
                    value={resolutionSummary} onChange={(e) => setResolutionSummary(e.target.value)}
                    placeholder="Ghi rõ nguyên nhân (Root Cause) và các bước xử lý..."
                    className="w-full bg-[#050B14] text-emerald-100 text-[11px] p-3 rounded border border-emerald-900/50 focus:border-emerald-500 outline-none resize-none h-24 mb-3 font-mono"
                />
                <div className="flex justify-end gap-2">
                    <button onClick={() => setIsResolving(false)} className="px-3 py-1.5 text-slate-500 hover:text-white text-[10px] font-bold uppercase tracking-widest transition">Hủy</button>
                    <button onClick={() => handleSubmit('RESOLVE')} disabled={resolutionSummary.trim().length < 5 || sending} className="px-4 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-white text-[10px] font-black uppercase tracking-widest rounded flex items-center gap-2 shadow-lg disabled:opacity-50">
                        {sending ? <Loader2 size={14} className="animate-spin"/> : <CheckCircle size={14}/>} Đóng Case
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="p-4 bg-[#0A101D]">
            {/* KHU VỰC HIỂN THỊ FILE UPLOAD */}
            {files.length > 0 && (
                <div className="flex gap-2 mb-3 overflow-x-auto custom-scrollbar pb-1">
                    {files.map((file, idx) => (
                        <div key={idx} className="relative group shrink-0 w-12 h-12 rounded bg-slate-800 border border-slate-700 flex items-center justify-center overflow-hidden shadow-md">
                            {file.type.startsWith('image/') ? <img src={URL.createObjectURL(file)} alt="preview" className="w-full h-full object-cover"/> : <span className="text-[8px] text-slate-500 font-mono p-1 break-all text-center">{file.name}</span>}
                            <button onClick={() => removeFile(idx)} className="absolute top-0 right-0 bg-red-500 text-white p-0.5 opacity-0 group-hover:opacity-100 transition"><X size={10}/></button>
                        </div>
                    ))}
                </div>
            )}

            {/* KHU VỰC NHẬP TEXT */}
            <div className="flex flex-col gap-2">
                <div className="flex gap-2">
                    <textarea 
                        value={note} onChange={(e) => setNote(e.target.value)} disabled={sending || status === 'Resolved'}
                        placeholder={status === 'Resolved' ? "Sự cố đã đóng. Không thể cập nhật thêm." : "Nhập tiến trình điều tra..."}
                        className="flex-1 bg-[#050B14] text-slate-300 text-xs p-3 rounded border border-slate-800 focus:border-indigo-500/50 outline-none resize-none h-14"
                    />
                    <div className="flex flex-col gap-2 shrink-0">
                        <button onClick={() => fileInputRef.current?.click()} disabled={status === 'Resolved'} className="h-6 w-10 bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white rounded flex items-center justify-center transition disabled:opacity-50"><ImagePlus size={12}/></button>
                        <button onClick={() => handleSubmit('COMMENT')} disabled={(!note.trim() && files.length === 0) || sending || status === 'Resolved'} className="h-6 w-10 bg-indigo-600 hover:bg-indigo-500 text-white rounded flex items-center justify-center transition shadow-lg disabled:opacity-50 disabled:bg-slate-800">
                            {sending ? <Loader2 size={12} className="animate-spin"/> : <Send size={12}/>}
                        </button>
                        <input type="file" multiple ref={fileInputRef} className="hidden" onChange={handleFileSelect} accept="image/*,.pdf,.txt,.zip" />
                    </div>
                </div>

                {/* NÚT THAO TÁC (CHỈ HIỆN KHI CÓ QUYỀN) */}
                <div className="flex gap-2">
                    {status === 'Open' && (
                        <button onClick={() => handleSubmit('INVESTIGATE')} disabled={sending} className="flex-1 py-1.5 bg-blue-600/10 text-blue-400 hover:bg-blue-600 hover:text-white text-[10px] font-bold uppercase tracking-widest rounded border border-blue-600/30 transition">
                            Nhận điều tra (Triage)
                        </button>
                    )}
                    {status !== 'Resolved' && hasPermissionToResolve && (
                        <button onClick={() => setIsResolving(true)} disabled={sending} className="flex-1 py-1.5 bg-emerald-600/10 text-emerald-400 hover:bg-emerald-600 hover:text-white text-[10px] font-bold uppercase tracking-widest rounded border border-emerald-600/30 transition">
                            Kết thúc hồ sơ (Resolve)
                        </button>
                    )}
                </div>
            </div>
        </div>
    );
};

export default IncidentActionBox;