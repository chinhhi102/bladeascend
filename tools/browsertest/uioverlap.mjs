import { chromium } from 'playwright';
import { dimText } from './contrast.mjs';
import { mkdirSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

// Nơi ghi ảnh chụp. `OUT=/đường/dẫn node tools/browsertest/uioverlap.mjs` để đổi.
//
// Trước đây chỗ này là một đường dẫn cứng tới thư mục nháp của MỘT phiên làm việc đã
// kết thúc. Phiên đó không còn, nên tool ghi ảnh vào một thư mục không ai mở ra xem —
// và không báo lỗi, vì `mkdirSync` vẫn tạo được.
// Địa chỉ server. `node tools/browsertest/uioverlap.mjs http://127.0.0.1:9999` để đổi.
// Mặc định là :9999 — cổng deploy thật. Trước đây nó cứng là :9998, cổng của một server
// test tự dựng ở phiên khác; chạy lại ở phiên sau thì chỉ ra ERR_CONNECTION_REFUSED.
const TARGET = process.argv[2] ?? 'http://127.0.0.1:9999';
const OUT = process.env.OUT ?? join(tmpdir(), 'bladeascend-uioverlap');
mkdirSync(OUT, { recursive: true });
const sizes = [
  { name: 'dien thoai doc 360x800', w: 360, h: 800, dpr: 3, mobile: true },
  { name: 'dien thoai ngang 800x360', w: 800, h: 360, dpr: 3, mobile: true },
  { name: 'dien thoai nho 320x568', w: 320, h: 568, dpr: 2, mobile: true },
  { name: 'cua so hep PC 420x900', w: 420, h: 900, dpr: 1, mobile: false },
  // Màn RỘNG: ba luật kích thước chủ dự án chốt 31/10 chỉ lộ ra ở đây. Ảnh chụp báo lỗi
  // hôm đó là 960×540; thêm cả 1920×1080 vì đó là cỡ hay gặp nhất.
  { name: 'PC 960x540', w: 960, h: 540, dpr: 1, mobile: false },
  // 1366×768 — cỡ laptop phổ thông nhất, chủ dự án thêm 1/11. Nó rơi vào KHOẢNG TRỐNG
  // giữa hai tầng CSS cũ (≤720 và ≤560) nên không tầng nào áp: bảng trang chủ giữ nguyên
  // cỡ đầy đủ, cao 809 px trong khung 736 px, và 73 px cuối — nguyên dòng hướng dẫn điều
  // khiển — bị đẩy vào thanh cuộn trong panel mà gần như không ai thấy.
  { name: 'PC 1366x768', w: 1366, h: 768, dpr: 1, mobile: false },
  { name: 'PC 1920x1080', w: 1920, h: 1080, dpr: 1, mobile: false },
];

/**
 * kiemBangMo — các ca 4–7 chủ dự án duyệt 1/11, CHỈ trên điện thoại (bảng thu gọn, chạm mở):
 *  4. nền bảng mở đục ≥ 92% + làm mờ phía sau;
 *  5. bảng mở không CHỒNG CHỮ lên khung HUD trái — hoặc nằm ngoài, hoặc che kín hẳn;
 *  6. thông báo khi bảng mở: không hiện đè lên bảng, bảng đóng thì hiện;
 *  7. bảng mở thì ẩn nút kỹ năng + 🔥;
 * cộng: có nút ✕ đóng được, và 8 giây không chạm thì tự đóng.
 */
async function kiemBangMo(p, s, kiem) {
  await p.locator('.room-board').tap();
  await sleep(400);
  // Thông báo đến LÚC bảng đang mở: phím V đổi chất lượng → banner "Chất lượng: …".
  await p.keyboard.press('KeyV');
  await sleep(200);
  const mo = await p.evaluate(() => {
    const b = document.querySelector('.room-board');
    const st = getComputedStyle(b);
    const alpha = Number((st.backgroundColor.match(/rgba?\(([^)]+)\)/)?.[1] ?? '').split(',')[3] ?? 1);
    const blur = /blur/.test(st.backdropFilter || st.webkitBackdropFilter || '');
    const br = b.getBoundingClientRect();
    const shown = (sel) => {
      const e = document.querySelector(sel);
      if (!e) return false;
      const c = getComputedStyle(e);
      return c.display !== 'none' && c.visibility !== 'hidden' && e.getBoundingClientRect().width > 0;
    };
    // Mục 5: mọi khối HUD trái mà giao với bảng thì phải bị bảng che KÍN (thử 5 điểm).
    const chongChu = ['.status', '.zone-name', '.minimap', '.buffs'].filter((sel) => {
      const e = document.querySelector(sel);
      if (!e || !shown(sel)) return false;
      const r = e.getBoundingClientRect();
      const giao = Math.min(r.right, br.right) > Math.max(r.left, br.left) && Math.min(r.bottom, br.bottom) > Math.max(r.top, br.top);
      if (!giao) return false;
      // Lùi vào trong 25%: góc bảng bo tròn 12px, điểm sát góc rơi ra NGOÀI nền bảng dù mắt
      // nhìn vẫn thấy bị che — lần chạy đầu báo vi phạm giả đúng vì thế.
      const ix = r.width / 4, iy = r.height / 4;
      const pts = [[r.left + ix, r.top + iy], [r.right - ix, r.top + iy], [r.left + ix, r.bottom - iy], [r.right - ix, r.bottom - iy], [(r.left + r.right) / 2, (r.top + r.bottom) / 2]];
      return pts.some(([x, y]) => {
        const hitEl = document.elementFromPoint(Math.min(x, innerWidth - 1), Math.min(y, innerHeight - 1));
        return hitEl && !b.contains(hitEl) && Math.min(r.right, br.right) > x && x > Math.max(r.left, br.left) && Math.min(r.bottom, br.bottom) > y && y > Math.max(r.top, br.top);
      });
    });
    return {
      open: b.classList.contains('open'), alpha, blur, chongChu,
      banner: shown('.banner'), feed: shown('.hud-bottom-left'),
      skill: shown('.skill-btn'), burn: shown('.burn-btn'), close: shown('.rb-close'),
    };
  });
  kiem(mo.open, 'cham vao bang thu gon thi bang mo ra');
  kiem(mo.alpha >= 0.92 && mo.blur, `nen bang mo duc ${Math.round(mo.alpha * 100)}% (>=92%), lam mo phia sau: ${mo.blur}`);
  kiem(mo.chongChu.length === 0, `bang mo khong chong chu len HUD trai (${mo.chongChu.join(', ') || 'khong'})`);
  kiem(!mo.banner && !mo.feed, `thong bao khong hien de len bang mo (banner ${mo.banner}, tin ${mo.feed})`);
  kiem(!mo.skill && !mo.burn, `bang mo thi an nut ky nang + lua (ky nang ${mo.skill}, lua ${mo.burn})`);
  kiem(mo.close, 'bang mo co nut ✕');
  await p.screenshot({ path: `${OUT}/board_open_${s.w}x${s.h}.png` });
  await p.locator('.rb-close').tap();
  await sleep(200);
  const dong = await p.evaluate(() => ({
    open: document.querySelector('.room-board').classList.contains('open'),
    banner: getComputedStyle(document.querySelector('.banner')).display !== 'none',
  }));
  kiem(!dong.open, 'bam ✕ thi bang dong');
  kiem(dong.banner, 'thong bao bi hoan luc bang mo thi HIEN khi bang dong');
  await p.locator('.room-board').tap();
  await sleep(8600);
  // Chờ 8,6 s giữa trận có bot: nhân vật có thể bị chém chết, HUD biến mất. Khi đó KHÔNG
  // tính đạt mà cũng không tính trượt — ghi rõ là bỏ qua để người đọc biết ca này chưa chạy.
  const idle = await p.evaluate(() => document.querySelector('.room-board')?.classList.contains('open') ?? null);
  if (idle === null) {
    console.log('  BO QUA bang tu dong sau 8 giay: nhan vat chet giua luc cho (chay lai de kiem)');
    return;
  }
  kiem(!idle, 'bang mo 8 giay khong cham thi tu dong');
}

/** Vượt luật kích thước thì cả bài TRƯỢT — đếm ở đây. */
let viPham = 0;

const b = await chromium.launch({ headless: true, args: ['--use-gl=angle','--use-angle=gl','--enable-gpu','--ignore-gpu-blocklist'] });
for (const s of sizes) {
  const ctx = await b.newContext({ viewport: { width: s.w, height: s.h }, deviceScaleFactor: s.dpr, isMobile: s.mobile, hasTouch: s.mobile });
  const p = await ctx.newPage();
  // KHÔNG bật ?debug=true: bảng đo F3 che nửa màn hình và không phải thứ người chơi thấy.
  await p.goto(TARGET, { waitUntil: 'domcontentloaded' });
  await p.waitForSelector('input', { timeout: 60000 });

  // ── TRANG CHỦ: đo TRƯỚC khi bấm "Vào chơi" ────────────────────────────────────────
  //
  // Phần còn lại của bài này đo HUD trong trận. Nhưng màn hình đầu tiên người lạ nhìn
  // thấy là trang chủ, và nó hỏng theo một kiểu mà mắt thường khó bắt: panel có
  // `overflow-y: auto`, nên khi nội dung không vừa thì phần thừa KHÔNG tràn ra ngoài —
  // nó trôi vào một thanh cuộn bên trong gần như vô hình.
  //
  // Vì vậy KHÔNG đo bằng `display`/`visibility`: `getComputedStyle` vẫn báo dòng hướng
  // dẫn là `block` trong khi người ngồi trước máy không thấy dòng nào. Bằng chứng đúng
  // là `scrollHeight - clientHeight`.
  await p.waitForSelector('.home .panel', { timeout: 60000 });
  await sleep(900);
  const home = await p.evaluate(() => {
    const panel = document.querySelector('.home .panel');
    const hint = document.querySelector('.home .hint');
    const r = panel.getBoundingClientRect();
    const st = hint ? getComputedStyle(hint) : null;
    const hintShown = !!st && st.display !== 'none' && st.visibility !== 'hidden';
    // Dòng hướng dẫn nằm TRONG vùng nhìn thấy của panel hay đã trôi quá đáy nó?
    let hintInView = false;
    if (hintShown) {
      const hr = hint.getBoundingClientRect();
      hintInView = hr.bottom <= r.bottom + 1 && hr.top >= r.top - 1;
    }
    // Tên game: có chữ nào rơi lẻ một mình xuống dòng cuối không?
    const logo = document.querySelector('.home .logo');
    const tops = [];
    const walk = (n) => {
      if (n.nodeType === 3) {
        const rng = document.createRange();
        let i = 0;
        for (const word of n.textContent.split(' ')) {
          if (word) {
            rng.setStart(n, i); rng.setEnd(n, i + word.length);
            tops.push(Math.round(rng.getBoundingClientRect().top));
          }
          i += word.length + 1;
        }
        return;
      }
      for (const c of n.childNodes) walk(c);
    };
    if (logo) walk(logo);
    const maxTop = tops.length ? Math.max(...tops) : 0;
    return {
      cachMepDuoi: Math.round(innerHeight - r.bottom),
      biCuonTrongPanel: panel.scrollHeight - panel.clientHeight,
      hintShown, hintInView,
      soDongTen: new Set(tops).size,
      soChuDongCuoi: tops.filter((t) => t === maxTop).length,
      soH1: document.querySelectorAll('h1').length,
      theDaiDong: [...document.querySelectorAll('.char-card .card-name')].filter((e) => {
        const lh = parseFloat(getComputedStyle(e).lineHeight) || e.getBoundingClientRect().height;
        return e.getBoundingClientRect().height > lh * 1.4;
      }).length,
    };
  });
  console.log(`\n=== ${s.name} ===`);
  console.log(`  TRANG CHU: panel cach mep duoi ${home.cachMepDuoi}px · cuon trong panel ${home.biCuonTrongPanel}px · ten game ${home.soDongTen} dong`);
  const kiemHome = (ok, nhan) => {
    console.log(`  ${ok ? 'OK ' : 'VI PHAM'} ${nhan}`);
    if (!ok) viPham++;
  };
  kiemHome(home.biCuonTrongPanel <= 1,
    `bang trang chu khong bi cat vao thanh cuon (thua ${home.biCuonTrongPanel}px)`);
  kiemHome(home.cachMepDuoi >= 4, `bang trang chu khong cham mep duoi (cach ${home.cachMepDuoi}px)`);
  // Dòng hướng dẫn CHO PHÉP ẩn ở cửa sổ rất thấp (luật CSS ≤560px) — ở đó nó đọc được
  // trong trận. Nhưng hiện mà nằm ngoài vùng nhìn thấy thì là hỏng, không phải lựa chọn.
  kiemHome(!home.hintShown || home.hintInView, 'dong huong dan dieu khien nam trong vung nhin thay');
  if (!s.mobile && s.h >= 640) {
    kiemHome(home.hintShown && home.hintInView, 'man PC du cao thi PHAI thay dong huong dan');
  }
  kiemHome(home.soChuDongCuoi !== 1 || home.soDongTen === 1,
    `ten game khong bo le mot chu o dong cuoi (${home.soDongTen} dong, dong cuoi ${home.soChuDongCuoi} chu)`);
  kiemHome(home.soH1 === 1, `trang co dung 1 the h1 (dang co ${home.soH1})`);
  kiemHome(home.theDaiDong === 0, `the nhan vat deu 1 dong (${home.theDaiDong} the bi xuong dong)`);
  await p.screenshot({ path: `${OUT}/home_${s.w}x${s.h}.png` });

  await p.fill('input', 'UiTest');
  await p.getByText('Vào chơi', { exact: true }).click();
  await p.waitForSelector('.room-board', { timeout: 60000 });
  await sleep(3000);
  // Ép cho CẢ HAI loại tin cùng hiện. Không ép thì phép đo phụ thuộc vào việc trong lúc
  // chụp có ai bị hạ hay không — và nó đã báo "không chồng lấn" đúng một lượt chỉ vì lúc
  // ấy chưa có tin toàn server nào, trong khi dải tin đó đè thẳng lên nút lửa.
  await p.evaluate(() => {
    const feed = document.querySelector('.feed');
    const sf = document.querySelector('.server-feed');
    for (let i = feed.children.length; i < 5; i++) {
      const d = document.createElement('div');
      d.className = 'feed-line';
      d.textContent = '[Băng Nguyệt Hồ] Thím Ba Bầu nằm xuống, Cô Út Mắm phủi tay đi tiếp';
      feed.append(d);
    }
    // Nhét BA tin loan để kiểm luật "chỉ một cái tại một thời điểm" (chủ dự án chốt
    // 31/10). Nhét một cái thì không bao giờ biết luật có chạy không.
    for (let i = 0; i < 3; i++) {
      const big = document.createElement('div');
      big.className = 'server-line';
      big.textContent = '📢 TOÀN CÕI TU CHÂN · hạ Top 5: Ông Kẹ dịch chuyển hết một vòng để tới chỗ Tám Chuyện chết';
      sf.append(big);
    }
    const bn = document.querySelector('.banner');
    bn.textContent = 'Rương Báu sắp xuất hiện trong 10s!';
    bn.style.display = 'block';
  });
  // Nhét dư rồi bắt HUD tự cắt theo LUẬT THẬT (`trimLists` gắn vào sự kiện resize), thay vì
  // tự chọn số tin trong bài đo. Có thế thì con số "feed N tin" ở dưới mới kiểm được luật
  // chứ không phải kiểm chính cái mình vừa nhét vào.
  await p.evaluate(() => window.dispatchEvent(new Event('resize')));
  await sleep(300);

  const res = await p.evaluate(() => {
    const SEL = ['.status', '.buffs', '.zone-name', '.minimap', '.room-board', '.feed',
      '.server-feed', '.banner', '.burn-btn', '.skill-btn', '.corner', '.rotate-hint', '.fb-btn'];
    const boxes = [];
    for (const s of SEL) {
      const e = document.querySelector(s);
      if (!e) continue;
      const st = getComputedStyle(e);
      if (st.display === 'none' || st.visibility === 'hidden' || +st.opacity === 0) continue;
      const r = e.getBoundingClientRect();
      if (r.width < 1 || r.height < 1) continue;
      boxes.push({ s, x: r.x, y: r.y, w: r.width, h: r.height });
    }
    const over = [];
    for (let i = 0; i < boxes.length; i++) {
      for (let j = i + 1; j < boxes.length; j++) {
        const a = boxes[i], c = boxes[j];
        const ox = Math.min(a.x + a.w, c.x + c.w) - Math.max(a.x, c.x);
        const oy = Math.min(a.y + a.h, c.y + c.h) - Math.max(a.y, c.y);
        if (ox > 2 && oy > 2) over.push(`${a.s} ⨯ ${c.s} (${Math.round(ox)}×${Math.round(oy)}px)`);
      }
    }
    const W = innerWidth, H = innerHeight;
    // Phần màn hình bị HUD phủ, tính theo LƯỚI ô 8px nên vùng chồng nhau không bị đếm hai lần.
    const cell = 8;
    let hit = 0, total = 0;
    for (let y = 0; y < H; y += cell) {
      for (let x = 0; x < W; x += cell) {
        total++;
        if (boxes.some((b) => x + cell > b.x && x < b.x + b.w && y + cell > b.y && y < b.y + b.h)) hit++;
      }
    }
    return {
      chongLan: over,
      phanTramManHinhBiHudPhu: Math.round((hit / total) * 100),
      soDongBangXepHang: document.querySelectorAll('.room-board .rb-row:not(.rb-me)').length,
      soTinKillFeed: document.querySelectorAll('.feed-line').length,
      minimapPx: Math.round(document.querySelector('.minimap')?.getBoundingClientRect().width ?? 0),
      // Ba luật kích thước (chủ dự án chốt 31/10).
      soTinLoan: document.querySelectorAll('.server-line').length,
      rongChatPhanTram: Math.round(((document.querySelector('.feed')?.getBoundingClientRect().width ?? 0) / W) * 100),
      caoBangPhanTram: Math.round(((document.querySelector('.room-board')?.getBoundingClientRect().height ?? 0) / H) * 100),
      dangDoc: H > W,
      // Ca 8 (duyệt 1/11): nút góp ý 💬 không được đè lên tin loan / kill feed.
      tinBiNutGopYChe: (() => {
        const fb = document.querySelector('.fb-btn')?.getBoundingClientRect();
        if (!fb) return [];
        return ['.server-line', '.feed-line'].flatMap((sel) => [...document.querySelectorAll(sel)].filter((e) => {
          const r = e.getBoundingClientRect();
          return r.width > 0 && Math.min(r.right, fb.right) - Math.max(r.left, fb.left) > 2 &&
            Math.min(r.bottom, fb.bottom) - Math.max(r.top, fb.top) > 2;
        }).map(() => sel));
      })(),
      bangBiCat: (() => {
        const b = document.querySelector('.room-board');
        return b ? b.scrollHeight - b.clientHeight : 0;
      })(),
    };
  });
  console.log(`  HUD phu ${res.phanTramManHinhBiHudPhu}% man hinh · bang ${res.soDongBangXepHang} dong · feed ${res.soTinKillFeed} tin · minimap ${res.minimapPx}px`);
  console.log(res.chongLan.length ? '  CHONG LAN:\n    ' + res.chongLan.join('\n    ') : '  khong khoi nao chong len nhau');

  // ── Ba luật kích thước, chủ dự án chốt 31/10 ──────────────────────────────────────────
  const tranChat = res.dangDoc ? 33 : 25;
  const kiem = (ok, nhan) => {
    console.log(`  ${ok ? 'OK ' : 'VI PHAM'} ${nhan}`);
    if (!ok) viPham++;
  };
  // +1% dung sai: `getBoundingClientRect` trả số lẻ, và 25vw của 963px ra 240.75px.
  kiem(res.soTinLoan <= 1, `tin loan cung luc: ${res.soTinLoan} (toi da 1)`);
  kiem(res.rongChatPhanTram <= tranChat + 1,
    `khung chat rong ${res.rongChatPhanTram}% be ngang (toi da ${tranChat}% — man ${res.dangDoc ? 'doc' : 'ngang'})`);
  kiem(res.caoBangPhanTram <= 51,
    `bang xep hang cao ${res.caoBangPhanTram}% chieu cao (toi da 50%)`);
  // Giữ đúng 50% mà cắt ngang một dòng thì vẫn là hỏng: người chơi thấy nửa chữ.
  kiem(res.bangBiCat <= 1, `bang xep hang khong bi cat ngang (thua ${res.bangBiCat}px)`);
  kiem(res.tinBiNutGopYChe.length === 0, `nut gop y khong de len tin (${res.tinBiNutGopYChe.join(', ') || 'khong'})`);
  await p.screenshot({ path: `${OUT}/ui_${s.w}x${s.h}.png` });
  if (s.mobile) {
    await kiemBangMo(p, s, kiem);
  }

  // Doan gioi thieu §20: phai DOC DUOC o trang chu.
  //
  // No nam duoi khung game tren nen trang mau toi, va lan dau toi viet `color: var(--ink)`
  // — tuc chu #1d1a16 tren nen #2b2b2b, 1,1:1. Anh chup trong y nhu mot khoang trong.
  await p.evaluate(() => {
    document.body.classList.remove('in-game');
    document.getElementById('intro')?.scrollIntoView();
  });
  const mo = await dimText(p, '#intro', 4.5);
  kiem(mo.length === 0, `doan gioi thieu doc duoc (tuong phan >= 4,5:1)${mo.length ? ' — mo: ' + mo.join(', ') : ''}`);

  await ctx.close();
}
await b.close();
if (viPham > 0) {
  console.error(`\nTRUOT: ${viPham} lan vi pham luat kich thuoc HUD.`);
  process.exit(1);
}
console.log('\nDAT: HUD khong khoi nao vuot luat kich thuoc.');
process.exit(0);
