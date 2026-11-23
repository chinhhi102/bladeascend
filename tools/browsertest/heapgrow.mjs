// Heap JS phình ở LOẠI NÀO — GAME_SPEC §13.9 (v6.57, việc số 7: "rò rỉ thì vẫn phải tìm").
//
//   node --max-old-space-size=8192 tools/browsertest/heapgrow.mjs http://127.0.0.1:9998
//   node tools/browsertest/heapgrow.mjs --analyze a.heapsnapshot b.heapsnapshot
//
// `leakcheck.mjs` trả lời "có dốc lên không" và "dốc bao nhiêu". Nó KHÔNG trả lời được
// "cái gì ở lại" — sổ tài nguyên chỉ đếm những thứ game tự khai (sprite, text, pool), mà
// 137 MB heap thì phần lớn không nằm trong mấy dòng đó.
//
// Bài này chụp hai ảnh heap — một ở đầu cửa sổ ổn định, một ở cuối — rồi so theo LỚP
// (constructor). Nó dùng đúng chu kỳ kịch bản cố định của `leakcheck.mjs`, nên cái nó
// soi chính là cái bài kia đo.
//
// ── Vì sao hai ảnh chứ không phải một ───────────────────────────────────────────────────
//
// Một ảnh heap chỉ nói "đang có 2 triệu object", mà một game đang chạy thì vốn phải có
// hàng triệu object. Chỉ phần CHÊNH giữa hai ảnh cùng kịch bản mới là thứ không nên có.
//
// ── Vì sao tách khỏi `leakcheck.mjs` ────────────────────────────────────────────────────
//
// Chụp ảnh heap ép dọn rác, dừng trang vài giây và cấp phát thêm trong chính trang đang
// đo. Nhét nó vào bài gác hằng đêm là làm hỏng đúng con số mà bài gác đang đo.
import { createWriteStream, readFileSync, mkdirSync, statSync } from 'node:fs';
import { cpus, loadavg } from 'node:os';
import { dirname } from 'node:path';
import { launch } from './browser.mjs';

const args = process.argv.slice(2);
const sleep = (ms) => new Promise((r) => setTimeout(r, Math.max(0, ms)));

// ── Đọc một ảnh heap, gom theo LỚP ──────────────────────────────────────────────────────
//
// Định dạng `.heapsnapshot`: `nodes` là một mảng số PHẲNG, mỗi nút chiếm `node_fields.length`
// ô liên tiếp; tên lớp là chỉ số vào `strings`. Đọc thẳng mảng phẳng thay vì dựng object
// cho từng nút — hai triệu object nữa trong chính tiến trình đang phân tích thì nó tự chết.
function census(path) {
  const snap = JSON.parse(readFileSync(path, 'utf8'));
  const meta = snap.snapshot.meta;
  const F = meta.node_fields.length;
  const iType = meta.node_fields.indexOf('type');
  const iName = meta.node_fields.indexOf('name');
  const iSize = meta.node_fields.indexOf('self_size');
  const types = meta.node_types[iType];
  const nodes = snap.nodes;
  const strings = snap.strings;
  const by = new Map();
  for (let i = 0; i < nodes.length; i += F) {
    const t = types[nodes[i + iType]];
    // Chuỗi và số gom theo KIỂU, không theo nội dung: 300 000 dòng chữ khác nhau thì
    // mỗi dòng một "lớp" và bảng kết quả thành vô dụng.
    const key = t === 'string' || t === 'number' || t === 'hidden' ? `(${t})` : `${t}:${strings[nodes[i + iName]]}`;
    const cur = by.get(key);
    if (cur) {
      cur.n++;
      cur.bytes += nodes[i + iSize];
    } else {
      by.set(key, { n: 1, bytes: nodes[i + iSize] });
    }
  }
  return { by, total: snap.snapshot.node_count };
}

/**
 * retainers: AI ĐANG GIỮ những đối tượng mới sinh ra của một lớp.
 *
 * Bảng "lớp nào phình" chỉ nói ra cái gì ở lại; nó không nói ai giữ, mà không biết ai giữ
 * thì không sửa được. DevTools có cột "Retainers" nhưng đó là việc tay trong một cửa sổ đồ
 * hoạ — không chạy được trong cron, không dán được vào báo cáo, và không ai làm lại được
 * sau ba tháng. Đường đi của retainer nằm sẵn trong chính tệp ảnh heap, nên tính ở đây.
 *
 * Cách làm: `edges` là mảng phẳng, các cung của nút i nằm liên tiếp nhau; cộng dồn
 * `edge_count` ra được chỗ bắt đầu của từng nút. Quét một lượt là có CHA của mọi nút (giữ
 * một cha đầu tiên cho mỗi nút — đủ để đi ngược lên, và rẻ hơn hẳn việc giữ mọi cha).
 *
 * Chỉ xét nút MỚI: id của V8 giữ nguyên qua các ảnh heap cho cùng một đối tượng, nên những
 * id không có trong ảnh đầu chính là những đối tượng sinh ra trong cửa sổ đo. Đó là thứ
 * đang cần tìm; 3 200 ArrayBuffer có từ trước không nói gì về rò rỉ.
 */
function retainers(bPath, { cls, aPath, limit = Number(process.env.LIMIT ?? 400), depth = 14 } = {}) {
  console.log(`đọc ${bPath} …`);
  const snap = JSON.parse(readFileSync(bPath, 'utf8'));
  const meta = snap.snapshot.meta;
  const NF = meta.node_fields.length;
  const iType = meta.node_fields.indexOf('type');
  const iName = meta.node_fields.indexOf('name');
  const iId = meta.node_fields.indexOf('id');
  const iSize = meta.node_fields.indexOf('self_size');
  const iEdges = meta.node_fields.indexOf('edge_count');
  const nodeTypes = meta.node_types[iType];
  const EF = meta.edge_fields.length;
  const eType = meta.edge_fields.indexOf('type');
  const eName = meta.edge_fields.indexOf('name_or_index');
  const eTo = meta.edge_fields.indexOf('to_node');
  const edgeTypes = meta.edge_types[eType];
  const nodes = snap.nodes;
  const edges = snap.edges;
  const strings = snap.strings;
  const count = nodes.length / NF;

  const label = (i) => {
    const t = nodeTypes[nodes[i * NF + iType]];
    const n = strings[nodes[i * NF + iName]];
    return t === 'string' || t === 'number' ? `(${t})` : `${t}:${n}`;
  };

  // Đường giữ NGẮN NHẤT TỪ GỐC, không phải "cha đầu tiên gặp".
  //
  // Bản đầu tôi viết lấy cha đầu tiên trong lượt quét tuần tự, và nó ra những chuỗi vô
  // nghĩa kiểu "ArrayBuffer ← ArrayBuffer ← ArrayBuffer [4] ← …": nút nào cũng có thể nằm
  // trong một mảng cùng loại, và lượt quét tuần tự gặp cái mảng đó trước. Thứ cần biết là
  // đường từ GỐC xuống — đúng cái DevTools gọi là "distance" và hiện trong cột Retainers.
  // BFS từ nút 0 cho đường ngắn nhất, và đường ngắn nhất gần như luôn là đường đọc ra được
  // tên của chủ sở hữu.
  //
  // Cần chỗ bắt đầu cung của TỪNG nút trước (cộng dồn `edge_count`), vì BFS không đi tuần tự.
  const firstEdge = new Int32Array(count + 1);
  for (let i = 0; i < count; i++) {
    firstEdge[i + 1] = firstEdge[i] + nodes[i * NF + iEdges];
  }
  const parent = new Int32Array(count).fill(-1);
  const viaEdge = new Int32Array(count).fill(-1);
  const queue = new Int32Array(count);
  let head = 0;
  let tail = 0;
  queue[tail++] = 0;
  const seen = new Uint8Array(count);
  seen[0] = 1;
  while (head < tail) {
    const i = queue[head++];
    for (let e2 = firstEdge[i]; e2 < firstEdge[i + 1]; e2++) {
      const to = edges[e2 * EF + eTo] / NF;
      if (!seen[to]) {
        seen[to] = 1;
        parent[to] = i;
        viaEdge[to] = e2;
        queue[tail++] = to;
      }
    }
  }
  const edgeLabel = (ei) => {
    if (ei < 0) {
      return '?';
    }
    const t = edgeTypes[edges[ei * EF + eType]];
    const v = edges[ei * EF + eName];
    return t === 'element' || t === 'hidden' ? `[${v}]` : String(strings[v] ?? v);
  };

  // id của ảnh ĐẦU, để biết nút nào là mới.
  let oldIds = null;
  if (aPath) {
    const a = JSON.parse(readFileSync(aPath, 'utf8'));
    const aNF = a.snapshot.meta.node_fields.length;
    const aIId = a.snapshot.meta.node_fields.indexOf('id');
    oldIds = new Set();
    for (let i = 0; i < a.nodes.length; i += aNF) {
      oldIds.add(a.nodes[i + aIId]);
    }
  }

  const paths = new Map();
  let hit = 0;
  let bytes = 0;
  for (let i = 0; i < count && hit < limit; i++) {
    if (label(i) !== cls) {
      continue;
    }
    if (oldIds && oldIds.has(nodes[i * NF + iId])) {
      continue;
    }
    hit++;
    bytes += nodes[i * NF + iSize];
    const chain = [];
    let cur = i;
    for (let d = 0; d < depth && cur > 0; d++) {
      chain.push(`${edgeLabel(viaEdge[cur])} ← ${label(parent[cur] >= 0 ? parent[cur] : 0)}`);
      cur = parent[cur];
      if (cur < 0) {
        break;
      }
    }
    // Gom theo đường đã CHUẨN HOÁ: bỏ số trong `[123]` và tên cung thuần số.
    //
    // Không chuẩn hoá thì 300 đối tượng cùng một đường sẽ thành 300 nhóm một-phần-tử, vì
    // mỗi cái nằm ở một ô khác nhau của cùng cái mảng — và bảng kết quả không chỉ ra được
    // chủ sở hữu nào đang giữ nhiều nhất.
    const key = chain.join('  ·  ').replace(/\[\d+\]/g, '[i]').replace(/(^|·  )\d+ /g, '$1# ');
    const got = paths.get(key) ?? { n: 0, bytes: 0 };
    got.n++;
    got.bytes += nodes[i * NF + iSize];
    paths.set(key, got);
  }

  console.log(`
== ${hit} đối tượng ${oldIds ? 'MỚI ' : ''}thuộc lớp "${cls}" · ${(bytes / 1048576).toFixed(1)} MB ==`);
  const rows = [...paths.entries()].sort((x, y) => y[1].bytes - x[1].bytes).slice(0, 10);
  for (const [path, v] of rows) {
    console.log(`
${v.n} đối tượng · ${(v.bytes / 1048576).toFixed(1)} MB`);
    for (const [d, step] of path.split('  ·  ').entries()) {
      console.log(`  ${'  '.repeat(d)}${step}`);
    }
  }
  if (rows.length === 0) {
    console.log('(không có đối tượng nào mới thuộc lớp này — thử bỏ --of, hoặc đổi --class)');
  }
}

/** compare in bảng lớp nào phình giữa hai ảnh. */
function compare(aPath, bPath) {
  console.log(`đọc ${aPath} (${(statSync(aPath).size / 1048576).toFixed(0)} MB) …`);
  const a = census(aPath);
  console.log(`đọc ${bPath} (${(statSync(bPath).size / 1048576).toFixed(0)} MB) …`);
  const b = census(bPath);
  const keys = [...new Set([...a.by.keys(), ...b.by.keys()])];
  const rows = keys.map((k) => {
    const x = a.by.get(k) ?? { n: 0, bytes: 0 };
    const y = b.by.get(k) ?? { n: 0, bytes: 0 };
    return { k, dn: y.n - x.n, dmb: (y.bytes - x.bytes) / 1048576, n0: x.n, n1: y.n };
  });
  console.log(`\nsố nút: ${a.total} → ${b.total} (${b.total - a.total >= 0 ? '+' : ''}${b.total - a.total})`);
  console.log('\n== lớp phình nhiều BYTE nhất ==');
  console.log('        ΔMB      Δsố   số đầu → số cuối   lớp');
  for (const r of rows.sort((p, q) => q.dmb - p.dmb).slice(0, 20)) {
    console.log(`${(r.dmb >= 0 ? '+' : '') + r.dmb.toFixed(2).padStart(10)} ${String((r.dn >= 0 ? '+' : '') + r.dn).padStart(8)}   ${String(r.n0).padStart(8)} → ${String(r.n1).padStart(8)}   ${r.k.slice(0, 70)}`);
  }
  console.log('\n== lớp phình nhiều ĐỐI TƯỢNG nhất ==');
  for (const r of rows.sort((p, q) => q.dn - p.dn).slice(0, 15)) {
    console.log(`${String((r.dn >= 0 ? '+' : '') + r.dn).padStart(8)} ${(r.dmb >= 0 ? '+' : '') + r.dmb.toFixed(2).padStart(9)} MB   ${String(r.n0).padStart(8)} → ${String(r.n1).padStart(8)}   ${r.k.slice(0, 70)}`);
  }
  console.log('\nBước sau: mở hai tệp này trong Chrome DevTools → Memory → Load profile →');
  console.log('chọn ảnh thứ hai, lọc theo lớp ở trên, rồi đọc "Retainers" để biết AI đang giữ.');
}

if (args[0] === '--analyze') {
  compare(args[1], args[2]);
  process.exit(0);
}
// --retainers <b.heapsnapshot> --class "<lớp>" [--of <a.heapsnapshot>]
if (args[0] === '--retainers') {
  const flag = (name, dflt) => {
    const i = args.indexOf(name);
    return i >= 0 ? args[i + 1] : dflt;
  };
  retainers(args[1], { cls: flag('--class', 'native:system / JSArrayBufferData'), aPath: flag('--of', '') });
  process.exit(0);
}

// ── Chạy kịch bản rồi chụp hai ảnh ──────────────────────────────────────────────────────
const TARGET = args[0] ?? 'http://127.0.0.1:9998';
/** WARMUP_MIN: chụp ảnh ĐẦU sau ngần này phút — cùng mốc với cửa sổ đo của leakcheck. */
const WARMUP_MIN = Number(process.env.WARMUP_MIN ?? 10);
/** WINDOW_MIN: chạy thêm ngần này phút rồi chụp ảnh thứ hai. */
const WINDOW_MIN = Number(process.env.WINDOW_MIN ?? 20);
/** ZONE_SECS: nhịp đổi khu, 60 s như chu kỳ của `leakcheck.mjs`. */
const ZONE_SECS = Number(process.env.ZONE_SECS ?? 60);
const OUT = process.env.OUT ?? 'server/data/leak-repeat/heap';
const MAX_LOAD = Number(process.env.MAX_LOAD ?? 1.5);

mkdirSync(dirname(OUT), { recursive: true });
const loadPerCPU = () => loadavg()[0] / Math.max(1, cpus().length);
if (loadPerCPU() > MAX_LOAD && process.env.IGNORE_LOAD !== '1') {
  console.log(`máy đang bận ${loadPerCPU().toFixed(2)}/nhân (> ${MAX_LOAD}) — luật 16, chạy lại lúc rảnh (IGNORE_LOAD=1 để kệ).`);
  process.exit(2);
}

const b = await launch({
  headless: true,
  args: ['--use-gl=angle', '--use-angle=gl', '--enable-gpu', '--ignore-gpu-blocklist', '--enable-precise-memory-info'],
});
const ctx = await b.newContext({ viewport: { width: 1280, height: 720 }, locale: 'vi-VN' });
const p = await ctx.newPage();
const cdp = await ctx.newCDPSession(p);
await cdp.send('HeapProfiler.enable').catch(() => undefined);

await p.goto(`${TARGET}/?debug=true`, { waitUntil: 'domcontentloaded' });
await p.waitForSelector('input', { timeout: 30000 });
await p.fill('input', 'KiemTraHeap');
await p.locator('.char-card').nth(0).click();
await p.locator('.home .btn.primary').click();
await p.waitForSelector('.room-board', { timeout: 60000 });
await sleep(4000);

const dbg = (cmd, n = 0) => p.evaluate(([c, k]) => window.kx?.conn?.send({ t: 'dbg', cmd: c, n: k }), [cmd, n]);

/**
 * snapshot: chụp một ảnh heap ra TỆP, không giữ trong bộ nhớ của bài đo.
 *
 * Ảnh của một heap 150 MB là vài trăm MB chữ JSON; nối chuỗi trong RAM rồi mới ghi là
 * cách chắc chắn nhất để chính bài đo chết vì hết bộ nhớ.
 */
async function snapshot(path) {
  await cdp.send('HeapProfiler.collectGarbage').catch(() => undefined);
  await sleep(500);
  const out = createWriteStream(path);
  const onChunk = (ev) => out.write(ev.chunk);
  cdp.on('HeapProfiler.addHeapSnapshotChunk', onChunk);
  await cdp.send('HeapProfiler.takeHeapSnapshot', { reportProgress: false, treatGlobalObjectsAsRoots: true });
  cdp.off('HeapProfiler.addHeapSnapshotChunk', onChunk);
  await new Promise((r) => out.end(r));
  console.log(`  đã chụp ${path} (${(statSync(path).size / 1048576).toFixed(0)} MB)`);
}

const KEYS = ['KeyW', 'KeyA', 'KeyS', 'KeyD'];
/** play: chạy kịch bản — đổi khu mỗi ZONE_SECS giây, đi lại giữa các lượt. */
async function play(mins, fromCycle) {
  const t0 = Date.now();
  let cycle = fromCycle;
  while (Date.now() - t0 < mins * 60000) {
    await dbg('zone', (cycle + 1) % 5);
    await dbg('nogate', 120);
    await dbg('invuln', 120);
    cycle++;
    const until = t0 + (cycle - fromCycle) * ZONE_SECS * 1000;
    while (Date.now() < until) {
      const k = KEYS[cycle % 4];
      await p.keyboard.down(k);
      await sleep(Math.min(520, until - Date.now()));
      await p.keyboard.up(k);
      await sleep(80);
    }
  }
  return cycle;
}

console.log(`khởi động ${WARMUP_MIN} phút (đổi khu mỗi ${ZONE_SECS} s) …`);
let cycle = await play(WARMUP_MIN, 0);
console.log('chụp ảnh heap ĐẦU cửa sổ …');
await snapshot(`${OUT}-a.heapsnapshot`);
console.log(`chạy tiếp ${WINDOW_MIN} phút …`);
cycle = await play(WINDOW_MIN, cycle);
console.log('chụp ảnh heap CUỐI cửa sổ …');
await snapshot(`${OUT}-b.heapsnapshot`);
const heap = await p.evaluate(() => performance.memory?.usedJSHeapSize / 1048576);
console.log(`JS heap cuối: ${heap?.toFixed(1)} MB · ${cycle} lượt đổi khu · tải máy ${loadPerCPU().toFixed(2)}/nhân`);
await b.close();

compare(`${OUT}-a.heapsnapshot`, `${OUT}-b.heapsnapshot`);
