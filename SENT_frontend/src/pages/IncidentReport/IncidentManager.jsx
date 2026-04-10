import React, { useState, useEffect, useCallback } from 'react';
import axiosInstance from '../../api/axios';
import { ShieldAlert, Radar, Flame, Lock, Usb, Activity } from 'lucide-react';
import IncidentCard from "./components/IncidentCard";
import IncidentDetail from "./IncidentDetail";

const IncidentManager = () => {
    const [incidents, setIncidents] = useState([]);
    const [loading, setLoading] = useState(true);
    const [selectedIncidentId, setSelectedIncidentId] = useState(null); 

    const fetchIncidents = useCallback(async () => {
        setLoading(true);
        try {
            const res = await axiosInstance.get('/incidents');
            setIncidents(Array.isArray(res.data) ? res.data : []);
        } catch (err) {
            console.error("Lỗi tải danh sách sự cố:", err);
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => { fetchIncidents(); }, [fetchIncidents]);

    if (selectedIncidentId) {
        return (
            <IncidentDetail 
                incidentId={selectedIncidentId} 
                onBack={() => {
                    setSelectedIncidentId(null);
                    fetchIncidents();
                }} 
            />
        );
    }

    // ĐỒNG BỘ: Màn hình Loading giống hệt Dashboard
    if (loading) return (
        <div className="h-[calc(100vh-60px)] bg-[#050B14] flex flex-col items-center justify-center gap-4">
            <Radar size={40} className="text-indigo-500 animate-spin-slow"/>
            <p className="text-indigo-500 font-mono text-sm tracking-widest animate-pulse uppercase">Syncing Incident Data...</p>
        </div>
    );

    return (
        // ĐỒNG BỘ: Wrapper bọc toàn màn hình giống Dashboard
        <div className="h-[calc(100vh-60px)] bg-[#050B14] text-slate-200 p-5 font-sans overflow-y-auto custom-scrollbar">
            
            {/* ĐỒNG BỘ: Header phong cách Command Center */}
            <div className="mb-6 flex items-end justify-between">
                <div>
                    <h1 className="text-xl font-black text-white flex items-center gap-2 uppercase tracking-tight">
                        <ShieldAlert className="text-indigo-500"/> INCIDENT WORKBENCH
                    </h1>
                    <p className="text-[10px] text-slate-500 mt-1 uppercase tracking-widest font-bold">Real-time Threat Response & Audit</p>
                </div>
                <div className="flex items-center gap-2 bg-[#0A101D] border border-slate-800 px-3 py-1.5 rounded-lg">
                    <div className="w-2 h-2 rounded-full bg-red-500 animate-pulse"></div>
                    <span className="text-[10px] font-mono text-red-400 font-bold tracking-widest">LIVE CASES: {incidents.length}</span>
                </div>
            </div>
            
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                {incidents.length === 0 ? (
                    <div className="col-span-full p-10 text-center text-slate-500 border border-slate-800 border-dashed rounded-lg bg-[#0A101D]">
                        NO ACTIVE INCIDENTS
                    </div>
                ) : (
                    incidents.map(inc => (
                        <IncidentCard 
                            key={inc.ID || inc.id} 
                            incident={inc} 
                            onClick={() => setSelectedIncidentId(inc.ID || inc.id)} 
                        />
                    ))
                )}
            </div>
        </div>
    );
};

export default IncidentManager;