import { defineConfig } from 'vite';
import { existsSync, readFileSync } from 'node:fs';
import { splitName } from './brand-split';

// Thương hiệu lấy từ `config/game.json` — GAME_SPEC §20 yêu cầu MỘT nguồn duy nhất.
//
// Vì sao thay lúc BUILD chứ không đọc `/api/config` lúc chạy: tên game nằm trong `<title>`
// và trong màn hình tải, mà màn hình tải phải hiện ngay từ byte đầu tiên của HTML — trước
// cả khi bundle JS tải về (§13.7 mục 4). Chờ một lượt gọi API để biết tên game thì mất đúng
// cái mà màn hình tải sinh ra để tránh.
const brand = JSON.parse(readFileSync(new URL('../config/game.json', import.meta.url), 'utf8')).brand;

const manifest = (): string =>
  JSON.stringify(
    {
      name: `${brand.name_vi} — ${brand.name}`,
      short_name: brand.name,
      description: brand.tagline_vi,
      lang: 'vi',
      start_url: '/',
      // `fullscreen` chứ không `standalone`: game dùng hết màn hình và không có thanh địa
      // chỉ nào đáng giữ. Trên iOS `display` bị bỏ qua, nhưng `apple-mobile-web-app-capable`
      // trong `index.html` lo phần đó.
      display: 'fullscreen',
      orientation: 'any',
      background_color: '#14110e',
      theme_color: '#1d1a16',
      icons: [
        // `/brand/` chứ không `/assets/`: art trong `/assets/` có hash trong tên tệp và
        // đổi mỗi lần build, còn icon PWA phải giữ nguyên địa chỉ — đổi là điện thoại tải
        // lại icon mỗi bản. Xem `tools/make_brandart.py`.
        { src: '/brand/icon_192.png', sizes: '192x192', type: 'image/png' },
        { src: '/brand/icon_512.png', sizes: '512x512', type: 'image/png' },
        { src: '/brand/icon_512.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
      ],
    },
    null,
    2,
  );

// The Go server (port 9999) owns the API, WebSocket and assets; Vite proxies to it in dev.
// host: true để điện thoại cùng Wi-Fi mở được http://<IP-PC>:5173.
// Override with KX_SERVER=http://localhost:8090 npm run dev when port 9999 is busy.
const target = process.env.KX_SERVER ?? 'http://localhost:9999';

// Telemetry chỉ có trong build test (`TELEMETRY=on npm run build`). Bản release build không
// có biến này nên `__TELEMETRY__` là `false` theo nghĩa đen, Rollup cắt nhánh `import()` và
// module telemetry không nằm trong gói — TELEMETRY_SPEC §1.6 yêu cầu đúng điều đó.
const telemetryOn = process.env.TELEMETRY === 'on';
// BUILD_ID để mọi số liệu truy được về đúng bản build đã deploy (§5).
const buildId = process.env.BUILD_ID ?? new Date().toISOString().slice(0, 16).replace(/[-:T]/g, '').replace(/(\d{8})(\d{4})/, '$1_$2');

/**
 * Logo hai dòng cho màn hình tải (chủ dự án chốt 02/11). Màn tải là HTML TĨNH hiện trước khi
 * JS chạy, nên không gọi được `assets.url()` — lấy địa chỉ có hash thẳng từ manifest lúc
 * build. Chưa dựng asset (máy dev mới) → '' và màn tải dùng chữ như cũ.
 */
function bootLogo(): string {
  const p = new URL('../assets/dist/manifest.json', import.meta.url);
  if (!existsSync(p)) {
    return '';
  }
  const man = JSON.parse(readFileSync(p, 'utf8')) as { files: Record<string, { url: string }> };
  const f = man.files['brand/logo_vi_2dong.png'];
  return f ? `<img class="boot-logo" src="/assets/${f.url}" width="1288" height="539" alt="${brand.name_vi}" fetchpriority="high" />` : '';
}

export default defineConfig({
  define: {
    __TELEMETRY__: JSON.stringify(telemetryOn),
    // Thương hiệu vào thẳng mã nguồn lúc build, không qua `/api/config`.
    //
    // Trang chủ vẽ tên game TRƯỚC khi `/api/config` về (nó chỉ cần tên và khẩu hiệu, không
    // cần luật chơi). Chờ API để biết tên game nghĩa là trang chủ nhấp nháy một lần: hiện
    // trống rồi mới có chữ. Cả `index.html` và chỗ này đọc cùng `config/game.json`.
    __BRAND__: JSON.stringify(brand),
    BUILD_ID: JSON.stringify(buildId),
  },
  server: {
    host: true,
    proxy: {
      '/api': target,
      '/assets': target,
      '/ws': { target: target.replace(/^http/, 'ws'), ws: true },
    },
  },
  // '/assets/' is the game art served by the Go server, so built JS/CSS go to /static/.
  build: { chunkSizeWarningLimit: 1200, assetsDir: 'static' },
  plugins: [
    {
      name: 'kx-brand',
      // Manifest PWA dựng TỪ CÙNG `config.brand` (§20), không phải một tệp tĩnh trong
      // `public/`. Một tệp tĩnh là chỗ thứ hai ghi tên game, và chỗ thứ hai là chỗ bị bỏ
      // sót — đúng như lần đổi tên này đã cho thấy.
      buildStart() {
        this.emitFile({ type: 'asset', fileName: 'manifest.webmanifest', source: manifest() });
      },
      configureServer(server) {
        server.middlewares.use('/manifest.webmanifest', (_req, res) => {
          res.setHeader('Content-Type', 'application/manifest+json; charset=utf-8');
          res.end(manifest());
        });
      },
      transformIndexHtml(html) {
        // Thay cả ở `npm run dev`, không chỉ lúc build — nếu không thì trang dev hiện
        // nguyên chuỗi `%BRAND_NAME_VI%` và không ai nhận ra mình đang xem bản hỏng.
        // Hai cụm tên game cho màn hình tải — cùng hàm `splitName` mà bảng trang chủ dùng,
        // nên hai chỗ ngắt dòng ở ĐÚNG một chỗ.
        const parts = splitName(brand.name_vi);
        const map: Record<string, string> = {
          '%BRAND_BOOT_LOGO%': bootLogo() || parts.map((p) => `<span class="boot-part">${p}</span>`).join(' '),
          '%BRAND_NAME_VI%': brand.name_vi,
          '%BRAND_NAME%': brand.name,
          '%BRAND_TAGLINE_VI%': brand.tagline_vi,
          '%BRAND_DOMAIN%': brand.domain,
        };
        let out = html;
        for (const [k, v] of Object.entries(map)) out = out.replaceAll(k, v);
        const left = out.match(/%BRAND_[A-Z_]+%/);
        if (left) throw new Error(`index.html dùng ${left[0]} mà config/game.json không có`);
        return out;
      },
    },
  ],
});
