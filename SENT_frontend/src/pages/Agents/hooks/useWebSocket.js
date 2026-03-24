import { useEffect } from 'react';
import { useAuth } from '../../../context/AuthContext.js';

export const useWebSocket = (onMessageReceived) => {
    const { user } = useAuth();

    useEffect(() => {
        if (!user || !user.token) return;

        const wsUrl = process.env.REACT_APP_WS_URL ;
        const socket = new WebSocket(`${wsUrl}?token=${user.token}`);

        socket.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (onMessageReceived) onMessageReceived(data);
        };

        socket.onclose = () => console.log("WebSocket Disconnected");

        return () => socket.close();
    }, [user, onMessageReceived]);
};