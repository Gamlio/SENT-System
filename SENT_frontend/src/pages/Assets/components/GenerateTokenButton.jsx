import React, { useState, useEffect } from 'react';
import { ShieldCheck, Clock, Copy, CheckCircle2, RefreshCw } from 'lucide-react';
import axios from '../../../api/axios';

const EnrollmentTokenDisplay = () => {
    const [tokenInfo, setTokenInfo] = useState(null);
    const [timeLeft, setTimeLeft] = useState(0);
    const [copied, setCopied] = useState(false);
    const [fetching, setFetching] = useState(false);

    // Hàm lấy mã từ Backend (Backend sẽ tự tạo nếu chưa có)
    const fetchToken = async () => {
        setFetching(true);
        try {
            const res = await axios.get('/assets/active-token'); 
            if (res.data && res.data.token) {
                setTokenInfo(res.data);
                setTimeLeft(res.data.expires_in || 0);
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

    // Logic đếm ngược và tự động lấy mã mới khi hết hạn
    useEffect(() => {
        let timer;
        if (timeLeft > 0) {
            timer = setInterval(() => {
                setTimeLeft((prev) => prev - 1);
            }, 1000);
        } else if (timeLeft === 0 && tokenInfo) {
            // Khi đếm ngược về 0, tự động gọi API lấy mã mới
            fetchToken();
        }
        return () => clearInterval(timer);
    }, [timeLeft, tokenInfo]);

    const handleCopy = () => {
        if (tokenInfo?.token) {
            navigator.clipboard.writeText(tokenInfo.token);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        }
    };

    const formatTime = (seconds) => {
        const m = Math.floor(seconds / 60).toString().padStart(2, '0');
        const s = (seconds % 60).toString().padStart(2, '0');
        return `${m}:${s}`;
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

                {timeLeft > 0 && (
                    <>
                        <div className="flex items-center gap-1.5 px-3 py-1.5 bg-amber-500/10 text-amber-400 border border-amber-500/20 rounded-lg text-xs font-bold font-mono">
                            <Clock size={14} className="animate-pulse" />
                            {formatTime(timeLeft)}
                        </div>
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