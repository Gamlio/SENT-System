import React, { useState, useRef } from 'react';
import { Plus, ShieldCheck, CheckCircle2, Globe, Users, Search, ListPlus, Type, Info, Upload, Loader2 } from 'lucide-react';
import Pagination from '../../../components/common/Pagination';
import { parseExcel, parseWord } from '../../../utils/fileParsers';

const PolicyForm = ({ currentConfig, groups, onAddPolicy, isLoading }) => {
    const [newTitle, setNewTitle] = useState('');
    const [policyType, setPolicyType] = useState('BLACKLIST');
    const [targetType, setTargetType] = useState('GLOBAL');
    const [selectedGroupID, setSelectedGroupID] = useState('');

    const [inputMode, setInputMode] = useState('SINGLE'); 
    const [singleValue, setSingleValue] = useState('');
    const [bulkValue, setBulkValue] = useState('');
    const [isReadingFile, setIsReadingFile] = useState(false);
    const fileInputRef = useRef(null);

    const handleFileImport = async (e) => {
        const file = e.target.files[0];
        if (!file) return;
        setIsReadingFile(true);
        const extension = file.name.split('.').pop().toLowerCase();
        let extractedText = "";
        try {
            if (['xlsx', 'xls'].includes(extension)) {
                extractedText = await parseExcel(file);
            } else if (['docx'].includes(extension)) {
                extractedText = await parseWord(file);
            } else {
                extractedText = await new Promise((resolve, reject) => {
                    const reader = new FileReader();
                    reader.onload = (ev) => resolve(ev.target.result);
                    reader.onerror = () => reject("Lỗi đọc file text");
                    reader.readAsText(file);
                });
            }
            if (extractedText) {
                setInputMode('BULK');
                setBulkValue(prev => {
                    const prefix = prev ? prev + "\n" : "";
                    return prefix + extractedText;
                });
                const count = extractedText.split('\n').filter(x => x.trim()).length;
                alert(`Đã nạp thành công ${count} mục từ file: ${file.name}`);
            }
        } catch (err) {
            console.error(err);
            alert("Không thể đọc file: " + err);
        } finally {
            setIsReadingFile(false);
            e.target.value = null;
        }
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        let valuesToSubmit = [];
        if (inputMode === 'SINGLE') {
            if (!singleValue.trim()) return alert("Vui lòng nhập giá trị!");
            valuesToSubmit = [singleValue.trim()];
        } else {
            valuesToSubmit = bulkValue.split(/[\n,;]+/) 
                .map(v => v.trim())
                .filter(v => v.length > 0);
            if (valuesToSubmit.length === 0) return alert("Danh sách trống!");
        }

        const basePayload = {
            title: newTitle,
            policy_type: policyType,
            target_type: targetType,
            group_id: targetType === 'GLOBAL' ? null : Number(selectedGroupID)
        };
        
        onAddPolicy({ ...basePayload, values: valuesToSubmit });

        setNewTitle('');
        setSingleValue('');
        setBulkValue('');
        setSelectedGroupID('');
    };

    return (
        <div className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl sticky top-24">
            {/* ... (Phần UI phía trên giữ nguyên) ... */}
            <div className="flex items-center gap-3 mb-6">
                <div className="p-3 bg-emerald-500/10 rounded-xl text-emerald-400"><Plus size={24}/></div>
                <div>
                    <h3 className="text-lg font-bold text-white">Thiết lập Quy tắc</h3>
                    <p className="text-xs text-slate-500">Tạo luật mới cho {currentConfig.label}</p>
                </div>
            </div>

            <form onSubmit={handleSubmit} className="space-y-5">
                <div>
                    <label className="text-[10px] font-bold text-slate-500 uppercase">Tên gợi nhớ (Nhóm)</label>
                    <input type="text" placeholder={inputMode === 'BULK' ? "VD: Danh sách đen T3/2026" : "VD: Chặn Game LOL"} value={newTitle} onChange={e => setNewTitle(e.target.value)} required className="w-full mt-1 p-3 bg-slate-900 border border-slate-700 rounded-xl text-sm text-white focus:border-emerald-500 outline-none transition" />
                </div>

                <div>
                    <div className="flex justify-between items-end mb-1">
                        <label className="text-[10px] font-bold text-slate-500 uppercase">Giá trị kỹ thuật</label>
                        <div className="flex gap-2">
                             <input type="file" ref={fileInputRef} onChange={handleFileImport} accept=".txt,.csv,.log,.xlsx,.xls,.docx" className="hidden" />
                            <div className="flex bg-slate-900 rounded-lg p-0.5 border border-slate-700">
                                <button type="button" onClick={() => setInputMode('SINGLE')} className={`px-2 py-1 rounded-md text-[10px] font-bold flex items-center gap-1 transition ${inputMode === 'SINGLE' ? 'bg-slate-700 text-white shadow' : 'text-slate-400 hover:text-white'}`}>
                                    <Type size={10}/> Đơn lẻ
                                </button>
                                <button type="button" onClick={() => setInputMode('BULK')} className={`px-2 py-1 rounded-md text-[10px] font-bold flex items-center gap-1 transition ${inputMode === 'BULK' ? 'bg-blue-600 text-white shadow' : 'text-slate-400 hover:text-white'}`}>
                                    <ListPlus size={10}/> Danh sách
                                </button>
                            </div>
                        </div>
                    </div>

                    {inputMode === 'SINGLE' ? (
                        <input type="text" placeholder={currentConfig.placeholder} value={singleValue} onChange={e => setSingleValue(e.target.value)} className="w-full p-3 bg-slate-900 border border-slate-700 rounded-xl text-sm font-mono text-white focus:border-emerald-500 outline-none transition" />
                    ) : (
                        <div className="relative group">
                            <textarea rows={6} disabled={isReadingFile} placeholder={isReadingFile ? "Đang đọc file..." : `Dán danh sách hoặc nạp file Excel/Word...\n${currentConfig.placeholder.replace('VD: ', '')}\n...`} value={bulkValue} onChange={e => setBulkValue(e.target.value)} className="w-full p-3 bg-slate-900 border border-slate-700 rounded-xl text-xs font-mono text-white focus:border-blue-500 outline-none transition custom-scrollbar resize-none disabled:opacity-50" />
                            <button type="button" onClick={() => fileInputRef.current.click()} disabled={isReadingFile} className="absolute bottom-3 right-3 flex items-center gap-1 bg-slate-800 hover:bg-slate-700 text-slate-300 px-2 py-1 rounded-lg border border-slate-600 text-[10px] font-bold transition shadow-lg disabled:opacity-50">
                                {isReadingFile ? <Loader2 size={10} className="animate-spin"/> : <Upload size={10} />} 
                                {isReadingFile ? " Đang xử lý..." : " Import (Excel/Doc)"}
                            </button>
                        </div>
                    )}
                </div>

                <div className="grid grid-cols-2 gap-3 p-1 bg-slate-900 rounded-xl border border-slate-700">
                    <button type="button" onClick={() => setPolicyType('BLACKLIST')} className={`py-2 rounded-lg text-xs font-bold transition flex items-center justify-center gap-2 ${policyType === 'BLACKLIST' ? 'bg-red-500 text-white shadow-lg' : 'text-slate-500 hover:text-white'}`}>
                        <ShieldCheck size={14}/> CẤM (Block)
                    </button>
                    <button type="button" onClick={() => setPolicyType('WHITELIST')} className={`py-2 rounded-lg text-xs font-bold transition flex items-center justify-center gap-2 ${policyType === 'WHITELIST' ? 'bg-emerald-500 text-white shadow-lg' : 'text-slate-500 hover:text-white'}`}>
                        <CheckCircle2 size={14}/> CHO PHÉP
                    </button>
                </div>

                <div className="space-y-3 pt-2 border-t border-slate-800">
                    <label className="text-[10px] font-bold text-slate-500 uppercase">Phạm vi áp dụng</label>
                    <div className="flex gap-3">
                        <div onClick={() => setTargetType('GLOBAL')} className={`flex-1 p-3 rounded-xl border cursor-pointer transition flex flex-col items-center gap-2 ${targetType === 'GLOBAL' ? 'bg-blue-500/10 border-blue-500 text-blue-400' : 'bg-slate-900 border-slate-700 text-slate-500 hover:border-slate-600'}`}>
                            <Globe size={20}/>
                            <span className="text-xs font-bold">Toàn hệ thống</span>
                        </div>
                        <div onClick={() => setTargetType('GROUP')} className={`flex-1 p-3 rounded-xl border cursor-pointer transition flex flex-col items-center gap-2 ${targetType === 'GROUP' ? 'bg-purple-500/10 border-purple-500 text-purple-400' : 'bg-slate-900 border-slate-700 text-slate-500 hover:border-slate-600'}`}>
                            <Users size={20}/>
                            <span className="text-xs font-bold">Theo Nhóm</span>
                        </div>
                    </div>
                </div>

                {targetType === 'GROUP' && (
                    <div className="bg-slate-900/50 rounded-xl border border-slate-700 animate-in fade-in slide-in-from-top-2 overflow-hidden">
                        <div className="p-3 space-y-2">
                            <select 
                                value={selectedGroupID}
                                onChange={(e) => setSelectedGroupID(e.target.value)}
                                className="w-full bg-[#050B14] border border-slate-700 rounded-lg px-3 py-2 text-xs text-white focus:border-purple-500 outline-none transition-all"
                                required
                            >
                                <option value="">-- Chọn một nhóm --</option>
                                {groups && groups.map(g => (
                                    <option key={g.id || g.ID} value={g.id || g.ID}>{g.name || g.Name}</option>
                                ))}
                            </select>
                        </div>
                    </div>
                )}

                <button type="submit" disabled={isLoading || isReadingFile} className="w-full py-3 bg-emerald-500 hover:bg-emerald-600 text-white font-bold rounded-xl shadow-lg shadow-emerald-500/20 transition disabled:opacity-50">
                    {isLoading ? 'Đang lưu...' : `Áp dụng ${inputMode === 'BULK' ? 'Danh sách' : 'Chính sách'}`}
                </button>
            </form>
        </div>
    );
};

export default PolicyForm;