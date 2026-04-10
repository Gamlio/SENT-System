// components/IncidentActionBox.jsx
const IncidentActionBox = ({ incident, onSuccess }) => {
    const [note, setNote] = useState("");
    const [evidence, setEvidence] = useState("");

    const handleResolve = async () => {
        try {
            // Gọi API /close với dữ liệu thực tế
            await axiosInstance.post('/incidents/close', {
                incident_id: incident.id,
                note: note,
                evidence_data: evidence // Bắt buộc theo logic Backend
            });
            onSuccess();
            alert("Sự cố đã được đóng và niêm phong AuditTrail.");
        } catch (err) {
            alert(err.response?.data?.error || "Lỗi khi đóng sự cố");
        }
    };

    return (
        <div className="bg-[#0A101D] border border-slate-800 rounded-lg p-4">
            <h4 className="text-[10px] font-black text-slate-500 uppercase mb-4">Submit Resolution Proof</h4>
            <textarea 
                className="w-full bg-slate-900 border border-slate-800 rounded p-3 text-xs mb-3"
                placeholder="Ghi chú điều tra..."
                value={note}
                onChange={(e) => setNote(e.target.value)}
            />
            <textarea 
                className="w-full bg-black border border-slate-800 rounded p-2 text-[10px] font-mono mb-4 text-emerald-500"
                placeholder='Paste Baseline Scan JSON here...'
                value={evidence}
                onChange={(e) => setEvidence(e.target.value)}
            />
            <button 
                onClick={handleResolve}
                className="w-full bg-emerald-600 hover:bg-emerald-700 text-white text-[10px] font-black py-2 rounded uppercase transition"
            >
                Confirm & Seal Audit
            </button>
        </div>
    );
};
export default IncidentActionBox;