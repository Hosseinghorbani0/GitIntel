import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

const apiTarget = process.env.GITINTEL_API_URL?.trim() || 'http://localhost:8080'
const frontendPort = Number(process.env.GITINTEL_FRONTEND_PORT || 5173)

export default defineConfig({
  plugins: [react()],
  server: {
    host: '127.0.0.1',
    port: frontendPort,
    strictPort: true,
    proxy: {
      '/api': {
        target: apiTarget,
        changeOrigin: true,
      },
    },
  },
})
