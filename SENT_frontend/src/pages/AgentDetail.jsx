// src/pages/AgentDetail.jsx
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Monitor, Cpu, Package, ArrowLeft } from 'lucide-react';
import axios from '../api/axios';

const AgentDetail = () => {
    const { hwid } = useParams();
    const navigate = useNavigate();
    const [agent, setAgent] = useState(null);

    useEffect(() => {
        const fetchDetail = async () => {
            try {
                const res = await axios.get(`/agents/${hwid}`); // Gọi API Backend
                setAgent(res.data);
            } catch (err) {
                console.error("Lỗi lấy chi tiết máy:", err);
            }
        };
        fetchDetail();
    }, [hwid]);

    if (!agent) return <div className="p-10 text-slate-400 font-bold uppercase tracking-widest">Đang truy vấn dữ liệu...</div>;

    return (
        <div className="text-slate-200 p-4">
            <button onClick={() => navigate('/agents')} className="flex items-center gap-2 text-slate-500 hover:text-emerald-400 mb-8 transition-all">
                <ArrowLeft size={18}/> Quay lại
            </button>

            <div className="bg-[#1e293b] p-8 rounded-3xl border border-slate-800 shadow-2xl">
                <div className="flex justify-between items-start mb-10">
                    <div className="flex gap-6">
                        <div className="p-5 bg-slate-900 rounded-2xl text-emerald-400"><Monitor size={40}/></div>
                        <div>
                          <h1 className="text-4xl font-black text-white">{agent.hostname || 'Chưa định danh'}</h1>
                          <p className="text-sm text-slate-500 font-mono mt-1">HWID: {agent.hwid}</p>
                        </div>
                    </div>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-2 gap-10">
                    {/* Thông tin phần cứng */}
                    <div className="space-y-6">
                        <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 tracking-widest">
                            <Cpu size={14}/> Cấu hình thiết bị
                        </h3>
                        <div className="bg-slate-900/50 p-6 rounded-2xl border border-slate-800 space-y-4">
                            <DetailItem label="CPU" value={agent.Inventory?.cpu_model} />
                            <DetailItem label="RAM" value={`${agent.Inventory?.ram_total_gb} GB`} />
                            <DetailItem label="Hệ điều hành" value={agent.Inventory?.os_info} />
                        </div>
                    </div>

                    {/* Danh sách phần mềm */}
                    <div className="space-y-6">
                        <h3 className="text-xs font-black text-slate-500 uppercase flex items-center gap-2 tracking-widest">
                            <Package size={14}/> Phần mềm (Compliance)
                        </h3>
                        <div className="bg-slate-900/50 p-4 rounded-2xl border border-slate-800 max-h-60 overflow-y-auto">
                            {agent.Software?.map((sw, i) => (
                                <div key={i} className="py-2 px-3 border-b border-slate-800 last:border-0 text-xs text-slate-400">
                                    {sw.SoftwareName}
                                </div>
                            ))}
                        </div>
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