import React, { useState } from 'react'; 
import { Building2, ArrowLeft, Loader2, Mail, User, Lock, ShieldPlus, Globe } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import axios from '../../api/axios';

const Register = () => {
    const navigate = useNavigate();
    
    const [formData, setFormData] = useState({
        company_name: '',
        email: '',
        username: '',
        password: '',
        confirmPassword: ''
    });
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const handleChange = (e) => {
        setFormData({ ...formData, [e.target.name]: e.target.value });
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (formData.password !== formData.confirmPassword) {
            return setError("Mật khẩu xác nhận không khớp!");
        }
        setError('');
        setLoading(true);

        try {
            const response = await axios.post('auth/register', {
                company_name: formData.company_name,
                email: formData.email,
                username: formData.username,
                password: formData.password
            });

           if (response.status === 200) {
                const generatedCode = response.data.company_code;
                alert(`KHỞI TẠO THÀNH CÔNG!\n\nMã Workspace của bạn là: ${generatedCode}\n\nHệ thống đã gửi hướng dẫn vào Email của bạn.`);
                navigate(`/login/${generatedCode}`); 
            }
        } catch (err) {
            setError(err.response?.data?.error || 'Lỗi kết nối hạ tầng SOC!');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="h-screen bg-[#050B14] flex items-center justify-center p-4 font-sans overflow-hidden">
            {/* Hiệu ứng trang trí Cyber cho nền */}
            <div className="absolute inset-0 opacity-5 pointer-events-none">
                <div className="absolute top-0 left-0 w-full h-full bg-[linear-gradient(to_right,#1e293b_1px,transparent_1px),linear-gradient(to_bottom,#1e293b_1px,transparent_1px)] bg-[size:40px_40px]"></div>
            </div>

            <div className="bg-[#0A101D] border border-slate-800 p-8 md:p-10 rounded-[2.5rem] w-full max-w-lg relative shadow-2xl z-10">
                
                {/* Nút Quay lại */}
                <button onClick={() => navigate('/login')} className="group flex items-center gap-2 text-slate-500 hover:text-indigo-400 transition-colors mb-6">
                    <ArrowLeft size={18} className="group-hover:-translate-x-1 transition-transform"/>
                    <span className="text-[10px] font-black uppercase tracking-widest">Quay lại đăng nhập</span>
                </button>

                <div className="text-center mb-8">
                    <div className="inline-flex p-4 bg-indigo-500/10 rounded-2xl mb-4 text-indigo-400 border border-indigo-500/20">
                        <ShieldPlus size={32} />
                    </div>
                    <h2 className="text-3xl font-black text-white tracking-tighter uppercase">Khởi tạo <span className="text-indigo-500">SOC.</span></h2>
                    <p className="text-slate-500 text-xs mt-2 font-medium tracking-wide">THIẾT LẬP KHÔNG GIAN GIÁM SÁT AN NINH CHO DOANH NGHIỆP</p>
                </div>

                {error && (
                    <div className="bg-red-500/5 border border-red-500/20 text-red-400 p-3 rounded-xl mb-6 text-xs font-bold flex items-center gap-2 animate-in fade-in duration-300">
                        <div className="w-1.5 h-1.5 rounded-full bg-red-500"></div>
                        {error}
                    </div>
                )}
                
                <form className="space-y-4" onSubmit={handleSubmit}>
                    {/* Tên công ty */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="space-y-1.5">
                            <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">Tổ chức / SME</label>
                            <div className="relative group">
                                <Building2 className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-indigo-400 transition-colors" size={16}/>
                                <input name="company_name" type="text" placeholder="Tên đơn vị" required
                                    onChange={handleChange} 
                                    className="w-full p-3.5 pl-11 bg-[#050B14] rounded-2xl border border-slate-800 text-white text-sm outline-none focus:border-indigo-500 transition-all font-bold"/>
                            </div>
                        </div>

                        {/* Email */}
                        <div className="space-y-1.5">
                            <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">Email hạ tầng</label>
                            <div className="relative group">
                                <Mail className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-indigo-400 transition-colors" size={16}/>
                                <input name="email" type="email" placeholder="admin@company.com" required
                                    onChange={handleChange} 
                                    className="w-full p-3.5 pl-11 bg-[#050B14] rounded-2xl border border-slate-800 text-white text-sm outline-none focus:border-indigo-500 transition-all font-bold"/>
                            </div>
                        </div>
                    </div>

                    {/* Tài khoản Admin */}
                    <div className="space-y-1.5">
                        <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">Tên đăng nhập (Root Admin)</label>
                        <div className="relative group">
                            <User className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-indigo-400 transition-colors" size={16}/>
                            <input name="username" type="text" placeholder="Ví dụ: admin_tech" required
                                onChange={handleChange} 
                                className="w-full p-3.5 pl-11 bg-[#050B14] rounded-2xl border border-slate-800 text-white text-sm outline-none focus:border-indigo-500 transition-all font-bold"/>
                        </div>
                    </div>

                    {/* Mật khẩu */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="space-y-1.5">
                            <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">Thiết lập mật khẩu</label>
                            <div className="relative group">
                                <Lock className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-indigo-400 transition-colors" size={16}/>
                                <input name="password" type="password" placeholder="••••••••" required
                                    onChange={handleChange} 
                                    className="w-full p-3.5 pl-11 bg-[#050B14] rounded-2xl border border-slate-800 text-white text-sm outline-none focus:border-indigo-500 transition-all"/>
                            </div>
                        </div>
                        <div className="space-y-1.5">
                            <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">Xác nhận</label>
                            <div className="relative group">
                                <Lock className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-600 group-focus-within:text-indigo-400 transition-colors" size={16}/>
                                <input name="confirmPassword" type="password" placeholder="••••••••" required
                                    onChange={handleChange} 
                                    className="w-full p-3.5 pl-11 bg-[#050B14] rounded-2xl border border-slate-800 text-white text-sm outline-none focus:border-indigo-500 transition-all"/>
                            </div>
                        </div>
                    </div>
                    
                    <button disabled={loading} className="w-full bg-indigo-600 hover:bg-indigo-500 text-white p-4 rounded-2xl font-black text-xs uppercase tracking-[0.2em] shadow-lg shadow-indigo-500/20 transition-all flex justify-center items-center gap-2 mt-4 active:scale-95 disabled:opacity-50">
                        {loading ? <Loader2 className="animate-spin" size={18} /> : (
                            <>THIẾT LẬP HỆ THỐNG <Globe size={18} /></>
                        )}
                    </button>
                </form>

                <p className="text-center mt-8 text-[9px] text-slate-600 font-bold uppercase tracking-[0.2em]">
                    Bằng việc đăng ký, bạn đồng ý với các chính sách an toàn của SENT SOC.
                </p>
            </div>
        </div>
    );
};

export default Register;