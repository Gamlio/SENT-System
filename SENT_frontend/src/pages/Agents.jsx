import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Search, Monitor, Cpu, ArrowUpRight, Globe } from 'lucide-react'; 
import axios from '../api/axios'; // Đảm bảo đã cấu hình axios hướng về port 8000

const Agents = () => {
    const navigate = useNavigate();
    const [agents, setAgents] = useState([]);

    useEffect(() => {
        const fetchAgents = async () => {
            try {
                // SỬA 1: Thêm dấu gạch chéo '/' vào đầu để URL chuẩn xác
                const res = await axios.get('/agents'); 
                setAgents(res.data);
            } catch (err) {
                console.error("Lỗi lấy danh sách máy trạm:", err);
            }
        };
        fetchAgents();
    }, []);

    return (
        <div className="text-slate-200">
            {/* ... Header giữ nguyên ... */}

            <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
                {/* SỬA 2: Thêm kiểm tra độ dài mảng để báo nếu chưa có máy */}
                {agents.length === 0 ? (
                    <div className="col-span-2 text-center text-slate-500 py-10">
                        Chưa có dữ liệu. Hãy chạy Agent để kết nối.
                    </div>
                ) : (
                    agents.map(agent => (
                        <div key={agent.hwid} onClick={() => navigate(`/agents/${agent.hwid}`)} 
                            className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 hover:border-emerald-500/50 transition-all cursor-pointer group">
                            <div className="flex justify-between items-start">
                                <div className="flex gap-4">
                                    <div className="p-4 bg-slate-900 rounded-2xl text-emerald-400"><Monitor size={28}/></div>
                                    <div>
                                        {/* Bây giờ Backend đã trả về hostname (thường) nên code này sẽ chạy đúng */}
                                        <h3 className="text-xl font-bold text-white group-hover:text-emerald-400 transition">{agent.hostname || "Unknown"}</h3>
                                        <p className="text-xs text-slate-500 font-mono">HWID: {agent.hwid?.substring(0, 15)}...</p>
                                    </div>
                                </div>
                                <span className={`flex items-center gap-1.5 text-[10px] font-bold px-3 py-1 rounded-full uppercase ${agent.status === 'online' ? 'text-emerald-400 bg-emerald-500/10' : 'text-red-400 bg-red-500/10'}`}>
                                    <span className={`w-1.5 h-1.5 rounded-full ${agent.status === 'online' ? 'bg-emerald-400 animate-pulse' : 'bg-red-400'}`}></span> 
                                    {agent.status}
                                </span>
                            </div>
                            <ArrowUpRight className="absolute bottom-6 right-6 text-slate-700 group-hover:text-emerald-400 transition" size={20}/>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};
export default Agents;