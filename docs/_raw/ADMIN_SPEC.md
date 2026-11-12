# ADMIN CONSOLE — Spec v1.6

**v1.6:** bộ dọn dữ liệu hằng ngày theo chính sách giữ (§13).

**v1.5:** bộ thu gom số liệu phút/giờ/ngày (min/max/avg/p95) + trang Hiệu năng (§12).

**v1.4:** cấu hình trần phòng / sàn khu / ngưỡng tự tắt bot ngay trên trang admin (§4).

**v1.3:** điều khiển bot toàn server + từng phòng (§4).

**v1.2:** bản đồ admin = bản đồ thật trong game để xem toàn cảnh (bỏ chấm người chơi live); giao diện admin dùng Ant Design; góp ý cho phép tải ảnh lên (§5, §2.1, §7).

**v1.1:** `?debug=true` chỉ hoạt động khi đã đăng nhập admin (§1.1).

> Đặt tại `docs/_raw/ADMIN_SPEC.md`. Dùng chung stack hiện tại (Go server + Vite/TypeScript). **Riêng trang admin** được dùng React + Ant Design (§2.1, bundle tách khỏi game); game không thêm framework UI. Mọi con số cấu hình nằm trong `config/game.json` → khối `admin`.

## 0. Mục tiêu
Một trang quản trị chỉ chủ dự án vào được, để:
1. Xem **bố cục bản đồ** từng khu, kể cả xem trực tiếp người đang chơi trong khu đó.
2. Xem **số liệu hậu trường** từng phòng / khu.
3. **Dashboard live** (chỉ số, không cần biểu đồ), **danh sách người chơi thật**, **kick / cấm thủ công** khi nghi hack (không bao giờ tự động kick).
4. Đọc **hòm thư góp ý** người chơi gửi từ trong game.
5. Toàn bộ phải **đăng nhập** mới xem được.

---

## 1. Đăng nhập & bảo mật (làm trước tiên)

| Mục | Quy định |
|---|---|
| Tài khoản | 1 tài khoản admin. Mật khẩu **không** nằm trong code/config: server đọc `ADMIN_PASSWORD_HASH` (bcrypt, cost ≥ 12) từ biến môi trường. Có lệnh `go run ./cmd/adminpass` để tạo hash từ mật khẩu nhập vào. |
| Phiên | Đăng nhập đúng → cookie phiên ngẫu nhiên 32 byte, `HttpOnly; Secure; SameSite=Strict; Path=/`, hết hạn 12 giờ, lưu ở bảng `admin_sessions`. Đăng xuất xoá phiên. |
| Chống dò mật khẩu | Tối đa 5 lần sai / 15 phút / IP, sau đó khoá IP 15 phút. Ghi mọi lần đăng nhập (thành công/thất bại) vào nhật ký. |
| Đường dẫn | `/admin/login` công khai. Mọi `/admin/*` khác và mọi `/api/admin/*` trả 401 nếu không có phiên hợp lệ. **HTML và JS của trang admin chỉ phục vụ sau khi đăng nhập** (không để mã admin công khai). |
| CSRF | Ngoài SameSite=Strict, mọi request thay đổi dữ liệu (POST/DELETE) phải có header `X-Admin-Token` (lấy từ `/api/admin/me`). |
| Giới hạn IP (tuỳ chọn) | Nếu đặt `ADMIN_IP_ALLOWLIST` thì chỉ IP trong danh sách mới vào được. |
| HTTPS | Bắt buộc khi deploy (đã có Caddy). Trên HTTP thường thì không cho đăng nhập (trừ `localhost` để phát triển). |
| Nhật ký thao tác | Mọi hành động admin (đăng nhập, kick, cấm, gỡ cấm, đổi trạng thái góp ý) ghi vào bảng `admin_audit`: thời gian, hành động, đối tượng, lý do, IP admin. Có trang xem nhật ký (§7). |
| XSS | Nội dung góp ý và tên người chơi là dữ liệu do người chơi nhập → **luôn hiển thị dạng text** (`textContent`), không bao giờ `innerHTML`. Có test chèn `<script>` vào góp ý và tên → không chạy. |
| **Cờ debug trong game** | `?debug=true` **chỉ có tác dụng khi trình duyệt đang có phiên admin hợp lệ** (đã đăng nhập ở `/admin/login`). Xem §1.1. |

Về sau có thể thêm 2 lớp (TOTP 2FA) — chưa cần ở v1.

### 1.1 Cờ debug gắn với đăng nhập admin
- Người chơi mở game với `?debug=true` mà **không có phiên admin** → game chạy như bình thường, **không** bật phím debug, F3, H, B…; không báo lỗi gì (không lộ là có chế độ debug).
- Có phiên admin (cùng tên miền, cookie gửi kèm khi mở WebSocket game) → server đánh dấu kết nối đó là `debug_allowed` và báo cho client; client **chỉ bật** phím debug khi server xác nhận. Tham số trên URL một mình không bao giờ đủ.
- Server **từ chối mọi lệnh debug** (+EXP, lên cảnh giới, +kiếm, buff, tự chết, thả bot, nạp Tiên khí…) từ kết nối không có `debug_allowed` — kể cả khi client bị sửa để gửi lệnh. Lệnh bị từ chối tính vào cờ nghi hack "Hành động bị từ chối" (§6.1).
- Mỗi lệnh debug được dùng → ghi `admin_audit` (lệnh, phòng/khu, nhân vật).
- Mạng chơi có dùng lệnh debug → gắn nhãn `debug`, **không ghi vào bảng xếp hạng toàn cục** và không tính vào telemetry cân bằng.
- Phiên admin hết hạn giữa trận → tắt quyền debug ngay ở lần kiểm tiếp theo (kiểm lại mỗi 60s).
- **Môi trường phát triển**: khi server chạy với `APP_ENV=dev` trên `localhost`, cho phép debug không cần đăng nhập (để làm việc nhanh). Công cụ tự động (`loadtest`, `browsertest`, `netcode-bench`) dùng biến môi trường `DEBUG_TOKEN` chỉ hợp lệ khi `APP_ENV=dev` hoặc chạy trên server cục bộ; **production không bao giờ nhận `DEBUG_TOKEN`**.

---

## 2. Bố cục trang admin

Thanh bên trái: **Dashboard · Phòng & Khu · Bản đồ · Người chơi · Góp ý (số chưa đọc) · Nhật ký · Đăng xuất**.

### 2.1 Giao diện — dùng Ant Design
- Trang admin là **bundle riêng** (không nằm trong bundle của game, không làm nặng game): **React + Ant Design (antd v5)** — đã duyệt cho riêng trang admin.
- Bố cục `Layout` + `Sider` (thu gọn được) + `Header` (tên build đang chạy, nút đăng xuất) + `Content`. Hỗ trợ chế độ sáng/tối (`ConfigProvider` theme algorithm), mặc định theo máy.
- **Dashboard**: lưới `Card` + `Statistic` (số to, đơn vị, mũi tên tăng/giảm so với 1 phút trước), màu theo ngưỡng (`Tag`/`Badge` xanh/vàng/đỏ), nhóm theo các mục §3. Không biểu đồ phức tạp; tuỳ chọn sparkline nhỏ 10 phút.
- **Mọi bảng dùng `Table` của antd**: header cố định, sắp xếp + lọc theo cột, phân trang (20/50/100), tìm kiếm, cột trạng thái là `Tag` màu, số căn phải, thời gian hiển thị tương đối ("3 phút trước", rê chuột thấy giờ đầy đủ), hàng bấm được → `Drawer` chi tiết bên phải (dùng `Descriptions`).
- Thao tác nguy hiểm (kick, cấm, ẩn bảng xếp hạng): `Modal.confirm` + ô lý do bắt buộc; kết quả báo bằng `message`.
- Chạy được trên điện thoại (Sider tự thu gọn, bảng cuộn ngang).

Dữ liệu live đi qua 1 kết nối **`/api/admin/ws`** (WebSocket, xác thực bằng cookie phiên). Admin **không** được tính là người chơi, không vào phòng, không ảnh hưởng bot fill.

---

## 3. Dashboard live (chỉ số, cập nhật mỗi 2 giây)

| Nhóm | Chỉ số |
|---|---|
| Người chơi | Người thật online · bot · số phòng · chế độ bot toàn server · người vào / phút · người rời / phút · phiên đang chạy theo thiết bị (PC / Android / iPhone) |
| Server | Tick ms TB / p95 **theo từng khu** (tô đỏ nếu > 4 ms) · CPU % · RAM MB · số goroutine · uptime · build đang chạy |
| Mạng | KB/s tổng gửi đi · KB/s TB mỗi client · snapshot TB / lớn nhất (KB) · ping người chơi p50 / p95 · số client đang nghẽn (bỏ snapshot) |
| Client | FPS TB / p10 (từ telemetry) · lỗi client / phút · số lần mất WebGL / phút · thời gian chờ vào trận p50 |
| Trò chơi | Phân bố cảnh giới người thật (bao nhiêu người ở mỗi cảnh giới) · số người Tiên giới · lần chém trúng / phút · lần chết / phút · Tiên khí phun ra / nhặt / phút |
| Hộp thư | Góp ý chưa đọc · cờ nghi hack đang mở |

Mỗi ô có ngưỡng màu (xanh / vàng / đỏ) đọc từ config `admin.thresholds`.

---

## 4. Phòng & Khu (số liệu hậu trường)

Bảng phòng → bấm mở ra 5 khu:

| Cột (mỗi khu) | |
|---|---|
| Người thật / bot / tổng | |
| Tick ms TB / p95 | |
| Cảnh giới cao nhất + tên | |
| Số quái / viên Linh khí / viên Tiên khí / vật phẩm trên đất | |
| Vật cản đang bị chém (chờ mọc lại) | |
| Đài Tiên Khí: giây tới đợt phun lớn tiếp theo; Thiên Tiên Đài: giây tới lần phun | |
| Lượt qua cổng / phút (vào, ra), lượt qua Thiên Môn / phút | |
| Seed bản đồ + phiên bản mapgen | |
| Nút **"Xem bản đồ"** → §5 | |

**Cấu hình phòng** (GAME_SPEC §9.5) — thẻ ở đầu trang Phòng & Khu, chỉnh được ngay:
- Trần người thật/phòng (mặc định production 50, staging 20), sàn nhân vật **từng khu** (production Thiên Đình 30 · khu rìa 20; staging 15 · 10), ngưỡng tự tắt bot (production 40, staging 20).
- Hiện môi trường đang chạy + mức đã qua loadtest; đặt cao hơn → cảnh báo vàng nhưng vẫn cho lưu.
- Lưu → `Modal.confirm` + lý do + audit; áp dụng ngay, sống qua deploy.

**Điều khiển bot** (GAME_SPEC §9.5):
- Ở đầu trang: công tắc **toàn server** Tự động / Bật / Tắt.
- Mỗi phòng: chế độ đang áp dụng (vd "Tự động — đang tắt vì 43 người"), nút đổi riêng cho phòng (Tự động / Bật / Tắt / Theo toàn server), nút **"Rút bot ngay"**.
- Cột người thật / bot / tổng và "giới hạn 50" cho từng phòng.
- Mọi thay đổi: `Modal.confirm` + ghi `admin_audit`.

---

## 5. Xem bản đồ (toàn cảnh bản đồ thật trong game)

**Mục đích**: nhìn toàn cảnh mỗi khu **đúng như người chơi thấy** (art thật, nền, decal, vật cản, cổng, Đài, Thiên Môn) để đánh giá bố cục và thẩm mỹ. **Không** vẽ người chơi trực tiếp (bỏ luồng dữ liệu 5 Hz của bản v1.1 — không tốn tài nguyên server).

- Chọn **phòng → khu** (mặc định: bản đồ seed đang chạy của khu đó) hoặc nhập **khu + seed** bất kỳ để xem bố cục mới trước khi deploy.
- Vẽ bằng **chính bộ vẽ bản đồ của game** (`render/mapview.ts` + decal + ambient + vật cản), chế độ "toàn cảnh": thu nhỏ vừa khung, kéo/phóng to mượt tới mức cận cảnh như trong game. Có nút **Ảnh chụp toàn khu** (xuất PNG độ phân giải cao).
- Lớp phủ bật/tắt (mặc định tắt, để nhìn như người chơi): bậc vật cản (tô màu), lưới 1 000, kết quả `--check` theo cảnh giới r = 0..13 (thanh trượt), vùng (bụi/bùn/băng/nước).
- Số liệu live của khu đặt ở thanh bên cạnh bản đồ (cập nhật 2s): người thật / bot, tick ms, viên Tiên khí trên đất, đếm ngược Đài — **chỉ số, không chấm vị trí**.

## 6. Người chơi (người thật đang online)

Bảng, sắp xếp được, tìm theo tên:

| Cột |
|---|
| Tên · nhân vật · `guest_id` (8 ký tự đầu) |
| Thiết bị / trình duyệt · IP (che 1 phần, vd `113.161.x.x`) · quốc gia (nếu có) |
| Phòng / khu · cảnh giới + tầng · số kiếm · Uy Danh · số lần chém trúng · thời gian sống |
| Ping · FPS · số lần hiệu chỉnh vị trí / phút |
| **Cờ nghi hack** (số + biểu tượng ⚠) |

Bấm vào 1 người → bảng chi tiết: lịch sử phiên gần đây (từ telemetry), chi tiết từng cờ, góp ý người này đã gửi, nút **Xem trên bản đồ**.

### 6.1 Cờ nghi hack (chỉ gắn cờ, **không tự kick**)
Server là bên quyết định luật nên hack khó gây hại trực tiếp; cờ dùng để phát hiện người đang thử phá:

| Cờ | Điều kiện |
|---|---|
| Input bất thường | > 2× tần suất input cho phép trong 10 giây |
| Gói tin lỗi | Gói không giải mã được / sai schema > 5 lần / phút |
| Hành động bị từ chối | Dùng kỹ năng khi không có lượt, nhặt đồ ngoài tầm, phím debug ở production… > 10 lần / phút |
| Nhiều kết nối | > 3 kết nối cùng IP hoặc cùng `guest_id` |
| Tăng điểm bất thường | Uy Danh / Linh khí tăng nhanh hơn 5× trung vị cùng cảnh giới trong 2 phút |
| Ping giả | Pong trả về sai mẫu / thời gian âm |

Cờ hiện ở dashboard và danh sách; mỗi cờ lưu bằng chứng (số liệu lúc gắn cờ). Admin có thể **bỏ cờ** (đánh dấu nhầm).

### 6.2 Thao tác (thủ công, phải nhập lý do)
- **Kick**: ngắt kết nối, người chơi thấy màn hình "Bạn đã bị mời ra khỏi phòng" kèm lý do ngắn. Vào lại được ngay.
- **Kick + cấm**: chọn thời hạn 1 giờ / 24 giờ / 7 ngày / vĩnh viễn; cấm theo `guest_id` **và** IP. Người bị cấm thấy thông báo + thời điểm hết hạn khi vào trang.
- **Gỡ cấm**: ở trang danh sách cấm.
- Hộp xác nhận trước mỗi thao tác. Mọi thao tác ghi `admin_audit`.

---

## 7. Hòm thư góp ý

### 7.1 Phía người chơi
- Nút nhỏ **💬 Góp ý** ở góc dưới (trang chủ và trong trận; trong trận không che HUD, không tạm dừng game).
- Form: **loại** (Báo lỗi · Cân bằng · Ý tưởng · Khác), **nội dung** (5–1 000 ký tự), ô tick **"Kèm ảnh chụp màn hình"** (mặc định bật khi chọn Báo lỗi; chụp canvas, nén WebP ≤ 300 KB).
- **Tải ảnh lên từ máy**: nút "Thêm ảnh" (`<input type="file" accept="image/*" multiple>` — trên iPhone/Android mở thư viện ảnh hoặc máy ảnh), **kéo-thả** và **dán (Ctrl+V)** trên PC. Tối đa **3 ảnh** (tính cả ảnh chụp màn hình), xem trước dạng ô nhỏ có nút xoá. Trình duyệt tự thu nhỏ (cạnh dài ≤ 1 600 px) và nén WebP ≤ 300 KB/ảnh trước khi gửi; nhận JPG/PNG/WebP/HEIC (HEIC không đọc được thì báo "Chưa hỗ trợ định dạng này").
- Server: kiểm định dạng bằng nội dung tệp (không tin đuôi), **mã hoá lại ảnh** → xoá toàn bộ EXIF (ảnh điện thoại có thể chứa toạ độ GPS nhà người chơi), giới hạn 3 ảnh × 300 KB / góp ý.
- Tự đính kèm (người chơi được thấy dòng "Sẽ gửi kèm: phiên bản, thiết bị, khu, cảnh giới, ping, FPS"): build, thiết bị/trình duyệt, kích thước màn hình, phòng/khu, cảnh giới, ping, FPS, `guest_id`, tên, 20 lỗi console gần nhất (nếu là Báo lỗi).
- Gửi xong: "Cảm ơn! Đã nhận góp ý." Giới hạn 3 góp ý / 10 phút / `guest_id`; người bị cấm không gửi được.

### 7.2 Phía admin
- Danh sách góp ý mới nhất trước, lọc theo loại / trạng thái / thiết bị / build; tìm theo nội dung.
- Trạng thái: **Mới · Đã xem · Đang làm · Xong · Bỏ qua** + ghi chú riêng của admin.
- Xem ảnh (chụp màn hình + ảnh tải lên, bấm để phóng to dạng `Image.PreviewGroup`), thông tin đính kèm, bấm sang thông tin người chơi.
- Xuất CSV.
- Số góp ý **Mới** hiện ở thanh bên và dashboard.

---

## 8. Dữ liệu (SQLite, cùng file với leaderboard)

| Bảng | Cột chính |
|---|---|
| `admin_sessions` | id (hash), tạo lúc, hết hạn, IP, user agent |
| `admin_audit` | thời gian, hành động, đối tượng, lý do, IP admin, chi tiết JSON |
| `admin_login_attempts` | thời gian, IP, thành công |
| `bans` | guest_id, IP, lý do, từ, đến (null = vĩnh viễn), admin tạo |
| `cheat_flags` | thời gian, guest_id, loại cờ, bằng chứng JSON, đã xử lý |
| `feedback` | id, thời gian, guest_id, tên, loại, nội dung, ngữ cảnh JSON, đường dẫn ảnh, trạng thái, ghi chú admin |

Ảnh góp ý lưu ở thư mục `data/feedback/` (ngoài thư mục phục vụ công khai), chỉ đọc được qua `/api/admin/feedback/{id}/image` khi đã đăng nhập. Kiểm định dạng ảnh thật (không tin đuôi file).

---

## 9. API (tóm tắt)

| Phương thức | Đường dẫn | |
|---|---|---|
| POST | `/api/admin/login` · `/api/admin/logout` | |
| GET | `/api/admin/me` | trả CSRF token |
| WS | `/api/admin/ws` | kênh live: dashboard 2s, phòng/khu 2s (chỉ số, không gửi vị trí người chơi) |
| GET | `/api/admin/rooms` · `/api/admin/players` · `/api/admin/players/{guest_id}` | |
| POST | `/api/admin/players/{guest_id}/kick` · `/ban` | lý do bắt buộc |
| DELETE | `/api/admin/bans/{id}` | |
| GET | `/api/admin/maps/{zone}?seed=` | bố cục + kết quả check |
| GET/PATCH | `/api/admin/feedback` · `/api/admin/feedback/{id}` | |
| GET | `/api/admin/audit` | |
| POST | `/api/feedback` | **công khai** (người chơi gửi), có rate limit |

---

## 10. Thứ tự làm

| Bước | Nội dung | Xong khi |
|---|---|---|
| A1 | Đăng nhập + phiên + chặn dò + audit + khung trang + Dashboard + **khoá cờ debug theo phiên admin (§1.1)** | Vào `/admin` bị đẩy về login; đăng nhập xong thấy dashboard số nhảy mỗi 2s |
| A2 | Phòng & Khu + Người chơi + cờ nghi hack + kick/cấm/gỡ cấm | Kick được 1 tab thử; tab đó thấy thông báo; nhật ký ghi lại |
| A3 | Hòm thư góp ý (nút trong game + trang admin) | Gửi góp ý kèm ảnh từ điện thoại, thấy ở admin |
| A4 | Bản đồ: toàn cảnh bằng bộ vẽ thật của game + số liệu khu | Xem đủ 5 khu như người chơi thấy, xuất được PNG |

## 11. Checklist
- [ ] Không đăng nhập: `/admin`, `/admin/*.js`, `/api/admin/*`, ảnh góp ý → đều 401/redirect.
- [ ] Sai mật khẩu 5 lần → khoá IP 15 phút.
- [ ] Góp ý chứa `<script>alert(1)</script>` và tên người chơi chứa HTML → hiển thị nguyên văn, không chạy.
- [ ] Trang Bản đồ hiện đúng art như trong game, phóng to/kéo mượt, xuất PNG toàn khu; không có luồng vị trí người chơi.
- [ ] Không có đoạn code nào tự kick theo cờ.
- [ ] Production, chưa đăng nhập admin: mở `?debug=true` → không có phím debug, F3 không hiện; gửi lệnh debug thủ công qua WebSocket → bị từ chối + gắn cờ.
- [ ] Đã đăng nhập admin: `?debug=true` bật đủ phím debug; dùng lệnh → có trong nhật ký; mạng đó không lên bảng xếp hạng toàn cục.
- [ ] Gửi góp ý kèm 3 ảnh tải lên từ iPhone (thư viện ảnh) → admin thấy đủ 3 ảnh, ảnh đã mất EXIF/GPS.
- [ ] Mọi bảng admin: sắp xếp, lọc, phân trang, Drawer chi tiết; dashboard dạng Card/Statistic.
- [ ] Người bị cấm 1 giờ: vào trang thấy thông báo + giờ hết hạn; sau 1 giờ vào lại được.

---

## 12. Bộ thu gom số liệu hiệu năng (v1.5)

**Mục tiêu**: xem hiệu năng game theo **phút / giờ / ngày**, mỗi khoảng có **min / max / avg / p95** của mọi chỉ số, để thấy xu hướng và gắn được sự cố với bản deploy hay với lúc máy bận.

### 12.1 Thu những gì (mỗi chỉ số kèm "phạm vi": toàn server / phòng / khu / loại thiết bị)

| Nhóm | Chỉ số |
|---|---|
| Server | tick ms (theo **khu**), thời gian từng hệ thống trong tick (va chạm, bot AI, kỹ năng, AOI, gửi gói), GC pause ms, heap MB, goroutine, CPU % **của tiến trình server** |
| Máy chủ | **load average 1 phút, CPU % toàn máy, RAM trống** — để phân biệt "game chậm" với "máy đang bận việc khác" |
| Phòng / khu | người thật, bot, quái, viên EXP trên đất, vật cản đang chờ mọc, lượt kỹ năng/phút, lượt chết/phút |
| Mạng | KB/s gửi đi (tổng + TB mỗi client), kích thước snapshot, số client đang nghẽn, ping p50/p95 (từ client) |
| Client (telemetry) | FPS p50 / p1, ms/khung, ms từng hệ thống vẽ, bộ nhớ GPU ước tính, lỗi JS — tách theo **PC / Android / iPhone** và **mức chất lượng** |
| Người chơi | vào/ra/phút, thời gian sống, cảnh giới cao nhất mỗi mạng, tỉ lệ chọn nhân vật |

### 12.2 Cách gom
- Server lấy mẫu mỗi **1 s** (tick thì gom mọi tick), giữ trong bộ nhớ; **hết mỗi phút** ghi 1 dòng/chỉ số/phạm vi: `count, min, max, sum (→ avg), p95` (p95 bằng histogram cố định hoặc t-digest — không lưu từng mẫu).
- **Chỉ ghi phút có ít nhất 1 người thật online.** Phút không có người thì không ghi (bot chạy một mình không tính hiệu năng).
- Cuộn lên: **giờ** = gộp các phút (min của min, max của max, avg theo trọng số count, p95 gộp từ histogram); **ngày** = gộp các giờ. Chạy mỗi khi hết giờ / hết ngày.
- Giữ dữ liệu: phút **14 ngày**, giờ **180 ngày**, ngày **giữ mãi**.
- Mỗi dòng gắn **build id** đang chạy; deploy → ghi 1 mốc "deploy" để vẽ vạch dọc trên biểu đồ.
- Chi phí: thu gom không làm tick tăng quá **0,1 ms** (đo bằng bench trước/sau). Ghi SQLite theo lô 1 lần/phút, không ghi mỗi giây.
- Bảng SQLite: `metrics_minute`, `metrics_hour`, `metrics_day` (`ts, metric, scope, build, count, min, max, sum, p95`) + `deploy_marks`.

### 12.3 Trang "Hiệu năng" trong admin (antd)
- Chọn **chỉ số**, **phạm vi** (khu / thiết bị…), **độ phân giải** (phút / giờ / ngày) và **khoảng thời gian** (24 giờ, 7 ngày, 30 ngày, tuỳ chọn).
- Biểu đồ đường: dải min–max tô mờ + đường avg + đường p95 + **đường ngưỡng ngân sách** (vd tick 4 ms) + vạch dọc ở các lần deploy.
- Đặt **2 chỉ số chồng lên nhau** để so (vd tick Thiên Đình với load máy, hoặc tick với số bot) — trục phải riêng.
- Bảng dưới biểu đồ: từng khoảng với min / max / avg / p95, sắp xếp được; **xuất CSV**.
- Thẻ tóm tắt đầu trang: 24 giờ qua — giờ tệ nhất của từng chỉ số chính, kèm số người thật lúc đó.
- Cảnh báo dashboard (§3) đọc từ đây: ngưỡng vượt liên tục ≥ 5 phút → thẻ cảnh báo, kèm gợi ý (vd "load máy 3,2 — có thể do máy bận, không phải game").

### 12.4 Checklist
- [ ] 1 người vào chơi 10 phút → có 10 dòng phút cho mọi chỉ số; rời đi 5 phút → không có dòng nào trong 5 phút đó.
- [ ] Hết giờ → dòng giờ khớp gộp từ các phút (min/max đúng, avg theo trọng số).
- [ ] Biểu đồ tick Thiên Đình + load máy chồng nhau, có vạch deploy; xuất CSV được.
- [ ] Bench: thu gom làm tick tăng ≤ 0,1 ms.

---

## 13. Bộ dọn dữ liệu (janitor) (v1.6)

Dữ liệu chỉ được giữ khi còn dùng vào việc gì. Job dọn chạy **mỗi ngày 04:30**, ghi kết quả vào `admin_audit` + trang "Dọn dữ liệu".

| Dữ liệu | Giữ | Sau đó |
|---|---|---|
| Mẫu telemetry thô (10 s/lần, client + server) | 14 ngày | đã cuộn vào metrics phút/giờ/ngày (§12) → xoá thô |
| Metrics phút / giờ / ngày | 14 ngày / 180 ngày / mãi | theo §12 |
| Mạng chơi (`lives`) | 90 ngày | giữ mãi các mạng đang nằm trên bảng xếp hạng toàn cục (tuần/tháng/mọi thời đại) + thống kê gộp; mạng thường quá 90 ngày → xoá |
| Lỗi client/server | 30 ngày | gộp theo chữ ký (số lần, build đầu/cuối) giữ mãi |
| Ảnh góp ý | theo Q28 (thư Xong/Bỏ qua > 90 ngày) | chữ góp ý giữ mãi |
| Phiên admin hết hạn, lần đăng nhập | 30 ngày | xoá |
| Cấm đã hết hạn | 30 ngày sau khi hết | xoá |
| Cờ nghi hack đã xử lý | 90 ngày | xoá |
| Nhật ký admin (`admin_audit`) | 1 năm | xoá |
| Mã phiên (§7.1) hết hạn | 1 giờ | xoá |
| Log file server | xoay vòng 50 MB × 5 tệp | xoá tệp cũ nhất |
| Image Docker | đã có cron prune 7 ngày | — |

- Sau khi dọn: `VACUUM` SQLite mỗi tuần (giờ vắng nhất); không chạy khi có > 0 người thật online.
- **Không bao giờ đụng** art / `assets/library` (luật 13), bảng xếp hạng đang hiện, lượt thích, góp ý (chữ).
- Trang **"Dọn dữ liệu"**: dung lượng từng bảng + tệp DB + thư mục ảnh/log, lần chạy gần nhất (đã xoá bao nhiêu dòng/MB), nút **chạy thử (dry-run)** liệt kê sẽ xoá gì trước khi xoá thật.
- Cảnh báo dashboard khi DB > 1 GB hoặc tăng > 20%/tuần.
- Checklist: dry-run in đúng số dòng sẽ xoá; chạy thật xoá đúng số đó; mạng trên bảng xếp hạng toàn cục còn nguyên; DB giảm dung lượng sau VACUUM.
