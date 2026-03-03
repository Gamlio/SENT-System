import { useState, useCallback } from 'react';
import axios from '../api/axios';

export const usePolicies = () => {
    const [policies, setPolicies] = useState([]);
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

    // 2. THÊM LUẬT MỚI (JSON)
    const addPolicy = async (policyData) => {
        try {
            await axios.post('/policies', policyData);
            // Sau khi thêm, nên gọi fetchPolicies lại ở component cha hoặc cập nhật state cục bộ
            return { success: true };
        } catch (err) {
            console.error("Lỗi thêm chính sách:", err);
            return { success: false, error: err.response?.data?.error || "Lỗi lưu dữ liệu" };
        }
    };

    // 3. XÓA LUẬT
    const deletePolicy = async (id) => {
        try {
            await axios.delete(`/policies/${id}`);
            setPolicies(prev => prev.filter(p => p.ID !== id)); // Cập nhật UI ngay
            return { success: true };
        } catch (err) {
            alert("Không thể xóa chính sách này!");
            return { success: false };
        }
    };

    return {
        policies,
        loading,
        error,
        fetchPolicies,
        addPolicy,
        deletePolicy
    };
};