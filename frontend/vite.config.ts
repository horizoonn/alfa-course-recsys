import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// В dev-режиме /api проксируется на локальный Go-сервис.
// Если бэкенд ещё не поднят — работает мок-режим (VITE_USE_MOCK=true).
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
