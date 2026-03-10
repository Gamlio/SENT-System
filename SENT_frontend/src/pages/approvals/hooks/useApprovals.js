import { useState, useCallback } from 'react';
import axios from '../../../api/axios'; // Đảm bảo đường dẫn file axios của bạn đúng

export const useApprovals = () => {
    const [tickets, setTickets] = useState([]);
    const [loading, setLoading] = useState(false);

    const fetchTickets = useCallback(async (moduleType = '') => {
        setLoading(true);
        try {
            const endpoint = moduleType ? `/approvals?status=PENDING&module_type=${moduleType}` : '/approvals?status=PENDING';
            const res = await axios.get(endpoint);
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
            // Cập nhật state UI ngay lập tức
            setTickets(prev => prev.filter(ticket => ticket.id !== id));
            return { success: true };
        } catch (error) {
            console.error("Lỗi duyệt đơn:", error);
            return { success: false, error: "Không thể xử lý yêu cầu" };
        }
    };

    return { tickets, loading, fetchTickets, reviewTicket };
};