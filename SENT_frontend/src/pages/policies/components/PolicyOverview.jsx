import React from 'react';
import { ShieldCheck, Globe, Laptop, List, Usb, Clock } from 'lucide-react'; // Đã xóa Activity
import { StatCard } from './PolicyShared';

const PolicyOverview = ({ policies }) => {
    // Đã xóa dòng tính toán usbCount thừa
    const softwareCount = policies.filter(p => p.category === 'SOFTWARE').length;
    const globalCount = policies.filter(p => p.target_type === 'GLOBAL').length;
    const specificCount = policies.filter(p => p.target_type === 'SPECIFIC').length;

    return (
        <div className="space-y-6 animate-fade-in">
            {/* Thống kê nhanh */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                <StatCard icon={<ShieldCheck size={32}/>} label="Tổng quy tắc" value={policies.length} color="text-emerald-400" bg="bg-emerald-500/10" />
                <StatCard icon={<Globe size={32}/>} label="Luật Toàn cục" value={globalCount} color="text-blue-400" bg="bg-blue-500/10" />
                <StatCard icon={<Laptop size={32}/>} label="Luật Ngoại lệ" value={specificCount} color="text-purple-400" bg="bg-purple-500/10" />
                <StatCard icon={<List size={32}/>} label="Phần mềm cấm" value={softwareCount} color="text-red-400" bg="bg-red-500/10" />
            </div>

            {/* Danh sách mới nhất */}
            <div className="bg-[#1e293b] rounded-3xl border border-slate-800 shadow-xl overflow-hidden">
                <div className="p-5 border-b border-slate-800 bg-slate-800/30 flex items-center gap-3">
                    <Clock className="text-slate-400" size={20}/>
                    <h3 className="font-bold text-white">Hoạt động chính sách gần đây</h3>
                </div>
                <div className="p-6 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {policies.slice(0, 6).map(p => (
                        <div key={p.ID} className="bg-slate-900/50 border border-slate-700/50 p-4 rounded-2xl flex items-start gap-3 relative overflow-hidden group hover:border-slate-600 transition">
                            <div className={`p-2.5 rounded-xl shrink-0 ${p.category === 'SOFTWARE' ? 'bg-blue-500/10 text-blue-400' : 'bg-orange-500/10 text-orange-400'}`}>
                                {p.category === 'SOFTWARE' ? <List size={18}/> : <Usb size={18}/>}
                            </div>
                            <div className="min-w-0">
                                <div className="flex items-center gap-2">
                                    <p className="font-bold text-white text-sm truncate">{p.title}</p>
                                    {p.target_type === 'GLOBAL' ? 
                                        <Globe size={12} className="text-slate-500"/> : 
                                        <Laptop size={12} className="text-purple-400"/>
                                    }
                                </div>
                                <p className="text-xs text-slate-500 truncate mt-0.5">{p.value}</p>
                                <span className={`text-[9px] font-bold uppercase mt-2 inline-block px-1.5 py-0.5 rounded ${p.policy_type === 'BLACKLIST' ? 'bg-red-500/10 text-red-400' : 'bg-emerald-500/10 text-emerald-400'}`}>
                                    {p.policy_type}
                                </span>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default PolicyOverview;