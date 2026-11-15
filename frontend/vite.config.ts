import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig(({ mode }) => {
  const rootEnv = loadEnv(mode, '..', '');
  const env = loadEnv(mode, '.', 'VITE_');

  return {
    plugins: [react()],
    server: {
      port: 5173,
      host: true,
      // Разрешаем внешние домены туннелей (localhost.run, cloudflared и т.п.),
      // иначе Vite отвечает "Blocked request. This host is not allowed".
      allowedHosts: true,
      proxy: {
        '/api': {
          target: env.VITE_PROXY_TARGET || `http://127.0.0.1:${rootEnv.APP_PORT || '8080'}`,
          changeOrigin: true,
        },
      },
    },
  };
});
