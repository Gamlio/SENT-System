import React from 'react';
import { Building2 } from 'lucide-react';

const Register = () => {
    return (
        <div className="h-screen bg-[#0f172a] flex items-center justify-center p-4">
            <div className="bg-[#1e293b] p-8 rounded-2xl w-full max-w-md border border-slate-700">
                <h2 className="text-3xl font-bold text-emerald-400 mb-2">Đăng ký SME</h2>
                <p className="text-slate-400 mb-8">Bắt đầu bảo vệ doanh nghiệp của bạn với SENT</p>
                
                <form className="space-y-4">
                    <div className="relative">
                        <Building2 className="absolute left-3 top-3.5 text-slate-500" size={20}/>
                        <input type="text" placeholder="Tên công ty / Tổ chức" 
                            className="w-full p-3 pl-10 bg-slate-900 rounded-xl border border-slate-700 outline-none focus:border-emerald-500 text-white"/>
                    </div>
                    <input type="text" placeholder="Tài khoản Admin" className="w-full p-3 bg-slate-900 rounded-xl border border-slate-700 outline-none focus:border-emerald-500 text-white"/>
                    <input type="password" placeholder="Mật khẩu" className="w-full p-3 bg-slate-900 rounded-xl border border-slate-700 outline-none focus:border-emerald-500 text-white"/>
                    <input type="password" placeholder="Xác nhận mật khẩu" className="w-full p-3 bg-slate-900 rounded-xl border border-slate-700 outline-none focus:border-emerald-500 text-white"/>
                    
                    <button className="w-full bg-emerald-500 hover:bg-emerald-600 text-white p-4 rounded-xl font-bold mt-4 transition">
                        Tạo tài khoản hệ thống
                    </button>
                </form>
            </div>
        </div>
    );
};
export default Register;