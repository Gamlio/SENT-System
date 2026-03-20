import React, { useState } from 'react';
import { Key } from 'lucide-react';
import axios from '../../../api/axios'; // Đảm bảo đường dẫn đúng

const GenerateTokenButton = () => {
    const [tokenInfo, setTokenInfo] = useState(null);
    const [loading, setLoading] = useState(false);

    const handleGenerate = async () => {
        setLoading(true);
        try {
            const res = await axios.post('/agents/generate-token');
            setTokenInfo(res.data);
        } catch (error) {
            alert("Lỗi tạo mã: " + error.response?.data?.error);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            <button 
                onClick={handleGenerate}
                disabled={loading}
                className="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg transition shadow-lg text-sm font-bold"
            >
                <Key size={18} />
                {loading ? "Đang tạo..." : "Tạo mã cài đặt SENT"}
            </button>

            {tokenInfo && (
                <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-[9999] backdrop-blur-sm p-4 animate-in fade-in zoom-in duration-200">
                    <div className="bg-slate-800 p-6 rounded-2xl border border-slate-700 max-w-md w-full shadow-2xl relative overflow-hidden">
                        {/* Background Deco */}
                        <div className="absolute -top-10 -right-10 text-indigo-500/10 pointer-events-none">
                            <Key size={120} />
                        </div>

                        <h3 className="text-xl font-bold text-white mb-2 relative z-10">Mã Cài Đặt (Enrollment Token)</h3>
                        <p className="text-slate-400 text-sm mb-6 relative z-10">
                            Sử dụng mã này để xác thực khi cài đặt SENT Sensor trên máy tính mới.
                        </p>
                        
                        <div className="bg-slate-900 p-5 rounded-xl text-center border border-indigo-500/50 mb-6 shadow-inner relative z-10">
                            <span className="text-3xl font-mono text-indigo-400 font-bold select-all tracking-wider">
                                {tokenInfo.enroll_token}
                            </span>
                        </div>

                        <div className="bg-red-500/10 border border-red-500/30 p-3.5 rounded-xl mb-6 relative z-10">
                            <p className="text-red-400 text-sm flex items-center gap-2 font-bold mb-1 uppercase tracking-widest">
                                <span className="animate-pulse">⚠️</span> Cảnh báo bảo mật
                            </p>
                            <p className="text-xs text-red-300 leading-relaxed">
                                Mã này sẽ <strong>tự động hủy sau đúng 15 phút</strong>. Vui lòng không chia sẻ mã này ra ngoài phạm vi phòng IT.
                            </p>
                        </div>

                        <button 
                            onClick={() => setTokenInfo(null)}
                            className="w-full py-3 bg-slate-700 hover:bg-slate-600 text-white font-bold rounded-xl transition relative z-10 uppercase tracking-widest text-sm"
                        >
                            Đã hiểu & Đóng
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default GenerateTokenButton;