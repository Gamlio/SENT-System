import { useState, useEffect, useMemo } from 'react';
// import axios from '../api/axios'; // Mở ra khi có API thật

export const useUsers = () => {
    const [users, setUsers] = useState([]);
    const [searchQuery, setSearchQuery] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const [isLoading, setIsLoading] = useState(false);
    const itemsPerPage = 5;

    // 1. Lấy dữ liệu (Tạm dùng dữ liệu mẫu để bạn test giao diện, khi có Backend sẽ thay bằng axios)
    const fetchUsers = async () => {
        try {
            setUsers([
                { id: 1, username: 'admin_master', role_level: 2, created_at: new Date().toISOString() },
                { id: 2, username: 'nguyen.vana', role_level: 1, created_at: new Date().toISOString() },
                { id: 3, username: 'tran.thib', role_level: 1, created_at: new Date(Date.now() - 86400000).toISOString() },
                { id: 4, username: 'le.vanc', role_level: 1, created_at: new Date().toISOString() },
                { id: 5, username: 'pham.thid', role_level: 1, created_at: new Date().toISOString() },
                { id: 6, username: 'hoang.vane', role_level: 1, created_at: new Date().toISOString() },
            ]);
        } catch (err) { console.error("Lỗi:", err); }
    };

    useEffect(() => { fetchUsers(); }, []);

    // 2. Logic Tìm kiếm & Phân trang tự động
    const filteredUsers = useMemo(() => {
        if (!searchQuery) return users;
        return users.filter(u => 
            u.username.toLowerCase().includes(searchQuery.toLowerCase())
        );
    }, [users, searchQuery]);

    const totalPages = Math.ceil(filteredUsers.length / itemsPerPage);
    const indexOfLastItem = currentPage * itemsPerPage;
    const indexOfFirstItem = indexOfLastItem - itemsPerPage;
    const currentUsers = filteredUsers.slice(indexOfFirstItem, indexOfLastItem);

    useEffect(() => { setCurrentPage(1); }, [searchQuery]);

    // 3. Logic thao tác (Thêm/Xóa)
    const createUser = async (formData) => {
        setIsLoading(true);
        setTimeout(() => { // Giả lập đợi API
            alert("Tạo tài khoản thành công!");
            fetchUsers();
            setCurrentPage(1);
            setIsLoading(false);
        }, 800);
    };

    const deleteUser = async (id) => {
        if (!window.confirm("Bạn có chắc muốn xóa nhân viên này?")) return;
        alert("Đã xóa thành công!");
        fetchUsers();
    };

    return {
        users, currentUsers, filteredUsers,
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages, indexOfFirstItem, indexOfLastItem,
        isLoading, createUser, deleteUser
    };
};