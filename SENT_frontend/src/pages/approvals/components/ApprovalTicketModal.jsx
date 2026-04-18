import React, { useState } from 'react';
import { X, Check, XCircle, FileText, Clock, User, Info, ArrowRight } from 'lucide-react';

const ApprovalTicketModal = ({ ticket, onClose, onReview }) => {
    const [reviewNote, setReviewNote] = useState('');
    const [isSubmitting, setIsSubmitting] = useState(false);

    if (!ticket) return null;

    let snapshot = {};
    try {
        snapshot = JSON.parse(ticket.snapshot_data || '{}');
    } catch (e) {
        console.error("Failed to parse snapshot data", e);
    }

    const handleAction = async (status) => {
        setIsSubmitting(true);
        await onReview(ticket.id, status, reviewNote);
        setIsSubmitting(false);
        onClose();
    };

    const renderSnapshotDetails = () => {
        if (!snapshot || Object.keys(snapshot).length === 0) {
            return <div className="text-slate-500 italic text-sm">Không có dữ liệu chi tiết.</div>;
        }

        switch (ticket.module_type) {
            case 'ASSET_ENROLL':
                return (
                    <div className="bg-slate-900 p-4 rounded-lg border border-slate-800">
                        <h4 className="text-sm font-bold text-white mb-2">Thông tin thiết bị:</h4>
                        <div className="grid grid-cols-2 gap-4 text-sm">
                            <div><span className="text-slate-500">Hostname:</span> <span className="text-slate-300 font-mono">{snapshot.hostname || 'N/A'}</span></div>
                            <div><span className="text-slate-500">IP Address:</span> <span className="text-slate-300 font-mono">{snapshot.ip_address || 'N/A'}</span></div>
                            <div className="col-span-2"><span className="text-slate-500">HWID:</span> <span className="text-slate-300 font-mono text-xs">{snapshot.hwid || ticket.target_name}</span></div>
                            {snapshot.is_re_enroll !== undefined && (
                                <div className="col-span-2">
                                    <span className="text-slate-500">Trạng thái:</span> 
                                    <span className={`ml-2 px-2 py-0.5 rounded text-xs ${snapshot.is_re_enroll ? 'bg-amber-500/20 text-amber-400' : 'bg-emerald-500/20 text-emerald-400'}`}>
                                        {snapshot.is_re_enroll ? 'Thiết bị cũ đăng ký lại' : 'Thiết bị mới hoàn toàn'}
                                    </span>
                                </div>
                            )}
                        </div>
                    </div>
                );
            case 'ASSET_BULK_DELETE':
            case 'ASSET_DELETE':
                return (
                    <div className="bg-slate-900 p-4 rounded-lg border border-slate-800">
                        <h4 className="text-sm font-bold text-red-400 mb-2">Danh sách tài sản bị ảnh hưởng:</h4>
                        {snapshot.hwids && snapshot.hwids.length > 0 ? (
                            <ul className="list-disc pl-5 text-slate-300 text-sm font-mono space-y-1">
                                {snapshot.hwids.map((hwid, idx) => (
                                    <li key={idx}>{hwid}</li>
                                ))}
                            </ul>
                        ) : (
                            <div className="text-slate-300 text-sm font-mono">{ticket.target_name}</div>
                        )}
                        {snapshot.reason && (
                            <div className="mt-4">
                                <span className="text-slate-500 text-sm">Lý do từ người yêu cầu:</span>
                                <div className="bg-slate-800/50 p-2 mt-1 rounded text-slate-300 text-sm">{snapshot.reason}</div>
                            </div>
                        )}
                    </div>
                );
            case 'USER_CREATE':
            case 'USER_UPDATE':
                // Render comparison or details
                return (
                    <div className="bg-slate-900 p-4 rounded-lg border border-slate-800">
                        <h4 className="text-sm font-bold text-white mb-3">Thông tin {ticket.module_type === 'USER_CREATE' ? 'tài khoản mới' : 'cập nhật'}:</h4>
                        <div className="grid grid-cols-2 gap-y-2 text-sm mb-4">
                            <div><span className="text-slate-500">Họ tên:</span> <span className="text-slate-300">{snapshot.full_name || snapshot.FullName}</span></div>
                            <div><span className="text-slate-500">Username:</span> <span className="text-slate-300 font-mono">{snapshot.username || snapshot.Username}</span></div>
                            <div><span className="text-slate-500">Email:</span> <span className="text-slate-300">{snapshot.email || snapshot.Email}</span></div>
                            <div><span className="text-slate-500">SĐT:</span> <span className="text-slate-300">{snapshot.phone || snapshot.Phone || 'N/A'}</span></div>
                        </div>
                        
                        <h4 className="text-sm font-bold text-white mt-4 mb-2">Quyền hạn mong muốn:</h4>
                        <div className="grid grid-cols-2 gap-2 text-xs">
                            {Object.entries(snapshot).filter(([k]) => k.startsWith('perm_') || k.startsWith('Perm')).map(([k, v]) => {
                                // Simple format
                                const label = k.replace(/perm_/i, '').replace(/_/g, ' ').toUpperCase();
                                return (
                                    <div key={k} className="flex items-center justify-between bg-slate-800/50 px-2 py-1 rounded">
                                        <span className="text-slate-400">{label}</span>
                                        {v ? <Check size={14} className="text-emerald-500" /> : <X size={14} className="text-red-500" />}
                                    </div>
                                )
                            })}
                        </div>
                    </div>
                );
            default:
                return (
                    <div className="bg-slate-900 p-4 rounded-lg border border-slate-800">
                        <pre className="text-xs text-slate-400 overflow-x-auto custom-scrollbar">
                            {JSON.stringify(snapshot, null, 2)}
                        </pre>
                    </div>
                );
        }
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-4">
            <div className="bg-[#0A101D] border border-slate-700 rounded-xl w-full max-w-2xl flex flex-col max-h-[90vh] shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-200">
                {/* Header */}
                <div className="px-6 py-4 border-b border-slate-800 flex justify-between items-center bg-slate-900/50">
                    <div className="flex items-center gap-3">
                        <div className="p-2 bg-indigo-500/20 text-indigo-400 rounded-lg">
                            <FileText size={20} />
                        </div>
                        <div>
                            <h2 className="text-lg font-black text-white tracking-tight">Chi tiết Đơn từ</h2>
                            <p className="text-[10px] text-slate-500 uppercase tracking-widest font-bold font-mono">ID: {ticket.id} | {ticket.module_type}</p>
                        </div>
                    </div>
                    <button onClick={onClose} className="p-2 text-slate-500 hover:text-white hover:bg-slate-800 rounded-lg transition-colors">
                        <X size={20} />
                    </button>
                </div>

                {/* Body */}
                <div className="p-6 overflow-y-auto custom-scrollbar flex-1 space-y-6">
                    {/* Meta Info */}
                    <div className="grid grid-cols-2 gap-4 bg-slate-800/30 p-4 rounded-xl border border-slate-800/50">
                        <div className="flex items-start gap-3">
                            <User size={16} className="text-slate-500 mt-0.5" />
                            <div>
                                <div className="text-[10px] text-slate-500 uppercase font-bold tracking-wider">Người làm đơn</div>
                                <div className="text-sm text-slate-200 font-medium">{ticket.requested_by || 'Hệ thống'}</div>
                            </div>
                        </div>
                        <div className="flex items-start gap-3">
                            <Clock size={16} className="text-slate-500 mt-0.5" />
                            <div>
                                <div className="text-[10px] text-slate-500 uppercase font-bold tracking-wider">Thời gian tạo</div>
                                <div className="text-sm text-slate-200">{new Date(ticket.created_at).toLocaleString('vi-VN')}</div>
                            </div>
                        </div>
                        <div className="col-span-2 flex items-start gap-3">
                            <Info size={16} className="text-slate-500 mt-0.5" />
                            <div>
                                <div className="text-[10px] text-slate-500 uppercase font-bold tracking-wider">Lý do / Mô tả</div>
                                <div className="text-sm text-slate-300 bg-slate-800/50 p-2 mt-1 rounded-md border border-slate-700/50">{ticket.request_reason || ticket.target_name}</div>
                            </div>
                        </div>
                    </div>

                    {/* Snapshot Data */}
                    <div>
                        <h3 className="text-xs font-bold text-slate-400 uppercase tracking-widest mb-3 flex items-center gap-2">
                            Nội dung chi tiết <ArrowRight size={12}/>
                        </h3>
                        {renderSnapshotDetails()}
                    </div>

                    {/* Audit / Review Note Input (Only if pending) */}
                    {ticket.status === 'PENDING' && (
                        <div>
                            <h3 className="text-xs font-bold text-slate-400 uppercase tracking-widest mb-2">Ghi chú phê duyệt (Bắt buộc nếu từ chối)</h3>
                            <textarea 
                                value={reviewNote}
                                onChange={(e) => setReviewNote(e.target.value)}
                                placeholder="Nhập lý do phê duyệt hoặc từ chối để lưu Audit Trail..."
                                className="w-full bg-slate-900 border border-slate-700 rounded-lg p-3 text-sm text-slate-200 placeholder-slate-600 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all min-h-[80px]"
                            />
                        </div>
                    )}

                    {/* Show existing review note if already processed */}
                    {ticket.status !== 'PENDING' && (
                        <div className={`p-4 rounded-xl border ${ticket.status === 'APPROVED' ? 'bg-emerald-900/10 border-emerald-500/20' : 'bg-red-900/10 border-red-500/20'}`}>
                            <div className="text-[10px] uppercase font-bold tracking-wider mb-1 flex justify-between">
                                <span className={ticket.status === 'APPROVED' ? 'text-emerald-500' : 'text-red-500'}>
                                    Đã {ticket.status === 'APPROVED' ? 'Duyệt' : 'Từ chối'} bởi {ticket.reviewed_by}
                                </span>
                            </div>
                            <div className="text-sm text-slate-300">
                                <span className="text-slate-500">Lý do/Ghi chú: </span>
                                {ticket.review_note || <span className="italic">Không có ghi chú</span>}
                            </div>
                        </div>
                    )}
                </div>

                {/* Footer Actions */}
                {ticket.status === 'PENDING' && (
                    <div className="px-6 py-4 border-t border-slate-800 bg-slate-900/80 flex justify-end gap-3">
                        <button 
                            disabled={isSubmitting}
                            onClick={() => handleAction('REJECTED')}
                            className="px-4 py-2 rounded-lg text-sm font-bold bg-slate-800 text-red-400 hover:bg-red-500 hover:text-white transition-colors border border-slate-700 hover:border-red-500 disabled:opacity-50"
                        >
                            Từ chối
                        </button>
                        <button 
                            disabled={isSubmitting}
                            onClick={() => handleAction('APPROVED')}
                            className="px-6 py-2 rounded-lg text-sm font-bold bg-emerald-600 text-white hover:bg-emerald-500 transition-colors shadow-lg shadow-emerald-500/20 disabled:opacity-50"
                        >
                            {isSubmitting ? 'Đang xử lý...' : 'Phê duyệt & Thực thi'}
                        </button>
                    </div>
                )}
            </div>
        </div>
    );
};

export default ApprovalTicketModal;