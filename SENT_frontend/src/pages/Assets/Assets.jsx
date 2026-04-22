import React, { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { 
    Search, Monitor, ShieldAlert, Smartphone, User, 
    ChevronRight, X, UserCheck, LayoutList, ArrowUpDown, Trash2,
    ShieldCheck, Laptop, Activity, Tag, Server, Briefcase, UserX
} from 'lucide-react'; 
import { useassets, getTimeAgo } from './hooks/useAssets'; 
import { useUsers } from '../User/hooks/useUsers'; 
import AssetsActions from './components/AssetsActions';
import GenerateTokenButton from './components/GenerateTokenButton';
import AssetsBulkActions from './components/AssetsBulkActions';import { useSocketSubscription } from '../../context/useSocketSubscription';
import axios from '../../api/axios';

const riskScoringLevels = [
    { label: 'Rủi Ro Thấp', value: 'Low', color: '#10b981', icon: <ShieldCheck /> },
    { label: 'Rủi Ro Trung Bình', value: 'Medium', color: '#f59e0b', icon: <Monitor /> },
    { label: 'Rủi Ro Cao', value: 'High', color: '#ef4444', icon: <ShieldAlert /> },
    { label: 'Nguy Hiểm', value: 'Critical', color: '#be123c', icon: <ShieldAlert /> },
];

const assetStatusTag = React.memo(({ status }) => {
    const isOnline = status === 'online';
    return (
        <div className={`px-2 py-0.5 rounded text-[9px] font-black uppercase tracking-widest border flex items-center gap-1.5 w-max ${
            isOnline ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30 shadow-[0_0_8px_rgba(16,185,129,0.2)]' : 'bg-slate-800 text-slate-500 border-slate-700'
        }`}>
            <div className={`w-1.5 h-1.5 rounded-full ${isOnline ? 'bg-emerald-400 animate-pulse' : 'bg-slate-500'}`}></div>
            {isOnline ? 'Online' : 'Offline'}
        </div>
    );
});

const RiskScoreDisplay = React.memo(({ score }) => {
    const level = riskScoringLevels.find(l => {
        if (l.value === 'Low' && score <= 30) return true;
        if (l.value === 'Medium' && score > 30 && score <= 70) return true;
        if (l.value === 'High' && score > 70 && score <= 90) return true;
        if (l.value === 'Critical' && score > 90) return true;
        return false;
    });

    if (!level) return <span className="text-slate-500 text-xs font-mono">{score}</span>;

    return (
        <div className="flex items-center gap-2">
            <span className={`text-lg font-black w-8 text-right font-mono ${score > 70 ? 'text-red-500' : score > 30 ? 'text-amber-500' : 'text-emerald-500'}`}>{score}</span>
            <div style={{ color: level.color, borderColor: `${level.color}40`, backgroundColor: `${level.color}10` }} className="flex items-center gap-1.5 text-[9px] font-black uppercase tracking-widest px-2 py-0.5 rounded border">
                {React.cloneElement(level.icon, { size: 10 })} {level.label}
            </div>
        </div>
    );
});

const assets = () => {
    const navigate = useNavigate();
    const {
        currentassets = [], searchQuery, setSearchQuery, 
        sortConfig, setSortConfig, assets = [], fetchassets
    } = useassets();

    const { users = [] } = useUsers();
    const [selectedassets, setSelectedassets] = useState([]);
    const [showAssignModal, setShowAssignModal] = useState(false);
    const [targetassetHwid, setTargetassetHwid] = useState(null);
    const [dialogConfig, setDialogConfig] = useState({ isOpen: false, type: '', hwids: [] });

    const debounceRef = useRef(null);
    const handleassetUpdate = useCallback(() => {
        if (debounceRef.current) clearTimeout(debounceRef.current);
        debounceRef.current = setTimeout(() => fetchassets(), 500);
    }, [fetchassets]);

    useSocketSubscription(['asset_STATUS_CHANGED', 'REFRESH_asset_LIST', 'BASELINE_COMPLETED'], handleassetUpdate);

    const handleAssignManager = async (userId) => {
        try {
            if (targetassetHwid) {
                await axios.put(`/assets/${targetassetHwid}/assign`, { user_id: userId });
            } else {
                await axios.post('/assets/bulk-assign', { hwids: selectedassets, user_id: userId });
            }
            setShowAssignModal(false);
            setSelectedassets([]);
            fetchassets();
        } catch (err) { alert("Lỗi phân công!"); }
    };

    const tableColumns = useMemo(() => [
        {
            key: 'hostname', label: 'Asset / IP', className: 'w-[30%]', 
            render: (a) => {
                const isOnline = a.status === 'online';
                let IconComponent = Monitor;
                let iconColor = isOnline ? 'text-emerald-400 bg-emerald-500/10 border-emerald-500/30' : 'text-slate-500 bg-slate-800 border-slate-700';
                
                if (a.device_type === 'SERVER') { IconComponent = Server; iconColor = isOnline ? 'text-purple-400 bg-purple-500/10 border-purple-500/30' : 'text-purple-900 bg-slate-800 border-slate-700'; }
                if (a.device_type === 'IT_ADMIN') { IconComponent = Briefcase; iconColor = isOnline ? 'text-blue-400 bg-blue-500/10 border-blue-500/30' : 'text-blue-900 bg-slate-800 border-slate-700'; }

                return (
                    <div className="flex items-center gap-3">
                        <div className={`p-2 rounded-lg border transition-colors ${iconColor}`}>
                            <IconComponent size={14} />
                        </div>
                        <div>
                            <h4 className="font-bold text-white text-xs truncate group-hover:text-indigo-400 transition-colors cursor-pointer" title={a.hostname}>
                                {a.hostname}
                            </h4>
                            <div className="flex items-center gap-2 mt-0.5">
                                <span className="text-[10px] font-mono text-slate-500">{a.ip_address}</span>
                                
                            </div>
                        </div>
                    </div>
                );
            }
        },
        {
            key: 'manager', label: 'Owner / Dept', className: 'w-[20%]',
            render: (a) => (
                <div>
                    <p className="text-[11px] font-bold text-slate-300 truncate">{a.manager?.full_name || 'Unassigned'}</p>
                    <p className="text-[9px] text-slate-500 uppercase font-black tracking-widest mt-0.5 border border-slate-700 bg-[#050B14] w-max px-1.5 py-0.5 rounded">
                        {a.department_tag || 'OFFICE'}
                    </p>
                </div>
            )
        },
        {
            key: 'trust', label: 'Trust', className: 'w-[15%]', sortable: true,
            render: (a) => {
                const trust = a.trust_score ?? 100;
                const color = trust >= 80 ? 'bg-emerald-500' : trust >= 50 ? 'bg-yellow-500' : 'bg-red-500';
                return (
                    <div className="w-24">
                        <span className={`text-[10px] font-bold font-mono ${trust >= 80 ? 'text-emerald-400' : trust >= 50 ? 'text-yellow-400' : 'text-red-400'}`}>{trust}/100</span>
                        <div className="w-full bg-[#050B14] rounded-full h-1 mt-1 border border-slate-800"><div className={`h-full rounded-full ${color}`} style={{ width: `${trust}%` }}></div></div>
                    </div>
                );
            }
        },
        {
            key: 'risk', label: 'Risk Score', className: 'w-[15%]', sortable: true,
            render: (a) => <RiskScoreDisplay score={a.risk_score} />
        },
        {
            key: 'last_seen', label: 'Status / Seen', className: 'w-[15%]', sortable: true,
            render: (a) => (
                <div className="flex flex-col gap-1">
                    <assetStatusTag status={a.status} />
                    <span className="text-slate-500 text-[9px] font-mono uppercase tracking-widest truncate">{getTimeAgo(a.last_seen, a.status)}</span>
                </div>
            )
        }
    ], []);

    return (
        <div className="p-6 h-[calc(100vh-60px)] flex flex-col text-slate-200 bg-[#050B14] font-sans">
            
            <div className="flex justify-between items-end mb-4 shrink-0">
                <div>
                    <h1 className="text-xl font-black text-white flex items-center gap-2 uppercase tracking-tight">
                        <Monitor className="text-indigo-500"/> ENDPOINT ASSETS
                    </h1>
                    <p className="text-[11px] text-slate-500 mt-1 uppercase tracking-widest font-bold">Total Enrolled Devices: {assets.length}</p>
                </div>
                <GenerateTokenButton />
            </div>

            <div className="flex-1 bg-[#0A101D] rounded-lg border border-slate-800 shadow-2xl flex flex-col overflow-hidden">
                <div className="p-3 border-b border-slate-800 bg-[#111827] flex justify-between items-center shrink-0">
                    <div className="relative w-72">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" size={14} />
                        <input type="text" placeholder="Search IP, Hostname..." value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)}
                            className="w-full bg-[#050B14] border border-slate-800 text-xs text-white rounded pl-9 pr-4 py-1.5 outline-none focus:border-indigo-500 font-mono transition-colors" />
                    </div>
                    <div className="flex items-center gap-2">
                        <Tag className="text-emerald-500" size={12}/> <span className="text-[10px] font-mono font-bold text-slate-400">ONLINE: {assets.filter(a=>a.status === 'online').length}</span>
                    </div>
                </div>

                {/* KHUNG BẢNG VỚI MIN-HEIGHT CHỐNG CẮT DROPDOWN */}
                <div className="overflow-x-auto custom-scrollbar flex-1 min-h-[450px] relative pb-20">
                    <table className="w-full text-left border-collapse whitespace-nowrap">
                        <thead className="sticky top-0 z-10 bg-[#111827]">
                            <tr className="border-b border-slate-800 text-[10px] uppercase tracking-widest text-slate-500">
                                <th className="p-3 w-10 text-center"><input type="checkbox" className="w-3.5 h-3.5 rounded border-slate-700 bg-[#050B14] accent-indigo-500" onChange={(e) => setSelectedassets(e.target.checked ? currentassets.map(a => a.asset_hwid) : [])}/></th>
                                {tableColumns.map(col => (
                                    <th key={col.key} className={`p-3 font-black ${col.className}`}>
                                        <div className="flex items-center gap-1.5 cursor-pointer hover:text-white" onClick={() => col.sortable && setSortConfig({ key: col.key, direction: sortConfig.direction === 'asc' ? 'desc' : 'asc' })}>
                                            {col.label} {col.sortable && <ArrowUpDown size={10}/>}
                                        </div>
                                    </th>
                                ))}
                                <th className="p-3 text-right font-black">ACTION</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-800/50">
                            {currentassets.map((asset) => (
                                <tr key={asset.hwid || asset.asset_hwid} className="hover:bg-slate-800/30 transition-colors cursor-pointer group" onClick={() => navigate(`/assets/${asset.hwid || asset.asset_hwid}`)}>
                                    <td className="p-3 text-center" onClick={e => e.stopPropagation()}>
                                        <input type="checkbox" checked={selectedassets.includes(asset.hwid || asset.asset_hwid)} onChange={() => {
                                            const id = asset.hwid || asset.asset_hwid;
                                            setSelectedassets(prev => prev.includes(id) ? prev.filter(itemId => itemId !== id) : [...prev, id]);
                                        }} className="w-3.5 h-3.5 rounded border-slate-700 bg-[#050B14] accent-indigo-500" />
                                    </td>
                                    {tableColumns.map(col => <td key={col.key} className={`p-3 ${col.className}`}>{col.render(asset)}</td>)}
                                    <td className="p-3 text-right" onClick={e => e.stopPropagation()}>
                                        <AssetsActions asset={asset} onRefresh={fetchassets} onOpenAssignModal={() => { setTargetassetHwid(asset.hwid || asset.asset_hwid); setShowAssignModal(true); }} />
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>

            {selectedassets.length > 0 && (
                <div className="fixed bottom-0 left-6 right-6 z-50 p-3 bg-[#0A101D] border-t border-slate-800 shadow-[0_-10px_30px_rgba(0,0,0,0.5)] flex justify-center">
                    <AssetsBulkActions selectedassets={selectedassets} assets={assets} clearSelection={() => setSelectedassets([])} onRefresh={fetchassets} onOpenApproveModal={() => { setTargetassetHwid(null); setShowAssignModal(true); }} onOpenDeleteModal={(hwids) => setDialogConfig({ isOpen: true, type: 'BULK_DELETE', hwids })}/>
                </div>
            )}

            {/* MODAL PHÂN CÔNG */}
            {showAssignModal && (
                <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-[100] flex items-center justify-center p-4">
                    <div className="bg-[#0A101D] w-full max-w-md rounded-xl border border-slate-700 shadow-2xl overflow-hidden">
                        <div className="p-4 border-b border-slate-800 flex justify-between items-center bg-[#111827]">
                            <h3 className="text-[11px] font-black uppercase tracking-widest text-white flex items-center gap-2"><UserCheck size={14} className="text-indigo-400"/> Phân công quản lý</h3>
                            <button onClick={() => setShowAssignModal(false)} className="text-slate-500 hover:text-white transition"><X size={16}/></button>
                        </div>
                        <div className="p-2 max-h-[300px] overflow-y-auto custom-scrollbar">
                            {users.map(u => (
                                <button key={u.id} onClick={() => handleAssignManager(u.id)} className="w-full flex items-center justify-between p-3 hover:bg-indigo-500/10 rounded-lg border border-transparent hover:border-indigo-500/30 transition-all group">
                                    <div className="flex items-center gap-3">
                                        <div className="p-1.5 bg-[#050B14] border border-slate-800 rounded text-slate-500 group-hover:text-indigo-400 transition-colors"><User size={14}/></div>
                                        <div className="text-left">
                                            <p className="text-xs font-bold text-white">{u.full_name}</p>
                                            <p className="text-[9px] text-slate-500 font-mono">@{u.username}</p>
                                        </div>
                                    </div>
                                    <ChevronRight size={14} className="text-slate-700 group-hover:text-indigo-400 transition-transform group-hover:translate-x-1"/>
                                </button>
                            ))}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default assets;