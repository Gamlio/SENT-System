import { useState, useEffect, useMemo } from 'react';
import axios from '../api/axios';

export const useAgents = () => {
    const [agents, setAgents] = useState([]);
    const [searchQuery, setSearchQuery] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 8; 


    const fetchAgents = async () => {
        try {
            const res = await axios.get('/agents');
            setAgents(res.data || []);
        } catch (err) {
            console.error("Lỗi lấy danh sách máy trạm:", err);
        }
    };
    // 1. Gọi API lấy dữ liệu Realtime
    useEffect(() => {
        fetchAgents();
        const interval = setInterval(fetchAgents, 30000); 
        return () => clearInterval(interval);
    }, []);
    
    // 2. Logic Tìm kiếm
    const filteredAgents = useMemo(() => {
        if (!searchQuery) return agents;
        const lowerQuery = searchQuery.toLowerCase();
        return agents.filter(agent => 
            (agent.hostname && agent.hostname.toLowerCase().includes(lowerQuery)) ||
            (agent.ip_address && agent.ip_address.includes(lowerQuery)) ||
            (agent.user_name && agent.user_name.toLowerCase().includes(lowerQuery)) ||
            (agent.user_phone && agent.user_phone.includes(lowerQuery))
        );
    }, [agents, searchQuery]);

    // 3. Logic Phân trang
    const totalPages = Math.ceil(filteredAgents.length / itemsPerPage);
    const indexOfLastItem = currentPage * itemsPerPage;
    const indexOfFirstItem = indexOfLastItem - itemsPerPage;
    const currentAgents = filteredAgents.slice(indexOfFirstItem, indexOfLastItem);

    // Tự động về trang 1 khi gõ tìm kiếm
    useEffect(() => { setCurrentPage(1); }, [searchQuery]);

    // 4. Logic Biểu đồ
    const onlineCount = agents.filter(a => a.status === 'online').length;
    const offlineCount = agents.length - onlineCount;
    
    const statusChartData = [
        { name: 'Online', value: onlineCount, color: '#10b981' },
        { name: 'Offline', value: offlineCount, color: '#64748b' }
    ];

    const osChartData = [
        { name: 'Win 10', count: agents.length > 0 ? Math.ceil(agents.length * 0.6) : 0 },
        { name: 'Win 11', count: agents.length > 0 ? Math.floor(agents.length * 0.3) : 0 },
        { name: 'macOS', count: agents.length > 0 ? Math.floor(agents.length * 0.1) : 0 },
    ];

    return {
        agents, filteredAgents, currentAgents,
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages, indexOfFirstItem, indexOfLastItem,
        statusChartData, osChartData, onlineCount, offlineCount, fetchAgents
    };
};

// Xuất riêng hàm tính thời gian để các component khác có thể tái sử dụng
export const getTimeAgo = (dateString, status) => {
    if (!dateString) return "Không rõ";
    const diffInSeconds = Math.floor((new Date() - new Date(dateString)) / 1000);
    if (status === 'online') {
        if (diffInSeconds < 60) return `Nhịp đập: vài giây trước`;
        return `Nhịp đập: ${Math.floor(diffInSeconds / 60)} phút trước`;
    }
    if (diffInSeconds < 60) return `Offline vài giây trước`;
    const mins = Math.floor(diffInSeconds / 60);
    if (mins < 60) return `Offline ${mins} phút trước`;
    const hours = Math.floor(mins / 60);
    if (hours < 24) return `Offline ${hours} giờ trước`;
    return `Offline ${Math.floor(hours / 24)} ngày trước`;
};