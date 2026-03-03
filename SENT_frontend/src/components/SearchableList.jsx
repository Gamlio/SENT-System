import React, { useState, useEffect } from 'react';
import { Search, ChevronLeft, ChevronRight } from 'lucide-react';

const SearchableList = ({ items, searchKeys, renderItem, title, icon: Icon, placeholder = "Tìm kiếm..." }) => {
    const [searchTerm, setSearchTerm] = useState("");
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 10;

    useEffect(() => { setCurrentPage(1); }, [searchTerm]);

    const filteredItems = items.filter(item => 
        searchKeys.some(key => item[key]?.toLowerCase().includes(searchTerm.toLowerCase()))
    );

    const totalPages = Math.ceil(filteredItems.length / itemsPerPage);
    const displayedItems = filteredItems.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);

    return (
        <div className="space-y-4">
            <div className="flex flex-col sm:flex-row justify-between items-center gap-4">
                <h3 className="text-[10px] font-black text-slate-500 uppercase flex items-center gap-2 tracking-widest">
                    <Icon size={14}/> {title} ({filteredItems.length})
                </h3>
                <div className="relative w-full sm:w-48">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" size={12}/>
                    <input 
                        type="text" placeholder={placeholder} value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        className="w-full pl-9 pr-4 py-1.5 bg-slate-900 border border-slate-800 rounded-xl text-[11px] text-white outline-none focus:border-emerald-500/50"
                    />
                </div>
            </div>
            
            <div className="bg-slate-900/50 rounded-2xl border border-slate-800 overflow-hidden flex flex-col min-h-[450px]">
                <div className="flex-1">
                    {displayedItems.length === 0 ? (
                        <p className="p-10 text-[11px] text-slate-600 italic text-center">Không có dữ liệu phù hợp</p>
                    ) : (
                        displayedItems.map((item, index) => renderItem(item, index))
                    )}
                </div>

                {totalPages > 1 && (
                    <div className="p-3 border-t border-slate-800 flex justify-between items-center bg-slate-900/80">
                        <span className="text-[9px] text-slate-600 font-bold">Trang {currentPage}/{totalPages}</span>
                        <div className="flex gap-1">
                            <button disabled={currentPage === 1} onClick={() => setCurrentPage(p => p - 1)} className="p-1.5 rounded-lg bg-slate-800 text-slate-400 disabled:opacity-20"><ChevronLeft size={14}/></button>
                            <button disabled={currentPage === totalPages} onClick={() => setCurrentPage(p => p + 1)} className="p-1.5 rounded-lg bg-slate-800 text-slate-400 disabled:opacity-20"><ChevronRight size={14}/></button>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default SearchableList;