import { useState, useEffect } from 'react';
import axios from '../../../api/axios';

export const useGroups = () => {
    const [groups, setGroups] = useState([]);

    const fetchGroups = async () => {
        try {
            const res = await axios.get('/groups'); // API tương ứng với HandleGetGroups
            setGroups(res.data.data || []);
        } catch (err) {
            console.error("Lỗi lấy danh sách nhóm:", err);
        }
    };

    useEffect(() => { fetchGroups(); }, []);

    return { groups };
};