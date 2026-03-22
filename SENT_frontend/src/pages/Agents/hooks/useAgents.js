import { useState, useEffect, useMemo } from 'react';
import axios from '../../../api/axios';

export const useAgents = () => {
    const [agents, setAgents] = useState([]);
    const [searchQuery, setSearchQuery] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    
    // 1. THÊM STATE SẮP XẾP
    // Mặc định: Mới nhất (last_seen giảm dần)
    const [sortConfig, setSortConfig] = useState({ key: 'last_seen', direction: 'desc' });
    
    const itemsPerPage = 8; 

    const fetchAgents = async () => {
        try {
            const res = await axios.get('/agents');
            setAgents(res.data || []);
        } catch (err) {
            console.error("Lỗi lấy danh sách máy trạm:", err);
        }
    };

    useEffect(() => {
        fetchAgents();
        const interval = setInterval(fetchAgents, 30000); 
        return () => clearInterval(interval);
    }, []);
    
    // --- PIPELINE XỬ LÝ DỮ LIỆU ---
    const processedAgents = useMemo(() => {
        // A. Lọc tìm kiếm
        let result = [...agents]; // Tạo bản sao để không ảnh hưởng state gốc

        if (searchQuery) {
            const lowerQuery = searchQuery.toLowerCase();
            result = result.filter(agent => 
                (agent.hostname && agent.hostname.toLowerCase().includes(lowerQuery)) ||
                (agent.ip_address && agent.ip_address.includes(lowerQuery)) ||
                (agent.manager?.full_name && agent.manager.full_name.toLowerCase().includes(lowerQuery)) ||
                (agent.manager?.phone && agent.manager.phone.includes(lowerQuery))
            );
        }

        // B. Sắp xếp (Sorting)
        if (sortConfig.key) {
            result.sort((a, b) => {
                let aValue = a[sortConfig.key];
                let bValue = b[sortConfig.key];

                // Xử lý riêng cho cột Manager (Object lồng nhau)
                if (sortConfig.key === 'manager') {
                    aValue = a.manager?.full_name || '';
                    bValue = b.manager?.full_name || '';
                }

                // Xử lý riêng cho IP Address (Chuyển về số để so sánh chuẩn)
                if (sortConfig.key === 'ip_address') {
                    const ipToNum = (ip) => {
                        if (!ip || ip === '::1' || ip === '127.0.0.1') return 0;
                        return Number(ip.split('.').map(d => ("000" + d).slice(-3)).join(""));
                    };
                    aValue = ipToNum(aValue);
                    bValue = ipToNum(bValue);
                }

                // Xử lý riêng cho Thời gian
                if (sortConfig.key === 'last_seen') {
                    aValue = new Date(aValue).getTime();
                    bValue = new Date(bValue).getTime();
                }

                if (aValue < bValue) return sortConfig.direction === 'asc' ? -1 : 1;
                if (aValue > bValue) return sortConfig.direction === 'asc' ? 1 : -1;
                return 0;
            });
        }

        return result;
    }, [agents, searchQuery, sortConfig]); // Khi agents, search, hoặc sort thay đổi thì tính toán lại

    // C. Phân trang (Pagination) dựa trên danh sách đã sắp xếp
    const totalPages = Math.ceil(processedAgents.length / itemsPerPage);
    const indexOfLastItem = currentPage * itemsPerPage;
    const indexOfFirstItem = indexOfLastItem - itemsPerPage;
    const currentAgents = processedAgents.slice(indexOfFirstItem, indexOfLastItem);

    // Reset về trang 1 khi thay đổi bộ lọc
    useEffect(() => { setCurrentPage(1); }, [searchQuery, sortConfig, agents.length]);

    // D. Dữ liệu biểu đồ
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
        currentAgents, // Dữ liệu đã được Lọc + Sắp xếp + Cắt trang
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages,
        sortConfig, setSortConfig, // Xuất hàm setSortConfig ra ngoài để UI dùng
        statusChartData, osChartData, onlineCount, offlineCount, fetchAgents
    };
};

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