import { useState, useEffect, useCallback } from 'react';
import axios from '../../../api/axios';

export const useDocuments = () => {
    const [documents, setDocuments] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [isUploading, setIsUploading] = useState(false);

    const [searchQuery, setSearchQuery] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 8;

    const fetchDocuments = useCallback(async () => {
        setLoading(true);
        try {
            const res = await axios.get('/docs');
            setDocuments(res.data || []);
            setError(null);
        } catch (err) {
            console.error("Lỗi tải tài liệu:", err);
            setError("Không thể tải danh sách tài liệu.");
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchDocuments();
    }, [fetchDocuments]);

    const uploadDoc = async (formData) => {
        setIsUploading(true);
        try {
            await axios.post('/docs/upload', formData, {
                headers: { 'Content-Type': 'multipart/form-data' }
            });
            await fetchDocuments();
            return { success: true };
        } catch (err) {
            console.error("Upload thất bại:", err);
            return { success: false, error: err.response?.data?.error || "Lỗi upload file" };
        } finally {
            setIsUploading(false);
        }
    };

    const deleteDoc = async (id, reason) => {
        try {
            await axios.delete(`/docs/${id}`, {
                data: { reason }
            });
            await fetchDocuments();
            return { success: true };
        } catch (err) {
            console.error("Lỗi khi gửi yêu cầu xóa:", err);
            return { success: false, error: err.response?.data?.error || "Lỗi xóa tài liệu" };
        }
    };

    const updateDoc = async (id, data, reason) => {
        try {
            const formData = new FormData();
            formData.append('title', data.title);
            formData.append('category', data.category);
            formData.append('reason', reason); 
            if (data.file) formData.append('file', data.file);

            await axios.put(`/docs/${id}`, formData);
            await fetchDocuments();
            return { success: true };
        } catch (err) {
            console.error("Lỗi cập nhật:", err);
            return { success: false, error: err.response?.data?.error || "Lỗi cập nhật" };
        }
    };

    const downloadDoc = async (id, fileName) => {
        try {
            const response = await axios.get(`/docs/${id}/download`, {
                responseType: 'blob',
            });
            const url = window.URL.createObjectURL(new Blob([response.data]));
            const link = document.createElement('a');
            link.href = url;
            link.setAttribute('download', fileName);
            document.body.appendChild(link);
            link.click();
            link.remove();
            window.URL.revokeObjectURL(url);
        } catch (err) {
            console.error("Lỗi tải file gốc:", err);
            alert("Không thể tải file gốc!");
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
        currentDocuments,
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
        updateDoc,
        downloadDoc
    };
};