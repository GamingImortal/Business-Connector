import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5174,
    // Forward API calls to the Go backend.
    proxy: { '/api': 'http://localhost:8080' },
  },
})
