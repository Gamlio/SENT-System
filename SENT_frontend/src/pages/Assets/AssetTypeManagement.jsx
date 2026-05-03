import React, { useState } from 'react';
import { Layers, Edit3, ShieldAlert, Info, ArrowLeft } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAssetTypes } from './hooks/useAssetTypes';
import TypeFormModal from './components/TypeFormModal';

const AssetTypeManagement = () => {
    const navigate = useNavigate();
    const { types, isLoading, updateType } = useAssetTypes();
    const [selectedType, setSelectedType] = useState(null);
    const [showModal, setShowModal] = useState(false);

    return (
        <div className="p-6 bg-[#050B14] min-h-full text-slate-200 font-sans">
            <div className="mb-8 flex justify-between items-start">
                <div>
                    <button 
                        onClick={() => navigate('/assets')}
                        className="text-slate-500 hover:text-white flex items-center gap-2 text-xs font-black uppercase mb-4 transition"
                    >
                        <ArrowLeft size={14}/> Trở lại danh sách máy
                    </button>
                    <h1 className="text-2xl font-black text-white flex items-center gap-3 uppercase tracking-tighter">
                        <Layers className="text-indigo-500" size={28}/> 
                        Quản lý Loại tài sản
                    </h1>
                    <p className="text-xs text-slate-500 font-bold uppercase tracking-widest mt-1">
                        Thiết lập hệ số rủi ro (Risk Weight) cho từng nhóm thiết bị
                    </p>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                {types.map((type) => (
                    <div key={type.id} className="bg-[#0A101D] border border-slate-800 rounded-2xl p-5 hover:border-indigo-500/50 transition-all group relative overflow-hidden">
                        <div className="flex justify-between items-start mb-4">
                            <div className="p-3 bg-indigo-500/10 rounded-xl border border-indigo-500/20 text-indigo-400">
                                {/* Render Icon động dựa trên type.icon */}
                                <ShieldAlert size={24}/>
                            </div>
                            <button 
                                onClick={() => { setSelectedType(type); setShowModal(true); }}
                                className="p-2 hover:bg-slate-800 rounded-lg text-slate-500 hover:text-white transition"
                            >
                                <Edit3 size={16}/>
                            </button>
                        </div>

                        <h3 className="font-black text-white uppercase tracking-tight text-lg">{type.name}</h3>
                        <p className="text-xs text-slate-500 mb-4 line-clamp-2">{type.description || 'Chưa có mô tả'}</p>
                        
                        <div className="flex items-center justify-between border-t border-slate-800 pt-4">
                            <span className="text-[10px] font-black text-slate-500 uppercase tracking-widest">Hệ số rủi ro</span>
                            <span className="text-xl font-mono font-black text-indigo-400">x{type.risk_weight}</span>
                        </div>
                    </div>
                ))}
            </div>

            {showModal && (
                <TypeFormModal 
                    type={selectedType} 
                    onClose={() => setShowModal(false)} 
                    onSubmit={async (data) => {
                        const res = await updateType(selectedType.id, data);
                        if (res.success) setShowModal(false);
                    }}
                />
            )}
        </div>
    );
};

export default AssetTypeManagement;