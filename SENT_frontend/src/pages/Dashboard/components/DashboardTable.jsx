import React from 'react';

const DashboardTable = ({ title, icon: Icon, data, columns = [], colorClass, onRowClick }) => {
    const safeData = data || [];
    
    return (
        <div className="bg-[#0A101D] rounded-xl border border-slate-800 shadow-lg flex flex-col overflow-hidden h-[300px]">
            {/* TABLE HEADER */}
            <div className="p-3 border-b border-slate-800 bg-[#111827] flex justify-between items-center shrink-0">
                <h3 className="text-[10px] font-black text-slate-400 uppercase tracking-widest flex items-center gap-2">
                    {Icon && <Icon size={14} className={colorClass} />} {title}
                </h3>
                <span className="bg-[#050B14] border border-slate-700 px-2 py-0.5 rounded text-[9px] font-mono text-slate-500">
                    {safeData.length} RECORDS
                </span>
            </div>

            {/* TABLE BODY TRÀN VIỀN */}
            <div className="flex-1 overflow-x-auto custom-scrollbar">
                <table className="w-full text-left whitespace-nowrap">
                    <thead className="sticky top-0 bg-[#0A101D] z-10 border-b border-slate-800">
                        <tr>
                            {columns.map((col, idx) => (
                                <th key={idx} className="px-4 py-2 text-[9px] font-black uppercase text-slate-600 tracking-widest">
                                    {col.label}
                                </th>
                            ))}
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800/50">
                        {safeData.length === 0 ? (
                            <tr>
                                <td colSpan={columns.length} className="px-4 py-8 text-center text-[10px] font-mono font-bold text-slate-600 uppercase tracking-widest">
                                    NO DATA DETECTED
                                </td>
                            </tr>
                        ) : (
                            safeData.map((row, idx) => (
                                <tr 
                                    key={idx} 
                                    onClick={() => onRowClick && onRowClick(row)}
                                    className={`group transition-colors ${onRowClick ? 'cursor-pointer hover:bg-slate-800/40' : 'hover:bg-[#111827]'}`}
                                >
                                    {columns.map((col, colIdx) => (
                                        <td key={colIdx} className="px-4 py-2.5 text-[11px]">
                                            {col.render(row)}
                                        </td>
                                    ))}
                                </tr>
                            ))
                        )}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default DashboardTable;