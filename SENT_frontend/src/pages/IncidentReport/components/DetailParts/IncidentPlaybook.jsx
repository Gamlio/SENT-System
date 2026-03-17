import React, { useState, useEffect } from 'react';
import { Bot, TerminalSquare, Loader2, CheckCircle, Lock, Cpu, ArrowRight } from 'lucide-react';
import axiosInstance from '../../../../api/axios';

const IncidentPlaybook = ({ incident, onUpdate }) => {
    const [analyzing, setAnalyzing] = useState(false);
    const [aiSteps, setAiSteps] = useState(null);

    // Giả lập AI phân tích sự cố khi vừa mở lên
    useEffect(() => {
        if (!incident || incident.status === 'Resolved') return;
        
        // Nếu chưa có Playbook AI, tự động chạy hiệu ứng phân tích
        setAnalyzing(true);
        const timer = setTimeout(() => {
            // Sinh kịch bản động dựa trên loại sự cố
            let steps = [];
            if (incident.type?.includes('Malware') || incident.type?.includes('AI_')) {
                steps = [
                    { id: 1, action: 'ISOLATE', text: 'Cô lập máy trạm khỏi mạng LAN để tránh lây lan.', icon: Lock, color: 'text-red-400', btn: 'Cô lập ngay' },
                    { id: 2, action: 'SCAN', text: 'Ra lệnh cho Agent quét toàn bộ hệ thống (Full Scan).', icon: Cpu, color: 'text-orange-400', btn: 'Quét hệ thống' }
                ];
            } else if (incident.type?.includes('USB')) {
                steps = [
                    { id: 1, action: 'EJECT', text: 'Ngắt kết nối cổng USB trái phép ngay lập tức.', icon: TerminalSquare, color: 'text-blue-400', btn: 'Ngắt cổng USB' },
                ];
            } else {
                steps = [
                    { id: 1, action: 'NOTIFY', text: 'Gửi cảnh báo đến màn hình người dùng.', icon: TerminalSquare, color: 'text-indigo-400', btn: 'Gửi cảnh báo' }
                ];
            }
            
            setAiSteps(steps);
            setAnalyzing(false);
        }, 2000); // 2s giả lập AI thinking

        return () => clearTimeout(timer);
    }, [incident.type]);

    // Hàm thực thi lệnh trực tiếp từ hướng dẫn của AI
    const executeAIAction = async (step) => {
        if(window.confirm(`Thực thi lệnh: ${step.text}? Hành động này sẽ can thiệp trực tiếp vào máy trạm.`)) {
            try {
                // Gọi API ExecuteLiveAction của Backend Go
                await axiosInstance.post(`/incidents/${incident.ID}/execute`, { command: step.action });
                alert(`Đã gửi lệnh [${step.action}] thành công!`);
                onUpdate(); // Reload lại màn hình để cập nhật Timeline
            } catch (err) {
                alert("Lỗi khi gửi lệnh xuống Agent.");
            }
        }
    };

    if (incident.status === 'Resolved') {
        return (
            <div className="bg-emerald-900/20 border border-emerald-500/30 p-4 rounded-xl text-center">
                <CheckCircle size={30} className="mx-auto text-emerald-500 mb-2"/>
                <p className="text-emerald-400 font-bold text-sm">Hồ sơ đã được đóng</p>
                <p className="text-emerald-500/70 text-xs mt-1">Playbook đã hoàn tất vòng đời.</p>
            </div>
        );
    }

    return (
        <div className="bg-[#0f172a] rounded-xl border border-indigo-500/30 overflow-hidden shadow-[0_0_20px_rgba(99,102,241,0.1)] relative">
            
            {/* Header của AI */}
            <div className="bg-indigo-900/40 p-3 border-b border-indigo-500/30 flex justify-between items-center">
                <div className="flex items-center gap-2 text-indigo-400 font-bold text-xs uppercase tracking-widest">
                    <Bot size={16} className={analyzing ? 'animate-pulse' : ''} /> 
                    AI Copilot Guidance
                </div>
                {!analyzing && <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse shadow-[0_0_8px_rgba(16,185,129,1)]"></div>}
            </div>

            <div className="p-4">
                {analyzing ? (
                    <div className="flex flex-col items-center justify-center py-6 text-indigo-400 space-y-3">
                        <Loader2 size={28} className="animate-spin"/>
                        <p className="text-xs font-mono animate-pulse">Đang phân tích ngữ cảnh sự cố...</p>
                    </div>
                ) : (
                    <div className="space-y-4 animate-in fade-in duration-500">
                        <p className="text-xs text-slate-300 leading-relaxed">
                            Dựa trên cảnh báo <span className="font-bold text-white">[{incident.type}]</span>, hệ thống AI khuyến nghị các bước phản ứng tức thời sau:
                        </p>
                        
                        <div className="space-y-3">
                            {aiSteps?.map((step) => (
                                <div key={step.id} className="bg-slate-900 border border-slate-700 p-3 rounded-lg flex flex-col gap-3 group hover:border-indigo-500/50 transition">
                                    <div className="flex items-start gap-2">
                                        <step.icon size={16} className={`${step.color} mt-0.5 shrink-0`} />
                                        <p className="text-[13px] text-slate-200 leading-snug">{step.text}</p>
                                    </div>
                                    
                                    {/* Nút Thực thi Lệnh động */}
                                    <div className="flex justify-end">
                                        <button 
                                            onClick={() => executeAIAction(step)}
                                            className="px-3 py-1.5 bg-indigo-600/20 text-indigo-400 border border-indigo-600/30 hover:bg-indigo-600 hover:text-white rounded-lg text-xs font-bold transition flex items-center gap-1.5 group-hover:shadow-[0_0_10px_rgba(99,102,241,0.3)]"
                                        >
                                            {step.btn} <ArrowRight size={12}/>
                                        </button>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

export default IncidentPlaybook;