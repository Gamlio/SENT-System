import React, { useState } from 'react';
import { X, Save, Zap } from 'lucide-react';

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
                        <label className="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-2 block">Hệ số rủi ro (Risk Weight)</label>
                        <div className="flex gap-3 items-center">
                            <input 
                                type="number" step="0.1" min="0.1" max="5.0"
                                value={formData.risk_weight}
                                onChange={(e) => setFormData({...formData, risk_weight: parseFloat(e.target.value)})}
                                className="flex-1 bg-[#050B14] border border-slate-700 rounded-lg px-4 py-2 text-white font-mono outline-none focus:border-indigo-500"
                            />
                            {/* Mốc chọn nhanh */}
                            <div className="flex gap-1">
                                {[1.0, 1.5, 2.0].map(val => (
                                    <button 
                                        type="button" key={val}
                                        onClick={() => setFormData({...formData, risk_weight: val})}
                                        className={`px-2 py-1 text-[10px] font-bold rounded border ${formData.risk_weight === val ? 'bg-indigo-500 border-indigo-500 text-white' : 'border-slate-700 text-slate-500'}`}
                                    >
                                        {val}x
                                    </button>
                                ))}
                            </div>
                        </div>
                        <p className="text-[10px] text-slate-600 mt-2 italic flex items-center gap-1">
                            <Zap size={10}/> Điểm rủi ro thực tế = (Điểm gốc) x {formData.risk_weight}[cite: 20]
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