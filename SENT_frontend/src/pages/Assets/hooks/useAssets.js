import { useState, useEffect, useMemo, useCallback } from 'react';
import axios from '../../../api/axios';
import { useSocketSubscription } from '../../../context/useSocketSubscription'; 


export const useassets = () => {
    const [assets, setassets] = useState([]);
    const [searchQuery, setSearchQuery] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const [sortConfig, setSortConfig] = useState({ key: 'last_seen', direction: 'desc' });
    
    const itemsPerPage = 8; 

    // Bọc fetchassets vào useCallback để tránh re-render vô hạn
    const fetchassets = useCallback(async () => {
        try {
            const res = await axios.get('/assets');
            // Lọc bỏ các máy trạm đã bị xóa mềm (trạng thái RETIRED)
            const activeAssets = (res.data || []).filter(asset => asset.status !== 'RETIRED');
            setassets(activeAssets);
        } catch (err) {
            console.error("Lỗi lấy danh sách máy trạm:", err);
        }
    }, []);

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
    
    // --- PIPELINE XỬ LÝ DỮ LIỆU ---
    const processedassets = useMemo(() => {
        // A. Lọc tìm kiếm
        let result = [...assets]; // Tạo bản sao để không ảnh hưởng state gốc

        if (searchQuery) {
            const lowerQuery = searchQuery.toLowerCase();
            result = result.filter(asset => 
                (asset.hostname && asset.hostname.toLowerCase().includes(lowerQuery)) ||
                (asset.ip_address && asset.ip_address.includes(lowerQuery)) ||
                (asset.manager?.full_name && asset.manager.full_name.toLowerCase().includes(lowerQuery)) ||
                (asset.manager?.phone && asset.manager.phone.includes(lowerQuery))
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
    }, [assets, searchQuery, sortConfig]); // Khi assets, search, hoặc sort thay đổi thì tính toán lại

    // C. Phân trang (Pagination) dựa trên danh sách đã sắp xếp
    const totalPages = Math.ceil(processedassets.length / itemsPerPage);
    const indexOfLastItem = currentPage * itemsPerPage;
    const indexOfFirstItem = indexOfLastItem - itemsPerPage;
    const currentassets = processedassets.slice(indexOfFirstItem, indexOfLastItem);

    // Reset về trang 1 khi thay đổi bộ lọc
    useEffect(() => { setCurrentPage(1); }, [searchQuery, sortConfig, assets.length]);

    // D. Dữ liệu biểu đồ
    const onlineCount = assets.filter(a => a.status === 'online').length;
    const offlineCount = assets.length - onlineCount;
    
    const statusChartData = [
        { name: 'Online', value: onlineCount, color: '#10b981' },
        { name: 'Offline', value: offlineCount, color: '#64748b' }
    ];

    const osChartData = [
        { name: 'Win 10', count: assets.length > 0 ? Math.ceil(assets.length * 0.6) : 0 },
        { name: 'Win 11', count: assets.length > 0 ? Math.floor(assets.length * 0.3) : 0 },
        { name: 'macOS', count: assets.length > 0 ? Math.floor(assets.length * 0.1) : 0 },
    ];

    return {
        currentassets, // Dữ liệu đã được Lọc + Sắp xếp + Cắt trang
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