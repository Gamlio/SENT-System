import { useState, useCallback } from 'react';
import axios from '../../../api/axios';

export const usePolicies = () => {
    const [policies, setPolicies] = useState([]);
    const [groups, setGroups] = useState([]); // Quản lý danh sách nhóm máy trạm
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    // 1. LẤY DANH SÁCH LUẬT (Hỗ trợ lọc theo category)
    const fetchPolicies = useCallback(async (category = '') => {
        setLoading(true);
        try {
            // Nếu có category thì thêm query param, không thì lấy hết
            const endpoint = category ? `/policies?category=${category}` : '/policies';
            const res = await axios.get(endpoint);
            setPolicies(res.data || []);
            setError(null);
        } catch (err) {
            console.error("Lỗi tải chính sách:", err);
            setError("Không thể tải danh sách chính sách.");
        } finally {
            setLoading(false);
        }
    }, []);

    // 2. LẤY DANH SÁCH NHÓM (Dùng cho dropdown trong PolicyForm)
    const fetchGroups = useCallback(async () => {
        try {
            // Sửa đường dẫn từ '/groups' thành '/policies/groups'
            const res = await axios.get('/policies/groups'); 
            
            // Dùng cơ chế bảo vệ dữ liệu như bạn đã viết
            setGroups(Array.isArray(res.data) ? res.data : []); 
        } catch (err) {
            console.error("Lỗi tải danh sách nhóm:", err);
                setGroups([]); 
            }
        }, []);

    // 3. THÊM LUẬT MỚI (JSON)
    const addPolicy = async (policyData) => {
        try {
            // policyData giờ đây sẽ chứa group_id thay vì target_asset_hwids
            await axios.post('/policies', policyData);
            // Re-fetch sẽ được thực hiện thông qua Socket hoặc gọi thủ công ở component
            return { success: true };
        } catch (err) {
            console.error("Lỗi thêm chính sách:", err);
            return { 
                success: false, 
                error: err.response?.data?.error || "Lỗi lưu dữ liệu chính sách" 
            };
        }
    };

    // 4. XÓA LUẬT LẺ
    const deletePolicy = async (id) => {
        try {
            await axios.delete(`/policies/${id}`);
            // Cập nhật state cục bộ để UI mượt mà hơn
            setPolicies(prev => prev.filter(p => p.ID !== id && p.id !== id));
            return { success: true };
        } catch (err) {
            console.error("Lỗi xóa chính sách:", err);
            return { success: false, error: "Không thể xóa chính sách này" };
        }
    };

    // 5. XÓA NHIỀU LUẬT (Gửi đơn phê duyệt)
    const deleteBulkPolicies = async (payload) => {
        try {
            // payload: { ids: [1, 2, 3], reason: "..." }
            await axios.post('/policies/bulk-delete', payload);
            return { success: true };
        } catch (err) {
            console.error("Lỗi gửi đơn xóa hàng loạt:", err);
            return { 
                success: false, 
                error: err.response?.data?.error || "Lỗi gửi yêu cầu xóa hàng loạt" 
            };
        }
    };

    return {
        policies,
        groups,      // Danh sách Group để hiển thị trong Form/List
        loading,
        error,
        fetchPolicies,
        fetchGroups, // Hàm nạp dữ liệu Group
        addPolicy,
        deletePolicy,
        deleteBulkPolicies
    };
};