import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [
    tailwindcss(),
    react()
  ],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: env.VITE_PROXY_TARGET, 
        changeOrigin: true,
        secure: false,
      },
      '/ws': {
        target: env.VITE_PROXY_TARGET,
        ws: true,
      },
    },
  },
  build: {
    outDir: 'build',
    minify: 'terser', 
  },
})