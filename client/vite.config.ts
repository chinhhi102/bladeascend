import { defineConfig } from 'vite';

// The Go server (port 8080) owns the API, WebSocket and assets; Vite proxies to it in dev.
// host: true lets phones on the same Wi-Fi open http://<PC-IP>:5173.
const target = 'http://localhost:8080';

export default defineConfig({
  server: {
    host: true,
    proxy: {
      '/api': target,
      '/assets': target,
      '/ws': { target: 'ws://localhost:8080', ws: true },
    },
  },
  build: { chunkSizeWarningLimit: 1200 },
});
