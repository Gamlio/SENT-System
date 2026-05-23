import { useState, useEffect, useRef, useCallback } from 'react';
import axios from '../../../api/axios'; 

export const useChat = () => {
    const [sessions, setSessions] = useState([]);
    const [currentSession, setCurrentSession] = useState(null);
    const [messages, setMessages] = useState([]);
    const [isLoading, setIsLoading] = useState(false);

    const isFetched = useRef(false);
    const messagesEndRef = useRef(null);


    const fetchMessages = useCallback(async (sessionId) => {
        if (!sessionId) return;
        
        try {
            // setIsLoading(true); // Có thể bật loading nếu muốn
            const res = await axios.get(`/ai/chat/${sessionId}`);
            
            // Map dữ liệu từ DB (snake_case) sang Frontend (camelCase)
            const history = (res.data || []).map(log => ({
                sender: log.role,      // "user" hoặc "ai"
                text: log.content
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

        setMessages(prev => [...prev, { sender: 'ai', text: "" }]);

         try {
             const token = localStorage.getItem('sent_token');
             const response = await fetch(`${import.meta.env.VITE_API_URL || '/api/v1'}/ai/chat/${activeSession.id}`, {
                 method: 'POST',
                 headers: {
                     'Content-Type': 'application/json',
                     ...(token ? { 'Authorization': `Bearer ${token}` } : {})
                 },
                 body: JSON.stringify({ message: text })
             });
 
             if (!response.ok) throw new Error(`Lỗi kết nối: ${response.status} ${response.statusText}`);
 
             const reader = response.body.getReader();
             const decoder = new TextDecoder("utf-8");
             let accumulatedRawText = "";
 
             while (true) {
                 const { value, done } = await reader.read();
                 if (done) break;
 
                 const chunk = decoder.decode(value, { stream: true });
                 const lines = chunk.split("\n");
 
                 for (const line of lines) {
                    if (line.trim().startsWith("data:")) {
                        try {
                            const cleanedLine = line.replace("data:", "").trim();
                            if (!cleanedLine) continue;
                            const parsedData = JSON.parse(cleanedLine);

                            if (parsedData.text) {
                                accumulatedRawText += parsedData.text;                                
                                
                                // Thinking logic removed. The entire response is the answer.
                                setMessages(prev => {
                                    const updated = [...prev];
                                    const targetIndex = updated.length - 1;
                                    if (updated[targetIndex]) {
                                        updated[targetIndex].text = accumulatedRawText;
                                    }
                                    return updated;
                                });
                            }
                        } catch (jsonErr) {
                            // Bỏ qua lỗi parse JSON
                        }
                    }
                 }
             }
 
         } catch (error) {
             console.error("Chat Error:", error);
             setMessages(prev => {
                 const updated = [...prev];
                 const lastMsgIndex = updated.length - 1;
                 if (lastMsgIndex >= 0 && updated[lastMsgIndex].sender === 'ai') {
                     updated[lastMsgIndex] = {
                         sender: 'ai',
                         text: `⚠️ Lỗi: ${error.message || 'Không thể kết nối luồng dữ liệu AI.'}`
                     };
                 }
                 return updated;
             });
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