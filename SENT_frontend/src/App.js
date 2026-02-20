import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from './context/AuthContext';
import Navbar from './components/Navbar.jsx';

// Import Pages (Bao gồm cả trang cũ và mới)
import Login from './pages/Auth/Login.jsx';
import Register from './pages/Auth/Register.jsx';
import Dashboard from './pages/Dashboard/Dashboard.jsx';
import Organizations from './pages/Admin/System/Organizations.jsx'; 
import UserManagement from './pages/Admin/System/UserManagement.jsx';
import AgentList from './pages/Agents/Agents.jsx';
import AgentDetail from './pages/Agents/AgentDetail.jsx'; 

import PolicyCenter from './pages/Admin/policies/PolicyCenter.jsx';
import PolicyDocuments from './pages/Admin/policies/PolicyDocuments.jsx'; // Trang nạp tài liệu cho AI
import AIChatAssistant from './pages/AI/AIChatAssistant.jsx'; // Trang giao tiếp với AI

const ProtectedRoute = ({ children, requiredLevel }) => {
    const { user, loading } = useAuth();
    if (loading) return <div className="h-screen bg-[#0f172a]"></div>;
    
    // 1. Nếu chưa đăng nhập -> Đẩy về trang Login
    if (!user) return <Navigate to="/login" />;
    
    // 2. Kiểm tra phân quyền: Level 1 (User), Level 2 (Admin)
    // Nếu trang yêu cầu Level 2 mà User chỉ có Level 1 -> Đẩy về trang chủ
    if (requiredLevel && user.level < requiredLevel) return <Navigate to="/" />;

    return (
        <div className="min-h-screen bg-[#0f172a] flex flex-col">
            <Navbar /> 
            
            <div className="flex-1 overflow-auto p-8">
                <div className="max-w-7xl mx-auto"> 
                    {children}
                </div>
            </div>
        </div>
    );
};

function App() {
    return (
        <Router>
            <Routes>
                {/* ========================================== */}
                {/* 1. KHU VỰC PUBLIC (Không cần đăng nhập)    */}
                {/* ========================================== */}
                <Route path="/login" element={<Login />} />
                <Route path="/register" element={<Register />} />


                {/* ========================================== */}
                {/* 2. KHU VỰC USER & ADMIN (Level 1 & 2)      */}
                {/* ========================================== */}
                <Route path="/" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
                <Route path="/agents" element={<ProtectedRoute><AgentList /></ProtectedRoute>} />
                <Route path="/agents/:hwid" element={<ProtectedRoute><AgentDetail /></ProtectedRoute>} />
                
                {/* Nhân viên được quyền chat với AI để hỏi đáp nội quy */}
                <Route path="/chat-ai" element={<ProtectedRoute><AIChatAssistant /></ProtectedRoute>} />


                {/* ========================================== */}
                {/* 3. KHU VỰC CHỈ DÀNH CHO ADMIN (Level 2)    */}
                {/* ========================================== */}
                
                {/* Quản lý các công ty SME (Nếu bạn là Admin tổng) */}
                <Route path="/admin/orgs" element={<ProtectedRoute requiredLevel={2}><Organizations /></ProtectedRoute>} />
                {/* Cấp tài khoản cho Nhân viên */}
                <Route path="/admin/users" element={<ProtectedRoute requiredLevel={2}><UserManagement /></ProtectedRoute>} />
                {/* Nạp tài liệu PDF/Word vào cho AI Agent đọc */}
                <Route path="/admin/docs" element={<ProtectedRoute requiredLevel={2}><PolicyDocuments /></ProtectedRoute>} />

                <Route path="/admin/policy-center" element={<ProtectedRoute requiredLevel={2}><PolicyCenter /></ProtectedRoute>} />
                
                {/* Quản lý các quy định cấm USB, Phần mềm */}

               

                {/* Route dự phòng */}
                <Route path="*" element={<Navigate to="/" />} />
            </Routes>
        </Router>
    );
}

export default App;