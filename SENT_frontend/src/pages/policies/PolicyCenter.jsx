import React, { useState, useEffect, useCallback } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { ShieldCheck, LayoutDashboard, List, Usb, Activity } from 'lucide-react';
import axios from '../../api/axios';

// Import các component con đã tách
import PolicyOverview from './components/PolicyOverview';
import PolicyForm from './components/PolicyForm';
import PolicyList from './components/PolicyList';

const PolicyCenter = () => {
    const location = useLocation();
    const navigate = useNavigate();
    const queryParams = new URLSearchParams(location.search);
    const activeTab = queryParams.get('tab') || 'overview';

    const [policies, setPolicies] = useState([]);
    const [agents, setAgents] = useState([]);
    const [isLoading, setIsLoading] = useState(false);

    // Cấu hình Tabs
    const tabConfig = {
        overview: { label: 'Tổng quát', icon: <LayoutDashboard size={18}/>, category: '' },
        software: { label: 'Phần mềm', icon: <List size={18}/>, category: 'SOFTWARE', placeholder: 'VD: utorrent.exe, game.exe...', hint: 'Nhập tên tiến trình hoặc phần mềm cần chặn/cho phép.' },
        usb: { label: 'Thiết bị USB', icon: <Usb size={18}/>, category: 'USB', placeholder: 'VD: VID_0951&PID_1666', hint: 'Nhập Device Instance Path hoặc Serial Number.' },
        network: { label: 'Quy tắc Mạng', icon: <Activity size={18}/>, category: 'NETWORK', placeholder: 'VD: 3389, 192.168.1.100', hint: 'Nhập Port hoặc IP cần kiểm soát.' }
    };

    const current = tabConfig[activeTab] || tabConfig['overview'];

    // Load dữ liệu
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
    }, [current.category]);

    useEffect(() => { fetchData(); }, [fetchData]);

    // Xử lý Thêm mới (Logic gọi API nằm ở đây để component con gọi lên)
   const handleAddPolicy = async (formData) => {
    // Validate
    if (formData.target_type === 'SPECIFIC' && formData.target_hwids.length === 0) {
        alert("Vui lòng chọn ít nhất một máy trạm áp dụng!");
        return;
    }

    setIsLoading(true);
    try {
        // Lấy danh sách giá trị (Nếu từ form đơn thì tạo mảng 1 phần tử)
        const valuesList = formData.values || [formData.value];
        
        // Gọi API Bulk (Chỉ 1 request duy nhất)
        await axios.post('/policies/bulk', {
            title: formData.title,
            category: current.category,
            policy_type: formData.policy_type,
            target_type: formData.target_type,
            target_hwids: formData.target_hwids,
            values: valuesList // Gửi mảng lên server
        });

        fetchData();
        alert(`Đã áp dụng ${valuesList.length} quy tắc thành công!`);
    } catch (err) { 
        console.error(err);
        alert("Lỗi: " + (err.response?.data?.error || "Không thể lưu chính sách")); 
    } finally {
        setIsLoading(false);
    }
};
    // Xử lý Xóa
    const handleDelete = async (id) => {
        if (!window.confirm("Bạn chắc chắn muốn xóa quy tắc này?")) return;
        try {
            await axios.delete(`/policies/${id}`);
            fetchData();
        } catch (err) { alert("Lỗi khi xóa!"); }
    };

    const handleTabChange = (tabKey) => navigate(`/admin/policy-center?tab=${tabKey}`);

    return (
        <div className="text-slate-200 pb-20">
            {/* Header */}
            <header className="mb-8">
                <h1 className="text-3xl font-bold text-white flex items-center gap-3">
                    <ShieldCheck className="text-emerald-400" size={32}/> Trung tâm Chính sách
                </h1>
                <p className="text-slate-400 text-sm mt-2">Quản lý và phân cấp quy tắc bảo mật cho toàn bộ hệ thống</p>
                
                <div className="mt-6 flex flex-wrap gap-2 border-b border-slate-800 pb-4">
                    {Object.keys(tabConfig).map(key => (
                        <button key={key} onClick={() => handleTabChange(key)} className={`flex items-center gap-2 px-5 py-2.5 rounded-xl text-sm font-bold transition ${activeTab === key ? 'bg-emerald-500 text-white shadow-lg shadow-emerald-500/20' : 'bg-slate-800/50 text-slate-400 hover:text-white hover:bg-slate-700'}`}>
                            {tabConfig[key].icon} {tabConfig[key].label}
                        </button>
                    ))}
                </div>
            </header>

            {/* Nội dung chính */}
            {activeTab === 'overview' ? (
                <PolicyOverview policies={policies} handleTabChange={handleTabChange} />
            ) : (
                <div className="grid grid-cols-1 xl:grid-cols-12 gap-8 animate-fade-in items-start">
                    {/* Cột Trái: Form */}
                    <div className="xl:col-span-4">
                        <PolicyForm 
                            currentConfig={current} 
                            agents={agents} 
                            onAddPolicy={handleAddPolicy} 
                            isLoading={isLoading} 
                        />
                    </div>
                    {/* Cột Phải: Danh sách */}
                    <div className="xl:col-span-8">
                        <PolicyList 
                            policies={policies} 
                            onDelete={handleDelete} 
                            agents={agents} 
                            currentConfig={current}
                        />
                    </div>
                </div>
            )}
        </div>
    );
};

export default PolicyCenter;