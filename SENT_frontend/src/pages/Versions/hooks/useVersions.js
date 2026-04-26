import { useState, useEffect } from 'react';
import axios from '../../../api/axios';

export const useVersions = () => {
    const [versions, setVersions] = useState([]);
    const [loading, setLoading] = useState(true);

    const fetchVersions = async () => {
        try {
            const res = await axios.get('/api/v1/versions');
            setVersions(res.data);
        } catch (err) {
            console.error("Lỗi tải phiên bản:", err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => { fetchVersions(); }, []);

    return { versions, loading, refresh: fetchVersions };
};