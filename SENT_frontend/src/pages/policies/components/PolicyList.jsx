import React, { useState } from 'react';
import { Laptop, Globe} from 'lucide-react';
import { PolicyItem } from './PolicyShared';
import Pagination from '../../../components/common/Pagination'; // Import

const PolicyList = ({ policies, onDelete, assets, currentConfig }) => {
    // Thêm State phân trang cho phần Specific
    const [specPage, setSpecPage] = useState(1);
    const itemsPerPage = 5;

    // Lọc dữ liệu
    const specificPolicies = policies.filter(p => p.target_type === 'SPECIFIC');
    const globalPolicies = policies.filter(p => p.target_type === 'GLOBAL');

    // Tính toán phân trang cho Specific
    const specTotalPages = Math.ceil(specificPolicies.length / itemsPerPage);
    const displaySpecific = specificPolicies.slice((specPage - 1) * itemsPerPage, specPage * itemsPerPage);

    return (
        <div className="space-y-6">
            {/* --- KHU VỰC LUẬT RIÊNG (SPECIFIC) - CÓ PHÂN TRANG --- */}
            {specificPolicies.length > 0 && (
                <div className="bg-[#1e293b] rounded-3xl border border-purple-500/30 shadow-xl overflow-hidden">
                    <div className="p-4 bg-purple-500/10 border-b border-purple-500/20 flex items-center justify-between">
                        <h3 className="font-bold text-purple-400 flex items-center gap-2 text-sm uppercase tracking-wider">
                            <Laptop size={18}/> Ngoại lệ & Ghi đè (Specific)
                        </h3>
                        <span className="text-[10px] font-bold bg-purple-500 text-white px-2 py-0.5 rounded">Ưu tiên cao</span>
                    </div>
                    <div className="divide-y divide-slate-800">
                        {displaySpecific.map(p => (
                            <PolicyItem key={p.ID} policy={p} onDelete={onDelete} assets={assets} currentConfig={currentConfig} />
                        ))}
                    </div>
                    
                    {/* Thêm thanh phân trang vào đây */}
                    <div className="p-4 bg-slate-900/50">
                        <Pagination 
                            currentPage={specPage} 
                            totalPages={specTotalPages} 
                            onPageChange={setSpecPage} 
                        />
                    </div>
                </div>
            )}

            {/* --- KHU VỰC LUẬT CHUNG (GLOBAL) --- */}
            {/* (Phần này thường ít nên có thể giữ nguyên hoặc thêm phân trang tương tự nếu muốn) */}
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
                            <PolicyItem key={p.ID} policy={p} onDelete={onDelete} assets={assets} currentConfig={currentConfig} />
                        ))
                    )}
                </div>
            </div>
        </div>
    );
};

export default PolicyList;