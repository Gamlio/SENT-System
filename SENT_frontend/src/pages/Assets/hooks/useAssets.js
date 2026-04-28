import { useState, useEffect, useCallback } from 'react';
import axios from '../../../api/axios';
import { useSocketSubscription } from '../../../context/useSocketSubscription'; 


export const useassets = () => {
    const [assets, setassets] = useState([]);
    const [total, setTotal] = useState(0);
    const [searchQuery, setSearchQuery] = useState('');
    const [debouncedSearch, setDebouncedSearch] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const [sortConfig, setSortConfig] = useState({ key: 'last_seen', direction: 'desc' });
    const [stats, setStats] = useState({ online: 0, offline: 0, total: 0 });
    
    const itemsPerPage = 8; 

    // Xử lý Debounce cho ô Search (Đợi 500ms sau khi ngừng gõ mới gọi API)
    useEffect(() => {
        const timer = setTimeout(() => {
            setDebouncedSearch(searchQuery);
            setCurrentPage(1); // Về trang 1 khi thay đổi từ khóa
        }, 500);
        return () => clearTimeout(timer);
    }, [searchQuery]);

    // Bọc fetchassets vào useCallback để tránh re-render vô hạn
    const fetchassets = useCallback(async () => {
        try {
            // Gọi API với các tham số phân trang và tìm kiếm (Server-Side)
            const res = await axios.get('/assets', {
                params: {
                    page: currentPage,
                    limit: itemsPerPage,
                    search: debouncedSearch
                }
            });
            // Lọc bỏ các máy trạm đã bị xóa mềm (trạng thái RETIRED)
            const activeAssets = (res.data.items || []).filter(asset => asset.status !== 'RETIRED');
            setassets(activeAssets);
            setTotal(res.data.total || 0);

            // Lấy thêm thống kê tổng quan cho biểu đồ
            const statsRes = await axios.get('/assets/stats');
            setStats({
                total: statsRes.data.total || 0,
                online: statsRes.data.online || 0,
                offline: (statsRes.data.total || 0) - (statsRes.data.online || 0)
            });
        } catch (err) {
            console.error("Lỗi lấy danh sách máy trạm:", err);
        }
    }, [currentPage, debouncedSearch]);

    useEffect(() => {
        fetchassets();
        // Giữ lại fallback 60s phòng trường hợp WebSocket rớt mạng
        const interval = setInterval(fetchassets, 60000); 
        return () => clearInterval(interval);
    }, [fetchassets]);

    // [QUAN TRỌNG]: LẮNG NGHE WEBSOCKET TỪ BACKEND
    useSocketSubscription((msg) => {
        // Lắng nghe các lệnh làm mới danh sách (Máy mới đăng ký, Máy đổi trạng thái, Máy sập nguồn)
        if (msg.type === 'REFRESH_asset_LIST' || msg.type === 'asset_STATUS_CHANGED') {
            fetchassets();
        }
    });
    
    // C. Phân trang (Pagination) - Tính toán từ Total của Backend
    const totalPages = Math.ceil(total / itemsPerPage);

    // D. Dữ liệu biểu đồ
    const onlineCount = stats.online;
    const offlineCount = stats.offline;
    
    const statusChartData = [
        { name: 'Online', value: onlineCount, color: '#10b981' },
        { name: 'Offline', value: offlineCount, color: '#64748b' }
    ];

    const osChartData = [
        { name: 'Win 10', count: stats.total > 0 ? Math.ceil(stats.total * 0.6) : 0 },
        { name: 'Win 11', count: stats.total > 0 ? Math.floor(stats.total * 0.3) : 0 },
        { name: 'macOS', count: stats.total > 0 ? Math.floor(stats.total * 0.1) : 0 },
    ];

    return {
        currentassets: assets, // Trả trực tiếp dữ liệu từ backend đã phân trang
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages,
        sortConfig, setSortConfig, // Xuất hàm setSortConfig ra ngoài để UI dùng
        statusChartData, osChartData, onlineCount, offlineCount, fetchassets,
        assets // [FIX] Bổ sung biến assets để ngoài giao diện Assets.jsx hiển thị đúng Tổng số lượng
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