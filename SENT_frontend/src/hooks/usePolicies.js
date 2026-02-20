import { useState, useEffect, useMemo } from 'react';
import axios from '../api/axios'; // Đảm bảo đường dẫn chuẩn

export const usePolicies = () => {
    const [documents, setDocuments] = useState([]);
    const [searchQuery, setSearchQuery] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const [isUploading, setIsUploading] = useState(false);
    const [editingDoc, setEditingDoc] = useState(null);
    const itemsPerPage = 5;

    // 1. Lấy dữ liệu
        const fetchDocs = async () => {
        try {
            const res = await axios.get('/docs'); // Đổi từ /knowledge hoặc /policies
            setDocuments(res.data || []);
        } catch (err) { console.error("Lỗi lấy danh sách tài liệu:", err); }
    };

    useEffect(() => { fetchDocs(); }, []);

    // 2. Logic Tìm kiếm & Phân trang tự động (Tự tính toán lại khi data hoặc search thay đổi)
    const filteredDocs = useMemo(() => {
        if (!searchQuery) return documents;
        return documents.filter(doc => 
            doc.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
            doc.category.toLowerCase().includes(searchQuery.toLowerCase())
        );
    }, [documents, searchQuery]);

    const totalPages = Math.ceil(filteredDocs.length / itemsPerPage);
    const indexOfLastItem = currentPage * itemsPerPage;
    const indexOfFirstItem = indexOfLastItem - itemsPerPage;
    const currentDocuments = filteredDocs.slice(indexOfFirstItem, indexOfLastItem);

    // Reset về trang 1 nếu người dùng đang gõ tìm kiếm
    useEffect(() => { setCurrentPage(1); }, [searchQuery]);

    // 3. Logic Thao tác API
    const uploadDoc = async (formData) => {
        try {
            await axios.post('/docs/upload', formData, { // Đổi route upload
                headers: { 'Content-Type': 'multipart/form-data' }
            });
            await fetchDocs();
            setCurrentPage(1); // Upload xong về trang 1
            return { success: true };
        } catch (err) {
            return { success: false, error: err.response?.data?.error || "Lỗi Server" };
        } finally { setIsUploading(false); }
    };

    const deleteDoc = async (id) => {
        if (!window.confirm("Bạn có chắc muốn xóa tài liệu này?")) return;
        try {
            await axios.delete(`/policies/${id}`);
            await fetchDocs();
            if (currentDocuments.length === 1 && currentPage > 1) setCurrentPage(currentPage - 1);
        } catch (err) { alert("Lỗi khi xóa!"); }
    };

    const updateDoc = async (id, data) => {
        try {
            await axios.put(`/policies/${id}`, data);
            setEditingDoc(null);
            await fetchDocs();
        } catch (err) { alert("Lỗi khi cập nhật!"); }
    };

    // Trả về tất cả state và hàm để giao diện sử dụng
    return {
        documents, filteredDocs, currentDocuments,
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages, indexOfFirstItem, indexOfLastItem,
        isUploading, editingDoc, setEditingDoc,
        uploadDoc, deleteDoc, updateDoc
    };
};