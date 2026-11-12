# Bộ dọn dữ liệu — chạy thử lần đầu (ADMIN_SPEC §13)

Chủ dự án yêu cầu: *"Chạy dry-run lần đầu → gửi tôi danh sách sẽ xoá trước khi chạy thật."*
Đây là danh sách đó. **Chưa xoá một dòng nào.**

Chạy lại bất cứ lúc nào:

```bash
cd server && go run ./cmd/server -janitor-dryrun -addr 127.0.0.1:9995
```

hoặc mở trang quản trị → **Dọn dữ liệu** (bản xem trước chạy sẵn khi mở trang).

---

## Kết quả ngày 12/11/2026, trên DB thật của `:9999`

```
-- dung lượng theo tệp --
  leaderboard.db         1.2 MB
  telemetry.db           6.4 MB
  metrics.db             1.6 MB
  ảnh góp ý                4 KB
                       ────────
  tổng                   9.2 MB   (ngưỡng cảnh báo: 1 GB)

-- dung lượng theo bảng (ước lượng) --
  server_samples               28 680 dòng      2.5 MB
  samples                       5 823 dòng      1.2 MB
  admin_audit                   5 328 dòng      520 KB
  metrics_minute                2 543 dòng      298 KB
  events                          597 dòng      233 KB
  lives                         1 059 dòng      165 KB
  sessions                        384 dòng      158 KB
  metrics_hour                    196 dòng       23 KB
  hidden_lives                     42 dòng        4 KB
  deploy_marks                     10 dòng        1 KB
  feedback                          1 dòng        1 KB
  admin_login_attempts              4 dòng        0 KB
  admin_sessions                    1 dòng        0 KB

-- SẼ XOÁ --
  Mẫu telemetry client (10 s/lần)                giữ 14 ngày               0 dòng
  Mẫu telemetry server (10 s/lần)                giữ 14 ngày               0 dòng
  Phiên telemetry không còn mẫu                  giữ 14 ngày               0 dòng
  Dòng lỗi thô (chữ ký đã gộp, giữ mãi)          giữ 30 ngày               0 dòng
  Số liệu hiệu năng theo PHÚT                    giữ 14 ngày               0 dòng
  Số liệu hiệu năng theo GIỜ                     giữ 180 ngày              0 dòng
  Mạng chơi cũ (giữ mãi mạng trên bảng xếp hạng) giữ 90 ngày               0 dòng
  Phiên đăng nhập quản trị đã hết hạn            giữ 30 ngày               0 dòng
  Nhật ký lần đăng nhập quản trị                 giữ 30 ngày               0 dòng
  Lệnh cấm đã hết hạn                            giữ 30 ngày sau khi hết   0 dòng
  Nhật ký thao tác quản trị                      giữ 1 năm                 0 dòng

TỔNG: 0 dòng.
```

## Đọc kết quả

**Chưa có gì để xoá, và điều đó đúng.** Dữ liệu cũ nhất trong DB mới khoảng ba tuần
(`server_samples` từ 30/10), trong khi hạn giữ ngắn nhất là 14 ngày cho mẫu telemetry thô —
nhưng 14 ngày đó tính từ **hôm nay ngược về**, nên lô đầu tiên chỉ tới hạn khoảng **26/11**.

Bảng to nhất là `server_samples` (28 680 dòng / 2,5 MB): mỗi 10 giây một dòng cho mỗi khu,
tức ~43 000 dòng mỗi ngày khi có người chơi. Đây sẽ là thứ bộ dọn cắt nhiều nhất, và nó
không mất thông tin — số liệu đã được cuộn vào `metrics_minute/hour/day` (§12).

## Khi nào nó bắt đầu xoá thật

**Chưa bao giờ, cho tới khi chủ dự án bấm.** Job 04:30 hằng đêm hiện chỉ **chạy thử** và ghi
danh sách vào Nhật ký quản trị (hành động `janitor_dryrun`). Bấm **"Dọn ngay"** một lần ở
trang *Dọn dữ liệu* là vừa dọn ngay vừa bật chế độ xoá thật cho các đêm sau.

Lý do có thêm cái chốt này (spec không đòi): một bộ xoá tự động chạy lần đầu lúc 4 giờ sáng,
trên dữ liệu thật, mà chưa ai nhìn qua danh sách — là cách tốt nhất để mất thứ không lấy
lại được.

## Những thứ bộ dọn KHÔNG BAO GIỜ đụng

- Art, `assets/library/` (quy tắc 13 của CLAUDE.md) — nó không đọc thư mục art.
- Mạng chơi đang nằm trên bảng xếp hạng toàn cục: top 50 của **cả ba bảng** (Uy Danh, số
  chém, Linh khí đỉnh) được giữ mãi, dù bao nhiêu tuổi.
- Mạng chơi đã bị ẩn (`hidden_lives`) — đó là bằng chứng, không phải rác.
- Lượt thích nhân vật.
- **Chữ** góp ý (chỉ ảnh của thư đã khép quá 90 ngày mới bị dọn — Q28, đã chạy từ trước).
- Chữ ký lỗi (`error_sigs`) và thống kê mạng chơi theo ngày (`lives_rollup`): cả hai được
  **gộp lại trước khi xoá dòng thô** và giữ mãi.
