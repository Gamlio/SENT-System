import React, { useState, useEffect } from 'react';
import { Layers, FileText, Shield, X, Loader2 } from 'lucide-react';

const riskWeightOptions = [
    { label: 'Thấp (Máy trạm thông thường)', value: 1.0 },
    { label: 'Trung bình (Máy trạm IT/DEV)', value: 1.5 },
    { label: 'Cao (Máy chủ ứng dụng)', value: 2.0 },
    { label: 'Rất cao (Máy chủ dữ liệu/DC)', value: 3.0 },
];

const TypeFormModal = ({ type, onSubmit, onClose }) => {
    const [name, setName] = useState('');
    const [description, setDescription] = useState('');
    const [riskWeight, setRiskWeight] = useState(1.0);
    const [isSaving, setIsSaving] = useState(false);

    useEffect(() => {
        if (type) {
            setName(type.name || '');
            setDescription(type.description || '');
            setRiskWeight(type.risk_weight || 1.0);
        }
    }, [type]);

    const handleSubmit = async (e) => {
        e.preventDefault();
        setIsSaving(true);
        
        // Tạo cấu trúc payload đồng bộ chuẩn hóa cho database.AssetType trên Backend
        await onSubmit({
            name: name.trim(),
            description: description.trim(),
            risk_weight: parseFloat(riskWeight),
        });
        setIsSaving(false);
        onClose();
    };

    return (
        <div className="fixed inset-0 bg-black bg-opacity-70 flex justify-center items-center z-50 p-4">
            <div className="bg-[#0A101D] border border-slate-800 rounded-xl w-full max-w-lg p-6">
                <div className="flex justify-between items-center mb-4">
                    <h2 className="text-lg font-bold text-white">
                        {type ? 'Chỉnh sửa Phân loại Tài sản' : 'Tạo Phân loại Tài sản mới'}
                    </h2>
                    <button onClick={onClose} className="text-slate-400 hover:text-white">
                        <X size={20} />
                    </button>
                </div>

                <form onSubmit={handleSubmit}>
                    <div className="space-y-4">
                        <div>
                            <label htmlFor="name" className="block text-sm font-medium text-slate-300 mb-1">
                                Tên loại tài sản (Asset Tier)
                            </label>
                            <div className="relative">
                                <Layers className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" size={16} />
                                <input
                                    type="text"
                                    id="name"
                                    value={name}
                                    onChange={(e) => setName(e.target.value)}
                                    className="w-full bg-slate-900/50 border border-slate-700 rounded-md pl-10 pr-4 py-2 text-white focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                                    placeholder="VD: SERVER, DEV_MACHINE..."
                                    required
                                />
                            </div>
                        </div>

                        <div>
                            <label htmlFor="description" className="block text-sm font-medium text-slate-300 mb-1">
                                Mô tả
                            </label>
                            <div className="relative">
                                <FileText className="absolute left-3 top-3 text-slate-500" size={16} />
                                <textarea
                                    id="description"
                                    value={description}
                                    onChange={(e) => setDescription(e.target.value)}
                                    rows="3"
                                    className="w-full bg-slate-900/50 border border-slate-700 rounded-md pl-10 pr-4 py-2 text-white focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
                                    placeholder="Mô tả chức năng, vai trò của loại tài sản này"
                                />
                            </div>
                        </div>

                        <div>
                            <label htmlFor="riskWeight" className="block text-sm font-medium text-slate-300 mb-1">
                                Mức độ quan trọng / Rủi ro (Hệ số C)
                            </label>
                            <div className="relative">
                                <Shield className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" size={16} />
                                <select
                                    id="riskWeight"
                                    value={riskWeight}
                                    onChange={(e) => setRiskWeight(parseFloat(e.target.value))}
                                    className="w-full bg-slate-900/50 border border-slate-700 rounded-md pl-10 pr-4 py-2 text-white focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 appearance-none"
                                >
                                    {riskWeightOptions.map(option => (
                                        <option key={option.value} value={option.value}>
                                            {option.label} ({option.value}x)
                                        </option>
                                    ))}
                                </select>
                            </div>
                             <p className="text-xs text-slate-500 mt-1">Hệ số này sẽ nhân trực tiếp vào điểm rủi ro cuối cùng của tài sản khi hệ thống tính toán điểm rủi ro.</p>
                        </div>
                    </div>

                    <div className="mt-6 flex justify-end gap-3">
                        <button type="button" onClick={onClose} className="px-4 py-2 bg-slate-700 text-slate-200 rounded-md hover:bg-slate-600 transition-colors">Hủy</button>
                        <button type="submit" disabled={isSaving} className="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-500 transition-colors flex items-center gap-2 disabled:bg-indigo-800 disabled:cursor-not-allowed">
                            {isSaving ? <Loader2 className="animate-spin" size={16} /> : null}
                            {isSaving ? 'Đang lưu...' : 'Lưu thay đổi'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default TypeFormModal;