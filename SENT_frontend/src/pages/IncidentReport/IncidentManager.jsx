import React, { useState, useEffect } from 'react';
import axiosInstance from '../../api/axios';
import IncidentTable from './components/IncidentTable';
import IncidentDetailPanel from './components/IncidentDetailPanel';

const IncidentManager = () => {
    // --- STATE QUẢN LÝ ---
    const [incidents, setIncidents] = useState([]);
    const [loading, setLoading] = useState(true);
    const [selectedIncident, setSelectedIncident] = useState(null); 
    const [loadingDetail, setLoadingDetail] = useState(false);

    // --- API CALLS ---
    const fetchIncidents = async () => {
        try {
            const response = await axiosInstance.get('/incidents');
            // Đảm bảo dữ liệu là mảng, sắp xếp mới nhất lên đầu
            const data = response.data.data || response.data || [];
            setIncidents(data);
        } catch (error) {
            console.error("Lỗi tải danh sách sự cố:", error);
        } finally {
            setLoading(false);
        }
    };

    // Auto refresh mỗi 15s
    useEffect(() => {
        fetchIncidents();
        const interval = setInterval(fetchIncidents, 15000); 
        return () => clearInterval(interval);
    }, []);

    // --- HANDLERS ---
    const handleViewDetail = async (id) => {
        setLoadingDetail(true);
        try {
            const response = await axiosInstance.get(`/incidents/${id}`);
            setSelectedIncident(response.data.data || response.data);
        } catch (error) {
            alert("Không thể tải chi tiết sự cố");
        } finally {
            setLoadingDetail(false);
        }
    };

    const handleCloseDetail = () => {
        setSelectedIncident(null);
        fetchIncidents(); // Refresh lại danh sách
    };

    // Xử lý đóng Case (Có kèm theo ghi chú)
    const handleResolve = async (id, resolutionNote) => {
        try {
            // TODO: Bạn cần viết API PUT /api/v1/incidents/:id ở Backend để nhận data này
            // await axiosInstance.put(`/incidents/${id}`, { 
            //     status: 'Resolved', 
            //     resolution: resolutionNote 
            // });
            
            alert("Đã ghi nhận và đóng hồ sơ sự cố (Chế độ Demo)!\nGhi chú: " + resolutionNote);
            handleCloseDetail();
        } catch (err) {
            alert("Lỗi khi đóng sự cố!");
        }
    };

    if (loading) return <div className="p-6 text-slate-400 font-bold animate-pulse">Đang tải dữ liệu...</div>;

    return (
        <div className="p-6 text-slate-200 h-full flex flex-col relative overflow-hidden">
            <IncidentTable rawData={incidents} onViewDetail={handleViewDetail} />

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