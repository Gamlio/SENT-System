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
            
            setLoading(false); 
        };
        checkAuth();
    }, []);

    const login = async (loginData) => {
        try {
            // loginData sẽ bao gồm { username, password }
            const res = await API.post('auth/login', loginData); 
            
            // 1. CẬP NHẬT: Thêm permissions và lấy đúng tên trường token
            const userData = { 
                username: res.data.username || loginData.username, 
                role: res.data.role, 
                org_id: res.data.org_id,
                company_code: res.data.company_code, 
                permissions: res.data.permissions || {}
            };
            
            // 2. CẬP NHẬT: Backend trả về 'token' chứ không phải 'access_token'
            localStorage.setItem('sent_token', res.data.token);
            localStorage.setItem('sent_user', JSON.stringify(userData));
            
            setUser(userData);
            return { success: true, data: res.data };
        } catch (error) {
            console.error("Login Error:", error);
            // Xử lý ném lỗi ra ngoài để form login bắt được
            throw error; 
        }
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