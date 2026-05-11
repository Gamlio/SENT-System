import { useState, useEffect } from 'react';
import axios from '../../../api/axios';

export const useAssetTypes = () => {
    const [types, setTypes] = useState([]);
    const [isLoading, setIsLoading] = useState(false);

    const fetchTypes = async () => {
        setIsLoading(true);
        try {
            const res = await axios.get('/assets/types');
            setTypes(res.data || []);
        } catch (err) {
            console.error("Lỗi lấy danh sách loại tài sản:", err);
        } finally {
            setIsLoading(false);
        }
    };

    useEffect(() => { fetchTypes(); }, []);

    const updateType = async (id, data) => {
        try {
            await axios.put(`/assets/types/${id}`, data);
            fetchTypes();
            return { success: true };
        } catch (err) {
            return { success: false, error: err.response?.data?.error };
        }
    };
    const createType = async (data) => {
    try {
        await axios.post(`/assets/types`, data);
        fetchTypes(); 
        return { success: true };
    } catch (err) {
        return { success: false, error: err.response?.data?.error };
    }
};
    return { types, isLoading, fetchTypes, updateType, createType };
};