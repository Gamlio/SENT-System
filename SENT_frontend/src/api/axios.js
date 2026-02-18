import axios from 'axios';
// Lấy URL từ .env, nếu không có thì dùng localhost mặc định
const baseURL = process.env.REACT_APP_API_URL

const instance = axios.create({
    baseURL: baseURL,
    withCredentials: true, // Quan trọng: Để gửi cookie/token nếu có
    headers: {
        'Content-Type': 'application/json',
    }
});

export default instance;