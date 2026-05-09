import { useEffect, useRef } from 'react';
import { useSocket } from '../context/WebSocketContext';

/**
 * Hook tiện ích để một component đăng ký và hủy đăng ký lắng nghe sự kiện WebSocket.
 * @param {string|string[]} eventType - Tên sự kiện hoặc một mảng các tên sự kiện để lắng nghe.
 * @param {function} callback - Hàm sẽ được gọi khi có tin nhắn, nhận (data, message) làm tham số.
 * @param {number} [debounceMs=0] - Thời gian debounce (ms) chống spam render (Mặc định: 0).
 */
export const useSocketSubscription = (eventType, callback, debounceMs = 0) => {
    const { subscribe } = useSocket();
    
    // Sử dụng ref để luôn giữ instance mới nhất của callback mà không làm kích hoạt lại useEffect
    const callbackRef = useRef(callback);
    const timeoutRef = useRef(null);

    useEffect(() => {
        callbackRef.current = callback;
    }, [callback]);

    useEffect(() => {
        const eventTypes = Array.isArray(eventType) ? eventType : [eventType];
        
        const handler = (data, message) => {
            if (debounceMs <= 0) {
                callbackRef.current(data, message);
                return;
            }
            
            if (timeoutRef.current) clearTimeout(timeoutRef.current);
            timeoutRef.current = setTimeout(() => {
                callbackRef.current(data, message);
            }, debounceMs);
        };

        // Đăng ký tất cả các sự kiện và thu thập các hàm hủy đăng ký
        const unsubscribers = eventTypes.map(type => subscribe(type, handler));

        // Khi component unmount, dọn dẹp bộ nhớ
        return () => {
            unsubscribers.forEach(unsubscribe => unsubscribe());
            if (timeoutRef.current) clearTimeout(timeoutRef.current);
        };
    }, [eventType, debounceMs, subscribe]);
};