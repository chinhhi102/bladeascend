import { defineConfig } from 'vite';

// The Go server (port 8080) owns the API, WebSocket and assets; Vite proxies to it in dev.
// Chỉ nghe trên máy này: không mở ra mạng LAN. Muốn cho điện thoại cùng Wi-Fi vào thì chạy
// `npm run dev -- --host` (và server Go cũng phải chạy với `-addr :8080`).
// Override with KX_SERVER=http://localhost:8090 npm run dev when port 8080 is busy.
const target = process.env.KX_SERVER ?? 'http://localhost:8080';

export default defineConfig({
  server: {
    proxy: {
      '/api': target,
      '/assets': target,
      '/ws': { target: target.replace(/^http/, 'ws'), ws: true },
    },
  },
  // '/assets/' is the game art served by the Go server, so built JS/CSS go to /static/.
  build: { chunkSizeWarningLimit: 1200, assetsDir: 'static' },
});
