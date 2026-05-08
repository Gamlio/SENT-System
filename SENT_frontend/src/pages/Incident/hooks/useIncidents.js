import { useState, useCallback } from 'react';
import axiosInstance from '../../../api/axios';

export const useIncidents = () => {
    const [incidents, setIncidents] = useState([]);
    const [detail, setDetail] = useState(null);
    const [total, setTotal] = useState(0);
    const [loading, setLoading] = useState(false);

    const fetchList = useCallback(async (page = 1) => {
    setLoading(true);
    try {
        // Đảm bảo endpoint khớp với router Backend[cite: 60]
        const res = await axiosInstance.get(`/incidents?page=${page}&limit=12`);
        
        console.log("Data nhận được:", res.data); // Debug để xem cấu trúc thực tế

        // Nếu Backend đã sửa như trên, dòng này sẽ chạy đúng[cite: 62]
        setIncidents(res.data.items || []); 
        setTotal(res.data.total || 0);
    } catch (err) {
        console.error("Lỗi fetch:", err);
    } finally {
        setLoading(false);
    }
}, []);
    // Lấy chi tiết và Audit Logs[cite: 34]
    const fetchDetail = useCallback(async (id) => {
        setLoading(true);
        try {
            const res = await axiosInstance.get(`/incidents/${id}`);
            setDetail(res.data);
        } catch (err) {
            console.error("Lỗi tải bằng chứng:", err);
        } finally {
            setLoading(false);
        }
    }, []);

    return { incidents, detail, total, loading, fetchList, fetchDetail };
};