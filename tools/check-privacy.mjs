#!/usr/bin/env node
// Kiểm dữ liệu người chơi lọt vào repo công khai (CLAUDE.md quy tắc 15).
//
//   node tools/check-privacy.mjs          # chạy từ gốc repo; mã thoát 1 nếu có vi phạm
//
// Quét mọi tệp CHỮ đang được git theo dõi trong `docs/` (trừ `docs/private/` — đã gitignore —
// và `docs/_raw/` — chỉ chủ dự án sửa). Bắt hai thứ máy nhận ra chắc chắn:
//   1. UUID đầy đủ (8-4-4-4-12 hex) — dạng `guest_id` mà trình duyệt sinh ra.
//   2. IPv4 công khai — địa chỉ thật của người chơi / máy chủ.
// Tên người chơi và nội dung góp ý thì máy không phân biệt được với chữ thường — phần đó
// vẫn phải người xem. Không thêm thư viện: chỉ dùng Node có sẵn.
import { execFileSync } from 'node:child_process';
import { readFileSync } from 'node:fs';

const UUID_RE = /\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b/gi;
// Không ăn khi đứng sát chữ số/dấu chấm khác (tránh bắt một mẩu giữa "1.2.3.4.5").
const IPV4_RE = /(?<![\d.])(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})(?![\d]|\.\d)/g;
const SKIP_PREFIXES = ['docs/private/', 'docs/_raw/'];
const BINARY_EXT = /\.(png|jpe?g|gif|webp|avif|ico|pdf|zip|gz|ogg|mp3|wav|woff2?|ttf|otf)$/i;

// Dải KHÔNG phải địa chỉ công khai (RFC 1918, loopback, link-local, CGNAT, tài liệu RFC 5737,
// multicast/dự trữ). Địa chỉ tài liệu 203.0.113.x / 198.51.100.x / 192.0.2.x là cố ý: các
// hướng dẫn deploy dùng chúng làm ví dụ, và chúng không bao giờ là máy thật.
function isNonPublic(a, b, c) {
  if (a === 0 || a === 10 || a === 127 || a >= 224) {
    return true;
  }
  if (a === 169 && b === 254) {
    return true;
  }
  if (a === 172 && b >= 16 && b <= 31) {
    return true;
  }
  if (a === 192 && b === 168) {
    return true;
  }
  if (a === 100 && b >= 64 && b <= 127) {
    return true;
  }
  return (a === 192 && b === 0 && c === 2) || (a === 198 && b === 51 && c === 100) ||
    (a === 203 && b === 0 && c === 113);
}

// Chuỗi trông như IP nhưng là SỐ PHIÊN BẢN:
//  - đứng sau "/" hoặc "v" ("Chrome/151.0.0.0", "v1.2.3.4") — User-Agent trình duyệt;
//  - dạng "N.0.0.0" — Chrome giản lược UA thành đúng dạng này, IP thật hầu như không có.
function isVersionLike(line, index, parts) {
  const prev = index > 0 ? line[index - 1] : '';
  if (prev === '/' || prev === 'v' || prev === 'V') {
    return true;
  }
  return parts[1] === 0 && parts[2] === 0 && parts[3] === 0;
}

function ipViolation(line, m) {
  const parts = m.slice(1, 5).map(Number);
  if (parts.some((n) => n > 255)) {
    return false;
  }
  if (isNonPublic(parts[0], parts[1], parts[2])) {
    return false;
  }
  return !isVersionLike(line, m.index, parts);
}

function trackedDocs() {
  const out = execFileSync('git', ['ls-files', '-z', '--', 'docs'], { encoding: 'utf8' });
  return out.split('\0').filter((f) => f && !BINARY_EXT.test(f) &&
    !SKIP_PREFIXES.some((p) => f.startsWith(p)));
}

function scanFile(file) {
  const buf = readFileSync(file);
  if (buf.subarray(0, 8192).includes(0)) {
    return [];
  }
  const hits = [];
  buf.toString('utf8').split('\n').forEach((line, i) => {
    for (const m of line.matchAll(UUID_RE)) {
      hits.push(`${file}:${i + 1}: UUID đầy đủ "${m[0]}" — cắt còn 6 ký tự đầu`);
    }
    for (const m of line.matchAll(IPV4_RE)) {
      if (ipViolation(line, m)) {
        hits.push(`${file}:${i + 1}: IPv4 công khai "${m[0]}" — xoá hoặc che`);
      }
    }
  });
  return hits;
}

function main() {
  const files = trackedDocs();
  const hits = files.flatMap(scanFile);
  if (hits.length === 0) {
    console.log(`check-privacy: ${files.length} tệp trong docs/ — không thấy UUID đầy đủ hay IP công khai.`);
    return;
  }
  console.error(hits.join('\n'));
  console.error(`\ncheck-privacy: ${hits.length} vi phạm. Dữ liệu người chơi thật chỉ được nằm ở ` +
    'docs/private/ (CLAUDE.md quy tắc 15).');
  process.exit(1);
}

main();
