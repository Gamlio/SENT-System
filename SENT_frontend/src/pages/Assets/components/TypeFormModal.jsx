import React, { useState } from 'react';
import { X, Save, Zap } from 'lucide-react';

const RISK_LEVELS = [
    { label: 'THẤP', value: 1.0, color: 'bg-emerald-500' },
    { label: 'TRUNG BÌNH', value: 1.5, color: 'bg-amber-500' },
    { label: 'CAO', value: 2.0, color: 'bg-orange-500' },
    { label: 'RẤT CAO', value: 3.0, color: 'bg-red-500' }
];

const TypeFormModal = ({ type, onClose, onSubmit }) => {
    const [formData, setFormData] = useState({
        name: type?.name || '',
        risk_weight: type?.risk_weight || 1.0,
        description: type?.description || ''
    });

    return (
        <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-[100] flex items-center justify-center p-4">
            <div className="bg-[#0A101D] w-full max-w-md rounded-2xl border border-slate-700 shadow-2xl overflow-hidden animate-in zoom-in duration-200">
                <div className="p-4 border-b border-slate-800 flex justify-between items-center bg-[#111827]">
                    <h3 className="text-xs font-black uppercase tracking-widest text-white">Chỉnh sửa: {type?.name}</h3>
                    <button onClick={onClose} className="text-slate-500 hover:text-white"><X size={18}/></button>
                </div>

                <form className="p-6 space-y-5" onSubmit={(e) => { e.preventDefault(); onSubmit(formData); }}>
                    <div>
                        <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-2 block">Tên loại tài sản (Asset Tier)</label>
                        <input 
                            type="text" required
                            value={formData.name}
                            onChange={(e) => setFormData({...formData, name: e.target.value})}
                            placeholder="VD: SERVER, DEV_MACHINE..."
                            className="w-full bg-[#050B14] border border-slate-700 rounded-lg px-4 py-2 text-white text-xs outline-none focus:border-indigo-500"
                        />
                    </div>

                    <div>
                        <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-3 block">Mức độ quan trọng / Rủi ro</label>
                        <div className="grid grid-cols-4 gap-2">
                            {RISK_LEVELS.map(level => (
                                <button 
                                    type="button" key={level.label}
                                    onClick={() => setFormData({...formData, risk_weight: level.value})}
                                    className={`flex flex-col items-center p-2 rounded-lg border transition-all ${
                                        formData.risk_weight === level.value 
                                        ? 'border-indigo-500 bg-indigo-500/10' 
                                        : 'border-slate-800 bg-[#050B14] hover:border-slate-600'
                                    }`}
                                >
                                    <div className={`w-2 h-2 rounded-full mb-2 ${level.color}`} />
                                    <span className={`text-[9px] font-black ${formData.risk_weight === level.value ? 'text-indigo-400' : 'text-slate-500'}`}>
                                        {level.label}
                                    </span>
                                    <span className="text-[10px] font-mono text-white mt-1">{level.value}x</span>
                                </button>
                            ))}
                        </div>
                        <p className="text-[10px] text-slate-600 mt-2 italic flex items-center gap-1">
                            <Zap size={10}/> Điểm rủi ro thực tế = (Điểm gốc) x {formData.risk_weight}
                        </p>
                    </div>

                    <div>
                        <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-2 block">Mô tả loại tài sản</label>
                        <textarea 
                            value={formData.description}
                            onChange={(e) => setFormData({...formData, description: e.target.value})}
                            className="w-full bg-[#050B14] border border-slate-700 rounded-lg px-4 py-2 text-white text-xs outline-none focus:border-indigo-500 h-24 resize-none"
                        />
                    </div>

                    <button type="submit" className="w-full bg-indigo-600 hover:bg-indigo-500 text-white py-3 rounded-xl font-black text-xs uppercase tracking-widest flex items-center justify-center gap-2 transition-all shadow-lg shadow-indigo-500/20">
                        <Save size={16}/> Lưu thay đổi
                    </button>
                </form>
            </div>
        </div>
    );
};
export default TypeFormModal;