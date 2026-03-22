import React from 'react';
import ReactDOM from 'react-dom/client';
import './styles/index.css';        // Nền tảng Tailwind
import './styles/animations.css';   // Các hiệu ứng động
import './styles/auth.css';         // Giao diện đăng nhập
import App from './App';
import { AuthProvider } from './context/AuthContext'; // Nhập AuthProvider

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <AuthProvider> {/* Bọc App vào đây là cực kỳ quan trọng */}
      <App />
    </AuthProvider>
  </React.StrictMode>
);