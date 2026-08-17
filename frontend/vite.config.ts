/// <reference types="vitest/config" />
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
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'lcov'],
      include: ['src/**/*.{ts,tsx}'],
      exclude: [
        'src/**/*.test.{ts,tsx}',
        'src/test/**',
        'src/main.tsx',
        // Vendored shadcn component, not app logic.
        'src/components/ui/**',
      ],
      thresholds: {
        lines: 80,
        functions: 80,
        branches: 80,
        statements: 80,
      },
    },
  },
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
