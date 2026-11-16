import { defineConfig, loadEnv, type ProxyOptions } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig(({ mode }) => {
  const rootEnv = loadEnv(mode, '..', '');
  const env = loadEnv(mode, '.', 'VITE_');

  const proxy: Record<string, ProxyOptions> = {
    '/api': {
      target: env.VITE_PROXY_TARGET || `http://127.0.0.1:${rootEnv.APP_PORT || '8080'}`,
      changeOrigin: true,
      configure(server) {
        server.on('error', (_error, _request, response) => {
          if ('writeHead' in response && !response.headersSent) {
            response.writeHead(502, { 'Content-Type': 'application/json; charset=utf-8' });
            response.end(JSON.stringify({ error: {
              code: 'DATA_UNAVAILABLE',
              message: 'Каталог временно недоступен. Попробуйте подключиться позже.',
            } }));
          }
        });
      },
    },
  };

  return {
    plugins: [react()],
    server: {
      port: 5173,
      host: true,
      // Разрешаем внешние домены туннелей (localhost.run, cloudflared и т.п.),
      // иначе Vite отвечает "Blocked request. This host is not allowed".
      allowedHosts: true,
      proxy,
    },
    // Production-сборка через туннель: один бандл вместо десятков модулей dev-сервера,
    // поэтому страница открывается быстро даже на медленном канале.
    preview: {
      port: 4173,
      host: true,
      allowedHosts: true,
      proxy,
    },
  };
});
