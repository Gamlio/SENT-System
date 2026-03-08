import React, { useState, useRef } from 'react';
import { Send, CheckCircle, FileText, Loader2, Paperclip, X } from 'lucide-react';

const IncidentActionBox = ({ status, onAction, sending }) => {
    const [note, setNote] = useState('');
    const [files, setFiles] = useState([]); // State lưu file đã chọn
    const fileInputRef = useRef(null);

    const handleFileSelect = (e) => {
        if (e.target.files) {
            setFiles([...files, ...Array.from(e.target.files)]);
        }
    };

    const removeFile = (index) => {
        setFiles(files.filter((_, i) => i !== index));
    };

    const handleSubmit = (type) => {
        // Gửi cả text và files lên cha
        onAction(type, note, files);
        
        // Reset form sau khi gửi (nếu là comment)
        if (type === 'COMMENT') {
            setNote('');
            setFiles([]);
        }
    };

    return (
        <div className="p-5 bg-[#0f172a] border-t border-slate-800 shrink-0">
            {/* Vùng hiển thị file đã chọn */}
            {files.length > 0 && (
                <div className="flex gap-2 mb-3 overflow-x-auto pb-2">
                    {files.map((file, idx) => (
                        <div key={idx} className="relative group shrink-0">
                            <div className="w-16 h-16 rounded-lg bg-slate-800 border border-slate-700 flex items-center justify-center overflow-hidden">
                                {file.type.startsWith('image/') ? (
                                    <img src={URL.createObjectURL(file)} alt="preview" className="w-full h-full object-cover"/>
                                ) : (
                                    <span className="text-[10px] text-slate-500 font-mono p-1 break-all">{file.name}</span>
                                )}
                            </div>
                            <button 
                                onClick={() => removeFile(idx)}
                                className="absolute -top-1.5 -right-1.5 bg-red-500 text-white rounded-full p-0.5 shadow-md opacity-0 group-hover:opacity-100 transition"
                            >
                                <X size={10}/>
                            </button>
                        </div>
                    ))}
                </div>
            )}

            <div className="relative">
                <textarea 
                    value={note}
                    onChange={(e) => setNote(e.target.value)}
                    placeholder="Nhập ghi chú hoặc dán ảnh bằng chứng..."
                    disabled={sending}
                    className="w-full bg-[#1e293b] text-slate-200 text-sm p-4 pr-12 rounded-xl border border-slate-700 focus:border-indigo-500 outline-none resize-none h-24 mb-4 shadow-inner"
                />
                
                {/* Nút đính kèm file */}
                <button 
                    onClick={() => fileInputRef.current?.click()}
                    className="absolute top-3 right-3 p-1.5 text-slate-400 hover:text-white hover:bg-slate-700 rounded-lg transition"
                    title="Đính kèm ảnh/file"
                >
                    <Paperclip size={18}/>
                </button>
                <input 
                    type="file" 
                    multiple 
                    ref={fileInputRef} 
                    className="hidden" 
                    onChange={handleFileSelect}
                    accept="image/*,.pdf,.txt" // Giới hạn loại file
                />
            </div>
            
            <div className="flex justify-between items-center">
                {/* (Giữ nguyên các nút bấm cũ) */}
                <div className="flex gap-3">
                    {status !== 'Resolved' && (
                        <button onClick={() => handleSubmit('RESOLVE')} disabled={sending} className="px-4 py-2 bg-emerald-600/10 text-emerald-500 text-xs font-bold rounded-xl flex items-center gap-2 border border-emerald-600/30 hover:bg-emerald-600 hover:text-white transition">
                            <CheckCircle size={16}/> Xử lý xong
                        </button>
                    )}
                    {status === 'Open' && (
                        <button onClick={() => handleSubmit('INVESTIGATE')} disabled={sending} className="px-4 py-2 bg-blue-600/10 text-blue-500 text-xs font-bold rounded-xl flex items-center gap-2 border border-blue-600/30 hover:bg-blue-600 hover:text-white transition">
                            <FileText size={16}/> Điều tra
                        </button>
                    )}
                </div>

                <button 
                    onClick={() => handleSubmit('COMMENT')}
                    disabled={(!note.trim() && files.length === 0) || sending}
                    className="px-6 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold rounded-xl flex items-center gap-2 shadow-lg disabled:opacity-50"
                >
                    {sending ? <Loader2 size={16} className="animate-spin"/> : <Send size={16}/>}
                    Gửi Ghi chú
                </button>
            </div>
        </div>
    );
};

export default IncidentActionBox;