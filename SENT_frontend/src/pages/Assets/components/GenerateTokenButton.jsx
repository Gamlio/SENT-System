import React, { useState, useEffect, memo } from 'react';
import { ShieldCheck, Clock, Copy, CheckCircle2, RefreshCw } from 'lucide-react';
import axios from '../../../api/axios';

// Tách riêng logic đếm giờ thành Component con và bọc React.memo
// Điều này giúp GenerateTokenButton (Cha) KHÔNG bị re-render mỗi giây!
const CountdownTimer = memo(({ initialSeconds, onExpire }) => {
    const [timeLeft, setTimeLeft] = useState(initialSeconds);

    useEffect(() => {
        setTimeLeft(initialSeconds);
    }, [initialSeconds]);

    useEffect(() => {
        if (timeLeft <= 0) return;
        const timer = setInterval(() => {
            setTimeLeft((prev) => {
                if (prev <= 1) {
                    clearInterval(timer);
                    onExpire();
                    return 0;
                }
                return prev - 1;
            });
        }, 1000);
        return () => clearInterval(timer);
    }, [timeLeft, onExpire]);

    if (timeLeft <= 0) return null;
    
    const m = Math.floor(timeLeft / 60).toString().padStart(2, '0');
    const s = (timeLeft % 60).toString().padStart(2, '0');

    return (
        <div className="flex items-center gap-1.5 px-3 py-1.5 bg-amber-500/10 text-amber-400 border border-amber-500/20 rounded-lg text-xs font-bold font-mono">
            <Clock size={14} className="animate-pulse" />
            {m}:{s}
        </div>
    );
});

const EnrollmentTokenDisplay = () => {
    const [tokenInfo, setTokenInfo] = useState(null);
    const [copied, setCopied] = useState(false);
    const [fetching, setFetching] = useState(false);

    // Hàm lấy mã từ Backend (Backend sẽ tự tạo nếu chưa có)
    const fetchToken = async () => {
        setFetching(true);
        try {
            const res = await axios.get('/assets/active-token'); 
            if (res.data && res.data.token) {
                setTokenInfo(res.data);
            }
        } catch (error) {
            console.error("Lỗi lấy mã cài đặt:", error);
        } finally {
            setFetching(false);
        }
    };

    useEffect(() => {
        fetchToken();
    }, []);

    const handleCopy = () => {
        if (tokenInfo?.token) {
            navigator.clipboard.writeText(tokenInfo.token);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        }
    };

    return (
        <div className="bg-[#1e293b] p-4 rounded-2xl border border-slate-800 shadow-lg flex flex-col md:flex-row items-start md:items-center gap-4 max-w-3xl mb-6">
            {/* Nhãn trạng thái */}
            <div className="flex items-center gap-2 px-4 py-3 bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 rounded-xl text-sm font-bold whitespace-nowrap">
                <ShieldCheck size={18} />
                Mã cài đặt hệ thống
            </div>

            {/* Khu vực hiển thị mã */}
            <div className="flex-1 flex items-center gap-3 bg-slate-900 border border-slate-700 px-4 py-3 rounded-xl w-full relative overflow-hidden">
                {fetching && <div className="absolute inset-0 bg-slate-900/50 flex items-center justify-center"><RefreshCw size={16} className="animate-spin text-indigo-400" /></div>}
                
                <span className="font-mono text-lg md:text-xl tracking-widest font-bold flex-1 text-indigo-400 select-all">
                    {tokenInfo ? tokenInfo.token : "Đang tải..."}
                </span>

                {tokenInfo && tokenInfo.expires_in > 0 && (
                    <>
                        <CountdownTimer initialSeconds={tokenInfo.expires_in} onExpire={fetchToken} />
                        <button 
                            onClick={handleCopy}
                            className="p-2 bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white rounded-lg transition"
                        >
                            {copied ? <CheckCircle2 size={18} className="text-emerald-400" /> : <Copy size={18} />}
                        </button>
                    </>
                )}
            </div>
        </div>
    );
};

export default EnrollmentTokenDisplay;