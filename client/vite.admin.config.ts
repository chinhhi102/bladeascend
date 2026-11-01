import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';

import { defineConfig } from 'vite';

// Trang quản trị — ADMIN_SPEC §2.1: React + antd, BUNDLE RIÊNG.
//
// Dựng vào `dist/admin/` (server phục vụ `<client>/admin` tại `/admin/`, và CHỈ sau khi đăng
// nhập — xem `adminPageGuard`). `index.html` của game không trỏ tới gì ở đây, nên không một
// byte nào của React/antd vào gói tải của người chơi.
//
// Chạy SAU `vite build` của game: bản đó dọn sạch `dist/` trước khi ghi, bản này chỉ dọn
// `dist/admin/`.
const brand = JSON.parse(readFileSync(new URL('../config/game.json', import.meta.url), 'utf8')).brand;
const target = process.env.KX_SERVER ?? 'http://localhost:9999';

export default defineConfig({
  root: fileURLToPath(new URL('./admin', import.meta.url)),
  base: '/admin/',
  // Không dùng `public/` của game: favicon/manifest đó là của trang người chơi.
  publicDir: false,
  define: {
    __BRAND__: JSON.stringify(brand),
  },
  // JSX qua esbuild sẵn có của Vite — không thêm plugin React (spec chỉ duyệt react, react-dom, antd).
  esbuild: { jsx: 'automatic' },
  server: {
    host: true,
    proxy: {
      '/api': target,
      '/assets': target,
    },
  },
  build: {
    outDir: fileURLToPath(new URL('./dist/admin', import.meta.url)),
    emptyOutDir: true,
    assetsDir: 'static',
    chunkSizeWarningLimit: 2000,
    rollupOptions: {
      // antd gắn `"use client"` (dành cho React Server Components) ở đầu mỗi module; Rollup
      // bỏ qua nó và in hàng trăm dòng cảnh báo vô hại, che mất cảnh báo thật.
      onwarn(w, warn) {
        if (w.code !== 'MODULE_LEVEL_DIRECTIVE') {
          warn(w);
        }
      },
    },
  },
});
