import React, { useState, useEffect } from 'react';
import { useAuth } from '../../context/AuthContext';
import { ShieldCheck, RefreshCw, User, Lock, Building2, LogIn, Radar, Shield } from 'lucide-react';
import { Link, useNavigate, useParams } from 'react-router-dom';

const generateCaptcha = () => {
    const chars = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    let result = '';
    for (let i = 0; i < 4; i++) result += chars[Math.floor(Math.random() * chars.length)];
    return result;
};

const Login = () => {
    const { companyCode: urlCompanyCode } = useParams();
    const [actualCaptcha, setActualCaptcha] = useState(generateCaptcha());
    const cachedCompanyCode = localStorage.getItem('sent_company_code') || '';
    const [loading, setLoading] = useState(false);

    const [form, setForm] = useState({ 
        company_code: urlCompanyCode || cachedCompanyCode, 
        username: '', 
        password: '', 
        captcha: '' 
    });

    const { login } = useAuth();

    const handleSubmit = async (e) => {
        e.preventDefault();
        
        if (form.captcha.toUpperCase() !== actualCaptcha) {
            alert("Mã bảo vệ (Captcha) không chính xác!");
            setActualCaptcha(generateCaptcha());
            return;
        }

        setLoading(true);
        try {
            // Lưu cache workspace
            localStorage.setItem('sent_company_code', form.company_code);
            await login(form); 
            window.location.href = "/";
        } catch (err) { 
            alert(err.response?.data?.error || "Đăng nhập thất bại. Vui lòng kiểm tra lại thông tin!");
            setActualCaptcha(generateCaptcha());
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="flex h-screen bg-[#050B14] overflow-hidden font-sans">
            {/* Banner trái: Giữ phong cách Dashboard Radar */}
            <div className="hidden lg:flex lg:w-1/2 xl:w-2/3 relative items-center justify-center border-r border-slate-800/50">
                <div className="absolute inset-0 opacity-10">
                    <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] border border-indigo-500 rounded-full animate-pulse"></div>
                    <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] h-[500px] border border-indigo-500/50 rounded-full"></div>
                </div>
                
                <div className="relative z-10 p-20">
                    <div className="flex items-center gap-3 mb-6">
                        <div className="p-3 bg-indigo-500/10 rounded-2xl text-indigo-500">
                            <Radar size={48} className="animate-spin-slow" />
                        </div>
                        <h1 className="text-4xl font-black text-white tracking-tighter">SENT <span className="text-indigo-500">SYSTEM</span></h1>
                    </div>
                    <h2 className="text-6xl font-black text-white leading-none mb-6">
                        COMMAND <br/>
                        <span className="text-transparent bg-clip-text bg-gradient-to-r from-indigo-400 to-cyan-400">CENTER.</span>
                    </h2>
                    <p className="text-slate-500 max-w-md font-medium tracking-wide uppercase text-xs">
                        Hệ thống giám sát an ninh mạng & thực thi chính sách Zero-Trust tập trung.
                    </p>
                </div>
            </div>

            {/* Form Đăng nhập */}
            <div className="w-full lg:w-1/2 xl:w-1/3 flex items-center justify-center p-8 bg-[#050B14]">
                <div className="w-full max-w-md">
                    <div className="bg-[#0A101D] border border-slate-800 p-10 rounded-[2rem] shadow-2xl">
                        <div className="mb-8">
                            <div className="flex items-center gap-2 text-indigo-400 mb-2">
                                <Shield size={16} />
                                <span className="text-[10px] font-black uppercase tracking-[0.3em]">Secure Access</span>
                            </div>
                            <h2 className="text-2xl font-bold text-white tracking-tight">Xác thực danh tính</h2>
                        </div>

                        <form onSubmit={handleSubmit} className="space-y-5">
                            {/* Workspace */}
                            <div className="space-y-1.5">
                                <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">Mã không gian làm việc</label>
                                <div className="relative group">
                                    <Building2 className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-indigo-400 transition-colors" size={18}/>
                                    <input type="text" required
                                        value={form.company_code}
                                        className="w-full p-4 pl-12 bg-[#050B14] rounded-2xl border border-slate-800 text-white focus:border-indigo-500 outline-none transition-all font-bold"
                                        placeholder="COMPANY-CODE"
                                        onChange={e => setForm({...form, company_code: e.target.value.toUpperCase()})} />
                                </div>
                            </div>

                            {/* Username */}
                            <div className="space-y-1.5">
                                <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">Tài khoản quản trị</label>
                                <div className="relative group">
                                    <User className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-indigo-400 transition-colors" size={18}/>
                                    <input type="text" required
                                        value={form.username}
                                        className="w-full p-4 pl-12 bg-[#050B14] rounded-2xl border border-slate-800 text-white focus:border-indigo-500 outline-none transition-all"
                                        placeholder="Admin username"
                                        onChange={e => setForm({...form, username: e.target.value})} />
                                </div>
                            </div>

                            {/* Password */}
                            <div className="space-y-1.5">
                                <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">Mật khẩu</label>
                                <div className="relative group">
                                    <Lock className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-indigo-400 transition-colors" size={18}/>
                                    <input type="password" required
                                        value={form.password}
                                        className="w-full p-4 pl-12 bg-[#050B14] rounded-2xl border border-slate-800 text-white focus:border-indigo-500 outline-none transition-all"
                                        placeholder="••••••••"
                                        onChange={e => setForm({...form, password: e.target.value})} />
                                </div>
                            </div>

                            {/* Captcha */}
                            <div className="flex items-center gap-3 bg-[#050B14] p-2 rounded-2xl border border-slate-800 group focus-within:border-indigo-500 transition-all">
                                <div className="bg-indigo-500/10 px-4 py-2 rounded-xl font-mono text-xl font-black tracking-[0.2rem] text-indigo-400 select-none border border-indigo-500/20">
                                    {actualCaptcha}
                                </div>
                                <input type="text" placeholder="Captcha" required
                                    value={form.captcha}
                                    className="flex-1 bg-transparent p-2 outline-none text-white text-sm uppercase font-bold"
                                    onChange={e => setForm({...form, captcha: e.target.value.toUpperCase()})} />
                                <button type="button" onClick={() => setActualCaptcha(generateCaptcha())} className="p-2 text-slate-600 hover:text-indigo-400 transition">
                                    <RefreshCw size={18} />
                                </button>
                            </div>

                            <button type="submit" disabled={loading} className="w-full bg-indigo-600 hover:bg-indigo-500 text-white p-4 rounded-2xl font-black text-xs uppercase tracking-[0.2em] shadow-lg shadow-indigo-500/20 transition-all flex justify-center items-center gap-2 group active:scale-95 disabled:opacity-50">
                                {loading ? <RefreshCw className="animate-spin" size={18} /> : (
                                    <>KÍCH HOẠT PHIÊN LÀM VIỆC <LogIn size={18} className="group-hover:translate-x-1 transition-transform" /></>
                                )}
                            </button>

                            <div className="pt-4 flex flex-col items-center gap-4">
                                <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">
                                    Chưa có tài khoản? 
                                    <Link to="/register" className="text-indigo-400 ml-2 hover:underline">Đăng ký SME</Link>
                                </p>
                            </div>
                        </form>
                    </div>
                    
                    <p className="text-center mt-8 text-[9px] text-slate-600 font-bold uppercase tracking-[0.3em]">
                        SENT SYSTEM V4.0 &copy; 2026 ALPHA SECURITY
                    </p>
                </div>
            </div>
        </div>
    );
};

export default Login;