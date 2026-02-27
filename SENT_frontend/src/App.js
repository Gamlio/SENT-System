import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from './context/AuthContext';
import Sidebar from './components/Sidebar.jsx';
import Navbar from './components/Navbar.jsx';

import Login from './pages/Auth/Login.jsx';
import Register from './pages/Auth/Register.jsx';
import Dashboard from './pages/Dashboard/Dashboard.jsx';
import UserManagement from './pages/Admin/System/UserManagement.jsx';
import AgentList from './pages/Agents/Agents.jsx';
import AgentDetail from './pages/Agents/AgentDetail.jsx'; 
import PolicyCenter from './pages/Admin/policies/PolicyCenter.jsx';
import PolicyDocuments from './pages/Admin/policies/PolicyDocuments.jsx';
import IncidentReport from './pages/IncidentReport/IncidentReport.jsx';
import IncidentList from './pages/IncidentReport/IncidentList.jsx';
import AIChatAssistant from './pages/AI/AIChatAssistant.jsx';

const ProtectedRoute = ({ children, requiredPermission }) => {
    const { user, loading } = useAuth();
    if (loading) return <div className="h-screen bg-[#0f172a]"></div>;
    
    if (!user) return <Navigate to="/login" />;
    
    // Kiểm tra phân quyền chi tiết (Nếu có yêu cầu)
    if (requiredPermission && user.permissions && !user.permissions[requiredPermission]) {
        return <Navigate to="/" />; // Không có quyền thì đẩy về trang chủ
    }

    return (
        <div className="flex h-screen overflow-hidden bg-[#0f172a]">
            {/* Sidebar bên trái */}
            <Sidebar />
            
            {/* Khu vực nội dung chính */}
            <div className="flex-1 flex flex-col overflow-hidden">
                {/* Header (Navbar rút gọn) */}
                <Navbar />
                
                {/* Nội dung trang */}
                <main className="flex-1 overflow-x-hidden overflow-y-auto p-8">
                    <div className="max-w-7xl mx-auto"> 
                        {children}
                    </div>
                </main>
            </div>
        </div>
    );
};

function App() {
    return (
        <Router>
            <Routes>
                <Route path="/login" element={<Login />} />
                <Route path="/register" element={<Register />} />

                {/* Dashboard ai cũng xem được */}
                <Route path="/" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
                
                {/* AI Chat (Chung) */}
                <Route path="/chat-ai" element={<ProtectedRoute><AIChatAssistant /></ProtectedRoute>} />

                {/* Các Route có check quyền riêng biệt */}
                <Route path="/agents" element={<ProtectedRoute requiredPermission="view_agents"><AgentList /></ProtectedRoute>} />
                <Route path="/agents/:hwid" element={<ProtectedRoute requiredPermission="view_agents"><AgentDetail /></ProtectedRoute>} />
                
                <Route path="/admin/policy-center" element={<ProtectedRoute requiredPermission="manage_policies"><PolicyCenter /></ProtectedRoute>} />
                <Route path="/admin/docs" element={<ProtectedRoute requiredPermission="view_docs"><PolicyDocuments /></ProtectedRoute>} />
                <Route path="/admin/users" element={<ProtectedRoute requiredPermission="manage_users"><UserManagement /></ProtectedRoute>} />
                <Route path="/incidents" element={<ProtectedRoute requiredPermission="manage_incidents"><IncidentList /></ProtectedRoute>} />
                <Route path="/incidents/:id" element={<ProtectedRoute requiredPermission="manage_incidents"><IncidentReport /></ProtectedRoute>} />
            </Routes>
        </Router>
    );
}

export default App;