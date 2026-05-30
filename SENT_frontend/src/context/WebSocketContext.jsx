import React, { createContext, useContext, useEffect, useRef, useState, useCallback, useMemo } from 'react';

const WebSocketContext = createContext(null);

/**
 * Hook tiện ích để truy cập vào context của WebSocket.
 */
export const useSocket = () => {
    const context = useContext(WebSocketContext);
    if (!context) {
        throw new Error('useSocket must be used within a WebSocketProvider');
    }
    return context;
};


export const WebSocketProvider = ({ children }) => {
    const [isConnected, setIsConnected] = useState(false);
    
    // 1. Dùng useMemo để ổn định URL, tránh reconnect liên tục
    const socketUrl = useMemo(() => {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const apiUrl = import.meta.env.VITE_API_URL || '/api/v1';
        const apiPath = apiUrl.startsWith('http') ? new URL(apiUrl).pathname : apiUrl;
        return `${protocol}//${window.location.host}${apiPath.replace(/\/$/, '')}/ws`;
    }, []);

    const listeners = useRef({}); 
    const socket = useRef(null);

    useEffect(() => {
        const connect = () => {
            // 1. Lấy token từ một nguồn đáng tin cậy (thường là 'accessToken' theo AuthContext)
            const token = localStorage.getItem('sent_token');
                          
            if (!token) {
                console.warn('⚠️ [WebSocket] Không tìm thấy Token, tạm hoãn kết nối...');
                setTimeout(connect, 5000); // Thử lại sau 5s nếu user chưa đăng nhập
                return;
            }
            
            // 2. Tạo URL với token
            const url = new URL(socketUrl);
            url.searchParams.append('token', token);

            // Cách tiếp cận an toàn hơn (yêu cầu backend hỗ trợ):
            // socket.current = new WebSocket(socketUrl, ['Authorization', token]);
            // Backend sẽ cần đọc token từ header `Sec-WebSocket-Protocol`.
            
            socket.current = new WebSocket(url.toString());

            socket.current.onopen = () => {
                console.log('✅ WebSocket Connected');
                setIsConnected(true);
            };

            socket.current.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    const { type, data } = message;

                    // Gửi tin nhắn đến tất cả các listener đã đăng ký cho event type này
                    if (listeners.current[type]) {
                        listeners.current[type].forEach(callback => callback(data, message));
                    }
                } catch (error) {
                    console.error('Lỗi xử lý tin nhắn WebSocket:', error);
                }
            };

            socket.current.onerror = (error) => {
                console.error('Lỗi WebSocket:', error);
            };

            socket.current.onclose = () => {
                console.log('❌ WebSocket Disconnected. Tự động kết nối lại sau 5 giây...');
                setIsConnected(false);
                setTimeout(connect, 5000); // Tự động kết nối lại
            };
        };

        connect();

        // Hàm dọn dẹp khi component unmount
        return () => {
            if (socket.current) {
                socket.current.onclose = null; // Ngăn việc tự động kết nối lại khi unmount
                socket.current.close();
            }
        };
    }, [socketUrl]); // Chỉ chạy lại nếu URL thay đổi

    // Hàm cho phép các component đăng ký lắng nghe một loại sự kiện
    const subscribe = useCallback((eventType, callback) => {
        if (!listeners.current[eventType]) {
            listeners.current[eventType] = [];
        }
        listeners.current[eventType].push(callback);

        // Trả về một hàm để hủy đăng ký
        return () => {
            if (listeners.current[eventType]) {
                listeners.current[eventType] = listeners.current[eventType].filter(cb => cb !== callback);
            }
        };
    }, []);

    const value = { isConnected, subscribe };

    return (
        <WebSocketContext.Provider value={value}>
            {children}
        </WebSocketContext.Provider>
    );
};