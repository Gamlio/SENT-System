import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from './context/AuthContext';
import Sidebar from './components/Sidebar.jsx';
import Navbar from './components/Navbar.jsx';

import Login from './pages/Auth/Login.jsx';
import Register from './pages/Auth/Register.jsx';
import Dashboard from './pages/Dashboard/Dashboard.jsx';

import AgentList from './pages/Agents/Agents.jsx';
import AgentDetail from './pages/Agents/AgentDetail.jsx'; 
import IncidentList from './pages/IncidentReport/IncidentManager.jsx';
import IncidentDetailPanel from './pages/IncidentReport/components/IncidentDetailPanel.jsx';
import AIChatPage from './pages/AIChat/AIChatPage.jsx';

import PolicyCenter from './pages/PolicyCenter/PolicyCenter.jsx'; 
import Documents from './pages/KnowledgeBase/Documents.jsx';     
import UserManagement from './pages/Admin/UserManagement.jsx';   

const ProtectedRoute = ({ children, requiredPermission }) => {
    const { user, loading } = useAuth();
    
    if (loading) return <div className="h-screen bg-[#0f172a]"></div>;
    if (!user) return <Navigate to="/login" />;
    
    // Kiểm tra phân quyền chi tiết
    if (requiredPermission && user.permissions && !user.permissions[requiredPermission]) {
        return <Navigate to="/" />; 
    }

    // KHUNG GIAO DIỆN CHÍNH NẰM Ở ĐÂY (Tuyệt vời!)
    return (
        <div className="flex h-screen overflow-hidden bg-[#0f172a]">
            {/* Sidebar bên trái */}
            <Sidebar />
            
            {/* Khu vực nội dung chính */}
            <div className="flex-1 flex flex-col overflow-hidden">
                {/* Header */}
                <Navbar />
                
                {/* Nội dung trang */}
                <main className="flex-1 overflow-x-hidden overflow-y-auto p-8 relative">
                    <div className="max-w-7xl mx-auto h-full"> 
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
                {/* Public Routes */}
                <Route path="/login" element={<Login />} />
                <Route path="/register" element={<Register />} />

                {/* Core Routes */}
                <Route path="/" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
                <Route path="/chat-ai" element={<ProtectedRoute><AIChatPage /></ProtectedRoute>} />
                
                <Route path="/agents" element={<ProtectedRoute requiredPermission="view_agents"><AgentList /></ProtectedRoute>} />
                <Route path="/agents/:hwid" element={<ProtectedRoute requiredPermission="view_agents"><AgentDetail /></ProtectedRoute>} />
                
                <Route path="/incidents" element={<ProtectedRoute requiredPermission="manage_incidents"><IncidentList /></ProtectedRoute>} />
                <Route path="/incidents/:id" element={<ProtectedRoute requiredPermission="manage_incidents"><IncidentDetailPanel /></ProtectedRoute>} />

                {/* Compliance & Knowledge Base Routes */}
                <Route path="/admin/policy-center" element={<ProtectedRoute requiredPermission="manage_policies"><PolicyCenter /></ProtectedRoute>} />
                <Route path="/admin/docs" element={<ProtectedRoute requiredPermission="view_docs"><Documents /></ProtectedRoute>} />
                
                {/* System Admin Routes */}
                <Route path="/admin/users" element={<ProtectedRoute requiredPermission="manage_users"><UserManagement /></ProtectedRoute>} />
            </Routes>
        </Router>
    );
}

export default App;