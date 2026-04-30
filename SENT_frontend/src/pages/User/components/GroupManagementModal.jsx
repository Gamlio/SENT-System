
import React, { useState, useEffect, useCallback } from 'react';
import { X, Layers, Plus, Edit, Trash2, Loader2, Save, ChevronLeft, ChevronRight } from 'lucide-react';
import axios from '../../../api/axios';

// Hook nội bộ để quản lý các API call liên quan đến Group
const useGroupApi = () => {
    const [groups, setGroups] = useState([]);
    const [pagination, setPagination] = useState({ page: 1, total: 0, limit: 5 });
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState(null);
    const [success, setSuccess] = useState(null);

    const fetchGroups = useCallback(async (page = 1) => {
        setIsLoading(true);
        setError(null);
        try {
            const res = await axios.get(`/groups?page=${page}&limit=${pagination.limit}`);
            setGroups(res.data.data || []);
            setPagination(prev => ({ ...prev, total: res.data.total, page: res.data.page }));
        } catch (err) {
            setError(err.response?.data?.error || "Không thể tải danh sách phòng ban.");
            console.error("Lỗi lấy groups:", err);
        } finally {
            setIsLoading(false);
        }
    }, [pagination.limit]);

    const createGroup = async (groupData) => {
        setIsLoading(true);
        setError(null);
        setSuccess(null);
        try {
            const res = await axios.post('/groups', groupData);
            setSuccess(res.data.message || "Thao tác thành công!");
            await fetchGroups(1); // Quay về trang đầu sau khi tạo
            return true;
        } catch (err) {
            setError(err.response?.data?.error || "Tạo phòng ban thất bại.");
            return false;
        } finally {
            setIsLoading(false);
        }
    };

    const updateGroup = async (id, groupData) => {
        setIsLoading(true);
        setError(null);
        setSuccess(null);
        try {
            const res = await axios.put(`/groups/${id}`, groupData);
            setSuccess(res.data.message || "Cập nhật thành công!");
            await fetchGroups(pagination.page);
            return true;
        } catch (err) {
            setError(err.response?.data?.error || "Cập nhật thất bại.");
            return false;
        } finally {
            setIsLoading(false);
        }
    };

    const deleteGroup = async (id, reason) => {
        setIsLoading(true);
        setError(null);
        setSuccess(null);
        try {
            const res = await axios.delete(`/groups/${id}`, { data: { reason } });
            setSuccess(res.data.message || "Yêu cầu xóa đã được gửi.");
            await fetchGroups(pagination.page);
            return true;
        } catch (err) {
            setError(err.response?.data?.error || "Yêu cầu xóa thất bại.");
            return false;
        } finally {
            setIsLoading(false);
        }
    };

    return {
        groups, pagination, isLoading, error, success,
        fetchGroups, createGroup, updateGroup, deleteGroup,
        setError, setSuccess
    };
};

const GroupManagementModal = ({ isOpen, onClose, currentUser }) => {
    const {
        groups, pagination, isLoading, error, success,
        fetchGroups, createGroup, updateGroup, deleteGroup,
        setError, setSuccess
    } = useGroupApi();

    const [isFormVisible, setIsFormVisible] = useState(false);
    const [editingGroup, setEditingGroup] = useState(null);
    const [deletingGroup, setDeletingGroup] = useState(null);
    const [deleteReason, setDeleteReason] = useState('');

    useEffect(() => {
        if (isOpen) {
            fetchGroups(1);
            setIsFormVisible(false);
            setEditingGroup(null);
            setDeletingGroup(null);
            setError(null);
            setSuccess(null);
        }
    }, [isOpen, fetchGroups]);

    useEffect(() => {
        if (error || success) {
            const timer = setTimeout(() => {
                setError(null);
                setSuccess(null);
            }, 4000);
            return () => clearTimeout(timer);
        }
    }, [error, success, setError, setSuccess]);

    const handleAddNew = () => {
        setEditingGroup({ name: '', description: '' });
        setIsFormVisible(true);
    };

    const handleEdit = (group) => {
        setEditingGroup(group);
        setIsFormVisible(true);
    };

    const handleDeleteRequest = (group) => {
        setDeletingGroup(group);
        setDeleteReason('');
    };

    const confirmDelete = async () => {
        if (!deleteReason.trim()) {
            setError("Vui lòng cung cấp lý do xóa.");
            return;
        }
        const isSuccess = await deleteGroup(deletingGroup.ID, deleteReason);
        if (isSuccess) {
            setDeletingGroup(null);
        }
    };

    const handleFormSubmit = async (e) => {
        e.preventDefault();
        const formData = { name: editingGroup.name, description: editingGroup.description };
        const isSuccess = editingGroup.ID ? await updateGroup(editingGroup.ID, formData) : await createGroup(formData);
        if (isSuccess) {
            setIsFormVisible(false);
            setEditingGroup(null);
        }
    };

    if (!isOpen) return null;

    const totalPages = Math.ceil(pagination.total / pagination.limit);

    return (
        <div className="fixed inset-0 bg-black bg-opacity-70 z-50 flex justify-center items-center animate-fade-in p-4">
            <div className="bg-[#161d2b] rounded-2xl shadow-2xl w-full max-w-3xl border border-slate-700 transform transition-all duration-300 scale-95 animate-modal-pop-in flex flex-col max-h-[90vh]">
                <div className="flex justify-between items-center p-5 border-b border-slate-800">
                    <h3 className="text-xl font-bold text-white flex items-center gap-3">
                        <Layers className="text-sky-400" />
                        Quản lý Phòng ban / Nhóm
                    </h3>
                    <button onClick={onClose} className="text-slate-400 hover:text-white transition-colors"><X size={24} /></button>
                </div>

                <div className="p-6 flex-1 overflow-y-auto">
                    {isFormVisible ? (
                        <div className="animate-in fade-in-50">
                            <h4 className="text-lg font-semibold text-emerald-400 mb-4">{editingGroup?.ID ? 'Chỉnh sửa Phòng ban' : 'Thêm Phòng ban mới'}</h4>
                            <form onSubmit={handleFormSubmit} className="space-y-4">
                                <div>
                                    <label className="block text-sm font-medium text-slate-300 mb-1">Tên phòng ban</label>
                                    <input type="text" value={editingGroup.name} onChange={(e) => setEditingGroup({ ...editingGroup, name: e.target.value })} required className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 text-sm text-white focus:border-emerald-500 outline-none" />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-slate-300 mb-1">Mô tả</label>
                                    <textarea value={editingGroup.description} onChange={(e) => setEditingGroup({ ...editingGroup, description: e.target.value })} rows="3" className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 text-sm text-white focus:border-emerald-500 outline-none"></textarea>
                                </div>
                                <div className="flex justify-end gap-3 pt-4">
                                    <button type="button" onClick={() => setIsFormVisible(false)} className="px-4 py-2 text-sm font-bold text-slate-300 rounded-lg hover:bg-slate-700">Hủy</button>
                                    <button type="submit" disabled={isLoading} className="flex items-center gap-2 bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2 rounded-lg text-sm font-bold shadow-lg transition disabled:bg-slate-500">
                                        {isLoading ? <Loader2 size={16} className="animate-spin" /> : <Save size={16} />} {editingGroup?.ID ? 'Lưu thay đổi' : 'Tạo mới'}
                                    </button>
                                </div>
                            </form>
                        </div>
                    ) : (
                        <div>
                            <div className="flex justify-between items-center mb-4">
                                <p className="text-sm text-slate-400">Danh sách các phòng ban hiện có.</p>
                                <button onClick={handleAddNew} className="flex items-center gap-2 bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2 rounded-lg text-sm font-bold shadow-lg transition"><Plus size={16} /> Thêm mới</button>
                            </div>
                            <div className="border border-slate-800 rounded-lg overflow-hidden">
                                <table className="w-full text-sm text-left text-slate-400">
                                    <thead className="text-xs text-slate-500 uppercase bg-slate-900/50"><tr><th scope="col" className="px-6 py-3">Tên phòng ban</th><th scope="col" className="px-6 py-3">Mô tả</th><th scope="col" className="px-6 py-3 text-right">Thao tác</th></tr></thead>
                                    <tbody>
                                        {isLoading && !groups.length ? (<tr><td colSpan="3" className="text-center p-6"><Loader2 className="mx-auto animate-spin text-slate-500" /></td></tr>) : groups.length > 0 ? (groups.map(group => (<tr key={group.ID} className="bg-slate-800/20 border-b border-slate-800 hover:bg-slate-800/50"><th scope="row" className="px-6 py-4 font-medium text-white whitespace-nowrap">{group.name}</th><td className="px-6 py-4">{group.description || <i className="text-slate-500">Không có mô tả</i>}</td><td className="px-6 py-4 text-right"><button onClick={() => handleEdit(group)} className="p-2 text-blue-400 hover:text-white"><Edit size={16} /></button><button onClick={() => handleDeleteRequest(group)} className="p-2 text-red-400 hover:text-white"><Trash2 size={16} /></button></td></tr>))) : (<tr><td colSpan="3" className="text-center p-6 text-slate-500">Chưa có phòng ban nào.</td></tr>)}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    )}
                    {deletingGroup && (<div className="fixed inset-0 bg-black/60 flex items-center justify-center z-[60]"><div className="bg-[#1e293b] p-6 rounded-lg shadow-xl border border-slate-700 w-full max-w-md"><h4 className="text-lg font-bold text-red-400">Xác nhận Xóa</h4><p className="text-sm text-slate-300 mt-2">Bạn sắp gửi yêu cầu xóa phòng ban <strong className="text-white">{deletingGroup.name}</strong>. Vui lòng cung cấp lý do.</p><textarea value={deleteReason} onChange={(e) => setDeleteReason(e.target.value)} placeholder="Nhập lý do xóa..." rows="3" className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2.5 text-sm text-white focus:border-red-500 outline-none mt-4"></textarea><div className="flex justify-end gap-3 mt-4"><button onClick={() => setDeletingGroup(null)} className="px-4 py-2 text-sm font-bold text-slate-300 rounded-lg hover:bg-slate-700">Hủy</button><button onClick={confirmDelete} disabled={isLoading || !deleteReason.trim()} className="flex items-center gap-2 bg-red-600 hover:bg-red-500 text-white px-4 py-2 rounded-lg text-sm font-bold shadow-lg transition disabled:opacity-50">{isLoading ? <Loader2 size={16} className="animate-spin" /> : <Trash2 size={16} />} Gửi yêu cầu</button></div></div></div>)}
                </div>

                <div className="flex justify-between items-center p-4 border-t border-slate-800 bg-slate-900/50 shrink-0">
                    <div className="text-sm h-5">{success && <p className="text-emerald-400 animate-in fade-in">{success}</p>}{error && <p className="text-red-400 animate-in fade-in">{error}</p>}</div>
                    {!isFormVisible && totalPages > 1 && (<div className="flex items-center gap-2 text-xs text-slate-400"><button onClick={() => fetchGroups(pagination.page - 1)} disabled={pagination.page <= 1 || isLoading} className="p-2 rounded hover:bg-slate-700 disabled:opacity-50"><ChevronLeft size={16} /></button><span>Trang {pagination.page} / {totalPages}</span><button onClick={() => fetchGroups(pagination.page + 1)} disabled={pagination.page >= totalPages || isLoading} className="p-2 rounded hover:bg-slate-700 disabled:opacity-50"><ChevronRight size={16} /></button></div>)}
                </div>
            </div>
        </div>
    );
};

export default GroupManagementModal;