import React from 'react';
import { BookOpen, Terminal, ShieldAlert, Key, CheckCircle2, Copy, ExternalLink, Monitor } from 'lucide-react';

const InstallationGuide = () => {
    const steps = [
        {
            title: "Tải bản thực thi (Binary)",
            description: "Chọn phiên bản phù hợp với hệ điều hành của máy trạm phía trên và tải về. Đảm bảo file không bị chặn bởi trình duyệt.",
            icon: <Terminal size={18} className="text-indigo-400" />,
            badge: "Step 01"
        },
        {
            title: "Lấy mã Enrollment Token",
            description: "Truy cập vào menu 'Assets' trên Dashboard, nhấn 'Add Asset' để lấy mã Token kích hoạt duy nhất cho máy trạm.",
            icon: <Key size={18} className="text-emerald-400" />,
            isWarning: true,
            badge: "Step 02"
        },
        {
            title: "Triển khai & Kích hoạt",
            description: "Thực hiện chạy file với quyền quản trị cao nhất (sudo/Administrator) để Agent bắt đầu thu thập dữ liệu.",
            icon: <ShieldAlert size={18} className="text-orange-400" />,
            badge: "Step 03"
        }
    ];

    return (
        <div className="mt-16 relative">
             {/* Background Glow */}
            <div className="absolute -top-24 -left-24 w-64 h-64 bg-indigo-500/10 blur-[100px] rounded-full pointer-events-none"></div>

            <div className="relative bg-[#0B1224]/60 backdrop-blur-xl border border-slate-800/60 rounded-[2rem] overflow-hidden shadow-2xl">
                {/* Header Section */}
                <div className="px-8 py-6 border-b border-slate-800/60 bg-gradient-to-r from-indigo-500/10 via-transparent to-transparent flex items-center justify-between">
                    <div className="flex items-center gap-4">
                        <div className="p-2.5 bg-indigo-500/20 rounded-xl border border-indigo-500/30">
                            <BookOpen className="text-indigo-400" size={20} />
                        </div>
                        <div>
                            <h2 className="text-sm font-black uppercase tracking-[0.2em] text-white">Deployment Guide</h2>
                            <p className="text-[10px] text-slate-500 font-bold uppercase mt-0.5">Hướng dẫn cài đặt SENT Agent v4.2.6</p>
                        </div>
                    </div>
                    <div className="flex items-center gap-2 px-3 py-1 bg-slate-900/50 border border-slate-800 rounded-full">
                        <span className="w-1.5 h-1.5 bg-emerald-500 rounded-full animate-pulse"></span>
                        <span className="text-[9px] font-black text-slate-400 uppercase tracking-tighter">Documentation Live</span>
                    </div>
                </div>

                <div className="p-8 lg:p-12 grid grid-cols-1 lg:grid-cols-12 gap-12">
                    {/* Left Column: Steps */}
                    <div className="lg:col-span-5 space-y-10">
                        {steps.map((step, idx) => (
                            <div key={idx} className="group relative flex gap-6">
                                {idx !== steps.length - 1 && (
                                    <div className="absolute left-[22px] top-12 bottom-[-40px] w-px bg-gradient-to-b from-indigo-500/30 to-transparent"></div>
                                )}
                                <div className="relative z-10 flex-shrink-0 w-11 h-11 bg-slate-900 border border-slate-800 rounded-2xl flex items-center justify-center group-hover:border-indigo-500/50 transition-colors shadow-lg shadow-black/50">
                                    {step.icon}
                                </div>
                                <div className="pt-1">
                                    <div className="flex items-center gap-3 mb-2">
                                        <span className="text-[9px] font-black text-indigo-500/80 uppercase tracking-widest">{step.badge}</span>
                                        <h3 className="text-sm font-black text-slate-100 uppercase tracking-tight">{step.title}</h3>
                                    </div>
                                    <p className="text-xs text-slate-400 leading-relaxed font-medium">
                                        {step.description}
                                    </p>
                                    {step.isWarning && (
                                        <div className="mt-4 p-3 bg-emerald-500/5 border border-emerald-400/20 rounded-xl border-l-2 border-l-emerald-500">
                                            <p className="text-[10px] text-emerald-400/80 font-medium italic leading-normal">
                                                Lưu ý: Mã Token có giá trị trong vòng 24h và chỉ sử dụng được 1 lần cho 1 máy trạm duy nhất.
                                            </p>
                                        </div>
                                    )}
                                </div>
                            </div>
                        ))}
                    </div>

                    {/* Right Column: Visuals & Code */}
                    <div className="lg:col-span-7 space-y-8">
                        {/* Terminal Emulator */}
                        <div className="rounded-2xl border border-slate-800/80 bg-[#020617] overflow-hidden shadow-2xl group/term">
                            <div className="px-4 py-3 bg-slate-900/50 border-b border-slate-800 flex items-center justify-between">
                                <div className="flex gap-1.5">
                                    <div className="w-2.5 h-2.5 rounded-full bg-[#FF5F56]"></div>
                                    <div className="w-2.5 h-2.5 rounded-full bg-[#FFBD2E]"></div>
                                    <div className="w-2.5 h-2.5 rounded-full bg-[#27C93F]"></div>
                                </div>
                                <span className="text-[9px] font-mono text-slate-500 uppercase tracking-widest">SENT-CLI v4.2.6</span>
                                <Copy size={12} className="text-slate-600 hover:text-indigo-400 cursor-pointer transition-colors" />
                            </div>
                            <div className="p-6 font-mono text-[11px] leading-[1.8]">
                                <div className="flex gap-3">
                                    <span className="text-slate-600 select-none">1</span>
                                    <span className="text-emerald-500/70"># Cấp quyền thực thi file binary</span>
                                </div>
                                <div className="flex gap-3">
                                    <span className="text-slate-600 select-none">2</span>
                                    <span><span className="text-indigo-400">chmod</span> +x SENT_v4.2.6_linux</span>
                                </div>
                                <div className="flex gap-3 mt-4">
                                    <span className="text-slate-600 select-none">3</span>
                                    <span className="text-emerald-500/70"># Khởi chạy Agent bằng quyền Root</span>
                                </div>
                                <div className="flex gap-3">
                                    <span className="text-slate-600 select-none">4</span>
                                    <span><span className="text-indigo-400">sudo</span> ./SENT_v4.2.6_linux</span>
                                </div>
                            </div>
                        </div>

                        {/* Image / UI Preview */}
                        <div className="group relative rounded-2xl border border-slate-800 bg-slate-900/40 p-1.5 overflow-hidden transition-all hover:border-indigo-500/30">
                            <div className="aspect-video bg-[#050B14] rounded-xl border border-slate-800/50 flex flex-col items-center justify-center overflow-hidden relative">
                                <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent opacity-60"></div>
                                <div className="z-10 flex flex-col items-center gap-3">
                                    <div className="w-12 h-12 bg-indigo-500/10 rounded-full flex items-center justify-center border border-indigo-500/20 text-indigo-400 group-hover:scale-110 transition-transform">
                                        <Monitor size={24} />
                                    </div>
                                    <span className="text-[9px] font-black text-slate-500 uppercase tracking-widest group-hover:text-indigo-400 transition-colors">Giao diện điều khiển trung tâm (Dashboard)</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Footer Link */}
                <div className="px-8 py-5 bg-[#020617] border-t border-slate-800/60 flex flex-col sm:flex-row items-center justify-between gap-4">
                    <p className="text-[10px] text-slate-500 font-medium">
                        © 2026 SENT Security Ecosystem. Mọi thắc mắc liên hệ: <span className="text-indigo-400 hover:underline cursor-pointer">Thanh097359@gmail.com</span>
                    </p>
                    <button className="flex items-center gap-2 text-[10px] font-black text-indigo-400 uppercase tracking-tighter hover:text-white transition-colors">
                        Mở tài liệu chi tiết <ExternalLink size={12} />
                    </button>
                </div>
            </div>
        </div>
    );
};

export default InstallationGuide;