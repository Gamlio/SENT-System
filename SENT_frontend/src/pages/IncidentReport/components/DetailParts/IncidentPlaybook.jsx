import React, { useState, useEffect } from 'react';
import { CheckSquare, Square, ShieldCheck } from 'lucide-react';
import axiosInstance from '../../../../api/axios';

const IncidentPlaybook = ({ incident, onUpdate }) => {
    const [steps, setSteps] = useState([]);

    // Parse JSON từ DB khi load
    useEffect(() => {
        if (incident.playbook_progress) {
            try {
                const data = JSON.parse(incident.playbook_progress);
                setSteps(data.steps || []);
            } catch (e) {
                setSteps([]);
            }
        }
    }, [incident]);

    // Hàm toggle checkbox
    const toggleStep = async (index) => {
        const newSteps = [...steps];
        newSteps[index].done = !newSteps[index].done;
        setSteps(newSteps);

        // Gọi API lưu ngay lập tức
        try {
            await axiosInstance.put(`/incidents/${incident.ID}/playbook`, { steps: newSteps });
            onUpdate(); // Refresh để cha biết
        } catch (err) {
            alert("Lỗi lưu tiến độ!");
        }
    };

    // Tính % hoàn thành
    const progress = Math.round((steps.filter(s => s.done).length / steps.length) * 100) || 0;

    return (
        <div className="bg-[#1e293b] rounded-xl border border-slate-700 p-5 mb-6 shadow-lg">
            <div className="flex justify-between items-center mb-4">
                <h3 className="text-xs font-black text-emerald-400 uppercase tracking-widest flex items-center gap-2">
                    <ShieldCheck size={16}/> Quy trình xử lý (Playbook)
                </h3>
                <span className="text-[10px] font-bold bg-slate-800 px-2 py-1 rounded text-white border border-slate-600">
                    {progress}% Hoàn thành
                </span>
            </div>

            {/* Thanh Progress Bar */}
            <div className="w-full h-1.5 bg-slate-800 rounded-full mb-4 overflow-hidden">
                <div 
                    className="h-full bg-emerald-500 transition-all duration-500 ease-out" 
                    style={{ width: `${progress}%` }}
                ></div>
            </div>

            {/* Danh sách Steps */}
            <div className="space-y-3">
                {steps.map((step, idx) => (
                    <div 
                        key={idx} 
                        onClick={() => toggleStep(idx)}
                        className={`flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-all ${
                            step.done 
                            ? 'bg-emerald-900/10 border-emerald-500/30 text-emerald-100' 
                            : 'bg-slate-900/50 border-slate-700 hover:border-slate-500 text-slate-300'
                        }`}
                    >
                        <div className={`mt-0.5 ${step.done ? 'text-emerald-400' : 'text-slate-500'}`}>
                            {step.done ? <CheckSquare size={18}/> : <Square size={18}/>}
                        </div>
                        <div className="text-sm select-none">
                            <span className="font-bold mr-2 text-xs opacity-50">BƯỚC {step.id}:</span>
                            <span className={step.done ? 'line-through opacity-70' : ''}>{step.text}</span>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default IncidentPlaybook;