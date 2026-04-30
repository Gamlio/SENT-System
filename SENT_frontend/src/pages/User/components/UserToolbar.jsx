import React from 'react';
import { Search, UserPlus, FileSpreadsheet, Download, Layers } from 'lucide-react';

const UserToolbar = ({ 
    searchQuery, 
    setSearchQuery, 
    onAddUser, 
    canManageUsers,
    onManageGroups,
    canManageGroups
}) => {
    return (
        <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-6">
            {/* Cụm Tìm kiếm */}
            <div className="relative w-full md:w-96 group">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 group-focus-within:text-emerald-400 transition-colors" size={18}/>
                <input 
                    type="text" 
                    placeholder="Tìm theo Tên, Username hoặc ID..." 
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="w-full bg-[#1e293b] text-white pl-10 pr-4 py-2.5 rounded-xl border border-slate-700 focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 outline-none transition-all shadow-inner"
                />
            </div>

            {/* Cụm Nút Hành động */}
            {(canManageUsers || canManageGroups) && (
                <div className="flex items-center gap-3 w-full md:w-auto">
                    
                    {canManageGroups && (
                        <button 
                            onClick={onManageGroups}
                            className="flex-1 md:flex-none flex items-center justify-center gap-2 bg-slate-800 hover:bg-slate-700 text-slate-300 px-4 py-2.5 rounded-xl text-sm font-bold border border-slate-700 transition"
                        >
                            <Layers size={18}/> 
                            Phòng ban
                        </button>
                    )}

                    {canManageUsers && (
                        <button 
                            onClick={onAddUser}
                            className="flex-1 md:flex-none flex items-center justify-center gap-2 bg-emerald-600 hover:bg-emerald-500 text-white px-4 py-2.5 rounded-xl text-sm font-bold shadow-[0_0_15px_rgba(16,185,129,0.3)] transition"
                        >
                            <UserPlus size={18}/> 
                            Thêm Nhân Sự
                        </button>
                    )}
                </div>
            )}
        </div>
    );
};

export default UserToolbar;