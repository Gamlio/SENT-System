import React, { useState } from 'react';
import { 
    FileText, UploadCloud, Trash2, Edit, CheckCircle, 
    Clock, RefreshCw, Eye, Download, Search, X, ShieldAlert 
} from 'lucide-react';
import { useDocuments } from './hooks/useDocuments';
import { useAuth } from '../../context/AuthContext';
import Pagination from '../../components/common/Pagination';

const Documents = () => {
    const { user } = useAuth(); // Lấy thông tin user từ Context[cite: 44]
    
    // Kiểm tra quyền quản lý tài liệu[cite: 44]
    const canManageDocs = user?.permissions?.doc_manage === true;

    const {
        currentDocuments, filteredDocs, 
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages,
        isUploading, editingDoc, setEditingDoc,
        uploadDoc, deleteDoc, updateDoc
    } = useDocuments();

    const [file, setFile] = useState(null);
    const [title, setTitle] = useState('');
    const [category, setCategory] = useState('Internal');
    const [viewingDoc, setViewingDoc] = useState(null);

    const handleUploadSubmit = async (e) => {
        e.preventDefault();
        if (!file) return alert("Vui lòng chọn file!");
        
        const formData = new FormData();
        formData.append('file', file);
        formData.append('title', title);
        formData.append('category', category);

        const res = await uploadDoc(formData);
        if (res.success) {
            alert("Tải lên thành công! Tài liệu đang chờ phê duyệt.");
            setTitle(''); setFile(null);
        } else {
            alert(res.error);
        }
    };

    const handleViewDocument = (doc) => setViewingDoc(doc);

    const handleDownloadDocument = (filePath, fileName) => {
        const formattedPath = filePath.replace(/\\/g, '/');
        const fileUrl = `http://localhost:8000/${formattedPath}`;
        const link = document.createElement('a');
        link.href = fileUrl;
        link.setAttribute('download', fileName || 'Tai_Lieu.pdf'); 
        link.setAttribute('target', '_blank'); 
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    };

    return (
        <div className="text-slate-200 relative">
            <header className="mb-8">
                <h1 className="text-3xl font-bold text-white">Quản lý Chính sách & Tài liệu</h1>
                <p className="text-slate-400 text-sm">Nạp dữ liệu pháp lý và quy định nội bộ cho AI Security Agent</p>
            </header>

            <div className="grid grid-cols-1 xl:grid-cols-3 gap-8">
                {/* --- CỘT TRÁI: FORM UPLOAD (Chỉ hiện nếu có quyền manage) --- */}
                {canManageDocs ? (
                    <div className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl h-fit">
                        <h3 className="text-lg font-bold mb-6 flex items-center gap-2">
                            <UploadCloud className="text-emerald-400" size={20}/> Tải lên tài liệu
                        </h3>
                        <form onSubmit={handleUploadSubmit} className="space-y-4">
                            <div>
                                <label className="text-xs font-bold text-slate-500 uppercase">Tiêu đề</label>
                                <input type="text" value={title} onChange={(e) => setTitle(e.target.value)} required
                                    className="w-full mt-2 p-3 bg-slate-900/50 border border-slate-700 rounded-xl outline-none focus:border-emerald-500 text-white" />
                            </div>
                            <div>
                                <label className="text-xs font-bold text-slate-500 uppercase">Phân loại</label>
                                <select value={category} onChange={(e) => setCategory(e.target.value)}
                                    className="w-full mt-2 p-3 bg-slate-900/50 border border-slate-700 rounded-xl outline-none focus:border-emerald-500 text-white">
                                    <option value="Internal">Quy định nội bộ</option>
                                    <option value="ISO27001">Tiêu chuẩn ISO 27001</option>
                                    <option value="Law">Luật An ninh mạng</option>
                                </select>
                            </div>
                            <div className="border-2 border-dashed border-slate-700 rounded-2xl p-6 text-center relative hover:border-emerald-500/50 cursor-pointer transition">
                                <input type="file" onChange={(e) => setFile(e.target.files[0])} 
                                    className="absolute inset-0 opacity-0 cursor-pointer" accept=".pdf,.doc,.docx" />
                                <FileText className="mx-auto text-slate-500 mb-2" size={32}/>
                                <p className="text-sm text-slate-400">{file ? file.name : "Kéo thả hoặc chọn file PDF/Word"}</p>
                            </div>
                            <button type="submit" disabled={isUploading}
                                className="w-full bg-emerald-500 hover:bg-emerald-600 text-white font-bold py-3 rounded-xl transition shadow-lg shadow-emerald-500/20 disabled:opacity-50">
                                {isUploading ? "Đang xử lý..." : "Bắt đầu nạp tài liệu"}
                            </button>
                        </form>
                    </div>
                ) : (
                    <div className="bg-slate-900/30 p-8 rounded-3xl border border-slate-800/50 border-dashed text-center h-fit">
                        <ShieldAlert className="mx-auto text-slate-700 mb-4" size={48}/>
                        <p className="text-slate-500 text-sm italic font-medium">
                            Bạn chỉ có quyền xem tài liệu. <br/> Vui lòng liên hệ Admin để được cấp quyền quản lý.
                        </p>
                    </div>
                )}

                {/* --- CỘT PHẢI: DANH SÁCH TÀI LIỆU --- */}
                <div className="xl:col-span-2 flex flex-col h-full">
                    <div className="mb-6 relative">
                        <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500" size={20}/>
                        <input 
                            type="text" 
                            placeholder="Tìm kiếm theo tiêu đề hoặc phân loại..." 
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="w-full pl-12 pr-4 py-4 bg-[#1e293b] border border-slate-800 rounded-2xl text-white outline-none focus:border-emerald-500 transition shadow-lg"
                        />
                    </div>

                    <div className="space-y-4 flex-1">
                        {currentDocuments.map((doc) => (
                            <div key={doc.ID} className="bg-[#1e293b] p-5 rounded-2xl border border-slate-800 flex justify-between items-center group hover:border-slate-700 transition">
                                <div className="flex gap-4 items-center">
                                    <div className="p-3 bg-slate-900 rounded-xl text-emerald-400"><FileText size={24}/></div>
                                    <div>
                                        <h4 className="font-bold text-white truncate max-w-[200px] sm:max-w-xs">{doc.title}</h4>
                                        <div className="flex items-center gap-2 mt-1">
                                            <span className="text-[10px] px-2 py-0.5 bg-slate-800 text-slate-400 rounded uppercase font-bold">{doc.category}</span>
                                            <span className="text-[10px] text-slate-500 font-medium"><Clock size={10} className="inline mr-1"/>{new Date(doc.CreatedAt).toLocaleDateString()}</span>
                                        </div>
                                    </div>
                                </div>
                                <div className="flex items-center gap-4 sm:gap-6">
                                    {/* Trạng thái xử lý AI */}
                                    {doc.is_processed ? (
                                        <span className="hidden sm:flex items-center gap-1 text-[10px] font-bold text-emerald-400 bg-emerald-500/10 px-3 py-1 rounded-full border border-emerald-500/20"><CheckCircle size={12}/> AI READY</span>
                                    ) : (
                                        <span className="hidden sm:flex items-center gap-1 text-[10px] font-bold text-amber-400 bg-amber-500/10 px-3 py-1 rounded-full border border-amber-500/20"><RefreshCw size={12} className="animate-spin"/> PROCESSING</span>
                                    )}
                                    
                                    <div className="flex gap-2 opacity-100 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity">
                                        <button onClick={() => handleViewDocument(doc)} className="p-2 bg-slate-800 text-emerald-400 hover:bg-emerald-500/20 rounded-lg transition" title="Xem trên màn hình"><Eye size={16}/></button>
                                        <button onClick={() => handleDownloadDocument(doc.file_path, doc.file_name)} className="p-2 bg-slate-800 text-indigo-400 hover:bg-indigo-500/20 rounded-lg transition" title="Tải xuống máy"><Download size={16}/></button>
                                        
                                        {/* Nút hành động chỉ dành cho quyền Manage */}
                                        {canManageDocs && (
                                            <>
                                                <button onClick={() => setEditingDoc(doc)} className="p-2 bg-slate-800 text-blue-400 hover:bg-blue-500/20 rounded-lg transition" title="Sửa thông tin"><Edit size={16}/></button>
                                                <button onClick={() => deleteDoc(doc.ID)} className="p-2 bg-slate-800 text-red-400 hover:bg-red-500/20 rounded-lg transition" title="Xóa tài liệu"><Trash2 size={16}/></button>
                                            </>
                                        )}
                                    </div>
                                </div>
                            </div>
                        ))}
                        {filteredDocs.length === 0 && (
                            <div className="text-center py-20 text-slate-600 border-2 border-dashed border-slate-800 rounded-3xl font-medium italic">Không tìm thấy tài liệu nào phù hợp.</div>
                        )}
                    </div>

                    <div className="px-4 pb-4">
                        <Pagination 
                            currentPage={currentPage}
                            totalPages={totalPages}
                            onPageChange={setCurrentPage}
                        />
                    </div>
                </div>
            </div>

            {/* Modal Xem Tài Liệu */}
            {viewingDoc && (
                <div className="fixed inset-0 bg-black/80 backdrop-blur-md flex items-center justify-center z-[9999] p-6 animate-in fade-in">
                    <div className="bg-[#1e293b] rounded-3xl border border-slate-700 w-full max-w-5xl h-[90vh] flex flex-col shadow-2xl overflow-hidden">
                        <div className="p-5 border-b border-slate-700 flex justify-between items-center bg-slate-800/50">
                            <div className="flex items-center gap-3">
                                <div className="p-2 bg-emerald-500/10 rounded-lg text-emerald-400"><Eye size={20}/></div>
                                <div>
                                    <h2 className="text-lg font-bold text-white leading-tight">{viewingDoc.title}</h2>
                                    <p className="text-xs text-slate-400">{viewingDoc.file_name}</p>
                                </div>
                            </div>
                            <button onClick={() => setViewingDoc(null)} className="p-2 bg-slate-700 hover:bg-red-500 text-white rounded-xl transition">
                                <X size={20}/>
                            </button>
                        </div>
                        
                        <div className="flex-1 bg-slate-100 overflow-hidden">
                            {viewingDoc.file_name.toLowerCase().endsWith('.pdf') ? (
                                <iframe
                                    src={`http://localhost:8000/${viewingDoc.file_path.replace(/\\/g, '/')}`}
                                    className="w-full h-full border-none shadow-inner"
                                    title="Document Viewer"
                                />
                            ) : (
                                <div className="flex flex-col items-center justify-center h-full text-slate-800">
                                    <div className="p-6 bg-white rounded-3xl shadow-xl flex flex-col items-center max-w-sm text-center">
                                        <FileText size={64} className="text-blue-500 mb-4 opacity-20"/>
                                        <h3 className="font-bold text-lg mb-2">Định dạng không hỗ trợ xem trực tiếp</h3>
                                        <p className="text-slate-500 text-sm mb-6 px-4">Hiện tại chỉ hỗ trợ xem file PDF trực tiếp trên trình duyệt. Vui lòng tải về máy để xem nội dung.</p>
                                        <button 
                                            onClick={() => handleDownloadDocument(viewingDoc.file_path, viewingDoc.file_name)}
                                            className="w-full bg-emerald-500 hover:bg-emerald-600 text-white px-6 py-3 rounded-xl font-bold flex items-center justify-center gap-2 transition"
                                        >
                                            <Download size={18}/> Tải về máy chủ ngay
                                        </button>
                                    </div>
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            )}

            {/* Modal Sửa Tài Liệu */}
            {editingDoc && (
                <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-[9999] p-4 animate-in zoom-in-95">
                    <div className="bg-[#1e293b] p-8 rounded-3xl border border-slate-700 w-full max-w-md shadow-2xl relative">
                        <button onClick={() => setEditingDoc(null)} className="absolute right-4 top-4 text-slate-500 hover:text-white transition"><X size={20}/></button>
                        <h2 className="text-xl font-bold text-white mb-2 flex items-center gap-2">
                            <Edit className="text-blue-400" size={24}/> Cập nhật thông tin
                        </h2>
                        <p className="text-slate-400 text-xs mb-6 font-medium">Chỉ sửa đổi thông tin mô tả, nội dung file sẽ được giữ nguyên.</p>
                        <form onSubmit={(e) => { e.preventDefault(); updateDoc(editingDoc.ID, {title: editingDoc.title, category: editingDoc.category}); setEditingDoc(null); }} className="space-y-5">
                            <div>
                                <label className="text-xs font-black text-slate-500 uppercase tracking-widest pl-1">Tiêu đề tài liệu</label>
                                <input type="text" value={editingDoc.title} onChange={(e) => setEditingDoc({...editingDoc, title: e.target.value})} className="w-full mt-2 p-3.5 bg-slate-900 border border-slate-700 rounded-xl outline-none focus:border-blue-500 text-white font-medium" />
                            </div>
                            <div>
                                <label className="text-xs font-black text-slate-500 uppercase tracking-widest pl-1">Phân loại nhóm</label>
                                <select value={editingDoc.category} onChange={(e) => setEditingDoc({...editingDoc, category: e.target.value})} className="w-full mt-2 p-3.5 bg-slate-900 border border-slate-700 rounded-xl outline-none focus:border-blue-500 text-white font-medium">
                                    <option value="Internal">Quy định nội bộ</option>
                                    <option value="ISO27001">Tiêu chuẩn ISO 27001</option>
                                    <option value="Law">Luật An ninh mạng</option>
                                </select>
                            </div>
                            <div className="flex gap-4 pt-4">
                                <button type="button" onClick={() => setEditingDoc(null)} className="flex-1 py-3 text-slate-400 bg-slate-800 hover:bg-slate-700 rounded-xl font-bold transition">Hủy bỏ</button>
                                <button type="submit" className="flex-1 py-3 text-white bg-blue-600 hover:bg-blue-500 rounded-xl font-bold transition shadow-lg shadow-blue-600/20">Lưu thay đổi</button>
                            </div>
                        </form>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Documents;