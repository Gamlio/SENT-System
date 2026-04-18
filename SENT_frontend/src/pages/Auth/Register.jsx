import React, { useState } from 'react'; 
import { Building2, ArrowLeft, Loader2,Mail } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import axios from '../../api/axios';

const Register = () => {
    const navigate = useNavigate();
    
 const [formData, setFormData] = useState({
        company_name: '',
        email: '', // [MỚI]
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
        setError('');

        setLoading(true);
        try {
            const response = await axios.post('auth/register', {
                company_name: formData.company_name,
                email: formData.email, // [MỚI]
                username: formData.username,
                password: formData.password
            });

           if (response.status === 200) {
                // 3. Sửa thông báo thành công
                const generatedCode = response.data.company_code;
                alert(`ĐĂNG KÝ THÀNH CÔNG!\n\nMã Công ty của bạn là: ${generatedCode}\n\nHệ thống đã gửi Link Đăng Nhập riêng tư vào Email của bạn. Vui lòng kiểm tra hộp thư!`);
                // Chuyển hướng thẳng đến trang đăng nhập riêng của công ty đó
                navigate(`/login/${generatedCode}`); 
            }
        } catch (err) {
            setError(err.response?.data?.error || 'Lỗi kết nối hệ thống!');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="h-screen bg-[#0f172a] flex items-center justify-center p-4 font-sans">
            <div className="bg-[#1e293b] p-8 rounded-2xl w-full max-w-md border border-slate-700 relative shadow-2xl">
                
                <button onClick={() => navigate('/login')} className="absolute top-6 left-6 text-slate-500 hover:text-emerald-400">
                    <ArrowLeft size={24} />
                </button>

                <div className="mt-4 text-center">
                    <h2 className="text-3xl font-bold text-emerald-400 mb-2">Đăng ký SME</h2>
                    <p className="text-slate-400 mb-6 text-sm">Trở thành đối tác của hệ thống SENT</p>
                </div>

                {error && <div className="bg-red-500/10 border border-red-500/50 text-red-400 p-3 rounded-lg mb-4 text-sm">{error}</div>}
                
                <form className="space-y-4" onSubmit={handleSubmit}>
                    <div className="relative">
                        <Building2 className="absolute left-3 top-3.5 text-slate-500" size={20}/>
                        <input name="company_name" type="text" placeholder="Tên công ty / Tổ chức" required
                            onChange={handleChange} className="w-full p-3 pl-10 bg-slate-900 rounded-xl border border-slate-700 text-white outline-none focus:border-emerald-500"/>
                    </div>

                   <div className="relative mt-4">
                        <Mail className="absolute left-3 top-3.5 text-slate-500" size={20}/>
                        <input name="email" type="email" placeholder="Email nhận thông báo hệ thống" required
                            onChange={handleChange} className="w-full p-3 pl-10 bg-slate-900 rounded-xl border border-slate-700 text-white outline-none focus:border-emerald-500"/>
                    </div>

                    <input name="username" type="text" placeholder="Tài khoản Admin" required
                        onChange={handleChange} className="w-full p-3 bg-slate-900 rounded-xl border border-slate-700 text-white outline-none focus:border-emerald-500"/>
                    <input name="password" type="password" placeholder="Mật khẩu" required
                        onChange={handleChange} className="w-full p-3 bg-slate-900 rounded-xl border border-slate-700 text-white outline-none focus:border-emerald-500"/>
                    <input name="confirmPassword" type="password" placeholder="Xác nhận mật khẩu" required
                        onChange={handleChange} className="w-full p-3 bg-slate-900 rounded-xl border border-slate-700 text-white outline-none focus:border-emerald-500"/>
                    
                    <button disabled={loading} className="w-full bg-emerald-500 hover:bg-emerald-600 text-white p-4 rounded-xl font-bold mt-4 transition flex justify-center items-center">
                        {loading ? <Loader2 className="animate-spin" /> : 'Tạo tài khoản hệ thống'}
                    </button>
                </form>

                <div className="mt-6 text-center text-slate-400 text-sm">
                    Đã có tài khoản? <button onClick={() => navigate('/login')} className="text-emerald-500 hover:underline">Quay lại</button>
                </div>
            </div>
        </div>
    );
};

export default Register;