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
        'APPROVED': { color: '#10b981', bg: 'bg-emerald-500/10', icon: <ShieldCheck size={11}/>, label: 'Đã Duyệt' },
        'PENDING': { color: '#f59e0b', bg: 'bg-orange-500/10', icon: <Clock size={11}/>, label: 'Chờ Duyệt' },
        'REJECTED': { color: '#ef4444', bg: 'bg-red-500/10', icon: <ShieldAlert size={11}/>, label: 'Bị Từ Chối' }
    };
    const s = config[status] || config['PENDING'];
    return (
        <div className={`flex items-center gap-1 px-2 py-0.5 rounded-full ${s.bg} border border-${s.color}/20 animate-in fade-in duration-500`}>
            <span style={{ color: s.color }}>{s.icon}</span>
            <span className="text-[9px] font-black uppercase tracking-wider" style={{ color: s.color }}>{s.label}</span>
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
        { label: 'Tổng tài liệu', value: filteredDocs.length, color: 'text-indigo-400', icon: <FileText size={18}/> },
        { label: 'Chờ phê duyệt', value: filteredDocs.filter(d => d.approval_status === 'PENDING').length, color: 'text-orange-400', icon: <Clock size={18}/> },
        { label: 'Đã ban hành', value: filteredDocs.filter(d => d.approval_status === 'APPROVED').length, color: 'text-emerald-400', icon: <ShieldCheck size={18}/> }
    ], [filteredDocs]);

    const handleUploadSubmit = async (e) => {
        e.preventDefault();
        if (!file) return;
        const formData = new FormData();
        formData.append('file', file);
        formData.append('title', title);
        formData.append('category', category);
        const res = await uploadDoc(formData);
        if (res.success) {
            setIsSidePanelOpen(false);
            setFile(null); setTitle('');
        }
    };

    return (
        <div className="flex flex-col h-full bg-[#050B14] text-slate-300 font-sans p-4 overflow-hidden animate-in fade-in duration-700">
            
            {/* HEADER & ACTION BAR */}
            <div className="flex justify-between items-end mb-4">
                <div>
                    <h2 className="text-xl font-black text-white tracking-tight flex items-center gap-2">
                        <div className="p-1.5 bg-indigo-500/10 rounded-lg text-indigo-400"><FileSearch size={20}/></div>
                        TRUNG TÂM TÀI LIỆU
                    </h2>
                    <p className="text-[10px] text-slate-500 font-bold uppercase tracking-[0.15em] mt-0.5 ml-1">
                        Quản lý văn bản & Quy trình bảo mật nội bộ
                    </p>
                </div>
                {canManageDocs && (
                    <button 
                        onClick={() => setIsSidePanelOpen(true)}
                        className="flex items-center gap-1.5 bg-indigo-600 hover:bg-indigo-500 text-white px-4 py-2 rounded-xl font-black text-[10px] uppercase tracking-widest transition-all shadow-lg shadow-indigo-600/20 active:scale-95"
                    >
                        <Plus size={14}/> Tải lên mới
                    </button>
                )}
            </div>

            {/* QUICK STATS GRID */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
                {stats.map((s, i) => (
                    <div key={i} className="bg-[#0A101D] border border-slate-800 p-3.5 rounded-2xl flex items-center gap-4 transition-transform hover:scale-[1.01]">
                        <div className={`p-2.5 rounded-xl bg-slate-900 ${s.color}`}>{s.icon}</div>
                        <div>
                            <p className="text-[9px] font-black uppercase text-slate-500 tracking-widest">{s.label}</p>
                            <h3 className="text-xl font-black text-white mt-0.5">{s.value}</h3>
                        </div>
                    </div>
                ))}
            </div>

            {/* CONTROLS AREA */}
            <div className="bg-[#0A101D] border border-slate-800 rounded-2xl p-1.5 mb-4 flex flex-wrap items-center gap-3">
                <div className="relative flex-1 min-w-[250px]">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-600" size={16} />
                    <input 
                        type="text"
                        placeholder="Tìm kiếm theo tiêu đề hoặc danh mục..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-full bg-[#050B14] border border-transparent focus:border-indigo-500/50 rounded-xl py-2 pl-10 pr-4 text-xs outline-none transition-all placeholder:text-slate-700"
                    />
                </div>
                <div className="flex items-center gap-1.5 px-2">
                    <Filter size={14} className="text-slate-500" />
                    <span className="text-[9px] font-black uppercase text-slate-600">Lọc nhanh:</span>
                    {['Internal', 'ISO27001', 'Law'].map(cat => (
                        <button key={cat} onClick={() => setSearchQuery(cat)} className="px-2.5 py-1 rounded-md bg-slate-900 hover:bg-slate-800 text-[9px] font-bold transition-colors uppercase italic">{cat}</button>
                    ))}
                </div>
            </div>

            {/* DOCUMENT LIST */}
            <div className="flex-1 overflow-y-auto pr-1 custom-scrollbar space-y-3">
                {currentDocuments.map((doc) => (
                    <div key={doc.ID} className="group relative bg-[#0A101D] border border-slate-800 p-3.5 rounded-2xl hover:border-indigo-500/50 transition-all flex items-center justify-between overflow-hidden">
                        <div className="flex items-center gap-4 flex-1 min-w-0">
                            <div className="p-3 bg-[#050B14] border border-slate-800 rounded-xl text-slate-500 group-hover:text-indigo-400 group-hover:border-indigo-500/30 transition-all shadow-inner">
                                <FileText size={20}/>
                            </div>
                            <div className="min-w-0">
                                <div className="flex items-center gap-2 mb-1">
                                    <h4 className="font-black text-white text-sm truncate tracking-tight">{doc.title}</h4>
                                    <DocStatusTag status={doc.approval_status} />
                                </div>
                                <div className="flex items-center gap-3 text-[9px] text-slate-500 font-bold uppercase tracking-wider">
                                    <span className="flex items-center gap-1 bg-slate-900 px-1.5 py-0.5 rounded text-indigo-400 border border-indigo-500/10">
                                        <Tag size={10}/> {doc.category}
                                    </span>
                                    <span className="flex items-center gap-1"><Clock size={10}/> {new Date(doc.CreatedAt).toLocaleDateString()}</span>
                                    <span className="flex items-center gap-1 italic underline opacity-60">@{doc.uploaded_by || 'system'}</span>
                                </div>
                            </div>
                        </div>

                        <div className="flex items-center gap-1.5 opacity-0 group-hover:opacity-100 transition-all transform translate-x-2 group-hover:translate-x-0">
                            {doc.display_pdf_path && (
                                <button 
                                    onClick={() => window.open(`/${doc.display_pdf_path}`)}
                                     className="p-2 bg-blue-500/10 text-blue-400 hover:bg-blue-500 rounded-lg hover:text-white transition-all shadow-sm" title="Xem nhanh">
                                    <Eye size={15}/>
                                </button>
                            )}
                            <button onClick={() => downloadDoc(doc.ID, doc.original_name)} className="p-2 bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500 rounded-lg hover:text-white transition-all shadow-sm" title="Tải file Word">
                                <Download size={15}/>
                            </button>
                            {canManageDocs && (
                                <>
                                    <button onClick={() => setEditingDoc(doc)} className="p-2 bg-slate-800 text-slate-400 hover:bg-slate-700 rounded-lg transition-all shadow-sm">
                                        <Edit size={15}/>
                                    </button>
                                    <button onClick={() => handleOpenDelete(doc)} className="p-2 bg-red-500/10 text-red-400 hover:bg-red-500 rounded-lg hover:text-white transition-all shadow-sm">
                                        <Trash2 size={15}/>
                                    </button>
                                </>
                            )}
                        </div>
                        <ChevronRight className="text-slate-800 group-hover:text-indigo-500/50 transition-colors ml-3" size={18}/>
                    </div>
                ))}
            </div>

            {/* FOOTER: PAGINATION */}
            <div className="mt-4 pt-4 border-t border-slate-800 flex justify-between items-center">
                <p className="text-[9px] font-black uppercase text-slate-600 tracking-widest">
                    Hiển thị <span className="text-white italic">{indexOfFirstItem + 1}-{Math.min(indexOfLastItem, filteredDocs.length)}</span> trên <span className="text-white">{filteredDocs.length}</span> tài liệu
                </p>
                <Pagination currentPage={currentPage} totalPages={totalPages} onPageChange={setCurrentPage} />
            </div>

            {/* RIGHT SIDE PANEL (UPLOAD/EDIT) */}
            {(isSidePanelOpen || editingDoc) && (
                <div className="fixed inset-0 z-[100] flex justify-end animate-in fade-in duration-300">
                    <div className="absolute inset-0 bg-[#050B14]/80 backdrop-blur-sm" onClick={() => {setIsSidePanelOpen(false); setEditingDoc(null)}} />
                    <div className="relative w-full max-w-sm bg-[#0A101D] border-l border-slate-800 shadow-2xl p-6 flex flex-col animate-in slide-in-from-right duration-500">
                        <div className="flex justify-between items-center mb-6">
                            <h3 className="text-lg font-black text-white uppercase tracking-tighter flex items-center gap-2">
                                <div className="p-1.5 bg-indigo-500/10 rounded-lg text-indigo-400">
                                    {editingDoc ? <Edit size={16}/> : <Plus size={16}/>}
                                </div>
                                {editingDoc ? 'Cập nhật tài liệu' : 'Tải lên tài liệu mới'}
                            </h3>
                            <button onClick={() => {setIsSidePanelOpen(false); setEditingDoc(null)}} className="p-1.5 hover:bg-slate-800 rounded-lg transition-colors"><X size={18}/></button>
                        </div>

                        <form onSubmit={handleUploadSubmit} className="space-y-4 flex-1 overflow-y-auto pr-1 custom-scrollbar">
                            <div className="space-y-1.5">
                                <label className="text-[9px] font-black text-slate-500 uppercase tracking-widest">Tiêu đề tài liệu</label>
                                <input 
                                    type="text" required 
                                    placeholder="Nhập tên tài liệu..."
                                    value={editingDoc ? editingDoc.title : title}
                                    onChange={(e) => editingDoc ? setEditingDoc({...editingDoc, title: e.target.value}) : setTitle(e.target.value)}
                                    className="w-full bg-[#050B14] border border-slate-800 rounded-xl px-4 py-3 text-xs focus:border-indigo-500 outline-none transition-all"
                                />
                            </div>

                            <div className="space-y-1.5">
                                <label className="text-[9px] font-black text-slate-500 uppercase tracking-widest">Phân loại danh mục</label>
                                <select 
                                    value={editingDoc ? editingDoc.category : category}
                                    onChange={(e) => editingDoc ? setEditingDoc({...editingDoc, category: e.target.value}) : setCategory(e.target.value)}
                                    className="w-full bg-[#050B14] border border-slate-800 rounded-xl px-4 py-3 text-xs focus:border-indigo-500 outline-none appearance-none cursor-pointer"
                                >
                                    <option value="Internal">Quy định nội bộ</option>
                                    <option value="ISO27001">Tiêu chuẩn ISO 27001</option>
                                    <option value="Law">Luật An ninh mạng</option>
                                </select>
                            </div>

                            {!editingDoc && (
                                <div className="space-y-1.5">
                                    <label className="text-[9px] font-black text-slate-500 uppercase tracking-widest text-indigo-400">Tệp tin đính kèm (Word only)</label>
                                    <label className="flex flex-col items-center justify-center w-full h-36 border-2 border-dashed border-slate-800 rounded-2xl bg-[#050B14] hover:bg-indigo-500/5 hover:border-indigo-500/30 transition-all cursor-pointer group">
                                        <div className="flex flex-col items-center justify-center pt-3 pb-4">
                                            <UploadCloud className="w-8 h-8 mb-2 text-slate-700 group-hover:text-indigo-500 transition-colors" />
                                            <p className="mb-1 text-[11px] font-bold text-slate-400 px-2 text-center truncate max-w-[200px]">{file ? file.name : "Kéo thả hoặc nhấn để chọn file"}</p>
                                            <p className="text-[8px] text-slate-600 font-black uppercase tracking-widest">Chỉ chấp nhận .doc, .docx</p>
                                        </div>
                                        <input type="file" accept=".doc,.docx" onChange={(e) => setFile(e.target.files[0])} className="hidden" />
                                    </label>
                                </div>
                            )}

                            <div className="bg-indigo-500/5 border border-indigo-500/10 rounded-xl p-3 flex gap-2">
                                <ShieldCheck className="text-indigo-400 shrink-0" size={16}/>
                                <p className="text-[9px] leading-relaxed italic text-slate-500">
                                    Hệ thống sẽ tự động chuyển đổi file của bạn sang định dạng PDF bảo mật và lưu vết người cập nhật để phục vụ công tác giám sát.
                                </p>
                            </div>
                        </form>

                        <div className="pt-4 mt-2 border-t border-slate-800 flex gap-3">
                            <button onClick={() => {setIsSidePanelOpen(false); setEditingDoc(null)}} className="flex-1 py-3 bg-slate-900 hover:bg-slate-800 text-[10px] font-black uppercase tracking-widest rounded-xl transition-all">Đóng</button>
                            <button 
                                onClick={editingDoc ? () => updateDoc(editingDoc.ID, editingDoc) : handleUploadSubmit}
                                className="flex-1 py-3 bg-indigo-600 hover:bg-indigo-500 text-white text-[10px] font-black uppercase tracking-widest rounded-xl transition-all shadow-xl shadow-indigo-600/20 active:scale-95"
                            >
                                {isUploading ? <RefreshCw className="animate-spin inline mr-1" size={12}/> : (editingDoc ? 'Lưu thay đổi' : 'Xác nhận')}
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