import React, { useEffect, useState } from 'react';
import axios from '../../../api/axios';
import { CheckCircle, PlusCircle } from 'lucide-react';

const AssetsBaseline = ({ hwid, onRefresh }) => {
    const [items, setItems] = useState([]);
    const [selected, setSelected] = useState([]);
    const [loading, setLoading] = useState(false);

    const fetchBaseline = async () => {
        try {
            setLoading(true);
            const res = await axios.get(`/policies?status=BASELINE`);
            const all = res.data || [];
            const assetItems = all.filter(p => (p.asset_hwid || p.assetHWID || '') === hwid);
            setItems(assetItems);
        } catch (err) { console.error(err); }
        finally { setLoading(false); }
    };

    useEffect(() => { fetchBaseline(); }, [hwid]);

    const toggle = (id) => setSelected(s => s.includes(id) ? s.filter(x=>x!==id) : [...s, id]);

    const approveSelected = async () => {
        if (selected.length === 0) return alert('Chọn ít nhất một mục để phê duyệt');
        try {
            setLoading(true);
            await axios.post('/policies/bulk-approve', { ids: selected });
            setSelected([]);
            fetchBaseline();
            if (onRefresh) onRefresh();
        } catch (err) { alert('Lỗi khi phê duyệt'); console.error(err); }
        finally { setLoading(false); }
    };

    const approveAllForAsset = async () => {
        if (!window.confirm('Phê duyệt tất cả đề xuất Baseline cho máy này?')) return;
        try {
            setLoading(true);
            await axios.post('/policies/approve-by-asset', { asset_hwid: hwid });
            fetchBaseline();
            if (onRefresh) onRefresh();
        } catch (err) { alert('Lỗi khi phê duyệt toàn bộ'); console.error(err); }
        finally { setLoading(false); }
    };

    const approveOne = async (id) => {
        try {
            await axios.put(`/policies/${id}/approve`);
            fetchBaseline();
            if (onRefresh) onRefresh();
        } catch (err) { alert('Lỗi khi phê duyệt'); console.error(err); }
    };

    return (
        <div className="h-full flex flex-col">
            <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-black uppercase tracking-wider">Proposed Baseline</h3>
                <div className="flex items-center gap-2">
                    <button onClick={approveSelected} className="px-3 py-1 bg-emerald-600 rounded text-xs font-bold">Bulk Approve</button>
                    <button onClick={approveAllForAsset} className="px-3 py-1 bg-indigo-600 rounded text-xs font-bold">Approve All</button>
                </div>
            </div>

            <div className="overflow-y-auto bg-[#0B1220] border border-slate-800 rounded p-3 flex-1">
                {loading && <p className="text-sm text-slate-500">Loading...</p>}
                {!loading && items.length === 0 && <p className="text-sm text-slate-500">No proposed baseline items.</p>}

                <table className="w-full text-sm">
                    <thead>
                        <tr className="text-left text-xs text-slate-400">
                            <th></th>
                            <th>Category</th>
                            <th>Value</th>
                            <th>Title</th>
                            <th className="text-right">Action</th>
                        </tr>
                    </thead>
                    <tbody>
                        {items.map(it => (
                            <tr key={it.id} className="border-t border-slate-800">
                                <td className="py-2"><input type="checkbox" checked={selected.includes(it.id)} onChange={() => toggle(it.id)} /></td>
                                <td className="py-2 font-mono text-[12px]">{it.category}</td>
                                <td className="py-2 font-mono text-[12px] truncate max-w-[280px]">{it.value}</td>
                                <td className="py-2 text-[12px]">{it.title}</td>
                                <td className="py-2 text-right">
                                    <button onClick={() => approveOne(it.id)} className="inline-flex items-center gap-2 px-2 py-1 rounded bg-emerald-500 text-xs font-bold">
                                        <CheckCircle size={14}/> Approve
                                    </button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default AssetsBaseline;
