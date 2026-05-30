import React, { useState, useEffect, useCallback, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Cpu, ArrowLeft, User, ShieldCheck, Fingerprint, Tag, Network, Monitor, AlertTriangle, ShieldAlert, Zap, Usb, Package, PlusCircle } from 'lucide-react';
import axios from '../../api/axios';

import AssetsUSB from './components/AssetsUSB';
import AssetsSoftware from './components/AssetsSoftware';
import AssetsLogs from './components/AssetsLogs';
import AssetsPort from './components/AssetsPort';
import AssetsBaseline from './components/AssetsBaseline';
import { useSocketSubscription } from '../../context/useSocketSubscription';

const AssetDetail = () => {
    const { hwid } = useParams();
    const navigate = useNavigate();
    const [asset, setasset] = useState(null);
    const [logs, setLogs] = useState([]);
    const [activeTab, setActiveTab] = useState('LOGS'); // Tabs điều hướng khung phải
    const [counts, setCounts] = useState({ software: 0, usb: 0, ports: 0 });
    const lastFetchTime = useRef(0);

    const fetchDetail = useCallback(async (force = false) => {
        const now = Date.now();
        if (!force && now - lastFetchTime.current < 2000) return;
        try {
            const res = await axios.get(`/assets/${hwid}`);
            setasset(res.data);
            const logRes = await axios.get(`/assets/${hwid}/logs`);
            setLogs(logRes.data);

            // Tách riêng truy vấn đếm số lượng (Total Count) trực tiếp từ MongoDB 
            // giúp độc lập hoàn toàn khỏi mảng dữ liệu đính kèm, tăng tốc Backend
            const [swRes, usbRes, portRes] = await Promise.all([
                axios.get(`/assets/${hwid}/software?limit=1`).catch(() => ({ data: { total: 0 } })),
                axios.get(`/assets/${hwid}/usb?limit=1`).catch(() => ({ data: { total: 0 } })),
                axios.get(`/assets/${hwid}/ports?limit=1`).catch(() => ({ data: { total: 0 } }))
            ]);
            setCounts({
                software: swRes.data.total || 0,
                usb: usbRes.data.total || 0,
                ports: portRes.data.total || 0
            });
            lastFetchTime.current = Date.now();
        } catch (err) { console.error(err); }
    }, [hwid]);

    useEffect(() => { fetchDetail(true); }, [fetchDetail]);
    
    useSocketSubscription(['REFRESH_DATA', 'ASSET_UPDATE'], (data) => { 
        if (data?.hwid === hwid || data?.asset_hwid === hwid) fetchDetail(); 
    });

    const handleUpdateDepartment = async (tag) => {
        try {
            await axios.put(`/assets/${hwid}/department`, { department_tag: tag });
            fetchDetail(true); 
        } catch (err) { alert("Lỗi cập nhật phòng ban!"); }
    };

    if (!asset) return (
        <div className="h-screen bg-[#050B14] flex flex-col items-center justify-center gap-4">
            <Zap size={40} className="text-indigo-500 animate-pulse"/>
            <p className="text-indigo-500 font-mono text-sm tracking-widest animate-pulse uppercase">Retrieving Asset Profile...</p>
        </div>
    );
    
    const inv = asset.inventory || {}; 
    // [SỬA LỖI LOGIC] Tính toán trạng thái Online dựa trên last_seen thay vì trường status của DB
    const isOnline = asset.last_seen ? (new Date() - new Date(asset.last_seen)) < 120000 : false; // 2 phút

    return (
        <div className="h-[calc(100vh-60px)] bg-[#050B14] text-slate-200 flex flex-col font-sans overflow-hidden">
            
            {/* --- HEADER TỔNG --- */}
            <div className="h-14 border-b border-slate-800 bg-[#0A101D] px-6 flex justify-between items-center shrink-0 z-10 shadow-md">
                <div className="flex items-center gap-4">
                    <button onClick={() => navigate('/assets')} className="text-slate-500 hover:text-white transition"><ArrowLeft size={16}/></button>
                    <div className="h-6 w-px bg-slate-800"></div>
                    <Monitor size={16} className={isOnline ? 'text-emerald-400' : 'text-slate-500'}/>
                    <h1 className="text-sm font-black text-white uppercase tracking-widest">{asset.hostname}</h1>
                    <span className="text-[10px] text-slate-500 font-mono bg-[#050B14] px-1.5 py-0.5 rounded border border-slate-800">{asset.asset_hwid}</span>
                </div>
                
                <div className="flex items-center gap-3">
                    
                    <div className={`px-2 py-0.5 rounded text-[10px] font-black uppercase tracking-widest border flex items-center gap-1.5 ${isOnline ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30 shadow-[0_0_8px_rgba(16,185,129,0.2)]' : 'bg-slate-800 text-slate-500 border-slate-700'}`}>
                        <div className={`w-1.5 h-1.5 rounded-full ${isOnline ? 'bg-emerald-400 animate-pulse' : 'bg-slate-500'}`}></div> {isOnline ? 'ONLINE' : 'OFFLINE'}
                    </div>
                </div>
            </div>

            {/* --- BODY CHIA ĐÔI MÀN HÌNH --- */}
            <div className="flex-1 flex overflow-hidden">
                
                {/* NỬA TRÁI (35%): THÔNG TIN TỔNG QUAN DÀY ĐẶC */}
                <div className="w-[380px] h-full overflow-y-auto custom-scrollbar border-r border-slate-800 bg-[#0A101D]/50 p-5 flex flex-col gap-5 shrink-0">
                    
                    {/* RISK & TRUST SCORE TIGHT GRID */}
                    <div className="grid grid-cols-2 gap-3">
                        <div className="bg-[#111827] border border-slate-800 rounded-lg p-3 text-center shadow-inner">
                            <p className="text-[9px] font-black text-slate-500 uppercase tracking-widest mb-1">Risk Score</p>
                            <div className="flex items-baseline justify-center gap-0.5">
                                <span className={`text-3xl font-black tracking-tighter ${asset.risk_score > 70 ? 'text-red-500' : asset.risk_score > 30 ? 'text-amber-500' : 'text-emerald-500'}`}>{asset.risk_score || 0}</span>
                                <span className="text-[10px] text-slate-600 font-bold">/100</span>
                            </div>
                        </div>
                        <div className="bg-[#111827] border border-slate-800 rounded-lg p-3 text-center shadow-inner relative overflow-hidden">
                            <ShieldCheck size={40} className="absolute -right-3 -bottom-3 text-emerald-500/5 rotate-12" />
                            <p className="text-[9px] font-black text-slate-500 uppercase tracking-widest mb-1 relative z-10">Trust Score</p>
                            <div className="flex items-baseline justify-center gap-0.5 relative z-10">
                                <span className={`text-3xl font-black tracking-tighter ${(asset.trust_score ?? 100) >= 80 ? 'text-emerald-500' : (asset.trust_score ?? 100) >= 50 ? 'text-yellow-500' : 'text-red-500'}`}>{asset.trust_score ?? 100}</span>
                                <span className="text-[10px] text-slate-600 font-bold">/100</span>
                            </div>
                        </div>
                    </div>

                    {/* TAGS & QUẢN LÝ PHÒNG BAN */}
                    <div>
                        <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-2 flex items-center gap-1.5 border-b border-slate-800 pb-1.5"><Tag size={12}/> Security Context</h3>
                        <div className="bg-[#111827] border border-slate-800 rounded-lg p-3 flex flex-col gap-2">
                            <div className="flex justify-between items-center">
                                <span className="text-[9px] font-bold text-slate-400 uppercase">Department / Role</span>
                                <select 
                                    value={asset.department_tag || 'OFFICE'} 
                                    onChange={(e) => handleUpdateDepartment(e.target.value)}
                                    className="bg-[#050B14] text-[10px] font-black uppercase tracking-widest text-indigo-400 px-2 py-1 rounded border border-slate-700 outline-none cursor-pointer"
                                >
                                    <option value="OFFICE">OFFICE (Default)</option>
                                    <option value="DEV">DEV (IT/Code)</option>
                                    <option value="FINANCE">FINANCE (Risk: High)</option>
                                    <option value="PROD">PROD (Server)</option>
                                </select>
                            </div>
                            <div className="flex justify-between items-center">
                                <span className="text-[9px] font-bold text-slate-400 uppercase">Device Tier</span>
                                <span className="text-[10px] font-black uppercase text-slate-300">{asset.device_type || 'OFFICE'}</span>
                            </div>
                        </div>
                    </div>

                    {/* THÔNG SỐ CẤU HÌNH (SPECS) */}
                    <div>
                        <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-2 flex items-center gap-1.5 border-b border-slate-800 pb-1.5"><Cpu size={12}/> Hardware Specs</h3>
                        <div className="bg-[#111827] border border-slate-800 rounded-lg p-3 space-y-2">
                            <InfoRow label="IP Address" value={asset.ip_address} icon={<Network size={10} className="text-emerald-500"/>} />
                            <InfoRow label="CPU" value={inv.cpu_model} />
                            <InfoRow label="RAM" value={`${inv.ram_total_gb} GB`} />
                            <InfoRow label="OS" value={inv.os_info} />
                        </div>
                    </div>

                    {/* NGƯỜI QUẢN LÝ */}
                    <div>
                        <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-2 flex items-center gap-1.5 border-b border-slate-800 pb-1.5"><User size={12}/> Assigned Owner</h3>
                        <div className="bg-[#111827] border border-slate-800 rounded-lg p-3">
                            {asset.manager ? (
                                <div className="flex items-center gap-3">
                                    <div className="w-8 h-8 rounded bg-slate-800 border border-slate-700 flex items-center justify-center text-slate-500"><User size={14}/></div>
                                    <div>
                                        <p className="text-xs font-bold text-white mb-0.5">{asset.manager.full_name}</p>
                                        <p className="text-[10px] text-slate-500 font-mono">@{asset.manager.username} • {asset.manager.phone}</p>
                                    </div>
                                </div>
                            ) : (
                                <p className="text-[10px] font-mono text-slate-500 italic">No owner assigned. Asset is orphaned.</p>
                            )}
                        </div>
                    </div>

                </div>

                {/* NỬA PHẢI (65%): TABS CHI TIẾT DỮ LIỆU */}
                <div className="flex-1 flex flex-col bg-[#050B14] relative border-l border-slate-800/50 min-w-0">
                    
                    {/* TABS HEADER TRÀN NGANG */}
                    <div className="flex border-b border-slate-800 bg-[#0A101D] shrink-0 overflow-x-auto custom-scrollbar">
                        <TabButton active={activeTab === 'LOGS'} onClick={() => setActiveTab('LOGS')} icon={ShieldAlert} label={`Alert Logs (${logs?.length})`} color="text-red-400" />
                        <TabButton active={activeTab === 'SOFTWARE'} onClick={() => setActiveTab('SOFTWARE')} icon={Package} label={`Software (${counts.software})`} color="text-blue-400" />
                        <TabButton active={activeTab === 'USB'} onClick={() => setActiveTab('USB')} icon={Usb} label={`USB History (${counts.usb})`} color="text-emerald-400" />
                        <TabButton active={activeTab === 'PORTS'} onClick={() => setActiveTab('PORTS')} icon={Network} label={`Open Ports (${counts.ports})`} color="text-amber-400" />
                        <TabButton active={activeTab === 'BASELINE'} onClick={() => setActiveTab('BASELINE')} icon={PlusCircle} label={`Proposed Baseline`} color="text-indigo-400" />
                    </div>

                    {/* TABS CONTENT (KHUNG ĐỦ TO ĐỂ CHỨA CÁC COMPONENT CON) */}
                    <div className="flex-1 overflow-hidden p-4">
                        {/* Mình bọc css cho các thẻ con bên trong giãn 100% height */}
                        <div className="h-full w-full [&>div]:h-full [&>div]:border-0 [&>div]:bg-transparent [&>div]:shadow-none">
                            {activeTab === 'LOGS' && <AssetsLogs logs={logs} />}
                            {activeTab === 'SOFTWARE' && <AssetsSoftware hwid={hwid} />}
                            {activeTab === 'USB' && <AssetsUSB hwid={hwid} />}
                            {activeTab === 'PORTS' && <AssetsPort hwid={hwid} />}
                            {activeTab === 'BASELINE' && <AssetsBaseline hwid={hwid} onRefresh={() => fetchDetail(true)} />}
                        </div>
                    </div>
                </div>

            </div>
        </div> 
    );
};

// --- COMPONENT NHỎ HỖ TRỢ HIỂN THỊ ---
const InfoRow = ({ label, value, icon }) => (
    <div className="flex justify-between items-center text-[10px]">
        <span className="text-slate-500 font-bold uppercase tracking-wider">{label}</span>
        <span className="text-slate-300 font-mono truncate max-w-[60%] flex items-center gap-1.5">{icon} {value || 'N/A'}</span>
    </div>
);

const TabButton = ({ active, onClick, icon: Icon, label, color }) => (
    <button 
        onClick={onClick}
        className={`flex items-center gap-2 px-6 py-3 text-[10px] font-black uppercase tracking-widest border-b-2 transition-colors whitespace-nowrap ${
            active ? `border-indigo-500 text-white bg-slate-800/30` : `border-transparent text-slate-500 hover:text-slate-300 hover:bg-slate-800/10`
        }`}
    >
        <Icon size={14} className={active ? color : 'text-slate-500'}/> {label}
    </button>
);

export default AssetDetail;