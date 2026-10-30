import { chromium } from 'playwright';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const OUT = '/tmp/claude-1000/-home-chinhdc-git-kiemxoaytutien/c75d6ed2-ce95-4f53-af5f-f49ca5b2db8c/scratchpad';
const sizes = [
  { name: 'dien thoai doc 360x800', w: 360, h: 800, dpr: 3, mobile: true },
  { name: 'dien thoai ngang 800x360', w: 800, h: 360, dpr: 3, mobile: true },
  { name: 'dien thoai nho 320x568', w: 320, h: 568, dpr: 2, mobile: true },
  { name: 'cua so hep PC 420x900', w: 420, h: 900, dpr: 1, mobile: false },
];

const b = await chromium.launch({ headless: true, args: ['--use-gl=angle','--use-angle=gl','--enable-gpu','--ignore-gpu-blocklist'] });
for (const s of sizes) {
  const ctx = await b.newContext({ viewport: { width: s.w, height: s.h }, deviceScaleFactor: s.dpr, isMobile: s.mobile, hasTouch: s.mobile });
  const p = await ctx.newPage();
  // KHÔNG bật ?debug=true: bảng đo F3 che nửa màn hình và không phải thứ người chơi thấy.
  await p.goto('http://127.0.0.1:9998/', { waitUntil: 'domcontentloaded' });
  await p.waitForSelector('input', { timeout: 60000 });
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
    const big = document.createElement('div');
    big.className = 'server-line';
    big.textContent = '📢 TOÀN CÕI TU CHÂN · hạ Top 5: Ông Kẹ dịch chuyển hết một vòng để tới chỗ Tám Chuyện chết';
    sf.append(big);
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
      '.server-feed', '.banner', '.burn-btn', '.skill-btn', '.corner', '.rotate-hint'];
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
      soDongBangXepHang: document.querySelectorAll('.room-board .lb-line').length,
      soTinKillFeed: document.querySelectorAll('.feed-line').length,
      minimapPx: Math.round(document.querySelector('.minimap')?.getBoundingClientRect().width ?? 0),
    };
  });
  console.log(`\n=== ${s.name} ===`);
  console.log(`  HUD phu ${res.phanTramManHinhBiHudPhu}% man hinh · bang ${res.soDongBangXepHang} dong · feed ${res.soTinKillFeed} tin · minimap ${res.minimapPx}px`);
  console.log(res.chongLan.length ? '  CHONG LAN:\n    ' + res.chongLan.join('\n    ') : '  khong khoi nao chong len nhau');
  await p.screenshot({ path: `${OUT}/ui_${s.w}x${s.h}.png` });
  await ctx.close();
}
await b.close(); process.exit(0);
