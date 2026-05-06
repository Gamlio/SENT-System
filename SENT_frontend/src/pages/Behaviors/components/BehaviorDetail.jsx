import React, { useEffect, useState } from 'react';
import { X, ShieldAlert, Activity, Monitor, Clock, FileText, Database, Loader2 } from 'lucide-react';
import axiosInstance from '../../../api/axios';

const BehaviorDetailModal = ({ behaviorId, onClose, onCreated }) => {
    const [data, setData] = useState(null);
    const [loading, setLoading] = useState(true);
    const [isCreating, setIsCreating] = useState(false); // Trạng thái khi đang tạo sự cố

    useEffect(() => {
        const fetchDetail = async () => {
            try {
                const res = await axiosInstance.get(`/behaviors/${behaviorId}`);
                setData(res.data);
            } catch (err) {
                console.error("Lỗi tải chi tiết:", err);
            } finally {
                setLoading(false);
            }
        };
        fetchDetail();
    }, [behaviorId]);

    // Hàm xử lý lập hồ sơ sự cố chính thức
    const handleCreateIncident = async () => {
        if (!window.confirm("Xác nhận lập hồ sơ sự cố dựa trên hành vi này?")) return;

        setIsCreating(true);
        try {
            // Gọi endpoint rút gọn đã thống nhất
            await axiosInstance.post('/behaviors/create-incident', {
                alert_id: behaviorId,
                severity: data.severity, // Lấy severity mặc định từ hành vi
                description: `Lập hồ sơ thủ công từ điều tra viên cho hành vi: ${data.title}`
            });

            alert("Đã lập hồ sơ sự cố thành công!");
            if (onCreated) onCreated(); // Gọi callback để load lại danh sách bên ngoài
            onClose(); // Đóng modal sau khi thành công
        } catch (err) {
            alert(err.response?.data?.error || "Lỗi khi lập hồ sơ");
        } finally {
            setIsCreating(false);
        }
    };

    if (!data && !loading) return null;

    return (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm">
            <div className="bg-[#0A101D] border border-slate-800 w-full max-w-2xl rounded-2xl overflow-hidden shadow-2xl">
                {/* Header */}
                <div className="flex justify-between items-center p-4 border-b border-slate-800 bg-slate-900/50">
                    <div className="flex items-center gap-3">
                        <Activity className="text-indigo-500" size={20}/>
                        <h3 className="text-sm font-black text-white uppercase tracking-widest text-[10px]">Điều tra chi tiết</h3>
                    </div>
                    <button onClick={onClose} className="text-slate-500 hover:text-white transition-colors">
                        <X size={20}/>
                    </button>
                </div>

                <div className="p-6 max-h-[70vh] overflow-y-auto">
                    {loading ? (
                        <div className="space-y-4 animate-pulse">
                            <div className="h-4 bg-slate-800 rounded w-3/4"></div>
                            <div className="h-24 bg-slate-800 rounded"></div>
                        </div>
                    ) : (
                        <div className="space-y-6">
                            {/* Thông tin máy & Thời gian[cite: 48] */}
                            <div className="grid grid-cols-2 gap-4">
                                <div className="bg-slate-900/50 p-3 rounded-xl border border-slate-800">
                                    <span className="text-[10px] text-slate-500 uppercase font-bold flex items-center gap-2 mb-2">
                                        <Monitor size={12}/> Asset HWID
                                    </span>
                                    <code className="text-[10px] text-indigo-400 font-mono break-all">{data.asset_hwid}</code>
                                </div>
                                <div className="bg-slate-900/50 p-3 rounded-xl border border-slate-800">
                                    <span className="text-[10px] text-slate-500 uppercase font-bold flex items-center gap-2 mb-2">
                                        <Clock size={12}/> Ghi nhận
                                    </span>
                                    <span className="text-[10px] text-slate-200">{new Date(data.created_at).toLocaleString('vi-VN')}</span>
                                </div>
                            </div>

                            {/* Nội dung chi tiết[cite: 48] */}
                            <div>
                                <span className="text-[10px] text-slate-500 uppercase font-bold flex items-center gap-2 mb-3 font-mono">
                                    <FileText size={12}/> Bằng chứng vi phạm
                                </span>
                                <div className="bg-red-500/5 border border-red-500/10 p-4 rounded-xl">
                                    <h4 className="text-xs font-bold text-white mb-2">{data.title}</h4>
                                    <p className="text-[11px] text-slate-400 leading-relaxed">{data.description}</p>
                                </div>
                            </div>
                            
                            {/* Phân tích hệ thống[cite: 48] */}
                            <div className="bg-[#050B14] p-4 rounded-xl border border-slate-800 border-dashed">
                                <span className="text-[10px] text-indigo-500 uppercase font-bold flex items-center gap-2 mb-3 font-mono">
                                    <Database size={12}/> Metadata Analytics
                                </span>
                                <div className="grid grid-cols-2 gap-y-3 text-[10px]">
                                    <span className="text-slate-500">Loại:</span>
                                    <span className="text-slate-200 font-mono text-right">{data.alert_type}</span>
                                    <span className="text-slate-500">Priority:</span>
                                    <span className="text-orange-500 font-bold text-right">{data.priority}</span>
                                    <span className="text-slate-500">Trạng thái:</span>
                                    <span className={`text-right ${data.is_resolved ? "text-emerald-500" : "text-amber-500"}`}>
                                        {data.is_resolved ? "Đã lập hồ sơ" : "Chờ xử lý"}
                                    </span>
                                </div>
                            </div>
                        </div>
                    )}
                </div>

                {/* Footer Modal[cite: 48] */}
                <div className="p-4 border-t border-slate-800 flex justify-end gap-3 bg-slate-900/30">
                    <button onClick={onClose} className="px-4 py-2 text-[10px] font-bold text-slate-500 hover:text-white transition-colors uppercase">
                        Đóng
                    </button>
                    {!data?.is_resolved && (
                        <button 
                            onClick={handleCreateIncident}
                            disabled={isCreating}
                            className="px-4 py-2 bg-red-600 hover:bg-red-500 text-white text-[10px] font-black uppercase rounded-lg transition-all flex items-center gap-2 shadow-lg shadow-red-500/10 disabled:opacity-50"
                        >
                            {isCreating ? <Loader2 size={14} className="animate-spin"/> : <ShieldAlert size={14}/>}
                            Lập hồ sơ sự cố
                        </button>
                    )}
                </div>
            </div>
        </div>
    );
};

export default BehaviorDetailModal;