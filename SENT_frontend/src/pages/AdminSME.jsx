import { useState, useEffect } from 'react';
import { fetchOrgs } from '../api/adminSME';
// Giao diện quản lý danh sách SME
const AdminSME = () => {
    const [orgs, setOrgs] = useState([]);

    // Gọi API lấy danh sách
    useEffect(() => {
        fetchOrgs();
    }, []);

    const toggleStatus = async (id) => {
        // Logic Khóa/Mở SME để bảo trì hoặc khóa License
    };

    return (
        <div className="p-8 bg-[#0f172a] min-height-screen">
            <h1 className="text-3xl font-bold text-emerald-400 mb-6">Quản lý đối tác SME</h1>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {orgs.map(org => (
                    <div key={org.id} className="p-6 bg-slate-800/50 rounded-2xl border border-slate-700">
                        <h3 className="text-xl font-bold text-white">{org.name}</h3>
                        <p className="text-slate-400 text-sm">Ngày tham gia: {new Date(org.created_at).toLocaleDateString()}</p>
                        <div className="mt-4 flex justify-between items-center">
                            <span className={`px-3 py-1 rounded-full text-xs ${org.is_active ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'}`}>
                                {org.is_active ? 'Hoạt động' : 'Đã khóa'}
                            </span>
                            <button className="text-slate-400 hover:text-white transition">Cấu hình Rules</button>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};