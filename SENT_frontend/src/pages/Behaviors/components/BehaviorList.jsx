import React, { useEffect, useState } from 'react';
import { useBehaviors } from '../hooks/useBehaviors';
import { AlertTriangle, ShieldAlert, ArrowUpRight } from 'lucide-react';

const BehaviorList = () => {
  const { behaviors, loading, error, fetchBehaviors, escalateToIncident, getSeverityColor } = useBehaviors();
  const [page, setPage] = useState(1);

  useEffect(() => {
    fetchBehaviors(page, 20);
  }, [page, fetchBehaviors]);

  const handleEscalate = async (alertId) => {
    if (!window.confirm('Bạn có chắc chắn muốn lập hồ sơ sự cố cho hành vi này?')) return;
    
    const res = await escalateToIncident(alertId);
    if (res.success) {
      alert(`Đã lập sự cố thành công! ID: ${res.incident.id}`);
    } else {
      alert(`Lỗi: ${res.error}`);
    }
  };

  return (
    <div className="p-6 bg-[#050B14] min-h-screen text-slate-300 font-sans">
      <div className="flex items-center gap-3 mb-6 border-b border-slate-800 pb-4">
        <ShieldAlert className="w-8 h-8 text-emerald-500" />
        <h1 className="text-2xl font-black uppercase tracking-widest text-emerald-400">Giám Sát Hành Vi</h1>
      </div>

      {error && <div className="mb-4 p-3 bg-red-500/10 border border-red-500/30 text-red-500 rounded">{error}</div>}

      <div className="bg-slate-900 border border-slate-800 rounded-lg overflow-hidden">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-slate-950 uppercase text-xs tracking-wider text-slate-500 border-b border-slate-800">
              <th className="p-4">Thời gian</th>
              <th className="p-4">Mức độ</th>
              <th className="p-4">Máy trạm (HWID)</th>
              <th className="p-4">Hành vi vi phạm</th>
              <th className="p-4">Trạng thái</th>
              <th className="p-4 text-right">Thao tác</th>
            </tr>
          </thead>
          <tbody>
            {loading && behaviors.length === 0 ? (
              <tr><td colSpan="6" className="p-8 text-center text-slate-500">Đang quét dữ liệu...</td></tr>
            ) : (
              behaviors.map((b) => (
                <tr key={b.id} className="border-b border-slate-800/50 hover:bg-slate-800/20 transition-colors">
                  <td className="p-4 text-sm text-slate-400">{new Date(b.created_at).toLocaleString()}</td>
                  <td className="p-4">
                    <span className={`px-2 py-1 rounded text-xs font-bold border ${getSeverityColor(b.severity)}`}>
                      {b.priority} - {b.severity}
                    </span>
                  </td>
                  <td className="p-4 font-mono text-sm text-indigo-400">{b.asset_hwid}</td>
                  <td className="p-4">
                    <div className="font-bold text-slate-200">{b.title}</div>
                    <div className="text-xs text-slate-500 mt-1 truncate max-w-md">{b.description}</div>
                  </td>
                  <td className="p-4 text-sm">
                    {b.is_resolved ? (
                      <span className="text-slate-500">Đã xử lý (Incident #{b.incident_id})</span>
                    ) : (
                      <span className="text-amber-500 animate-pulse">Cần chú ý</span>
                    )}
                  </td>
                  <td className="p-4 text-right">
                    {!b.is_resolved && (
                      <button onClick={() => handleEscalate(b.id)} className="flex items-center gap-1 ml-auto px-3 py-1.5 bg-red-600/20 text-red-400 border border-red-600/30 rounded hover:bg-red-600/40 transition-colors text-sm font-bold uppercase tracking-wider">
                        <AlertTriangle className="w-4 h-4" /> Lập sự cố
                      </button>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
export default BehaviorList;