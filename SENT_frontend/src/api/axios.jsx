import axios from 'axios';

const baseURL = import.meta.env.VITE_API_URL


const instance = axios.create({
    baseURL: baseURL,
    headers: {
        'Content-Type': 'application/json',
    }
});
// Tự động chèn Token vào Header trước khi gửi request đi
instance.interceptors.request.use(
    (config) => {
        // Lấy token đã lưu lúc đăng nhập (Key phải trùng với bên AuthContext)
        const token = localStorage.getItem('sent_token'); 
        
        if (token) {
            // Đính kèm vào Header theo chuẩn JWT: "Bearer <token>"
            config.headers['Authorization'] = `Bearer ${token}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

export default instance;