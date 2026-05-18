import { useState, useCallback } from 'react';
import axios from '../../../api/axios';

export const usePolicies = () => {
    const [policies, setPolicies] = useState([]);
    const [groups, setGroups] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const fetchPolicies = useCallback(async (category = '') => {
        setLoading(true);
        try {
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

    const fetchGroups = useCallback(async () => {
        try {
            const res = await axios.get('/groups');
            setGroups(res.data.data || []);
        } catch (err) {
            console.error("Lỗi tải danh sách nhóm:", err);
            setGroups([]);
        }
    }, []);

    const addPolicy = async (policyData) => {
        try {
            await axios.post('/policies', policyData);
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
            setPolicies(prev => prev.filter(p => p.ID !== id && p.id !== id));
            return { success: true };
        } catch (err) {
            console.error("Lỗi xóa chính sách:", err);
            return { success: false, error: "Không thể xóa chính sách này" };
        }
    };

    const deleteBulkPolicies = async (payload) => {
        try {
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
        groups,      
        loading,
        error,
        fetchPolicies,
        fetchGroups, 
        addPolicy,
        deletePolicy,
        deleteBulkPolicies
    };
};