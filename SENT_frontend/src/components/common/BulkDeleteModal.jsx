import React, { useState } from 'react';
import { X, ShieldCheck, Clock, User, FileText, CheckCircle2 } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';
import Pagination from './Pagination';

const BulkDeleteModal = ({
    isOpen,
    onClose,
    onConfirm,
    items,
    type // 'ASSET' or 'POLICY'
}) => {
    const { user } = useAuth();
    const [reason, setReason] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 5;

    if (!isOpen) return null;

    // Pagination logic
    const indexOfLastItem = currentPage * itemsPerPage;
    const indexOfFirstItem = indexOfLastItem - itemsPerPage;
    const currentItems = items.slice(indexOfFirstItem, indexOfLastItem);
    const totalPages = Math.ceil(items.length / itemsPerPage);

    const handleSubmit = (e) => {
        e.preventDefault();
        if (reason.trim().length < 5) {
            alert("Vui lòng nhập lý do hợp lệ (ít nhất 5 ký tự).");
            return;
        }
        onConfirm(reason);
    };

    return (
        <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-[100] flex items-center justify-center p-4">
            <div className="bg-[#0A101D] w-full max-w-3xl rounded-xl border border-slate-700 shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
                {/* Header */}
                <div className="p-4 border-b border-slate-800 flex justify-between items-center bg-[#111827]">
                    <h3 className="text-sm font-black uppercase tracking-widest text-white flex items-center gap-2">
                        <FileText size={16} className="text-red-400" />
                        Đơn yêu cầu xóa {type === 'ASSET' ? 'thiết bị' : 'chính sách'} hàng loạt
                    </h3>
                    <button onClick={onClose} className="text-slate-500 hover:text-white transition">
                        <X size={20} />
                    </button>
                </div>

                <div className="p-6 overflow-y-auto custom-scrollbar">
                    <form id="bulk-delete-form" onSubmit={handleSubmit} className="space-y-6">
                        
                        {/* Thông tin đơn */}
                        <div className="grid grid-cols-2 gap-4 bg-[#111827] p-4 rounded-xl border border-slate-800">
                            <div>
                                <label className="text-[10px] font-bold text-slate-500 uppercase mb-1 block">Tài khoản làm đơn</label>
                                <div className="text-sm text-white font-mono flex items-center gap-2">
                                    <User size={14} className="text-slate-400"/>
                                    {user?.email || user?.username || 'admin'}
                                </div>
                            </div>
                            <div>
                                <label className="text-[10px] font-bold text-slate-500 uppercase mb-1 block">Người làm đơn</label>
                                <div className="text-sm text-white font-medium">
                                    {user?.full_name || 'Quản trị viên'}
                                </div>
                            </div>
                            <div>
                                <label className="text-[10px] font-bold text-slate-500 uppercase mb-1 block">Thời gian tạo đơn</label>
                                <div className="text-sm text-slate-300 flex items-center gap-2">
                                    <Clock size={14} className="text-slate-400"/>
                                    {new Date().toLocaleString('vi-VN')}
                                </div>
                            </div>
                            <div>
                                <label className="text-[10px] font-bold text-slate-500 uppercase mb-1 block">Người phê duyệt</label>
                                <div className="text-sm text-emerald-400 flex items-center gap-2 font-bold">
                                    <ShieldCheck size={14} />
                                    Trung tâm Phê duyệt (SOC Admin)
                                </div>
                            </div>
                        </div>

                        {/* Danh sách mục cần xóa */}
                        <div>
                            <label className="text-xs font-bold text-white uppercase mb-2 flex items-center justify-between">
                                <span>Danh sách mục cần xóa ({items.length})</span>
                            </label>
                            <div className="bg-slate-900 border border-slate-700 rounded-xl overflow-hidden">
                                <table className="w-full text-left border-collapse text-xs">
                                    <thead>
                                        <tr className="bg-slate-800/50 border-b border-slate-700 text-slate-400">
                                            <th className="p-3">#</th>
                                            {type === 'ASSET' ? (
                                                <>
                                                    <th className="p-3">Hostname</th>
                                                    <th className="p-3">IP / MAC</th>
                                                    <th className="p-3">HWID</th>
                                                </>
                                            ) : (
                                                <>
                                                    <th className="p-3">Tên chính sách</th>
                                                    <th className="p-3">Loại</th>
                                                    <th className="p-3">Phạm vi</th>
                                                </>
                                            )}
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {currentItems.map((item, index) => (
                                            <tr key={index} className="border-b border-slate-800/50 hover:bg-slate-800/20 transition text-slate-300">
                                                <td className="p-3">{indexOfFirstItem + index + 1}</td>
                                                {type === 'ASSET' ? (
                                                    <>
                                                        <td className="p-3 font-medium text-white">{item.hostname || 'N/A'}</td>
                                                        <td className="p-3 font-mono text-slate-400">{item.ip_address || 'N/A'}</td>
                                                        <td className="p-3 font-mono text-[10px] truncate max-w-[150px]">{item.hwid || item.asset_hwid}</td>
                                                    </>
                                                ) : (
                                                    <>
                                                        <td className="p-3 font-medium text-white">{item.title}</td>
                                                        <td className="p-3">
                                                            <span className="px-2 py-0.5 bg-slate-800 rounded text-[10px] text-slate-300">
                                                                {item.policy_type}
                                                            </span>
                                                        </td>
                                                        <td className="p-3 font-mono text-[10px]">{item.target_type}</td>
                                                    </>
                                                )}
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                                {totalPages > 1 && (
                                    <div className="p-3 border-t border-slate-800 flex justify-center bg-slate-900">
                                        <Pagination currentPage={currentPage} totalPages={totalPages} onPageChange={setCurrentPage} />
                                    </div>
                                )}
                            </div>
                        </div>

                        {/* Lý do */}
                        <div>
                            <label className="text-xs font-bold text-white uppercase mb-2 block">Lý do yêu cầu xóa <span className="text-red-500">*</span></label>
                            <textarea
                                value={reason}
                                onChange={(e) => setReason(e.target.value)}
                                placeholder="Nhập lý do chi tiết..."
                                className="w-full p-3 bg-slate-900 border border-slate-700 rounded-xl text-sm text-white focus:border-red-500 focus:ring-1 focus:ring-red-500 outline-none transition resize-none h-24"
                                required
                            />
                        </div>
                    </form>
                </div>

                {/* Footer Actions */}
                <div className="p-4 border-t border-slate-800 flex justify-end gap-3 bg-[#111827]">
                    <button
                        type="button"
                        onClick={onClose}
                        className="px-5 py-2.5 rounded-xl font-bold text-xs uppercase tracking-wider text-slate-400 hover:text-white hover:bg-slate-800 transition"
                    >
                        Hủy
                    </button>
                    <button
                        type="submit"
                        form="bulk-delete-form"
                        className="flex items-center gap-2 bg-red-500 hover:bg-red-600 text-white px-5 py-2.5 rounded-xl transition font-bold text-xs uppercase tracking-wider shadow-lg shadow-red-500/20"
                    >
                        <CheckCircle2 size={16} />
                        Gửi Đơn Yêu Cầu Xóa
                    </button>
                </div>
            </div>
        </div>
    );
};

export default BulkDeleteModal;
