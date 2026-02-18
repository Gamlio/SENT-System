import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import {  useAuth } from './context/AuthContext';
import Navbar from './components/Navbar.jsx';

// Import Pages
import Login from './pages/Login.jsx';
import Register from './pages/Register.jsx';
import Dashboard from './pages/Dashboard.jsx';
import Organizations from './pages/Organizations.jsx'; // Cho R1/R2
import AgentList from './pages/Agents.jsx';
import AgentDetail from './pages/AgentDetail.jsx'; // Trang mới chi tiết
import USBWhitelist from './pages/USBWhitelist.jsx'; // Quản lý tuân thủ
import SoftwarePolicies from './pages/SoftwarePolicies.jsx';

const ProtectedRoute = ({ children, requiredLevel }) => {
    const { user, loading } = useAuth();
    if (loading) return <div className="h-screen bg-[#0f172a]"></div>;
    
    // Kiểm tra quyền truy cập dựa trên RoleLevel
    if (!user) return <Navigate to="/login" />;
    if (requiredLevel && user.level > requiredLevel) return <Navigate to="/" />;

    return (
        <div className="min-h-screen bg-[#0f172a] flex flex-col"> {/* Đổi thành flex-col để xếp dọc */}
            <Navbar /> {/* Menu nằm trên cùng */}
            
            {/* Phần nội dung bên dưới */}
            <div className="flex-1 overflow-auto p-8">
                <div className="max-w-7xl mx-auto"> {/* Căn giữa nội dung cho đẹp (tùy chọn) */}
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
                {/* Trang đăng nhập chung cho tất cả các Role Level */}
                <Route path="/login" element={<Login />} />
                {/* Đăng ký SME mới (Chỉ R1/R2) */}
                <Route path="/register" element={<Register />} />
                {/* Dashboard chung */}
                <Route path="/" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />

                {/* Quản lý SME (Chỉ R1/R2) */}
                <Route path="/admin/orgs" element={<ProtectedRoute requiredLevel={2}><Organizations /></ProtectedRoute>} />

                {/* Quản lý thiết bị & Chi tiết */}
                <Route path="/agents" element={<ProtectedRoute><AgentList /></ProtectedRoute>} />
                <Route path="/agents/:hwid" element={<ProtectedRoute><AgentDetail /></ProtectedRoute>} />

                {/* Quản lý chính sách (Chỉ R3) */}
                <Route path="/policies/usb" element={<ProtectedRoute requiredLevel={3}><USBWhitelist /></ProtectedRoute>} />
                <Route path="/policies/software" element={<ProtectedRoute requiredLevel={3}><SoftwarePolicies /></ProtectedRoute>} />

                <Route path="*" element={<Navigate to="/" />} />
            </Routes>
        </Router>
    );
}

export default App;