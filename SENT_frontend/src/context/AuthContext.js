import React, { createContext, useState, useContext, useEffect } from 'react';
import API from '../api/axios';

const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
    const checkAuth = () => {
        const token = localStorage.getItem('sent_token');
        const savedUser = localStorage.getItem('sent_user');
        if (token && savedUser) {
            try {
                setUser(JSON.parse(savedUser));
            } catch (e) {
                console.error("Lỗi parse dữ liệu User:", e);
                localStorage.clear();
            }
        }
        // Luôn luôn phải có dòng này ở cuối để tắt màn hình chờ
        setLoading(false); 
    };
    checkAuth();
}, []);

    const login = async (loginData) => {
    // loginData sẽ bao gồm { username, password, captcha, otp }
    const res = await API.post('/auth/login', loginData); 
    
    const userData = { 
        username: loginData.username, 
        level: res.data.level, 
        org_id: res.data.org_id 
    };
    
    localStorage.setItem('sent_token', res.data.access_token);
    localStorage.setItem('sent_user', JSON.stringify(userData));
    setUser(userData);
    return res.data;
};

    const logout = () => {
        localStorage.clear();
        setUser(null);
    };

    return (
        <AuthContext.Provider value={{ user, login, logout, loading }}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => useContext(AuthContext);