import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  
  build: {
    // Output to dist/ directory (will be embedded in Go binary)
    outDir: 'dist',
    // Clean the output directory before building
    emptyOutDir: true,
  },
  
  server: {
    // Proxy API requests to PocketBase during development
    proxy: {
      // Forward /api/* requests to PocketBase
      '/api': {
        target: 'http://localhost:8090',
        changeOrigin: true,
      },
      // Forward /_/* requests (PocketBase admin UI) to PocketBase
      '/_': {
        target: 'http://localhost:8090',
        changeOrigin: true,
      },
    },
  },
})
