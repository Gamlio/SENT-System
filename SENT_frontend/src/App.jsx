import React, { useState } from 'react'; // Bắt buộc thêm useState ở đây
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from './context/AuthContext';
import { WebSocketProvider } from './context/WebSocketContext';
import Sidebar from './components/Sidebar.jsx';
import Navbar from './components/Navbar.jsx';
import Profile from './components/Profile.jsx';
import VersionsPage from './pages/Versions/VersionsPage.jsx';
import Login from './pages/Auth/Login.jsx';
import ResetPassword from './pages/Auth/ResetPassword.jsx';
import Register from './pages/Auth/Register.jsx';
import Dashboard from './pages/Dashboard/Dashboard.jsx';
import AssetList from './pages/Assets/Assets.jsx';
import AssetDetail from './pages/Assets/AssetsDetail.jsx';
import AssetTypeManagement from './pages/Assets/AssetTypeManagement.jsx';
import IncidentList from './pages/Incident/IncidentManager.jsx';
import AIChatPage from './pages/AIChat/AIChatPage.jsx';
import ApprovalCenter from './pages/approvals/ApprovalCenter.jsx';
import PolicyCenter from './pages/policies/PolicyCenter.jsx'; 
import IncidentDetail from './pages/Incident/IncidentDetail.jsx';
import BehaviorManager from './pages/Behaviors/BehaviorManager.jsx';
import Documents from './pages/Documents/Documents.jsx';     
import UserManagement from './pages/User/UserManagement.jsx';   

import GlobalCopilotDrawer from './components/GlobalCopilotDrawer.jsx'; 

const ProtectedRoute = ({ children, requiredPermission }) => {
    const { user, loading } = useAuth();

    const [isCopilotOpen, setIsCopilotOpen] = useState(false);
    
    if (loading) return <div className="h-screen bg-[#0f172a]"></div>;
    if (!user) return <Navigate to="/login" />;
    if (requiredPermission && user.permissions && !user.permissions[requiredPermission]) {
        return <Navigate to="/" />; 
    }
    return (
        <div className="flex h-screen overflow-hidden bg-[#0f172a] relative">

            <Sidebar />
            <div className="flex-1 flex flex-col overflow-hidden">
                <Navbar onOpenCopilot={() => setIsCopilotOpen(true)} />
                <main className="flex-1 overflow-x-hidden overflow-y-auto p-8 relative">
                    <div className="max-w-7xl mx-auto h-full"> 
                        {children}
                    </div>
                </main>
            </div>

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
            <WebSocketProvider>
                <Routes>
                    {/* Public Routes */}
                    <Route path="/login" element={<Login />} />
                    <Route path="/login/:companyCode" element={<Login />} />
                    <Route path="/register" element={<Register />} />
                    <Route path="/reset-password/:token" element={<ResetPassword />} />
                    {/* Protected Routes */}
                    {/* Core Routes */}
                    <Route path="/" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
                    <Route path="/chat-ai" element={<ProtectedRoute><AIChatPage /></ProtectedRoute>} />
                    <Route path="/profile" element={<ProtectedRoute><Profile /></ProtectedRoute>} />
                    <Route path="/assets" element={<ProtectedRoute requiredPermission="asset_view"><AssetList /></ProtectedRoute>} />
                    <Route path="/assets/:hwid" element={<ProtectedRoute requiredPermission="asset_view"><AssetDetail /></ProtectedRoute>} /> 
<Route path="/assets/types" element={<ProtectedRoute requiredPermission="asset_move"><AssetTypeManagement /></ProtectedRoute>} />                    <Route path="/profile" element={<ProtectedRoute><Profile /></ProtectedRoute>} />
                    <Route path="/incidents" element={<ProtectedRoute requiredPermission="incident_view"><IncidentList /></ProtectedRoute>} />
                    <Route path="/incidents/:id" element={<IncidentDetail />} />
                    <Route path="/behaviors" element={<ProtectedRoute requiredPermission="incident_view"><BehaviorManager /></ProtectedRoute>} />
                    <Route path="/approvals" element={<ProtectedRoute requiredPermission="approval_view"><ApprovalCenter /></ProtectedRoute>} />

                    {/* Compliance & Knowledge Base Routes */}
                    <Route path="/policy-center" element={<ProtectedRoute requiredPermission="policy_view"><PolicyCenter /></ProtectedRoute>} />
                    <Route path="/docs" element={<ProtectedRoute><Documents requiredPermission="doc_view" /></ProtectedRoute>} />
                    <Route path="/users" element={<ProtectedRoute requiredPermission="user_view"><UserManagement /></ProtectedRoute>} />
                    <Route path="/versions" element={<ProtectedRoute><VersionsPage /></ProtectedRoute>} />
                </Routes>
            </WebSocketProvider>
        </Router>
    );
}

export default App;