import React, { useState, useEffect } from 'react';
import axiosInstance from '../../api/axios';
import IncidentTable from './components/IncidentTable';
import IncidentDetailPanel from './components/IncidentDetailPanel';

const IncidentManager = () => {
    // --- STATE QUẢN LÝ ---
    const [incidents, setIncidents] = useState([]);
    const [loading, setLoading] = useState(true);
    
    // State quản lý xem chi tiết
    const [selectedIncident, setSelectedIncident] = useState(null); 
    const [loadingDetail, setLoadingDetail] = useState(false);

    // --- API CALLS ---
    const fetchIncidents = async () => {
        try {
            const response = await axiosInstance.get('/incidents');
            setIncidents(response.data.data || []);
        } catch (error) {
            console.error("Lỗi tải danh sách:", error);
        } finally {
            setLoading(false);
        }
    };

    // Auto refresh
    useEffect(() => {
        fetchIncidents();
        const interval = setInterval(fetchIncidents, 15000); 
        return () => clearInterval(interval);
    }, []);

    // --- HANDLERS (SỰ KIỆN) ---
    
    // 1. Khi bấm vào 1 dòng trong bảng -> Gọi API lấy chi tiết
    const handleViewDetail = async (id) => {
        setLoadingDetail(true);
        try {
            const response = await axiosInstance.get(`/incidents/${id}`);
            setSelectedIncident(response.data.data);
        } catch (error) {
            alert("Không thể tải chi tiết sự cố");
        } finally {
            setLoadingDetail(false);
        }
    };

    // 2. Khi đóng panel
    const handleCloseDetail = () => {
        setSelectedIncident(null);
        fetchIncidents(); // Refresh lại danh sách để cập nhật trạng thái nếu có thay đổi
    };

    // 3. Khi bấm nút "Đánh dấu đã xử lý" (Logic giả lập, bạn cần thêm API backend)
    const handleResolve = async (id) => {
        if(window.confirm("Xác nhận đóng hồ sơ sự cố này?")) {
            // Gọi API Update Status (Cần viết thêm API PUT /incidents/:id)
            // await axiosInstance.put(`/incidents/${id}`, { status: 'Resolved' });
            alert("Đã đánh dấu xử lý (Demo)");
            handleCloseDetail();
        }
    };

    // --- RENDER ---
    if (loading) return <div className="p-6 text-slate-400 font-bold animate-pulse">Đang tải dữ liệu...</div>;

    return (
        <div className="p-6 text-slate-200 h-full flex flex-col relative overflow-hidden">
            {/* Component Danh Sách */}
            <IncidentTable 
                rawData={incidents} 
                onViewDetail={handleViewDetail} 
            />

            {/* Component Chi Tiết (Chỉ hiện khi selectedIncident != null) */}
            {selectedIncident && (
                <IncidentDetailPanel 
                    incident={selectedIncident} 
                    onClose={handleCloseDetail}
                    onResolve={handleResolve}
                />
            )}
        </div>
    );
};

export default IncidentManager;