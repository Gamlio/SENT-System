import { useEffect } from 'react';
import { useSocket } from '../context/WebSocketContext';

/**
 * Hook tiện ích để một component đăng ký và hủy đăng ký lắng nghe sự kiện WebSocket.
 * @param {string|string[]} eventType - Tên sự kiện hoặc một mảng các tên sự kiện để lắng nghe.
 * @param {function} callback - Hàm sẽ được gọi khi có tin nhắn, nhận (data, message) làm tham số.
 */
export const useSocketSubscription = (eventType, callback) => {
    const { subscribe } = useSocket();

    useEffect(() => {
        const eventTypes = Array.isArray(eventType) ? eventType : [eventType];
        
        // Đăng ký tất cả các sự kiện và thu thập các hàm hủy đăng ký
        const unsubscribers = eventTypes.map(type => subscribe(type, callback));

        // Khi component unmount, gọi tất cả các hàm hủy đăng ký
        return () => unsubscribers.forEach(unsubscribe => unsubscribe());

    }, [eventType, callback, subscribe]);
};