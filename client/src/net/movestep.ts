import { contains, pushOut } from './geom.ts';
import type { GameConfig } from '../config.ts';
import { TIER_BORDER, type GameMap, type MapObstacle, type Shape } from './protocol.ts';

/**
 * Va chạm khi di chuyển — spec §8.2a.
 *
 * Module này **không phụ thuộc Pixi, DOM hay trạng thái trận đấu** để hai nơi dùng chung
 * đúng một đoạn mã: vòng lặp dự đoán của client và script đối chiếu
 * `scripts/collision-check.ts`. Nó là bản port của `game.resolveBlocking` bên Go, và có bài
 * kiểm tra 1 000 bước đi ngẫu nhiên sát vật cản để bắt mọi chỗ hai bên lệch nhau.
 *
 * Vì sao phải khớp đến từng con số: client dự đoán vị trí của chính mình để nhân vật nhúc
 * nhích ngay khi bấm phím, không chờ server. Lệch một chút thôi là server kéo nhân vật về
 * mỗi tick — 20 lần/giây — và người chơi thấy rung liên tục. Đo thật trước khi sửa: đi tới
 * rìa bản đồ cho ra **19,7 lần hiệu chỉnh/giây**, lệch tới 14,5 đơn vị.
 */

export interface Vec {
  x: number;
  y: number;
}

/** Số vòng đẩy ra trong một tick, đủ để xử lý góc giữa hai vật cản (spec §8.2a). */
const MAX_ITER = 3;

/** sorted sắp vật cản theo id tăng dần, nhớ lại kết quả vì danh sách không đổi trong trận. */
const sortCache = new WeakMap<object, MapObstacle[]>();

function sorted(obstacles: MapObstacle[]): MapObstacle[] {
  let out = sortCache.get(obstacles);
  if (!out) {
    out = [...obstacles].sort((a, b) => a.id - b.id);
    sortCache.set(obstacles, out);
  }
  return out;
}

/** blockShape: hình dùng để CHẶN, ưu tiên footprint đã tính sẵn (mirror của Obstacle.Block). */
function blockShape(o: MapObstacle): MapObstacle['s'] {
  return o.f && (o.f.w ?? 0) > 0 ? o.f : o.s;
}

/**
 * clampPlay giữ một hình tròn bán kính r bên trong **vùng chơi được**, tức là đã trừ lề
 * tường biên (`play_margin`), đúng như `game.clampToPlay` bên server.
 *
 * Đây chính là chỗ từng sai: client kẹp theo cả bản đồ `[r, size−r]` còn server kẹp theo
 * `[margin+r, size−margin−r]`. Bản đồ 12 000 với lề 420 nghĩa là **cả vành ngoài rộng 420
 * đơn vị** — khoảng 13% diện tích — chỗ nào client cũng dự đoán ra ngoài vùng server cho
 * phép, nên bị kéo về mỗi tick.
 */
export function clampPlay(p: Vec, r: number, size: number, playMargin: number): Vec {
  let lo = playMargin + r;
  let hi = size - playMargin - r;
  if (hi < lo) {
    // Bản đồ bé hơn hai lần lề: lùi về kẹp theo cả bản đồ, giống hệt server.
    lo = r;
    hi = size - r;
  }
  return { x: Math.min(Math.max(p.x, lo), hi), y: Math.min(Math.max(p.y, lo), hi) };
}

/**
 * resolveBlocking đẩy một hình tròn bán kính r ra khỏi mọi vật cản đặc, rồi kẹp vào vùng
 * chơi được. Bản port của `game.resolveBlocking`.
 *
 * Đẩy ra theo **pháp tuyến tiếp xúc** chính là trượt dọc mà spec §8.2a yêu cầu: chỉ bỏ phần
 * chuyển động đâm vào mặt vật cản, phần chạy song song mặt được giữ nguyên. Không bao giờ
 * bật ngược.
 */
export function resolveBlocking(
  p: Vec,
  r: number,
  map: Pick<GameMap, 'obstacles' | 'size'> & { play_margin?: number },
  broken: (id: number) => boolean,
  /** only: lọc vật cản nào được xét. Dùng cho Gom Đơn — chỉ còn biên khu chặn. */
  only?: (o: MapObstacle) => boolean,
): Vec {
  const margin = map.play_margin ?? 0;
  let pos = clampPlay(p, r, map.size, margin);
  // Đẩy ra theo THỨ TỰ ID TĂNG DẦN, giống hệt server. Đẩy lần lượt ra khỏi nhiều vật cản
  // là phép KHÔNG GIAO HOÁN: đứng ở góc hai vật cản, đẩy A trước rồi B cho ra chỗ khác hẳn
  // đẩy B trước rồi A. Bài đối chiếu 1 000 bước bắt được 4 ca lệch vì đúng chuyện này, ca
  // nặng nhất 173,7 đơn vị.
  const byId = sorted(map.obstacles);
  for (let iter = 0; iter < MAX_ITER; iter++) {
    let moved = false;
    for (const o of byId) {
      if (broken(o.id) || (only && !only(o))) {
        continue;
      }
      const np = pushOut(blockShape(o), pos, r);
      if (np) {
        pos = np;
        moved = true;
      }
    }
    if (!moved) {
      break;
    }
  }
  return clampPlay(pos, r, map.size, margin);
}


/**
 * zoneMult: hệ số tốc độ của vùng địa hình đang đứng (spec §8.3) — bản port của
 * `game.ZoneMult`.
 *
 * Hai chỗ từng lệch so với server, cả hai đều làm server kéo nhân vật về mỗi tick:
 *   - Vùng **nước** cũng làm chậm như bùn (§8.6 mục 4); client trước đây chỉ xét `mud`.
 *   - Vùng bây giờ là **đa giác**, không phải hình chữ nhật — xem `contains`.
 */
export function zoneMult(cfg: GameConfig, map: Pick<GameMap, 'zones'>, p: Vec, dir: Vec): number {
  let mult = 1;
  for (const z of map.zones ?? []) {
    if (!contains(z.s, p.x, p.y)) {
      continue;
    }
    if (z.k === 'mud' || z.k === 'water') {
      mult *= cfg.zones.mud_mult;
    } else if (z.k === 'snow') {
      mult *= cfg.zones.snow_mult;
    } else if (z.k === 'wind') {
      mult *= windMult(cfg, dir, z.wd ?? 0);
    }
  }
  return mult;
}

/** windMult: theo gió thì nhanh, ngược gió thì chậm, ngang gió giữ nguyên (spec §8.3). */
export function windMult(cfg: GameConfig, dir: Vec, windDir: number): number {
  const cos = Math.max(-1, Math.min(1, dir.x * Math.cos(windDir) + dir.y * Math.sin(windDir)));
  const deg = (Math.acos(cos) * 180) / Math.PI;
  if (deg < cfg.zones.wind_with_max_deg) {
    return cfg.zones.wind_with;
  }
  if (deg > cfg.zones.wind_against_min_deg) {
    return cfg.zones.wind_against;
  }
  return 1;
}

/** onIce: điểm này có nằm trên mặt băng không (spec §8.0). */
export function onIce(map: Pick<GameMap, 'zones'>, p: Vec): boolean {
  for (const z of map.zones ?? []) {
    if (z.k === 'ice' && contains(z.s, p.x, p.y)) {
      return true;
    }
  }
  return false;
}

/**
 * iceStep: một bước trượt trên băng (spec §8.0) — bản port của `game.slideOnIce`.
 *
 *     vận tốc mới = vận tốc cũ × friction + hướng input × accel × dt
 *
 * Trả về cả vị trí lẫn vận tốc mới, vì vận tốc là TRẠNG THÁI: bước sau phụ thuộc bước
 * trước. Đó cũng là lý do server phải gửi kèm `vx/vy` cho chính chủ — không có nó thì mỗi
 * lần nhận snapshot client lại phát lại dự đoán từ vận tốc 0 và nhân vật giật trên cả hồ.
 */
export function iceStep(
  cfg: GameConfig,
  from: Vec,
  vel: Vec,
  angle: number,
  moving: boolean,
  dt: number,
  speedMult: number,
): { pos: Vec; vel: Vec } {
  const ic = cfg.ice;
  let vx = vel.x * ic.friction;
  let vy = vel.y * ic.friction;
  if (moving) {
    const a = ic.accel * speedMult * dt;
    vx += Math.cos(angle) * a;
    vy += Math.sin(angle) * a;
  }
  return { pos: { x: from.x + vx * dt, y: from.y + vy * dt }, vel: { x: vx, y: vy } };
}

/**
 * moveStep chạy đúng MỘT bước di chuyển, khớp từng con số với `game.movePlayer` bên server.
 *
 * Đây là hàm mà `scripts/collision-check.ts` phát lại trên dữ liệu do server sinh ra, nên
 * mọi thứ ảnh hưởng tới vị trí — tốc độ gốc, hệ số vùng, phép đẩy ra khỏi vật cản, kẹp vào
 * vùng chơi được — đều phải nằm trong đây, không được rải rác ở chỗ khác.
 */
export function moveStep(
  cfg: GameConfig,
  map: Pick<GameMap, 'obstacles' | 'zones' | 'size'> & { play_margin?: number },
  from: Vec,
  angle: number,
  scale: number,
  dt: number,
  broken: (id: number) => boolean,
  speedMult = 1,
  vel: Vec = { x: 0, y: 0 },
  moving = true,
  ghost = false,
  /** stalls: dãy sạp tạm của Cô Ba đang chặn NGƯỜI NÀY (spec v6.10 §6.3). */
  stalls: Shape[] = [],
): { pos: Vec; vel: Vec } {
  const r = cfg.player.body_r * scale;
  // Gom Đơn Hỏa Tốc (spec v6.9 §6.3): xuyên MỌI vật cản, chỉ còn **biên khu** chặn. Phải
  // có ở đây chứ không chỉ ở server — client dự đoán bị tường chặn trong khi server cho đi
  // xuyên là lệch vị trí mỗi tick, đúng loại lỗi giật mà §8.2a vừa sửa xong.
  //
  // `clampPlay` một mình KHÔNG đủ: nó chỉ kẹp vào rìa bản đồ, còn biên khu là vật cản loại
  // `border` nằm bên trong. Bản đầu chỉ kẹp rìa, nên server đẩy Shipper ra khỏi biên khu mà
  // client cho đi xuyên — lệch đúng bề dày bức tường.
  const settle = (p: Vec): Vec => {
    let out = ghost
      ? resolveBlocking(p, r, map, broken, (o) => o.t >= TIER_BORDER)
      : resolveBlocking(p, r, map, broken);
    // Sạp xét SAU vật cản tĩnh, cùng thứ tự với `Game.resolveBlockingFor`: ai bị kẹp giữa
    // sạp và tường thì được đẩy ra khỏi sạp, vì sạp mới là thứ biến mất sau 6 giây.
    for (let iter = 0; iter < MAX_ITER && stalls.length > 0; iter++) {
      let moved = false;
      for (const sh of stalls) {
        const np = pushOut(sh, out, r);
        if (np) {
          out = np;
          moved = true;
        }
      }
      if (!moved) {
        break;
      }
    }
    return out;
  };
  if (onIce(map, from)) {
    const step = iceStep(cfg, from, vel, angle, moving, dt, speedMult);
    return { pos: settle(step.pos), vel: step.vel };
  }
  if (!moving) {
    return { pos: settle(from), vel: { x: 0, y: 0 } };
  }
  const dir = { x: Math.cos(angle), y: Math.sin(angle) };
  const speed = cfg.player.move_base * speedMult * zoneMult(cfg, map, from, dir);
  const to = { x: from.x + dir.x * speed * dt, y: from.y + dir.y * speed * dt };
  return { pos: settle(to), vel: { x: 0, y: 0 } };
}
