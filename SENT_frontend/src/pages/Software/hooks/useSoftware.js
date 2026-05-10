import { useState, useEffect } from 'react';
import axios from '../../../api/axios';

export const useVersions = () => {
    const [softwares, setVersions] = useState([]);
    const [loading, setLoading] = useState(true);

    const fetchVersions = async () => {
        try {
            const res = await axios.get('/softwares');
            setVersions(res.data);
        } catch (err) {
            console.error("Lỗi tải phiên bản:", err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => { fetchVersions(); }, []);

    return { softwares, loading, refresh: fetchVersions };
};