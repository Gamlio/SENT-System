import React, { useState, useEffect, memo } from 'react';
import { ShieldCheck, Clock, Copy, CheckCircle2, RefreshCw } from 'lucide-react';
import axios from '../../../api/axios';

// CountdownTimer được tinh chỉnh để giống các Tag trạng thái trong Assets.jsx
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
        <div className="flex items-center gap-1.5 px-2 py-0.5 bg-amber-500/10 text-amber-400 border border-amber-500/30 rounded text-[10px] font-black font-mono tracking-tighter">
            <Clock size={10} className="animate-pulse" />
            {m}:{s}
        </div>
    );
});

const EnrollmentTokenDisplay = () => {
    const [tokenInfo, setTokenInfo] = useState(null);
    const [copied, setCopied] = useState(false);
    const [fetching, setFetching] = useState(false);

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
        <div className="bg-[#0A101D] p-1.5 pl-3 rounded-xl border border-slate-800 shadow-2xl flex items-center gap-4 group transition-all hover:border-slate-700">
            {/* Label trái - Font style giống tiêu đề nhỏ trong Assets */}
            <div className="flex items-center gap-2">
                <div className="p-1.5 bg-indigo-500/10 border border-indigo-500/30 rounded-lg text-indigo-400">
                    <ShieldCheck size={14} />
                </div>
                <div className="hidden md:block">
                    <p className="text-[9px] font-black text-slate-500 uppercase tracking-widest leading-none mb-0.5">Enrollment</p>
                    <p className="text-[10px] font-bold text-white uppercase tracking-tight leading-none">Mã cài đặt</p>
                </div>
            </div>

            {/* Khu vực Token - Sử dụng font mono và style của IP address */}
            <div className="flex items-center gap-3 bg-[#050B14] border border-slate-800 px-3 py-1.5 rounded-lg relative overflow-hidden min-w-[180px]">
                {fetching && (
                    <div className="absolute inset-0 bg-[#050B14]/80 flex items-center justify-center z-10">
                        <RefreshCw size={12} className="animate-spin text-indigo-500" />
                    </div>
                )}
                
                <span className="font-mono text-xs md:text-sm tracking-[0.2em] font-black text-indigo-400 select-all">
                    {tokenInfo ? tokenInfo.token : "--------"}
                </span>

                <div className="flex items-center gap-2 ml-auto">
                    {tokenInfo && tokenInfo.expires_in > 0 && (
                        <CountdownTimer initialSeconds={tokenInfo.expires_in} onExpire={fetchToken} />
                    )}
                    
                    <button 
                        onClick={handleCopy}
                        className="p-1.5 hover:bg-slate-800 text-slate-500 hover:text-white rounded-md transition-colors border border-transparent hover:border-slate-700"
                        title="Copy to clipboard"
                    >
                        {copied ? <CheckCircle2 size={14} className="text-emerald-400" /> : <Copy size={14} />}
                    </button>
                </div>
            </div>
        </div>
    );
};

export default EnrollmentTokenDisplay;