import React, { useState } from 'react';
import { Bot, Loader2, Sparkles, ShieldAlert, CheckCircle, Activity } from 'lucide-react';
import axiosInstance from '../../../../api/axios';

const IncidentPlaybook = ({ incident, onUpdate }) => {
    const [analyzing, setAnalyzing] = useState(false);
    const [aiReport, setAiReport] = useState(null);

    // Gọi API thực tế xuống Backend để hỏi AI (Qwen 4B)
    const handleAskAI = async () => {
        setAnalyzing(true);
        try {
            const res = await axiosInstance.post(`/incidents/${incident.ID}/ai-analyze`);
            setAiReport(res.data.analysis);
        } catch (error) {
            console.error("Lỗi gọi AI:", error);
            setAiReport("[LỖI] Hệ thống AI Copilot đang bận hoặc không phản hồi. Vui lòng kiểm tra lại kết nối Ollama trên Server.");
        } finally {
            setAnalyzing(false);
        }
    };

    // Hàm parse Markdown/Text của AI thành UI Đẹp mắt
    const renderAIReport = (text) => {
        if (!text) return null;
        
        // Chia text thành các khối dựa trên thẻ [TÊN_THẺ]
        const blocks = text.split(/(?=\[.*?\])/g).filter(b => b.trim() !== '');

        return blocks.map((block, idx) => {
            const lines = block.trim().split('\n');
            const headerLine = lines[0].trim();
            const content = lines.slice(1).join('\n').trim();
            
            let colorClass = 'text-blue-400';
            let bgClass = 'bg-blue-500/10 border-blue-500/20';
            let Icon = Activity;

            if (headerLine.includes('TÓM TẮT')) {
                colorClass = 'text-emerald-400'; bgClass = 'bg-emerald-500/10 border-emerald-500/20'; Icon = CheckCircle;
            } else if (headerLine.includes('ĐÁNH GIÁ') || headerLine.includes('RỦI RO')) {
                colorClass = 'text-orange-400'; bgClass = 'bg-orange-500/10 border-orange-500/20'; Icon = ShieldAlert;
            } else if (headerLine.includes('ĐỀ XUẤT') || headerLine.includes('XỬ LÝ')) {
                colorClass = 'text-indigo-400'; bgClass = 'bg-indigo-500/10 border-indigo-500/20'; Icon = Sparkles;
            }

            return (
                <div key={idx} className={`mb-3 p-3 rounded-xl border ${bgClass}`}>
                    <h4 className={`text-xs font-black uppercase flex items-center gap-1.5 mb-2 ${colorClass}`}>
                        <Icon size={14}/> {headerLine.replace(/\[|\]/g, '')}
                    </h4>
                    <div className="text-[12px] text-slate-300 leading-relaxed whitespace-pre-wrap font-medium">
                        {content}
                    </div>
                </div>
            );
        });
    };

    return (
        <div className="mb-8">
            <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-3 flex items-center gap-2">
                <Bot size={14} className="text-indigo-400"/> AI Copilot Advisor
            </h3>
            
            <div className="bg-[#1e293b] p-1 rounded-2xl border border-indigo-500/30 shadow-[0_0_15px_rgba(99,102,241,0.15)] relative overflow-hidden">
                {/* Header dải màu gradient */}
                <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500"></div>
                
                <div className="p-4 pt-5">
                    {!aiReport && !analyzing && (
                        <div className="text-center py-6">
                            <Bot size={40} className="mx-auto text-indigo-500/50 mb-3"/>
                            <p className="text-xs text-slate-400 mb-4 px-4">
                                Khởi động AI Copilot để đọc hàng ngàn dòng log và tổng hợp báo cáo chi tiết cho sự cố này.
                            </p>
                            <button 
                                onClick={handleAskAI}
                                className="bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-bold px-6 py-2.5 rounded-xl transition shadow-lg flex items-center gap-2 mx-auto"
                            >
                                <Sparkles size={16}/> Phân tích Case này
                            </button>
                        </div>
                    )}

                    {analyzing && (
                        <div className="text-center py-8">
                            <Loader2 size={32} className="animate-spin text-indigo-400 mx-auto mb-3"/>
                            <p className="text-xs font-bold text-indigo-300 animate-pulse uppercase tracking-wider">
                                Đang trích xuất Logs & Đánh giá rủi ro...
                            </p>
                        </div>
                    )}

                    {aiReport && !analyzing && (
                        <div className="space-y-1">
                            {renderAIReport(aiReport)}
                            
                            <div className="mt-4 pt-3 border-t border-slate-700/50 flex justify-between items-center">
                                <span className="text-[10px] text-slate-500 italic">* AI chỉ mang tính chất cố vấn. Quyết định cuối cùng thuộc về Admin.</span>
                                <button 
                                    onClick={handleAskAI}
                                    className="text-[10px] font-bold text-indigo-400 hover:text-indigo-300 bg-indigo-500/10 px-3 py-1.5 rounded-lg transition"
                                >
                                    Phân tích lại
                                </button>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default IncidentPlaybook;