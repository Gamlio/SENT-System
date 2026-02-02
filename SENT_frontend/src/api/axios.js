import axios from 'axios';

const instance = axios.create({
    baseURL: process.env.REACT_APP_API_URL, // Tự động lấy từ .env
});

// ... giữ nguyên phần interceptors ...
export default instance;