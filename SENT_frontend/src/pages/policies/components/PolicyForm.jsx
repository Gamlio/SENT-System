import React, { useState, useRef } from 'react';
import { Plus, ShieldCheck, CheckCircle2, Globe, Laptop, Search, ListPlus, Type, Info, Upload, Loader2 } from 'lucide-react';
import Pagination from '../../../components/common/Pagination';
// IMPORT HÀM XỬ LÝ FILE
import { parseExcel, parseWord } from '../../../utils/fileParsers';

const PolicyForm = ({ currentConfig, agents, onAddPolicy, isLoading }) => {
    // --- STATE FORM CƠ BẢN ---
    const [newTitle, setNewTitle] = useState('');
    const [policyType, setPolicyType] = useState('BLACKLIST');
    const [targetType, setTargetType] = useState('GLOBAL');
    const [selectedHWIDs, setSelectedHWIDs] = useState([]);

    // --- STATE NHẬP LIỆU (BULK/SINGLE) ---
    const [inputMode, setInputMode] = useState('SINGLE'); 
    const [singleValue, setSingleValue] = useState('');
    const [bulkValue, setBulkValue] = useState('');
    
    // State loading riêng cho việc đọc file
    const [isReadingFile, setIsReadingFile] = useState(false);
    
    const fileInputRef = useRef(null);

    // --- STATE TÌM KIẾM MÁY ---
    const [agentSearch, setAgentSearch] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 5;

    // Logic lọc và phân trang máy trạm
    const filteredAgents = agents.filter(a => 
        a.hostname?.toLowerCase().includes(agentSearch.toLowerCase()) || 
        a.ip_address?.includes(agentSearch)
    );
    const totalPages = Math.ceil(filteredAgents.length / itemsPerPage);
    const displayedAgents = filteredAgents.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);

    // --- HÀM XỬ LÝ IMPORT FILE ---
    const handleFileImport = async (e) => {
        const file = e.target.files[0];
        if (!file) return;

        setIsReadingFile(true);
        const extension = file.name.split('.').pop().toLowerCase();
        let extractedText = "";

        try {
            if (['xlsx', 'xls'].includes(extension)) {
                // Xử lý Excel
                extractedText = await parseExcel(file);
            } else if (['docx'].includes(extension)) {
                // Xử lý Word (.docx)
                extractedText = await parseWord(file);
            } else {
                // Xử lý Text/CSV/Log (Mặc định)
                extractedText = await new Promise((resolve, reject) => {
                    const reader = new FileReader();
                    reader.onload = (ev) => resolve(ev.target.result);
                    reader.onerror = () => reject("Lỗi đọc file text");
                    reader.readAsText(file);
                });
            }

            if (extractedText) {
                // Tự động chuyển sang chế độ Danh sách (BULK) và điền dữ liệu
                setInputMode('BULK');
                setBulkValue(prev => {
                    const prefix = prev ? prev + "\n" : "";
                    return prefix + extractedText;
                });
                
                // Thông báo kết quả
                const count = extractedText.split('\n').filter(x => x.trim()).length;
                alert(`Đã nạp thành công ${count} mục từ file: ${file.name}`);
            }
        } catch (err) {
            console.error(err);
            alert("Không thể đọc file: " + err);
        } finally {
            setIsReadingFile(false);
            e.target.value = null; // Reset input để người dùng có thể chọn lại file cũ nếu muốn
        }
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        
        let valuesToSubmit = [];
        if (inputMode === 'SINGLE') {
            if (!singleValue.trim()) return alert("Vui lòng nhập giá trị!");
            valuesToSubmit = [singleValue.trim()];
        } else {
            // Tách dòng thông minh (hỗ trợ dấu phẩy, chấm phẩy, xuống dòng)
            valuesToSubmit = bulkValue.split(/[\n,;]+/) 
                .map(v => v.trim())
                .filter(v => v.length > 0);
            
            if (valuesToSubmit.length === 0) return alert("Danh sách trống!");
        }

        const basePayload = {
            title: newTitle,
            policy_type: policyType,
            target_type: targetType,
            target_hwids: selectedHWIDs
        };
        
        // Gọi hàm từ cha để xử lý API (gửi mảng values)
        onAddPolicy({ ...basePayload, values: valuesToSubmit });

        // Reset Form sau khi gửi
        setNewTitle('');
        setSingleValue('');
        setBulkValue('');
        setSelectedHWIDs([]);
    };

    return (
        <div className="bg-[#1e293b] p-6 rounded-3xl border border-slate-800 shadow-xl sticky top-24">
            <div className="flex items-center gap-3 mb-6">
                <div className="p-3 bg-emerald-500/10 rounded-xl text-emerald-400"><Plus size={24}/></div>
                <div>
                    <h3 className="text-lg font-bold text-white">Thiết lập Quy tắc</h3>
                    <p className="text-xs text-slate-500">Tạo luật mới cho {currentConfig.label}</p>
                </div>
            </div>

            <form onSubmit={handleSubmit} className="space-y-5">
                {/* 1. TÊN GỢI NHỚ */}
                <div>
                    <label className="text-[10px] font-bold text-slate-500 uppercase">Tên gợi nhớ (Nhóm)</label>
                    <input 
                        type="text" 
                        placeholder={inputMode === 'BULK' ? "VD: Danh sách đen T3/2026" : "VD: Chặn Game LOL"} 
                        value={newTitle} 
                        onChange={e => setNewTitle(e.target.value)} 
                        required 
                        className="w-full mt-1 p-3 bg-slate-900 border border-slate-700 rounded-xl text-sm text-white focus:border-emerald-500 outline-none transition" 
                    />
                </div>

                {/* 2. GIÁ TRỊ KỸ THUẬT */}
                <div>
                    <div className="flex justify-between items-end mb-1">
                        <label className="text-[10px] font-bold text-slate-500 uppercase">Giá trị kỹ thuật</label>
                        
                        <div className="flex gap-2">
                             {/* Input file ẩn hỗ trợ nhiều định dạng */}
                             <input 
                                type="file" 
                                ref={fileInputRef} 
                                onChange={handleFileImport} 
                                accept=".txt,.csv,.log,.xlsx,.xls,.docx" 
                                className="hidden" 
                            />

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
                        <input 
                            type="text" 
                            placeholder={currentConfig.placeholder} 
                            value={singleValue} 
                            onChange={e => setSingleValue(e.target.value)} 
                            className="w-full p-3 bg-slate-900 border border-slate-700 rounded-xl text-sm font-mono text-white focus:border-emerald-500 outline-none transition" 
                        />
                    ) : (
                        <div className="relative group">
                            <textarea 
                                rows={6}
                                disabled={isReadingFile}
                                placeholder={isReadingFile ? "Đang đọc file..." : `Dán danh sách hoặc nạp file Excel/Word...\n${currentConfig.placeholder.replace('VD: ', '')}\n...`}
                                value={bulkValue}
                                onChange={e => setBulkValue(e.target.value)}
                                className="w-full p-3 bg-slate-900 border border-slate-700 rounded-xl text-xs font-mono text-white focus:border-blue-500 outline-none transition custom-scrollbar resize-none disabled:opacity-50"
                            />
                            
                            {/* Nút Nạp File Nổi */}
                            <button 
                                type="button"
                                onClick={() => fileInputRef.current.click()}
                                disabled={isReadingFile}
                                className="absolute bottom-3 right-3 flex items-center gap-1 bg-slate-800 hover:bg-slate-700 text-slate-300 px-2 py-1 rounded-lg border border-slate-600 text-[10px] font-bold transition shadow-lg disabled:opacity-50"
                                title="Nạp từ file Excel, Word, Text"
                            >
                                {isReadingFile ? <Loader2 size={10} className="animate-spin"/> : <Upload size={10} />} 
                                {isReadingFile ? " Đang xử lý..." : " Import (Excel/Doc)"}
                            </button>
                        </div>
                    )}
                    
                    <p className="text-[10px] text-slate-500 mt-1 italic flex items-center gap-1">
                        <Info size={10}/> 
                        {inputMode === 'SINGLE' ? currentConfig.hint : "Hỗ trợ nhập tay, hoặc nạp file Excel (.xlsx), Word (.docx), Text."}
                    </p>
                </div>

                {/* 3. LOẠI HÀNH ĐỘNG */}
                <div className="grid grid-cols-2 gap-3 p-1 bg-slate-900 rounded-xl border border-slate-700">
                    <button type="button" onClick={() => setPolicyType('BLACKLIST')} className={`py-2 rounded-lg text-xs font-bold transition flex items-center justify-center gap-2 ${policyType === 'BLACKLIST' ? 'bg-red-500 text-white shadow-lg' : 'text-slate-500 hover:text-white'}`}>
                        <ShieldCheck size={14}/> CẤM (Block)
                    </button>
                    <button type="button" onClick={() => setPolicyType('WHITELIST')} className={`py-2 rounded-lg text-xs font-bold transition flex items-center justify-center gap-2 ${policyType === 'WHITELIST' ? 'bg-emerald-500 text-white shadow-lg' : 'text-slate-500 hover:text-white'}`}>
                        <CheckCircle2 size={14}/> CHO PHÉP
                    </button>
                </div>

                {/* 4. PHẠM VI ÁP DỤNG */}
                <div className="space-y-3 pt-2 border-t border-slate-800">
                    <label className="text-[10px] font-bold text-slate-500 uppercase">Phạm vi áp dụng</label>
                    <div className="flex gap-3">
                        <div onClick={() => setTargetType('GLOBAL')} className={`flex-1 p-3 rounded-xl border cursor-pointer transition flex flex-col items-center gap-2 ${targetType === 'GLOBAL' ? 'bg-blue-500/10 border-blue-500 text-blue-400' : 'bg-slate-900 border-slate-700 text-slate-500 hover:border-slate-600'}`}>
                            <Globe size={20}/>
                            <span className="text-xs font-bold">Toàn hệ thống</span>
                        </div>
                        <div onClick={() => setTargetType('SPECIFIC')} className={`flex-1 p-3 rounded-xl border cursor-pointer transition flex flex-col items-center gap-2 ${targetType === 'SPECIFIC' ? 'bg-purple-500/10 border-purple-500 text-purple-400' : 'bg-slate-900 border-slate-700 text-slate-500 hover:border-slate-600'}`}>
                            <Laptop size={20}/>
                            <span className="text-xs font-bold">Máy cụ thể</span>
                        </div>
                    </div>
                </div>

                {/* 5. CHỌN MÁY (Nếu Specific) */}
                {targetType === 'SPECIFIC' && (
                    <div className="bg-slate-900/50 rounded-xl border border-slate-700 animate-in fade-in slide-in-from-top-2 overflow-hidden">
                        <div className="p-3 border-b border-slate-800 bg-slate-900">
                            <div className="relative">
                                <Search className="absolute left-2 top-1/2 -translate-y-1/2 text-slate-500" size={14}/>
                                <input type="text" placeholder="Tìm tên máy hoặc IP..." value={agentSearch} onChange={(e) => setAgentSearch(e.target.value)} className="w-full pl-8 pr-2 py-1.5 bg-slate-800 border border-slate-700 rounded-lg text-xs text-white outline-none focus:border-purple-500 transition" />
                            </div>
                        </div>
                        <div className="p-2 space-y-1 min-h-[150px]">
                            {displayedAgents.length === 0 ? (
                                <div className="text-center py-4 text-slate-500 text-xs italic">Không tìm thấy máy phù hợp</div>
                            ) : (
                                displayedAgents.map(agent => (
                                    <label key={agent.hwid} className={`flex items-center gap-3 p-2 rounded-lg cursor-pointer transition border border-transparent ${selectedHWIDs.includes(agent.hwid) ? 'bg-purple-500/20 border-purple-500/30' : 'hover:bg-slate-800'}`}>
                                        <input type="checkbox" checked={selectedHWIDs.includes(agent.hwid)} onChange={() => {
                                            if (selectedHWIDs.includes(agent.hwid)) setSelectedHWIDs(prev => prev.filter(id => id !== agent.hwid));
                                            else setSelectedHWIDs(prev => [...prev, agent.hwid]);
                                        }} className="accent-purple-500 w-4 h-4 rounded" />
                                        <div className="min-w-0">
                                            <p className={`text-xs font-bold truncate ${selectedHWIDs.includes(agent.hwid) ? 'text-white' : 'text-slate-300'}`}>{agent.hostname}</p>
                                            <p className="text-[9px] text-slate-500 font-mono truncate">{agent.ip_address}</p>
                                        </div>
                                    </label>
                                ))
                            )}
                        </div>
                        <Pagination currentPage={currentPage} totalPages={totalPages} onPageChange={setCurrentPage} />
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