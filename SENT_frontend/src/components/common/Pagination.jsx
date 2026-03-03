import React, { useState, useEffect } from 'react';
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react';

const Pagination = ({ currentPage, totalPages, onPageChange }) => {
    const [inputPage, setInputPage] = useState("");

    useEffect(() => setInputPage(""), [currentPage]);

    if (totalPages <= 1) return null;

    const handleJump = (e) => {
        e.preventDefault();
        const page = parseInt(inputPage);
        if (page >= 1 && page <= totalPages) {
            onPageChange(page);
        }
    };

    const getPageNumbers = () => {
        const pages = [];
        const maxVisible = 3; // Giảm xuống 3 số để gọn hơn

        if (totalPages <= maxVisible) {
            for (let i = 1; i <= totalPages; i++) pages.push(i);
        } else {
            if (currentPage <= 2) {
                pages.push(1, 2, '...', totalPages);
            } else if (currentPage >= totalPages - 1) {
                pages.push(1, '...', totalPages - 1, totalPages);
            } else {
                pages.push(1, '...', currentPage, '...', totalPages);
            }
        }
        return pages;
    };

    return (
        <div className="flex flex-col items-center gap-3 pt-3 mt-2 border-t border-slate-800/50 w-full">
            
            {/* HÀNG 1: THÔNG TIN TRANG & Ô NHẬP NHANH */}
            <div className="flex items-center justify-between w-full px-1">
                <div className="text-[10px] text-slate-500 font-bold uppercase">
                    Trang <span className="text-white">{currentPage}</span> / {totalPages}
                </div>
                
                {/* Form nhập số trang nhỏ gọn */}
                <form onSubmit={handleJump} className="flex items-center gap-1">
                    <span className="text-[9px] text-slate-600">Go to:</span>
                    <input
                        type="number"
                        min="1"
                        max={totalPages}
                        value={inputPage}
                        onChange={(e) => setInputPage(e.target.value)}
                        className="w-8 h-6 bg-slate-900 border border-slate-700 rounded text-center text-[10px] text-white outline-none focus:border-emerald-500 transition"
                    />
                </form>
            </div>

            {/* HÀNG 2: THANH ĐIỀU HƯỚNG CHÍNH */}
            <div className="flex items-center justify-center bg-slate-900/50 p-1 rounded-lg border border-slate-800 w-full">
                {/* Về đầu */}
                <button
                    onClick={() => onPageChange(1)}
                    disabled={currentPage === 1}
                    className="p-1.5 rounded-md text-slate-500 hover:text-white hover:bg-slate-800 disabled:opacity-20 transition"
                >
                    <ChevronsLeft size={12} />
                </button>
                
                {/* Trang trước */}
                <button
                    onClick={() => onPageChange(currentPage - 1)}
                    disabled={currentPage === 1}
                    className="p-1.5 rounded-md text-slate-500 hover:text-white hover:bg-slate-800 disabled:opacity-20 transition mr-1"
                >
                    <ChevronLeft size={12} />
                </button>

                {/* Dãy số trang */}
                <div className="flex items-center gap-1 px-2 border-x border-slate-800/50">
                    {getPageNumbers().map((page, index) => (
                        <button
                            key={index}
                            onClick={() => typeof page === 'number' && onPageChange(page)}
                            disabled={page === '...'}
                            className={`min-w-[24px] h-6 flex items-center justify-center rounded text-[10px] font-bold transition ${
                                page === currentPage
                                    ? 'bg-emerald-500 text-white shadow-sm'
                                    : page === '...'
                                    ? 'text-slate-600 cursor-default'
                                    : 'text-slate-400 hover:bg-slate-800 hover:text-white'
                            }`}
                        >
                            {page}
                        </button>
                    ))}
                </div>

                {/* Trang sau */}
                <button
                    onClick={() => onPageChange(currentPage + 1)}
                    disabled={currentPage === totalPages}
                    className="p-1.5 rounded-md text-slate-500 hover:text-white hover:bg-slate-800 disabled:opacity-20 transition ml-1"
                >
                    <ChevronRight size={12} />
                </button>

                {/* Đến cuối */}
                <button
                    onClick={() => onPageChange(totalPages)}
                    disabled={currentPage === totalPages}
                    className="p-1.5 rounded-md text-slate-500 hover:text-white hover:bg-slate-800 disabled:opacity-20 transition"
                >
                    <ChevronsRight size={12} />
                </button>
            </div>
        </div>
    );
};

export default Pagination;