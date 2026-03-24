import React, { useState, useEffect, useRef } from 'react';
import { MoreVertical, ShieldCheck, Terminal, Lock, Usb, Cpu, Server, Briefcase, UserX,Trash2, X,Fingerprint } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import AppDialog from '../../../components/AppDialog';
import axios from '../../../api/axios'; // Đảm bảo đường dẫn axios chuẩn

const AgentActions = ({ agent }) => {
    const navigate = useNavigate();
    const [isOpen, setIsOpen] = useState(false);
    const [showTypeModal, setShowTypeModal] = useState(false);
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

    // Gọi API cập nhật phân loại
   const handleChangeType = async (type) => {
        try {
            await axios.put(`/agents/${agent.hwid}/device-type`, { device_type: type });
            setShowTypeModal(false);
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
        }
    };
const handleDeleteAgent = () => {
        setIsOpen(false); // Đóng menu thả xuống
        
        // Mở Dialog Hỏi "Bạn có chắc chắn?"
        setDialogConfig({
            isOpen: true,
            title: 'Yêu cầu gỡ bỏ máy trạm?',
            message: `Bạn đang gửi yêu cầu gỡ bỏ hệ thống giám sát trên máy ${agent.hostname} (${agent.hwid}). Thao tác này cần SOC Admin phê duyệt.`,
            type: 'danger',
            isAlertOnly: false,
            confirmText: 'Gửi yêu cầu xóa',
            onConfirm: async () => {
                closeDialog();
                try {
                    await axios.post(`/agents/${agent.hwid}/request-delete`);
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
                <button
                onClick={(e) => { 
                    e.stopPropagation(); 
                    setIsOpen(false);
                    handleTriggerBaseline(); // Hàm mới để gọi API Push qua WebSocket hoặc REST
                }}
                className="w-full text-left px-4 py-3 text-sm text-slate-300 hover:bg-slate-700 hover:text-emerald-400 flex items-center gap-3 transition border-l-2 border-emerald-500/50 bg-emerald-500/5"
            >
                <Fingerprint size={16} /> Quét & Thiết lập Baseline
            </button>
            const handleTriggerBaseline = async () => {
            setDialogConfig({
                isOpen: true,
                title: 'Thiết lập Zero Trust?',
                message: 'Hệ thống sẽ ra lệnh cho Agent quét toàn bộ phần mềm và USB hiện tại để làm danh sách Whitelist mặc định.',
                type: 'info',
                onConfirm: async () => {
                    try {
                        await axios.post(`/agents/${agent.hwid}/trigger-baseline`);
                        alert("Đã gửi lệnh quét tới máy trạm qua WebSocket!");
                    } catch (err) {
                        alert("Lỗi khi gửi lệnh.");
                    }
                }
            });
        };
    return (
        <div className="relative flex items-center" ref={menuRef}>
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
                <div className="absolute right-0 top-full mt-2 w-56 bg-[#1e293b] border border-slate-700 rounded-xl shadow-2xl z-50 overflow-hidden animate-in fade-in zoom-in duration-200">
                    <div className="px-4 py-2 border-b border-slate-700/50 bg-slate-800/50">
                        <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Tiện ích mở rộng</p>
                    </div>
                    
                    <button
                        onClick={(e) => { 
                            e.stopPropagation(); 
                            setIsOpen(false);
                            navigate(`/agents/${agent.hwid}/software`); 
                        }}
                        className="w-full text-left px-4 py-3 text-sm text-slate-300 hover:bg-slate-700 hover:text-emerald-400 flex items-center gap-3 transition"
                    >
                        <ShieldCheck size={16} /> Cấu hình Phần mềm
                    </button>

                    {/* [MỚI] NÚT ĐỔI PHÂN LOẠI THIẾT BỊ */}
                    <button
                        onClick={(e) => { 
                            e.stopPropagation(); 
                            setIsOpen(false);
                            setShowTypeModal(true); // Bật Modal
                        }}
                        className="w-full text-left px-4 py-3 text-sm text-slate-300 hover:bg-slate-700 hover:text-purple-400 flex items-center gap-3 transition border-l-2 border-purple-500/50 bg-purple-500/5"
                    >
                        <Cpu size={16} /> Phân loại thiết bị
                    </button>

                    <button className="w-full text-left px-4 py-3 text-sm text-slate-600 flex items-center gap-3 cursor-not-allowed" title="Sắp ra mắt">
                        <Usb size={16} /> Quản lý USB (Sắp có)
                    </button>
                    <button className="w-full text-left px-4 py-3 text-sm text-slate-600 flex items-center gap-3 cursor-not-allowed" title="Sắp ra mắt">
                        <Terminal size={16} /> Remote Terminal
                    </button>
                    <button className="w-full text-left px-4 py-3 text-sm text-red-500/50 flex items-center gap-3 cursor-not-allowed" title="Sắp ra mắt">
                        <Lock size={16} /> Khóa máy từ xa
                    </button>
                    <button 
                        onClick={(e) => {
                            e.stopPropagation();
                            handleDeleteAgent(); // Gọi hàm xóa mới
                        }}
                        className="w-full text-left px-4 py-3 text-sm text-red-400 hover:bg-red-500/10 hover:text-red-300 flex items-center gap-3 transition"
                    >
                        <Trash2 size={16} /> Yêu cầu Gỡ bỏ
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
                                <p className="text-[10px] text-slate-500 uppercase font-black mt-1">Máy: {agent.hostname}</p>
                            </div>
                            <button onClick={() => setShowTypeModal(false)} className="text-slate-500 hover:text-white transition"><X size={20}/></button>
                        </div>
                        
                        <div className="p-4 space-y-3">
                            <button onClick={() => handleChangeType('SERVER')} className={`w-full flex flex-col p-4 rounded-2xl border transition group relative overflow-hidden ${agent.device_type === 'SERVER' ? 'bg-purple-500/20 border-purple-500' : 'border-slate-700 bg-slate-800/50 hover:bg-purple-500/10 hover:border-purple-500/50'}`}>
                                <div className="flex items-center gap-3">
                                    <div className="p-2 bg-purple-500/20 text-purple-400 rounded-lg"><Server size={20}/></div>
                                    <div className="text-left">
                                        <p className={`text-sm font-bold transition ${agent.device_type === 'SERVER' ? 'text-purple-400' : 'text-white group-hover:text-purple-400'}`}>Máy chủ (Tier 1)</p>
                                        <p className="text-[10px] text-slate-400 mt-0.5">Hệ số rủi ro: <span className="font-bold text-white">x1.5</span> | Dành cho Server, Giám đốc</p>
                                    </div>
                                </div>
                            </button>

                            <button onClick={() => handleChangeType('IT_ADMIN')} className={`w-full flex flex-col p-4 rounded-2xl border transition group ${agent.device_type === 'IT_ADMIN' ? 'bg-blue-500/20 border-blue-500' : 'border-slate-700 bg-slate-800/50 hover:bg-blue-500/10 hover:border-blue-500/50'}`}>
                                <div className="flex items-center gap-3">
                                    <div className="p-2 bg-blue-500/20 text-blue-400 rounded-lg"><Briefcase size={20}/></div>
                                    <div className="text-left">
                                        <p className={`text-sm font-bold transition ${agent.device_type === 'IT_ADMIN' ? 'text-blue-400' : 'text-white group-hover:text-blue-400'}`}>IT Admin (Tier 2)</p>
                                        <p className="text-[10px] text-slate-400 mt-0.5">Hệ số rủi ro: <span className="font-bold text-white">x1.2</span> | Dành cho IT, Quản trị viên</p>
                                    </div>
                                </div>
                            </button>

                            <button onClick={() => handleChangeType('OFFICE')} className={`w-full flex flex-col p-4 rounded-2xl border transition group ${(!agent.device_type || agent.device_type === 'OFFICE') ? 'bg-slate-700 border-slate-500' : 'border-slate-700 bg-slate-800/50 hover:bg-slate-700'}`}>
                                <div className="flex items-center gap-3">
                                    <div className="p-2 bg-slate-600 text-white rounded-lg"><Cpu size={20}/></div>
                                    <div className="text-left">
                                        <p className="text-sm font-bold text-white">Văn phòng (Tier 3)</p>
                                        <p className="text-[10px] text-slate-400 mt-0.5">Hệ số rủi ro: <span className="font-bold text-white">x1.0</span> | Máy tính làm việc tiêu chuẩn</p>
                                    </div>
                                </div>
                            </button>

                            <button onClick={() => handleChangeType('GUEST')} className={`w-full flex flex-col p-4 rounded-2xl border transition group ${agent.device_type === 'GUEST' ? 'bg-stone-500/20 border-stone-500' : 'border-slate-700 bg-slate-800/50 hover:bg-stone-500/10 hover:border-stone-500/50'}`}>
                                <div className="flex items-center gap-3">
                                    <div className="p-2 bg-stone-500/20 text-stone-400 rounded-lg"><UserX size={20}/></div>
                                    <div className="text-left">
                                        <p className={`text-sm font-bold transition ${agent.device_type === 'GUEST' ? 'text-stone-400' : 'text-white group-hover:text-stone-400'}`}>Máy Khách (Tier 4)</p>
                                        <p className="text-[10px] text-slate-400 mt-0.5">Hệ số rủi ro: <span className="font-bold text-white">x0.8</span> | Máy Public, Lễ tân</p>
                                    </div>
                                </div>
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default AgentActions;