import React, { useState, useMemo } from 'react';
import { 
    FileText, UploadCloud, Trash2, Edit, CheckCircle2, 
    Clock, RefreshCw, Eye, Download, Search, X, ShieldAlert, Tag,
    ChevronRight, LayoutList, Filter, Plus, FileSearch, ShieldCheck, Globe
} from 'lucide-react';
import { useDocuments } from './hooks/useDocuments';
import { useAuth } from '../../context/AuthContext';
import Pagination from '../../components/common/Pagination';
import AppDialog from '../../components/AppDialog';

const DocStatusTag = ({ status }) => {
    const config = {
        'APPROVED': { color: '#10b981', bg: 'bg-emerald-500/10', icon: <ShieldCheck size={12}/>, label: 'Đã Duyệt' },
        'PENDING': { color: '#f59e0b', bg: 'bg-orange-500/10', icon: <Clock size={12}/>, label: 'Chờ Duyệt' },
        'REJECTED': { color: '#ef4444', bg: 'bg-red-500/10', icon: <ShieldAlert size={12}/>, label: 'Bị Từ Chối' }
    };
    const s = config[status] || config['PENDING'];
    return (
        <div className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full ${s.bg} border border-${s.color}/20 animate-in fade-in duration-500`}>
            <span style={{ color: s.color }}>{s.icon}</span>
            <span className="text-[10px] font-black uppercase tracking-widest" style={{ color: s.color }}>{s.label}</span>
        </div>
    );
};

const Documents = () => {
    const { user } = useAuth(); 
    const canManageDocs = user?.permissions?.doc_manage === true; 

    const {
        currentDocuments, filteredDocs, 
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages,
        isUploading, editingDoc, setEditingDoc,
        uploadDoc, deleteDoc, updateDoc, downloadDoc,
        indexOfFirstItem, 
        indexOfLastItem
    } = useDocuments();

    const [isSidePanelOpen, setIsSidePanelOpen] = useState(false);
    const [file, setFile] = useState(null);
    const [title, setTitle] = useState('');
    const [category, setCategory] = useState('Internal');

    const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
    const [targetDeleteDoc, setTargetDeleteDoc] = useState(null);
    const [deleteReason, setDeleteReason] = useState('');

    const handleOpenDelete = (doc) => {
        setTargetDeleteDoc(doc);
        setDeleteReason('');
        setIsDeleteModalOpen(true);
    };

    const handleConfirmDelete = async () => {
        if (!targetDeleteDoc) return;
        const res = await deleteDoc(targetDeleteDoc.ID, deleteReason);
        if (res && res.success) {
            setIsDeleteModalOpen(false);
            setTargetDeleteDoc(null);
            setDeleteReason('');
        } else if (res && res.error) {
            alert(res.error);
        }
    };

    const stats = useMemo(() => [
        { label: 'Tổng tài liệu', value: filteredDocs.length, color: 'text-indigo-400', icon: <FileText/> },
        { label: 'Chờ phê duyệt', value: filteredDocs.filter(d => d.approval_status === 'PENDING').length, color: 'text-orange-400', icon: <Clock/> },
        { label: 'Đã ban hành', value: filteredDocs.filter(d => d.approval_status === 'APPROVED').length, color: 'text-emerald-400', icon: <ShieldCheck/> }
    ], [filteredDocs]);

    const handleUploadSubmit = async (e) => {
        e.preventDefault();
        if (!file) return;
        const formData = new FormData();
        formData.append('file', file);
        formData.append('title', title);
        formData.append('category', category);
        const res = await uploadDoc(formData); //
        if (res.success) {
            setIsSidePanelOpen(false);
            setFile(null); setTitle('');
        }
    };

    return (
        <div className="flex flex-col h-full bg-[#050B14] text-slate-300 font-sans p-6 overflow-hidden animate-in fade-in duration-700">
            
            {/* HEADER & ACTION BAR */}
            <div className="flex justify-between items-end mb-8">
                <div>
                    <h2 className="text-2xl font-black text-white tracking-tight flex items-center gap-3">
                        <div className="p-2 bg-indigo-500/10 rounded-xl text-indigo-400"><FileSearch size={24}/></div>
                        TRUNG TÂM TÀI LIỆU
                    </h2>
                    <p className="text-[11px] text-slate-500 font-bold uppercase tracking-[0.2em] mt-1 ml-1">
                        Quản lý văn bản & Quy trình bảo mật nội bộ
                    </p>
                </div>
                {canManageDocs && (
                    <button 
                        onClick={() => setIsSidePanelOpen(true)}
                        className="flex items-center gap-2 bg-indigo-600 hover:bg-indigo-500 text-white px-6 py-3 rounded-2xl font-black text-[11px] uppercase tracking-widest transition-all shadow-lg shadow-indigo-600/20 active:scale-95"
                    >
                        <Plus size={16}/> Tải lên mới
                    </button>
                )}
            </div>

            {/* QUICK STATS GRID */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                {stats.map((s, i) => (
                    <div key={i} className="bg-[#0A101D] border border-slate-800 p-5 rounded-3xl flex items-center gap-5 transition-transform hover:scale-[1.02]">
                        <div className={`p-3 rounded-2xl bg-slate-900 ${s.color}`}>{s.icon}</div>
                        <div>
                            <p className="text-[10px] font-black uppercase text-slate-500 tracking-widest">{s.label}</p>
                            <h3 className="text-2xl font-black text-white mt-1">{s.value}</h3>
                        </div>
                    </div>
                ))}
            </div>

            {/* CONTROLS AREA */}
            <div className="bg-[#0A101D] border border-slate-800 rounded-3xl p-2 mb-6 flex flex-wrap items-center gap-4">
                <div className="relative flex-1 min-w-[300px]">
                    <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-600" size={18} />
                    <input 
                        type="text"
                        placeholder="Tìm kiếm theo tiêu đề hoặc danh mục..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)} //
                        className="w-full bg-[#050B14] border border-transparent focus:border-indigo-500/50 rounded-2xl py-3.5 pl-12 pr-4 text-sm outline-none transition-all placeholder:text-slate-700"
                    />
                </div>
                <div className="flex items-center gap-2 px-4">
                    <Filter size={16} className="text-slate-500" />
                    <span className="text-[10px] font-black uppercase text-slate-600">Lọc nhanh:</span>
                    {['Internal', 'ISO27001', 'Law'].map(cat => (
                        <button key={cat} onClick={() => setSearchQuery(cat)} className="px-3 py-1.5 rounded-lg bg-slate-900 hover:bg-slate-800 text-[10px] font-bold transition-colors uppercase italic">{cat}</button>
                    ))}
                </div>
            </div>

            <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar space-y-4">
                {currentDocuments.map((doc) => (
                    <div key={doc.ID} className="group relative bg-[#0A101D] border border-slate-800 p-5 rounded-3xl hover:border-indigo-500/50 transition-all flex items-center justify-between overflow-hidden">
                        <div className="flex items-center gap-5 flex-1 min-w-0">
                            <div className="p-4 bg-[#050B14] border border-slate-800 rounded-2xl text-slate-500 group-hover:text-indigo-400 group-hover:border-indigo-500/30 transition-all shadow-inner">
                                <FileText size={24}/>
                            </div>
                            <div className="min-w-0">
                                <div className="flex items-center gap-3 mb-1.5">
                                    <h4 className="font-black text-white text-base truncate tracking-tight">{doc.title}</h4>
                                    <DocStatusTag status={doc.approval_status} />
                                </div>
                                <div className="flex items-center gap-4 text-[10px] text-slate-500 font-bold uppercase tracking-wider">
                                    <span className="flex items-center gap-1.5 bg-slate-900 px-2 py-0.5 rounded text-indigo-400 border border-indigo-500/10">
                                        <Tag size={12}/> {doc.category}
                                    </span>
                                    <span className="flex items-center gap-1.5"><Clock size={12}/> {new Date(doc.CreatedAt).toLocaleDateString()}</span>
                                    <span className="flex items-center gap-1.5 italic underline opacity-60">@{doc.uploaded_by || 'system'}</span>
                                </div>
                            </div>
                        </div>

                        <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-all transform translate-x-4 group-hover:translate-x-0">
                            {doc.display_pdf_path && (
                                <button 
                                    onClick={() => window.open(`/${doc.display_pdf_path}`)}
                                     className="p-3 bg-blue-500/10 text-blue-400 hover:bg-blue-500 rounded-xl hover:text-white transition-all shadow-sm" title="Xem nhanh">
                                    <Eye size={18}/>
                                </button>
                            )}
                            <button onClick={() => downloadDoc(doc.ID, doc.original_name)} className="p-3 bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500 rounded-xl hover:text-white transition-all shadow-sm" title="Tải file Word">
                                <Download size={18}/>
                            </button>
                            {canManageDocs && (
                                <>
                                    <button onClick={() => setEditingDoc(doc)} className="p-3 bg-slate-800 text-slate-400 hover:bg-slate-700 rounded-xl transition-all shadow-sm">
                                        <Edit size={18}/>
                                    </button>
                                    <button onClick={() => handleOpenDelete(doc)} className="p-3 bg-red-500/10 text-red-400 hover:bg-red-500 rounded-xl hover:text-white transition-all shadow-sm">
                                        <Trash2 size={18}/>
                                    </button>
                                </>
                            )}
                        </div>
                        <ChevronRight className="text-slate-800 group-hover:text-indigo-500/50 transition-colors ml-4" size={24}/>
                    </div>
                ))}
            </div>

            {/* FOOTER: PAGINATION */}
            <div className="mt-6 pt-6 border-t border-slate-800 flex justify-between items-center">
                <p className="text-[10px] font-black uppercase text-slate-600 tracking-widest">
                    Hiển thị <span className="text-white italic">{indexOfFirstItem + 1}-{Math.min(indexOfLastItem, filteredDocs.length)}</span> trên <span className="text-white">{filteredDocs.length}</span> tài liệu
                </p>
                <Pagination currentPage={currentPage} totalPages={totalPages} onPageChange={setCurrentPage} />
            </div>

            {/* RIGHT SIDE PANEL (UPLOAD/EDIT) */}
            {(isSidePanelOpen || editingDoc) && (
                <div className="fixed inset-0 z-[100] flex justify-end animate-in fade-in duration-300">
                    <div className="absolute inset-0 bg-[#050B14]/80 backdrop-blur-sm" onClick={() => {setIsSidePanelOpen(false); setEditingDoc(null)}} />
                    <div className="relative w-full max-w-md bg-[#0A101D] border-l border-slate-800 shadow-2xl p-8 flex flex-col animate-in slide-in-from-right duration-500">
                        <div className="flex justify-between items-center mb-8">
                            <h3 className="text-xl font-black text-white uppercase tracking-tighter flex items-center gap-3">
                                <div className="p-2 bg-indigo-500/10 rounded-lg text-indigo-400">
                                    {editingDoc ? <Edit size={20}/> : <Plus size={20}/>}
                                </div>
                                {editingDoc ? 'Cập nhật tài liệu' : 'Tải lên tài liệu mới'}
                            </h3>
                            <button onClick={() => {setIsSidePanelOpen(false); setEditingDoc(null)}} className="p-2 hover:bg-slate-800 rounded-xl transition-colors"><X/></button>
                        </div>

                        <form onSubmit={handleUploadSubmit} className="space-y-6 flex-1 overflow-y-auto pr-2 custom-scrollbar">
                            <div className="space-y-2">
                                <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest">Tiêu đề tài liệu</label>
                                <input 
                                    type="text" required 
                                    placeholder="Nhập tên tài liệu..."
                                    value={editingDoc ? editingDoc.title : title}
                                    onChange={(e) => editingDoc ? setEditingDoc({...editingDoc, title: e.target.value}) : setTitle(e.target.value)}
                                    className="w-full bg-[#050B14] border border-slate-800 rounded-2xl px-5 py-4 text-sm focus:border-indigo-500 outline-none transition-all"
                                />
                            </div>

                            <div className="space-y-2">
                                <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest">Phân loại danh mục</label>
                                <select 
                                    value={editingDoc ? editingDoc.category : category}
                                    onChange={(e) => editingDoc ? setEditingDoc({...editingDoc, category: e.target.value}) : setCategory(e.target.value)}
                                    className="w-full bg-[#050B14] border border-slate-800 rounded-2xl px-5 py-4 text-sm focus:border-indigo-500 outline-none appearance-none cursor-pointer"
                                >
                                    <option value="Internal">Quy định nội bộ</option>
                                    <option value="ISO27001">Tiêu chuẩn ISO 27001</option>
                                    <option value="Law">Luật An ninh mạng</option>
                                </select>
                            </div>

                            {!editingDoc && (
                                <div className="space-y-2">
                                    <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest text-indigo-400">Tệp tin đính kèm (Word only)</label>
                                    <label className="flex flex-col items-center justify-center w-full h-48 border-2 border-dashed border-slate-800 rounded-3xl bg-[#050B14] hover:bg-indigo-500/5 hover:border-indigo-500/30 transition-all cursor-pointer group">
                                        <div className="flex flex-col items-center justify-center pt-5 pb-6">
                                            <UploadCloud className="w-12 h-12 mb-4 text-slate-700 group-hover:text-indigo-500 transition-colors" />
                                            <p className="mb-2 text-xs font-bold text-slate-400">{file ? file.name : "Kéo thả hoặc nhấn để chọn file"}</p>
                                            <p className="text-[9px] text-slate-600 font-black uppercase tracking-widest">Chỉ chấp nhận .doc, .docx</p>
                                        </div>
                                        <input type="file" accept=".doc,.docx" onChange={(e) => setFile(e.target.files[0])} className="hidden" />
                                    </label>
                                </div>
                            )}

                            <div className="bg-indigo-500/5 border border-indigo-500/10 rounded-2xl p-4 flex gap-3">
                                <ShieldCheck className="text-indigo-400 shrink-0" size={18}/>
                                <p className="text-[10px] leading-relaxed italic text-slate-500">
                                    Hệ thống sẽ tự động chuyển đổi file của bạn sang định dạng PDF bảo mật và lưu vết người cập nhật để phục vụ công tác giám sát.
                                </p>
                            </div>
                        </form>

                        <div className="pt-8 mt-4 border-t border-slate-800 flex gap-4">
                            <button onClick={() => {setIsSidePanelOpen(false); setEditingDoc(null)}} className="flex-1 py-4 bg-slate-900 hover:bg-slate-800 text-[11px] font-black uppercase tracking-widest rounded-2xl transition-all">Đóng</button>
                            <button 
                                onClick={editingDoc ? () => updateDoc(editingDoc.ID, editingDoc) : handleUploadSubmit}
                                className="flex-1 py-4 bg-indigo-600 hover:bg-indigo-500 text-white text-[11px] font-black uppercase tracking-widest rounded-2xl transition-all shadow-xl shadow-indigo-600/20 active:scale-95"
                            >
                                {isUploading ? <RefreshCw className="animate-spin inline mr-2" size={14}/> : (editingDoc ? 'Lưu thay đổi' : 'Xác nhận tải lên')}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            <AppDialog
                isOpen={isDeleteModalOpen}
                onClose={() => {
                    setIsDeleteModalOpen(false);
                    setTargetDeleteDoc(null);
                    setDeleteReason('');
                }}
                onConfirm={handleConfirmDelete}
                title="Xác nhận xóa tài liệu"
                message={`Bạn có chắc chắn muốn xóa tài liệu "${targetDeleteDoc?.title}"? Hành động này sẽ được ghi vào Audit Log.`}
                type="danger"
                confirmText="Xác nhận xóa"
                showInput={true}
                inputValue={deleteReason}
                onInputChange={setDeleteReason}
                inputPlaceholder="Nhập lý do xóa (bắt buộc)..."
            />
        </div>
    );
};

export default Documents;