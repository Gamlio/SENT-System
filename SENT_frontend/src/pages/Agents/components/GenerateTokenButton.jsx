import React, { useState, useEffect } from 'react';
import { Key, Clock, Copy, CheckCircle2 } from 'lucide-react';
import axios from '../../../api/axios'; // Đảm bảo đường dẫn đúng với dự án của bạn

const GenerateTokenButton = () => {
    const [tokenInfo, setTokenInfo] = useState(null);
    const [loading, setLoading] = useState(false);
    const [timeLeft, setTimeLeft] = useState(0); // Thời gian còn lại tính bằng giây
    const [copied, setCopied] = useState(false);

    // 1. Lấy mã hiện tại đang active (nếu có) khi vừa vào trang
    const fetchActiveToken = async () => {
        try {
            // Thay đổi URL này cho đúng với Router GET GetActiveEnrollmentToken của bạn
            const res = await axios.get('/agents/active-token'); 
            if (res.data && res.data.token) {
                setTokenInfo(res.data);
                
                // Sử dụng expires_in từ server (nếu có) để đếm ngược chính xác, bỏ qua lỗi lệch giờ
                if (res.data.expires_in !== undefined) {
                    setTimeLeft(res.data.expires_in);
                } else {
                    calculateTimeLeft(res.data.expires_at);
                }
            }
        } catch (error) {
            console.log("Không có token active hoặc lỗi lấy token:", error);
        }
    };

    useEffect(() => {
        fetchActiveToken();
    }, []);

    // 2. Hàm tính toán số giây còn lại
    const calculateTimeLeft = (expiresAt) => {
        const expiryTime = new Date(expiresAt).getTime();
        const now = new Date().getTime();
        const diff = Math.floor((expiryTime - now) / 1000);
        setTimeLeft(diff > 0 ? diff : 0);
    };

    // 3. Logic đếm ngược (Countdown)
    useEffect(() => {
        let timer;
        if (timeLeft > 0) {
            timer = setInterval(() => {
                setTimeLeft((prevTime) => {
                    if (prevTime <= 1) {
                        setTokenInfo(null); // Xóa token khi hết hạn
                        return 0;
                    }
                    return prevTime - 1;
                });
            }, 1000);
        }
        return () => clearInterval(timer);
    }, [timeLeft]);

    // 4. Xử lý tạo mã mới
    const handleGenerate = async () => {
        setLoading(true);
        try {
            const res = await axios.post('/agents/generate-token');
            setTokenInfo(res.data);
            if (res.data.expires_in !== undefined) {
                setTimeLeft(res.data.expires_in);
            } else {
                calculateTimeLeft(res.data.expires_at);
            }
        } catch (error) {
            alert("Lỗi tạo mã: " + (error.response?.data?.error || "Đã xảy ra lỗi"));
        } finally {
            setLoading(false);
        }
    };

    // 5. Tính năng Copy mã
    const handleCopy = () => {
        if (tokenInfo?.token) {
            navigator.clipboard.writeText(tokenInfo.token);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        }
    };

    // Hàm chuyển đổi giây sang định dạng MM:SS
    const formatTime = (seconds) => {
        const m = Math.floor(seconds / 60).toString().padStart(2, '0');
        const s = (seconds % 60).toString().padStart(2, '0');
        return `${m}:${s}`;
    };

    const isTokenActive = timeLeft > 0 && tokenInfo !== null;

    return (
        <div className="bg-[#1e293b] p-4 rounded-2xl border border-slate-800 shadow-lg flex flex-col md:flex-row items-start md:items-center gap-4 max-w-3xl mb-6">
            {/* Nút Tạo Mã (Bị Disable nếu mã đang active) */}
            <button 
                onClick={handleGenerate}
                disabled={loading || isTokenActive}
                className={`flex items-center justify-center gap-2 px-5 py-3 rounded-xl transition shadow-lg text-sm font-bold whitespace-nowrap ${
                    isTokenActive 
                        ? 'bg-slate-800 text-slate-500 cursor-not-allowed border border-slate-700' 
                        : 'bg-indigo-600 hover:bg-indigo-700 text-white'
                }`}
            >
                <Key size={18} />
                {loading ? "Đang tạo..." : "Tạo mã cài đặt SENT"}
            </button>

            {/* Khu vực hiển thị mã */}
            <div className="flex-1 flex items-center gap-3 bg-slate-900 border border-slate-700 px-4 py-3 rounded-xl w-full">
                <span className={`font-mono text-lg md:text-xl tracking-widest font-bold flex-1 ${isTokenActive ? 'text-indigo-400 select-all' : 'text-slate-700 select-none'}`}>
                    {isTokenActive ? tokenInfo.token : "---- ---- ----"}
                </span>

                {/* Giao diện đếm ngược & nút copy (Chỉ hiện khi có mã) */}
                {isTokenActive && (
                    <>
                        <div className="flex items-center gap-1.5 px-3 py-1.5 bg-red-500/10 text-red-400 border border-red-500/20 rounded-lg text-xs font-bold font-mono">
                            <Clock size={14} className="animate-pulse" />
                            {formatTime(timeLeft)}
                        </div>
                        <button 
                            onClick={handleCopy}
                            className="p-2 bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white rounded-lg transition"
                            title="Copy mã"
                        >
                            {copied ? <CheckCircle2 size={18} className="text-emerald-400" /> : <Copy size={18} />}
                        </button>
                    </>
                )}
            </div>
        </div>
    );
};

export default GenerateTokenButton;
