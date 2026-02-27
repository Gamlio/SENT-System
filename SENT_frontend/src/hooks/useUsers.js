import { useState, useEffect, useMemo } from 'react';
import axios from '../api/axios'; // Đã mở khóa import API thật

export const useUsers = () => {
    const [users, setUsers] = useState([]);
    const [searchQuery, setSearchQuery] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const [isLoading, setIsLoading] = useState(false);
    const itemsPerPage = 5;

    // 1. LẤY DỮ LIỆU TỪ BACKEND
    const fetchUsers = async () => {
        try {
            const res = await axios.get('/users'); 
            setUsers(res.data || []);
        } catch (err) { 
            console.error("Lỗi lấy danh sách người dùng:", err); 
        }
    };

    // Tự động lấy dữ liệu khi trang được load
    useEffect(() => { fetchUsers(); }, []);

    // 2. LOGIC TÌM KIẾM TỐI ƯU (Hỗ trợ tìm theo cả Username và Full Name)
    const filteredUsers = useMemo(() => {
        if (!searchQuery) return users;
        const query = searchQuery.toLowerCase();
        return users.filter(u => 
            (u.username && u.username.toLowerCase().includes(query)) ||
            (u.full_name && u.full_name.toLowerCase().includes(query))
        );
    }, [users, searchQuery]);

    const totalPages = Math.ceil(filteredUsers.length / itemsPerPage);
    const indexOfLastItem = currentPage * itemsPerPage;
    const indexOfFirstItem = indexOfLastItem - itemsPerPage;
    const currentUsers = filteredUsers.slice(indexOfFirstItem, indexOfLastItem);

    useEffect(() => { setCurrentPage(1); }, [searchQuery]);

    // 3. LOGIC TẠO NGƯỜI DÙNG THẬT
    const createUser = async (formData) => {
        setIsLoading(true);
        try {
            // Gửi dữ liệu (bao gồm họ tên, sđt, email) xuống Backend
            await axios.post('/users', formData);
            alert("Tạo tài khoản thành công!");
            fetchUsers(); // Tải lại danh sách mới
            setCurrentPage(1);
        } catch (err) {
            // Hiển thị lỗi từ Backend (VD: "Tên đăng nhập đã tồn tại")
            alert(err.response?.data?.error || "Lỗi khi tạo tài khoản!");
        } finally {
            setIsLoading(false);
        }
    };

    // 4. LOGIC XÓA NGƯỜI DÙNG THẬT
    const deleteUser = async (id) => {
        if (!window.confirm("Bạn có chắc muốn xóa nhân viên này khỏi hệ thống?")) return;
        try {
            await axios.delete(`/users/${id}`);
            fetchUsers(); // Tải lại danh sách sau khi xóa
        } catch (err) {
            alert("Lỗi khi xóa tài khoản!");
        }
    };

    // 5. LOGIC CẬP NHẬT NGƯỜI DÙNG THẬT
    const updateUser = async (id, updateData) => {
        setIsLoading(true);
        try {
            await axios.put(`/users/${id}`, updateData);
            alert("Cập nhật thông tin thành công!");
            fetchUsers(); // Tải lại danh sách
        } catch (err) {
            alert(err.response?.data?.error || "Lỗi khi cập nhật!");
        } finally {
            setIsLoading(false);
        }
    };

    return {
        users, currentUsers, filteredUsers,
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages, indexOfFirstItem, indexOfLastItem,
        isLoading, createUser, deleteUser, updateUser
    };
};