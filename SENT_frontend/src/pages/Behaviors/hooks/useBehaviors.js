import { useState, useCallback } from 'react';
import axios from '../../../api/axios';

export const useBehaviors = () => {
  const [behaviors, setBehaviors] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const fetchBehaviors = useCallback(async (page = 1, limit = 20) => {
    setLoading(true);
    setError(null);
    try {
      const response = await axios.get(`/behaviors?page=${page}&limit=${limit}`);
      // Xử lý dữ liệu trả về từ API
      setBehaviors(response.data || []);
    } catch (err) {
      setError(err.response?.data?.error || 'Lỗi khi tải danh sách hành vi');
    } finally {
      setLoading(false);
    }
  }, []);

  const escalateToIncident = async (alertId) => {
    try {
      const response = await axios.post('/behaviors/escalate', { alert_id: alertId });
      
      // Cập nhật lại UI: Xóa hành vi đã được nâng cấp khỏi danh sách (hoặc đánh dấu là đã Resolved)
      setBehaviors((prev) => 
        prev.map((b) => b.id === alertId ? { ...b, is_resolved: true, incident_id: response.data.incident.id } : b)
      );

      return { success: true, incident: response.data.incident };
    } catch (err) {
      const errMsg = err.response?.data?.error || 'Lỗi khi nâng cấp sự cố';
      setError(errMsg);
      return { success: false, error: errMsg };
    }
  };

  const getSeverityColor = (severity) => {
    switch (severity) {
      case 'Critical': return 'text-red-500 bg-red-500/10 border-red-500/20';
      case 'High': return 'text-orange-500 bg-orange-500/10 border-orange-500/20';
      case 'Medium': return 'text-yellow-500 bg-yellow-500/10 border-yellow-500/20';
      default: return 'text-emerald-500 bg-emerald-500/10 border-emerald-500/20';
    }
  };

  return {
    behaviors, loading, error, fetchBehaviors, escalateToIncident, getSeverityColor
  };
};