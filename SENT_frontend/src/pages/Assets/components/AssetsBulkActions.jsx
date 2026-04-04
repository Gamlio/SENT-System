import React, { useState } from 'react';
import { ShieldCheck, Trash2, X } from 'lucide-react';
import axios from '../../../api/axios';
import AppDialog from '../../../components/AppDialog'; // Đảm bảo đường dẫn đúng

const assetBulkActions = ({ 
    selectedassets, 
    clearSelection, 
    onRefresh, 
    onOpenApproveModal 
}) => {
    // Mang State của Dialog sang đây
    const [deleteReason, setDeleteReason] = useState('');
    const [dialogConfig, setDialogConfig] = useState({ isOpen: false });

    // Nếu không chọn máy nào thì ẩn component này đi
    if (selectedassets.length === 0) return null;

    const closeDialog = () => {
        setDialogConfig({ ...dialogConfig, isOpen: false });
        setDeleteReason('');
    };

    // Hàm gọi API xóa hàng loạt
    const handleBulkDeleteClick = () => {
        setDialogConfig({
            isOpen: true,
            title: `Yêu cầu gỡ bỏ ${selectedassets.length} thiết bị?`,
            message: `Hành động này sẽ ngắt kết nối và xóa dữ liệu của ${selectedassets.length} máy trạm. Vui lòng nhập lý do.`,
            type: 'danger',
            confirmText: 'Gửi yêu cầu xóa',
            showInput: true,
            onConfirm: async () => {
                if (deleteReason.trim().length < 5) {
                    alert("Vui lòng nhập lý do rõ ràng (ít nhất 5 ký tự).");
                    return;
                }
                closeDialog();
                try {
                    await axios.post('/assets/bulk-request-delete', {
                        hwids: selectedassets,
                        reason: deleteReason 
                    });
                    
                    clearSelection(); // Bỏ tick chọn toàn bộ
                    onRefresh();      // Load lại bảng
                    
                    setDialogConfig({
                        isOpen: true, 
                        title: 'Đã gửi yêu cầu', 
                        message: `Đơn xin gỡ bỏ ${selectedassets.length} thiết bị đã được chuyển đến Trung tâm Phê duyệt.`, 
                        type: 'success', 
                        isAlertOnly: true 
                    });
                } catch (err) {
                    setDialogConfig({
                        isOpen: true, 
                        title: 'Từ chối yêu cầu', 
                        message: err.response?.data?.error || 'Lỗi gửi yêu cầu xóa.', 
                        type: 'warning', 
                        isAlertOnly: true
                    });
                }
            }
        });
    };

    return (
        <>
            {/* Thanh công cụ nổi */}
            <div className="fixed bottom-10 left-1/2 -translate-x-1/2 bg-slate-800 text-white px-6 py-4 rounded-2xl shadow-2xl border border-emerald-500/50 flex items-center gap-6 z-50 animate-in slide-in-from-bottom-5 duration-300">
                <span className="text-sm font-bold">
                    <span className="bg-emerald-500 px-2.5 py-1 rounded-lg mr-2 text-xs font-black">
                        {selectedassets.length}
                    </span> 
                    máy đã chọn
                </span>
                
                <button 
                    onClick={onOpenApproveModal} 
                    className="flex items-center gap-2 bg-emerald-500 hover:bg-emerald-600 px-5 py-2.5 rounded-xl transition font-bold text-xs uppercase tracking-wider shadow-lg"
                >
                    <ShieldCheck size={16}/> Cấp phép hàng loạt
                </button>

                <button 
                    onClick={handleBulkDeleteClick} 
                    className="flex items-center gap-2 bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/30 px-5 py-2.5 rounded-xl transition font-bold text-xs uppercase tracking-wider shadow-lg"
                >
                    <Trash2 size={16}/> Gỡ bỏ ({selectedassets.length})
                </button>

                <button 
                    onClick={clearSelection} 
                    className="p-2 text-slate-400 hover:text-red-400 transition-colors"
                >
                    <X size={20}/>
                </button>
            </div>

            {/* Hộp thoại xác nhận */}
            <AppDialog 
                {...dialogConfig}
                inputValue={deleteReason}
                onInputChange={setDeleteReason}
                onClose={closeDialog}
            />
        </>
    );
};

export default assetBulkActions;