import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  /* import.meta.dirname, not node:path — Node 20.11+ and Vite's config
     loader both define it natively, so no extra import is needed. */
  resolve: {
    alias: {
      '@': `${import.meta.dirname}/src`,
    },
  },
  // SPA: every unknown path falls back to index.html, rendered fully in the browser.
  appType: 'spa',
  server: {
    /* Keeps the API same-origin in development, so CORS is only exercised in
       production. Run: CALCULATOR_ALLOWED_ORIGINS=http://localhost:5173 go run ./cmd/api */
    proxy: {
      '/evaluations': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: false,
      },
    },
  },
})
