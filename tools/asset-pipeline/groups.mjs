// Nhóm tải theo GAME_SPEC §13.7 mục 3, và kích thước xuất của từng loại asset (mục 2).
//
// Nguyên tắc chia nhóm: **chỉ nhóm `boot` chặn người chơi**, mọi nhóm khác tải nền trong lúc
// chơi. Nên câu hỏi cho mỗi asset là "KHÔNG có nó thì giây đầu tiên của trận có hỏng không?".
// Thiếu ảnh thần thú Tiên giới thì 99% người chơi không nhận ra (họ còn ở Luyện Khí); thiếu
// font thì game đứng ở màn hình chờ.
import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';

/** Thứ tự khu phải khớp `server/internal/room/zone.go: ZoneMapIDs` — lệch là tải sai khu. */
export const ZONE_MAP_IDS = ['thien_dinh', 'linh_son_coc', 'truc_lam', 'pham_tran', 'bang_nguyet_ho'];

/**
 * zoneObstacleKinds đọc `maps/*.tmj` để biết khu nào dùng loại vật cản nào.
 *
 * Đọc bản đồ chứ không gõ tay vào đây: `type` của đối tượng Tiled là ĐÚNG tên thư mục
 * `assets/src/obstacles/<kind>/`, nên phép ánh xạ không thể lệch. Gõ tay thì lần sau mapgen
 * thêm một loại vật cản là người chơi lại thấy placeholder mà không ai biết tại sao.
 */
export function zoneObstacleKinds(mapsDir) {
  const out = [];
  for (const id of ZONE_MAP_IDS) {
    const m = JSON.parse(readFileSync(join(mapsDir, `${id}.tmj`), 'utf8'));
    const kinds = new Set();
    for (const l of m.layers ?? []) {
      if (l.type !== 'objectgroup' || l.name !== 'obstacles') {
        continue;
      }
      for (const o of l.objects ?? []) {
        if (o.type) {
          kinds.add(o.type);
        }
      }
    }
    out.push({ id, kinds: [...kinds].sort() });
  }
  return out;
}

/** charIds đọc trực tiếp `assets/src/characters/` — thêm nhân vật là tự có nhóm. */
export function charIds(srcDir) {
  return readdirSync(join(srcDir, 'characters'), { withFileTypes: true })
    .filter((e) => e.isDirectory())
    .map((e) => e.name)
    .sort();
}

/**
 * Nhóm `home` — thứ phải có để HIỆN ĐƯỢC TRANG CHỦ, không phải để vào trận.
 *
 * Tách ra khỏi `boot` vì hai mốc chờ khác nhau hẳn: trang chủ chỉ cần font + giao diện +
 * mười chân dung (0,9 MB), còn `boot` còn phải gánh nền bản đồ, vật cản, cổng, quái (3,4 MB).
 * Gộp lại là người chơi ngồi nhìn màn hình trắng trong lúc tải cả cái bản đồ mà họ còn chưa
 * bấm "Vào chơi".
 */
const VFX_BOOT = ['clash_spark', 'run_dust', 'death_smoke', 'sword_shatter', 'shield_break', 'stun_stars', 'particles'];
/** VFX chỉ Tiên giới mới thấy — để `fx_high`. */
const VFX_HIGH = ['levelup_aura', 'lightning'];
/** Kiếm cảnh giới thấp vào `boot`; 9 cảnh giới thường, 0–3 là đủ cho phút đầu. */
const SWORD_BOOT_REALMS = 4;

/**
 * KHÔNG PHÁT HÀNH. Đây là chỗ đáng nhìn kỹ nhất của cả §13.7: ba thư mục này đang nằm
 * trong `manifest.json` cũ và người chơi TẢI THẬT, mà không mã nào đọc tới.
 *
 * - `frames/`, `frames_norm/`: ảnh từng khung, nguồn của `atlas.png`. Game chỉ đọc atlas.
 * - `ref_sheet.png`: tờ tham chiếu cho bên vẽ. Bộ nạp cũ đã lọc tên này rồi, nhưng nó vẫn
 *   bị liệt kê trong manifest nên vẫn tính vào dung lượng.
 * - `body/`, `faces/`: art của hướng A (một thân nhiều mặt) — **chưa có mã nào dùng**.
 *   Chủ dự án chốt hướng A làm sau cùng. Phát hành bây giờ là bắt người chơi tải 1,7 MB
 *   cho một tính năng chưa tồn tại.
 * - `_raw/`: ảnh thô tải từ bên sinh ảnh.
 */
const NEVER_SHIP = [
  /(^|\/)_raw\//,
  /(^|\/)frames\//,
  /(^|\/)frames_norm\//,
  /(^|\/)ref_sheet\.png$/,
  /^characters\/[^/]+\/(body|faces)\//,
  /(^|\/)CREDITS\.md$/,
  // manifest cũ (v1: `files: string[]`). Pipeline tự ghi manifest v2 vào `assets/dist`;
  // phát hành cả bản cũ thì bộ nạp có hai nguồn sự thật và chắc chắn có ngày đọc nhầm bản sai.
  /^manifest\.json$/,
  // Art chưa khu nào dùng (Q31 nhóm 2) và `characters/dan_em` từng được chặn ở đây; từ 1/11
  // chúng nằm hẳn trong `assets/library/` (luật 13, spec §13.7-3b) — pipeline không bao giờ
  // đọc thư mục đó, nên không cần dòng chặn nào. Cần lại thì `git mv` về `assets/src/`.
  // Tệp cấu hình của `cmd/artgen` (công cụ DỰNG art), không phải art. Client không đọc
  // chúng — `npm run check:dead` tìm ra. Giữ ở `assets/src` vì artgen đọc từ đó.
  /^anim_overrides\.json$/,
  /^jitter_exceptions\.json$/,
  /(^|\/)LICENSE[^/]*$/,
  // Logo (02/11): khung gốc 2752 px và font Pattaya chỉ là NGUYÊN LIỆU của `tools/make_logo.py`
  // — chữ đã vẽ sẵn vào PNG, người chơi không cần tải font 410 KB hay khung trơn.
  /^brand\/brand_logo_frame\.png$/,
  /^fonts\/Pattaya/,
];

export function neverShip(rel) {
  return NEVER_SHIP.some((re) => re.test(rel));
}

/**
 * plan trả về `{ group, maxPx, quality, kind }` cho một đường dẫn logic trong `assets/src`.
 *
 * `maxPx` là cạnh dài nhất sau khi thu nhỏ — lấy theo KÍCH THƯỚC HIỂN THỊ (§13.7 mục 2),
 * không theo kích thước bên vẽ gửi tới. `kind`: `image` | `audio` | `data` | `font`.
 */
export function plan(rel, ctx) {
  const seg = rel.split('/');
  const top = seg[0];

  // Atlas (nhân vật và quái) — mọi atlas đều dùng ô 128 px, tức ĐÃ nhỏ hơn mức 256 px §13.7
  // yêu cầu. Nên không thu nhỏ, chỉ đổi PNG → WebP. Thu nhỏ nữa là làm mờ nhân vật mà
  // không đổi lấy được gì: WebP đã lấy hết phần dung lượng.
  if (seg.at(-1) === 'atlas.json' || seg.at(-1) === 'atlas.png') {
    const group = top === 'characters' ? `char_${seg[1]}` : 'boot';
    return seg.at(-1) === 'atlas.json'
      ? { groups: [group], kind: 'atlas-json' }
      : { groups: [group], kind: 'image', maxPx: 0, quality: 82 };
  }

  if (top === 'brand') {
    // Logo tên game (02/11): trang chủ nạp bằng <picture>, trình duyệt chỉ tải ĐÚNG MỘT bản
    // (PC 1 dòng / điện thoại 2 dòng / tiếng nước ngoài) — nên không nằm trong nhóm tải nào.
    // 1288 px = gấp đôi cỡ hiển thị lớn nhất (~640 px) cho màn Retina.
    return { groups: ['brand'], kind: 'image', maxPx: 1288, quality: 82 };
  }

  if (top === 'fonts') {
    // Font chữ Thái (GAME_SPEC §19): chỉ người chọn tiếng Thái mới tải — nhóm riêng, không
    // nằm trong nhóm tải nào khác ("chỉ tải font của ngôn ngữ đang chọn").
    if (/NotoSansThai/i.test(rel)) {
      return { groups: ['lang_th'], kind: 'font' };
    }
    // Font chặn khởi động (`loadFonts` await trước mọi thứ) → `home`, không đụng vào.
    return { groups: ['home'], kind: 'font' };
  }
  // Tệp chữ (giấy phép CC0 của bộ hạt Kenney…): phải đi kèm asset nó bảo hộ, nhưng game
  // không bao giờ tải. Nhóm `legal` không nằm trong bất kỳ nhóm tải nào.
  if (/\.(txt|md|LICENSE)$/i.test(rel)) {
    return { groups: ['legal'], kind: 'data' };
  }
  if (top === 'audio') {
    const bgm = rel.includes('/bgm/');
    return { groups: ['audio'], kind: 'audio', bitrate: bgm ? '128k' : '96k', mono: !bgm };
  }
  if (rel.endsWith('.json') && seg.length === 1) {
    // jitter.json / anim_overrides.json / jitter_exceptions.json: công cụ gỡ lỗi (phím J).
    return { groups: ['debug'], kind: 'data' };
  }

  if (top === 'characters') {
    const char = seg[1];
    const file = seg.slice(2).join('/');
    if (file === 'portrait_neutral.png') {
      // Màn chọn nhân vật cần ảnh của CẢ MƯỜI người ngay — không có thì người chơi chọn mù.
      return { groups: ['home'], kind: 'image', maxPx: 256, quality: 80 };
    }
    if (file.startsWith('portrait_')) {
      // Các tâm trạng khác chỉ hiện trên HUD lúc đang chơi (ô 40–64 px).
      return { groups: [`char_${char}`], kind: 'image', maxPx: 256, quality: 80 };
    }
    return { groups: [`char_${char}`], kind: 'image', maxPx: 256, quality: 80 };
  }

  if (top === 'obstacles') {
    // Vật cản là nhóm NẶNG NHẤT (3,5 MB sau nén) và cũng là nhóm chia theo khu rõ nhất:
    // Linh Sơn Cốc chỉ có vách đá và tảng đá, Phàm Trấn chỉ có sạp/rào/gạch. Gộp hết vào
    // `boot` là bắt người chơi Linh Sơn tải cả cái chợ Phàm Trấn.
    //
    // Một tệp thuộc NHIỀU nhóm khu được: nhóm chỉ là danh sách, tải trùng thì bộ nạp tự bỏ.
    // Loại nào có ở CẢ 5 khu (`stone_wall`) thì để `boot` — tách ra 5 nhóm cũng không tiết
    // kiệm được gì mà còn làm mỗi lần qua cổng lại kiểm tra thừa.
    const kind = seg[1];
    const zones = ctx.zoneKinds.filter((z) => z.kinds.includes(kind)).map((z) => z.id);
    // Vật cản to nhất trong game ~500 đơn vị ≈ 500 px trên màn hình ở scale 1.
    const img = { kind: 'image', maxPx: 512, quality: 80 };
    if (zones.length === ctx.zoneKinds.length) {
      return { groups: ['boot'], ...img };
    }
    if (zones.length === 0) {
      // Ảnh vật cản lẻ ở `obstacles/*.png` (vách đá, mái nhà, cầu treo…) — mapgen chưa
      // sinh ra loại nào dùng chúng, nên chưa khu nào cần. Tải nền.
      return { groups: ['fx_high'], ...img };
    }
    return { groups: zones.map((z) => `zone_${z}`), ...img };
  }

  if (top === 'tiles') {
    // Texture nền LẶP: không thu nhỏ (thu nhỏ là mất chi tiết trên cả mặt bản đồ), q75.
    return { groups: ['boot'], kind: 'image', maxPx: 0, quality: 75 };
  }

  if (top === 'ui') {
    if (rel === 'ui/home_bg.png') {
      // Nền trang chủ phủ toàn màn hình: 1920 là đủ cho mọi màn hình chơi game .io.
      return { groups: ['home'], kind: 'image', maxPx: 1920, quality: 72 };
    }
    // Khung giấy cuộn màn Tổng kết là ảnh **9 ô** (`border-image`): nó bị kéo giãn tới cỡ
    // cả panel, không phải một icon nhỏ. Thu về 256 như icon thì viền mờ nhoè — và nếu chỗ
    // nào còn cắt theo pixel thì hỏng hẳn khung (xem `ui/summary.ts`). Giữ 1024.
    if (rel === 'ui/scroll.png') {
      return { groups: ['home'], kind: 'image', maxPx: 1024, quality: 78 };
    }
    // Huy hiệu, ruy băng, vương miện, icon kỹ năng: ô 24–128 px trên HUD.
    return { groups: ['home'], kind: 'image', maxPx: 256, quality: 80 };
  }

  if (top === 'vfx') {
    const name = seg[1]?.replace(/\.png$/, '');
    if (VFX_HIGH.includes(name)) {
      return { groups: ['fx_high'], kind: 'image', maxPx: 256, quality: 80 };
    }
    return { groups: [VFX_BOOT.includes(name) ? 'boot' : 'fx_high'], kind: 'image', maxPx: 256, quality: 80 };
  }

  if (top === 'swords') {
    const m = /^realm_(\d+)\.png$/.exec(seg[1] ?? '');
    const realm = m ? Number(m[1]) : 0;
    return { groups: [realm < SWORD_BOOT_REALMS ? 'boot' : 'fx_high'], kind: 'image', maxPx: 256, quality: 82 };
  }

  if (top === 'beasts') {
    // Thần thú chỉ xuất hiện từ Chân Tiên trở lên (§3.5).
    return { groups: ['fx_high'], kind: 'image', maxPx: 256, quality: 80 };
  }

  if (top === 'gates' || top === 'zones' || top === 'decals' || top === 'decor' || top === 'ambient' || top === 'monsters' || top === 'items' || top === 'buffs') {
    // Tất cả đều có ở mọi khu và thấy trong vài giây đầu → `boot`.
    const maxPx = top === 'gates' ? 256 : top === 'zones' || top === 'decals' ? 512 : 256;
    return { groups: ['boot'], kind: 'image', maxPx, quality: 80 };
  }

  // Loại chưa biết: đưa vào `fx_high` (tải nền) chứ không vào `boot`. Thêm asset mới mà
  // quên khai báo thì nó làm game thiếu ảnh một lúc, không làm người chơi chờ thêm.
  return { groups: ['fx_high'], kind: rel.endsWith('.json') ? 'data' : 'image', maxPx: 512, quality: 80 };
}
