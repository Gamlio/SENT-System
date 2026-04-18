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
          target: env.VITE_PROXY_TARGET , 
          changeOrigin: true,
          secure: false,
        },
        '/ws': {
          target: env.VITE_PROXY_TARGET ,
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