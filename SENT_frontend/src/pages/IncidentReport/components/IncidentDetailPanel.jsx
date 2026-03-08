// src/pages/IncidentReport/components/IncidentDetailPanel.jsx
import React, { useState, useEffect } from 'react';
import axiosInstance from '../../../api/axios';
import IncidentHeader from './DetailParts/IncidentHeader';
import IncidentInfoSidebar from './DetailParts/IncidentInfoSidebar';
import IncidentTimeline from './DetailParts/IncidentTimeline';
import IncidentActionBox from './DetailParts/IncidentActionBox';

const IncidentDetailPanel = ({ incidentId, onClose, onUpdate }) => {
    const [incident, setIncident] = useState(null);
    const [loading, setLoading] = useState(true);
    const [sending, setSending] = useState(false);

    // Fetch dữ liệu
    const fetchDetail = async () => {
        try {
            const res = await axiosInstance.get(`/incidents/${incidentId}`);
            setIncident(res.data);
        } catch (err) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (incidentId) fetchDetail();
    }, [incidentId]);

    // Xử lý logic hành động
    const handleAction = async (type, content, files = []) => {
        // 1. [CHẶN ĐÓNG CASE] Nếu chọn Resolve, phải kiểm tra Playbook
        if (type === 'RESOLVE') {
            try {
                const progressData = JSON.parse(incident.playbook_progress || '{}');
                const steps = progressData.steps || [];
                // Kiểm tra xem có bước nào chưa done (false) không
                const notDone = steps.some(s => !s.done);
                
                if (notDone) {
                    alert("⛔ KHÔNG THỂ ĐÓNG SỰ CỐ!\n\nBạn chưa hoàn thành hết các bước trong quy trình Playbook (Cột trái).\nVui lòng thực hiện đầy đủ trước khi xác nhận xử lý xong.");
                    return; 
                }
            } catch (e) {
                // Nếu lỗi parse JSON thì cho qua hoặc chặn tùy ý
            }
            
            // Confirm lần cuối
            if (!window.confirm("Xác nhận: Bạn đã hoàn tất xử lý và muốn đóng hồ sơ này?")) return;
        }

        // 2. Validate nội dung gửi
        if (!content.trim() && files.length === 0 && type === 'COMMENT') return;

        setSending(true);
        try {
            const formData = new FormData();
            formData.append('action_type', type);
            formData.append('content', content);
            files.forEach(f => formData.append('files', f));

            await axiosInstance.post(`/incidents/${incidentId}/activity`, formData, {
                headers: { 'Content-Type': 'multipart/form-data' }
            });

            await fetchDetail(); // Load lại để hiện cái vừa gửi
            if (onUpdate) onUpdate(); // Refresh bảng bên ngoài
        } catch (err) {
            alert("Lỗi kết nối Server!");
        } finally {
            setSending(false);
        }
    };

    if (!incident) return null;

    return (
        <div className="absolute inset-0 z-50 flex justify-end bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
            <div className="w-full max-w-6xl h-full bg-[#0f172a] border-l border-slate-700 shadow-2xl flex flex-col animate-in slide-in-from-right duration-300">
                <IncidentHeader incident={incident} onClose={onClose} />
                <div className="flex-1 flex overflow-hidden">
                    <IncidentInfoSidebar incident={incident} onUpdate={fetchDetail} /> {/* Truyền fetchDetail để khi tick checkbox thì update lại state incident */}
                    
                    <div className="flex-1 flex flex-col min-w-0 bg-[#0f172a]">
                        <IncidentTimeline incident={incident} />
                        
                        {/* Chỉ hiện ô nhập liệu nếu chưa Resolved. Nếu Resolved thì ẩn đi để chỉ xem được thôi (Giải quyết vấn đề 1) */}
                        {incident.status !== 'Resolved' ? (
                            <IncidentActionBox 
                                status={incident.status} 
                                onAction={handleAction} 
                                sending={sending}
                            />
                        ) : (
                            <div className="p-4 bg-emerald-900/20 border-t border-emerald-500/30 text-center text-emerald-400 text-sm font-bold">
                                ✅ Hồ sơ sự cố đã đóng. Chỉ có thể xem lịch sử.
                            </div>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default IncidentDetailPanel;