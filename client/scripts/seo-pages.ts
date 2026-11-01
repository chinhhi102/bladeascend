// Trang chủ tĩnh theo ngôn ngữ (GAME_SPEC §19 "SEO đa ngôn ngữ"): `/`, `/en/`, `/pt/`,
// `/es/`, `/id/`, `/th/` + thẻ hreflang + `sitemap.xml`.
//
// Chạy SAU `vite build`: lấy `dist/index.html` (đã có bundle có hash) làm khuôn, mỗi ngôn
// ngữ thay `<html lang>`, title, description, OG, canonical, và cả khối `#below` (giới
// thiệu · 3 ảnh · cách chơi · FAQ · chân trang) bằng chữ trong `src/i18n/<lang>.json → seo`.
// Nội dung là HTML TĨNH: máy tìm kiếm đọc được mà không cần chạy JavaScript.
//
// Ngôn ngữ nào chưa có file thì bỏ qua (không sinh trang trống). Khoá thiếu rơi về English
// rồi Tiếng Việt, giống game.
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

const ROOT = new URL('../', import.meta.url).pathname;
const DIST = join(ROOT, 'dist');
const I18N = join(ROOT, 'src/i18n');

interface Lang { code: string; name: string; path: string; og: string }
// Cùng thứ tự với `src/i18n/index.ts → LANGS` (file đó dùng import.meta.glob nên Node không nạp được).
const LANGS: Lang[] = [
  { code: 'vi', name: 'Tiếng Việt', path: '/', og: 'vi_VN' },
  { code: 'en', name: 'English', path: '/en/', og: 'en_US' },
  { code: 'pt-BR', name: 'Português', path: '/pt/', og: 'pt_BR' },
  { code: 'es', name: 'Español', path: '/es/', og: 'es_ES' },
  { code: 'id', name: 'Bahasa Indonesia', path: '/id/', og: 'id_ID' },
  { code: 'th', name: 'ไทย', path: '/th/', og: 'th_TH' },
];

interface Shot { alt: string; caption: string }
interface Howto { b: string; t: string }
interface Faq { q: string; a: string }
interface Seo {
  title: string; description: string; intro: string[]; shotsTitle: string; shots: Shot[];
  howtoTitle: string; howto: Howto[]; faqTitle: string; faq: Faq[]; langLabel: string; feedback: string; privacy: string;
}
type J = Record<string, unknown>;

const brand = (JSON.parse(readFileSync(join(ROOT, '../config/game.json'), 'utf8')) as { brand: { name: string; name_vi: string; domain: string; tagline_vi: string } }).brand;

function readJson(code: string): J | null {
  const p = join(I18N, `${code}.json`);
  return existsSync(p) ? (JSON.parse(readFileSync(p, 'utf8')) as J) : null;
}

function isObj(v: unknown): v is J {
  return typeof v === 'object' && v !== null && !Array.isArray(v);
}
function merge(base: J, over: J | null): J {
  if (!over) {
    return base;
  }
  const out: J = { ...base };
  for (const [k, v] of Object.entries(over)) {
    if (v === '' || v === null || (Array.isArray(v) && v.length === 0)) {
      continue;
    }
    out[k] = isObj(base[k]) && isObj(v) ? merge(base[k] as J, v) : v;
  }
  return out;
}

/** Thoát ký tự HTML cho THUỘC TÍNH. Nội dung đoạn văn giữ thẻ <em>/<b> có chủ đích. */
const attr = (s: string): string => s.replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/</g, '&lt;');

function below(seo: Seo, lang: Lang, have: Lang[]): string {
  const name = lang.code === 'vi' ? brand.name_vi : brand.name;
  const langs = have
    .map((l) => (l.code === lang.code ? `<b>${l.name}</b>` : `<a href="${l.path}" hreflang="${l.code}">${l.name}</a>`))
    .join(' · ');
  return `<div id="below">
      <section id="intro">
        <h2>${name}${lang.code === 'vi' ? ` — <span lang="en">${brand.name}</span>` : ''}</h2>
${seo.intro.map((p) => `        <p>${p}</p>`).join('\n')}
      </section>
      <section id="shots">
        <h2>${seo.shotsTitle}</h2>
        <div class="shot-row">
${seo.shots.map((s, i) => `          <figure>
            <img src="/shots/shot_${i + 1}.webp" width="960" height="540" loading="lazy" decoding="async" alt="${attr(s.alt)}" />
            <figcaption>${s.caption}</figcaption>
          </figure>`).join('\n')}
        </div>
      </section>
      <section id="howto">
        <h2>${seo.howtoTitle}</h2>
        <ol>
${seo.howto.map((h) => `          <li><b>${h.b}</b> ${h.t}</li>`).join('\n')}
        </ol>
      </section>
      <section id="faq">
        <h2>${seo.faqTitle}</h2>
${seo.faq.map((f) => `        <details>
          <summary>${f.q}</summary>
          <p>${f.a}</p>
        </details>`).join('\n')}
      </section>
      <footer id="foot">
        <p class="foot-lang">${seo.langLabel} ${langs}</p>
        <p class="foot-links">
          <button type="button" id="foot-feedback">${seo.feedback}</button>
          <a href="/privacy.html">${seo.privacy}</a>
        </p>
        <p class="foot-brand">${brand.name} · ${brand.domain}</p>
      </footer>
    </div>`;
}

function setMeta(html: string, re: RegExp, value: string): string {
  if (!re.test(html)) {
    throw new Error(`seo-pages: không tìm thấy ${re} trong dist/index.html`);
  }
  return html.replace(re, (_m, pre: string, post: string) => `${pre}${attr(value)}${post}`);
}

function page(base: string, d: J, lang: Lang, have: Lang[]): string {
  const seo = d.seo as Seo;
  const name = lang.code === 'vi' ? brand.name_vi : brand.name;
  const title = seo.title.replace('{name}', name);
  const tagline = (d.tagline as string) || brand.tagline_vi;
  const url = `https://${brand.domain}${lang.path}`;
  let h = base.replace(/<html lang="[^"]*"/, `<html lang="${lang.code}"`);
  h = h.replace(/<title>[^<]*<\/title>/, `<title>${attr(title)}</title>`);
  h = setMeta(h, /(<meta name="description" content=")[^"]*(")/, seo.description);
  h = setMeta(h, /(<meta property="og:title" content=")[^"]*(")/, title);
  h = setMeta(h, /(<meta property="og:description" content=")[^"]*(")/, tagline);
  h = setMeta(h, /(<meta property="og:url" content=")[^"]*(")/, url);
  const alt = have.map((l) => `    <link rel="alternate" hreflang="${l.code}" href="https://${brand.domain}${l.path}" />`).join('\n');
  const head = `    <link rel="canonical" href="${url}" />\n${alt}\n    <link rel="alternate" hreflang="x-default" href="https://${brand.domain}/" />\n    <meta property="og:locale" content="${lang.og}" />\n`;
  h = h.replace('</head>', `${head}  </head>`);
  if (lang.code !== 'vi') {
    // Màn hình tải hiện TRƯỚC khi JS nạp từ điển: không đổi thì trang /th/ chớp chữ Việt
    // vài giây trên mạng chậm (đo 1/11: iPhone WebKit ~5 s).
    h = h.replace(/(<div class="boot-name">)[\s\S]*?(<\/div>)/, `$1${name}$2`);
    h = h.replace(/(<div class="tagline">)[^<]*(<\/div>)/, `$1${attr(tagline)}$2`);
    h = h.replace(/(<div class="step" id="boot-step">)[^<]*(<\/div>)/, `$1${attr(d.loadUi as string)}$2`);
  }
  const start = h.indexOf('<!--seo-below');
  const end = h.indexOf('<!--/seo-below-->');
  if (start < 0 || end < 0) {
    throw new Error('seo-pages: thiếu cặp chú thích <!--seo-below--> … <!--/seo-below--> trong index.html');
  }
  // Giữ cặp chú thích trong trang ra: chạy lại `seo:pages` (không qua vite build) vẫn được.
  return `${h.slice(0, start)}<!--seo-below-->\n    ${below(seo, lang, have)}\n    <!--/seo-below-->${h.slice(end + '<!--/seo-below-->'.length)}`;
}

const base = readFileSync(join(DIST, 'index.html'), 'utf8');
const vi = readJson('vi') as J;
const en = readJson('en');
const have = LANGS.filter((l) => l.code === 'vi' || existsSync(join(I18N, `${l.code}.json`)));
for (const lang of have) {
  const d = lang.code === 'vi' ? vi : merge(merge(vi, en), readJson(lang.code));
  const out = join(DIST, lang.path, 'index.html');
  mkdirSync(join(DIST, lang.path), { recursive: true });
  writeFileSync(out, page(base, d, lang, have));
}

const today = new Date().toISOString().slice(0, 10);
const urls = have.map((l) => `  <url>
    <loc>https://${brand.domain}${l.path}</loc>
    <lastmod>${today}</lastmod>
${have.map((a) => `    <xhtml:link rel="alternate" hreflang="${a.code}" href="https://${brand.domain}${a.path}" />`).join('\n')}
  </url>`).join('\n');
writeFileSync(join(DIST, 'sitemap.xml'), `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">
${urls}
</urlset>
`);
if (!existsSync(join(DIST, 'robots.txt'))) {
  writeFileSync(join(DIST, 'robots.txt'), `User-agent: *\nAllow: /\nDisallow: /admin/\nSitemap: https://${brand.domain}/sitemap.xml\n`);
}
console.log(`seo-pages: ${have.map((l) => l.path).join(' ')} + sitemap.xml`);
