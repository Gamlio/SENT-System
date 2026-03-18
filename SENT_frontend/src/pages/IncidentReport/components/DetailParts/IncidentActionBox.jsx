import React, { useState, useRef } from 'react';
import { Send, CheckCircle, FileText, Loader2, Paperclip, X, ShieldCheck } from 'lucide-react';

const IncidentActionBox = ({ status, onAction, sending }) => {
    const [note, setNote] = useState('');
    const [files, setFiles] = useState([]);
    const [isResolving, setIsResolving] = useState(false); // Bật chế độ viết Post-Mortem
    const [resolutionSummary, setResolutionSummary] = useState('');
    const fileInputRef = useRef(null);

    const handleFileSelect = (e) => {
        if (e.target.files) setFiles([...files, ...Array.from(e.target.files)]);
    };
    const removeFile = (index) => setFiles(files.filter((_, i) => i !== index));

    const handleSubmit = (type) => {
        // Gửi kèm resolutionSummary nếu là chế độ Đóng Case
        onAction(type, type === 'RESOLVE' ? `[BÁO CÁO HẬU KIỂM]\n${resolutionSummary}` : note, files, resolutionSummary);
        if (type === 'COMMENT') { setNote(''); setFiles([]); }
        setIsResolving(false);
    };

    // NẾU ĐANG BẬT CHẾ ĐỘ ĐÓNG CASE (VIẾT REPORT)
    if (isResolving) {
        return (
            <div className="p-5 bg-emerald-950/20 border-t border-emerald-900/50 shrink-0 animate-in slide-in-from-bottom-2">
                <div className="flex items-center gap-2 mb-3 text-emerald-400 font-bold text-sm">
                    <ShieldCheck size={18}/> Báo cáo Tổng kết Sự cố (Post-Mortem)
                </div>
                <p className="text-xs text-slate-400 mb-3">Vui lòng tóm tắt nguyên nhân gốc rễ (Root Cause) và các bước đã khắc phục để lưu vào hồ sơ lưu trữ.</p>
                <textarea 
                    value={resolutionSummary}
                    onChange={(e) => setResolutionSummary(e.target.value)}
                    placeholder="- Nguyên nhân: Người dùng tải phần mềm lậu.&#10;- Khắc phục: Đã xóa file, quét lại toàn bộ ổ C..."
                    className="w-full bg-[#0A101D] text-emerald-100 text-sm p-4 rounded-xl border border-emerald-900/50 focus:border-emerald-500 outline-none resize-none h-32 mb-4"
                />
                <div className="flex justify-end gap-3">
                    <button onClick={() => setIsResolving(false)} className="px-4 py-2 text-slate-400 hover:text-white text-xs font-bold transition">Hủy đóng</button>
                    <button 
                        onClick={() => handleSubmit('RESOLVE')}
                        disabled={resolutionSummary.trim().length < 10 || sending}
                        className="px-6 py-2 bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-bold rounded-xl flex items-center gap-2 shadow-lg disabled:opacity-50"
                    >
                        {sending ? <Loader2 size={16} className="animate-spin"/> : <CheckCircle size={16}/>}
                        Xác nhận Đóng Hồ Sơ
                    </button>
                </div>
            </div>
        );
    }

    // CHẾ ĐỘ CHAT/LOG BÌNH THƯỜNG
    return (
        <div className="p-5 bg-[#0A101D] border-t border-slate-800 shrink-0">
            {files.length > 0 && (
                <div className="flex gap-2 mb-3 overflow-x-auto pb-2">
                    {files.map((file, idx) => (
                        <div key={idx} className="relative group shrink-0 w-16 h-16 rounded-lg bg-slate-800 border border-slate-700 flex items-center justify-center overflow-hidden">
                            {file.type.startsWith('image/') ? <img src={URL.createObjectURL(file)} alt="preview" className="w-full h-full object-cover"/> : <span className="text-[10px] text-slate-500 font-mono p-1 break-all">{file.name}</span>}
                            <button onClick={() => removeFile(idx)} className="absolute -top-1.5 -right-1.5 bg-red-500 text-white rounded-full p-0.5 shadow-md opacity-0 group-hover:opacity-100 transition"><X size={10}/></button>
                        </div>
                    ))}
                </div>
            )}

            <div className="relative">
                <textarea 
                    value={note} onChange={(e) => setNote(e.target.value)} disabled={sending}
                    placeholder="Ghi chú điều tra hoặc phân tích..."
                    className="w-full bg-[#111827] text-slate-200 text-sm p-4 pr-12 rounded-xl border border-slate-700 focus:border-indigo-500 outline-none resize-none h-20 mb-4 shadow-inner"
                />
                <button onClick={() => fileInputRef.current?.click()} className="absolute top-3 right-3 p-1.5 text-slate-400 hover:text-white hover:bg-slate-700 rounded-lg transition"><Paperclip size={18}/></button>
                <input type="file" multiple ref={fileInputRef} className="hidden" onChange={handleFileSelect} accept="image/*,.pdf,.txt" />
            </div>
            
            <div className="flex justify-between items-center">
                <div className="flex gap-3">
                    {status !== 'Resolved' && (
                        <button onClick={() => setIsResolving(true)} disabled={sending} className="px-4 py-2 bg-emerald-600/10 text-emerald-500 text-xs font-bold rounded-xl flex items-center gap-2 border border-emerald-600/30 hover:bg-emerald-600 hover:text-white transition">
                            <CheckCircle size={16}/> Đóng Case (Post-Mortem)
                        </button>
                    )}
                    {status === 'Open' && (
                        <button onClick={() => handleSubmit('INVESTIGATE')} disabled={sending} className="px-4 py-2 bg-blue-600/10 text-blue-500 text-xs font-bold rounded-xl flex items-center gap-2 border border-blue-600/30 hover:bg-blue-600 hover:text-white transition">
                            <FileText size={16}/> Bắt đầu Điều tra
                        </button>
                    )}
                </div>
                <button onClick={() => handleSubmit('COMMENT')} disabled={(!note.trim() && files.length === 0) || sending} className="px-6 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold rounded-xl flex items-center gap-2 shadow-lg disabled:opacity-50">
                    {sending ? <Loader2 size={16} className="animate-spin"/> : <Send size={16}/>} Gửi
                </button>
            </div>
        </div>
    );
};

export default IncidentActionBox;