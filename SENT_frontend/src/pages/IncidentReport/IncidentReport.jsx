import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import axiosInstance from '../../api/axios';

const IncidentReport = () => {
    const { id } = useParams();
    const navigate = useNavigate();
    const [incident, setIncident] = useState(null);
    const [activeTab, setActiveTab] = useState('report'); // 'report', 'details', 'playbook'
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchIncident = async () => {
            try {
                const response = await axiosInstance.get(`/api/v1/incidents/${id}`);
                setIncident(response.data.data);
            } catch (error) {
                console.error("Lỗi tải sự cố:", error);
            } finally {
                setLoading(false);
            }
        };
        fetchIncident();
    }, [id]);

    // Giả lập nội dung Playbook dựa trên Loại sự cố (Sau này có thể kéo từ Database)
    const getPlaybookSteps = (type) => {
        if (type.includes('Network')) {
            return [
                { step: 1, action: "Kiểm tra IP nguồn và đích của kết nối mạng." },
                { step: 2, action: "Sử dụng Firewall để block cổng mạng trái phép (VD: 5432, 4444)." },
                { step: 3, action: "Cách ly máy trạm khỏi mạng nội bộ nếu có dấu hiệu truyền dữ liệu lạ." }
            ];
        }
        if (type.includes('Brute Force')) {
            return [
                { step: 1, action: "Khóa tạm thời tài khoản người dùng đang bị tấn công." },
                { step: 2, action: "Chặn địa chỉ IP thực hiện brute-force trên tường lửa." },
                { step: 3, action: "Yêu cầu người dùng đổi mật khẩu mạnh và bật 2FA." }
            ];
        }
        return [
            { step: 1, action: "Xác minh lại tính chính xác của cảnh báo (False Positive?)." },
            { step: 2, action: "Cập nhật chữ ký phần mềm diệt virus." },
            { step: 3, action: "Khởi động lại tiến trình Agent trên máy trạm." }
        ];
    };

    if (loading) return <div className="p-6 text-gray-400">Đang thu thập dữ liệu báo cáo...</div>;
    if (!incident) return <div className="p-6 text-red-400">Không tìm thấy Hồ sơ Sự cố.</div>;

    const playbookSteps = getPlaybookSteps(incident.type);

    return (
        <div className="p-6 text-gray-200 min-h-screen bg-gray-900">
            {/* Header Báo Cáo */}
            <div className="flex justify-between items-start mb-6 border-b border-gray-700 pb-4">
                <div>
                    <button onClick={() => navigate(-1)} className="text-blue-400 hover:text-blue-300 text-sm mb-3 flex items-center">
                        ← Trở về danh sách
                    </button>
                    <h1 className="text-3xl font-bold text-white tracking-wide">Báo Cáo Sự Cố: #{incident.id}</h1>
                    <p className="text-gray-400 mt-2">Máy ảnh hưởng: <span className="text-blue-300 font-mono">{incident.device?.hostname}</span></p>
                </div>
                <div className="text-right">
                    <span className={`px-4 py-2 rounded font-bold uppercase tracking-wider text-sm shadow-lg ${incident.status === 'Open' ? 'bg-red-600/20 text-red-500 border border-red-500' : 'bg-green-600/20 text-green-500 border border-green-500'}`}>
                        {incident.status}
                    </span>
                    <button className="block mt-4 bg-blue-600 hover:bg-blue-500 text-white px-4 py-2 rounded text-sm font-bold transition">
                        ✓ Đánh dấu Đã Xử Lý
                    </button>
                </div>
            </div>

            {/* Điều hướng Tabs */}
            <div className="flex space-x-1 mb-6 bg-gray-800 p-1 rounded-lg w-fit">
                <button onClick={() => setActiveTab('report')} className={`px-6 py-2 rounded-md text-sm font-bold transition ${activeTab === 'report' ? 'bg-gray-700 text-white shadow' : 'text-gray-400 hover:text-gray-200'}`}>
                    📑 Tổng Quan Báo Cáo
                </button>
                <button onClick={() => setActiveTab('details')} className={`px-6 py-2 rounded-md text-sm font-bold transition ${activeTab === 'details' ? 'bg-gray-700 text-white shadow' : 'text-gray-400 hover:text-gray-200'}`}>
                    🔎 Chi Tiết Kỹ Thuật (Logs)
                </button>
                <button onClick={() => setActiveTab('playbook')} className={`px-6 py-2 rounded-md text-sm font-bold transition ${activeTab === 'playbook' ? 'bg-blue-600/20 text-blue-400 shadow border border-blue-500/30' : 'text-gray-400 hover:text-gray-200'}`}>
                    🛡️ Playbook Ứng Phó
                </button>
            </div>

            {/* Nội dung Tabs */}
            <div className="bg-gray-800 border border-gray-700 rounded-xl p-6 shadow-xl">
                
                {/* TAB 1: BÁO CÁO TỔNG QUAN */}
                {activeTab === 'report' && (
                    <div className="space-y-6">
                        <h2 className="text-xl font-bold text-white border-b border-gray-700 pb-2">Thông tin Tóm tắt</h2>
                        <div className="grid grid-cols-2 gap-6 text-sm">
                            <div>
                                <p className="text-gray-400 mb-1">Loại đe dọa:</p>
                                <p className="font-semibold text-lg text-gray-200">{incident.type}</p>
                            </div>
                            <div>
                                <p className="text-gray-400 mb-1">Mức độ nghiêm trọng:</p>
                                <p className={`font-bold text-lg ${incident.severity === 'High' ? 'text-red-500' : 'text-yellow-500'}`}>{incident.severity}</p>
                            </div>
                            <div>
                                <p className="text-gray-400 mb-1">Thời gian phát hiện:</p>
                                <p className="font-mono text-gray-300">{new Date(incident.created_at).toLocaleString('vi-VN')}</p>
                            </div>
                            <div>
                                <p className="text-gray-400 mb-1">ID Thiết bị (HWID):</p>
                                <p className="font-mono text-xs bg-gray-900 p-2 rounded text-gray-400 border border-gray-700">{incident.device?.hwid}</p>
                            </div>
                        </div>
                        <div className="mt-4">
                            <p className="text-gray-400 mb-2">Mô tả tự động:</p>
                            <div className="bg-gray-900 p-4 rounded border border-gray-700 text-gray-300 leading-relaxed">
                                {incident.description}
                            </div>
                        </div>
                    </div>
                )}

                {/* TAB 2: CHI TIẾT KỸ THUẬT */}
                {activeTab === 'details' && (
                    <div>
                        <h2 className="text-xl font-bold text-white border-b border-gray-700 pb-2 mb-4">Bằng Chứng / Alerts ({incident.alerts?.length || 0})</h2>
                        <div className="space-y-4">
                            {incident.alerts?.map(alert => (
                                <div key={alert.id} className="bg-gray-900 p-4 rounded-lg border-l-4 border-red-500">
                                    <div className="flex justify-between items-center mb-2">
                                        <span className="font-bold text-red-400">{alert.type}</span>
                                        <span className="text-xs font-mono text-gray-500">{new Date(alert.created_at).toLocaleString('vi-VN')}</span>
                                    </div>
                                    <p className="text-sm text-gray-300 font-mono bg-black/50 p-2 rounded">{alert.message}</p>
                                </div>
                            ))}
                        </div>
                    </div>
                )}

                {/* TAB 3: PLAYBOOK ỨNG PHÓ */}
                {activeTab === 'playbook' && (
                    <div>
                        <h2 className="text-xl font-bold text-blue-400 border-b border-gray-700 pb-2 mb-4 flex items-center">
                            <span className="mr-2">⚡</span> Kịch Bản Xử Lý Đề Xuất
                        </h2>
                        <p className="text-sm text-gray-400 mb-6">Thực hiện các bước sau để cô lập và khắc phục sự cố dựa trên loại đe dọa <strong className="text-gray-200">{incident.type}</strong>.</p>
                        
                        <div className="space-y-3">
                            {playbookSteps.map((item, index) => (
                                <div key={index} className="flex items-start bg-gray-900 p-4 rounded-lg border border-gray-700 hover:border-blue-500/50 transition">
                                    <div className="flex-shrink-0 w-8 h-8 rounded-full bg-blue-900 text-blue-300 flex items-center justify-center font-bold mr-4">
                                        {item.step}
                                    </div>
                                    <div className="pt-1">
                                        <p className="text-gray-200 text-sm leading-relaxed">{item.action}</p>
                                    </div>
                                    <div className="ml-auto pl-4">
                                        <input type="checkbox" className="w-5 h-5 accent-blue-600 rounded bg-gray-800 border-gray-600" />
                                    </div>
                                </div>
                            ))}
                        </div>
                        
                        <div className="mt-8 pt-6 border-t border-gray-700">
                            <label className="block text-sm text-gray-400 mb-2">Ghi chú thực thi (Resolution Notes):</label>
                            <textarea className="w-full bg-gray-900 border border-gray-700 rounded p-3 text-sm text-white focus:outline-none focus:border-blue-500" rows="3" placeholder="Nhập tóm tắt hành động đã thực hiện trước khi đóng case..."></textarea>
                        </div>
                    </div>
                )}

            </div>
        </div>
    );
};

export default IncidentReport;