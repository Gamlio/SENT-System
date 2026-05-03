import React, { useState, useEffect, useRef } from 'react';
import { MoreVertical, Cpu, Server, Briefcase, UserX, Trash2, X, UserCheck } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import AppDialog from '../../../components/AppDialog';
import axios from '../../../api/axios'; // Đảm bảo đường dẫn axios chuẩn

const assetActions = ({ asset, onRefresh, onOpenAssignModal }) => {
    const navigate = useNavigate();
    const [isOpen, setIsOpen] = useState(false);
    const [showTypeModal, setShowTypeModal] = useState(false);
    const [isLoadingType, setIsLoadingType] = useState(false);
    const [assetTypes, setAssetTypes] = useState([]);
    const [isLoadingTypes, setIsLoadingTypes] = useState(false);
    const menuRef = useRef(null);
    const [dialogConfig, setDialogConfig] = useState({
        isOpen: false,
        title: '',
        message: '',
        type: 'info',
        isAlertOnly: false,
        onConfirm: null
    });
    const closeDialog = () => setDialogConfig({ ...dialogConfig, isOpen: false });
    // Xử lý click ra ngoài để đóng menu
    useEffect(() => {
        const handleClickOutside = (event) => {
            if (menuRef.current && !menuRef.current.contains(event.target)) {
                setIsOpen(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    // Lấy danh sách từ Backend khi mở Modal
    useEffect(() => {
        const fetchTypes = async () => {
            try {
                setIsLoadingTypes(true);
                const res = await axios.get('/assets/types');
                setAssetTypes(res.data.data || res.data || []);
            } catch (error) {
                console.error('Lỗi khi tải danh sách loại thiết bị:', error);
            } finally {
                setIsLoadingTypes(false);
            }
        };
        if (showTypeModal) fetchTypes();
    }, [showTypeModal]);

    // Gọi API cập nhật phân loại
   const handleChangeType = async (typeId) => {
        if (isLoadingType) return; // Ngăn chặn người dùng bấm nhiều lần (Race condition)
        setIsLoadingType(true);
        try {
            await axios.put(`/assets/${asset.asset_hwid}/device-type`, { device_type: typeId });
            setShowTypeModal(false);
            if (onRefresh) onRefresh();
            setDialogConfig({
                isOpen: true,
                title: 'Thành công!',
                message: 'Đã cập nhật phân loại thiết bị. Điểm rủi ro (Risk Score) đã được tính toán lại.',
                type: 'success',
                isAlertOnly: true,
                onConfirm: () => window.location.reload()
            });
        } catch (err) {
            setDialogConfig({
                isOpen: true, title: 'Lỗi cập nhật', message: 'Không thể cập nhật phân loại thiết bị.', type: 'danger', isAlertOnly: true
            });
        } finally {
            setIsLoadingType(false);
        }
    };
const handleDeleteasset = () => {
        setIsOpen(false); // Đóng menu thả xuống
        
        // Mở Dialog Hỏi "Bạn có chắc chắn?"
        setDialogConfig({
            isOpen: true,
            title: 'Yêu cầu gỡ bỏ máy trạm?',
            message: `Bạn đang gửi yêu cầu gỡ bỏ hệ thống giám sát trên máy ${asset.hostname} (${asset.asset_hwid}). Thao tác này cần SOC Admin phê duyệt.`,
            type: 'danger',
            isAlertOnly: false,
            confirmText: 'Gửi yêu cầu xóa',
            onConfirm: async () => {
                closeDialog();
                try {
                    await axios.post(`/assets/${asset.asset_hwid}/request-delete`, {});
                    // Gọi API thành công -> Bật Dialog báo thành công
                    setDialogConfig({
                        isOpen: true,
                        title: 'Đã gửi yêu cầu',
                        message: 'Đơn xin gỡ bỏ thiết bị đã được chuyển đến Trung tâm Phê duyệt.',
                        type: 'success',
                        isAlertOnly: true,
                        onConfirm: () => window.location.reload()
                    });
                } catch (err) {
                    // Lỗi -> Bật Dialog báo lỗi
                    setDialogConfig({
                        isOpen: true,
                        title: 'Từ chối yêu cầu',
                        message: err.response?.data?.error || "Không thể gửi yêu cầu xóa lúc này.",
                        type: 'warning',
                        isAlertOnly: true
                    });
                }
            }
        });
    };

    return (
        <div className="relative flex items-center justify-end" ref={menuRef}>
            <button 
                onClick={(e) => {
                    e.stopPropagation(); 
                    setIsOpen(!isOpen);
                }}
                className={`p-2 rounded-lg transition ${isOpen ? 'bg-slate-700 text-emerald-400' : 'text-slate-400 hover:text-emerald-400 hover:bg-slate-800'}`}
            >
                <MoreVertical size={20} />
            </button>

            {/* DANH SÁCH MENU */}
           {isOpen && (
                <div className="absolute right-0 top-full mt-1 w-48 bg-[#0A101D] border border-slate-700 rounded-lg shadow-[0_10px_40px_rgba(0,0,0,0.8)] z-[99] overflow-hidden animate-in fade-in zoom-in duration-100">
                    <div className="px-3 py-1.5 border-b border-slate-800 bg-[#111827]">
                        <p className="text-[9px] font-black text-slate-500 uppercase tracking-widest">Action Menu</p>
                    </div>
                    
                    <button
                        onClick={(e) => { e.stopPropagation(); setIsOpen(false); if(onOpenAssignModal) onOpenAssignModal(); }}
                        className="w-full text-left px-3 py-2.5 text-[11px] font-bold text-slate-300 hover:bg-slate-800 hover:text-indigo-400 flex items-center gap-2.5 transition border-l-2 border-indigo-500/0 hover:border-indigo-500"
                    >
                        <UserCheck size={14} /> Assign Owner
                    </button>

                    <button
                        onClick={(e) => { e.stopPropagation(); setIsOpen(false); setShowTypeModal(true); }}
                        className="w-full text-left px-3 py-2.5 text-[11px] font-bold text-slate-300 hover:bg-slate-800 hover:text-purple-400 flex items-center gap-2.5 transition"
                    >
                        <Cpu size={14} /> Asset Tier
                    </button>

                    <button 
                        onClick={(e) => { e.stopPropagation(); handleDeleteasset(); }}
                        className="w-full text-left px-3 py-2.5 text-[11px] font-bold text-red-400 hover:bg-red-500/10 hover:text-red-300 flex items-center gap-2.5 transition border-t border-slate-800"
                    >
                        <Trash2 size={14} /> Remove asset
                    </button>
                </div>
            )}
            <AppDialog 
                {...dialogConfig} 
                onClose={closeDialog} 
            />
            {/* [MỚI] MODAL PHÂN LOẠI THIẾT BỊ */}
            {showTypeModal && (
                <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-[9999] backdrop-blur-sm p-4 animate-in fade-in duration-200" onClick={(e) => e.stopPropagation()}>
                    <div className="bg-[#1e293b] w-full max-w-md rounded-3xl border border-slate-700 shadow-2xl overflow-hidden" onClick={(e) => e.stopPropagation()}>
                        <div className="p-6 border-b border-slate-800 flex justify-between items-center bg-slate-800/50">
                            <div>
                                <h3 className="font-bold text-white flex items-center gap-2"><Cpu className="text-purple-400"/> Phân loại Tài sản (Asset Tier)</h3>
                                <p className="text-[10px] text-slate-500 uppercase font-black mt-1">Máy: {asset.hostname}</p>
                            </div>
                            <button onClick={() => setShowTypeModal(false)} className="text-slate-500 hover:text-white transition"><X size={20}/></button>
                        </div>
                        
                        <div className="p-4 space-y-3 max-h-[60vh] overflow-y-auto custom-scrollbar">
                            {isLoadingTypes ? (
                                <div className="flex flex-col items-center justify-center py-8 text-indigo-500 gap-3">
                                    <div className="w-6 h-6 border-2 border-purple-500 border-t-transparent rounded-full animate-spin"></div>
                                    <span className="text-[10px] font-black uppercase tracking-widest animate-pulse">Đang tải phân loại...</span>
                                </div>
                            ) : assetTypes.length > 0 ? (
                                assetTypes.map((type) => (
                                    <button 
                                        key={type.ID || type.id}
                                        disabled={isLoadingType}
                                        onClick={() => handleChangeType(type.ID || type.id)}
                                        className={`w-full flex items-center p-4 rounded-2xl border transition group ${isLoadingType ? 'opacity-50 cursor-not-allowed' : ''} ${asset.device_type === (type.ID || type.id) ? 'bg-purple-500/20 border-purple-500' : 'border-slate-700 bg-slate-800/50 hover:bg-purple-500/10 hover:border-purple-500/50'}`}
                                    >
                                        <div className="text-left flex-1">
                                            <p className={`text-sm font-bold transition ${asset.device_type === (type.ID || type.id) ? 'text-purple-400' : 'text-white group-hover:text-purple-400'}`}>
                                                {type.name}
                                            </p>
                                            <p className="text-[10px] text-slate-400 mt-0.5 italic line-clamp-1">
                                                {type.description || 'Không có mô tả chi tiết'}
                                            </p>
                                        </div>
                                        <div className="flex flex-col items-end justify-center ml-3 shrink-0">
                                            <span className="text-[9px] font-black tracking-widest text-slate-500 uppercase mb-0.5">
                                                Risk Weight
                                            </span>
                                            <span className={`px-2 py-0.5 rounded text-[10px] font-black font-mono border ${
                                                (type.risk_weight || 1) >= 1.5 
                                                    ? 'bg-red-500/10 text-red-400 border-red-500/30' 
                                                    : (type.risk_weight || 1) > 1.0 
                                                        ? 'bg-amber-500/10 text-amber-400 border-amber-500/30' 
                                                        : 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
                                            }`}>
                                                x{Number(type.risk_weight || 1).toFixed(1)}
                                            </span>
                                        </div>
                                    </button>
                                ))
                            ) : (
                                <div className="text-center py-6 text-slate-500 text-[10px] font-bold uppercase tracking-widest border-2 border-dashed border-slate-800 rounded-lg">
                                    Không tìm thấy phân loại thiết bị
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default assetActions;