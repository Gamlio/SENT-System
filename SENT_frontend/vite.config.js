import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000, // Giữ port 3000 cho quen thuộc
  },
  build: {
    outDir: 'build',
  },
  server: {
    port: 3000,
  }
})