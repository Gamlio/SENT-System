import React, { useState, useEffect } from 'react';
import axiosInstance from '../../api/axios';
import IncidentTable from './components/IncidentTable';
import IncidentDetailPanel from './components/IncidentDetailPanel';

const IncidentManager = () => {
    const [incidents, setIncidents] = useState([]);
    const [loading, setLoading] = useState(true);
    const [selectedIncidentId, setSelectedIncidentId] = useState(null); // Sửa: Chỉ lưu ID
    const [loadingDetail, setLoadingDetail] = useState(false); // Fix lỗi biến

    // Fetch danh sách
    const fetchIncidents = async () => {
        try {
            const response = await axiosInstance.get('/incidents');
            // Xử lý dữ liệu trả về linh hoạt (data.data hoặc data)
            const data = Array.isArray(response.data) ? response.data : (response.data.data || []);
            setIncidents(data);
        } catch (error) {
            console.error("Lỗi tải danh sách:", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchIncidents();
        const interval = setInterval(fetchIncidents, 15000); 
        return () => clearInterval(interval);
    }, []);

    // Khi bấm nút Chi tiết ở bảng
    const handleViewDetail = (id) => {
        setSelectedIncidentId(id); // Chỉ set ID để mở Panel
    };

    // Khi đóng Panel -> Refresh lại bảng để cập nhật trạng thái mới (VD: Open -> Resolved)
    const handleCloseDetail = () => {
        setSelectedIncidentId(null);
        fetchIncidents(); 
    };

    if (loading) return <div className="p-6 text-slate-400 font-bold animate-pulse">Đang tải dữ liệu...</div>;

    return (
        <div className="p-6 text-slate-200 h-full flex flex-col relative overflow-hidden">
            <IncidentTable rawData={incidents} onViewDetail={handleViewDetail} />

            {/* Chỉ hiện Panel khi có ID */}
            {selectedIncidentId && (
                <IncidentDetailPanel 
                    incidentId={selectedIncidentId}  // Truyền ID vào đây
                    onClose={handleCloseDetail}
                    onUpdate={fetchIncidents} // Truyền hàm refresh
                />
            )}
        </div>
    );
};

export default IncidentManager;