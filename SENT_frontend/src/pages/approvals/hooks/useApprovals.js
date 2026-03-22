import { useState, useCallback } from 'react';
import axios from '../../../api/axios'; 

export const useApprovals = () => {
    const [tickets, setTickets] = useState([]);
    const [loading, setLoading] = useState(false);

    const fetchTickets = useCallback(async (moduleType = '', status = 'PENDING') => {
        setLoading(true);
        try {
            let url = `/approvals?module_type=${moduleType}`;
            if (status !== 'ALL') {
                url += `&status=${status}`;
            }
            const res = await axios.get(url);
            setTickets(res.data || []);
        } catch (error) {
            console.error("Lỗi lấy danh sách phê duyệt:", error);
        } finally {
            setLoading(false);
        }
    }, []);

    const reviewTicket = async (id, status, reviewNote = '') => {
        try {
            await axios.put(`/approvals/${id}/review`, { status, review_note: reviewNote });
            return { success: true };
        } catch (error) {
            return { success: false, error: error.response?.data?.error || "Lỗi xử lý" };
        }
    };

    return { tickets, loading, fetchTickets, reviewTicket };
};