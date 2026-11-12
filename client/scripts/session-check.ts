// Mã phiên vào lại trận — GAME_SPEC §7.1 (v6.53).
//
//   npm run check:session
//
// Kiểm bốn điều, tất cả đều đã từng là lỗi thật ở cơ chế `pendingRejoin` cũ:
//   1. Lưu rồi đọc lại được, kèm tên + nhân vật (gói `join` cần chúng để qua bộ lọc tên).
//   2. Quá hạn thì TRẢ NULL và tự xoá — không gửi một mã từ nửa tiếng trước.
//   3. `clearSession` xoá cả bản trong bộ nhớ, không chỉ sessionStorage: ván đã kết thúc
//      mà bản nhớ còn thì một lần nối lại sẽ cố giành lại nhân vật đã chết.
//   4. sessionStorage ném lỗi (ẩn danh / iOS chặn) → vẫn chạy bằng bản dự phòng trong bộ
//      nhớ. Spec ghi thẳng yêu cầu này.

/** Node không có sessionStorage. Dựng một bản tối thiểu, có công tắc để giả chế độ ẩn danh. */
const store = new Map<string, string>();
let blocked = false;
(globalThis as unknown as { sessionStorage: unknown }).sessionStorage = {
  getItem: (k: string): string | null => {
    if (blocked) {
      throw new Error('ẩn danh');
    }
    return store.get(k) ?? null;
  },
  setItem: (k: string, v: string): void => {
    if (blocked) {
      throw new Error('ẩn danh');
    }
    store.set(k, v);
  },
  removeItem: (k: string): void => {
    if (blocked) {
      throw new Error('ẩn danh');
    }
    store.delete(k);
  },
};

const { saveSession, loadSession, clearSession } = await import('../src/net/session.ts');

let fails = 0;
function expect(name: string, got: unknown, want: unknown): void {
  const ok = got === want;
  if (!ok) {
    fails++;
  }
  console.log(`${ok ? 'đạt ' : 'TRƯỢT'}  ${name}: ${String(got)}${ok ? '' : ` (mong ${String(want)})`}`);
}

const KEY = 'a'.repeat(64);
const P = { name: 'Tèo', char: 'a_hon' };

// 1. Lưu → đọc lại
saveSession(KEY, P);
expect('đọc lại đúng mã', loadSession()?.key, KEY);
expect('đọc lại đúng tên', loadSession()?.name, 'Tèo');
expect('đọc lại đúng nhân vật', loadSession()?.char, 'a_hon');

// 2. Quá hạn (120 s) → null, và tự xoá
const raw = JSON.parse(store.get('kx.session') as string) as { at: number };
raw.at -= 121;
store.set('kx.session', JSON.stringify({ ...raw, key: KEY, name: 'Tèo', char: 'a_hon' }));
expect('mã quá 120 s → không dùng nữa', loadSession(), null);
expect('và đã bị xoá khỏi sessionStorage', store.has('kx.session'), false);

// 3. clearSession xoá cả bản trong bộ nhớ
saveSession(KEY, P);
clearSession();
expect('clearSession xoá sạch', loadSession(), null);

// 4. sessionStorage bị chặn → vẫn chạy bằng bộ nhớ
blocked = true;
saveSession(KEY, P);
expect('ẩn danh: vẫn nhớ được mã', loadSession()?.key, KEY);
clearSession();
expect('ẩn danh: vẫn xoá được', loadSession(), null);
blocked = false;

// Mã rỗng không bao giờ được lưu: gửi `rs: ""` là tín hiệu "người mới", gửi nhầm thì server
// phải quét bảng tra cho một chuỗi rỗng.
saveSession('', P);
expect('mã rỗng không được lưu', loadSession(), null);

if (fails > 0) {
  console.error(`\n${fails} điều TRƯỢT`);
  process.exit(1);
}
console.log('\nMã phiên vào lại trận (§7.1): đạt hết');
