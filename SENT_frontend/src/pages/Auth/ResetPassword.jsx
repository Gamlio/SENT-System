import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Lock, ShieldCheck, Loader2 } from 'lucide-react';
import axios from '../../api/axios';

const ResetPassword = () => {
    const { token } = useParams();
    const navigate = useNavigate();
    const [password, setPassword] = useState('');
    const [confirm, setConfirm] = useState('');
    const [loading, setLoading] = useState(false);

    const handleReset = async (e) => {
        e.preventDefault();
        if (password !== confirm) return alert("Mật khẩu không khớp!");

        setLoading(true);
        try {
            await axios.post('auth/reset-password', { token, new_password: password });
            alert("Đổi mật khẩu thành công!");
            navigate('/login');
        } catch (err) {
            alert(err.response?.data?.error || "Lỗi thực thi!");
        } finally { setLoading(false); }
    };

    return (
        <div className="h-screen bg-[#050B14] flex items-center justify-center p-4">
            <div className="bg-[#0A101D] border border-slate-800 p-10 rounded-[2.5rem] w-full max-w-md">
                <div className="text-center mb-8">
                    <ShieldCheck className="mx-auto text-indigo-500 mb-4" size={48} />
                    <h2 className="text-2xl font-black text-white">MẬT KHẨU MỚI</h2>
                    <p className="text-slate-500 text-xs mt-2 uppercase tracking-widest">Thiết lập lại quyền truy cập SOC</p>
                </div>

                <form onSubmit={handleReset} className="space-y-4">
                    <div className="space-y-1.5">
                        <label className="text-[10px] font-black text-slate-500 uppercase ml-1">Mật khẩu mới</label>
                        <div className="relative">
                            <Lock className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-600" size={18}/>
                            <input type="password" required value={password}
                                onChange={e => setPassword(e.target.value)}
                                className="w-full p-4 pl-12 bg-[#050B14] rounded-2xl border border-slate-800 text-white outline-none focus:border-indigo-500" />
                        </div>
                    </div>
                    <div className="space-y-1.5">
                        <label className="text-[10px] font-black text-slate-500 uppercase ml-1">Xác nhận mật khẩu</label>
                        <input type="password" required value={confirm}
                            onChange={e => setConfirm(e.target.value)}
                            className="w-full p-4 bg-[#050B14] rounded-2xl border border-slate-800 text-white outline-none focus:border-indigo-500" />
                    </div>
                    <button disabled={loading} className="w-full bg-indigo-600 hover:bg-indigo-500 text-white p-4 rounded-2xl font-black text-xs transition-all">
                        {loading ? <Loader2 className="animate-spin mx-auto" /> : "CẬP NHẬT MẬT KHẨU"}
                    </button>
                </form>
            </div>
        </div>
    );
};

export default ResetPassword;