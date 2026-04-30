import { useState, useEffect, useMemo } from 'react';
import axios from '../../../api/axios'; 

export const useUsers = () => {
    const [users, setUsers] = useState([]);
    const [groups, setGroups] = useState([]); // [MỚI] Lưu danh sách phòng ban
    const [searchQuery, setSearchQuery] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const [isLoading, setIsLoading] = useState(false);
    const itemsPerPage = 6;

    const fetchUsers = async () => {
        try {
            const res = await axios.get('/users'); 
            const approvedUsers = (res.data || []).filter(u => u.approval_status === 'APPROVED');
            setUsers(approvedUsers);
        } catch (err) { 
            console.error("Lỗi lấy danh sách người dùng:", err); 
        }
    };

    // [MỚI] Lấy danh sách phòng ban
    const fetchGroups = async () => {
        try {
            const res = await axios.get('/groups'); // Gọi HandleGetGroups
            setGroups(res.data.data || []);
        } catch (err) { console.error("Lỗi lấy groups:", err); }
    };

    useEffect(() => { 
        fetchUsers();
        fetchGroups();
    }, []);

    const filteredUsers = useMemo(() => {
        if (!searchQuery) return users;
        const query = searchQuery.toLowerCase();
        return users.filter(u => 
            (u.username && u.username.toLowerCase().includes(query)) ||
            (u.full_name && u.full_name.toLowerCase().includes(query)) ||
            (u.employee_id && u.employee_id.toLowerCase().includes(query)) // Hỗ trợ tìm theo ID
        );
    }, [users, searchQuery]);

    const totalPages = Math.ceil(filteredUsers.length / itemsPerPage);
    const indexOfLastItem = currentPage * itemsPerPage;
    const indexOfFirstItem = indexOfLastItem - itemsPerPage;
    const currentUsers = filteredUsers.slice(indexOfFirstItem, indexOfLastItem);

    const createUser = async (formData) => {
        setIsLoading(true);
        try {
            await axios.post('/users', formData);
            fetchUsers(); 
            setCurrentPage(1);
            return { success: true };
        } catch (err) {
            return { success: false, error: err.response?.data?.error || "Lỗi tạo tài khoản" };
        } finally {
            setIsLoading(false);
        }
    };

    const updateUser = async (id, updateData) => {
        setIsLoading(true);
        try {
            await axios.put(`/users/${id}`, updateData);
            fetchUsers();
            return { success: true };
        } catch (err) {
            return { success: false, error: err.response?.data?.error || "Lỗi cập nhật" };
        } finally {
            setIsLoading(false);
        }
    };

    const deleteUser = async (id) => {
        try {
            await axios.delete(`/users/${id}`);
            fetchUsers();
            return { success: true };
        } catch (err) {
            return { success: false };
        }
    };

    return {
        users, groups, fetchGroups, currentUsers, filteredUsers,
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages, indexOfFirstItem, indexOfLastItem,
        isLoading, createUser, deleteUser, updateUser 
    };
};