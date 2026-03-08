import React, { useEffect, useState } from 'react';
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react';

const Pagination = ({ currentPage, totalPages, onPageChange }) => {
    const [inputPage, setInputPage] = useState("");

    // Reset input khi trang đổi
    useEffect(() => setInputPage(""), [currentPage]);

    if (totalPages <= 1) return null;

    // Logic hiển thị số trang thông minh (1 ... 4 5 6 ... 10)
    const getPageNumbers = () => {
        const pages = [];
        const maxVisible = 1; // Số trang hiện bên cạnh trang hiện tại

        // Luôn hiện trang 1
        pages.push(1);

        // Logic dấu ... đầu
        if (currentPage > maxVisible + 2) {
            pages.push('...');
        } else if (currentPage > 2) {
            // Lấp lỗ hổng nếu khoảng cách nhỏ (VD: 1, 2, [3]...)
            for (let i = 2; i < currentPage - maxVisible; i++) pages.push(i);
        }

        // Các trang xung quanh current
        for (let i = Math.max(2, currentPage - maxVisible); i <= Math.min(totalPages - 1, currentPage + maxVisible); i++) {
            pages.push(i);
        }

        // Logic dấu ... cuối
        if (currentPage < totalPages - maxVisible - 1) {
            pages.push('...');
        } else if (currentPage < totalPages - 1) {
            for (let i = currentPage + maxVisible + 1; i < totalPages; i++) pages.push(i);
        }

        // Luôn hiện trang cuối
        if (totalPages > 1) pages.push(totalPages);

        // Lọc trùng lặp (Set) và sort lại cho chắc chắn
        return [...new Set(pages)].sort((a, b) => (typeof a === 'number' && typeof b === 'number' ? a - b : 0));
    };

    return (
        <div className="flex flex-wrap justify-between items-center gap-4 text-xs font-medium select-none">
            
            {/* Info text */}
            <span className="text-slate-500">
                Trang <span className="text-white font-bold">{currentPage}</span> / {totalPages}
            </span>

            <div className="flex items-center gap-1">
                {/* Nút về đầu */}
                <button
                    onClick={() => onPageChange(1)}
                    disabled={currentPage === 1}
                    className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 disabled:opacity-30 disabled:cursor-not-allowed transition"
                >
                    <ChevronsLeft size={16} />
                </button>

                {/* Nút lùi */}
                <button
                    onClick={() => onPageChange(currentPage - 1)}
                    disabled={currentPage === 1}
                    className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 disabled:opacity-30 disabled:cursor-not-allowed transition mr-1"
                >
                    <ChevronLeft size={16} />
                </button>

                {/* Danh sách số trang */}
                <div className="flex items-center gap-1">
                    {getPageNumbers().map((page, index) => (
                        <button
                            key={index}
                            onClick={() => typeof page === 'number' && onPageChange(page)}
                            disabled={page === '...'}
                            className={`min-w-[28px] h-7 px-1 flex items-center justify-center rounded-lg transition-all ${
                                page === currentPage
                                    ? 'bg-emerald-500 text-white shadow-lg shadow-emerald-500/20 font-bold'
                                    : page === '...'
                                    ? 'text-slate-600 cursor-default'
                                    : 'text-slate-400 hover:bg-slate-800 hover:text-white border border-transparent hover:border-slate-700'
                            }`}
                        >
                            {page}
                        </button>
                    ))}
                </div>

                {/* Nút tiến */}
                <button
                    onClick={() => onPageChange(currentPage + 1)}
                    disabled={currentPage === totalPages}
                    className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 disabled:opacity-30 disabled:cursor-not-allowed transition ml-1"
                >
                    <ChevronRight size={16} />
                </button>

                {/* Nút về cuối */}
                <button
                    onClick={() => onPageChange(totalPages)}
                    disabled={currentPage === totalPages}
                    className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 disabled:opacity-30 disabled:cursor-not-allowed transition"
                >
                    <ChevronsRight size={16} />
                </button>
            </div>
        </div>
    );
};

export default Pagination;