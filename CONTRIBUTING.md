# Đóng góp cho Đạo Hữu Xin Dừng Bước

> **English summary.** Thanks for helping! Fork the repo, create a branch, make a small
> focused change, and open a Pull Request against `main`. Before opening it, run
> `cd server && go vet ./... && go test -race ./... && go run ./cmd/mapgen --check`,
> `cd client && npm run build`, and `node tools/check-privacy.mjs` — CI runs the same checks.
> The game is **server-authoritative**: all game rules live in `server/internal/game` and
> need unit tests; every balance number comes from `config/game.json`. The source of truth
> for game rules is `docs/_raw/GAME_SPEC.md` (only the owner edits it — if your change
> conflicts with the spec, open an Issue to discuss first). Project rules are in
> [CLAUDE.md](CLAUDE.md). Player-facing text is Vietnamese and lives in
> `client/src/ui/strings.ts`. Never commit real player data (names, full `guest_id`, IPs,
> feedback text). Discussions and docs are mostly in Vietnamese; English is welcome.

Cảm ơn bạn đã muốn góp tay! Tài liệu này nói ngắn gọn: làm thế nào để thay đổi của bạn được
nhận nhanh nhất.

## 1. Đọc trước

- [`docs/_raw/GAME_SPEC.md`](docs/_raw/GAME_SPEC.md) — luật chơi và kiến trúc. **Đây là nguồn
  sự thật.** Chỉ chủ dự án sửa tệp này (và mọi thứ trong `docs/_raw/`).
- [`CLAUDE.md`](CLAUDE.md) — quy tắc bắt buộc của dự án (cả người lẫn AI đều theo).
- [`docs/_raw/ART_PROMPTS.md`](docs/_raw/ART_PROMPTS.md) — quy chuẩn asset (kích thước,
  pivot, tên tệp) nếu bạn đóng góp art.

Thấy code và spec mâu thuẫn, hoặc spec thiếu? **Đừng tự đoán.** Mở một Issue mô tả chỗ
mâu thuẫn và các phương án — chủ dự án sẽ quyết và tự sửa spec.

## 2. Quy trình Pull Request

1. **Fork** repo và tạo nhánh mới từ `main` (`git checkout -b sua-loi-cong-dich-chuyen`).
2. Làm **một thay đổi nhỏ, một mục đích**. PR nhỏ được xem nhanh hơn nhiều.
3. Chạy đủ các bài kiểm ở mục 3 trên máy mình — tất cả phải qua.
4. Commit với lời nhắn rõ ràng (tiếng Việt hoặc tiếng Anh đều được).
5. Mở PR vào `main`, điền mẫu PR (danh sách kiểm tra có sẵn).
6. CI sẽ chạy lại các bài kiểm. CI không dùng secret nào, nên PR từ fork chạy bình thường.
7. Chủ dự án xem và phản hồi. Có thể sẽ nhờ bạn sửa thêm — chuyện bình thường.

Thay đổi lớn (luật chơi mới, thư viện mới, đổi giao thức mạng)? **Mở Issue bàn trước** khi
viết code, để khỏi phí công.

## 3. Bài kiểm phải qua trước khi mở PR

```bash
# Server
cd server
go vet ./...
go test -race ./...
go run ./cmd/mapgen --check      # bản đồ đi lọt ở cả 14 cảnh giới (§8.1b)
cd ..

# Client (lần đầu: cd tools/asset-pipeline && npm ci; cd client && npm ci)
cd client
npm run build                    # tự chạy asset-pipeline + ngân sách + tsc + vite
cd ..

# Không có dữ liệu người chơi thật trong docs/
node tools/check-privacy.mjs
```

## 4. Quy tắc chính (tóm tắt từ CLAUDE.md)

- **Server authoritative:** mọi luật chơi chạy trên server (`server/internal/game`). Client chỉ
  gửi input và vẽ.
- **Không hardcode con số cân bằng:** lấy từ `config/game.json`.
- **Luật mới phải có unit test** (đặc biệt: EXP, scale, bể kiếm, rớt cảnh giới, buff, vật cản
  mọc lại).
- **Chữ hiển thị cho người chơi bằng tiếng Việt**, gom trong `client/src/ui/strings.ts`.
- **Không thêm thư viện/framework mới** khi chưa hỏi trong Issue.
- **Asset:** trong code luôn dùng đường dẫn logic (`ui/medal_gold.png`) qua `assets.url(...)`,
  không ghép tay `/assets/...`. Chưa có art thì dùng placeholder nhưng giữ đúng tên tệp.
- **Không bao giờ xoá art đã sinh** — chuyển vào `assets/library/` (xem CLAUDE.md quy tắc 13).
- **Không commit mật khẩu, token, khoá API.**
- **Không commit dữ liệu người chơi thật:** tên/biệt danh, `guest_id` đầy đủ, IP, nội dung góp
  ý. Ví dụ trong tài liệu thì dùng "Người chơi A" và id cắt 6 ký tự.

## 5. Báo lỗi, đề xuất

- Lỗi game, đề xuất tính năng: mở **Issue** theo mẫu có sẵn.
- **Lỗ hổng bảo mật:** KHÔNG mở Issue công khai — xem [SECURITY.md](SECURITY.md).

Mọi người tham gia đều theo [Quy tắc ứng xử](CODE_OF_CONDUCT.md).
