import React, { useState } from 'react'; //
import { Building2, ArrowLeft, Loader2 } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios'; // Đừng quên cài đặt: npm install axios

const Register = () => {
    const navigate = useNavigate();
    
    // 1. Quản lý trạng thái form
    const [formData, setFormData] = useState({
        company_name: '',
        username: '',
        password: '',
        confirmPassword: ''
    });
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    // 2. Hàm xử lý thay đổi input
    const handleChange = (e) => {
        setFormData({ ...formData, [e.target.name]: e.target.value });
    };

    // 3. Hàm đẩy dữ liệu về Backend (FastAPI)
    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');

        // Kiểm tra bảo mật cơ bản (Đúng chất dân Security Foxconn)
        if (formData.password !== formData.confirmPassword) {
            setError('Mật khẩu xác nhận không khớp!');
            return;
        }

        setLoading(true);
        try {
            // Đẩy dữ liệu qua cổng 8000
            const response = await axios.post('http://localhost:8000/api/v1/auth/register', {
                company_name: formData.company_name,
                username: formData.username,
                password: formData.password
            });

            if (response.status === 200) {
                alert('Đăng ký SME thành công! Hãy đăng nhập.');
                navigate('/login'); // Chuyển về trang đăng nhập
            }
        } catch (err) {
            setError(err.response?.data?.detail || 'Lỗi kết nối hệ thống!');
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