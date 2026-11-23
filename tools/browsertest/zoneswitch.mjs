// Đổi khu trong bài đo, và XÁC NHẬN là đã đổi thật — GAME_SPEC §13.9 (v6.58, mục 4).
//
// ── Vì sao có tệp này ───────────────────────────────────────────────────────────────────
//
// Mọi bài đo trước đây tự viết `conn.send({t:'dbg',cmd:'zone',n})` rồi `sleep` một lúc, và
// coi như xong. Không bài nào hỏi lại "đã sang khu đó chưa". Hai lỗi đã trốn trong khe đó:
//
//   1. `dbg('zone')` thiếu tham số → server đọc n = 0 → "sang khu 0", mà lệnh bỏ qua khi đã
//      ở khu 0. Sáu lượt đo rò rỉ 30–52 phút tự nhận "50 lượt đổi khu" trong khi gần như
//      không đổi khu lần nào.
//   2. Từ v6.56 cổng bắn đi NGẪU NHIÊN khắp mạng, mà lệnh test thả người vào giữa bản đồ —
//      chỗ thường có cổng. Xin sang khu 1 thì vài giây sau đang đứng ở khu 0.
//
// Cả hai đều là "thành công giả": lệnh gửi đi, không ai báo lỗi, bài đo xanh. Nên từ v6.58
// mọi bài đo đổi khu đều phải đi qua đây.
//
// `__stats().zone` là khu client ĐANG VẼ, không phải khu nó vừa xin — nên nó mới là bằng
// chứng. Chờ tới khi thấy đúng khu đó, hoặc trả về false.

/**
 * switchZone gửi lệnh đổi khu rồi chờ client thật sự vẽ khu đó.
 *
 * @returns true nếu đã sang đúng khu; false nếu hết thời gian chờ mà chưa sang.
 */
export async function switchZone(page, zone, { timeoutMs = 8000, settleMs = 600 } = {}) {
  if (!Number.isInteger(zone) || zone < 0) {
    // Ném chứ không trả false: đây là lỗi của chính bài đo, không phải của game.
    throw new Error(`switchZone: khu phải là số nguyên ≥ 0, nhận được ${zone}`);
  }
  await page.evaluate((z) => window.kx?.conn?.send({ t: 'dbg', cmd: 'zone', n: z }), zone);
  const t0 = Date.now();
  while (Date.now() - t0 < timeoutMs) {
    const now = await page.evaluate(() => window.__stats?.().zone ?? -1);
    if (now === zone) {
      // Chờ thêm một nhịp: khu đã đổi nhưng art của khu mới còn đang nạp, và phần lớn bài
      // đo quan tâm đúng khoảng đó (lỗi vẽ 13/11 nằm ở đây).
      await new Promise((r) => setTimeout(r, settleMs));
      return true;
    }
    await new Promise((r) => setTimeout(r, 250));
  }
  return false;
}

/**
 * mustSwitchZone như trên nhưng NÉM khi không sang được.
 *
 * Dùng cho các bài trong cổng deploy: một bài đo không làm được việc của chính nó thì phải
 * đỏ, không được báo xanh. "Không kiểm được" và "kiểm thấy ổn" là hai kết quả khác nhau.
 */
export async function mustSwitchZone(page, zone, opts) {
  if (!(await switchZone(page, zone, opts))) {
    const now = await page.evaluate(() => window.__stats?.().zone ?? -1);
    throw new Error(`đổi khu THẤT BẠI: xin sang khu ${zone} nhưng client vẫn ở khu ${now}`);
  }
}
