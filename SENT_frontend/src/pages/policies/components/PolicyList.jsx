import React, { useState } from 'react';
import { Users, Globe, Trash2, X } from 'lucide-react';
import { PolicyItem } from './PolicyShared';
import Pagination from '../../../components/common/Pagination';
import BulkDeleteModal from '../../../components/common/BulkDeleteModal';

const PolicyList = ({ policies, onDelete, onBulkDelete, groups, currentConfig }) => {
    // Thêm State phân trang cho phần Specific
    const [specPage, setSpecPage] = useState(1);
    const itemsPerPage = 5;

    // State chọn nhiều (Bulk Selection)
    const [selectedIds, setSelectedIds] = useState([]);
    const [isBulkDeleteModalOpen, setIsBulkDeleteModalOpen] = useState(false);

    // Lọc dữ liệu
    const specificPolicies = policies.filter(p => p.target_type === 'GROUP' || p.group_id);
    const globalPolicies = policies.filter(p => p.target_type === 'GLOBAL');

    // Tính toán phân trang cho Specific
    const specTotalPages = Math.ceil(specificPolicies.length / itemsPerPage);
    const displaySpecific = specificPolicies.slice((specPage - 1) * itemsPerPage, specPage * itemsPerPage);

    const handleSelect = (id) => {
        setSelectedIds(prev => 
            prev.includes(id) ? prev.filter(item => item !== id) : [...prev, id]
        );
    };

    const handleConfirmBulkDelete = async (reason) => {
        setIsBulkDeleteModalOpen(false);
        if (onBulkDelete) {
            await onBulkDelete({ ids: selectedIds, reason });
        }
        setSelectedIds([]);
    };

    const clearSelection = () => setSelectedIds([]);

    const selectedPolicies = policies.filter(p => selectedIds.includes(p.ID));

    return (
        <div className="space-y-6 relative pb-20">
            {/* --- KHU VỰC LUẬT RIÊNG (SPECIFIC) - CÓ PHÂN TRANG --- */}
            {specificPolicies.length > 0 && (
                <div className="bg-[#1e293b] rounded-3xl border border-purple-500/30 shadow-xl overflow-hidden">
                    <div className="p-4 bg-purple-500/10 border-b border-purple-500/20 flex items-center justify-between">
                        <h3 className="font-bold text-purple-400 flex items-center gap-2 text-sm uppercase tracking-wider">
                            <Users size={18}/> Chính sách Nhóm (Group)
                        </h3>
                        <span className="text-[10px] font-bold bg-purple-500 text-white px-2 py-0.5 rounded">Ưu tiên cao</span>
                    </div>
                    <div className="divide-y divide-slate-800">
                        {displaySpecific.map(p => (
                            <PolicyItem 
                                key={p.ID} 
                                policy={p} 
                                onDelete={onDelete} 
                                groups={groups} 
                                currentConfig={currentConfig} 
                                isSelected={selectedIds.includes(p.ID)}
                                onSelect={handleSelect}
                            />
                        ))}
                    </div>
                    
                    {/* Thêm thanh phân trang vào đây */}
                    {specTotalPages > 1 && (
                        <div className="p-4 bg-slate-900/50">
                            <Pagination 
                                currentPage={specPage} 
                                totalPages={specTotalPages} 
                                onPageChange={setSpecPage} 
                            />
                        </div>
                    )}
                </div>
            )}

            {/* --- KHU VỰC LUẬT CHUNG (GLOBAL) --- */}
            <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-xl overflow-hidden">
                <div className="p-4 bg-slate-800/50 border-b border-slate-800 flex items-center justify-between">
                    <h3 className="font-bold text-white flex items-center gap-2 text-sm uppercase tracking-wider">
                        <Globe size={18} className="text-blue-400"/> Chính sách Toàn cục (Global)
                    </h3>
                    <span className="text-[10px] font-bold bg-slate-700 text-slate-300 px-2 py-0.5 rounded">Mặc định</span>
                </div>
                <div className="divide-y divide-slate-800">
                    {globalPolicies.length === 0 ? (
                        <div className="p-10 text-center text-slate-500 text-sm italic">Chưa có luật chung nào.</div>
                    ) : (
                        globalPolicies.map(p => (
                            <PolicyItem 
                                key={p.ID} 
                                policy={p} 
                                onDelete={onDelete} 
                                groups={groups} 
                                currentConfig={currentConfig} 
                                isSelected={selectedIds.includes(p.ID)}
                                onSelect={handleSelect}
                            />
                        ))
                    )}
                </div>
            </div>

            {/* Bulk Actions Bar */}
            {selectedIds.length > 0 && (
                <div className="fixed bottom-10 left-1/2 -translate-x-1/2 bg-slate-800 text-white px-6 py-4 rounded-2xl shadow-2xl border border-indigo-500/50 flex items-center gap-6 z-50 animate-in slide-in-from-bottom-5 duration-300">
                    <span className="text-sm font-bold">
                        <span className="bg-indigo-500 px-2.5 py-1 rounded-lg mr-2 text-xs font-black">
                            {selectedIds.length}
                        </span> 
                        luật đã chọn
                    </span>

                    <button 
                        onClick={() => setIsBulkDeleteModalOpen(true)} 
                        className="flex items-center gap-2 bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/30 px-5 py-2.5 rounded-xl transition font-bold text-xs uppercase tracking-wider shadow-lg"
                    >
                        <Trash2 size={16}/> Gỡ bỏ ({selectedIds.length})
                    </button>

                    <button 
                        onClick={clearSelection} 
                        className="p-2 text-slate-400 hover:text-red-400 transition-colors"
                    >
                        <X size={20}/>
                    </button>
                </div>
            )}

            <BulkDeleteModal
                isOpen={isBulkDeleteModalOpen}
                onClose={() => setIsBulkDeleteModalOpen(false)}
                onConfirm={handleConfirmBulkDelete}
                items={selectedPolicies}
                type="POLICY"
            />
        </div>
    );
};

export default PolicyList;