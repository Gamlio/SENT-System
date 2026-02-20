import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Monitor, Cpu, Package, ArrowLeft, ShieldAlert, ShieldCheck, ShieldBan } from 'lucide-react';
import axios from '../../api/axios';

const AgentDetail = () => {
    const { hwid } = useParams();
    const navigate = useNavigate();
    const [agent, setAgent] = useState(null);
    const [logs, setLogs] = useState([]);   
    const [whitelist, setWhitelist] = useState([]); // Thêm state lưu Whitelist

    useEffect(() => {
        const fetchDetail = async () => {
            try {
                const res = await axios.get(`/agents/${hwid}`);
                setAgent(res.data);
                
                const logRes = await axios.get(`/agents/${hwid}/logs`);
                setLogs(logRes.data);

                // Gọi thêm API lấy Whitelist của riêng máy này
                const whiteRes = await axios.get(`/agents/${hwid}/whitelist`);
                setWhitelist(whiteRes.data || []);
            } catch (err) {
                console.error("Lỗi lấy chi tiết máy:", err);
            }
        };
        fetchDetail();
    }, [hwid]);

    if (!agent) return <div className="p-10 text-slate-400 font-bold uppercase tracking-widest">Đang truy vấn dữ liệu...</div>;

    const inv = agent.inventory || {}; 
    const soft = agent.software || [];
    const isZeroTrustMode = whitelist.length > 0; // Nếu có whitelist -> Bật mode Zero Trust

    // Hàm kiểm tra phần mềm có nằm trong whitelist không
    const isWhitelisted = (swName) => {
        if (!isZeroTrustMode) return true; // Nếu ko bật Zero Trust thì tạm coi là an toàn (chờ Blacklist xử)
        return whitelist.some(w => swName.toLowerCase().includes(w.name.toLowerCase()));
    };

    return (
        <div className="text-slate-200 p-4">
            <div className="flex justify-between items-center mb-8">
                <button onClick={() => navigate('/agents')} className="flex items-center gap-2 text-slate-500 hover:text-emerald-400 transition-all">
                    <ArrowLeft size={18}/> Quay lại
                </button>
                
                {/* Nút chuyển nhanh sang trang Cấu hình Whitelist */}
                <button 
                    onClick={() => navigate(`/agents/${hwid}/software`)}
                    className="flex items-center gap-2 bg-slate-800 hover:bg-slate-700 text-emerald-400 px-4 py-2 rounded-xl transition font-bold text-sm"
                >
                    <ShieldCheck size={16}/> Cấu hình Zero Trust
                </button>
            </div>

            <div className="bg-[#1e293b] p-8 rounded-3xl border border-slate-800 shadow-2xl">
                {/* HEADER CHI TIẾT */}
                <div className="flex justify-between items-start mb-10">
                    <div className="flex gap-6">
                        <div className="p-5 bg-slate-900 rounded-2xl text-emerald-400"><Monitor size={40}/></div>
                        <div>
                            <h1 className="text-4xl font-black text-white">{agent.hostname || 'Chưa định danh'}</h1>
                            <p className="text-sm text-slate-500 font-mono mt-1">HWID: {agent.hwid}</p>
                            <p className="text-sm text-slate-500 font-mono mt-1">IP: {agent.ip_address}</p>
                        </div>
                    </div>
                    {/* HUY HIỆU TRẠNG THÁI BẢO MẬT */}
                    <div className={`px-4 py-2 rounded-xl border flex items-center gap-2 font-bold text-sm shadow-lg ${isZeroTrustMode ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400' : 'bg-slate-800 border-slate-700 text-slate-400'}`}>
                        {isZeroTrustMode ? <><ShieldCheck size={18}/> Đang bật Zero Trust</> : <><ShieldBan size={18}/> Dùng Blacklist chung</>}
                    </div>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-2 gap-10">
                    {/* CẤU HÌNH PHẦN CỨNG */}
                    <div className="space-y-6">
                        <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 tracking-widest">
                            <Cpu size={14}/> Cấu hình thiết bị
                        </h3>
                        <div className="bg-slate-900/50 p-6 rounded-2xl border border-slate-800 space-y-4">
                            <DetailItem label="Địa chỉ IP mạng" value={agent.ip_address} />
                            <DetailItem label="CPU" value={inv.cpu_model} />
                            <DetailItem label="RAM" value={`${inv.ram_total_gb || 0} GB`} />
                            <DetailItem label="Hệ điều hành" value={inv.os_info} />
                        </div>
                    </div>

                    {/* DANH SÁCH PHẦN MỀM CÀI ĐẶT */}
                    <div className="space-y-6">
                        <div className="flex justify-between items-center">
                            <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 tracking-widest">
                                <Package size={14}/> Phần mềm hiện tại ({soft.length})
                            </h3>
                        </div>
                        
                        <div className="bg-slate-900/50 p-4 rounded-2xl border border-slate-800 max-h-60 overflow-y-auto">
                            {soft.length === 0 ? (
                                <p className="text-xs text-slate-500 italic">Chưa có dữ liệu phần mềm</p>
                            ) : (
                                soft.map((sw, i) => {
                                    const safe = isWhitelisted(sw.software_name);
                                    return (
                                        <div key={i} className={`py-2.5 px-3 border-b border-slate-800/50 last:border-0 text-xs font-medium flex justify-between items-center ${!safe && isZeroTrustMode ? 'bg-red-500/5' : ''}`}>
                                            <span className={safe ? 'text-slate-300' : 'text-red-400 font-bold'}>
                                                {sw.software_name}
                                            </span>
                                            {isZeroTrustMode && (
                                                safe 
                                                ? <span className="text-[9px] bg-emerald-500/20 text-emerald-400 px-2 py-0.5 rounded uppercase font-bold">Hợp lệ</span>
                                                : <span className="text-[9px] bg-red-500/20 text-red-400 px-2 py-0.5 rounded uppercase font-bold">Vi phạm</span>
                                            )}
                                        </div>
                                    )
                                })
                            )}
                        </div>
                    </div>
                </div>

                {/* NHẬT KÝ CẢNH BÁO */}
                <div className="mt-10">
                    <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
                        <ShieldAlert className="text-red-400" size={20}/> 
                        Nhật ký Cảnh báo & Vi phạm
                    </h3>
                    
                    <div className="bg-[#1e293b] rounded-3xl border border-slate-800 overflow-hidden shadow-2xl">
                        <table className="w-full text-left">
                            <thead className="bg-slate-900/50 text-xs uppercase text-slate-500 font-bold">
                                <tr>
                                    <th className="p-4">Thời gian</th>
                                    <th className="p-4">Loại cảnh báo</th>
                                    <th className="p-4">Mức độ</th>
                                    <th className="p-4">Nội dung chi tiết</th>
                                    <th className="p-4">Trạng thái</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-slate-800">
                                {logs.length === 0 ? (
                                    <tr>
                                        <td colSpan="5" className="p-8 text-center text-slate-500 italic">
                                            Máy trạm này chưa có ghi nhận vi phạm nào.
                                        </td>
                                    </tr>
                                ) : (
                                    logs.map((log) => (
                                        <tr key={log.ID} className="hover:bg-slate-800/50 transition">
                                            <td className="p-4 text-sm text-slate-400 font-mono">
                                                {new Date(log.CreatedAt).toLocaleString('vi-VN')}
                                            </td>
                                            <td className="p-4">
                                                <span className="text-xs font-bold text-white bg-slate-700 px-2 py-1 rounded border border-slate-600">
                                                    {log.alert_type || "SECURITY"}
                                                </span>
                                            </td>
                                            <td className="p-4">
                                                <span className={`text-xs font-bold px-2 py-1 rounded ${
                                                    log.severity === 'Critical' ? 'bg-red-500/20 text-red-400 border border-red-500/30' :
                                                    log.severity === 'High' ? 'bg-orange-500/10 text-orange-400' : 
                                                    'bg-blue-500/10 text-blue-400'
                                                }`}>
                                                    {log.severity || "Low"}
                                                </span>
                                            </td>
                                            <td className="p-4 text-sm text-slate-300">
                                                {log.description}
                                            </td>
                                            <td className="p-4">
                                                <span className={`text-[10px] font-bold uppercase ${log.is_resolved ? 'text-emerald-400' : 'text-red-400'}`}>
                                                    {log.is_resolved ? 'Đã xử lý' : 'Chưa xử lý'}
                                                </span>
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div> 
    );
};

const DetailItem = ({ label, value }) => (
    <div className="flex justify-between items-center text-sm">
        <span className="text-slate-500">{label}</span>
        <span className="font-bold text-white">{value || 'N/A'}</span>
    </div>
);

export default AgentDetail;