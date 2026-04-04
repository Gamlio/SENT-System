import { useState, useEffect, useRef, useCallback } from 'react';
import axios from '../../../api/axios'; 

export const useChat = () => {
    const [sessions, setSessions] = useState([]);
    const [currentSession, setCurrentSession] = useState(null);
    const [messages, setMessages] = useState([]);
    const [isLoading, setIsLoading] = useState(false);
    
    // useRef để chặn spam API
    const isFetched = useRef(false);
    const messagesEndRef = useRef(null);

    // --- 1. HÀM TẢI TIN NHẮN TỪ SERVER ---
    // (Đưa lên đầu để useEffect có thể gọi được)
    const fetchMessages = useCallback(async (sessionId) => {
        if (!sessionId) return;
        
        try {
            // setIsLoading(true); // Có thể bật loading nếu muốn
            const res = await axios.get(`/ai/chat/${sessionId}`);
            
            // Map dữ liệu từ DB (snake_case) sang Frontend (camelCase)
            const history = (res.data || []).map(log => ({
                sender: log.role,      // "user" hoặc "ai"
                text: log.content,
                thought: log.thought   // Suy luận (nếu có)
            }));
            
            setMessages(history);
        } catch (err) {
            console.error("Lỗi tải lịch sử tin nhắn:", err);
            setMessages([]);
        } finally {
            // setIsLoading(false);
        }
    }, []);

    // --- 2. LOGIC TẠO PHIÊN MỚI ---
    const createNewSession = async () => {
        try {
            const res = await axios.post('/ai/sessions', {});
            const newSession = res.data;
            
            setSessions(prev => [newSession, ...prev]);
            setCurrentSession(newSession);
            setMessages([]);
            return newSession; 
        } catch (err) { 
            console.error("Lỗi tạo phiên:", err);
            return null;
        }
    };

    // --- 3. LOGIC TẢI DANH SÁCH PHIÊN ---
    const fetchSessions = useCallback(async () => {
        try {
            const res = await axios.get('/ai/sessions');
            const data = res.data || [];
            setSessions(data);
            
            if (data.length > 0) {
                // Chọn phiên mới nhất mặc định
                setCurrentSession(prev => prev || data[0]);
            } else {
                await createNewSession();
            }
        } catch (err) { 
            console.error("Lỗi tải Sessions:", err); 
        }
    }, []);

    // Chạy 1 lần khi vào trang
    useEffect(() => {
        if (!isFetched.current) {
            fetchSessions();
            isFetched.current = true;
        }
    }, [fetchSessions]);

    // --- 4. KHI CHỌN PHIÊN KHÁC -> TẢI TIN NHẮN CỦA PHIÊN ĐÓ ---
    useEffect(() => {
        if (currentSession?.id) {
            fetchMessages(currentSession.id);
        } else {
            setMessages([]);
        }
    }, [currentSession?.id, fetchMessages]); // <--- Dependency quan trọng

    const selectSession = (sessionId) => {
        const sess = sessions.find(s => s.id === sessionId);
        if (sess) setCurrentSession(sess);
    };

    // --- 5. GỬI TIN NHẮN ---
    const sendMessage = async (text) => {
        if (!text.trim()) return;

        let activeSession = currentSession;

        if (!activeSession) {
            activeSession = await createNewSession();
            if (!activeSession) return;
        }

        const userMsg = { sender: 'user', text };
        setMessages(prev => [...prev, userMsg]);
        setIsLoading(true);

        try {
            const res = await axios.post(`/ai/chat/${activeSession.id}`, { message: text });
            
            const aiMsg = { 
                sender: 'ai', 
                text: res.data.response,
                thought: res.data.thought
            };
            setMessages(prev => [...prev, aiMsg]);
        } catch (error) {
            console.error("Chat Error:", error);
            setMessages(prev => [...prev, { sender: 'ai', text: "⚠️ Lỗi kết nối AI." }]);
        } finally {
            setIsLoading(false);
        }
    };

    const renameSession = async (sessionId, newTitle) => {
        try {
            await axios.put(`/ai/sessions/${sessionId}`, { title: newTitle });
            setSessions(prev => prev.map(s => 
                s.id === sessionId ? { ...s, title: newTitle } : s
            ));
            if (currentSession?.id === sessionId) {
                setCurrentSession(prev => ({ ...prev, title: newTitle }));
            }
        } catch (err) { console.error("Lỗi đổi tên:", err); }
    };

    const deleteSession = async (sessionId) => {
        if (!window.confirm("Bạn chắc chắn muốn xóa cuộc trò chuyện này?")) return;
        try {
            await axios.delete(`/ai/sessions/${sessionId}`);
            const remaining = sessions.filter(s => s.id !== sessionId);
            setSessions(remaining);
            if (currentSession?.id === sessionId) {
                if (remaining.length > 0) setCurrentSession(remaining[0]);
                else createNewSession();
            }
        } catch (err) { console.error("Lỗi xóa phiên:", err); }
    };

    return { 
        messages, isLoading, sendMessage, messagesEndRef,
        sessions, currentSession, createNewSession, selectSession, 
        renameSession, deleteSession 
    };
};