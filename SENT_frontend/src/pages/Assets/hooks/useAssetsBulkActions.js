import { useState } from 'react';
import axios from '../../../api/axios';

export const useassetBulkActions = (fetchassets) => {
    const [selectedassets, setSelectedassets] = useState([]);
    const [deleteReason, setDeleteReason] = useState('');
    const [dialogConfig, setDialogConfig] = useState({ isOpen: false });

    // Hàm chọn tất cả máy trạm trên trang hiện tại
    const handleSelectAll = (currentassets, isChecked) => {
        if (isChecked) {
            setSelectedassets(currentassets.map(a => a.asset_hwid));
        } else {
            setSelectedassets([]);
        }
    };

    // Hàm chọn một máy trạm
    const handleSelectOne = (hwid) => {
        if (selectedassets.includes(hwid)) {
            setSelectedassets(selectedassets.filter(id => id !== hwid));
        } else {
            setSelectedassets([...selectedassets, hwid]);
        }
    };

    const clearSelection = () => setSelectedassets([]);

    const closeDialog = () => {
        setDialogConfig(prev => ({ ...prev, isOpen: false }));
        setDeleteReason(''); // Reset lý do
    };

    // Hàm xử lý logic Xóa hàng loạt
    const initiateBulkDelete = () => {
        setDialogConfig({
            isOpen: true,
            title: `Yêu cầu gỡ bỏ ${selectedassets.length} thiết bị?`,
            message: `Hành động này sẽ ngắt kết nối và xóa dữ liệu của ${selectedassets.length} máy trạm. Vui lòng nhập lý do.`,
            type: 'danger',
            confirmText: 'Gửi yêu cầu xóa',
            showInput: true,
            onConfirm: async (reason) => {
                if (reason.trim().length < 5) {
                    alert("Vui lòng nhập lý do rõ ràng (ít nhất 5 ký tự).");
                    return;
                }
                
                closeDialog();
                try {
                    await axios.post('/assets/bulk-request-delete', {
                        hwids: selectedassets,
                        reason: reason 
                    });
                    
                    clearSelection();
                    fetchassets(); // Tải lại danh sách
                    
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

    return {
        selectedassets,
        handleSelectAll,
        handleSelectOne,
        clearSelection,
        
        dialogConfig,
        closeDialog,
        
        deleteReason,
        setDeleteReason,
        
        initiateBulkDelete
    };
};