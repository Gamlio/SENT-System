import React, { useState } from 'react';
import { useAuth } from '../../context/AuthContext';
import '../../styles/auth.css';
import { ShieldCheck, Key, RefreshCw, User, Lock, ArrowRight } from 'lucide-react';
import { Link } from 'react-router-dom';

const Login = () => {
    const [step, setStep] = useState(1); // 1: Login, 2: 2FA
    const [form, setForm] = useState({ username: '', password: '', captcha: '', otp: '' });
    const { login } = useAuth();

    const handleNext = (e) => {
        e.preventDefault();
        if (form.captcha !== "X8R3") {
            alert("Mã CAPTCHA không đúng!");
            return;
        }
        setStep(2);
    };

    const handleFinalLogin = async (e) => {
    e.preventDefault();
    try {
        // Truyền toàn bộ object form (có chứa cả username, password, captcha, otp)
        await login(form); 
        window.location.href = "/";
    } catch (err) { 
        alert("Mã OTP sai hoặc lỗi hệ thống"); 
    }
};

    return (
        <div className="flex h-screen bg-[#0f172a] overflow-hidden">
            
            {/* PHẦN BÊN TRÁI: BACKGROUND IMAGE (Ẩn trên mobile) */}
            <div className="hidden lg:flex lg:w-1/2 xl:w-2/3 relative">
                <img 
                    src="https://images.unsplash.com/photo-1550751827-4bd374c3f58b?q=80&w=2070" 
                    alt="Cyber Security" 
                    className="absolute inset-0 w-full h-full object-cover"
                />
                <div className="absolute inset-0 bg-gradient-to-r from-emerald-900/40 to-[#0f172a]"></div>
                <div className="relative z-10 flex flex-col justify-center p-20 text-white">
                    <h1 className="text-6xl font-black mb-4 leading-tight">BẢO VỆ <br/><span className="text-emerald-400">TÀI SẢN SỐ.</span></h1>
                    <p className="text-lg text-slate-300 max-w-md">Hệ thống giám sát an ninh tập trung SENT mang lại sự an tâm tuyệt đối cho doanh nghiệp SME.</p>
                </div>
            </div>

            {/* PHẦN BÊN PHẢI: FORM LOGIN (Bọc trong Card) */}
            <div className="w-full lg:w-1/2 xl:w-1/3 flex items-center justify-center p-8">
                <div className="w-full max-w-md bg-slate-800/50 backdrop-blur-xl p-10 rounded-[2.5rem] border border-slate-700 shadow-2xl">
                    
                    {/* Header Card */}
                    <div className="text-center mb-8">
                        <div className="inline-flex p-4 bg-emerald-500/10 rounded-2xl mb-4 text-emerald-400">
                            <ShieldCheck size={40} />
                        </div>
                        <h2 className="text-3xl font-bold text-white tracking-tight">Chào mừng trở lại</h2>
                        <p className="text-slate-400 text-sm mt-2">Vui lòng nhập thông tin để truy cập hệ thống</p>
                    </div>

                    {step === 1 ? (
                        <form onSubmit={handleNext} className="space-y-6">
                            {/* Input Tài khoản */}
                            <div className="space-y-2">
                                <label className="text-xs font-bold text-slate-400 uppercase tracking-widest ml-1">Tài khoản</label>
                                <div className="relative">
                                    <User className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500" size={18}/>
                                    <input type="text" required
                                        className="w-full p-4 pl-12 bg-slate-900/50 rounded-2xl border border-slate-700 text-white focus:ring-2 focus:ring-emerald-500 outline-none transition-all"
                                        placeholder="Tên đăng nhập"
                                        onChange={e => setForm({...form, username: e.target.value})} />
                                </div>
                            </div>

                            {/* Input Mật khẩu */}
                            <div className="space-y-2">
                                <div className="flex justify-between items-center">
                                    <label className="text-xs font-bold text-slate-400 uppercase tracking-widest ml-1">Mật khẩu</label>
                                    <button type="button" className="text-xs text-emerald-400 hover:underline">Quên mật khẩu?</button>
                                </div>
                                <div className="relative">
                                    <Lock className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500" size={18}/>
                                    <input type="password" required
                                        className="w-full p-4 pl-12 bg-slate-900/50 rounded-2xl border border-slate-700 text-white focus:ring-2 focus:ring-emerald-500 outline-none transition-all"
                                        placeholder="••••••••"
                                        onChange={e => setForm({...form, password: e.target.value})} />
                                </div>
                            </div>

                            {/* CAPTCHA */}
                            <div className="flex items-center gap-4 bg-slate-900/50 p-2 rounded-2xl border border-slate-700">
                                <div className="bg-emerald-500/10 px-4 py-2 rounded-xl font-mono text-xl font-black text-emerald-400 select-none">
                                    X8R3
                                </div>
                                <input type="text" placeholder="Mã bảo vệ" required
                                    className="flex-1 bg-transparent p-2 outline-none text-white text-sm"
                                    onChange={e => setForm({...form, captcha: e.target.value})} />
                                <RefreshCw size={18} className="text-slate-500 cursor-pointer mr-2 hover:text-emerald-400 transition" />
                            </div>

                            {/* Nút Đăng nhập */}
                            <button className="w-full bg-gradient-to-r from-emerald-500 to-teal-600 hover:from-emerald-600 hover:to-teal-700 text-white p-4 rounded-2xl font-bold shadow-lg shadow-emerald-500/20 transition-all flex justify-center items-center gap-2 group">
                                Đăng nhập ngay <ArrowRight size={20} className="group-hover:translate-x-1 transition-transform" />
                            </button>

                            {/* Link Đăng ký */}
                            <p className="text-center text-sm text-slate-500">
                                Bạn chưa có tài khoản? 
                                <Link to="/register" className="text-emerald-400 font-bold ml-1 hover:underline">Đăng ký SME</Link>
                            </p>
                        </form>
                    ) : (
                        <form onSubmit={handleFinalLogin} className="space-y-8 animate-in fade-in duration-500">
                            <div className="text-center">
                                <div className="inline-block p-4 bg-emerald-500/10 rounded-full text-emerald-500 mb-4 border border-emerald-500/20">
                                    <Key size={32} />
                                </div>
                                <p className="text-slate-300 font-medium">Xác thực hai yếu tố</p>
                            </div>
                            
                            <input type="text" placeholder="000000" maxLength="6" required
                                className="w-full p-5 bg-slate-900/50 rounded-2xl border border-slate-700 text-center text-4xl tracking-[0.8rem] font-black text-emerald-400 outline-none focus:ring-2 focus:ring-emerald-500"
                                onChange={e => setForm({...form, otp: e.target.value})} />

                            <button className="w-full bg-emerald-500 hover:bg-emerald-600 text-white p-4 rounded-2xl font-bold transition-all shadow-lg shadow-emerald-500/20">
                                Xác nhận OTP
                            </button>
                            <button type="button" onClick={() => setStep(1)} className="w-full text-slate-500 hover:text-slate-300 text-xs font-bold uppercase tracking-widest transition">
                                Quay lại đăng nhập
                            </button>
                        </form>
                    )}
                </div>
            </div>
        </div>
    );
};

export default Login;