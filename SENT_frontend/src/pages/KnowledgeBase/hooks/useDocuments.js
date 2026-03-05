import { useState, useEffect } from 'react';
import axios from '../../../api/axios';

export const useDocuments = () => {
    const [documents, setDocuments] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [isUploading, setIsUploading] = useState(false);

    // State cho tìm kiếm & phân trang
    const [searchQuery, setSearchQuery] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 8;

    // 1. LẤY DANH SÁCH TÀI LIỆU
    const fetchDocuments = async () => {
        setLoading(true);
        try {
            const res = await axios.get('/docs'); // Gọi đúng endpoint docs
            setDocuments(res.data || []);
            setError(null);
        } catch (err) {
            console.error("Lỗi tải tài liệu:", err);
            setError("Không thể tải danh sách tài liệu.");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchDocuments();
    }, []);

    // 2. UPLOAD TÀI LIỆU (MULTIPART/FORM-DATA)
    const uploadDoc = async (formData) => {
        setIsUploading(true);
        try {
            await axios.post('/docs/upload', formData, {
                headers: { 'Content-Type': 'multipart/form-data' }
            });
            await fetchDocuments(); // Tải lại danh sách sau khi up
            return { success: true };
        } catch (err) {
            console.error("Upload thất bại:", err);
            return { success: false, error: err.response?.data?.error || "Lỗi upload file" };
        } finally {
            setIsUploading(false);
        }
    };

    // 3. XÓA TÀI LIỆU
    const deleteDoc = async (id) => {
        if (!window.confirm("Bạn chắc chắn muốn xóa tài liệu này?")) return;
        try {
            await axios.delete(`/docs/${id}`);
            setDocuments(prev => prev.filter(d => d.ID !== id));
        } catch (err) {
            alert("Lỗi khi xóa tài liệu!");
        }
    };

    // 4. CẬP NHẬT THÔNG TIN
    const updateDoc = async (id, data) => {
        try {
            await axios.put(`/docs/${id}`, data);
            await fetchDocuments();
            return { success: true };
        } catch (err) {
            return { success: false, error: "Lỗi cập nhật" };
        }
    };

    // --- LOGIC XỬ LÝ DỮ LIỆU TẠI CLIENT ---
    const filteredDocs = documents.filter(doc => 
        doc.title?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        doc.category?.toLowerCase().includes(searchQuery.toLowerCase())
    );

    const indexOfLastItem = currentPage * itemsPerPage;
    const indexOfFirstItem = indexOfLastItem - itemsPerPage;
    const currentDocuments = filteredDocs.slice(indexOfFirstItem, indexOfLastItem);
    const totalPages = Math.ceil(filteredDocs.length / itemsPerPage);

    return {
        documents,
        currentDocuments, // Dữ liệu đã phân trang để hiển thị
        filteredDocs,
        loading,
        error,
        isUploading,
        searchQuery, setSearchQuery,
        currentPage, setCurrentPage, totalPages,
        indexOfFirstItem, indexOfLastItem,
        fetchDocuments,
        uploadDoc,
        deleteDoc,
        updateDoc
    };
};