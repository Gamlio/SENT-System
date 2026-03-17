import React, { useState } from 'react'; // Bắt buộc thêm useState ở đây
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
import AIChatPage from './pages/AIChat/AIChatPage.jsx';
import ApprovalCenter from './pages/approvals/ApprovalCenter.jsx';
import PolicyCenter from './pages/policies/PolicyCenter.jsx'; 
import IncidentDetail from './pages/IncidentReport/IncidentDetail.jsx';
import Documents from './pages/KnowledgeBase/Documents.jsx';     
import UserManagement from './pages/User/UserManagement.jsx';   

// IMPORT DRAWER MỚI TẠO
import GlobalCopilotDrawer from './components/GlobalCopilotDrawer.jsx'; 

const ProtectedRoute = ({ children, requiredPermission }) => {
    const { user, loading } = useAuth();
    
    // STATE ĐIỀU KHIỂN AI DRAWER
    const [isCopilotOpen, setIsCopilotOpen] = useState(false);
    
    if (loading) return <div className="h-screen bg-[#0f172a]"></div>;
    if (!user) return <Navigate to="/login" />;
    
    if (requiredPermission && user.permissions && !user.permissions[requiredPermission]) {
        return <Navigate to="/" />; 
    }

    return (
        <div className="flex h-screen overflow-hidden bg-[#0f172a] relative">
            {/* Sidebar bên trái */}
            <Sidebar />
            
            {/* Khu vực nội dung chính */}
            <div className="flex-1 flex flex-col overflow-hidden">
                {/* Header truyền hàm mở Drawer vào Navbar */}
                <Navbar onOpenCopilot={() => setIsCopilotOpen(true)} />
                
                {/* Nội dung trang */}
                <main className="flex-1 overflow-x-hidden overflow-y-auto p-8 relative">
                    <div className="max-w-7xl mx-auto h-full"> 
                        {children}
                    </div>
                </main>
            </div>
            
            {/* NHÚNG DRAWER VÀO LAYOUT TOÀN CỤC */}
            <GlobalCopilotDrawer 
                isOpen={isCopilotOpen} 
                onClose={() => setIsCopilotOpen(false)} 
            />
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
                <Route path="/incidents/:id" element={<IncidentDetail />} />
                <Route path="/approvals" element={<ProtectedRoute ><ApprovalCenter /></ProtectedRoute>} />

                {/* Compliance & Knowledge Base Routes */}
                <Route path="/admin/policy-center" element={<ProtectedRoute requiredPermission="manage_policies"><PolicyCenter /></ProtectedRoute>} />
                <Route path="/admin/docs" element={<ProtectedRoute requiredPermission="view_docs"><Documents /></ProtectedRoute>} />
                {/* Approval Routes */}
                {/* System Admin Routes */}
                <Route path="/admin/users" element={<ProtectedRoute requiredPermission="manage_users"><UserManagement /></ProtectedRoute>} />
            </Routes>
        </Router>
    );
}

export default App;