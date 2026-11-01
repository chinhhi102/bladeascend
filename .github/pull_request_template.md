## Thay đổi gì

<!-- Một hai câu: PR này làm gì, vì sao. -->

Liên quan Issue: #

## Loại thay đổi

- [ ] Sửa lỗi
- [ ] Tính năng mới (đã bàn trong Issue trước)
- [ ] Luật chơi / cân bằng (ghi mục GAME_SPEC: §…)
- [ ] Giao diện / chữ hiển thị
- [ ] Art / âm thanh
- [ ] Tài liệu / công cụ / CI

## Đã kiểm tra

<!-- Đánh dấu [x] những gì đã chạy và QUA trên máy bạn. CI sẽ chạy lại. -->

- [ ] `cd server && go vet ./...`
- [ ] `cd server && go test -race ./...`
- [ ] `cd server && go run ./cmd/mapgen --check` (nếu đụng bản đồ / vật cản / `maps/`)
- [ ] `cd client && npm run build`
- [ ] `node tools/check-privacy.mjs`
- [ ] Đã chơi thử trên trình duyệt (ghi PC / điện thoại, trình duyệt nào):

## Quy tắc dự án

- [ ] Luật chơi chỉ chạy ở server (`server/internal/game`), luật mới có unit test
- [ ] Con số cân bằng lấy từ `config/game.json`, không hardcode
- [ ] Chữ cho người chơi bằng tiếng Việt, nằm trong `client/src/ui/strings.ts`
- [ ] Không thêm thư viện mới (hoặc đã được duyệt trong Issue)
- [ ] Không sửa gì trong `docs/_raw/`
- [ ] Không có mật khẩu, token, tên/biệt danh người chơi thật, `guest_id` đầy đủ, IP

## Ảnh / video (nếu đổi giao diện)
