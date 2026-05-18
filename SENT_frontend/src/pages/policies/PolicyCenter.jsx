import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { LayoutDashboard, List, Usb, Wifi, ShieldCheck } from 'lucide-react';
import { usePolicies } from './hooks/usePolicies';
import { useSocketSubscription } from '../../context/useSocketSubscription';

import PolicyOverview from './components/PolicyOverview';
import PolicyForm from './components/PolicyForm';
import PolicyList from './components/PolicyList';

const PolicyCenter = () => {
    const location = useLocation();
    const navigate = useNavigate();
    const queryParams = new URLSearchParams(location.search);
    const activeTab = queryParams.get('tab') || 'overview';

    const { 
        policies, groups, loading, 
        fetchPolicies, fetchGroups, addPolicy, deletePolicy, deleteBulkPolicies 
    } = usePolicies();

    // Lắng nghe sự kiện từ WebSocket và tải lại dữ liệu
    useSocketSubscription('POLICY_UPDATED', () => {
        console.log('[WS] Bộ luật chính sách đã thay đổi, đang tải lại...');
        fetchPolicies();
    });

    // Cấu hình Tabs
    const tabConfig = {
        overview: { label: 'Tổng quát', icon: <LayoutDashboard size={16}/>, category: '' },
        software: { label: 'Phần mềm', icon: <List size={16}/>, category: 'SOFTWARE', placeholder: 'VD: process.exe', hint: 'Tên tiến trình' },
        usb: { label: 'USB Device', icon: <Usb size={16}/>, category: 'USB', placeholder: 'VD: VID_045E', hint: 'Mã hoặc tên USB' },
        network: { label: 'Network', icon: <Wifi size={16}/>, category: 'NETWORK', placeholder: 'VD: 443', hint: 'Cổng mạng' },
    };

    const currentConfig = tabConfig[activeTab] || tabConfig.overview;

    // 1. [QUAN TRỌNG NHẤT] Luôn tải TẤT CẢ (Không truyền tham số category)
    // Để lấy cả những dòng "OTHER" về rồi Frontend tự lọc
    useEffect(() => {
        fetchPolicies(); 
        if (fetchGroups) fetchGroups();
    }, [fetchPolicies, fetchGroups]);

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
        <div className="p-6 h-[calc(100vh-60px)] flex flex-col text-slate-200 bg-[#050B14] font-sans overflow-hidden">
            <div className="flex justify-between items-end mb-4 shrink-0">
                <div>
                    <h1 className="text-xl font-black text-white flex items-center gap-2 uppercase tracking-tight">
                        <ShieldCheck className="text-indigo-500" size={24}/> Policy Management
                    </h1>
                    <p className="text-[11px] text-slate-500 mt-1 uppercase tracking-widest font-black opacity-70">
                        Total System Rules: {policies.length}
                    </p>
                </div>
                
                {/* Tabs styled like Asset controls */}
                <div className="flex bg-[#111827] p-1 rounded-xl border border-slate-800 shadow-lg">
                    {Object.keys(tabConfig).map(key => (
                        <button
                            key={key}
                            onClick={() => handleTabChange(key)}
                            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-[10px] font-black uppercase tracking-widest transition-all ${activeTab === key ? 'bg-indigo-600 text-white shadow-lg' : 'text-slate-500 hover:text-white hover:bg-slate-800'}`}
                        >
                            {tabConfig[key].icon} {tabConfig[key].label}
                        </button>
                    ))}
                </div>
            </div>

            <div className="flex-1 overflow-hidden">
                {activeTab === 'overview' ? (
                    <PolicyOverview policies={policies} handleTabChange={handleTabChange} />
                ) : (
                    <div className="grid grid-cols-1 xl:grid-cols-12 gap-6 h-full min-h-0 p-4">
                        <div className="xl:col-span-4 h-full overflow-y-auto pr-1">
                            <PolicyForm 
                                currentConfig={currentConfig} 
                                groups={groups || []} 
                                onAddPolicy={handleAddPolicy} 
                                isLoading={loading} 
                            />
                        </div>
                        <div className="xl:col-span-8 h-full flex flex-col min-h-0">
                            <PolicyList 
                                policies={filteredPolicies} // Giữ nguyên luồng lọc dữ liệu gốc để không mất USB
                                onDelete={handleDelete} 
                                onBulkDelete={handleBulkDelete} 
                                groups={groups || []} 
                            />
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default PolicyCenter;