import React, { useState, useEffect } from 'react';
import { useAuth } from '../../context/AuthContext';
import '../../styles/auth.css';
import { ShieldCheck, Key, RefreshCw, User, Lock, ArrowRight, Building2 } from 'lucide-react';
import { Link, useNavigate, useParams } from 'react-router-dom';

// 1. Hàm tạo chuỗi ngẫu nhiên 4 ký tự cho CAPTCHA
const generateCaptcha = () => {
    const chars = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    let result = '';
    for (let i = 0; i < 4; i++) {
        result += chars[Math.floor(Math.random() * chars.length)];
    }
    return result;
};

const Login = () => {
    const { companyCode: urlCompanyCode } = useParams();
    const navigate = useNavigate();

    const [step, setStep] = useState(1);
    
    // 2. Tạo State lưu mã CAPTCHA hiện tại trên màn hình
    const [actualCaptcha, setActualCaptcha] = useState(generateCaptcha());
    
    // Lấy mã công ty từ Cache (nếu có) để UX mượt hơn
    const cachedCompanyCode = localStorage.getItem('sent_company_code') || '';

    // Khởi tạo form, ưu tiên lấy companyCode từ URL -> Cache -> Rỗng
    const [form, setForm] = useState({ 
        company_code: urlCompanyCode || cachedCompanyCode, 
        username: '', 
        password: '', 
        captcha: '', 
        otp: '' 
    });

    const { login } = useAuth();

    useEffect(() => {
        if (urlCompanyCode) {
            setForm(prev => ({ ...prev, company_code: urlCompanyCode }));
        }
    }, [urlCompanyCode]);

    const handleNext = (e) => {
        e.preventDefault();
        
        if (!form.company_code) {
            alert("BẮT BUỘC nhập Mã Công Ty (Workspace) để tiếp tục!");
            return;
        }

        // 3. KIỂM TRA CAPTCHA (Ép viết hoa để dễ so sánh)
        if (form.captcha.toUpperCase() !== actualCaptcha) {
            alert("Mã bảo vệ không đúng! Vui lòng nhập lại.");
            // Tự động random mã mới và xóa ô input để bắt gõ lại
            setActualCaptcha(generateCaptcha());
            setForm(prev => ({ ...prev, captcha: '' }));
            return;
        }
        
        // 4. Nếu qua ải thành công -> Lưu luôn mã công ty vào Cache trình duyệt
        localStorage.setItem('sent_company_code', form.company_code);
        setStep(2);
    };

    const handleFinalLogin = async (e) => {
        e.preventDefault();
        try {
            await login(form); 
            window.location.href = "/";
        } catch (err) { 
            alert(err.response?.data?.error || "Mã OTP sai hoặc lỗi hệ thống"); 
        }
    };

    // Hàm làm mới Captcha khi bấm nút xoay
    const refreshCaptcha = () => {
        setActualCaptcha(generateCaptcha());
        setForm(prev => ({ ...prev, captcha: '' }));
    };

    return (
        <div className="flex h-screen bg-[#0f172a] overflow-hidden">
            {/* ... (Phần Banner bên trái giữ nguyên) ... */}
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

            <div className="w-full lg:w-1/2 xl:w-1/3 flex items-center justify-center p-8">
                <div className="w-full max-w-md bg-slate-800/50 backdrop-blur-xl p-10 rounded-[2.5rem] border border-slate-700 shadow-2xl relative overflow-hidden">
                    
                    <div className="text-center mb-8">
                        <div className="inline-flex p-4 bg-emerald-500/10 rounded-2xl mb-4 text-emerald-400">
                            <ShieldCheck size={40} />
                        </div>
                        <h2 className="text-3xl font-bold text-white tracking-tight">
                            {urlCompanyCode ? `Workspace: ${urlCompanyCode}` : 'Chào mừng trở lại'}
                        </h2>
                        <p className="text-slate-400 text-sm mt-2">Vui lòng nhập thông tin để truy cập hệ thống</p>
                    </div>

                    {step === 1 ? (
                        <form onSubmit={handleNext} className="space-y-6 animate-in slide-in-from-left-4 duration-300">
                            
                            <div className="space-y-2">
                                <label className="text-xs font-bold text-slate-400 uppercase tracking-widest ml-1">Mã công ty (Workspace)</label>
                                <div className="relative">
                                    <Building2 className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500" size={18}/>
                                    <input type="text" required
                                        disabled={!!urlCompanyCode} 
                                        value={form.company_code}
                                        className="w-full p-4 pl-12 bg-slate-900/50 rounded-2xl border border-slate-700 text-white focus:ring-2 focus:ring-emerald-500 outline-none transition-all uppercase disabled:opacity-50 disabled:cursor-not-allowed"
                                        placeholder="VD: FPT-123"
                                        onChange={e => setForm({...form, company_code: e.target.value.toUpperCase()})} />
                                </div>
                            </div>

                            <div className="space-y-2">
                                <label className="text-xs font-bold text-slate-400 uppercase tracking-widest ml-1">Tài khoản</label>
                                <div className="relative">
                                    <User className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500" size={18}/>
                                    <input type="text" required
                                        value={form.username}
                                        className="w-full p-4 pl-12 bg-slate-900/50 rounded-2xl border border-slate-700 text-white focus:ring-2 focus:ring-emerald-500 outline-none transition-all"
                                        placeholder="Tên đăng nhập"
                                        onChange={e => setForm({...form, username: e.target.value})} />
                                </div>
                            </div>

                            <div className="space-y-2">
                                <div className="flex justify-between items-center">
                                    <label className="text-xs font-bold text-slate-400 uppercase tracking-widest ml-1">Mật khẩu</label>
                                    <button type="button" className="text-xs text-emerald-400 hover:underline">Quên mật khẩu?</button>
                                </div>
                                <div className="relative">
                                    <Lock className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500" size={18}/>
                                    <input type="password" required
                                        value={form.password}
                                        className="w-full p-4 pl-12 bg-slate-900/50 rounded-2xl border border-slate-700 text-white focus:ring-2 focus:ring-emerald-500 outline-none transition-all"
                                        placeholder="••••••••"
                                        onChange={e => setForm({...form, password: e.target.value})} />
                                </div>
                            </div>

                            {/* 5. GIAO DIỆN CAPTCHA ĐÃ ĐƯỢC LÀM MỚI */}
                            <div className="flex items-center gap-4 bg-slate-900/50 p-2 rounded-2xl border border-slate-700">
                                <div className="bg-emerald-500/10 px-4 py-2 rounded-xl font-mono text-xl font-black tracking-[0.2rem] text-emerald-400 select-none">
                                    {actualCaptcha}
                                </div>
                                <input type="text" placeholder="Nhập mã bên trái" required
                                    value={form.captcha}
                                    className="flex-1 bg-transparent p-2 outline-none text-white text-sm uppercase"
                                    onChange={e => setForm({...form, captcha: e.target.value.toUpperCase()})} />
                                <button type="button" onClick={refreshCaptcha} className="p-2 text-slate-500 hover:text-emerald-400 transition" title="Đổi mã khác">
                                    <RefreshCw size={18} />
                                </button>
                            </div>

                            <button type="submit" className="w-full bg-gradient-to-r from-emerald-500 to-teal-600 hover:from-emerald-600 hover:to-teal-700 text-white p-4 rounded-2xl font-bold shadow-lg shadow-emerald-500/20 transition-all flex justify-center items-center gap-2 group">
                                Tiếp tục <ArrowRight size={20} className="group-hover:translate-x-1 transition-transform" />
                            </button>

                            <div className="pt-2 text-center space-y-3">
                                <p className="text-sm text-slate-500">
                                    Bạn chưa có tài khoản? 
                                    <Link to="/register" className="text-emerald-400 font-bold ml-1 hover:underline">Đăng ký SME</Link>
                                </p>
                                
                                {urlCompanyCode && (
                                    <p className="text-sm text-slate-500 border-t border-slate-800 pt-3">
                                        Đổi không gian làm việc? 
                                        <button type="button" onClick={() => {
                                            navigate('/login'); 
                                            localStorage.removeItem('sent_company_code'); // Xóa cache nếu muốn đổi
                                            setForm({...form, company_code: ''});
                                        }} className="text-emerald-400 font-bold ml-1 hover:underline">Nhập mã khác</button>
                                    </p>
                                )}
                            </div>

                        </form>
                        
                    ) : (

                        /* --- LUỒNG 2: XÁC THỰC OTP --- */
                        <form onSubmit={handleFinalLogin} className="space-y-8 animate-in slide-in-from-right-4 duration-300">
                            
                            <div className="text-center">
                                <div className="inline-block p-4 bg-emerald-500/10 rounded-full text-emerald-500 mb-4 border border-emerald-500/20">
                                    <Key size={32} />
                                </div>
                                <p className="text-slate-300 font-medium">Mã bảo mật (OTP)</p>
                                <p className="text-xs text-slate-500 mt-1">Vui lòng nhập 6 số từ ứng dụng Authenticator</p>
                            </div>
                            
                            <input type="text" placeholder="000000" maxLength="6" required
                                value={form.otp}
                                className="w-full p-5 bg-slate-900/50 rounded-2xl border border-slate-700 text-center text-4xl tracking-[0.8rem] font-black text-emerald-400 outline-none focus:ring-2 focus:ring-emerald-500"
                                onChange={e => setForm({...form, otp: e.target.value})} />

                            <button type="submit" className="w-full bg-emerald-500 hover:bg-emerald-600 text-white p-4 rounded-2xl font-bold transition-all shadow-lg shadow-emerald-500/20">
                                Xác nhận Đăng nhập
                            </button>
                            
                            <button type="button" onClick={() => setStep(1)} className="w-full text-slate-500 hover:text-slate-300 text-xs font-bold uppercase tracking-widest transition flex justify-center items-center gap-1">
                                Quay lại nhập thông tin
                            </button>
                        </form>

                    )}
                </div>
            </div>
        </div>
    );
};

export default Login;