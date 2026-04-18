import { defineConfig, loadEnv } from 'vite' // Thêm loadEnv vào đây
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig(({ mode }) => {
  // Khởi tạo biến env để Vite đọc được các biến VITE_ trong .env
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [
      tailwindcss(),
      react()
    ],
    server: {
      port: 3000,
      proxy: {
        '/api': {
          // Bây giờ biến env đã được định nghĩa và có thể sử dụng
          target: env.VITE_PROXY_TARGET || 'http://localhost:8000', 
          changeOrigin: true,
          secure: false,
        },
        '/ws': {
          target: env.VITE_PROXY_TARGET || 'http://localhost:8000',
          ws: true,
        },
      },
    },
    build: {
      outDir: 'build',
      minify: 'terser', 
    },
  }
})