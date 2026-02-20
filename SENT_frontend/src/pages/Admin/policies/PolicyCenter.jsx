import React, { useState, useEffect, useCallback } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { ShieldCheck, List, Usb, Activity, Plus, Trash2, CheckCircle2, Globe, Laptop, LayoutDashboard, Clock } from 'lucide-react';
import axios from '../../../api/axios';

const PolicyCenter = () => {
    const location = useLocation();
    const navigate = useNavigate();
    const queryParams = new URLSearchParams(location.search);
    const activeTab = queryParams.get('tab') || 'overview'; // Mặc định vào trang Tổng quát

    const [policies, setPolicies] = useState([]);
    const [agents, setAgents] = useState([]);
    const [selectedHWIDs, setSelectedHWIDs] = useState([]);
    const [targetType, setTargetType] = useState('GLOBAL');
    
    const [newTitle, setNewTitle] = useState('');
    const [newValue, setNewValue] = useState('');
    const [policyType, setPolicyType] = useState('BLACKLIST');

    // Cấu hình các Tab
    const tabConfig = {
        overview: { label: 'Tổng quát', icon: <LayoutDashboard size={18}/>, category: '' },
        software: { label: 'Phần mềm', icon: <List size={18}/>, category: 'SOFTWARE', placeholder: 'VD: zalo.exe, telegram...' },
        usb: { label: 'Thiết bị USB', icon: <Usb size={18}/>, category: 'USB', placeholder: 'VD: VID_0951&PID_1666' },
        network: { label: 'Quy tắc Mạng', icon: <Activity size={18}/>, category: 'NETWORK', placeholder: 'VD: 192.168.1.1 hoặc .torrent' }
    };

    const current = tabConfig[activeTab] || tabConfig['overview'];

    // Lấy dữ liệu (Nếu ở tab overview sẽ lấy toàn bộ, nếu ở tab khác sẽ lấy theo category)
  const fetchData = useCallback(async () => {
    try {
        const categoryQuery = current.category ? `?category=${current.category}` : '';
        const [policyRes, agentRes] = await Promise.all([
            axios.get(`/policies${categoryQuery}`),
            axios.get('/agents')
        ]);
        setPolicies(policyRes.data || []);
        setAgents(agentRes.data || []);
    } catch (err) { console.error("Lỗi tải dữ liệu:", err); }
}, [current.category]); // Đưa current.category vào list theo dõi

// Đưa fetchData vào mảng dependency của useEffect
useEffect(() => { fetchData(); }, [fetchData]);

    const handleAddPolicy = async (e) => {
        e.preventDefault();
        try {
            await axios.post('/policies', {
                title: newTitle,
                value: newValue,
                category: current.category,
                policy_type: policyType,
                target_type: targetType,
                target_hwids: targetType === 'SPECIFIC' ? selectedHWIDs : []
            });
            setNewTitle(''); setNewValue(''); setSelectedHWIDs([]);
            fetchData();
            alert("Đã áp dụng chính sách mới thành công!");
        } catch (err) { alert("Lỗi khi tạo chính sách!"); }
    };

    const handleDelete = async (id) => {
        if (!window.confirm("Xóa quy tắc này?")) return;
        try {
            await axios.delete(`/policies/${id}`);
            fetchData();
        } catch (err) { alert("Lỗi khi xóa!"); }
    };

    // Hàm chuyển Tab
    const handleTabChange = (tabKey) => {
        navigate(`/admin/policy-center?tab=${tabKey}`);
    };

    // --- GIAO DIỆN TỔNG QUÁT (OVERVIEW) ---
    const renderOverview = () => {
        const softwareCount = policies.filter(p => p.category === 'SOFTWARE').length;
        const usbCount = policies.filter(p => p.category === 'USB').length;
        const networkCount = policies.filter(p => p.category === 'NETWORK').length;

        return (
            <div className="space-y-8 animate-fade-in">
                {/* Các thẻ thống kê */}
                <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                    <div className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl flex items-center gap-5">
                        <div className="p-4 bg-emerald-500/10 text-emerald-400 rounded-2xl"><ShieldCheck size={32}/></div>
                        <div><p className="text-slate-400 text-sm font-bold">Tổng quy tắc</p><h3 className="text-3xl font-black text-white">{policies.length}</h3></div>
                    </div>
                    <div onClick={() => handleTabChange('software')} className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl flex items-center gap-5 cursor-pointer hover:border-emerald-500/50 transition group">
                        <div className="p-4 bg-blue-500/10 text-blue-400 rounded-2xl group-hover:scale-110 transition"><List size={32}/></div>
                        <div><p className="text-slate-400 text-sm font-bold">Phần mềm</p><h3 className="text-3xl font-black text-white">{softwareCount}</h3></div>
                    </div>
                    <div onClick={() => handleTabChange('usb')} className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl flex items-center gap-5 cursor-pointer hover:border-emerald-500/50 transition group">
                        <div className="p-4 bg-purple-500/10 text-purple-400 rounded-2xl group-hover:scale-110 transition"><Usb size={32}/></div>
                        <div><p className="text-slate-400 text-sm font-bold">Thiết bị USB</p><h3 className="text-3xl font-black text-white">{usbCount}</h3></div>
                    </div>
                    <div onClick={() => handleTabChange('network')} className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl flex items-center gap-5 cursor-pointer hover:border-emerald-500/50 transition group">
                        <div className="p-4 bg-orange-500/10 text-orange-400 rounded-2xl group-hover:scale-110 transition"><Activity size={32}/></div>
                        <div><p className="text-slate-400 text-sm font-bold">Mạng</p><h3 className="text-3xl font-black text-white">{networkCount}</h3></div>
                    </div>
                </div>

                {/* Danh sách quy tắc mới nhất */}
                <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-xl overflow-hidden">
                    <div className="p-5 border-b border-slate-800 bg-slate-800/30 flex items-center gap-3">
                        <Clock className="text-emerald-400" size={20}/>
                        <h3 className="font-bold text-white">Quy tắc vừa được áp dụng gần đây</h3>
                    </div>
                    <div className="p-6">
                        {policies.length === 0 ? (
                            <div className="text-center py-10 text-slate-500 italic">Chưa có chính sách nào trên hệ thống.</div>
                        ) : (
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                                {/* Lấy 6 policy mới nhất */}
                                {[...policies].reverse().slice(0, 6).map(p => (
                                    <div key={p.ID} className="bg-slate-900 border border-slate-700 p-4 rounded-2xl flex items-start gap-3">
                                        <div className="p-2 bg-slate-800 rounded-lg text-slate-400">
                                            {p.category === 'SOFTWARE' ? <List size={16}/> : p.category === 'USB' ? <Usb size={16}/> : <Activity size={16}/>}
                                        </div>
                                        <div>
                                            <p className="text-sm font-bold text-white truncate max-w-[200px]">{p.title}</p>
                                            <div className="flex gap-2 mt-1">
                                                <span className={`text-[9px] px-1.5 py-0.5 rounded font-bold uppercase ${p.policy_type === 'BLACKLIST' ? 'bg-red-500/20 text-red-400' : 'bg-emerald-500/20 text-emerald-400'}`}>{p.policy_type}</span>
                                                <span className="text-[9px] px-1.5 py-0.5 rounded font-bold uppercase bg-slate-800 text-slate-400">{p.target_type}</span>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                </div>
            </div>
        );
    };

    // --- GIAO DIỆN QUẢN LÝ TỪNG LOẠI (SOFTWARE, USB, NETWORK) ---
    const renderCategoryManager = () => (
        <div className="grid grid-cols-1 xl:grid-cols-3 gap-8 animate-fade-in">
            {/* CỘT TRÁI: FORM */}
            <div className="xl:col-span-1 space-y-6">
                <div className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl">
                    <h3 className="text-lg font-bold mb-6 flex items-center gap-2 text-emerald-400">
                        <Plus size={20}/> Tạo chính sách {current.label}
                    </h3>
                    <form onSubmit={handleAddPolicy} className="space-y-4">
                        <input type="text" placeholder="Tên quy tắc (VD: Chặn UltraView)" value={newTitle} onChange={e => setNewTitle(e.target.value)} required className="w-full p-3 bg-slate-900 border border-slate-700 rounded-xl outline-none focus:border-emerald-500 text-white" />
                        <input type="text" placeholder={current.placeholder} value={newValue} onChange={e => setNewValue(e.target.value)} required className="w-full p-3 bg-slate-900 border border-slate-700 rounded-xl outline-none focus:border-emerald-500 text-white" />

                        <div className="grid grid-cols-2 gap-3">
                            <button type="button" onClick={() => setPolicyType('BLACKLIST')} className={`py-2 rounded-xl border font-bold text-xs transition ${policyType === 'BLACKLIST' ? 'bg-red-500/20 border-red-500 text-red-400' : 'bg-slate-900 border-slate-700 text-slate-500'}`}>BLACKLIST (CẤM)</button>
                            <button type="button" onClick={() => setPolicyType('WHITELIST')} className={`py-2 rounded-xl border font-bold text-xs transition ${policyType === 'WHITELIST' ? 'bg-emerald-500/20 border-emerald-500 text-emerald-400' : 'bg-slate-900 border-slate-700 text-slate-500'}`}>WHITELIST (CHO PHÉP)</button>
                        </div>

                        <div className="pt-4 border-t border-slate-800">
                            <label className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Phạm vi áp dụng</label>
                            <div className="flex gap-4 mt-2">
                                <label className="flex items-center gap-2 cursor-pointer group">
                                    <input type="radio" checked={targetType === 'GLOBAL'} onChange={() => setTargetType('GLOBAL')} className="hidden" />
                                    <div className={`p-2 rounded-lg transition ${targetType === 'GLOBAL' ? 'bg-emerald-500 text-white' : 'bg-slate-800 text-slate-500'}`}><Globe size={16}/></div>
                                    <span className={`text-sm ${targetType === 'GLOBAL' ? 'text-white font-bold' : 'text-slate-500'}`}>Toàn hệ thống</span>
                                </label>
                                <label className="flex items-center gap-2 cursor-pointer group">
                                    <input type="radio" checked={targetType === 'SPECIFIC'} onChange={() => setTargetType('SPECIFIC')} className="hidden" />
                                    <div className={`p-2 rounded-lg transition ${targetType === 'SPECIFIC' ? 'bg-blue-500 text-white' : 'bg-slate-800 text-slate-500'}`}><Laptop size={16}/></div>
                                    <span className={`text-sm ${targetType === 'SPECIFIC' ? 'text-white font-bold' : 'text-slate-500'}`}>Chọn máy trạm</span>
                                </label>
                            </div>
                        </div>

                        {targetType === 'SPECIFIC' && (
                            <div className="bg-slate-900/50 p-4 rounded-2xl border border-slate-800 max-h-40 overflow-y-auto space-y-2">
                                {agents.map(agent => (
                                    <label key={agent.hwid} className="flex items-center gap-3 cursor-pointer p-2 hover:bg-slate-800 rounded-lg transition">
                                        <input type="checkbox" checked={selectedHWIDs.includes(agent.hwid)} onChange={() => selectedHWIDs.includes(agent.hwid) ? setSelectedHWIDs(selectedHWIDs.filter(id => id !== agent.hwid)) : setSelectedHWIDs([...selectedHWIDs, agent.hwid])} className="accent-blue-500" />
                                        <span className="text-xs text-slate-300 font-bold">{agent.hostname}</span>
                                    </label>
                                ))}
                            </div>
                        )}
                        <button type="submit" className="w-full bg-emerald-500 hover:bg-emerald-600 text-white font-bold py-3 rounded-xl transition shadow-lg shadow-emerald-500/20">Áp dụng luật</button>
                    </form>
                </div>
            </div>

            {/* CỘT PHẢI: DANH SÁCH */}
            <div className="xl:col-span-2 space-y-4">
                <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-xl overflow-hidden min-h-[400px]">
                    <div className="p-5 border-b border-slate-800 bg-slate-800/30 flex items-center justify-between">
                        <h3 className="font-bold text-white flex items-center gap-2"><CheckCircle2 className="text-emerald-400" size={18}/> Danh sách đang thực thi</h3>
                        <span className="text-xs bg-slate-700 px-3 py-1 rounded-full text-slate-300 font-bold">{policies.length} quy tắc</span>
                    </div>
                    <div className="p-6 grid grid-cols-1 md:grid-cols-2 gap-4">
                        {policies.length === 0 ? (
                            <div className="col-span-2 text-center py-20 text-slate-500 italic">Chưa có chính sách {current.label.toLowerCase()} nào.</div>
                        ) : (
                            policies.map(p => (
                                <div key={p.ID} className={`flex items-center justify-between bg-slate-900 border p-4 rounded-2xl group transition ${p.policy_type === 'BLACKLIST' ? 'border-red-500/20 hover:border-red-500/50' : 'border-emerald-500/20 hover:border-emerald-500/50'}`}>
                                    <div className="flex items-center gap-4 overflow-hidden">
                                        <div className={`p-2 rounded-lg shrink-0 ${p.policy_type === 'BLACKLIST' ? 'text-red-400 bg-red-400/10' : 'text-emerald-400 bg-emerald-400/10'}`}>
                                            {current.icon}
                                        </div>
                                        <div className="truncate">
                                            <div className="flex items-center gap-2">
                                                <p className="font-bold text-white text-sm truncate">{p.title}</p>
                                                <span className={`shrink-0 text-[8px] px-1.5 py-0.5 rounded font-black ${p.target_type === 'GLOBAL' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-blue-500/20 text-blue-400'}`}>{p.target_type}</span>
                                            </div>
                                            <p className="text-[10px] text-slate-500 font-mono mt-0.5 truncate">{p.value}</p>
                                        </div>
                                    </div>
                                    <button onClick={() => handleDelete(p.ID)} className="p-2 text-slate-600 hover:text-red-400 hover:bg-red-500/10 rounded-xl transition shrink-0"><Trash2 size={16}/></button>
                                </div>
                            ))
                        )}
                    </div>
                </div>
            </div>
        </div>
    );

    return (
        <div className="text-slate-200">
            <header className="mb-8">
                <h1 className="text-3xl font-bold text-white flex items-center gap-3">
                    <ShieldCheck className="text-emerald-400" size={32}/> Trung tâm Chính sách
                </h1>
                <p className="text-slate-400 text-sm mt-2">Thiết lập các quy tắc bảo mật kỹ thuật đa lớp áp dụng xuống toàn bộ máy trạm</p>
                
                {/* MENU ĐIỀU HƯỚNG TAB BÊN TRONG TRANG */}
                <div className="mt-6 flex flex-wrap gap-2 border-b border-slate-800 pb-4">
                    {Object.keys(tabConfig).map(key => (
                        <button 
                            key={key}
                            onClick={() => handleTabChange(key)}
                            className={`flex items-center gap-2 px-5 py-2.5 rounded-xl text-sm font-bold transition ${activeTab === key ? 'bg-emerald-500 text-white shadow-lg shadow-emerald-500/20' : 'bg-slate-800/50 text-slate-400 hover:text-white hover:bg-slate-700'}`}
                        >
                            {tabConfig[key].icon} {tabConfig[key].label}
                        </button>
                    ))}
                </div>
            </header>

            {/* Render nội dung dựa trên Tab đang chọn */}
            {activeTab === 'overview' ? renderOverview() : renderCategoryManager()}
        </div>
    );
};

export default PolicyCenter;