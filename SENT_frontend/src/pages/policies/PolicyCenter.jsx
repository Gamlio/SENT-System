import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { LayoutDashboard, List, Usb, Wifi } from 'lucide-react';
import { usePolicies } from './hooks/usePolicies';
import { useSocketSubscription } from '../../context/useSocketSubscription';

// Import Components
import PolicyOverview from './components/PolicyOverview';
import PolicyForm from './components/PolicyForm';
import PolicyList from './components/PolicyList';

const PolicyCenter = () => {
    const location = useLocation();
    const navigate = useNavigate();
    const queryParams = new URLSearchParams(location.search);
    const activeTab = queryParams.get('tab') || 'overview';

    const { 
        policies, loading, 
        fetchPolicies, addPolicy, deletePolicy, deleteBulkPolicies 
    } = usePolicies();

    // Lắng nghe sự kiện từ WebSocket và tải lại dữ liệu
    useSocketSubscription('POLICY_UPDATED', () => {
        console.log('[WS] Bộ luật chính sách đã thay đổi, đang tải lại...');
        fetchPolicies();
    });

    const [agents] = useState([]); 

    // Cấu hình Tabs
    const tabConfig = {
        overview: { label: 'Tổng quát', icon: <LayoutDashboard size={18}/>, category: '' },
        software: { label: 'Phần mềm', icon: <List size={18}/>, category: 'SOFTWARE', placeholder: 'VD: game.exe', hint: 'Tên tiến trình' },
        usb: { label: 'Thiết bị USB', icon: <Usb size={18}/>, category: 'USB', placeholder: 'VD: VID_1234', hint: 'Mã hoặc tên USB' },
        network: { label: 'Mạng & Port', icon: <Wifi size={18}/>, category: 'NETWORK', placeholder: 'VD: 80, 443', hint: 'Cổng mạng' },
    };

    const currentConfig = tabConfig[activeTab] || tabConfig.overview;

    // 1. [QUAN TRỌNG NHẤT] Luôn tải TẤT CẢ (Không truyền tham số category)
    // Để lấy cả những dòng "OTHER" về rồi Frontend tự lọc
    useEffect(() => {
        fetchPolicies(); 
    }, [fetchPolicies]);

    // 2. Logic Lọc Thông Minh (Fix lỗi không hiện USB)
    const filteredPolicies = useMemo(() => {
        if (activeTab === 'overview') return policies;

        return policies.filter(p => {
            const cat = (p.category || '').toUpperCase().trim();
            const val = (p.value || '').toUpperCase().trim();

            // --- TAB PHẦN MỀM ---
            if (activeTab === 'software') {
                return cat === 'SOFTWARE' || val.endsWith('.EXE') || val.endsWith('.MSI');
            }

            // --- TAB MẠNG ---
            if (activeTab === 'network') {
                // Là Network HOẶC là số (Port) HOẶC là IP
                return cat === 'NETWORK' || cat === 'PORT' || (!isNaN(val) && val.length < 6) || (val.includes('.') && !val.includes('EXE'));
            }

            // --- TAB USB (Quan trọng) ---
            if (activeTab === 'usb') {
                // Cách 1: Đúng là USB
                if (cat.includes('USB') || cat === 'DEVICE' || cat === 'STORAGE') return true;
                
                // Cách 2: Bị lưu là "OTHER" nhưng tên giống USB (Fix cho file Word của bạn)
                if (cat === 'OTHER' || cat === '') {
                    // Nếu nó KHÔNG phải phần mềm và KHÔNG phải Port -> Coi như là USB
                    const isSoftware = val.endsWith('.EXE') || val.endsWith('.MSI');
                    const isNetwork = !isNaN(val) && val.length < 6;
                    
                    if (!isSoftware && !isNetwork) return true; 
                }
                return false;
            }

            return false;
        });
    }, [policies, activeTab]);

    const handleTabChange = (key) => navigate(`?tab=${key}`);

    const handleAddPolicy = async (data) => {
        // Khi thêm mới thì gửi kèm category gợi ý của tab hiện tại
        const res = await addPolicy({ ...data, category: currentConfig.category });
        if (res.success) fetchPolicies(); // Load lại
        else alert(res.error);
    };

    const handleDelete = async (id) => {
        if(!window.confirm("Bạn có chắc muốn xóa?")) return;
        const res = await deletePolicy(id);
        if (res.success) fetchPolicies();
    };

    const handleBulkDelete = async (ids) => {
        const res = await deleteBulkPolicies(ids);
        if (res.success) fetchPolicies();
        else alert(res.error);
    };

    return (
        <div className="p-6 text-slate-200 h-[calc(100vh-60px)] flex flex-col bg-[#050B14] font-sans overflow-hidden">
            <div className="mb-4 flex justify-between items-end shrink-0">
                <div>
                    <h1 className="text-xl font-black text-white flex items-center gap-2 uppercase tracking-tight">
                        <LayoutDashboard className="text-indigo-500" size={24}/> CHÍNH SÁCH BẢN QUYỀN
                    </h1>
                    <p className="text-[11px] text-slate-500 mt-1 uppercase tracking-widest font-bold">Quản lý luật, chính sách và kiểm soát truy cập thiết bị.</p>
                </div>
            </div>

            <div className="mb-4">
                <div className="bg-[#1e293b] p-1.5 rounded-2xl border border-slate-800 flex shadow-lg">
                    {Object.keys(tabConfig).map(key => (
                        <button
                            key={key}
                            onClick={() => handleTabChange(key)}
                            className={`flex items-center gap-2 px-5 py-2.5 rounded-xl text-xs font-bold transition-all ${activeTab === key ? 'bg-emerald-500 text-white shadow-lg shadow-emerald-500/20' : 'bg-transparent text-slate-400 hover:text-white hover:bg-slate-700/50'}`}
                        >
                            {tabConfig[key].icon} {tabConfig[key].label}
                        </button>
                    ))}
                </div>
            </div>

            <div className="flex-1 bg-[#0A101D] rounded-lg border border-slate-800 shadow-2xl overflow-hidden">
                {activeTab === 'overview' ? (
                    <PolicyOverview policies={policies} handleTabChange={handleTabChange} />
                ) : (
                    <div className="grid grid-cols-1 xl:grid-cols-12 gap-6 h-full min-h-0 p-4">
                        <div className="xl:col-span-4 h-full overflow-y-auto pr-1">
                            <PolicyForm 
                                currentConfig={currentConfig} 
                                agents={agents} 
                                onAddPolicy={handleAddPolicy} 
                                isLoading={loading} 
                            />
                        </div>
                        <div className="xl:col-span-8 h-full min-h-0 flex flex-col">
                            <PolicyList 
                                policies={filteredPolicies} // <--- Danh sách đã được lọc kỹ
                                onDelete={handleDelete} 
                                onBulkDelete={handleBulkDelete} 
                                agents={agents} 
                            />
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default PolicyCenter;