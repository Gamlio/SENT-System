import React, { useState } from 'react';
import { Bot, Loader2, Sparkles, ShieldAlert, CheckCircle, Activity, Terminal } from 'lucide-react';
import axiosInstance from '../../../api/axios';

const IncidentPlaybook = ({ incident, onUpdate }) => {
    const [analyzing, setAnalyzing] = useState(false);
    const [aiReport, setAiReport] = useState(null);

    const handleAskAI = async () => {
        setAnalyzing(true);
        try {
            const res = await axiosInstance.post(`/incidents/${incident.ID}/ai-analyze`);
            setAiReport(res.data.analysis);
        } catch (error) {
            console.error("Lỗi gọi AI:", error);
            setAiReport("[SYSTEM ERROR] Kênh kết nối AI Copilot (Ollama) bị gián đoạn. Vui lòng kiểm tra lại Node Server.");
        } finally {
            setAnalyzing(false);
        }
    };

    const renderAIReport = (text) => {
        if (!text) return null;
        const blocks = text.split(/(?=\[.*?\])/g).filter(b => b.trim() !== '');

        return blocks.map((block, idx) => {
            const lines = block.trim().split('\n');
            const headerLine = lines[0].trim();
            const content = lines.slice(1).join('\n').trim();
            
            let colorClass = 'text-blue-400 border-blue-500/30 bg-blue-500/5';
            let Icon = Activity;

            if (headerLine.includes('TÓM TẮT')) {
                colorClass = 'text-emerald-400 border-emerald-500/30 bg-emerald-500/5'; Icon = CheckCircle;
            } else if (headerLine.includes('ĐÁNH GIÁ') || headerLine.includes('RỦI RO')) {
                colorClass = 'text-orange-400 border-orange-500/30 bg-orange-500/5'; Icon = ShieldAlert;
            } else if (headerLine.includes('ĐỀ XUẤT') || headerLine.includes('XỬ LÝ')) {
                colorClass = 'text-indigo-400 border-indigo-500/30 bg-indigo-500/5'; Icon = Sparkles;
            }

            return (
                <div key={idx} className={`mb-3 p-3 rounded-lg border-l-2 border-y border-r border-y-slate-800/50 border-r-slate-800/50 ${colorClass}`}>
                    <h4 className="text-[10px] font-black uppercase flex items-center gap-1.5 mb-2 tracking-widest">
                        <Icon size={12}/> {headerLine.replace(/\[|\]/g, '')}
                    </h4>
                    <div className="text-[11px] text-slate-300 leading-relaxed whitespace-pre-wrap font-sans">
                        {content}
                    </div>
                </div>
            );
        });
    };

    return (
        <div className="mb-6 w-full">
            <h3 className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-2 flex items-center gap-2 border-b border-slate-800 pb-2">
                <Bot size={12} className="text-indigo-400"/> AI Copilot Advisor
            </h3>
            
            <div className="bg-[#0A101D] rounded-xl border border-slate-800/80 shadow-inner relative overflow-hidden group">
                <div className="absolute top-0 left-0 right-0 h-[2px] bg-gradient-to-r from-indigo-500/50 via-purple-500/50 to-emerald-500/50 opacity-50 group-hover:opacity-100 transition"></div>
                
                <div className="p-4">
                    {!aiReport && !analyzing && (
                        <div className="text-center py-4">
                            <Terminal size={24} className="mx-auto text-slate-600 mb-2"/>
                            <p className="text-[11px] text-slate-500 mb-4 px-2 font-mono">
                                [WAITING_FOR_COMMAND] Khởi chạy AI để đọc raw logs và trích xuất phương án xử lý tự động.
                            </p>
                            <button 
                                onClick={handleAskAI}
                                className="bg-indigo-500/10 border border-indigo-500/30 hover:bg-indigo-600 hover:text-white text-indigo-400 text-[10px] font-black uppercase tracking-widest px-6 py-2 rounded transition flex items-center gap-2 mx-auto"
                            >
                                <Sparkles size={14}/> Run AI Analysis
                            </button>
                        </div>
                    )}

                    {analyzing && (
                        <div className="text-center py-6 flex flex-col items-center">
                            <div className="relative mb-3">
                                <Loader2 size={24} className="animate-spin text-indigo-500 absolute"/>
                                <Bot size={24} className="text-indigo-400/50 animate-pulse"/>
                            </div>
                            <p className="text-[10px] font-mono font-bold text-indigo-400 animate-pulse tracking-widest">
                                &gt; EXTRACTING LOGS & EVALUATING RISKS...
                            </p>
                        </div>
                    )}

                    {aiReport && !analyzing && (
                        <div className="space-y-1 animate-in fade-in slide-in-from-bottom-2">
                            {renderAIReport(aiReport)}
                            
                            <div className="mt-4 pt-3 border-t border-slate-800 flex justify-between items-center">
                                <span className="text-[9px] text-slate-600 font-mono italic flex items-center gap-1">
                                    <AlertTriangle size={10}/> AI generated content. Verify before action.
                                </span>
                                <button 
                                    onClick={handleAskAI}
                                    className="text-[9px] font-black uppercase tracking-widest text-slate-400 hover:text-indigo-400 transition flex items-center gap-1"
                                >
                                    <Activity size={10}/> Re-Analyze
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