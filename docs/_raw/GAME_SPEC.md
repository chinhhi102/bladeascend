# BLADEASCEND · ĐẠO HỮU XIN DỪNG BƯỚC — Game Design & Technical Spec (v2)

> Tên tạm (working title). File này dành cho Claude Code: đặt tại `docs/GAME_SPEC.md`.
> Mọi con số nằm trong `config/game.json` (mục 16) — **không hardcode**.
> Art/animation prompts: `docs/ART_PROMPTS.md`. Hướng dẫn setup: `docs/SETUP_GUIDE.md`.

**Thay đổi v6.33:** mỗi nhân vật có nội tại riêng ±15% cho tốc chạy / bán kính nhặt / tốc quay / độ bền (tổng = 0), cộng trên base (§6.4).

**Thay đổi v6.32:** luật chung mọi thưởng/phạt CỘNG trên base (bán kính nhặt, tốc chạy, tốc quay, độ bền, EXP) — không nhân dồn; Shipper +0.1 cộng dồn với Nam Châm và cảnh giới (§4.0).

**Thay đổi v6.31:** Shipper Gom Đơn: bán kính nhặt đồ chỉ +10% base (bỏ hút 3 U) (§6.3).

**Thay đổi v6.30:** bỏ hàng rào cân bằng theo tỉ lệ thắng tay đôi (cân bằng theo số liệu người chơi thật); Bà Tám toả tròn 360°; Ông Ba Võng thêm Trộm Kiếm 4s (kiếm không mẻ, kiếm địch bể thành của ông); sửa ví dụ Uy Danh 850 (§6.3, §11.1, trả lời Q36).

**Thay đổi v6.29:** chết = hết ván: ra màn hình Tổng kết, không hồi sinh, không tự vào lại; phải bấm "Chơi tiếp" (§5.1, §7).

**Thay đổi v6.28:** trần người thật/phòng, sàn 10 nhân vật/khu, ngưỡng tự tắt bot chỉnh ở admin; mặc định production 50/10/40, staging 20/10/20; bot lấp theo khu (trả lời Q34, Q35) (§9.1, §9.3, §9.5).

**Thay đổi v6.27:** khống chế không bao giờ vô hiệu hoá vòng kiếm (Ru Ngủ: kiếm vẫn đánh trả; hồi 25s ×2); Thằng Mập bịt mắt: kiếm người mù không chém trúng thân ai trong 5s (hồi 20s ×2) (§6.3).

**Thay đổi v6.26:** chém trúng thân = CHẾT ở mọi cảnh giới (bỏ rớt cảnh giới); người chém nhận 30% Linh khí nạn nhân (trần +2 đại cảnh giới, bot 15%), rơi 30%, mất 40%; Tiên khí: nhận 30% (trần +1 cảnh giới Tiên), rơi 30% (§3.4, §5.1).

**Thay đổi v6.25:** tốc độ cộng trên base + kẹp trần ×2 (Gom Đơn ×3), không nhân dồn (§4.1); kỹ năng tự hồi + tích lượt theo từng nhân vật, bỏ vật phẩm Bí Kíp; giảm Đại Ca (30%, 8s) và Bé Tí (4s); U kỹ năng giữ 90×scale (§6.3).

**Thay đổi v6.24:** §6.0 ghi nguyên nhân thật của lạm phát EXP (lỗi Bí Kíp cộng cả tu vi), bảng nhịp là chuẩn, hệ số chỉnh theo sim (Q33).

**Thay đổi v6.23:** art đã sinh không bao giờ xoá — chuyển vào thư viện `assets/library/`, không build, không tải cho người chơi (§13.7 mục 3b).

**Thay đổi v6.22:** bot sinh như người chơi mới (Luyện Khí 1); phòng tối đa 50 người thật; bật/tắt bot ở admin (toàn server + từng phòng), tự tắt khi > 40 người, bật lại < 35 (§9.1, §9.3, §9.5).

**Thay đổi v6.21:** sửa lạm phát EXP — viên EXP/quái × 1.5^cảnh giới, bán kính hút × √scale, Nam Châm × 1.5, cướp EXP từ bot 25%, bảng nhịp mục tiêu + kiểm CI (§6.0).

**Thay đổi v6.20:** chất lượng mặc định Vừa (PC GPU rời) / Thấp (điện thoại, GPU tích hợp), tự nâng Thấp→Vừa sau 10s, Vừa→Cao sau 20s ổn định ≥ 55 FPS, chỉ lưu lựa chọn tay (§13.6 mục 8, trả lời Q32).

**Thay đổi v6.19:** bảng xếp hạng phòng mới — tên, danh hiệu, khu đang đứng, cảnh giới; danh hiệu tính theo cả phòng (5 khu); Top 1/2/3 thiết kế bục vinh danh riêng (§11.3, §11.5).

**Thay đổi v6.18:** tên Tiếng Việt chốt **Đạo Hữu Xin Dừng Bước** + mô tả/khẩu hiệu/giới thiệu; bỏ hẳn "Kiếm Xoay Tu Tiên"; đổi tên project (§20).

**Thay đổi v6.17:** chốt tên game quốc tế **BladeAscend** (§20).

**Thay đổi v6.16:** thêm đa ngôn ngữ đợt 1 (vi, en, pt-BR, es, id, th) + SEO đa ngôn ngữ (§19); tên game & thương hiệu (§20).

**Thay đổi v6.15:** `?debug=true` chỉ hoạt động sau khi đăng nhập admin (§18, ADMIN_SPEC §1.1).

**Thay đổi v6.14:** duyệt bố cục lượt 3; mật độ hình vẽ chốt 8–14% (Băng Nguyệt 4–10%) (§8.1b).

**Thay đổi v6.13:** chốt mục tiêu bố cục theo lượt 2 (mật độ hình vẽ 12–20%, khoảng thở ≥ 20%, Băng Nguyệt ≥ 40%), mẫu cụm riêng từng khu, luật túi kín ≥ 4 ô (§8.1b).

**Thay đổi v6.12:** mật độ đo theo hình vẽ (giữ mức hiện tại, không nới corridor_min); thêm luật nhịp cụm – khoảng thở, đa dạng hình khối, bản sắc từng khu (§8.1b).

**Thay đổi v6.11:** sạp của Cô Ba có bậc = cảnh giới Cô Ba + 1, Cô Ba đi xuyên sạp của mình (§6.3).

**Thay đổi v6.10:** Cô Ba đổi sang Bày Hàng Chặn Lối (dựng dãy sạp làm vật cản tạm + vỏ chuối) để không trùng vai trò cơ động với Shipper (§6.3).

**Thay đổi v6.9:** làm lại 3 kỹ năng yếu — Ông Ba Võng: Ru Ngủ Trưa (ngủ gục, kiếm ngủ không đánh trả); Sư Phụ Mì: Hút Mì Đoạt Kiếm (cướp 25% kiếm); Shipper: Gom Đơn Hỏa Tốc (×3, xuyên vật cản, miễn khống chế, hút đồ 3 U) (§6.3).

**Thay đổi v6.8:** chia lại bậc vật cản trải đủ 14 cảnh giới (0–13), Tiên giới vẫn có vật cản không chém được, chỉ Đạo Tổ chém mọi thứ; thêm tảng băng vạn năm, cột ngọc, thạch bi cấm chế (§8.2).

**Thay đổi v6.7:** vật cản bỏ HP — chém được theo bậc cảnh giới, chém là biến mất ngay; không còn vật cản không phá được trong khu; luật sắp xếp bản đồ + kiểm tra liên thông theo từng cảnh giới (§5.3, §8.0, §8.1b, §8.2).

**Thay đổi v6.6:** Tiên khí giảm từ ×20 xuống **×10** bản gốc (Chân Tiên 1 000 … Đạo Tổ 15 000); Thiêu Đốt Tiên khí vẫn 8–40/giây (frac 0.008) (§3.4, §4.1c).

**Thay đổi v6.5:** phạm vi kỹ năng tính theo U = vòng hút EXP của người dùng (pickup_r × scale), đẩy lùi theo U người dùng; Tiên giới giữ tối đa 3 lượt Bí Kíp (§6.3).

**Thay đổi v6.4:** Tiên khí gấp 20 lần (Chân Tiên 2 000 … Đạo Tổ 30 000), Thiêu Đốt Tiên khí giữ lượng tuyệt đối cũ, viên rơi có mệnh giá lớn (§3.4, §4.1c); giảm dung lượng tải + tải theo nhóm + màn hình tải có thanh tiến độ (§13.7).

**Thay đổi v6.3:** điểm độ giật tách anim lặp / anim một lần (điểm "lạc lõng"); thứ tự sửa hoạt ảnh B → C → A (§12.5).

**Thay đổi v6.2:** va kiếm xét từ vòng ngoài vào cho tới khi đủ 1 vòng đầy (sửa lỗi mài tắc khi vòng ngoài còn kiếm lẻ) (§13.4). NETCODE_SPEC lên v1.1.

**Thay đổi v6.1:** tên bot giống người thật (1 000 tên trong config/bot_names.txt, luật chọn tên & hành vi khi chết, §9.4); hoạt ảnh liền mạch — một thân nhiều khuôn mặt, chuẩn hóa frame tự động, bộ điều khiển anim mượt (§12.5).

**Thay đổi v6.0:** tối ưu mạng/hiệu năng chuyển sang tài liệu riêng `NETCODE_SPEC.md` — đồng bộ theo khu, Room Shared Cache, Protocol v2, cache nhiều lớp, chống nhấp nháy/placeholder (§13.3).

**Thay đổi v5.9:** làm rõ: chém người Tiên giới luôn rơi 30% Tiên khí mất ra đất, bất kể người chém là ai (§3.4).

**Thay đổi v5.8:** bị chém rớt 1 đại cảnh giới và thanh về **mức sàn** (tầng 1, 0%) của cảnh giới mới, không giữ tầng nữa (§5.1, §3.4).

**Thay đổi v5.7:** Tiên giới tính theo tổng Tiên khí (Chân Tiên 1–100 … Đạo Tổ 1 001–1 500), vào Tiên giới cần Đại Thừa 10 đầy thanh; Đạo Tổ nạp đầy 1 500 mới đổi Uy Danh, Thiêu Đốt được; bỏ cân bằng theo tỉ lệ thắng 1-đấu-1 (§3.4, §4.1c, §6.3).

**Thay đổi v5.6:** Thiêu Đốt bật thêm bằng nháy đúp giữ / chạm 2 lần giữ; Tiên giới đốt Tiên khí (§4.1c); bảng hướng dẫn có dòng Kỹ năng (§18.1).

**Thay đổi v5.5:** chỉnh kỹ năng — A Hớn hồi 20% kiếm tối đa; Bà Tám đẩy 500; Ông Ba Võng bán kính 500; Thằng Mập bịt mắt 5s + không nhặt đồ; Cô Ba trượt = choáng 1s; Shipper ×3 trong 5s, xuyên mọi vật cản trừ biên (§6.3).

**Thay đổi v5.4:** chỉnh kỹ năng — Gọi Đàn Em 10s, mỗi đàn em 50% số kiếm của đại ca; Lưới Mì trói 5s; Gồng Cơ 5s; Rụt Cổ 5s (§6.3).

**Thay đổi v5.3:** Tiên khí dư của Đạo Tổ đổi Uy Danh 10:1; va kiếm giữ nhịp đúng sau khi sửa lỗi SegSegDist2 (§3.4, §5.2).

**Thay đổi v5.2:** thêm Thiên Môn ở tâm mỗi khu (trừ Thiên Đình) — cổng một chiều đích cố định tới Thiên Đình (§8.4a).

**Thay đổi v5.1:** giữ 20 người/phòng (bot lấp theo phòng, sinh vào khu ít người nhất); kỹ năng chỉ dùng được khi nhặt vật phẩm Bí Kíp (tối đa 2 lượt); Thằng Mập đổi thành Phun Trân Châu bịt mắt (§6.3, §9).

**Thay đổi v5.0:** thế giới 5 khu vực tách biệt (mỗi khu = bản đồ cũ, thiết kế riêng), chỉ đi qua bằng cổng dịch chuyển đích ngẫu nhiên; sửa giật khi chạm vật cản (dùng chung mã va chạm, trượt dọc, gộp collider, hiệu chỉnh mượt); kỹ năng riêng cho 10 nhân vật; Đạo Tổ Tiên khí dư → Uy Danh (§3.4, §6.3, §8.0, §8.2a, §8.4, §9).

**Thay đổi v4.9:** ảnh đại diện nhân vật ở HUD thay số cấp (§12.4); bản đồ sinh động — trộn texture theo nhiễu, vùng hình hữu cơ (server dùng chung đa giác), decal, nước, ambient; cấm khối chữ nhật/tròn lộ ra (§8.6).

**Thay đổi v4.8:** thiết kế lại bảng tên trên đầu: ruy băng danh hiệu theo hạng, huy hiệu cảnh giới màu theo cảnh giới + chỉ báo nguy hiểm, tên to có viền, nền chung, kích thước cố định trên màn hình (§12.3).

**Thay đổi v4.7:** thần thú Tiên giới luôn lượn lờ quanh người (quỹ đạo nền + động tác ngẫu nhiên + bám theo có độ trễ) thay cho bay 1 vòng mỗi 20s (§3.5).

**Thay đổi v4.6:** bộ công cụ VFX — blend cộng, hào quang nhiều lớp, bloom cho lớp VFX, pixi-filters, particle editor (revolt-fx), texture CC0 (§3.6).

**Thay đổi v4.5:** Linh khí / Tiên khí tách hẳn, không quy đổi; bỏ lá chắn chặn đường, chém thân chỉ khi còn 1 vòng; bể kiếm tính mỗi tick, đổi chiều cooldown 0.3s (§3.4, §5.2). Xác nhận đã bỏ Trường Kiếm (v4.3).

**Thay đổi v4.4:** thêm vật phẩm Cương Kiếm (độ cứng: bền ×1.5 khi va kiếm, sát thương quái/vật cản ×2); mục tối ưu hiệu năng bắt buộc §13.6 (vẽ mỗi vòng kiếm thành 1 texture, delta snapshot, LOD, đo ping đúng).

**Thay đổi v4.3:** bỏ vật phẩm Trường Kiếm; tầm chém chỉ tăng theo số vòng và cảnh giới (§4.1a, §6.2).

**Thay đổi v4.2:** VFX thần thú bay lên từ chân cho cảnh giới Tiên (§3.5).

**Thay đổi v4.1:** độ bền kiếm mỗi cảnh giới +20% (Trúc Cơ vs Luyện Khí đổi 10 : 12), thay cho tỉ lệ 1 : (d+1) (§5.2).

**Thay đổi v4.0:** gộp toàn bộ thay đổi đã nhắn trực tiếp cho agent (tốc chạy, bất tử 5s, realm_mult 2.0, Thiêu Đốt Tinh Huyết, bảng phòng theo Uy Danh, danh hiệu Top 1–10, tin loan toàn server, bot ngang hạng 11, debug, thanh máu quái); không đẩy văng khi va kiếm + lá chắn vòng kế ngoài cùng; Tiên khí từ Đài Tiên Khí thay cho sát mạng; thuật toán sinh bản đồ tu tiên; sửa lỗi Tốc Kiếm.

**Thay đổi v3.9:** hào quang sau lưng cho các cảnh giới Tiên, mỗi cảnh giới một màu (§3.4).

**Thay đổi v3.8:** số kiếm theo vòng — vòng k chứa 10k kiếm, mỗi cảnh giới thêm 1 vòng (Luyện Khí 10 → Đạo Tổ 1 050); thanh tiến độ gọi là Linh khí / Tiên khí (§3.1, §3.4, §4.1a, §8.2).

**Thay đổi v3.7:** thêm 5 cảnh giới Tiên giới (Chân Tiên → Đạo Tổ) lên bằng sát mạng, không dùng EXP (§3.4).

**Thay đổi v3.6:** vật cản cứng hơn — damage tính theo người (0.5s/lần) thay vì theo từng kiếm, HP tăng (§5.3, §8.2).

**Thay đổi v3.5:** clash_unit = 1; chỉ đổi chiều quay khi cùng cảnh giới; số kiếm có thể về 0 (không chết); nhịp Luyện Khí → Trúc Cơ 30–60s (§5.2, §6.1).

**Thay đổi v3.4:** thêm minimap (§12.1) và danh hiệu Top 1–3 phòng (§11.3); kiếm quay chậm lại còn 1 vòng/giây, Tốc Kiếm trả về ×1.6 (§4.1b).

**Thay đổi v3.3:** bảng xếp hạng dùng điểm Uy Danh thay EXP, thêm bảng Sát Thần (lần chém trúng) và Thần Tốc (lên Đại Thừa nhanh nhất); Ẩn Thân giảm còn 5s, tàng hình 80% (§6.2, §11).

**Thay đổi v3.2:** bán kính vòng kiếm cố định, không phụ thuộc số kiếm; chỉ vòng mới mới rộng hơn (§4.1a).

**Thay đổi v3.1:** vòng kiếm vẽ kiểu cánh quạt như quảng cáo — so le góc, lưỡi nghiêng 35°, mỗi vòng một màu, vòng ngoài lùi ra xa hơn (§4.1a).

**Thay đổi v3:** kiếm dài gấp đôi, quay 300ms/vòng, kiếm tối đa 60→400 xếp vòng lồng nhau (20 kiếm/vòng), bị chém rớt 1 đại cảnh giới, EXP rải khắp bản đồ, vật phẩm ×3, vật cản chỉ chặn ở chân đế + y-sort.

**Thay đổi v2 so với v1:** đổi stack (Go server + TypeScript/PixiJS client trên web) · kiếm có thể bị bể khi va nhau (tỉ lệ theo chênh lệch cảnh giới) · scale 1 → 5 và tầm kiếm dài theo · chết ra màn hình chờ rồi match lại · bản đồ có địa hình, vật cản phá được (mọc lại sau 30s), cổng dịch chuyển · không login, nhập tên là chơi, có leaderboard · bộ vật phẩm buff có thời hạn.

---

## 1. Tổng quan

- Thể loại: multiplayer `.io` top-down 2D chạy trên **trình duyệt** (PC + điện thoại), vào là chơi, không cài đặt, không đăng nhập.
- Fantasy: dân phố tấu hài tu tiên bằng cách cho kiếm bay vòng quanh người.
- Core loop: **di chuyển → chém quái/vật cản → lượm EXP & vật phẩm → lên cảnh giới (to ra, kiếm dài và nhiều hơn) → chém người khác → bị chém thì rớt tầng, mất kiếm → tận dụng địa hình để lật kèo**.
- Điều khiển: WASD / phím mũi tên / giữ chuột trên PC; joystick ảo trên điện thoại. Không có nút tấn công.

## 2. Luồng màn hình

```
[Trang chủ] nhập tên + chọn nhân vật + nút "Vào chơi" + nút "Bảng xếp hạng"
      │
      ▼
[Đang tìm phòng...] → [Trong trận] ──chết──▶ [Màn hình chờ / Tổng kết]
                                                 ├─ "Chơi tiếp"  → matchmaking lại (có thể sang phòng khác)
                                                 ├─ "Bảng xếp hạng"
                                                 └─ "Đổi nhân vật" → Trang chủ
```

- Tên 2–16 ký tự, cho phép tiếng Việt có dấu, lọc từ tục, cho trùng tên.
- Tên & nhân vật đã chọn được nhớ trên máy (localStorage) cho lần sau.

## 3. Hệ thống tu luyện

### 3.1 Cảnh giới, kiếm tối đa, kích thước

`level_index L = r*10 + s` (r = cảnh giới 0..8, s = tầng 0..9) → 90 level.

```
scale(L) = 1 + 4 * L / 89          // 1.00 (Luyện Khí 1) → 5.00 (Đại Thừa 10), to ra ở MỖI tầng
```

**Mỗi cảnh giới mở thêm 1 vòng kiếm.** Vòng thứ k chứa `10 × k` kiếm (vòng 1: 10, vòng 2: 20, ... vòng 10: 100).
→ Cảnh giới r (0-index) có `r + 1` vòng, kiếm tối đa `= 5 × (r+1) × (r+2)`.

| r | Cảnh giới | Số vòng | Kiếm tối đa | Scale tầng 1 → tầng 10 |
|---|---|---|---|---|
| 0 | Luyện Khí | 1 | 10 | 1.00 → 1.40 |
| 1 | Trúc Cơ | 2 | 30 | 1.45 → 1.85 |
| 2 | Kim Đan | 3 | 60 | 1.90 → 2.30 |
| 3 | Nguyên Anh | 4 | 100 | 2.35 → 2.75 |
| 4 | Hóa Thần | 5 | 150 | 2.80 → 3.20 |
| 5 | Luyện Hư | 6 | 210 | 3.25 → 3.65 |
| 6 | Hợp Thể | 7 | 280 | 3.70 → 4.10 |
| 7 | Độ Kiếp | 8 | 360 | 4.15 → 4.55 |
| 8 | Đại Thừa | 9 | 450 | 4.60 → 5.00 |

Tiên giới tiếp tục quy tắc này (§3.4). Config: `ring_capacity_step = 10`; mảng `max_swords` tính sẵn từ công thức.

### 3.2 EXP (hiển thị cho người chơi: **Linh khí**)

```
exp_needed(r, s) = round(100 * 2^r * (1 + 0.08 * s))      // realm_mult = 2.0: mỗi cảnh giới tốn gấp đôi cảnh giới trước
```

| Cảnh giới | EXP/tầng (1 → 10) | Tổng cả cảnh giới | Cộng dồn |
|---|---|---|---|
| Luyện Khí | 100 → 172 | 1 360 | 1 360 |
| Trúc Cơ | 200 → 344 | 2 720 | 4 080 |
| Kim Đan | 400 → 688 | 5 440 | 9 520 |
| Nguyên Anh | 800 → 1 376 | 10 880 | 20 400 |
| Hóa Thần | 1 600 → 2 752 | 21 760 | 42 160 |
| Luyện Hư | 3 200 → 5 504 | 43 520 | 85 680 |
| Hợp Thể | 6 400 → 11 008 | 87 040 | 172 720 |
| Độ Kiếp | 12 800 → 22 016 | 174 080 | 346 800 |
| Đại Thừa | 25 600 → 44 032 | 348 160 | 694 960 |

**Triết lý tiến độ**: farm chỉ đủ để qua các cảnh giới đầu. Từ Kim Đan trở lên, đường lên chính là **cướp đoạt Linh khí của người khác** (§5.1: chém trúng 1 người = nhận 50% lượng EXP họ mất, mà lượng mất là nguyên 1 cảnh giới). Chém 1 người cao hơn mình 1–2 cảnh giới có thể nhảy vọt nhiều cảnh giới. Farm thuần lên Đại Thừa gần như không thể (nhiều giờ) — đó là chủ ý.

- Lưu `total_exp` cộng dồn, level suy ra từ bảng ngưỡng tích lũy tính sẵn.
- **Tên hiển thị**: dưới Tiên giới, mọi chữ "EXP" người chơi nhìn thấy đổi thành **Linh khí** (thanh Linh khí, viên Linh khí, "+20 Linh khí"). Trong code vẫn giữ tên `exp`.
- Đại Thừa tầng 10 **đầy thanh** (`total_exp = 694 960`, trần): Linh khí nhặt thêm bị bỏ. Lên tiếp phải hấp thụ **Tiên khí** (§3.4). Đại Thừa tầng 10 chưa đầy thanh thì vẫn nhặt Linh khí như thường.
- Lên tầng: glow nhỏ. **Đột phá cảnh giới**: anim `breakthrough`, sét, camera rung, đổi skin kiếm theo cảnh giới, nhân vật quanh đó hiện emote kinh ngạc.

### 3.4 Tiên giới — lên cảnh giới bằng Tiên khí

Sau Đại Thừa có thêm 5 cảnh giới. **Không dùng Linh khí (EXP)**, chỉ lên được bằng **Tiên khí** — loại viên đặc biệt chỉ phun ra từ các **Đài Tiên Khí** trên bản đồ (§8.5). Tiên giới **không chia tầng**. Cảnh giới Tiên được xác định thẳng từ **tổng Tiên khí đang có** (`tien_khi`, cộng dồn):

| r | Cảnh giới | Tiên khí (cộng dồn) | Trần | Kiếm tối đa | Số vòng | Scale |
|---|---|---|---|---|---|---|
| 9 | Chân Tiên | 1 – 1 000 | 1 000 | 550 | 10 | 5.2 |
| 10 | Kim Tiên | 1 001 – 3 000 | 3 000 | 660 | 11 | 5.4 |
| 11 | Thái Ất Ngọc Tiên | 3 001 – 6 000 | 6 000 | 780 | 12 | 5.6 |
| 12 | Đại La Kim Tiên | 6 001 – 10 000 | 10 000 | 910 | 13 | 5.8 |
| 13 | Đạo Tổ | 10 001 – 15 000 | 15 000 | 1 050 | 14 | 6.0 |

(v6.6: **gấp 10 lần** bản gốc v5.7 — v6.4 từng thử ×20 nhưng ước tính ôm trọn 2 đài vẫn mất ~62 phút mới tới Đạo Tổ, quá lâu so với một ván; ×10 ≈ 30 phút khi ôm 2 đài. Nguồn Tiên khí từ đài giữ nguyên.)

- **Điều kiện vào Tiên giới**: Đại Thừa tầng 10 **đầy thanh Linh khí** (694 960). Chỉ khi đó mới nhặt được Tiên khí; **viên Tiên khí đầu tiên (tien_khi = 1) → Chân Tiên ngay**.
- Lên cảnh giới khi Tiên khí **vượt trần** cảnh giới hiện tại: 1 000 → 1 001 là Kim Tiên, 3 000 → 3 001 là Thái Ất, 6 000 → 6 001 là Đại La, **10 000 → 10 001 là Đạo Tổ**.
- **Mốc khởi điểm** mỗi cảnh giới (đáy): Chân Tiên 1 · Kim Tiên 1 001 · Thái Ất 3 001 · Đại La 6 001 · Đạo Tổ 10 001.
- Config: `tien_khi_ceilings: [1000, 3000, 6000, 10000, 15000]`.

**Ai nhặt được Tiên khí**
**Linh khí và Tiên khí tách hẳn nhau** (ranh giới: Đại Thừa tầng 10 đầy thanh):
- Từ **Đại Thừa tầng 10 đầy thanh** trở lên: chỉ hấp thụ **Tiên khí**. Viên Linh khí đi xuyên qua, không nhặt được.
- Dưới mức đó (kể cả Đại Thừa tầng 10 chưa đầy): chỉ hấp thụ **Linh khí**. Viên Tiên khí đi xuyên qua, không nhặt được.
- Viên không hấp thụ được vẫn nằm đó cho người dùng được tới nhặt; client vẽ mờ 30% với người không nhặt được để họ không chạy theo vô ích.
- Không có quy đổi qua lại giữa hai loại.

**Bị chém chết khi đang ở Tiên giới (cướp đoạt Tiên khí)**
- Chết như mọi cảnh giới (§5.1), vào lại Luyện Khí 1, mất toàn bộ Tiên khí.
- Người chém (nếu đang hấp thụ được Tiên khí) nhận **30%** Tiên khí của nạn nhân, trần: không vượt quá lượng đủ lên **thêm 1 cảnh giới Tiên**.
- **Luôn rơi 30% Tiên khí của nạn nhân thành viên Tiên khí** quanh xác — bất kể người chém là ai (phàm, Tiên, bot). Đây là nguồn Tiên khí quan trọng ngoài Đài.
  - Mệnh giá **500 / 100 / 10 / 2 / 1** (lớn trước), **tối đa 40 viên**, văng bán kính 200–500, tồn tại 60s; hiệu ứng "nổ Tiên khí" vàng kim + SFX.
  - Viên rơi tuân luật hấp thụ thường (chỉ người Tiên giới / Đại Thừa 10 đầy thanh nhặt được; người khác thấy mờ 30%).
- Người chém chưa hấp thụ được Tiên khí: không nhận Tiên khí, nhận Linh khí theo §5.1 + Uy Danh (§11.1).

**Đạo Tổ: nạp đầy rồi mới đổi Uy Danh**
- Đạo Tổ vẫn nhặt Tiên khí bình thường để nạp thanh từ 10 001 lên **trần 15 000**.
- **Chỉ khi thanh đang ở đúng 15 000** thì Tiên khí nhặt thêm mới đổi sang Uy Danh theo tỉ lệ **10 Tiên khí = 1 Uy Danh** (`tien_overflow_score` = 0.1). Thanh không vượt 15 000.
- Đạo Tổ dùng Thiêu Đốt → Tiên khí bị trừ như thường (từ 15 000 trở xuống, không dưới 10 001). Muốn đổi Uy Danh tiếp thì phải **nạp lại đầy 15 000** trước.

**Hiển thị & tấu hài**
- HUD: **giữ nguyên thanh tiến độ**, chỉ đổi tên: dưới Tiên giới gọi là **Linh khí**, từ Chân Tiên trở lên gọi là **Tiên khí** (màu vàng kim), thanh chạy từ đáy tới trần cảnh giới. Ví dụ: `Kim Tiên — Tiên khí 2 000/3 000`. Đạo Tổ đầy 15 000: thanh nhấp nháy vàng + chữ "Tiên khí dư → Uy Danh".
- Đột phá Tiên giới: hiệu ứng **Thiên Kiếp** (mây đen, sét đánh 3 lần liên tiếp, màn hình rung), kill feed toàn phòng: `"{A} vượt Thiên Kiếp, phi thăng {realm}!"`.
- **Hào quang sau lưng** (mọi cảnh giới Tiên, mỗi cảnh giới một màu):

  | Cảnh giới | Màu hào quang | Hiệu ứng thêm |
  |---|---|---|
  | Chân Tiên | Trắng bạc `#E8F1FF` | Vài đốm mây nhỏ trôi quanh |
  | Kim Tiên | Vàng kim `#FFD60A` | Tia sáng tỏa ra |
  | Thái Ất Ngọc Tiên | Xanh ngọc `#4CE0B3` | Hạt sáng bay lên |
  | Đại La Kim Tiên | Tím `#B15CFF` | Hai vòng sáng xoay ngược chiều nhau |
  | Đạo Tổ | Đen trắng âm dương + viền cầu vồng | Vòng âm dương xoay chậm, viền đổi màu liên tục |

  - Vẽ **sau lưng nhân vật** (dưới sprite nhân vật, trên mặt đất), tâm ở ngang ngực, đường kính ≈ 2.2 × chiều cao nhân vật, to theo `scale`. Dưới vòng kiếm.
  - Nhịp thở: phóng to/thu nhỏ ±5% và sáng/tối nhẹ theo chu kỳ 2s.
  - Ai cũng thấy. Ẩn Thân → mờ theo nhân vật (20%). Trong bụi → ẩn theo nhân vật.
  - Rớt cảnh giới → hào quang đổi màu ngay (về Đại Thừa thì tắt), kèm hiệu ứng vỡ vụn.
  - Màu nằm trong config `tien_aura_colors`.
- Đạo Tổ hiện trên minimap của cả phòng (giống Top 1) → cả phòng hợp sức săn.
- **Thần thú bay lên từ chân** (VFX riêng mỗi cảnh giới Tiên, xem §3.5).
- Uy Danh (§11.1) dùng nguyên công thức với `r` tới 13 (chém trúng Đạo Tổ = 1 400 × lật kèo). Đột phá Tiên giới vẫn `20 × r`.

**Lưu ý cân bằng**: Đạo Tổ có 14 vòng / 1 050 kiếm, tầm chém ≈ 1 900 đơn vị. Nếu thấy quá lớn, giảm `tien_scale_step` trong config.

### 3.5 VFX Thần thú Tiên giới

Mỗi cảnh giới Tiên có một thần thú, màu khớp với hào quang:

| Cảnh giới | Thần thú | Màu | Vệt theo sau |
|---|---|---|---|
| Chân Tiên | Tiên hạc | Trắng bạc | Mây trắng nhỏ |
| Kim Tiên | Kim long (rồng vàng) | Vàng kim | Tia lửa vàng |
| Thái Ất Ngọc Tiên | Ngọc long (rồng ngọc) | Xanh ngọc | Hạt sáng xanh |
| Đại La Kim Tiên | Tử phượng (phượng hoàng tím) | Tím | Lông vũ sáng |
| Đạo Tổ | **Song long âm dương** (1 trắng, 1 đen, quấn nhau) | Đen trắng + viền cầu vồng | Mây âm dương |

**Khi nào xuất hiện**
1. **Đột phá vào cảnh giới Tiên** (đi cùng Thiên Kiếp): thần thú to, chui lên từ dưới chân, **xoắn ốc quanh người 1.5 vòng** đi lên, bay vút lên trời rồi tan thành vệt sáng. Dài 2.5s. Camera rung nhẹ, mọi người trong tầm nhìn đều thấy.
2. **Luôn lượn lờ quanh người** (thường trực, liên tục, không tắt): sau màn đột phá, thần thú thu nhỏ còn 60% và **ở lại bay quanh chủ nhân suốt thời gian còn ở Tiên giới**:
   - **Quỹ đạo nền**: vòng elip quanh thân (bán kính 1.3–1.8 × chiều cao nhân vật, ép dẹt 0.35 theo góc nhìn), tốc độ chậm ~1 vòng / 4s; độ cao nhấp nhô lên xuống nhẹ (sóng sin chu kỳ 3s) → lúc vòng sau lưng, lúc lượn ra trước mặt.
   - **Đi theo chủ**: tâm quỹ đạo bám theo nhân vật có độ trễ (lerp 0.1/frame) → khi chạy, thần thú bị kéo dài ra phía sau như dải lụa rồi đuổi kịp.
   - **Động tác ngẫu nhiên** mỗi 8–15s, chọn 1: xoắn ốc vút lên rồi bổ nhào xuống · lượn hình số 8 quanh người · đổi chiều bay · lao ra xa 1 nhịp rồi vòng về · cuộn tròn quanh đầu chủ. Mỗi động tác 1.5–3s rồi về quỹ đạo nền (chuyển mượt, không giật).
   - **Phản ứng**: chủ bị chém trúng khi có khiên / nhặt Tiên khí → thần thú lóe sáng và lượn nhanh 1 vòng; chủ Thiêu Đốt Tinh Huyết → bay sát người, vệt đỏ.
   - Chiều sâu như cũ: nửa sau quỹ đạo vẽ dưới nhân vật, nửa trước vẽ trên.
   - Đạo Tổ: 2 con (trắng, đen) bay lệch pha nhau 180°, thỉnh thoảng quấn vào nhau.
3. **Bị chém rớt cảnh giới**: thần thú hiện ra gãy vụn thành mảnh sáng rơi xuống (0.8s).

**Cách làm: chuỗi đốt (procedural), không vẽ từng frame**
- Thần thú = **1 sprite đầu + 12–20 sprite thân (đốt) + 1 sprite đuôi** (hạc/phượng thì: thân + 2 cánh vỗ 4 frame + đuôi lông). Code cho đầu chạy theo quỹ đạo, các đốt sau đi theo vị trí đầu ở các thời điểm trước đó (như con rắn) → uốn lượn mượt, không cần sinh hàng chục frame, đổi màu bằng tint.
- **Quỹ đạo xoắn ốc**: `x = cx + R(t)·cos(θ(t))`, `y = feet_y − H(t) + R(t)·sin(θ(t))·0.35` (ép dẹt theo góc nhìn 3/4); `θ` tăng 1.5 vòng, `H` từ 0 lên `3 × chiều cao nhân vật`, `R` từ 0.6 → 1.0 × bán kính thân rồi bung ra khi bay vút lên.
- **Chiều sâu**: đốt nào đang ở nửa sau vòng xoắn (`sin θ < 0`) vẽ **dưới** sprite nhân vật, nửa trước vẽ **trên** → trông như rồng quấn quanh người thật.
- Mỗi đốt nhỏ dần về đuôi; đầu hơi lắc theo hướng bay; độ trong suốt tăng dần ở 0.4s cuối.
- Vệt theo sau: hạt particle sinh từ các đốt, tồn tại 0.5s.
- Tất cả kích thước nhân `scale` của nhân vật.
- Chỉ chạy ở client, kích hoạt bởi event `Breakthrough` / `RealmDrop` và đồng hồ client (bản thường trực) → không tốn băng thông.
- Ẩn Thân: thần thú mờ theo chủ (20%). Trong bụi: ẩn theo chủ.
- **Hiệu năng (vì giờ luôn bật)**: số đốt theo chất lượng — Cao 16 đốt + vệt hạt; Vừa 10 đốt, vệt hạt 50%; Thấp 6 đốt, không vệt. Thần thú của người ở xa (chủ nhỏ hơn 40 px trên màn hình) chỉ vẽ đầu + 4 đốt. Không giới hạn số con, nhưng vệt hạt chỉ bật cho tối đa 6 con gần mình nhất. Cập nhật quỹ đạo bằng toán thuần (không tạo object mới mỗi frame).
- Placeholder khi chưa có art: đầu = tam giác, đốt = hình tròn, tô màu theo bảng trên.

### 3.6 Bộ công cụ VFX (nâng chất lượng hiệu ứng Tiên giới)

**Vì sao đang xấu**: hào quang hiện là 1 đĩa phẳng + 1 vòng viền, blend thường, không phát sáng, không chuyển động nhiều lớp → nhìn như hình dán.

**Nguyên tắc** (áp cho mọi hiệu ứng ánh sáng: hào quang, thần thú, Thiên Kiếp, Đài Tiên Khí, viên Tiên khí):
1. **Blend cộng (`blendMode = 'add'`)** cho mọi phần phát sáng — màu chồng lên nhau sẽ sáng lên thay vì che nhau. Không tốn hiệu năng.
2. **Nhiều lớp, mỗi lớp chuyển động riêng** — không bao giờ dùng 1 hình tĩnh:
   - Lớp 1: quầng sáng mềm (radial gradient, blend cộng), thở ±8% chu kỳ 2s.
   - Lớp 2: vòng phù văn / bát quái quay chậm (1 vòng / 12s).
   - Lớp 3: tia sáng tỏa ra quay ngược chiều (1 vòng / 20s), độ trong 40%.
   - Lớp 4: hạt sáng bay lên từ chân (particle), tốc độ ngẫu nhiên, mờ dần.
   - Đạo Tổ: lớp âm dương quay + viền cầu vồng đổi màu theo thời gian (hue shift).
3. **Bloom chỉ cho lớp VFX**: đặt mọi hiệu ứng sáng vào 1 container riêng `vfxLayer`, gắn `AdvancedBloomFilter` (thư viện `pixi-filters`) **cho container đó thôi**, không gắn cho cả màn hình.
4. **Hiệu ứng khoảnh khắc** (`pixi-filters`): `ShockwaveFilter` (sóng xung kích khi đột phá Tiên / Thiên Tiên Đài phun), `GlowFilter` (viền sáng nhân vật Tiên giới), `GodrayFilter` (tia sáng trên Thiên Tiên Đài khi sắp phun).
5. **"Juice"**: đột phá có khựng hình 80ms (hit-stop), rung camera, chớp trắng 1 frame, chữ cảnh giới to bật ra (scale 0 → 1.3 → 1.0, easing back-out).

**Particle**: dùng thư viện particle có trình chỉnh sửa trực quan để thiết kế bằng mắt rồi xuất config JSON (không chỉnh số trong code):
- Ưu tiên **`revolt-fx`** (hỗ trợ PixiJS 8, có editor online, làm được chuỗi hiệu ứng theo thời gian).
- Phương án 2: `@spd789562/particle-emitter` (bản port `@pixi/particle-emitter` cho PixiJS 8, dùng ParticleContainer).
- Config hiệu ứng lưu ở `client/src/vfx/presets/*.json`.

**Texture hạt miễn phí (CC0, dùng thương mại được)**: Kenney Particle Pack, Brackeys VFX Bundle (itch.io, gồm flipbook). Ghi vào `assets/CREDITS.md`. Dùng các texture này cho tia lửa, khói, ánh sáng, vì chúng đẹp và ổn định hơn ảnh AI sinh cho hạt nhỏ; art AI giữ cho phần có "nhân vật" (đầu rồng, phượng, phù văn).

**Hiệu năng (gắn với §13.6)**:
- Chất lượng **Cao**: đủ bloom + filter + particle. **Vừa**: bloom chất lượng thấp, particle 50%. **Thấp**: tắt bloom/filter, chỉ blend cộng + particle 25%.
- Chỉ người trong khung nhìn mới có VFX; người ở xa (nhỏ hơn 40 px trên màn hình) chỉ còn lớp quầng sáng.
- Bloom filter chỉ 1 lần cho cả `vfxLayer`, không gắn filter riêng từng nhân vật.

## 4.0 Luật chung: mọi thưởng/phạt CỘNG trên base (v6.32)

Mọi chỉ số bị ảnh hưởng bởi cảnh giới, vật phẩm, kỹ năng, địa hình đều tính:

```
giá_trị = base × (1 + Σ phần_thưởng − Σ phần_phạt)        // KHÔNG nhân các hệ số với nhau
```

- Mỗi nguồn đóng góp **một phần trăm của base**, cộng lại với nhau. Cùng một nguồn không cộng 2 lần.
- Hệ số kiểu "× k" ghi trong bảng cũ được hiểu là **phần thưởng = k − 1** (vd Nam Châm ×1.5 → +0.5; Tốc Kiếm ×1.6 → +0.6; Song Bội Đan ×2 → +1.0; Cương Kiếm ×1.5 → +0.5; Gồng Cơ ×2 → +1.0).
- Phần đóng góp của **cảnh giới/scale** cũng là một khoản cộng như mọi nguồn khác (vd bán kính hút: `√scale − 1`), để Luyện Khí 1 đúng bằng base.

| Chỉ số | base | Các khoản cộng | Trần |
|---|---|---|---|
| Bán kính nhặt đồ | 90 | nội tại nhân vật (§6.4) · cảnh giới `√scale − 1` · Nam Châm +0.5 · Gom Đơn (Shipper) +0.1 | ×4 base |
| Tốc chạy | 195 | nội tại nhân vật · Thần Hành +0.4 · Thiêu Đốt +0.8 · Gom Đơn +2.0 · gió ±; bùn −0.4, tuyết −0.2 | ×2 (×3 khi Gom Đơn), sàn ×0.4 |
| Tốc quay kiếm ω | 2π/1s | nội tại nhân vật · Tốc Kiếm +0.6 | ×2 |
| Độ bền kiếm khi va (D) | `1.2^r` | nội tại nhân vật · Cương Kiếm +0.5 · Gồng Cơ +1.0 · (Trộm Kiếm của Ông Ba: không mẻ) | — |
| EXP nhận | giá trị viên/quái | Song Bội Đan +1.0 | — |

Ví dụ bán kính nhặt đồ của Shipper Đại Thừa 10 (scale 5) đang Gom Đơn và có Nam Châm:
`90 × (1 + (√5 − 1) + 0.5 + 0.1) = 90 × 2.836 ≈ 255` (cách nhân dồn cũ: 90 × 2.236 × 1.5 × 1.1 ≈ 332).

## 4. Kiếm, vòng quay, tầm đánh

### 4.1a Vòng kiếm nhiều lớp (kiểu cánh quạt như quảng cáo)

Mục tiêu hình ảnh: nhiều vòng kiếm **so le, nghiêng, mỗi vòng một màu**, nhìn thành một dải dày xoáy như chong chóng. Tất cả các lớp đều lộ ra, không nấp sau nhau.

**Quy tắc cố định bán kính (BẮT BUỘC)**
- Bán kính và chiều dài của một vòng **chỉ phụ thuộc vào thứ tự vòng `k` và `scale`**, **KHÔNG phụ thuộc vào số kiếm `n`**. Thêm kiếm vào vòng đang có → chỉ lấp kín khe, vòng **không rộng ra**.
- Tầm chém chỉ tăng khi (a) mở **vòng mới** (kiếm thứ 11, 31, 61, 101...), hoặc (b) nhân vật **lên tầng/cảnh giới** (scale tăng). Không có vật phẩm nào tăng tầm chém.
- Xóa mọi công thức cũ kiểu `ring_radius = max(..., n * spacing / 2π)` (từ spec v1) khỏi code.
- Chuôi kiếm vòng 1 luôn sát người (`inner_r = (BODY_R + 6) * scale`), không có khoảng trống lớn giữa thân và kiếm.

Ví dụ ở cùng 1 tầng (scale không đổi): 1 kiếm, 5 kiếm, 10 kiếm → cùng 1 vòng, cùng bán kính. 11 kiếm → xuất hiện vòng 2 xa hơn. 30 kiếm → vẫn 2 vòng, bán kính không đổi.

**Chia vòng**
- Vòng thứ `k` (k = 0 là vòng 1) chứa tối đa `10 × (k+1)` kiếm: vòng 1 = 10, vòng 2 = 20, vòng 3 = 30... Kiếm lấp đầy vòng trong trước rồi mới sang vòng ngoài.
- Số vòng đang hiện = số vòng cần để chứa n kiếm (`R` nhỏ nhất sao cho `5R(R+1) ≥ n`), không vượt quá `r + 1` của cảnh giới.
- Trong 1 vòng, các kiếm chia đều 360°: vòng đầy có khoảng cách `36° / (k+1)`; vòng ngoài cùng chưa đầy (m kiếm) thì chia đều `360° / m`.

**Hình học từng vòng** (`R` = tổng số vòng hiện có, `base = SWORD_LEN * scale`, SWORD_LEN = 80)

```
inner_r           = (BODY_R + 6) * scale                     // bán kính chuôi vòng 1
hilt_r(k)         = inner_r + HILT_STEP * k * base           // HILT_STEP = 0.1  → vòng ngoài lùi ra xa hơn
len(k)            = base * (1 + (RING_GROWTH - HILT_STEP) * k) // RING_GROWTH = 0.2 → mũi vòng sau xa hơn vòng trước 20% chiều dài gốc
stagger(k)        = k == 0 ? 0 : 0.5 * 360° / (10 * (k+1))    // so le nửa khe của chính vòng đó → kiếm vòng ngoài không trùng thẳng hàng với vòng 1
tilt              = TILT_DEG * dir                           // TILT_DEG = 35°, lưỡi ngả về phía sau so với chiều quay
angle(i, k, t)    = phase + dir * ω * t + slot_angle(i) + stagger(k)
ω                 = 2π / rev_secs * buff_spin                // rev_secs = 1.0 → 1 vòng/giây; buff_spin = 1.6 khi có Tốc Kiếm
```

- Mỗi thanh kiếm đặt chuôi tại điểm `(hilt_r(k), angle)` và chĩa ra theo hướng **bán kính xoay thêm `tilt`**, dài `len(k)`.
- Đổi chiều quay → `tilt` lật theo, tween 0.1s (lưỡi luôn ngả về phía sau, như kéo lê theo vòng xoáy).
- Tầm chém thực tế: `reach = max_k |tip(k)|` (khoảng cách từ tâm tới mũi kiếm xa nhất, tính hình học vì kiếm nghiêng). Tham khảo tầng 10, full kiếm: Luyện Khí ≈ 160 · Trúc Cơ ≈ 245 · Kim Đan ≈ 340 · Nguyên Anh ≈ 450 · Đại Thừa ≈ 1 200 · Đạo Tổ ≈ 1 900 đơn vị.
- Config `ring_growth_mode: "linear" | "compound"` vẫn giữ để thử.

**Màu từng vòng**
- Vòng k tô màu `ring_tints[k % số_màu]` (mặc định cam → tím → xanh lá → đỏ → xanh lơ → vàng). Skin kiếm theo cảnh giới vẫn giữ hình dáng, chỉ nhuộm màu theo vòng.
- Vòng 1 dùng màu gốc của skin (tint trắng) để người chơi nhận ra cảnh giới.

**Thứ tự vẽ**: vòng ngoài cùng vẽ trước (dưới), vòng 1 vẽ sau cùng (trên). Mỗi vòng có lớp spin blur cùng màu (mục 4.1b).

**Hitbox (server dùng đúng hình học trên)**
- Mỗi kiếm: capsule từ chuôi tới mũi theo hướng đã nghiêng, độ dày `sword_width = 12 * scale`.
- Thân: hình tròn `BODY_R * scale` (BODY_R = 30).
- Số kiếm tối thiểu **0** (có thể bị tước sạch kiếm). Bắt đầu: 1 kiếm. Thêm/bớt kiếm hoặc đổi số vòng (stagger thay đổi): tween vị trí 0.2s.

### 4.1b Va chạm và hiển thị khi kiếm quay
- **Tốc độ gốc: 1 vòng / 1 giây** (`rev_secs: 1.0`). Có buff Tốc Kiếm: 1 vòng / 0.625 giây.
- **Va chạm theo góc quét**: ở 30 tick/s, mỗi tick kiếm quay ~12° (~19° khi có buff). Vẫn kiểm tra **cung đã quét** từ tick trước tới tick này (sweep), không chỉ vị trí tức thời, để kiếm không "xuyên" qua người. Mục tiêu nằm trong vành khăn `[inner_r, reach]` và trong cung quét của bất kỳ kiếm nào → trúng.
- **Hiển thị**: ở 60 fps mỗi frame quay ~6°, nhỏ hơn nhiều so với khoảng cách 18° giữa 2 kiếm → nhìn rõ từng thanh kiếm quay, không còn hiệu ứng quay ngược. Vệt mờ (spin blur) chỉ cần nhẹ, mảnh, cùng màu vòng; đậm hơn khi có buff Tốc Kiếm.
- Dùng `ParticleContainer` của PixiJS cho kiếm (có thể tới hàng nghìn sprite trên màn hình).

### 4.1 Tốc độ di chuyển & camera

```
move_speed = 195 × clamp(1 + Σ thưởng + Σ phạt, 0.4, trần)        // CỘNG trên base, KHÔNG nhân dồn
  thưởng: Thần Hành +0.4 · Thiêu Đốt +0.8 · Gom Đơn (Shipper) +2.0 · theo gió +0.5
  phạt:   bùn −0.4 · bờ tuyết −0.2 · ngược gió −0.2 · hiệu ứng làm chậm của kỹ năng (theo kỹ năng)
  trần:   2.0 (×2 base) bình thường · 3.0 khi Shipper đang Gom Đơn
  cùng một nguồn không cộng 2 lần (nhặt 2 Thần Hành = vẫn +0.4); mặt băng: quán tính tính sau, không đổi trần
camera: chiều cao vùng nhìn thấy (đơn vị thế giới) = max(900 * scale^0.85, 2.6 * reach)
```

- Tốc độ chạy **giống nhau ở mọi cảnh giới** trong thế giới.
- **Mọi hệ số tốc độ CỘNG vào base rồi kẹp trần**, không nhân dồn. Ví dụ Thần Hành + Thiêu Đốt + theo gió: cũ 1.4 × 1.8 × 1.5 = **×3.78**; mới 1 + 0.4 + 0.8 + 0.5 = 2.7 → kẹp **×2.0**. Ví dụ Thiêu Đốt trong bùn: 1 + 0.8 − 0.4 = ×1.4.
- Camera lùi theo cảnh giới nên trên màn hình người to trông đi chậm hơn người nhỏ → người nhỏ né được.

### 4.1c Thiêu Đốt Tinh Huyết (chạy nước rút)
- Giữ **Shift**, hoặc **nháy đúp rồi giữ** chuột (PC); giữ nút **🔥**, hoặc **chạm 2 lần rồi giữ** (điện thoại): tốc chạy **+0.8 base** (cộng, §4.1).
- Trong lúc giữ: mỗi giây đốt **8% EXP của tầng hiện tại** (`exp_needed(r, s) × 0.08`). Có thể rớt tầng nhưng **không bao giờ rớt cảnh giới**: tiến độ chạm đáy cảnh giới hiện tại (tầng 1, 0%) thì tự tắt.
- **Ở Tiên giới** (Chân Tiên trở lên): đốt **Tiên khí** thay cho EXP — mỗi giây **giữ nguyên lượng tuyệt đối như trước** (Chân Tiên 8/giây, Kim Tiên 16, Thái Ất 24, Đại La 32, **Đạo Tổ 40**) — tức 0.8% độ rộng khoảng Tiên khí sau khi tăng ×10 (nếu để 8% thì 1 giây chạy nhanh tốn hàng chục viên Tiên khí, quá đắt). Không bao giờ xuống dưới đáy cảnh giới đang đứng. Đạo Tổ đầy 15 000 đốt được bình thường; sau đó phải nạp lại 15 000 mới đổi Uy Danh tiếp (§3.4).
- Hiệu ứng: người bốc khói đỏ, mặt đỏ gay, vệt lửa sau chân, emote "Liều mạng!".

## 5. Chiến đấu

Thứ tự xử lý mỗi tick server:

### 5.1 Kiếm A chạm thân B — B CHẾT NGAY (v6.26)
(A ≠ B, B không bất tử, không có khiên, không đang dịch chuyển, B chỉ còn ≤ 1 vòng kiếm — §5.2)
- **Chém trúng thân là chết ở MỌI cảnh giới** (bỏ luật rớt 1 đại cảnh giới). **Ván của B kết thúc ngay**: B bị đưa ra khỏi phòng tới **màn hình Tổng kết (end game)** — không hồi sinh tại chỗ, không tự vào lại (§7).
- **A nhận Linh khí** = `30% × B.total_exp`, **trần**: không vượt quá lượng đủ đưa A **lên thêm 2 đại cảnh giới** so với hiện tại (`trần = sàn(r_A + 2) − A.total_exp`, tối thiểu 0). Nạn nhân là bot → 15% (thay 30%). Kèm Uy Danh (§11.1) + `kill_bonus_exp` (30).
- **Rơi tại xác**: `30% × B.total_exp` thành viên Linh khí (mệnh giá lớn trước, tối đa 40 viên, ai nhặt được theo luật hấp thụ thường) + 50% số kiếm thành Bó Kiếm (§7) + Tiên khí nếu B ở Tiên giới (§3.4).
- 40% còn lại **biến mất** (chống lạm phát).
- Mỗi tick, mỗi nạn nhân tối đa 1 hit. Người chém không bị đẩy lùi.
- Nếu B có buff Khiên (Kim Cang): khiên vỡ, B **không chết**, bất tử 1s, bị đẩy lùi khỏi A.

### 5.2 Kiếm A chạm kiếm B — bể kiếm

**Không đẩy văng** (`clash_knockback = 0`): hai bên mài kiếm vào nhau tới khi một bên hết lớp. VFX tia lửa + SFX "keng" tại điểm va.
- **Bể kiếm tính mỗi tick** (`clash_cooldown = 0`): ở 30 tick/s, hai người ngang nhau mài 1 giây là rụng ~30 kiếm mỗi bên (vd Kim Đan 60 kiếm → còn 30 sau 1s).
- **Đổi chiều quay** có cooldown riêng `flip_cooldown = 0.3s` (để vòng không giật 30 lần/giây).

**Luật chém thân theo số vòng** (thay cho lá chắn chặn đường đi ở bản trước):
- Hai người **đi xuyên qua vòng kiếm của nhau tự do**, không bị chặn, không bật ngược.
- Kiếm đối thủ **chỉ gây chém thân (§5.1) khi nạn nhân còn đúng 1 vòng** (≤ 10 kiếm). Còn từ 2 vòng trở lên thì thân "được che", kiếm chạm thân không tính.
- Kiếm bị bể luôn trừ từ **vòng ngoài cùng** trước. Vòng ngoài cùng hết kiếm → số vòng giảm 1.
- Hệ quả: Bó Kiếm khi rớt cảnh giới gần như không xảy ra (nạn nhân đã ≤ 10 kiếm lúc bị chém); Bó Kiếm đến chủ yếu từ xác người chết và vật phẩm trên bản đồ.

**Đổi chiều quay chỉ khi cùng cảnh giới**: `rA == rB` → cả hai vòng đổi chiều (`dir *= -1`). Khác cảnh giới → **không ai đổi chiều**, cả hai giữ nguyên chiều quay.

Sau đó tính bể kiếm theo **độ bền kiếm**: mỗi cảnh giới kiếm bền hơn cảnh giới ngay dưới **20%**.

```
durability(r) = 1.2^r                      // Luyện Khí 1.0 · Trúc Cơ 1.2 · Kim Đan 1.44 · Nguyên Anh 1.73 ...
D(người) = durability(r) × (1 + nội_tại_bền + 0.5 nếu Cương Kiếm + 1.0 nếu Gồng Cơ)   // cộng, §4.0, §6.4
k = D(bên cứng hơn) / D(bên mềm hơn) ≥ 1      // không buff: k = 1.2^d
```

- "Bên cao" / "bên thấp" trong bảng dưới nghĩa là bên có **D lớn hơn / nhỏ hơn**. Ví dụ Luyện Khí có Cương Kiếm (D = 1.5) va Trúc Cơ không buff (D = 1.2) → Luyện Khí là bên cứng hơn.
- D bằng nhau → mỗi bên mất 1.
- **Đổi chiều quay vẫn chỉ xét cảnh giới** (cùng cảnh giới mới đổi), không xét buff.

| Trường hợp | Kết quả mỗi lần va (u = `clash_unit` = 1) |
|---|---|
| Cùng cảnh giới (d = 0) | Mỗi bên mất `u` kiếm, cả hai đổi chiều |
| Khác cảnh giới | Bên thấp mất `u` kiếm; bên cao nhận `crack += lost / k`. Khi `crack ≥ 1` → bên cao mất `floor(crack)` kiếm, `crack -= floor(crack)`. Không ai đổi chiều |

Số kiếm có thể về **0** (`min_count = 0`). Lý do: vòng kiếm dày chặn hết kiếm đối thủ trước khi chạm tới thân, nên phải tước được sạch kiếm thì mới chém được vào người.
- 0 kiếm **không chết**: chỉ bị tước vũ khí. Chém trúng thân vẫn theo §5.1 (Trúc Cơ trở lên rớt 1 cảnh giới, Luyện Khí chết).
- Người 0 kiếm hiện emote hoảng loạn, nhặt item Kiếm/Bó Kiếm để có lại vòng. Bot 0 kiếm luôn chuyển sang Flee.
- Bên 0 kiếm không còn kiếm để va, nên không còn tính bể kiếm với họ.

Bảng tỉ lệ đổi kiếm (bên thấp mất : bên cao mất):

| Chênh lệch | Ví dụ | Tỉ lệ | Diễn giải |
|---|---|---|---|
| 0 | Kim Đan vs Kim Đan | 1 : 1 | Mỗi lần va mỗi bên −1, đổi chiều |
| 1 | Trúc Cơ vs Luyện Khí | 1.2 : 1 | **12 kiếm Luyện Khí bể → 10 kiếm Trúc Cơ bể** |
| 2 | Kim Đan vs Luyện Khí | 1.44 : 1 | ≈ 14 đổi 10 |
| 3 | Nguyên Anh vs Luyện Khí | 1.73 : 1 | ≈ 17 đổi 10 |
| 8 | Đại Thừa vs Luyện Khí | 4.3 : 1 | ≈ 43 đổi 10 |
| 13 | Đạo Tổ vs Luyện Khí | 10.7 : 1 | ≈ 107 đổi 10 |

→ Cảnh giới cao chỉ nhỉnh hơn mỗi bậc, đông kiếm vẫn có thể mài chết kiếm của người cao hơn 1–2 cảnh giới.

`crack` lưu trên người, reset về 0 khi chết. Hiển thị vết nứt nhẹ trên vòng kiếm khi `crack > 0`.

Kiếm bị bể: VFX vỡ vụn + SFX, mảnh văng ra (không thành item). Chủ nhân hiện emote "Kiếm của taaa!".

### 5.3 Kiếm chạm quái / vật cản
- Quái: 1 damage mỗi lần chạm (2 nếu có Cương Kiếm), cooldown mỗi cặp (kiếm, quái) 0.25s. Quái chết → **toàn bộ EXP rơi thành viên EXP** tại chỗ. Trong `orb_owner_secs` (1.5s) đầu, chỉ người chém hạ quái được nhặt; sau đó ai cũng nhặt được.
- Vật cản: **không còn tính damage**. Kiếm của người đủ cảnh giới (bậc §8.2) chạm vào là vật cản biến mất ngay. Kiếm xuyên qua vật cản và không bị mất/bể.
- Kiếm không tương tác với chính chủ.

## 6. EXP rải khắp bản đồ & vật phẩm

### 6.0 Kinh tế Linh khí (EXP) — chống lạm phát (v6.21)

**Lỗi 1/11 — nguyên nhân thật (EXP_AUDIT, Q33)**: lỗi code ở Bí Kíp. Nhặt Bí Kíp khi đã đủ lượt, code cộng *toàn bộ EXP đang có + 20* thay vì +20 → mỗi cuốn gần như nhân đôi tu vi, chiếm 80–97% EXP thu được. Nam Châm chỉ làm nhặt Bí Kíp nhanh hơn. Đã sửa về **+20 Linh khí** (Tiên giới: +2 Tiên khí) đúng §6.3. Viên EXP và quái trong code khi đó **không** nhân theo cảnh giới.

**Bảng nhịp bên dưới là CHUẨN** (Q33-b phương án A). Các hệ số trong "Luật mới" là điểm xuất phát; giá trị cuối cùng do `cmd/sim --farm` chỉnh cho trúng bảng (±25%) và trúng ràng buộc Nam Châm (≤ 1,5 lần), rồi ghi lại vào đây.

**Luật mới (mọi nguồn EXP phải theo)**
1. **Giá trị viên EXP môi trường** = giá trị gốc (5 / 20 / 60) × `orb_value_mult^r`, với `orb_value_mult = 1.5` (không phải 2). EXP cần tăng ×2 mỗi cảnh giới nên mỗi cảnh giới farm lâu hơn cảnh giới trước một chút — đúng cảm giác "càng lên càng khó".
2. **Bán kính hút đồ** = `90 × (1 + (√scale − 1) + Σ thưởng)` theo luật cộng §4.0 (không nhân thẳng scale, không nhân dồn buff). Vẫn rộng ra khi to lên nhưng không tăng tốc hút theo cấp số.
3. **Nam Châm**: bán kính hút **+0.5 base** (cộng, §4.0), 15s.
4. **Quái**: EXP khi hạ = giá trị gốc × `orb_value_mult^r_người_chém` (cùng hệ số, không riêng).
5. **Cướp EXP khi chém người** (§5.1, v6.26): chém trúng = giết; người chém nhận **30%** EXP nạn nhân, trần +2 đại cảnh giới; nạn nhân là bot → 15%. Vẫn là đường lên cấp nhanh nhất — phần thưởng của giao tranh.
6. Song Bội Đan ×2 giữ nguyên (20s, hiếm).
7. **Không nguồn nào khác** được nhân EXP theo cảnh giới. Mọi hệ số nằm trong `config.exp_economy`.

**Nhịp mục tiêu — chỉ farm, không đánh nhau, không buff** (`cmd/sim --farm`, sai lệch cho phép ±25%):

| Cảnh giới | Thời gian farm qua cảnh giới đó | Cộng dồn |
|---|---|---|
| Luyện Khí | ~1,3 phút | 1,3 |
| Trúc Cơ | ~1,5 | 2,8 |
| Kim Đan | ~1,8 | 4,6 |
| Nguyên Anh | ~2,1 | 6,7 |
| Hóa Thần | ~2,6 | 9,3 |
| Luyện Hư | ~3,3 | 12,6 |
| Hợp Thể | ~4,1 | 16,7 |
| Độ Kiếp | ~5,2 | 21,9 |
| Đại Thừa (đủ 10 tầng) | ~6,6 | **~28 phút** |

- Có Nam Châm liên tục (giả định xấu nhất): không được nhanh hơn bảng trên quá **1,5 lần** (tức Đại Thừa 10 không dưới ~19 phút).
- Đánh nhau / cướp EXP được phép nhanh hơn nhiều — đó là phần thưởng của giao tranh.
- **CI**: `cmd/sim --farm` và `cmd/sim --farm --magnet-always` chạy mỗi build; trượt khỏi khoảng cho phép → build fail. Thêm bài kiểm "Luyện Hư → Chân Tiên có Nam Châm" phải ≥ 15 phút farm thuần.

### 6.1 Viên EXP môi trường (nguồn EXP chính)
- Bản đồ luôn phủ đều viên EXP: mật độ mặc định **1 viên / 140×140 đơn vị** (≈ 7 300 viên trên bản đồ 12 000²).
- Tỉ lệ: nhỏ +5 (80%) · vừa +20 (18%) · lớn +60 (2%).
- Viên bị nhặt → 5s sau mọc lại ở vị trí ngẫu nhiên **trong cùng ô lưới 1 000×1 000** (khu vực bị vét sạch sẽ tự đầy lại, không dồn về một chỗ).
- Bán kính hút đồ mặc định: `90 * sqrt(scale)` (§6.0).
- Nhịp độ farm thuần chỉ để tham khảo, **không phải mục tiêu cân bằng**: lên cấp chủ yếu nhờ cướp Linh khí (§3.2). Dùng `cmd/sim --farm` để ghi lại số đo hiện tại vào `docs/PROGRESS.md` sau mỗi lần đổi config.
- Mạng: viên EXP đứng yên → chỉ gửi theo AOI khi vào tầm nhìn, sau đó chỉ gửi sự kiện nhặt/mọc.

### 6.2 Vật phẩm

| Item | Hiệu ứng | Thời hạn | Weight spawn |
|---|---|---|---|
| Kiếm | +3 kiếm. Nếu đã max → +20 EXP | – | 40 |
| Bó Kiếm | +15 kiếm (cũng là dạng rơi ra khi rớt cảnh giới/chết, khi đó chứa đúng số kiếm rơi) | – | 6 |
| Tốc Kiếm | **Chỉ** tăng tốc độ quay: ω **+0.6 base** (cộng, §4.0). **Không** đổi chiều dài/bán kính/tầm chém | 12s | 12 |
| Thần Hành | Tốc chạy **+0.4 base** (cộng, §4.1) | 10s | 12 |
| Nam Châm | Bán kính hút đồ **+0.5 base** (cộng, §4.0) | 15s | 8 |
| **Cương Kiếm** (độ cứng) | Kiếm cứng hơn: độ bền khi va kiếm **+0.5** (cộng, §4.0, §5.2), sát thương lên quái × 2, **chém được vật cản cao hơn 1 bậc** (§8.2) | 10s | 7 |
| Kim Cang (khiên) | Chặn 1 lần bị chém vào thân | 8s hoặc tới khi vỡ | 7 |
| Song Bội Đan | EXP nhận **+1.0** (cộng, §4.0) | 20s | 5 |
| Ẩn Thân Phù | Tàng hình **80%**: người khác thấy mình và vòng kiếm mờ (độ đậm 20%), không thấy tên, không hiện trên minimap. Ở rất gần (120 × scale của người nhìn) thì hiện rõ 100%. Hết hiệu lực ngay khi kiếm mình chém trúng người | **5s** (tối đa, nhặt lại chỉ làm mới về 5s) | 3 |
| **Rương Báu** (hiếm) | Người đầu tiên **chạm vào** → nổ ra 8 item ngẫu nhiên văng xung quanh | – | 1 rương mỗi 60s, báo trước toàn phòng 10s |

- Số vật phẩm duy trì: **200/phòng**, item bị nhặt → 3s sau mọc lại chỗ khác.
- Nhặt lại buff đang có → làm mới thời gian (không cộng dồn).
- HUD hiển thị icon buff + vòng đếm ngược. Mỗi buff có aura màu riêng (trừ Ẩn Thân).
- Item spawn cách người chơi ≥ 300 đơn vị.

### 6.3 Kỹ năng nhân vật — hồi chiêu + tích trữ lượt (v6.25)

Mỗi nhân vật có **1 kỹ năng riêng**, **tự hồi** theo thời gian, **không còn nhặt Bí Kíp** (vật phẩm Bí Kíp đã bỏ; art chuyển vào `assets/library/`).
- Mỗi kỹ năng có **thời gian hồi 1 lượt** và **số lượt tích tối đa** riêng, tuỳ độ mạnh (bảng dưới). Lượt hồi lần lượt: dùng 1 lượt thì đồng hồ hồi chạy tiếp cho lượt kế, đầy thì dừng.
- **Tiên giới** (Chân Tiên trở lên): **+1 lượt tích tối đa** cho mọi nhân vật. Rớt về Đại Thừa → cắt về mức thường.
- Vào trận / hồi sinh: có sẵn **1 lượt**, đồng hồ hồi bắt đầu chạy.
- Giữa 2 lần dùng khoá **3s** (chống bấm dồn).
- Nút **Space** / nút kỹ năng: icon + **vòng hồi chiêu** + **chấm số lượt** (vd ●●○). 0 lượt thì xám và hiện giây còn lại.
- Chết mất hết lượt tích. Qua cổng, rớt cảnh giới không mất (trừ quy tắc Tiên giới ở trên).
- Không dùng được khi đang dịch chuyển; đang Ẩn Thân mà dùng thì lộ.
- **Mọi phạm vi kỹ năng tính theo đơn vị U của người dùng**, không dùng số cố định:
  `U = 90 × scale_người_dùng` (Luyện Khí 1 = 90 · Kim Đan 1 ≈ 171 · Đại Thừa 10 = 450 · Đạo Tổ = 540). **Lưu ý:** U là đơn vị riêng của kỹ năng, **không** đổi theo công thức bán kính hút mới `90 × √scale` (§6.0) — đổi bán kính hút không được làm co phạm vi kỹ năng.
  Mọi chiều dài, bán kính, khoảng đẩy trong bảng dưới ghi dạng `k × U`. **Khoảng đẩy lùi cũng tính theo U của người dùng**, không phụ thuộc kích thước người bị đẩy (người nhỏ hay to đều bị đẩy cùng một quãng).
  Đẩy lùi diễn ra trong 0.3s, trượt dọc vật cản (§8.2a), không xuyên vật cản.
  Thời gian hiệu ứng không đổi theo cảnh giới.
- Vì sao chọn U (vòng hút EXP) thay vì tầm chém: vòng hút chỉ phụ thuộc cảnh giới (scale), không đổi khi mất/thêm kiếm, nên kỹ năng ổn định và dễ đoán. Tham khảo: tầm chém full kiếm ≈ 1.8 U (Luyện Khí) → 2.7 U (Đại Thừa); hệ số kỹ năng chọn 3–8 U để kỹ năng vươn xa hơn vòng kiếm 1.5–3 lần.
- **Cân bằng** (v6.30): **không dùng tỉ lệ thắng đấu tay đôi** để cân bằng. Kỹ năng có giá trị theo nhiều cách (tranh tài nguyên, lên cấp nhanh, thoát thân, đánh đông người), tay đôi không đo được. `cmd/sim --duel` chỉ còn là **số tham khảo** trong báo cáo, không phải hàng rào CI.
  - Cân bằng dựa trên **số liệu người chơi thật** (≥ 2 ngày telemetry): tỉ lệ chọn nhân vật, Uy Danh/phút, thời gian sống, cảnh giới cao nhất mỗi mạng, số lần dùng kỹ năng. Nhân vật có tỉ lệ chọn quá thấp kéo dài → chủ dự án quyết chỉnh.
- **Không kỹ năng khống chế nào được vô hiệu hoá vòng kiếm** (v6.27): choáng, ngủ, trói, trượt chỉ được khoá **di chuyển / kỹ năng / Thiêu Đốt / tầm nhìn**; vòng kiếm của người bị khống chế **luôn quay và đánh trả bình thường** (vì chém trúng là chết — khoá kiếm = giết chắc). Ngoại lệ duy nhất: bị mù (Thằng Mập) thì kiếm không chém trúng **thân** người khác, nhưng vẫn va kiếm bình thường.
- Luật cân bằng chung: kỹ năng **không** trực tiếp đổi cảnh giới/EXP, **không** chém thân thay kiếm, **không** xuyên luật "còn 1 vòng mới chém được thân" (§5.2).
- Không dùng được khi đang dịch chuyển; đang Ẩn Thân mà dùng thì lộ (hết Ẩn Thân).

| Nhân vật | Kỹ năng | Hiệu ứng | Hồi 1 lượt | Tích tối đa |
|---|---|---|---|---|
| A Hớn Bánh Bao | **Bánh Bao Thần** | Ăn 1 cái bánh bao: **hồi ngay 20% số kiếm tối đa** của cảnh giới hiện tại (không vượt tối đa, lấp vào vòng trong trước), ợ một tiếng to | 30s | 2 |
| Bà Tám Loa | **Loa Phường** | Hét vào loa: sóng âm **toả tròn 360°** bán kính **6 U**, **đẩy lùi mọi địch xung quanh** ra **6 U** (theo U của Bà Tám) và làm vòng kiếm của họ **đổi chiều**. Mạnh nhất khi tranh Đài Tiên Khí, Rương Báu, xác vừa nổ đồ — dọn sạch quanh mình rồi nhặt | 18s | 2 |
| Ông Ba Võng | **Ru Ngủ Trưa — Trộm Kiếm** | Hát ru 1 câu (đứng yên 0.5s, bị chém trúng thì mất chiêu) rồi nằm võng **4s**: (1) mọi địch trong **4 U** **ngủ gục 2.5s** (không di chuyển/kỹ năng/Thiêu Đốt; kiếm vẫn quay và đánh trả); (2) trong 4s đó **kiếm của Ông Ba không bị mẻ** khi va kiếm, và **mỗi thanh kiếm của đối thủ bị bể khi va với Ông Ba bay về nhập vòng của ông** (tới trần; dư rơi Bó Kiếm). Ông Ba vẫn di chuyển được trên võng | 25s | 2 |
| Đại Ca Dép Tổ Ong | **Gọi Đàn Em** | Huýt sáo gọi 2 đàn em (NPC nhỏ hơn đại ca 30%, cùng cảnh giới) chạy theo **8s**, bám quanh đại ca trong **3 U**, tìm đối thủ trong **8 U**. **Mỗi đàn em có số kiếm = 30% số kiếm hiện tại của đại ca** (giảm từ 50% — tỉ lệ thắng cũ 87%) (làm tròn xuống, tối thiểu 1, xếp vòng như người chơi), lao vào mài kiếm đối thủ gần nhất. Đàn em không chém thân được, không nhặt đồ, không qua cổng; hết kiếm hoặc hết giờ thì biến mất (khói + câu "Đại ca ơi em đi đây"). Đại ca không bị trừ kiếm khi gọi | 40s | 1 |
| Thằng Mập Trà Sữa | **Phun Trân Châu (bịt mắt)** | Phun 1 loạt trân châu hình nón 45° dài **8 U** phía trước. Người trúng **bị mù 5s**: màn hình dính đầy trân châu, **không nhặt được gì**, và **kiếm của họ không chém trúng thân bất kỳ ai** (vẫn va kiếm, vẫn mài kiếm bình thường — chỉ không giết được người). Dùng để thoát thân hoặc cứu đồng đội đang còn 1 vòng | 20s | 2 |
| Cô Ba Gánh Hàng | **Bày Hàng Chặn Lối** | Quăng gánh bày ra **1 dãy 5 sạp hàng** hình vòng cung, cách **3 U** theo hướng đang nhắm (chuột / hướng joystick), dài **6 U**, tồn tại **6s**. Sạp là **vật cản tạm** có bậc = **cảnh giới Cô Ba + 1** (tối đa 13): người **cùng cảnh giới hoặc thấp hơn** không chém được, phải đi vòng; Cô Ba **tự đi xuyên** sạp của mình → bỏ chạy chắc chắn trước người ngang cấp. Dọc chân dãy sạp rải **vỏ chuối** rộng **1 U**: địch dẫm phải → **trượt theo quán tính 1s và choáng**. Sạp không đặt đè lên người (đoạn nào trùng người thì bỏ trống đoạn đó). Dùng để chặn đường chạy của con mồi, chắn lối vào Đài, hoặc đặt sau lưng khi bỏ chạy | 22s | 2 |
| Sư Phụ Mì Gói | **Hút Mì Đoạt Kiếm** | Ném sợi mì vào **1 địch gần nhất** trong nón 60° dài **5 U**, rồi húp "sùn sụt": **cướp 25% số kiếm hiện có** của họ (lấy từ vòng ngoài cùng), kiếm bay dọc sợi mì về **nhập vào vòng của Sư Phụ** (không vượt tối đa; tối đa cướp = 20% kiếm tối đa của Sư Phụ). Phần cướp vượt trần thì rơi xuống đất thành Bó Kiếm. Không có ai trong nón → sợi mì hụt, vẫn mất lượt | 30s | 2 |
| Bé Tí Cơ Bắp | **Gồng Cơ** | Gồng **4s** (giảm từ 5s — tỉ lệ thắng cũ 85%): độ bền kiếm **× 2** (cộng với Cương Kiếm theo §4.0: cả hai = +1.5), không bị đẩy lùi/chậm/trói | 35s | 1 |
| Anh Shipper Tia Chớp | **Gom Đơn Hỏa Tốc** | 4s: tốc chạy **+2.0 base** (trần ×3, §4.1), **đi xuyên mọi vật cản** (trừ biên), **miễn mọi khống chế** (chậm, trói, choáng, ngủ, đẩy lùi), và **bán kính nhặt đồ +10% base** (` +0.1 base, **cộng dồn** với cảnh giới và Nam Châm theo §4.0). Hết giờ mà kẹt trong vật cản → đẩy mượt ra chỗ trống gần nhất 0.3s. Điểm mạnh: chạy nhanh nhất game + xuyên vật cản + miễn khống chế để tới Đài / xác vừa nổ đồ trước người khác | 20s | 2 |
| Cụ Rùa Thiền | **Rụt Cổ** | Rụt vào mai **5s**: đứng yên, **không bị bể kiếm, không bị chém thân**; ra khỏi mai bật sóng bán kính **2 U**, đẩy địch ra **3 U** | 30s | 2 |

Bảng tham khảo (đã nhân U) — chỉ để hình dung, code tính từ hệ số:

| Kỹ năng (hệ số) | Luyện Khí 1 (U=90) | Kim Đan 1 (U≈171) | Đại Thừa 10 (U=450) | Đạo Tổ (U=540) |
|---|---|---|---|---|
| Loa Phường: bán kính & đẩy (6 U, toả tròn) | 540 | 1 026 | 2 700 | 3 240 |
| Ru Ngủ Trưa: bán kính ngủ (4 U) | 360 | 684 | 1 800 | 2 160 |
| Phun Trân Châu: nón (8 U) | 720 | 1 368 | 3 600 | 4 320 |
| Bày Hàng: khoảng cách / độ dài dãy sạp (3 U / 6 U) | 270 / 540 | 513 / 1 026 | 1 350 / 2 700 | 1 620 / 3 240 |
| Hút Mì: nón (5 U) | 450 | 855 | 2 250 | 2 700 |
| Rụt Cổ: sóng / đẩy (2 U / 3 U) | 180 / 270 | 342 / 513 | 900 / 1 350 | 1 080 / 1 620 |


**Nguyên tắc thiết kế kỹ năng (dùng khi chỉnh tiếp)** — một kỹ năng được người chơi chọn khi nó:
1. **Gắn vào vòng lặp chính** (kiếm, mài kiếm, cướp Linh/Tiên khí, Đài, cổng) — không chỉ là hiệu ứng phụ. Ví dụ bản cũ của Ông Ba Võng/Sư Phụ Mì chỉ làm chậm/trói nhưng địch vẫn mài kiếm bình thường → không đổi kết quả trận.
2. **Có khoảnh khắc dùng rõ ràng** ("đang mài kiếm thì ru ngủ", "Đài sắp phun thì gom đơn").
3. **Có cách chống lại**: Ru Ngủ phải đứng 0.5s và bị chém là mất chiêu; Hút Mì chỉ 1 mục tiêu, hụt vẫn mất lượt; Gom Đơn không gây sát thương.
4. **Tạo vai trò riêng, không trùng nhau**: Ông Ba Võng = mở giao tranh, Sư Phụ Mì = đoạt kiếm từ xa, Shipper = cơ động / cướp tài nguyên (nhân vật **duy nhất** có kỹ năng di chuyển nhanh), Cô Ba = **dựng địa hình** chặn đường / bẫy. Shipper đi xuyên được dãy sạp của Cô Ba → hai người khắc nhau.
5. **Đã mắt, tấu hài**: khoảnh khắc thắng phải nhìn thấy được (kiếm bay theo sợi mì, cả đám ngủ gục ngáy, Shipper hút đồ như máy hút bụi).

**Trân châu dính màn hình** (hiệu ứng của Thằng Mập):
- Chỉ hiện trên màn hình **người bị trúng** (lớp overlay phía client, server gửi event `Blinded {secs}`).
- 25–40 viên trân châu đen bóng kèm vệt trà sữa **bắn "bẹp" lên màn hình** ở vị trí ngẫu nhiên, phủ khoảng **55–65%** diện tích, dày hơn ở giữa; mỗi viên **trượt chậm xuống** để lại vệt, rồi rơi/tan dần trong 5s.
- Trong 5s đó người bị trúng **không nhặt được bất cứ thứ gì** và **kiếm không chém trúng thân ai** (server bỏ qua chém thân, va kiếm vẫn tính); quanh nhân vật hiện icon 🚫 nhỏ, vòng kiếm phủ lớp trân châu mờ để người khác biết họ đang "mù". Có tiếng "bẹp bẹp" + "rột rột".
- Không che HUD (thanh tiến độ, nút, minimap) để người chơi vẫn biết mình đang ở đâu.
- Bot bị trúng: tầm nhìn còn 40%, không nhặt đồ, và kiếm không chém trúng thân ai trong 5s.
- Trúng lần nữa khi đang bị: làm mới thời gian, thêm 10 viên.

- Mỗi kỹ năng có câu thoại tấu hài khi dùng (bong bóng thoại) + SFX riêng.
- Bot dùng kỹ năng khi có lượt, theo luật đơn giản (vd Cụ Rùa dùng Rụt Cổ khi còn 1 vòng và có địch gần; Bà Tám dùng Loa khi địch trong nón; Ông Ba Võng dùng Ru Ngủ khi ≥ 1 địch trong 4 U đang va kiếm với mình; Sư Phụ Mì dùng khi có địch nhiều kiếm trong nón; Shipper dùng khi Đài sắp phun / có Tiên khí rơi trong 8 U / bị khống chế; Cô Ba dùng khi con mồi yếu hơn đang chạy (đặt chặn phía trước nó) hoặc khi chính mình bỏ chạy (đặt sau lưng)).
- Server tính mọi hiệu ứng; client nhận event `SkillCast {id, skill, pos, dir}`, `SkillEffect`, `Blinded`.

### 6.4 Chỉ số gốc riêng từng nhân vật (nội tại, v6.33)

Mỗi nhân vật có **nội tại** chỉnh nhẹ 4 chỉ số gốc. Nội tại là **một khoản cộng** như mọi nguồn khác trong §4.0 (vd Shipper tốc chạy = 195 × (1 + 0.15 + các buff khác)).

| Nhân vật | Tốc chạy | Bán kính nhặt | Tốc quay kiếm | Độ bền kiếm | Tổng | Lý do |
|---|---|---|---|---|---|---|
| A Hớn Bánh Bao | -5% | +10% | -5% | 0 | +0 | Tham ăn: hút đồ xa hơn; bụng phệ chạy hơi chậm. |
| Bà Tám Loa | -5% | -5% | 0 | +10% | +0 | Lì đòn chợ búa: kiếm cứng; chen lấn chậm, mải nói nên nhặt kém. |
| Ông Ba Võng | -10% | 0 | -5% | +15% | +0 | Lười: chạy chậm, quay chậm; da dày nên kiếm bền — hợp Trộm Kiếm. |
| Đại Ca Dép Tổ Ong | -10% | 0 | +5% | +5% | +0 | Ra vẻ đại ca: đi khệnh khạng; kiếm quay nhanh, cứng. |
| Thằng Mập Trà Sữa | -10% | +15% | 0 | -5% | +0 | Hút trà sữa rột rột: hút đồ xa nhất; nặng nề, kiếm hơi giòn. |
| Cô Ba Gánh Hàng | +10% | +5% | -5% | -10% | +0 | Quen gánh chạy chợ: nhanh nhẹn, nhặt khéo; kiếm mỏng. |
| Sư Phụ Mì Gói | 0 | -5% | +10% | -5% | +0 | Võ sư: kiếm quay nhanh nhất; mắt lé nên nhặt kém. |
| Bé Tí Cơ Bắp | -5% | -10% | +5% | +10% | +0 | Cơ bắp: kiếm cứng, quay mạnh; chân ngắn, tay to vớ đồ kém. |
| Anh Shipper Tia Chớp | +15% | 0 | -5% | -10% | +0 | Chạy nhanh nhất game; kiếm đồ chơi giao kèm đơn nên giòn. |
| Cụ Rùa Thiền | -15% | 0 | 0 | +15% | +0 | Chậm nhất game; mai rùa nên kiếm bền nhất. |

- **Tổng mỗi nhân vật = 0** (ưu điểm đổi bằng nhược điểm), không ai mạnh hơn hẳn trên giấy. Mỗi chỉ số lệch tối đa **±15%**.
- Nội tại **hợp với kỹ năng**: Shipper chạy nhanh + Gom Đơn; Cụ Rùa/Ông Ba kiếm bền + kỹ năng phòng thủ/trộm kiếm; Thằng Mập/A Hớn hút xa để farm; Sư Phụ quay nhanh để mài kiếm.
- Độ bền ở đây cộng vào hệ số độ bền khi va kiếm (§5.2): `D = 1.2^r × (1 + nội_tại_bền + Cương Kiếm + Gồng Cơ)`.
- Trang chọn nhân vật: dưới khung chi tiết có **4 thanh chỉ số** (Tốc · Nhặt · Quay · Bền) tô xanh nếu dương, đỏ nếu âm, kèm số %, cạnh dòng "Hồi Xs · tích N lượt".
- Cân bằng như §6.3: theo số liệu người chơi thật; chỉnh nội tại theo bước 5%, giữ tổng = 0.

## 7. Chết & màn hình chờ

**Chết = hết ván (v6.29).** Người thật chết → rời phòng ngay, hiện **màn hình Tổng kết**. Không có hồi sinh tại chỗ, không đếm ngược tự vào lại, không tự ghép phòng mới. Muốn chơi tiếp phải tự bấm **"Chơi tiếp"** → ghép phòng như người mới (có thể vào phòng khác), bắt đầu Luyện Khí 1.
- Cơ chế tự vào lại khi trang bị tải lại (vd mất WebGL trên điện thoại, `pendingRejoin`) **chỉ áp dụng khi người chơi còn sống**; nếu ván đã kết thúc vì chết thì sau khi tải lại phải hiện lại màn hình Tổng kết, không tự vào trận.
- Bot chết vẫn theo luật hồi sinh/rút bot ở §9.3 (bot không có màn hình Tổng kết).

- **Rơi kiếm khi chết** (người và bot): `floor(n * death_sword_drop)` kiếm (mặc định 50%) văng ra quanh xác dưới dạng Bó Kiếm (mỗi bó ≤ 10). Ví dụ chết (ở Luyện Khí) với 10 kiếm → 1 bó 5 kiếm.
- **Người thật**: anim `death` 1s → rời phòng → **Màn hình Tổng kết** hiển thị:
  - Bị ai chém (tên + nhân vật), câu tấu hài random ("Bạn đã được tiễn về quê nuôi cá").
  - Cảnh giới cao nhất đạt được, tổng EXP đỉnh, số người đã chém, thời gian sống, hạng trong phòng.
  - Nút: **Chơi tiếp** (matchmaking lại ngay, giữ tên + nhân vật), **Bảng xếp hạng**, **Đổi nhân vật**.
- Khi vào lại: Luyện Khí tầng 1, 1 kiếm, không buff, **bất tử 5s**, spawn ở chỗ an toàn (không có ai trong 800 đơn vị).
- **Bot**: chết thì xử lý theo mục 9.3 (respawn tại chỗ hoặc bị xóa).
- Kill feed tấu hài trong trận: `"{A} vừa tiễn {B} về quê nuôi cá"`, `"{B} bị {A} chém rớt xuống {realm}"`, `"{A} đột phá {realm}, cả phòng run rẩy"`, `"{A} chui cổng ra liền bị {B} chém, đen thôi đỏ quên đi"`.

## 8. Bản đồ & địa hình

### 8.0 Thế giới 5 khu vực

Mỗi phòng là **1 thế giới gồm 5 khu vực tách biệt**. Mỗi khu vực to bằng toàn bộ bản đồ cũ (12 000 × 12 000) → tổng diện tích gấp 5 lần, giảm cảnh vừa vào đã bị dí.

- Các khu vực **không nối nhau bằng đường đi** (tường biên mỗi khu không qua được). **Chỉ đi qua khu khác bằng cổng dịch chuyển** (§8.4).
- Mỗi khu vực có thiết kế, bố cục và luật địa hình riêng:

| # | Khu vực | Bố cục | Điểm đặc biệt | Đài Tiên Khí |
|---|---|---|---|---|
| 1 | **Thiên Đình** | Bố cục hiện tại (§8.1): Thiên Tiên Đài giữa, 4 sân môn phái, đường chính + vành đai, 4 biome nhỏ | Khu "trung tâm", nơi duy nhất có Thiên Tiên Đài → điểm nóng giao tranh | 4 nhỏ + Thiên Tiên Đài |
| 2 | **Linh Sơn Cốc** | Các dãy vách đá (không phá được) chạy dài chia khu thành nhiều **hẻm núi**, nối bằng **đèo hẹp** và **cầu treo** (đoạn đường hẹp 200–300) | Đánh nhau ở đèo/cầu, người to khó qua đèo hẹp | 2 đài trên 2 "đỉnh núi" (cao nguyên nhỏ có 2 lối lên) |
| 3 | **Trúc Lâm Mê Cung** | Rặng tre dày (không phá được) tạo **mê cung** với nhiều ngõ cụt, bụi rậm khắp nơi, vài ao | Phục kích, tầm nhìn ngắn: trong khu này bụi rậm nhiều gấp 3 | 2 đài ở 2 khoảng trống giữa mê cung |
| 4 | **Phàm Trấn** | **Lưới phố** kiểu thị trấn: nhà (khung tường gạch có cửa), chợ đêm (sạp hàng tạo hẻm 150/250/350), quảng trường nhỏ | Nhiều vật cản phá được nhất → thay đổi địa hình liên tục | 2 đài ở 2 quảng trường |
| 5 | **Băng Nguyệt Hồ** | Hồ băng lớn ở giữa + đảo băng, bờ tuyết xung quanh, ít vật cản | **Mặt băng trơn**: trên băng di chuyển có quán tính (trượt, khó dừng/đổi hướng); bờ tuyết làm chậm 0.8 | 2 đài trên 2 đảo băng |

- **Hướng sửa từng khu theo bản kiểm tra 31/10** (người Đại Thừa, scale 5):
  - *Thiên Đình*: 2 sân môn phái bên phải bị **nhốt kín** (Đài bên trong không tới được), góc trên trái có túi kín; còn 51% chỗ trống. → Mở 3–4 cổng sân rộng ≥ 400, tường sân bậc 7; cột ngọc bậc 11 quanh Thiên Tiên Đài, 4–6 thạch bi bậc 13; bỏ 30–40% tường gạch lẻ giữa sân trung tâm và sân môn phái, gom phần còn lại thành đảo.
  - *Linh Sơn Cốc*: các dải vách chạy ngang cắt khu thành tầng, đèo quá hẹp → **cả nửa dưới (có 1 cổng, 1 Đài) bị cô lập**. → Mỗi dải vách ≥ 3 đèo rộng ≥ 300, vách bậc 7, đá tảng bậc 3, đá khổng lồ bậc 9 ở miệng đèo (khe ≥ 340).
  - *Trúc Lâm Mê Cung*: hành lang ~150 → người to gần như kẹt khắp nơi (trống còn 46%). → Hành lang ≥ 220; tre chia bậc theo vòng: ngoài bậc 1, giữa bậc 2, lõi quanh Thiên Môn bậc 5 (muốn đi tắt vào tâm phải đủ Luyện Hư, không thì tìm đường); vài thạch bi bậc 13 quanh Thiên Môn.
  - *Phàm Trấn*: ổn nhất; nới ngõ giữa các khối nhà ≥ 160 (nhà = gạch bậc 2), giữ đường phố chính trống.
  - *Băng Nguyệt Hồ*: ngược lại — **quá trống** (85%), chỉ có đá tảng lẻ. → Thêm 4–6 đảo băng nhỏ: đá tảng bậc 3 + tảng băng vạn năm bậc 9 quanh 2 Đài làm chỗ nấp, vẫn giữ mặt hồ rộng.
- Mỗi khu dùng thuật toán sinh riêng (`mapgen` nhận `zone_type`), vẫn tuân thủ: đường/quảng trường chính không vật cản, kiểm tra liên thông **trong khu**, §8.6 (không khối hình học lộ ra). Khu 1 giữ nguyên thuật toán §8.1.
- **Băng trơn**: vận tốc mới = vận tốc cũ × 0.92 + hướng input × gia tốc nhỏ mỗi tick (tức là trượt dài); thiết kế art cần mặt băng bóng, vết trượt.
- **Vào trận** (lần đầu hoặc "Chơi tiếp"): spawn ở **khu ít người thật nhất** (hòa → ngẫu nhiên), chỗ an toàn trong khu đó.
- HUD: tên khu vực hiện cạnh minimap; vào khu mới hiện chữ lớn mờ dần giữa màn hình ("— Trúc Lâm Mê Cung —", 1.5s).
- Minimap chỉ vẽ khu đang đứng. Nhấn **M** (phóng to): thêm cột bên cạnh liệt kê 5 khu + số người đang ở mỗi khu (không hiện vị trí).
- Kill feed: tin từ khu khác có tiền tố tên khu `[Linh Sơn Cốc]`. Bảng xếp hạng phòng, danh hiệu: tính chung cả thế giới.

### 8.1 Khu 1 — Thiên Đình: thuật toán sinh bản đồ "Tu Tiên Giới"

Bản đồ 12 000 × 12 000, tường biên không phá được. Mỗi phòng sinh 1 bản đồ từ `seed` (lưu lại để debug/tái hiện). Bản đồ Tiled vẽ tay (nếu có) được ưu tiên hơn bản đồ sinh.

**Bố cục cố định (giống một "tu tiên giới" có môn phái, đường núi, tiên đài)**

```
┌───────────────────────────────────────────┐
│ [Đài TK]  Linh Sơn         Trúc Lâm [Đài TK]│
│     ╲      (đá tảng)       (bụi, tre)  ╱     │
│      ╲___                         ___╱      │
│          ╲____   ┌─────────┐  ____╱          │
│               ╲__│ Thiên   │_╱               │
│                  │ Tiên Đài│                 │
│               ╱‾‾│ (giữa)  │‾╲               │
│          ____╱   └─────────┘  ╲____          │
│      ___╱                          ╲___      │
│     ╱      Phàm Trấn          Hắc Thị  ╲     │
│ [Đài TK]  (làng, rào tre)   (chợ, hẻm)[Đài TK]│
└───────────────────────────────────────────┘
```

- **Thiên Tiên Đài** ở trung tâm: quảng trường bán kính 900, nền đá sáng, không vật cản bên trong, viền là vòng cột/tường thấp có 4 lối vào (theo 4 đường chính).
- **4 Đài Tiên Khí** gần 4 góc (cách mỗi cạnh 15% bản đồ), mỗi đài nằm trong một **sân môn phái**: tường không phá được bao quanh, có 2–3 cổng ra vào, bán kính sân 700.
- **4 vùng (biome) theo 4 góc phần tư**, dùng lại art hiện có:
  - Linh Sơn (núi đá): nhiều đá tảng thành cụm, ít bụi.
  - Trúc Lâm (rừng tre): nhiều bụi rậm + rào tre, chỗ phục kích.
  - Phàm Trấn (làng): các "ngôi nhà" là khung tường gạch hình chữ nhật có cửa, rào tre quanh vườn.
  - Hắc Thị (chợ): lưới sạp hàng/xe đẩy tạo **hẻm hẹp** 150 / 250 / 350.
  - Ranh giới giữa các vùng dùng **Voronoi có nhiễu** (không phải đường thẳng) để trông tự nhiên.

**Các bước thuật toán**
1. **Khung**: đặt Thiên Tiên Đài, 4 sân môn phái + Đài Tiên Khí, 4 điểm vào của sân.
2. **Đường chính**: nối mỗi sân môn phái tới Thiên Tiên Đài bằng đường cong (Bezier + nhiễu nhẹ), rộng 500. Thêm **đường vành đai** vòng tròn bán kính ~3 500 nối 4 vùng. Đường lát đá sáng màu, **không đặt vật cản trên đường**.
3. **Đối xứng xoay 4 phía**: sinh nội dung 1 góc phần tư, xoay 90° cho 3 góc còn lại (công bằng cho mọi hướng), rồi mỗi góc áp **biome riêng** (đổi loại vật cản/decor) và xê dịch ngẫu nhiên nhỏ (±150) để không lộ là bản sao.
4. **Vật cản theo biome**: dùng **Poisson-disk sampling** (khoảng cách tối thiểu theo loại) + **cụm** (mỗi cụm 3–7 vật cản). Tường gạch xếp thành khung nhà có lỗ cửa; rào tre xếp thành dãy liền dọc mép đường; sạp hàng xếp lưới tạo hẻm.
5. **Vùng địa hình**: bụi rậm thành cụm ở mép rừng và cạnh đường (chỗ phục kích); bùn ở chỗ trũng giữa vùng; đường gió chạy dọc vài đoạn đường chính hướng về tâm.
6. **Cổng dịch chuyển**: 4 cổng mỗi khu (khu 1: 1 cổng trong mỗi sân môn phái; khu khác: rải đều 4 phía), thuộc mạng cổng chung của cả thế giới (§8.4). Cổng chỉ đặt ở chỗ trống, cách vật cản ≥ 300.
7. **Decor**: Poisson-disk thưa (≤ 1 món / 600×600), không đặt trên đường, không đè vật cản.
8. **Kiểm tra hợp lệ** (sai thì sinh lại với seed+1, tối đa 10 lần):
   - Lưới đi bộ 100×100: mọi ô trống phải **liên thông** (flood fill) — không có khu bị nhốt.
   - Diện tích trống ≥ 65%.
   - Không vật cản trong bán kính 300 quanh cổng, Đài Tiên Khí, lối vào sân.
   - Mỗi biome có ít nhất 1 hẻm/lối hẹp và 1 cụm bụi.
9. Xuất bản đồ ra cùng định dạng object như Tiled (`.tmj` JSON) để client/server dùng chung và để xem lại trong Tiled.

Thêm lệnh `go run ./cmd/mapgen --seed N --png out.png` xuất ảnh xem trước để duyệt nhanh.

### 8.2 Vật cản — chém được theo cảnh giới (không còn độ cứng)

**Bỏ hẳn HP / độ cứng.** Mỗi vật cản có 1 **bậc** = cảnh giới tối thiểu để chém được. Kiếm của người **đủ cảnh giới** (chỉ số cảnh giới ≥ bậc) chạm vào → vật cản **biến mất ngay lập tức**, không hoạt ảnh vỡ, không đống đổ nát (chỉ 1 tiếng "bụp" ngắn). Người chưa đủ cảnh giới: kiếm đi xuyên như cũ, không có tác dụng.
**Không còn vật cản nào không phá được** trong khu — chỉ **biên khu vực** là không qua được.

Bậc = **chỉ số cảnh giới tối thiểu** (0 = Luyện Khí … 8 = Đại Thừa, 9 = Chân Tiên … 13 = Đạo Tổ). Có vật cản ở cả Tiên giới; **chỉ Đạo Tổ mới chém được mọi thứ**.

| Bậc | Vật cản | Chém được từ | Hành lang tối thiểu (§8.1b) |
|---|---|---|---|
| 0 | Hàng rào tre, rào gỗ | Luyện Khí | – |
| 1 | Sạp hàng, xe đẩy, tre mỏng | Trúc Cơ | 150 |
| 2 | Tường gạch, tre dày | Kim Đan | 160 |
| 3 | Đá tảng | Nguyên Anh | 180 |
| 5 | Tre cổ thụ, tường đá thường | Luyện Hư | 240 |
| 7 | Tường sân môn phái, vách núi | Độ Kiếp | 290 |
| 9 | Đá khổng lồ, **tảng băng vạn năm** | Chân Tiên | 340 |
| 11 | **Cột ngọc Thiên Đình** (quanh Thiên Tiên Đài, bến Thiên Môn) | Thái Ất Ngọc Tiên | 370 |
| 13 | **Thạch bi cấm chế** (bia đá phù văn phát sáng) | Đạo Tổ | 390 |
| ∞ | Biên khu vực | Không bao giờ | – |

- **Mỗi khu phải có vật cản bậc Tiên** (9, 11 hoặc 13) để kỹ năng dựa trên vật cản (Shipper xuyên vật cản, đẩy lùi ép vào tường, Gánh Lướt, Rụt Cổ…) vẫn có tác dụng với người Tiên giới. Gợi ý: Thiên Đình — cột ngọc quanh Thiên Tiên Đài + 4–6 thạch bi cấm chế; Linh Sơn Cốc — đá khổng lồ ở các đèo; Băng Nguyệt Hồ — tảng băng vạn năm trên đảo; Trúc Lâm — thạch bi quanh Thiên Môn ở tâm; Phàm Trấn — 2–4 thạch bi ở quảng trường.
- Vật cản bậc cao là **điểm tựa chiến thuật**, không phải tường chắn đường: luôn xếp lẻ hoặc cụm nhỏ, khe giữa chúng ≥ hành lang tối thiểu → không ai bị kẹt.

- Bậc gán theo **loại vật cản** (config `obstacles.min_realm`), mapgen có thể ghi đè từng vật bằng thuộc tính Tiled `min_realm` (vd tre trong Trúc Lâm theo vòng).
- **Cương Kiếm** (§6.2): trong thời gian buff, chém được vật cản có bậc ≤ cảnh giới + 1 (vd Kim Đan chém được bậc 3).
- **Đạo Tổ** chém được mọi vật cản (trừ biên). Các cảnh giới Tiên khác vẫn có vật cản không chém được.
- **Nhận biết bậc**: vật cản cao hơn cảnh giới của mình hiện **viền tối mờ** khi nằm trong tầm kiếm; kiếm chạm vào → tia lửa nhỏ + chữ nổi "Cần {cảnh giới}" (tối đa 1 lần / 3s / vật cản). Vật cản mình chém được trông bình thường.
- **Mọc lại** sau **30s** ở đúng chỗ cũ: 3s trước khi mọc nhấp nháy mờ cảnh báo; có người/quái đứng trong vùng → chờ trống rồi mới mọc (đẩy mượt ra như §8.2a).
- Sạp hàng / xe đẩy bị chém: vẫn 30% rơi 1 item.
- Bot: tìm đường coi vật cản mình chém được là "đi qua được, tốn thêm 0.3s".

- **Hai vùng va chạm khác nhau cho mỗi vật cản:**
  - **Chân đế (footprint)** — chặn thân nhân vật và quái. Chỉ là phần sát mặt đất: mặc định hình elip/chữ nhật rộng 90% sprite, cao **30% sprite, nằm ở đáy**. Trong Tiled đặt được thủ công bằng thuộc tính `footprint` (nếu không có thì tính tự động như trên).
  - **Thân hình (visual bounds)** — vùng kiếm chém trúng để gây damage (đúng với cái người chơi nhìn thấy).
- **Sắp xếp lớp vẽ theo chân (y-sort)**: nhân vật đứng phía sau (y chân nhỏ hơn đáy vật cản) được vẽ khuất sau vật cản; phần vật cản che lên nhân vật mình thì mờ đi 50% để vẫn thấy mình.
- **Vật cản dạng dãy** (hàng rào, tường): art phải liền mạch, nối tiếp nhau không hở (tile ngang). Chỗ nào art có khe hở thì footprint cũng phải hở theo — người chơi thấy đi qua được thì phải đi qua được.
- Kiếm đi xuyên vật cản; chém được hay không theo bậc ở trên.

### 8.1b Luật sắp xếp bản đồ (áp cho cả 5 khu, mapgen phải kiểm)

**Nguyên tắc "không chém được thì phải đi lọt"**: người cảnh giới r chưa chém được vật cản bậc > r, nên mọi khe giữa hai vật cản phải đủ rộng cho người **to nhất** trong số những người không chém được chúng:
- Khe giữa vật cản bậc A và bậc B ≥ `corridor_min[min(A, B)]` (bảng §8.2). Vật cản bậc 0 không cần (ai cũng chém được).
- Lý do con số: bậc T chặn người cảnh giới < T, người to nhất trong nhóm đó là cảnh giới T−1 (tầng 10 nếu là phàm giới) → đường kính thân `60 × scale` + 40 dư, làm tròn lên.

**Khung bản đồ** (thay cho kiểu rải tường dài xuyên khu):
1. **3 cấp đường**: đường chính rộng ≥ 600, **không bao giờ có vật cản**; đường phụ ≥ 300; ngõ hẹp chỉ nằm bên trong cụm.
2. **Vật cản xếp thành "đảo"** giữa các đường (cụm nhà, cụm đá, cụm tre), mỗi đảo tối đa ~1 200 × 1 200. **Không tường liền dài quá 800**; tường dài hơn phải có khe ≥ `corridor_min` của bậc đó.
3. **Đảo phân tầng**: viền ngoài bậc thấp (rào, sạp), lõi bậc cao (gạch, đá) → người nhỏ chui vào nấp, người lớn chém thẳng vào.
4. **Quanh cổng, Thiên Môn, Đài Tiên Khí, bến Thiên Môn**: trống bán kính ≥ 600, có ≥ 3 lối vào rộng ≥ 400.
5. **Mật độ**: đo theo **diện tích hình vẽ** của vật cản (chân đế chỉ ≈ 27% hình vẽ). Mục tiêu hình vẽ: Băng Nguyệt Hồ 4–10%, các khu khác **8–14%** (chốt theo lượt bố cục 3 đã duyệt: cụm rỗng ruột — nhà có sân, bãi đá rời, lùm tre so le — nên mật độ hình vẽ thấp mà vẫn đủ đầy; trần cứng 26%). Không nới `corridor_min` để tăng mật độ.
6. **Nhịp "cụm – khoảng thở"**, không rải đều: **≥ 20%** diện tích khu (Băng Nguyệt Hồ ≥ 40%) là **khoảng trống** (không vật cản trong bán kính 400) để giao tranh; vật cản dồn thành cụm/đảo có hình dáng rõ (nhà, sân, lùm tre, bãi đá), giữa các cụm là khoảng trống. Cấm kiểu rải đều như giấy kẻ ô (Poisson đều khắp khu) — người chơi không có mốc để nhớ đường.
7. **Đa dạng hình khối & mẫu cụm riêng từng khu**: trộn kích thước, xoay, hướng. Mỗi khu có **mẫu cụm riêng**, không dùng chung một mẫu ô 2×2: Linh Sơn = bãi đá tròn/méo, xoay ngẫu nhiên; Trúc Lâm = lùm tre dài cong, xếp so le; Phàm Trấn = nhà chữ nhật có sân/cửa; Thiên Đình = điện/sân có tường bao; Băng Nguyệt = đá & băng tròn trên đảo và bờ.
8. **Bản sắc từng khu là luật, không phải trang trí**: Thiên Đình = trục đường chữ X + vành đai + sân; Linh Sơn Cốc = dải vách núi là nhân vật chính, thung lũng giữa các dải phần lớn trống, chỉ vài bãi đá; Trúc Lâm = **lối đi quanh co giữa các lùm tre** (mê cung hữu cơ, hành lang ≥ 220, không phải lưới); Phàm Trấn = lưới phố + khối nhà (tường gạch thành nhà có sân/cửa); Băng Nguyệt Hồ = **mặt hồ gần như trống** (chỉ đảo có vật cản), vật cản dồn ở bờ và trên đảo.
9. **Đối xứng** (xoay 4 phía hoặc gương) để mọi hướng vào công bằng.

**Kiểm tra bắt buộc** — `cmd/mapgen --check` (CI fail nếu trượt): với **mỗi cảnh giới r = 0..13**, dùng thân to nhất của r (tầng 10) và **bỏ các vật cản bậc ≤ r** (vì r chém được), flood fill lưới 50:
- Toàn bộ chỗ trống phải **liên thông** (không túi bị nhốt).
- Mọi cổng, Thiên Môn, Đài, bến, điểm spawn đều tới được.
- Diện tích trống ≥ 60%; khoảng thở (không vật cản trong bán kính 400) ≥ 20% (Băng Nguyệt Hồ ≥ 40%); vật cản trên mặt hồ Băng Nguyệt (ngoài đảo) ≤ 1%; **không túi kín nào ≥ 4 ô lưới 50** ở mọi cảnh giới.
- Xuất ảnh `docs/map_check/{khu}_r{r}.png`: xám = vật cản, cam = chỗ thân không lọt, đỏ = vùng bị nhốt.

### 8.2a Di chuyển mượt khi chạm vật cản (sửa lỗi giật)

Nguyên nhân thường gặp của giật/khựng khi đi sát vật cản: (1) client dự đoán va chạm khác server (khác hình footprint, quên góc xoay ±6°/lật) → server kéo nhân vật về → giật; (2) đẩy ra khỏi vật cản bằng cách "bật" theo pháp tuyến → rung qua lại giữa 2 vật cản sát nhau; (3) khe giữa các đoạn rào/tường liền nhau làm vướng như bậc thềm; (4) sửa sai vị trí bằng cách nhảy thẳng tới vị trí server.

Bắt buộc:
- **Dùng chung 1 mã va chạm** cho server và client dự đoán (port cùng thuật toán, cùng dữ liệu footprint đã tính sẵn **sau khi** xoay/lật, cùng số thực). Có test so khớp: 1 000 lần di chuyển ngẫu nhiên sát vật cản → vị trí client và server lệch < 0.5 đơn vị.
- **Trượt dọc** (slide): khi chạm, bỏ thành phần vận tốc đi vào vật cản, giữ thành phần song song bề mặt. Lặp tối đa 3 lần trong 1 tick để xử lý góc hai vật cản. Không bao giờ bật ngược.
- Footprint trơn: hình tròn / capsule / đa giác lồi có bo góc; các đoạn rào/tường liền nhau **gộp thành 1 collider** dài, không để khe/bậc giữa các đoạn.
- Vật cản vỡ/mọc lại: khi mọc lại trùng người, đẩy người ra mượt trong 0.2s (không dịch tức thời).
- **Hiệu chỉnh dự đoán mượt**: lệch < 30 đơn vị → kéo dần về trong 100ms; lệch lớn hơn mới nhảy thẳng.
- Debug `H`: vẽ footprint client (xanh) và server (đỏ) chồng lên nhau, có bộ đếm "số lần hiệu chỉnh/giây" trong F3.

### 8.3 Vùng địa hình (không chặn)

| Vùng | Hiệu ứng | Dùng để lật kèo |
|---|---|---|
| Bụi rậm | Ai đứng trong bụi thì người ngoài bụi không thấy (trừ khi đứng gần 150 × scale). Bị chém trúng thì lộ 2s | Phục kích, trốn |
| Bùn lầy | Tốc chạy × 0.6 | Dụ kẻ đuổi vào bùn |
| Đường gió | Lệch so với hướng gió < 60° → × 1.5 (theo gió); > 120° → × 0.8 (ngược gió); còn lại × 1.0 (ngang gió) | Đường thoát nhanh |
| Hẻm hẹp | Các lối đi rộng 150 / 250 / 350 giữa vật cản không phá được | Người to (thân rộng) không chui lọt, người nhỏ trốn vào được |

### 8.4 Cổng dịch chuyển (đích ngẫu nhiên, nối 5 khu)
- Mỗi khu có **4 cổng** → cả thế giới 20 cổng, cùng 1 mạng. **Không còn ghép cặp cố định.**
- Bước vào 1 cổng → ra ở **1 cổng ngẫu nhiên khác trong cả thế giới** (có thể cùng khu hoặc khu khác; không bao giờ ra lại chính cổng vừa vào). Với 20 cổng, ~80% lần dịch chuyển sẽ sang khu khác. Config `gates.exit_pool`: `"all"` (mặc định) | `"other_zone"` (luôn sang khu khác). Tắt ngẫu nhiên: `gates.random_exit: false` (khi đó ra cổng kế tiếp theo vòng).
- Kích hoạt khi tâm nhân vật đi vào bán kính cổng (150). Anim `teleport_out` 0.4s → xuất hiện ở cổng đích với anim `teleport_in` 0.4s.
- Cổng đích **lóe sáng 0.5s trước** khi người xuất hiện (cảnh báo nhẹ cho người đứng gần — vẫn dễ bị bất ngờ).
- **Không bất tử khi vừa ra cổng.** Trong lúc đang dịch chuyển (0.8s) không bị chém và không chém được.
- Cooldown mỗi người: 8s mới dùng cổng tiếp được.
- Người to cỡ nào cũng dùng được.
- Mỗi cổng có màu theo khu nó đang đứng (để biết mình đang ở khu nào), nhưng **không cho biết** sẽ ra ở đâu.
- Chuyển khu: server chuyển nhân vật sang mô phỏng của khu đích ở ranh giới tick (mọi trạng thái — kiếm, buff, crack, Tiên khí — giữ nguyên).

### 8.4a Thiên Môn — cổng thẳng tới Thiên Đình

Mỗi khu **trừ Thiên Đình** có thêm **1 Thiên Môn ở chính giữa khu** (tổng 4). Đây là lối đi **chắc chắn** tới Thiên Đình cho người muốn leo cao (Thiên Tiên Đài, 4 Đài Tiên Khí nhỏ), khác với 4 cổng thường đi ngẫu nhiên.

- **Một chiều, đích cố định**: bước vào Thiên Môn → luôn ra ở Thiên Đình. Từ Thiên Đình muốn rời đi thì dùng cổng thường (ngẫu nhiên).
- **Điểm ra ở Thiên Đình**: 4 **Bến Thiên Môn** nằm ngoài rìa quảng trường Thiên Tiên Đài (cách tâm ~1 400, mỗi bến ở đầu 1 con đường chính). Mỗi khu nối cố định với 1 bến (Linh Sơn Cốc → bến Bắc, Trúc Lâm → Đông, Phàm Trấn → Nam, Băng Nguyệt Hồ → Tây). Bến là bục đá nhỏ có cột sáng, không vật cản trong bán kính 400.
- Cùng luật cổng thường: kích hoạt khi tâm nhân vật vào bán kính (200), dịch chuyển 0.8s không chém/không bị chém, **không bất tử khi ra**, bến đích lóe sáng 0.5s trước. Cooldown chung với cổng thường (8s).
- **Vị trí trong từng khu** (mapgen phải chừa quảng trường bán kính 500 ở tâm khu, có ít nhất 2 đường dẫn tới, kiểm tra liên thông):
  - Linh Sơn Cốc: trên một bình nguyên giữa các hẻm núi, 2–3 đèo dẫn vào.
  - Trúc Lâm Mê Cung: khoảng trống trung tâm mê cung (đường vào ngoằn ngoèo — muốn lên Thiên Đình phải tìm đường).
  - Phàm Trấn: quảng trường chính giữa thị trấn.
  - Băng Nguyệt Hồ: đảo băng giữa hồ (phải trượt băng mới tới).
- **Nhận diện**: to gấp đôi cổng thường, cổng đá trắng – mái vàng, cột sáng vàng bắn lên trời (thấy từ xa), biển "Thiên Đình ↑". Minimap: icon vàng riêng, luôn hiện. Khi đứng trong khu, góc minimap có mũi tên nhỏ chỉ hướng tới Thiên Môn.
- Bot: dùng Thiên Môn khi đã ở Đại Thừa trở lên hoặc khi Thiên Tiên Đài sắp phun (xác suất cao), còn lại hiếm khi dùng.
- Kill feed khi người trong Top 3 bước qua Thiên Môn: `"{A} ({danh hiệu}) vừa bước qua Thiên Môn — Thiên Đình dậy sóng!"`.

### 8.5 Đài Tiên Khí

| Đài | Vị trí | Phun thường | Phun lớn |
|---|---|---|---|
| **Đài Tiên Khí nhỏ** (khu 1: ×4, mỗi khu khác: ×2 — tổng 12) | Khu 1: giữa 4 sân môn phái; khu khác: theo bảng §8.0 | Mỗi **5s**: 2–3 viên nhỏ (2 điểm) | Mỗi **30s**: 15 viên nhỏ + 2 viên vừa (10 điểm) |
| **Thiên Tiên Đài** (×1, chỉ khu 1) | Trung tâm khu Thiên Đình | – | Mỗi **150s**: 1 viên **khổng lồ** (150 điểm) |

- Viên văng ra ngẫu nhiên trong bán kính 300–700 quanh đài (theo cung, như đài phun nước).
- **Báo trước**: đài nhỏ rung + sáng dần 3s trước đợt phun lớn; Thiên Tiên Đài thông báo toàn phòng **10s trước** (`"Thiên Tiên Đài sắp giáng Tiên khí!"`) + đếm ngược trên đỉnh đài + icon trên minimap.
- Viên Tiên khí khổng lồ: to, sáng vàng kim, bay lơ lửng nhấp nhô; ai chạm trước nhận cả viên.
- Viên Tiên khí không mọc lại ở chỗ khác; tồn tại tối đa 60s rồi tan.
- Đài là địa hình không phá được, kiếm đi xuyên. Không vật cản quanh đài 300 đơn vị.
- Minimap: hiện 5 đài; đài sắp phun lớn thì nhấp nháy.

### 8.6 Hình ảnh bản đồ sinh động (không dùng khối hình học)

**Quy tắc cứng**: bản đồ hiển thị **không được có hình chữ nhật thẳng cạnh hay hình tròn hoàn hảo** nào lộ ra (vùng nền, bùn, bụi, đường, sân...). Hình học đơn giản chỉ được dùng cho va chạm và chế độ debug (`H`).

**1. Mặt đất: trộn texture theo nhiễu (texture splatting)**
- Mỗi biome có 2–3 texture nền (vd Linh Sơn: đá xám, sỏi, rêu; Trúc Lâm: cỏ, lá khô, đất ẩm). Trộn bằng **mặt nạ nhiễu** (Perlin/Simplex, 2 tầng tần số) → mảng màu loang tự nhiên thay vì ô vuông.
- Ranh giới giữa 2 biome: dải chuyển tiếp rộng 400–600, trộn theo nhiễu (không phải đường thẳng hay đường cong trơn).
- **Bake sẵn** lúc vào phòng thành các tấm `RenderTexture` 2048×2048 (chunk), chỉ vẽ chunk trong khung nhìn. Không trộn lại mỗi frame.

**2. Vùng địa hình có hình dạng hữu cơ**
- Bùn, ao, bụi, sân, vùng gió... sinh dạng **đa giác méo theo nhiễu**: `r(θ) = r0 × (1 + 0.25 × noise(θ))`, làm mượt (Chaikin 2 lần), 24–40 đỉnh.
- **Server dùng đúng đa giác này** cho logic (điểm-trong-đa-giác) → hình thấy và vùng tác dụng khớp nhau. `mapgen` xuất đỉnh đa giác trong file bản đồ.
- Mép vùng: viền mềm (feather 24–48 px bake sẵn) + 1 dải viền texture riêng (mép bùn có vệt nứt, mép ao có cỏ lau, mép đường có sỏi vụn).
- Đường đá: vẽ theo đường cong với mép răng cưa nhẹ, thỉnh thoảng có viên đá lát lệch, rêu mọc kẽ.

**3. Decal & trang trí rải theo biome** (bake vào chunk, không có va chạm)
- Mỗi chunk 30–60 decal ngẫu nhiên (Poisson-disk), **xoay / lật / co giãn 0.7–1.3 / đổi sắc nhẹ** để không lặp: vết nứt, khóm cỏ, lá rụng, sỏi, vũng nước nhỏ, rễ cây, hoa dại nhỏ màu nhạt, vết bánh xe, mảnh ngói vỡ...
- Quanh vật cản: bóng đổ mềm + vụn nhỏ dưới chân (tường gạch có gạch vỡ, đá tảng có sỏi, sạp hàng có rác) → vật cản "cắm" vào đất thay vì nổi lên.
- Vật cản cùng loại: xoay nhẹ ±6°, lật ngang ngẫu nhiên, 2–3 biến thể sprite.

**4. Nước & điểm nhấn**
- Mỗi bản đồ 2–4 **ao / suối nhỏ** (đa giác hữu cơ, chỉ là vùng làm chậm như bùn hoặc trang trí không va chạm), mặt nước có shader gợn sóng nhẹ + phản chiếu lấp lánh.
- Biome Linh Sơn có **vách đá / vực** dọc biên (thay tường biên thẳng), Trúc Lâm có hàng tre dày làm biên.

**5. Sự sống (ambient, chỉ client, không ảnh hưởng gameplay)**
- **Bóng mây** trôi chậm trên mặt đất (2–4 bóng mờ cỡ lớn).
- Theo biome: Trúc Lâm lá tre rơi + đom đóm; Linh Sơn sương mù mỏng trôi; Phàm Trấn gà chạy, bướm; Hắc Thị đèn lồng đung đưa; Thiên Tiên Đài hạt sáng bay lên.
- Chim bay ngang màn hình thỉnh thoảng (bóng chim dưới đất).
- Giới hạn tổng số hạt ambient theo chất lượng (Cao 150 / Vừa 60 / Thấp 0).

**6. Màu sắc** vẫn tuân thủ ART_PROMPTS §9.0: nền nhạt, tương phản thấp; vật chơi được (EXP, item, quái, nhân vật) luôn nổi bật hơn nền.

## 9. Phòng, matchmaking, bot

### 9.1 Phòng
- **Trần người thật / phòng: chỉnh được ở trang admin** (§9.5). Mặc định theo môi trường: **production 50**, **staging 20** (máy staging yếu). Bot không tính vào trần này. 1 phòng = 1 thế giới 5 khu, mỗi khu 1 goroutine.
- **Sàn nhân vật mỗi khu: 10** (người + bot). Bot lấp cho mỗi khu đủ 10 khi đang được phép sinh bot.
- Phòng tự hủy khi 0 người thật trong 60s.

### 9.2 Matchmaking
```
on_join(player):
  candidates = rooms where humans < room_max_humans   // giá trị admin đang đặt
  room = candidate có humans lớn nhất; hòa → phòng tạo sớm nhất
  if none: room = create_room(random_map)
```

### 9.3 Bot
```
every 2s per ZONE:
  if bot_spawn_enabled(room) and zone_humans + zone_bots < zone_min_chars (10):
      spawn min(zone_min_chars - zone_humans - zone_bots, 2) bots vào khu này
on human join/leave:  KHÔNG xóa bot
on bot death:
  if not bot_spawn_enabled(room) or zone_humans + (zone_bots - 1) >= zone_min_chars: remove bot
  else: respawn bot sau 3s cùng khu

bot_spawn_enabled(room):
  mode = room.bot_mode ?? global.bot_mode        // "auto" | "on" | "off" (admin đặt, §9.5)
  if mode == "off": false
  if mode == "on":  true
  // auto — ngưỡng chỉnh ở admin, mặc định production 40 · staging 20
  if humans >= bot_auto_off_at: room.auto_off = true
  if humans <  bot_auto_off_at - 5: room.auto_off = false   // trễ 5 người chống bật/tắt liên tục
  return not room.auto_off
```
- **Bot sinh ra như người chơi mới**: Luyện Khí tầng 1, 1 kiếm, không buff, 1 lượt kỹ năng — giống hệt người vừa bấm "Vào chơi" (bỏ luật "ngang hạng 11" và `max_spawn_realm`). Bot vẫn tu luyện lên dần trong trận như người thật. Mục đích: người mới vào không gặp toàn bot cấp cao.
- Bot vẫn có 5s bất tử lúc sinh như người chơi.
- Bot **không vào leaderboard toàn cục** (server biết, người chơi không biết).
- **Tên bot phải không phân biệt được với người thật** (§9.4).
- Bot cũng dùng cổng: khi bỏ chạy (Flee) gần cổng, hoặc khu đang có quá ít mục tiêu/vật phẩm (xác suất thấp, cooldown 60s).

### 9.5 Cấu hình phòng & bot (admin + tự động)

**Chỉnh ở trang admin, áp dụng ngay không cần khởi động lại** (lưu SQLite, ghi `admin_audit`):

| Thông số | Production mặc định | Staging mặc định | Ghi chú |
|---|---|---|---|
| `room_max_humans` — trần người thật/phòng | **50** | **20** | Bot không tính. Hạ trần không đuổi ai: phòng đang vượt trần chỉ ngừng nhận người mới. |
| `zone_min_chars` — sàn nhân vật mỗi khu | **10** | **10** | Bot lấp mỗi khu đủ số này khi được phép sinh. |
| `bot_auto_off_at` — tự tắt sinh bot khi người thật ≥ | **40** | **20** | Bật lại khi < ngưỡng − 5. |
| `bot_mode` — chế độ bot | Tự động | Tự động | Toàn server + ghi đè từng phòng: Tự động / Bật / Tắt. |

- Mặc định lấy theo `APP_ENV` (production / staging) từ `config/game.json` khối `room_env`; giá trị admin đã đặt **ghi đè** mặc định và **sống qua deploy**.
- Khi bot bị tắt (tay hoặc tự động): không sinh bot mới, bot chết không hồi sinh → vơi dần tự nhiên. Nút **"Rút bot ngay"**: bot không giao tranh trong 5s rời phòng bằng hiệu ứng khói.
- Trang admin hiện cảnh báo nếu đặt `room_max_humans` cao hơn mức đã qua loadtest của máy đang chạy (lấy từ PERF_LOG mới nhất: production 50 đạt trên i5-12400F; staging ARM 4 nhân chỉ đạt 20).

### 9.4 Đặt tên bot

- Nguồn: `config/bot_names.txt` — 1 000 tên theo đúng kiểu nickname game thủ Việt hay dùng: tên thật (có/không dấu, kèm năm sinh như `2k7`, `_09`), tên ở nhà (Tèo, Bống, Mít...), biệt danh tu tiên hài (Sư Huynh Nợ Tiền, Tán Tu Lang Thang...), câu nói đùa (Đừng gank em, Wifi hàng xóm...), tên tiếng Anh ghép (SleepyNinja, LazyPanda...), kiểu teencode chữ thường (`tho2k9`, `meo.22`), và một ít có ký tự trang trí ASCII (`_Max.`, `^^Vy_`). Không có chữ "bot", không có tên người nổi tiếng hay nhân vật có bản quyền.
- **Ký tự trang trí chỉ dùng ASCII** (`_ . ~ ^ x`) vì BitmapFont của bảng tên (§12.3) có thể không có glyph ★ ♥ ✿. Tên người chơi thật có ký tự lạ ngoài font → vẽ bằng font dự phòng, không được hiện ô vuông.
- **Luật chọn tên trong 1 room**:
  - Rút ngẫu nhiên không lặp; không trùng tên bất kỳ người/bot nào đang trong room hoặc vừa rời trong 10 phút.
  - 15% số lần: thêm biến thể nhỏ (đổi `_` ↔ `.`, thêm 1–2 chữ số) để tên không lặp y hệt giữa các room.
  - Phân bố nhân vật bot giống phân bố nhân vật người chơi chọn (lấy thống kê 7 ngày, chưa có thì đều).
- **Hành vi giống người khi chết** (người thật chết → màn Tổng kết → bấm Chơi tiếp mất vài giây):
  - 70%: hồi sinh **cùng tên** sau ngẫu nhiên 4–12s (như người bấm "Chơi tiếp"), vào khu ít người như người thật.
  - 30%: "rời phòng" — bot biến mất, bot mới (tên mới) vào sau 5–20s.
  - Thỉnh thoảng (5%) bot đứng yên 3–8s sau khi vào (như người đang đọc bảng hướng dẫn).
- Kill feed, tin loan, danh hiệu, Uy Danh hiển thị cho bot y hệt người thật.

## 10. Bot AI

Quyết định lại mỗi 0.3–0.6s (random), tìm đường bằng A* trên lưới 100 đơn vị (cập nhật khi vật cản vỡ/mọc).

| State | Điều kiện | Hành vi |
|---|---|---|
| Farm | Mặc định | Tới quái, item, vật cản yếu gần nhất |
| Hunt | Mục tiêu trong 600 × scale có `power` < 0.8× mình | Đuổi, nhắm vòng kiếm vào thân; tránh lao kiếm vào kiếm đối thủ cùng/cao cảnh giới hơn |
| Flee | Có kẻ `power` > 1.25× trong 500 × scale | Chạy ngược hướng; ưu tiên hẻm hẹp (nếu mình nhỏ hơn), bụi rậm, cổng |
| Ambush | Mình nhỏ, gần bụi rậm/cổng | Nấp chờ kẻ yếu đi ngang |
| Wander | Không có gì | Đi ngẫu nhiên |

- `power = sword_count * (r + 1) * reach`.
- Tham số tính cách random mỗi bot: `reaction_delay` 0.2–0.6s, `aim_noise`, `aggression`.
- Bot chỉ "thấy" những gì người chơi cùng vị trí thấy được (tôn trọng bụi rậm). Người đang Ẩn Thân: bot chỉ phát hiện trong 50% tầm nhìn bình thường.

## 11. Leaderboard (không cần login)

EXP **không** dùng để xếp hạng toàn cục vì có trần (Đại Thừa tầng 10) → ai chơi lâu cũng bằng nhau. Thay bằng điểm **Uy Danh** thưởng cho việc đánh nhau, lật kèo và giữ ngôi.

### 11.1 Điểm Uy Danh (tính trong 1 mạng, server tính)

| Hành động | Điểm |
|---|---|
| **Chém chết người** | `100 × (cảnh giới nạn nhân + 1) × lật_kèo` |
| Thưởng thêm khi chém chết | +50 (mọi lần chém trúng giờ đều là chém chết) |
| **Lật kèo**: nạn nhân cao cảnh giới hơn mình | `lật_kèo = 1 + 0.5 × (r_nạn_nhân − r_mình)`, tối thiểu 1 |
| **Đột phá** cảnh giới mới (lần đầu trong mạng đó) | `20 × r_mới` |
| **Giữ Top 1 phòng** | +1 điểm mỗi giây |

- Ví dụ: Trúc Cơ (r=1) chém trúng Nguyên Anh (r=3) → `100 × 4 × (1 + 0.5×2) + 50 = 850`. Đại Thừa chém Luyện Khí → `100 × 1 × 1 + 50 = 150`.
- Nạn nhân là **bot** → nhận 50% điểm.
- **Chống cày điểm**: chém trúng cùng 1 người (cùng `guest_id`) lần nữa trong vòng 60s → 0 điểm.
- HUD hiển thị Uy Danh hiện tại; mỗi lần cộng điểm hiện số bay lên (+800 LẬT KÈO!).
- Màn hình Tổng kết hiển thị bảng chi tiết điểm theo từng nguồn.

### 11.2 Các bảng xếp hạng

- **Trong phòng** (realtime, 1 lần/giây): top 10 xếp theo **Uy Danh** của mạng hiện tại, tính **chung cả phòng (5 khu)**. Thiết kế bảng ở §11.5.
- **Toàn cục**, tab **Hôm nay / Tuần này / Mọi thời đại**, 3 bảng:
  - **Uy Danh** (bảng chính): điểm Uy Danh cao nhất trong 1 mạng.
  - **Sát Thần**: số lần chém trúng người nhiều nhất trong 1 mạng.
  - **Thần Tốc**: thời gian ngắn nhất từ lúc vào trận tới Đại Thừa (chỉ ghi những mạng đã đạt Đại Thừa; xếp tăng dần).
- Lưu mỗi "mạng" vào SQLite: `guest_id, name, character, score, hits, kills, peak_level, secs_to_dai_thua (null nếu chưa đạt), top1_secs, survive_secs, map, ended_at`.
- API: `GET /api/leaderboard?period=day|week|all&board=score|hits|speed` → top 50, cache 10s.
- Điểm do server tính → client không gian lận được. Bot không được ghi.

### 11.3 Danh hiệu Top 1–10 phòng

| Hạng | Danh hiệu | Màu |
|---|---|---|
| Top 1 | **Bần tăng chưa ngán ai bao giờ** + vương miện | Vàng kim |
| Top 2 | **Ngon thì nhào vô** | Bạc |
| Top 3 | **Chỗ này nhiều máu nè** | Đồng |
| Top 4–10 | Theo mảng `titles` trong config (giữ đúng bộ danh hiệu đang chạy trong code) | Trắng |

- **Danh hiệu tính theo cả PHÒNG, không theo khu**: xếp hạng Uy Danh của mọi người thật + bot trong cả 5 khu của phòng, Top 1 của phòng chỉ có **1 người**. Người ở khu khác vẫn thấy đúng danh hiệu của họ khi gặp. (Bất kỳ chỗ nào trong code đang tính hạng trong 1 khu → sửa.)
- Danh hiệu hiện **phía trên tên** nhân vật trong trận và trong bảng phòng (§11.5).
- Chống nhấp nháy: giữ hạng mới liên tục 3s mới đổi danh hiệu.
- Kill feed khi lên ngôi / mất Top 1 (như cũ). Người Ẩn Thân / trong bụi: danh hiệu ẩn theo.

### 11.5 Thiết kế bảng xếp hạng phòng (góc trên phải)

**Mỗi người hiển thị**: tên · danh hiệu · **khu đang đứng** · **cảnh giới + tầng** · Uy Danh. Chân dung nhân vật cho Top 3.

**Bố cục — "bục vinh danh" cho Top 3, danh sách gọn cho 4–10**

```
┌─────────────────────────────────┐
│ ★ TOP PHÒNG            19 người │
├─────────────────────────────────┤
│ 👑 ┌────┐ Mít Tơ Bean      4 931 │  ← TOP 1: thẻ to, viền vàng kim chuyển sắc,
│    │chân│ Bần tăng chưa ngán…    │    vệt sáng lướt 3s/lần, vương miện, chân dung 40px
│    │dung│ [KIM ĐAN ⁵] ⛩Thiên Đình │
│    └────┘                         │
├────────────────┬────────────────┤
│ 🥈 ┌──┐ Nhỏ Bông│ 🥉 ┌──┐ Tí Sún │  ← TOP 2 & 3: hai thẻ nửa chiều rộng,
│    └──┘ 1 420  │    └──┘ 1 070  │    viền bạc / đồng, chân dung 28px,
│ Ngon thì nhào vô│ Chỗ này nhiều… │    danh hiệu 1 dòng (cắt "…")
│ [TRÚC CƠ ⁹] 🏔LS│ [TRÚC CƠ ³] 🎋TL│
├────────────────┴────────────────┤
│ 4  Thằng Cu Đất   [LK⁸] 🏘PT   970│  ← 4–10: mỗi người 1 dòng gọn,
│    Chưa gặp đối thủ xứng tầm      │    danh hiệu dòng phụ chữ nhỏ mờ
│ 5  …                              │
├─────────────────────────────────┤
│ ▶ 14  Bạn (Cô Ba)  [LK²] ❄BN   120│  ← dòng của mình luôn ghim cuối (nếu ngoài top 10),
└─────────────────────────────────┘    nền nổi bật
```

**Top 1 / 2 / 3 phải khác hẳn nhau và khác hẳn phần còn lại**

| | Top 1 | Top 2 | Top 3 | 4–10 |
|---|---|---|---|---|
| Khung | Thẻ to cả chiều rộng, nền tối ánh vàng | Thẻ nửa rộng | Thẻ nửa rộng | 1 dòng |
| Viền | Vàng kim **chuyển sắc động** + vệt sáng lướt 3s/lần | Bạc, vệt sáng 5s/lần | Đồng, tĩnh | Không |
| Biểu tượng | 👑 vương miện vẽ (không dùng emoji hệ thống) | Huy chương bạc | Huy chương đồng | Số hạng |
| Chân dung | 40 px, viền màu cảnh giới | 28 px | 28 px | Không |
| Tên | 16 px đậm, vàng | 14 px đậm, bạc | 14 px đậm, đồng | 13 px trắng |
| Danh hiệu | Dòng riêng, chữ trang trí, vàng | 1 dòng, bạc | 1 dòng, đồng | Dòng phụ nhỏ, mờ 70% |
| Uy Danh | 18 px | 14 px | 14 px | 13 px |
| Hiệu ứng | Vài đốm lấp lánh quanh thẻ | – | – | – |

- **Cảnh giới**: viên thuốc nhỏ cùng màu với huy hiệu cảnh giới §12.3 (`KIM ĐAN ⁵`; Tiên giới viền chảy). Hẹp quá thì viết tắt (`KĐ⁵`).
- **Khu đang đứng**: icon khu + tên ngắn (Thiên Đình · Linh Sơn · Trúc Lâm · Phàm Trấn · Băng Nguyệt), màu icon theo khu; người cùng khu với mình được đánh dấu chấm nhỏ "cùng khu". Người đang **Ẩn Thân** → khu hiện "???".
- **Hoạt ảnh khi đổi hạng**: dòng trượt lên/xuống 0.25s; lên hạng có mũi tên ▲ xanh 2s, xuống ▼ đỏ 2s. Ai mới vào Top 1: thẻ Top 1 lật (flip) sang người mới + chớp vàng. Top 1 bị hạ: thẻ rung nhẹ trước khi đổi.
- **Dòng của mình**: nếu trong top 10 → nền sáng hơn + viền mảnh; nếu ngoài → ghim dòng cuối kèm hạng thật.
- **Không tràn màn hình**: rộng tối đa 300 px (PC), neo góc trên phải có lề 12 px; tên/danh hiệu dài cắt "…" (rê chuột/nhấn giữ thấy đủ).
- **Điện thoại**: mặc định **thu gọn** = 1 hàng: 3 chân dung tròn Top 1–3 (vương miện trên Top 1) + "Bạn #14". Chạm để mở bảng đầy đủ, chạm ngoài để đóng.
- Đa ngôn ngữ (§19): tên khu, cảnh giới, danh hiệu theo ngôn ngữ người xem.
- **Hiệu năng**: cập nhật 1 lần/giây, chỉ vẽ lại dòng thay đổi; thẻ Top 1–3 cache, hiệu ứng động là lớp chồng; chất lượng Thấp tắt vệt sáng/lấp lánh.
- **Dữ liệu từ server** (gói Leaderboard 1 Hz): mỗi dòng `{id, name, character, title_id, rank, zone_id, realm, tier, score, stealth}` + hạng thật của người nhận. Server tính hạng 1 lần/giây cho cả phòng.

### 11.4 Tin loan toàn server
- Khi một người **đang trong Top 5 phòng** bị chém trúng (rớt cảnh giới hoặc chết), **mọi phòng trên server** đều nhận tin trên kill feed, ví dụ: `"[Phòng 3] {A} vừa hạ {B} — 'Ngon thì nhào vô' — rớt xuống {realm}!"`.
- Giới hạn tối đa 1 tin/giây toàn server (gộp nếu nhiều), không hiện tin của chính phòng mình 2 lần.

## 12. Trạng thái nhân vật, animation & tấu hài

Server tính `mood` và gửi xuống (mọi người thấy giống nhau). Giữ mood tối thiểu 0.5s để không nhấp nháy.

| Anim | Khi nào | Tấu hài |
|---|---|---|
| `idle` | Đứng yên | Ngáp, gãi đầu |
| `run` | Chạy, không có địch gần | La toáng "AAAA" |
| `chase` | Đang tiến lại kẻ yếu hơn trong tầm | Cười nham hiểm, xoa tay |
| `flee` | Đang chạy khỏi kẻ mạnh hơn | Giơ tay hoảng loạn, mồ hôi văng |
| `hit` | Bị chém (0.6s) | Mặt bẹp, sao quay quanh đầu |
| `levelup` | Lên tầng | Tạo dáng, lóe sáng |
| `breakthrough` | Lên cảnh giới | Sét đánh, tóc dựng |
| `realm_drop` | Rớt cảnh giới | Ngồi bệt khóc |
| `teleport_out` / `teleport_in` | Dịch chuyển | Bị hút xoáy / văng ra chóng mặt |
| `death` | Chết | Dẹp lép như bánh tráng, hồn bay lên |
| `respawn` | Vào trận | Rơi từ trời xuống, cắm mặt, đứng dậy phủi bụi |

- Emote bubble khi: mất kiếm, khiên vỡ, nhặt Rương Báu, bị lộ khỏi bụi. Mỗi loại có cooldown.
- Client tự lật sprite trái/phải theo hướng chạy.

### 12.3 Bảng tên trên đầu (nameplate): Danh hiệu · Cảnh giới · Tên

**Vấn đề hiện tại**: 3 dòng chữ trơn nằm đè lên vòng kiếm nhiều màu → khó đọc, không có cấp bậc thị giác.

**Bố cục** (từ trên xuống, căn giữa trên đầu nhân vật):

```
      ╭──────────────────────────────╮
      │ 👑  Bần tăng chưa ngán ai bao giờ │   ← Ruy băng danh hiệu (chỉ Top 1–10)
      ╰──────────────────────────────╯
          ╭───────────────╮
          │ ☯ ĐẠO TỔ   ▲ │                ← Huy hiệu cảnh giới (dạng viên thuốc)
          ╰───────────────╯
          TestThThuTianThu                ← Tên (to nhất, đậm nhất)
```

Cả cụm có **1 tấm nền chung** bo góc, đen 45%, viền mảnh sáng 1px → chữ luôn nổi trên mọi nền (kiếm, cỏ, đá). Khoảng cách các dòng 2–4px, cụm cách đỉnh đầu 8px.

**1. Ruy băng danh hiệu** (chỉ Top 1–10)

| Hạng | Nền ruy băng | Chữ | Hiệu ứng |
|---|---|---|---|
| Top 1 | Ruy băng hai đuôi xẻ, đỏ son viền vàng | Vàng kim chuyển sắc (sáng → đậm), viền nâu đậm 2px | Vương miện nhỏ bên trái; **vệt sáng lướt qua chữ mỗi 3s**; 2–3 đốm lấp lánh quanh ruy băng |
| Top 2 | Ruy băng xanh đen viền bạc | Bạc chuyển sắc | Vệt sáng lướt qua mỗi 5s |
| Top 3 | Ruy băng nâu viền đồng | Đồng chuyển sắc | Không |
| Top 4–10 | Dải mảnh trong suốt 60% | Trắng ngà, chữ nhỏ hơn | Không |

- Lên hạng: ruy băng **bật ra** (scale 0 → 1.2 → 1.0, 0.35s) + đốm sáng tung ra. Mất hạng: ruy băng rách đôi rơi xuống (0.4s).

**2. Huy hiệu cảnh giới** (dạng viên thuốc bo tròn)
- Bên trái: biểu tượng nhỏ của cảnh giới (dùng lại 14 huy hiệu cảnh giới, thu nhỏ). Giữa: tên cảnh giới **in hoa**, chữ đậm; phàm giới thêm số tầng nhỏ phía sau dạng chỉ số (`KIM ĐAN ⁵`).
- Màu nền theo cảnh giới (chữ là tông đậm cùng màu hoặc trắng nếu nền tối):

| Cảnh giới | Màu | | Cảnh giới | Màu |
|---|---|---|---|---|
| Luyện Khí | Xám xanh `#8D99AE` | | Hợp Thể | Cam `#F77F00` |
| Trúc Cơ | Xanh lá `#57CC99` | | Độ Kiếp | Đỏ sấm `#E63946` |
| Kim Đan | Vàng đồng `#E9C46A` | | Đại Thừa | Đen, viền & chữ vàng `#FFD60A` |
| Nguyên Anh | Xanh lam `#4EA8DE` | | Tiên giới (5 cấp) | Màu hào quang tương ứng (§3.4) |
| Hóa Thần | Tím `#9D4EDD` | | | |
| Luyện Hư | Xanh ngọc `#2EC4B6` | | | |

- **Tiên giới**: viền huy hiệu là dải màu chảy vòng quanh (gradient xoay), chữ phát sáng nhẹ; Đạo Tổ viền cầu vồng đổi màu liên tục.
- **Chỉ báo nguy hiểm** (chỉ hiện trên người khác, bên phải huy hiệu): ▲ đỏ = cao cảnh giới hơn bạn · ● vàng = ngang · ▼ xanh = thấp hơn. Giúp quyết định né hay săn trong nửa giây.
- Lên tầng: số tầng lăn lên (0.3s). Đột phá: huy hiệu chớp trắng + phình to 1.3 rồi về.

**3. Tên người chơi**
- Dòng **to nhất**: chữ đậm, trắng, viền đen 3px + bóng đổ nhẹ phía dưới.
- **Tên của chính bạn**: vàng ấm + mũi tên nhỏ ▼ phía trên cụm (dễ tìm mình giữa đám đông).
- Bot hiển thị giống người (không lộ là bot).

**Kích thước cố định trên màn hình** (không phóng to theo scale nhân vật, không co theo camera):

| | PC | Điện thoại |
|---|---|---|
| Tên | 18 px | 15 px |
| Cảnh giới | 13 px | 11 px |
| Danh hiệu | 13 px (Top 1: 15 px) | 11 px (Top 1: 13 px) |

**Font**: phải hỗ trợ đầy đủ dấu tiếng Việt. Tên + cảnh giới: **Be Vietnam Pro** (ExtraBold / Bold). Danh hiệu: 1 font trang trí có bộ ký tự Vietnamese trên Google Fonts (lọc theo ngôn ngữ Vietnamese; agent tự kiểm tra hiển thị đủ dấu, ví dụ thử chuỗi "Bần tăng chưa ngán ai bao giờ"), nếu không có thì dùng Be Vietnam Pro Italic.

**Gọn màn hình (declutter)**
- Nhân vật nhỏ hơn 40 px trên màn hình: chỉ hiện **tên** (+ vương miện nếu Top 1).
- Hơn 12 bảng tên trên màn hình: ẩn ruy băng Top 4–10 của người ở xa.
- 2 bảng tên chồng nhau: bảng của người ở xa mờ 50%.
- Ẩn Thân / trong bụi: ẩn cả cụm (như cũ).

**Hiệu năng**
- Dựng mỗi bảng tên thành **1 texture cache** (vẽ lại chỉ khi nội dung đổi: hạng, cảnh giới, tầng). Hiệu ứng động (vệt sáng, viền chảy, đốm lấp lánh) là sprite riêng chồng lên, không vẽ lại texture.
- Chữ dùng BitmapFont sinh sẵn (gồm đủ ký tự tiếng Việt), không dùng `Text` thường.
- Chất lượng Thấp: tắt vệt sáng, đốm lấp lánh, viền chảy (giữ màu tĩnh).

### 12.4 Ảnh đại diện ở góc thông tin (thay số cấp)

Thay vòng tròn số "14" ở góc trên trái bằng **ảnh đại diện nhân vật**:
- **Chân dung** nhân vật đang chơi (cắt từ đầu tới vai), trong khung tròn 64 px (PC) / 52 px (điện thoại). Biểu cảm đổi theo trạng thái: bình thường · cười nham hiểm khi đang đuổi · hoảng khi bị chém / còn 1 vòng kiếm · khóc khi rớt cảnh giới (dùng 4 chân dung).
- **Viền khung = màu cảnh giới** (bảng §12.3). Tiên giới: viền là hào quang động theo màu §3.4; Đạo Tổ viền âm dương xoay.
- **Huy hiệu nhỏ** góc dưới phải khung: biểu tượng cảnh giới + số tầng (phàm giới) — thay cho con số to cũ.
- Vương miện nhỏ trên khung nếu đang Top 1; ruy băng hạng nhỏ nếu Top 2–10.
- Bị chém: khung rung + chớp đỏ. Đột phá: khung phát sáng + phình ra.
- Chưa có art: dùng frame `idle_00` của nhân vật, tự cắt phần đầu (vùng trên 45% sprite) làm chân dung tạm.

### 12.5 Hoạt ảnh liền mạch & tự nhiên

**Vấn đề**: các frame sinh bằng AI lệch nhau (to nhỏ, vị trí, dáng, nét vẽ khác nhau) → nhân vật "giật, nhảy lung tung". Mỗi mood (run / chase / flee) lại là 1 bộ frame AI riêng → đổi mood là đổi cả dáng người, nhìn như 3 nhân vật khác nhau.

**Giải pháp gồm 3 lớp — thứ tự làm: B → C → A** (B chữa gốc vì phần lớn điểm giật đến từ frame lệch nhau; C làm mượt phần còn lại; A cần art mới nên làm sau):

**A. Một thân, nhiều khuôn mặt** (bỏ việc mỗi mood một bộ chạy riêng)
- Mỗi nhân vật chỉ có **1 chu kỳ chạy** (`run`, 6 frame) và **1 chu kỳ đứng** (`idle`, 4 frame) cho phần thân.
- Biểu cảm (bình thường, la hét, cười nham hiểm, hoảng, khóc, choáng...) là **lớp mặt riêng** chồng lên đầu (sprite mặt 6–8 cái, cùng kích thước, cùng vị trí neo). Đổi mood = **chỉ đổi mặt**, thân vẫn chạy liền mạch.
- `chase` / `flee` khác `run` bằng **tư thế thủ tục** (procedural): chase ngả người tới trước 8° + hai tay xoa (lớp tay nếu có); flee ngả ra sau 6° + nảy cao hơn + giọt mồ hôi văng (particle).
- Anim một lần (hit, levelup, breakthrough, death, respawn, teleport, skill) vẫn là frame riêng nhưng phải qua bước chuẩn hóa B.

**B. Chuẩn hóa frame tự động** (`tools/atlas-pack`, chạy trên mọi frame trước khi ghép atlas)
1. **Neo chân**: tìm đường đáy của 2 bàn chân, đặt về y = 460, căn giữa x theo **trung điểm hai chân** (không theo hộp bao — tay vung ra làm hộp bao lệch).
2. **Cùng tỉ lệ cho cả animation**: đo chiều cao phần đầu+thân (bỏ tay chân) từng frame, lấy **trung vị**, co giãn mọi frame của animation về cùng chiều cao đó (không co giãn từng frame riêng lẻ).
3. **Neo đầu**: vị trí đầu giữa các frame liên tiếp không được lệch quá 4 px (ngoài nhịp nhún); lệch hơn → căn lại hoặc đánh dấu frame hỏng.
4. **Cùng bảng màu & nét viền**: lượng tử hóa về palette của nhân vật (ART_PROMPTS §5), chuẩn hóa độ dày viền.
5. **Điểm "độ giật"** — tách 2 nhóm:
   - **Anim lặp** (idle, run, chase, flee, float...): so từng cặp frame liên tiếp (lệch tâm, lệch chiều cao, khác màu) → điểm 0–100, đọc **điểm trung bình**. Ngưỡng đỏ 45 (tạm, chỉnh theo mắt chủ dự án).
   - **Anim một lần** (hit, death, levelup, breakthrough, teleport, skill...): những anim này **cố ý biến hình mạnh**, nên dùng điểm **"lạc lõng"** — bước nhảy của 1 frame lớn gấp mấy lần bước nhảy thông thường của chính animation đó. Ngưỡng đỏ 45.
   - Frame vượt ngưỡng → báo đỏ trong `assets:report`, chỉ đích danh frame nên sinh lại.
6. Xuất GIF xem trước + ảnh **onion-skin** (chồng mờ các frame) để duyệt bằng mắt.

**C. Bộ điều khiển hoạt ảnh lúc chạy (client)**
- **Chuyển mượt**: đổi animation có crossfade 80–120 ms (không cắt cứng).
- **Giữ pha chu kỳ**: run ↔ chase ↔ flee dùng chung chu kỳ thân nên **không reset về frame 0** khi đổi mood.
- **Tốc độ anim theo tốc độ chạy thật**: fps chu kỳ chạy = tốc độ di chuyển / độ dài bước chân → không bị "trượt chân" (chân bước chậm mà người lướt nhanh). Thiêu Đốt, Thần Hành, băng trơn đều tự khớp.
- **Thời gian giữ tối thiểu**: mood giữ ít nhất 0.5 s (đã có), biểu cảm mặt giữ ít nhất 0.8 s.
- **Quay đầu trái/phải**: không lật tức thì — co ngang 0.08 s (scale X 1 → 0.2) rồi lật và giãn lại; chỉ lật khi hướng đổi > 30° và giữ > 0.12 s (chống lật qua lại khi chạy dọc).
- **Chuyển động phụ thủ tục** (áp lên mọi anim): nhún theo bước chạy (sóng sin đồng bộ pha chân), nghiêng người theo hướng rẽ (tối đa 6°), squash & stretch khi dừng/đáp đất (±6%), thở nhẹ khi đứng. Biên độ nhỏ, đồng nhất cho mọi nhân vật → cảm giác cùng 1 "chất".
- **Không giật theo mạng**: trạng thái anim lấy từ vị trí đã **nội suy** (không từ gói tin thô); nhận mood mới từ server thì đợi tới mốc frame gần nhất mới chuyển.
- Debug `?debug=true` phím `J`: hiện tên anim, frame, pha, điểm độ giật của nhân vật đang chọn.

**Nâng cấp sau (nếu A–C vẫn chưa đủ)**: chuyển sang **rig cắt mảnh 2D** — từ 1 ảnh nhân vật chuẩn, cắt đầu / thân / 2 tay / 2 chân thành sprite riêng có điểm khớp, animation bằng xoay/dịch các mảnh trong code. Mọi frame dùng **đúng 1 bộ hình** nên tuyệt đối đồng nhất; đổi skin chỉ cần thay bộ mảnh.

### 12.2 Máu quái
- **Thanh máu nhỏ** trên đầu quái (không còn số to đè lên mặt quái). Chỉ hiện khi quái đã mất máu; đổi màu xanh → vàng → đỏ theo % máu.
- Art quái không cần chừa vùng trống giữa bụng nữa.

### 12.1 Minimap

- Góc **trên bên trái** (dưới thanh cảnh giới), hình vuông bo góc, nền mờ 60%. Kích thước: 180px trên PC, 120px trên điện thoại. Chạm/nhấn phím `M` → phóng to giữa màn hình (bấm lại để thu nhỏ).
- **Hiển thị:**
  - Địa hình tĩnh vẽ sẵn 1 lần từ bản đồ Tiled: tường không phá được, đá tảng, bụi rậm, bùn, đường gió (màu nhạt, đơn giản).
  - Cổng dịch chuyển: chấm tròn theo màu cặp/mạng cổng.
  - **Mình**: mũi tên chỉ hướng di chuyển, màu nổi bật.
  - Khung chữ nhật mảnh thể hiện vùng camera đang nhìn.
  - **Top 1 phòng**: icon vương miện (cả phòng đều thấy → khuyến khích đi săn).
  - **Rương Báu**: icon rương nhấp nháy khi đang có, kể cả lúc báo trước 10s.
- **Không hiển thị** người chơi/bot khác (trừ Top 1), quái, item, viên EXP → giữ yếu tố bất ngờ và phục kích. Top 1 đang Ẩn Thân hoặc trong bụi thì ẩn khỏi minimap.
- Vật cản phá được không cần vẽ (đổi trạng thái liên tục).
- Dữ liệu: vị trí mình lấy ở client; vị trí Top 1 và Rương Báu gửi kèm gói `Leaderboard` (1 lần/giây).

## 13. Kiến trúc kỹ thuật

### 13.1 Stack đã chọn

| Phần | Công nghệ | Lý do |
|---|---|---|
| Server | **Go** (bản ổn định mới nhất) | Hiệu năng cao, code dễ đọc, compile nhanh, ra 1 file chạy duy nhất, Claude viết Go rất ổn định |
| WebSocket | `github.com/coder/websocket` | Gọn, hiện đại |
| Mã hóa gói tin | MessagePack (`vmihailenco/msgpack/v5` ↔ `@msgpack/msgpack`) | Nhỏ hơn JSON nhiều, dễ debug hơn binary tự chế |
| Database | SQLite qua `modernc.org/sqlite` | Không cần cài DB server, không cần CGO |
| Client | **TypeScript + Vite + PixiJS** | Chạy ngay trên trình duyệt PC & điện thoại (đúng kiểu `.io`), PixiJS vẽ hàng nghìn sprite 2D mượt bằng WebGL |
| Âm thanh | `howler.js` | Xử lý được việc điện thoại chặn auto-play |
| Bản đồ | Tiled (`.tmj`) | Vẽ map bằng kéo thả, không cần code |

Nguyên tắc: **server quyết định mọi thứ** (authoritative). Client chỉ gửi hướng di chuyển và vẽ lại những gì server gửi.

### 13.2 Cấu trúc repo
```
kiem-xoay/
├── CLAUDE.md
├── docs/            GAME_SPEC.md · ART_PROMPTS.md · SETUP_GUIDE.md · PROGRESS.md
├── config/game.json                 # mọi con số cân bằng; server đọc, client lấy qua GET /api/config
├── maps/            lang_tre.tmj · cho_dem.tmj · nui_da.tmj
├── server/
│   ├── go.mod
│   ├── cmd/server/main.go           # HTTP + WebSocket + phục vụ file client đã build
│   ├── cmd/sim/main.go              # chạy giả lập không giao diện với N bot để cân bằng game
│   └── internal/
│       ├── game/      # luật chơi thuần: entity, tick, va chạm, exp, kiếm, buff (có unit test)
│       ├── world/     # bản đồ, vật cản, vùng, cổng, lưới tìm đường
│       ├── bot/       # AI
│       ├── room/      # phòng, matchmaking, bot fill
│       ├── netproto/  # định nghĩa message
│       └── store/     # SQLite, leaderboard
├── client/
│   ├── package.json · vite.config.ts · index.html
│   └── src/
│       ├── main.ts
│       ├── net/       # kết nối WebSocket, giải mã, nội suy
│       ├── render/    # PixiJS: nhân vật, kiếm, bản đồ, VFX
│       ├── ui/        # trang chủ, HUD, màn hình chết, leaderboard, joystick
│       └── audio/
└── assets/          # xem ART_PROMPTS.md
```

### 13.3 Mạng

> **Đã được thay thế bởi `NETCODE_SPEC.md` (Protocol v2, cache nhiều lớp, đồng bộ theo khu).** Nội dung dưới đây chỉ giữ để tham khảo; chỗ nào mâu thuẫn thì theo NETCODE_SPEC.

- Tick server 30 Hz, mỗi phòng 1 goroutine. Snapshot gửi client 20 Hz.
- Client → Server: `Join{name, character, guest_id}`, `Input{seq, angle_u8, moving}` (30 Hz), `Ping`.
- Server → Client: `Welcome`, `Snapshot`, `Event`, `Leaderboard` (1 Hz).
- **Không gửi vị trí từng thanh kiếm**: gửi `phase, dir, n, ω, scale, phase_tick` (client tự tính số vòng, chiều dài từng vòng); client tự tính góc. Đổi chiều → event `RingFlip`.
- Event: `Hit, SwordBreak, Death, LevelUp, Breakthrough, RealmDrop, RingFlip, ObstacleDamaged, ObstacleBroken, ObstacleRegrow, TeleportOut, TeleportIn, ItemPicked, BuffStart, BuffEnd, ChestIncoming, KillFeed`.
- AOI: chỉ gửi những gì nằm trong vùng nhìn của người đó (+ lề). Người **trong bụi** không được gửi cho người không được thấy (chống hack nhìn xuyên). Người **Ẩn Thân** vẫn được gửi (vì vẫn hiện mờ 20%), kèm cờ `stealth`.
- Client: nội suy người khác (trễ 100ms); tự dự đoán chuyển động của mình + hiệu chỉnh theo server.

### 13.4 Va chạm (server)
- Lưới không gian (spatial hash) cho thân, quái, item, viên EXP, vật cản. Vật cản tĩnh index sẵn.
- **Không bao giờ duyệt từng thanh kiếm** để tìm va chạm. Mỗi vòng là 1 vật thể cứng (cùng tốc độ, cùng pha) → tính theo **vòng**: kiểm tra mục tiêu có nằm trong vành khăn của vòng không, rồi tính thẳng chỉ số kiếm gần nhất theo góc (O(1)), cộng cung quét trong tick.
- Kiếm–kiếm: xét va chạm cho **các vòng ngoài cùng tính từ ngoài vào cho tới khi gộp đủ số kiếm của 1 vòng đầy** (vd vòng ngoài chỉ còn 2 kiếm lẻ thì xét thêm vòng kế trong). Lý do: nếu chỉ xét đúng vòng ngoài cùng, lúc vòng đó còn vài kiếm lẻ thì gần như không chạm được nhau → mài tắc (lỗi gameplay đã phát hiện). Kiếm bể vẫn trừ từ vòng ngoài cùng trước. Các vòng sâu hơn bỏ qua.
- Kiếm–quái / vật cản / viên EXP: chỉ dùng vòng tròn `reach` để lọc qua lưới, rồi kiểm tra vành khăn.
- Không cấp phát bộ nhớ mỗi tick (dùng lại slice, `sync.Pool` cho buffer gói tin).

### 13.6 Ngân sách hiệu năng & tối ưu (bắt buộc)

**Mục tiêu** (phòng đầy 40 nhân vật, trong đó 10 ở Tiên giới đủ kiếm ~1 000 thanh):

| Chỉ số | Mục tiêu |
|---|---|
| Server: thời gian 1 tick / phòng | < 4 ms (tick 33 ms) |
| Gói snapshot / client | < 8 KB, tổng < 40 KB/s |
| Client: FPS máy tầm trung / điện thoại tầm trung | ≥ 60 / ≥ 40 |
| Ping hiển thị (không tính địa lý) | tăng < 20 ms so với phòng trống |

**Đo trước khi sửa**
- Phím `F3` (chỉ với `?debug=true`): FPS, ms/frame, số sprite đang vẽ, số draw call, KB/s nhận, kích thước snapshot, tick server ms (server gửi kèm), số entity trong AOI.
- **Ping phải đo đúng**: server gắn timestamp vào `Pong`, client trả lời ngay trong handler WebSocket (hoặc Web Worker), không chờ vòng lặp vẽ. Nếu ping chỉ tăng khi FPS tụt → nghẽn ở client, không phải mạng.
- `cmd/loadtest`: mở N client giả (headless) vào 1 phòng, ép M người lên Đạo Tổ đủ kiếm, in tick ms, KB/s, p95 ping.

**Client (thường là thủ phạm chính)**
1. **Mỗi vòng kiếm vẽ thành 1 texture, rồi chỉ xoay texture đó.** Vòng kiếm quay như 1 khối cứng → vẽ toàn bộ kiếm của vòng (kể cả nghiêng, so le, màu, spin blur) vào 1 `RenderTexture` **một lần**, mỗi frame chỉ đổi `rotation` của 1 sprite. 1 050 kiếm → 14 sprite. Chỉ vẽ lại texture khi số kiếm của vòng đổi (giới hạn tối đa 5 lần/giây/người), khi đổi chiều (lật tilt), khi đổi skin/màu.
2. **LOD theo kích thước trên màn hình**: kích thước texture vòng theo đường kính thực trên màn hình (tối đa 2048). Vật nhỏ hơn 3 px trên màn hình (viên EXP khi camera lùi xa) → không vẽ; thay bằng lớp "bụi sáng" chung.
3. **Cull**: không vẽ gì ngoài màn hình (+ lề).
4. Chữ (tên, danh hiệu, số điểm bay lên) dùng `BitmapText`, không dùng `Text` (Text tạo texture mới mỗi lần đổi chữ). Bảng xếp hạng/HUD cập nhật tối đa 1–2 lần/giây.
5. Viên EXP, item, hạt VFX dùng `ParticleContainer` + 1 atlas → ít draw call.
6. Giải mã MessagePack trong **Web Worker**, luồng chính chỉ nhận dữ liệu đã giải mã.
7. Không tạo object mới mỗi frame (pool cho sprite, particle, vector).
8. **Chất lượng đồ hoạ: mặc định + tự điều chỉnh**
   - **Mặc định khi vào game**: PC có GPU rời → **Vừa**; điện thoại, máy tính bảng, PC chỉ có GPU tích hợp → **Thấp**. Không bao giờ bắt đầu ở Cao.
   - **Tự hạ**: FPS < 45 liên tục 3s → hạ 1 bậc (Cao → Vừa → Thấp). Ở Thấp vẫn < 30 → thêm bước khẩn: resolution 1.0, tắt ambient, giảm decal.
   - **Tự nâng** khi FPS ≥ 55 liên tục: **Thấp → Vừa sau 10s**, **Vừa → Cao sau 20s**. Sau mỗi lần nâng mà bị hạ lại → khoá không tự nâng trong 5 phút (chống lên xuống liên tục).
   - Người chơi tự chọn mức (phím V / cài đặt) → tắt tự điều chỉnh cho tới khi chọn lại "Tự động".
   - **Chỉ lựa chọn tay của người chơi được lưu** (localStorage); mức do máy tự đổi không lưu — lần sau vào game lại bắt đầu từ mặc định.
   - Khác biệt chính giữa Cao và Vừa: mây, lá bay, bloom, vệt hạt thần thú (đo 31/10: Cao tốn hơn Vừa ~33%).

**Server & mạng**
9. Lượng tử hóa: vị trí `u16` (bản đồ 12 000 → bước ~0.2), góc `u8`/`u16`, scale `u8`. Không gửi float64.
10. **Delta snapshot**: chỉ gửi trường thay đổi so với snapshot client đã xác nhận; entity đứng yên không gửi lại.
11. Viên EXP/Tiên khí/item: chỉ gửi sự kiện xuất hiện/biến mất, không nằm trong snapshot định kỳ. Camera lùi xa (Tiên giới) → **không gửi viên EXP nhỏ** ngoài bán kính 1 500 quanh người chơi (dù nằm trong khung nhìn), chỉ gửi viên vừa/lớn, Tiên khí, item.
12. Tần suất theo khoảng cách: entity gần (trong 1/2 khung nhìn) 20 Hz, xa 10 Hz.
13. Giới hạn tối đa 150 entity động mỗi snapshot, ưu tiên gần và to.
14. Kiểm tra nghẽn ghi: nếu bộ đệm gửi của 1 client đầy (mạng yếu) → bỏ snapshot cũ, chỉ gửi bản mới nhất (không xếp hàng làm trễ dồn).

**Ghi chú**: nếu mọi người đều ping ~200 ms kể cả phòng trống thì nguyên nhân là **khoảng cách địa lý** tới server (ví dụ người chơi ở Việt Nam, server ở Mỹ/Âu) → cần đặt server gần người chơi (Singapore / Việt Nam), không phải tối ưu code.

### 13.5 An toàn cơ bản
- Kiểm tra & giới hạn mọi input từ client; rate limit tin nhắn và số kết nối mỗi IP.
- Lọc tên tục, giới hạn độ dài.


### 13.7 Dung lượng tải & màn hình tải (bắt buộc)

**Vấn đề**: lần tải đầu ~1 GB, chờ rất lâu. Nguyên nhân gần như chắc chắn: ảnh art gốc (PNG 512–2048 px, hàng nghìn frame) được tải nguyên trạng, và tải **tất cả** trước khi vào trận.

**Ngân sách**

| Mục | Ngân sách |
|---|---|
| Gói tải ban đầu (vào được trận) | **≤ 8 MB** |
| Toàn bộ game (tải dần trong lúc chơi) | ≤ 80 MB |
| Lần vào thứ 2 trở đi | gần như 0 (cache trình duyệt) |
| Thời gian vào trận, 4G ~10 Mbps, lần đầu | ≤ 10 s |

Thêm bước CI `npm run assets:budget`: vượt ngân sách → build fail, in 20 file nặng nhất.

**1. Đo trước** — `npm run assets:size`: tổng dung lượng theo thư mục, top 30 file nặng nhất, tổng số file. Ghi vào PERF_LOG trước khi sửa.

**2. Pipeline xử lý art lúc build** (`tools/asset-pipeline`, chạy tự động trước `npm run build`; art gốc giữ ở `assets/src/`, game chỉ dùng `assets/dist/`):
- **Thu nhỏ về đúng kích thước hiển thị**: trên màn hình nhân vật cao tối đa ~150 px → frame nhân vật lưu ở **256 px** (đủ nét màn hình độ phân giải cao), VFX theo kích thước hiển thị, icon 64–128 px, vật cản theo kích thước thực trong game. Không bao giờ phát hành ảnh 1024–2048 px trừ texture nền lặp.
- **Cắt viền trong suốt** (trim) và ghi lại offset để giữ pivot.
- **Ghép atlas**: mỗi nhân vật 1–2 atlas 2048×2048 chứa mọi anim; item/icon/decal gom chung atlas. Hàng nghìn file nhỏ → vài chục file.
- **Nén WebP** (có alpha, chất lượng ~80) thay PNG — thường nhỏ hơn 5–10 lần. Texture nền lặp: WebP chất lượng ~75.
- Âm thanh: OGG/MP3 96 kbps mono cho SFX, nhạc nền 128 kbps và phát dạng stream.
- Tên file kèm hash nội dung (`a_hon.3f9c2e.webp`) → server gửi header `Cache-Control: public, max-age=31536000, immutable`; HTML/manifest không cache lâu. JS/JSON nén gzip/brotli.
- Pipeline xuất `assets/dist/manifest.json` (đường dẫn, dung lượng byte, nhóm tải) — loader đọc file này.

**3. Tải theo nhóm (lazy loading)**

| Nhóm | Gồm | Khi nào tải |
|---|---|---|
| `boot` | Font, UI, HUD, nhân vật **mình đã chọn**, vật phẩm, viên EXP, kiếm cảnh giới thấp, nền + vật cản của **khu được xếp vào** | Trước khi vào trận (có thanh tiến độ) |
| `zone_*` | Nền, decal, vật cản, ambient của từng khu còn lại | Tải nền ngay sau khi vào trận; ưu tiên khu kề cổng gần nhất |
| `chars` | 9 nhân vật còn lại | Tải nền; nhân vật nào xuất hiện trước thì tải trước |
| `fx_high` | VFX Tiên giới, thần thú, hào quang, skin kiếm cảnh giới cao | Tải nền; ưu tiên khi có người Tiên giới trong khu |
| `audio` | Nhạc nền, SFX ít dùng | Tải nền |

- Asset chưa tải xong: dùng placeholder đã có, đổi sang ảnh thật khi tải xong (mờ dần 0.2s, không nhấp nháy).
- Qua cổng sang khu chưa tải xong: màn hình dịch chuyển giữ lâu hơn (tối đa 2s) kèm % tải của khu đó; quá 2s thì vào với placeholder.

**3b. Thư viện art (không xoá art đã sinh)**
- Art do `artgen` sinh ra là tài nguyên trả tiền → **không bao giờ xoá**.
- Art không còn dùng (hoạt ảnh vỡ cũ, buff đã bỏ, biến thể thừa, art chưa có chỗ dùng) → chuyển vào **`assets/library/<nhóm>/`**, ghi vào `assets/library/INDEX.md`: tên tệp, lý do ngừng dùng, ngày, job/prompt gốc, phiên bản spec liên quan.
- `assets/library/` **không** nằm trong pipeline build: không vào `assets/dist`, không vào manifest, không tải cho người chơi, không tính vào ngân sách dung lượng.
- `check:dead` khi thấy art không ai nạp: **đề xuất chuyển vào thư viện**, không bao giờ đề xuất xoá.
- Lấy lại: chuyển từ `assets/library/` về `assets/src/` rồi build như bình thường.

**4. Màn hình tải có thanh tiến độ**
- Hiện ngay khi mở trang (HTML + CSS thuần, < 50 KB, không đợi PixiJS).
- **Thanh tiến độ theo byte thật** (tổng dung lượng nhóm `boot` lấy từ manifest), hiện `%` và `x,x / y,y MB`. Không dùng % theo số file (file nặng nhẹ khác nhau làm thanh chạy giật).
- Dòng trạng thái theo bước: "Đang tải giao diện…" → "Đang gọi nhân vật…" → "Đang dựng {tên khu}…" → "Đang tìm phòng…".
- Trang trí tấu hài: nhân vật mình chọn chạy trên thanh tiến độ (dùng frame chạy nhỏ tải đầu tiên), mẹo chơi / câu bựa đổi mỗi 4s.
- Lần vào sau (đã có cache): thanh chạy vọt lên gần như ngay lập tức.
- Lỗi mạng: hiện nút "Thử lại", tải tiếp phần còn thiếu (không tải lại từ đầu).
- Trong trận: icon nhỏ góc dưới hiện % các nhóm tải nền đang chạy; xong thì ẩn.

## 14. Quái (PvE)

Quái **không gây sát thương**, đi lang thang, chạy khi bị chém, hiện số HP trên người.

| Loại | HP | EXP | Drop item | Weight |
|---|---|---|---|---|
| Nhỏ | 3 | 10 | 10% | 60 |
| Vừa | 8 | 30 | 25% | 30 |
| Lớn | 20 | 90 | 60% | 9 |
| Boss mini | 80 | 400 | 100% (3 item) | 1, tối đa 1/phòng |

Duy trì ~150 quái/phòng (bản đồ lớn hơn v1).

## 15. Milestones

Mỗi milestone là 1 phiên làm việc với Claude Code (prompt mẫu trong `SETUP_GUIDE.md`).

| # | Nội dung | Xong khi |
|---|---|---|
| M0 | Tạo repo, server Go "hello", client Vite + PixiJS kết nối WebSocket | Mở trình duyệt thấy chấm tròn di chuyển theo phím, vị trí do server gửi |
| M1 | Vòng kiếm, quái, item kiếm, EXP, 9 cảnh giới, scale + camera. Hình placeholder | Chơi 1 mình, lên Trúc Cơ thấy to ra, kiếm dài ra, camera lùi ra |
| M2 | PvP: chém thân, rớt tầng, bể kiếm theo cảnh giới, chết → màn hình Tổng kết → chơi tiếp | Mở 2 tab đánh nhau được |
| M3 | Phòng, matchmaking, bot fill, bot AI cơ bản, trang chủ nhập tên, leaderboard SQLite | 1 tab = 1 phòng có 19 bot; trang leaderboard có dữ liệu |
| M4 | Bản đồ Tiled: vật cản phá được + mọc lại, bụi, bùn, gió, hẻm, cổng; bot biết dùng địa hình | Chơi được trên 3 bản đồ |
| M5 | Toàn bộ vật phẩm buff + Rương Báu; netcode mượt (nội suy, dự đoán, AOI, ẩn thân) | Giả lập ping 150ms vẫn mượt |
| M6 | Ghép art, animation, mood, VFX, SFX, UI hoàn chỉnh | Đủ anim mục 12 |
| M7 | Tối ưu điện thoại, joystick, deploy lên VPS có HTTPS | Bạn bè vào được bằng link |

Sau mỗi milestone: chạy `go run ./cmd/sim` để xem cân bằng (thời gian trung bình lên mỗi cảnh giới, số kiếm trung bình, số lần chết).

## 16. Config mặc định (`config/game.json`)

```json
{
  "cultivation": {
    "realms": ["Luyện Khí","Trúc Cơ","Kim Đan","Nguyên Anh","Hóa Thần","Luyện Hư","Hợp Thể","Độ Kiếp","Đại Thừa"],
    "tien_realms": ["Chân Tiên","Kim Tiên","Thái Ất Ngọc Tiên","Đại La Kim Tiên","Đạo Tổ"],
    "tien_khi_ceilings": [1000, 3000, 6000, 10000, 15000], "tien_entry_requires_full_exp": true,
    "tien_rob_share": 0.5, "tien_drop_share": 0.3,
    "tien_scale_step": 0.2, "tien_overflow_score": 0.1,
    "tien_aura_colors": ["#E8F1FF", "#FFD60A", "#4CE0B3", "#B15CFF", "yin_yang"],
    "tiers_per_realm": 10,
    "exp_base": 100, "realm_mult": 2.0, "tier_step": 0.08,
    "scale_min": 1.0, "scale_max": 5.0
  },
  "sword": {
    "start_count": 1, "min_count": 0,
    "max_swords": [10, 30, 60, 100, 150, 210, 280, 360, 450, 550, 660, 780, 910, 1050],
    "ring_capacity_step": 10, "length": 80, "width": 12, "hilt_gap": 6,
    "ring_growth": 0.2, "ring_growth_mode": "linear", "hilt_step": 0.1, "tilt_deg": 35,
    "ring_tints": ["#FFFFFF", "#C77DFF", "#7CFC6A", "#FF6B6B", "#4CC9F0", "#FFD60A"],
    "rev_secs": 1.0, "flip_cooldown": 0.3,
    "clash_unit": 1, "flip_only_same_realm": true, "clash_knockback": 0, "clash_cooldown": 0, "body_hit_max_rings": 1, "drop_bundle_max": 10,
    "durability_per_realm": 1.2, "durability_mode": "compound"
  },
  "speed": { "mode": "additive", "cap": 2.0, "cap_gom_don": 3.0, "floor": 0.4, "haste": 0.4, "burn": 0.8, "gom_don": 2.0, "wind_with": 0.5, "wind_against": -0.2, "mud": -0.4, "snow": -0.2 },
  "player": {
    "body_r": 30, "move_base": 195, "move_scale_sqrt": false, "pickup_r": 90,
    "invuln_on_hit": 1.0, "invuln_on_spawn": 5.0,
    "burn_speed_mult": 1.8, "burn_exp_per_sec": 0.08, "burn_tien_frac_per_sec": 0.008,
    "attacker_share": 0.3, "drop_share": 0.3, "hit_kills": true, "rob_cap_realms": 2, "rob_cap_tien_realms": 1, "rob_share_bot_victim": 0.15, "kill_bonus_exp": 30,
    "death_sword_drop": 0.5, "orb_owner_secs": 1.5,
    "camera_view_h": 900, "camera_exp": 0.85, "camera_reach_mult": 2.6,
    "hit_drops_realm": true
  },
  "buffs": {
    "spin":   { "bonus": 0.6, "secs": 12 },
    "haste":  { "mult": 1.4, "secs": 10 },
    "magnet": { "bonus": 0.5, "secs": 15 },
    "hard":   { "durability_bonus": 0.5, "damage_bonus": 1.0, "secs": 10 },
    "shield": { "secs": 8 },
    "exp2":   { "bonus": 1.0, "secs": 20 },
    "stealth":{ "secs": 5, "max_secs": 5, "visible_alpha": 0.2, "reveal_r": 120, "bot_detect_mult": 0.5 }
  },
  "items": {
    "weights": { "sword": 40, "bundle": 6, "spin": 12, "haste": 12, "magnet": 8, "hard": 7, "shield": 7, "exp2": 5, "stealth": 3 },
    "sword_item_amount": 3, "bundle_amount": 15,
    "count_per_room": 200, "respawn_secs": 3, "overcap_exp": 20, "chest_every_secs": 60, "chest_items": 8
  },
  "stat_stacking": "additive_on_base",
  "character_traits": {"a_hon": {"move": -0.05, "pickup": 0.1, "spin": -0.05, "durability": 0.0}, "ba_tam_loa": {"move": -0.05, "pickup": -0.05, "spin": 0.0, "durability": 0.1}, "ong_ba_vong": {"move": -0.1, "pickup": 0.0, "spin": -0.05, "durability": 0.15}, "dai_ca_dep": {"move": -0.1, "pickup": 0.0, "spin": 0.05, "durability": 0.05}, "map_tra_sua": {"move": -0.1, "pickup": 0.15, "spin": 0.0, "durability": -0.05}, "co_ba_ganh": {"move": 0.1, "pickup": 0.05, "spin": -0.05, "durability": -0.1}, "su_phu_mi": {"move": 0.0, "pickup": -0.05, "spin": 0.1, "durability": -0.05}, "be_ti": {"move": -0.05, "pickup": -0.1, "spin": 0.05, "durability": 0.1}, "shipper": {"move": 0.15, "pickup": 0.0, "spin": -0.05, "durability": -0.1}, "cu_rua": {"move": -0.15, "pickup": 0.0, "spin": 0.0, "durability": 0.15}},
  "pickup": { "base": 90, "scale_term": "sqrt_minus_1", "cap": 4.0 },
  "exp_economy": { "orb_value_mult": 1.5, "pickup_scale_mode": "sqrt", "monster_value_mult": 1.5, "rob_share_bot_victim": 0.15, "sim_tolerance": 0.25, "magnet_max_speedup": 1.5 },
  "exp_orbs": {
    "cell": 140, "values": [5, 20, 60], "weights": [80, 18, 2],
    "respawn_secs": 5, "respawn_region": 1000
  },
  "obstacles": {
    "footprint_w": 0.9, "footprint_h": 0.3, "occlusion_alpha": 0.5,
    "regrow_secs": 30, "regrow_warn_secs": 3, "hard_buff_extra_tier": 1, "cut_hint_cooldown_secs": 3,
    "corridor_min": { "1": 150, "2": 160, "3": 180, "5": 240, "7": 290, "9": 340, "11": 370, "13": 390 },
    "min_realm": { "fence": 0, "stall": 1, "cart": 1, "bamboo_thin": 1, "brick": 2, "bamboo_wall": 2, "boulder": 3, "bamboo_ancient": 5, "stone_wall": 5, "sect_wall": 7, "cliff": 7, "boulder_huge": 9, "ice_eternal": 9, "jade_pillar": 11, "seal_stele": 13 },
    "stall_item_drop": 0.3
  },
  "zones": { "mud_mult": 0.6, "wind_with": 1.5, "wind_against": 0.8, "wind_with_max_deg": 60, "wind_against_min_deg": 120, "bush_reveal_r": 150, "bush_reveal_on_hit_secs": 2 },
  "gates": { "radius": 150, "transit_secs": 0.8, "exit_warn_secs": 0.5, "cooldown_secs": 8, "per_zone": 4, "random_exit": true, "exit_pool": "all",
    "heaven_gate": { "enabled": true, "radius": 200, "landing_r_from_center": 1400, "landing_clear_r": 400, "zone_center_plaza_r": 500, "map": { "linh_son_coc": "north", "truc_lam_me_cung": "east", "pham_tran": "south", "bang_nguyet_ho": "west" } } },
  "titles": ["Bần tăng chưa ngán ai bao giờ", "Ngon thì nhào vô", "Chỗ này nhiều máu nè"],
  "title_hold_secs": 3,
  "minimap": { "size_pc": 180, "size_mobile": 120, "show_top1": true, "show_chest": true, "show_others": false },
  "score": {
    "hit_base": 100, "kill_bonus": 50, "upset_per_realm": 0.5,
    "breakthrough_per_realm": 20, "top1_per_sec": 1,
    "bot_victim_mult": 0.5, "same_victim_cooldown_secs": 60
  },
  "altars": {
    "small": { "count": 4, "every_secs": 5, "orbs": [2, 3], "orb_value": 2, "burst_every_secs": 30, "burst_small": 15, "burst_medium": 2, "medium_value": 10, "warn_secs": 3 },
    "center": { "every_secs": 150, "orb_value": 150, "warn_secs": 10 },
    "spray_r": [300, 700], "orb_lifetime_secs": 60, "clear_r": 300
  },
  "mapgen": {
    "center_plaza_r": 900, "sect_r": 700, "sect_inset": 0.15, "road_w": 500, "ring_road_r": 3500,
    "symmetry": "rot4", "jitter": 150, "min_open_ratio": 0.65, "max_retries": 10
  },
  "bots": { "spawn_like_new_player": true, "mode_default": "auto", "auto_off_hysteresis": 5, "withdraw_idle_secs": 5 },
  "room_env": { "production": { "max_humans": 50, "zone_min_chars": 10, "bot_auto_off_at": 40 }, "staging": { "max_humans": 20, "zone_min_chars": 10, "bot_auto_off_at": 20 } },
  "zones_def": ["thien_dinh", "linh_son_coc", "truc_lam_me_cung", "pham_tran", "bang_nguyet_ho"],
  "ice": { "friction": 0.92, "accel": 0.25, "snow_mult": 0.8 },
  "skills": { "key": "Space", "source": "cooldown", "start_charges": 1, "tien_extra_charges": 1, "lockout_secs": 3,
    "cooldowns": { "a_hon": [30, 2], "ba_tam_loa": [18, 2], "ong_ba_vong": [25, 2], "dai_ca_dep": [40, 1], "map_tra_sua": [20, 2], "co_ba_ganh": [22, 2], "su_phu_mi": [30, 2], "be_ti": [35, 1], "shipper": [20, 2], "cu_rua": [30, 2] },
    "ranges_u": { "loa_radius": 6, "loa_shape": "circle", "loa_push": 6, "ru_ngu_r": 4, "ru_ngu_steal_secs": 4, "ru_ngu_secs": 2.5, "ru_ngu_channel": 0.5, "dan_em_follow": 3, "dan_em_share": 0.3, "dan_em_secs": 8, "gong_co_secs": 4, "dan_em_seek": 8, "tran_chau_cone": 8, "bay_sap_dist": 3, "bay_sap_len": 6, "bay_sap_count": 5, "bay_sap_secs": 6, "bay_sap_tier_bonus": 1, "bay_sap_owner_passes": true, "vo_chuoi_w": 1, "hut_mi_cone": 5, "hut_mi_steal": 0.25, "hut_mi_cap_of_own_max": 0.2, "gom_don_pickup_bonus": 0.1, "gom_don_secs": 4, "rut_co_wave": 2, "rut_co_push": 3 },
    "push_secs": 0.3, "lockout_secs": 3, "overcap_exp": 20, "overcap_tien_khi": 2, "blind": { "secs": 5, "no_pickup": true, "pearls": [25, 40], "cover": [0.55, 0.65], "bot_vision_mult": 0.4 } },
  "movement": { "slide_iterations": 3, "smooth_correct_max": 30, "smooth_correct_ms": 100 },
  "room": {
    "max_humans": 50, "zone_min_chars": 10, "zones": 5, "map_size": 12000,
    "monster_count": 150, "empty_close_secs": 60, "safe_spawn_r": 800
  },
  "net": { "tick_hz": 30, "snapshot_hz": 20, "interp_ms": 100 }
}
```

Độ bền kiếm: `durability_mode: "compound"` → `k = 1.2^d` (mặc định). Đổi sang `"linear"` → `k = 1 + 0.2 × d` nếu muốn khoảng cách giữa cảnh giới thấp và cao nhỏ hơn nữa (Đạo Tổ vs Luyện Khí chỉ 3.6 : 1).

## 18. Chế độ debug

- **Cần đăng nhập admin trước** (ADMIN_SPEC §1.1): `?debug=true` chỉ có tác dụng khi trình duyệt có phiên admin hợp lệ; server mới là bên cấp quyền debug, mọi lệnh debug từ kết nối không có quyền bị từ chối. Mạng chơi dùng debug không lên bảng xếp hạng toàn cục.
- Mở game với `?debug=true` trên URL → bật phím debug: `E` +100 EXP, `R` lên cảnh giới, `K` +1 kiếm, `1–7` bật buff, `X` tự chết, **`B` thả 1 bot ngay cạnh mình**, phím vẽ hitbox capsule. Không có `?debug=true` **hoặc không có phiên admin** thì mọi phím debug bị tắt (cả client lẫn server từ chối). Ngoại lệ duy nhất: `APP_ENV=dev` trên localhost.
- Có thể tắt hẳn debug trên server production bằng biến môi trường.

### 18.1 Bảng hướng dẫn điều khiển lúc vào trận
- Vừa vào trận: hiện bảng mờ (nền 50%) giữa-dưới màn hình liệt kê điều khiển (di chuyển; Thiêu Đốt: Shift / nháy đúp giữ / 🔥; **Kỹ năng: Space / nút kỹ năng** (tự hồi chiêu); M minimap...). Tự mờ dần sau 5s (trùng thời gian bất tử) hoặc biến mất khi người chơi bắt đầu di chuyển.


## 19. Đa ngôn ngữ (i18n)

**Ngôn ngữ đợt 1**: Tiếng Việt (gốc) · **English** · **Português (Brasil)** · **Español** · **Bahasa Indonesia** · **ไทย (Thái)**. Đây là các thị trường game .io trên trình duyệt đông người chơi; đợt 2 cân nhắc Русский, Türkçe, 日本語, 한국어 theo telemetry quốc gia.

**Cơ chế**
- Toàn bộ chữ người chơi thấy nằm trong `client/src/i18n/{lang}.json` (gồm cả chuỗi tấu hài, kill feed, danh hiệu Top 1–10, mẹo màn hình tải, tên nhân vật, tên kỹ năng, tên cảnh giới, tên khu, lời nhắc lỗi). Không chuỗi cứng nào trong code. CI: thiếu khoá ở ngôn ngữ nào → cảnh báo, rơi về English (rồi tới Tiếng Việt).
- Chọn ngôn ngữ: tự theo `navigator.language` lần đầu → lưu localStorage; nút chọn cờ ở trang chủ; URL `?lang=en` ghi đè.
- Server gửi **mã** sự kiện (vd `kill.realm_drop {a, b, realm}`), client dịch — kill feed / tin loan toàn server hiện đúng ngôn ngữ từng người xem.
- Bảng tên, danh hiệu, cảnh giới: dịch theo ngôn ngữ **người xem**. Tên người chơi giữ nguyên chữ họ nhập (mọi bảng chữ cái).
- Font: Be Vietnam Pro cho chữ Latin (đã đủ dấu Việt, Bồ, Tây Ban Nha, Indonesia); thêm **Noto Sans Thai** cho tiếng Thái. BitmapFont sinh sẵn theo bộ ký tự từng ngôn ngữ, **chỉ tải font của ngôn ngữ đang chọn**. Tên người chơi có ký tự ngoài bộ (vd chữ Hán) → vẽ bằng Text thường (có cache).
- Góp ý admin: lưu kèm ngôn ngữ người gửi; admin có nút "Dịch" (tuỳ chọn, để sau).

**Thuật ngữ tu tiên**: dùng tên chuẩn mà cộng đồng truyện tu tiên quốc tế quen dùng, không dịch máy từng chữ.

| Việt | English | Ghi chú cho các ngôn ngữ khác |
|---|---|---|
| Luyện Khí | Qi Refining | Dịch nghĩa theo English |
| Trúc Cơ | Foundation Establishment | |
| Kim Đan | Golden Core | |
| Nguyên Anh | Nascent Soul | |
| Hóa Thần | Deity Transformation | |
| Luyện Hư | Void Refinement | |
| Hợp Thể | Body Integration | |
| Độ Kiếp | Tribulation Transcendence | |
| Đại Thừa | Mahayana | |
| Chân Tiên / Kim Tiên | True Immortal / Golden Immortal | |
| Thái Ất Ngọc Tiên / Đại La Kim Tiên | Taiyi Jade Immortal / Daluo Golden Immortal | |
| Đạo Tổ | Dao Ancestor | |
| Linh khí / Tiên khí | Spirit Qi / Immortal Qi | |
| Uy Danh | Renown | |

- Câu tấu hài (kill feed, câu thoại, danh hiệu): **viết lại cho buồn cười theo từng ngôn ngữ**, không dịch sát chữ. Mỗi ngôn ngữ cần 1 người bản ngữ đọc duyệt trước khi phát hành (agent soạn bản nháp + ghi chú ngữ cảnh từng câu vào `i18n/NOTES.md`).
- Tên nhân vật: giữ tên Việt có dấu làm tên riêng (bản sắc), kèm biệt danh dịch nghĩa ngắn ở thẻ chọn nhân vật (vd "A Hớn — *the Bun Guy*").

**SEO đa ngôn ngữ**
- Mỗi ngôn ngữ một đường dẫn trang chủ riêng (`/`, `/en/`, `/pt/`, `/es/`, `/id/`, `/th/`), có `<title>`, `<meta description>`, ảnh OG, nội dung giới thiệu ngắn **dạng HTML tĩnh** (công cụ tìm kiếm đọc được, không phụ thuộc JS), thẻ `hreflang` liên kết các bản, `sitemap.xml`.
- Trang chủ có đoạn mô tả ngắn + 3 ảnh chụp game + cách chơi, đặt **phía dưới** khung game (không làm chậm vào game).

## 20. Tên game & thương hiệu

- **Tên đã chốt: BladeAscend** (tên miền dự kiến `bladeascend.io`).
  - **Tên Tiếng Việt: "Đạo Hữu Xin Dừng Bước"** — tên game cho người chơi Việt (logo trang `/`, màn hình tải, tiêu đề tab, ảnh chia sẻ). Chế từ câu kinh điển của truyện tu tiên: một bên chạy, một bên đuổi theo gọi với — đúng y gameplay.
  - **Bỏ hẳn cụm "Kiếm Xoay Tu Tiên"** ở mọi nơi người chơi thấy (tên, tiêu đề, mô tả, màn hình tải, admin).
  - `<title>` Việt: `Đạo Hữu Xin Dừng Bước — Game tu tiên .io chơi ngay trên trình duyệt`.
  - **Mô tả ngắn** (`<meta description>`, ≤ 160 ký tự): *"Tu tiên bằng cả trăm thanh kiếm bay. Chém người cướp tu vi, phi thăng tới Đạo Tổ — nếu chạy kịp. Chơi miễn phí ngay trên trình duyệt."*
  - **Khẩu hiệu dưới logo:** *"Kiếm bay vòng vòng, đạo hữu chạy vòng vòng."*
  - **Đoạn giới thiệu trên trang chủ** (HTML tĩnh dưới khung game):
    > Ngươi là một tu sĩ mới nhập môn, trong tay chỉ có một thanh kiếm. Nhặt linh khí, gom thêm kiếm, xếp thành vòng vòng quanh người — càng nhiều vòng càng chém xa. Gặp đạo hữu yếu hơn thì đuổi theo cướp tu vi; gặp đạo hữu mạnh hơn thì… chạy, kèm một câu thật lịch sự: *"Đạo hữu xin dừng bước!"*.
    > Chín đại cảnh giới từ Luyện Khí tới Đại Thừa, năm tầng Tiên giới tới Đạo Tổ. Năm khu vực, cổng dịch chuyển ngẫu nhiên, mười nhân vật dở hơi với mười bí kíp riêng. Không cần tải, không cần đăng ký — nhập tên là vào.
  - Từ khoá tìm kiếm (đặt tự nhiên trong mô tả/giới thiệu, không nhồi): "game io", "tu tiên", "chơi trên trình duyệt", "không cần tải".
  - Ngôn ngữ khác: `BladeAscend.io — {khẩu hiệu theo ngôn ngữ}` (vd en: *Spin your swords, ascend to immortality*; agent soạn khẩu hiệu cho từng ngôn ngữ trong i18n, người bản ngữ duyệt).
  - Cách viết: luôn **BladeAscend** (chữ B và A hoa, viết liền); dạng thường `bladeascend` cho tên miền, tên repo, id kỹ thuật.
  - Lưu ở `config.brand` = `{ "name": "BladeAscend", "domain": "bladeascend.io", "name_vi": "Đạo Hữu Xin Dừng Bước", "tagline_vi": "Kiếm bay vòng vòng, đạo hữu chạy vòng vòng." }`; mọi chỗ hiển thị tên game (trang chủ, màn hình tải, tiêu đề tab, OG, manifest PWA, trang admin, ảnh chia sẻ) lấy từ đây.
  - **Tên project / mã nguồn**: repo, thư mục gốc, module Go, `package.json`, tiêu đề các tài liệu trong `docs/` đổi thành `bladeascend` / "BladeAscend". Không còn chỗ nào trong code hay tài liệu dùng "kiem-xoay" / "Kiếm Xoay Tu Tiên" (trừ lịch sử thay đổi).
  - Có một game PC tên gần giống (*Darkblade Ascent*, roguelike góc nhìn thứ nhất) — khác thể loại, không trùng tên, nhưng nên luôn viết liền "BladeAscend" để phân biệt.
- Tiêu chí: ngắn (≤ 12 ký tự trước `.io`), dễ đánh vần bằng tiếng Anh, gợi "kiếm xoay" + "tu tiên", **không đụng thương hiệu có sẵn** (vd tránh "Halo" — thương hiệu Microsoft; tránh "Tiên Kiếm" — tên một dòng game nổi tiếng), còn tên miền `.io` và tên trên các cổng game.
- Việc chủ dự án cần làm: mua tên miền `bladeascend.io` (và nên giữ thêm `bladeascend.com`), kiểm tra trùng tên trên các cổng game .io lớn và cửa hàng ứng dụng, tra nhãn hiệu (WIPO Global Brand Database).

## 17. Checklist test chấp nhận

- [ ] Shipper Luyện Khí 1 không buff: tốc chạy = 195 × 1.15; Cụ Rùa = 195 × 0.85; nội tại cộng chung với buff theo §4.0 (vd Shipper + Thần Hành = 195 × 1.55).
- [ ] Trang chọn nhân vật hiện 4 thanh chỉ số đúng bảng §6.4.

- [ ] Grep code: không còn chỗ nào nhân 2 hệ số buff/kỹ năng/địa hình với nhau cho cùng một chỉ số; mọi chỉ số đi qua 1 hàm `stat(base, bonuses, cap)` dùng chung server + client.

- [ ] Bà Tám dùng Loa: mọi địch trong 6 U ở mọi hướng bị đẩy ra 6 U + đổi chiều vòng kiếm.
- [ ] Ông Ba Võng trong 4s Trộm Kiếm: va kiếm với người cùng cảnh giới → Ông Ba không mất kiếm, đối thủ mất kiếm và Ông Ba +đúng số đó (tới trần, dư rơi Bó Kiếm).

- [ ] Người thật bị chém chết → rời phòng ngay, màn hình Tổng kết; chờ 2 phút không tự vào lại; tải lại trang vẫn ở Tổng kết; bấm "Chơi tiếp" mới vào trận (Luyện Khí 1).

- [ ] Admin đổi trần phòng 50 → 30: phòng mới chỉ nhận 30 người; phòng đang 35 người không đuổi ai, chỉ ngừng nhận. Giá trị còn nguyên sau deploy.
- [ ] Production: người thật ≥ 40 → ngừng sinh bot, < 35 → sinh lại. Staging: ≥ 20 → ngừng. Mỗi khu được lấp đủ 10 nhân vật khi bot được phép.

- [ ] Người bị Ru Ngủ / choáng vỏ chuối: không di chuyển nhưng vòng kiếm vẫn quay, va kiếm hai chiều bình thường, vẫn chém chết được người lao vào thân mình.
- [ ] Người bị Thằng Mập bịt mắt: trong 5s kiếm chạm thân người khác không gây chết; va kiếm vẫn bể kiếm bình thường.

- [ ] Chém trúng thân người Kim Đan (còn 1 vòng) → chết ngay, vào lại Luyện Khí 1; người chém nhận 30% Linh khí nạn nhân nhưng không lên quá 2 đại cảnh giới; xác rơi 30% Linh khí + 50% kiếm.
- [ ] Người có Kim Cang bị chém → không chết, khiên vỡ.
- [ ] Đạo Tổ bị chém → chết, rơi 30% Tiên khí; người chém Tiên giới nhận 30% nhưng không lên quá 1 cảnh giới Tiên.

- [ ] Thần Hành + Thiêu Đốt + theo gió → tốc ×2.0 (không phải ×3.78); Shipper Gom Đơn + mọi buff → ≤ ×3.0; Thiêu Đốt trong bùn → ×1.4.
- [ ] Không còn vật phẩm Bí Kíp trên bản đồ. Kỹ năng tự hồi đúng bảng, tích đủ số lượt, Tiên giới +1 lượt; vào trận có 1 lượt.

- [ ] Bot mới sinh luôn Luyện Khí 1, 1 kiếm, không buff.
- [ ] loadtest 50 người thật: tick p95 mỗi khu < 4 ms trước khi mở giới hạn 50.

- [ ] 2 người ở 2 khu khác nhau: chỉ 1 người có danh hiệu Top 1 trong cả phòng; bảng phòng hiện đúng khu + cảnh giới từng người.
- [ ] Top 1 thẻ to viền vàng động + vương miện + chân dung; Top 2/3 thẻ nửa rộng bạc/đồng; 4–10 dạng dòng; dòng của mình ghim cuối khi ngoài top 10.
- [ ] Bảng không tràn ở 1366×768 và iPhone dọc; điện thoại mặc định thu gọn, chạm để mở.

- [ ] Cô Ba Kim Đan bày sạp (bậc 3): người Kim Đan và thấp hơn không chém được, phải đi vòng; người Nguyên Anh chém được; Cô Ba đi xuyên sạp của mình; dẫm vỏ chuối → trượt + choáng 1s; sau 6s sạp biến mất; Shipper đang Gom Đơn đi xuyên qua.

- [ ] Hút Mì: địch 100 kiếm → Sư Phụ cướp 25 (không vượt 20% tối đa của mình, dư rơi Bó Kiếm); không ai trong nón → hụt, mất lượt.
- [ ] Bán kính nhặt Shipper Đại Thừa 10 + Gom Đơn + Nam Châm ≈ 255 (cộng §4.0, không phải 332); đang bị trói/ngủ/choáng dùng được và thoát ngay.

- [ ] Luyện Khí chém rào tre → biến mất ngay, không hoạt ảnh; chém tường gạch → không có tác dụng + chữ "Cần Kim Đan"; 30s sau rào mọc lại đúng chỗ.
- [ ] `mapgen --check` pass cả 5 khu × 14 cảnh giới: không túi kín, mọi cổng/Đài/Thiên Môn tới được, trống ≥ 60%.
- [ ] Không còn vật cản nào trong khu có bậc ∞ (chỉ biên khu). Mỗi khu có ít nhất 1 vật cản bậc ≥ 9. Kim Tiên chém cột ngọc → không tác dụng; Đạo Tổ chém thạch bi → biến mất.

- [ ] Loa Phường của Bà Tám Luyện Khí 1 đẩy người Đạo Tổ đúng 540; Bà Tám Đạo Tổ đẩy người Luyện Khí đúng 3 240 (theo U người dùng, không theo người bị đẩy); đẩy không xuyên tường.
- [ ] Không còn số phạm vi cố định nào trong code kỹ năng — mọi phạm vi đọc từ `skills.ranges_u` × U.

- [ ] `assets:size` trước/sau ghi PERF_LOG; gói `boot` ≤ 8 MB, toàn bộ ≤ 80 MB; lần vào thứ 2 không tải lại art (kiểm tra tab Network: from cache).
- [ ] Màn hình tải hiện ngay khi mở trang, thanh tiến độ theo byte chạy đều, có % và MB; mất mạng giữa chừng → "Thử lại" tải tiếp.
- [ ] Qua cổng sang khu chưa tải: chờ tối đa 2s rồi vào, không màn hình đen.

- [ ] Không có tên bot nào chứa "bot"; 2 bot trong cùng room không trùng tên; bot chết thì 70% quay lại cùng tên sau 4–12s.
- [ ] Đổi run → chase → flee: thân chạy liên tục không reset frame, chỉ đổi mặt + tư thế; quay trái/phải không lật qua lại khi chạy dọc; tốc độ bước chân khớp tốc độ di chuyển.
- [ ] `assets:report` có điểm độ giật mọi animation; không frame nào vượt ngưỡng trong build thật.

- [ ] Thế giới 5 khu, mỗi khu 12 000²; không đi bộ sang khu khác được; 200 lượt qua cổng: ra đều các cổng còn lại, không bao giờ ra lại cổng vừa vào, ~80% sang khu khác.
- [ ] Người mới vào spawn ở khu ít người nhất; tổng người + bot mỗi phòng vẫn 20, bot sinh vào khu ít nhân vật nhất.
- [ ] Băng Nguyệt Hồ: trên băng trượt có quán tính.
- [ ] 4 khu (trừ Thiên Đình) có Thiên Môn ở tâm, đường tới thông; bước vào luôn ra ở đúng Bến Thiên Môn của khu đó ở Thiên Đình; minimap có icon + mũi tên chỉ hướng.
- [ ] Đi sát/đi dọc tường, rào, đá tảng: không giật, không rung; F3 'hiệu chỉnh/giây' ≈ 0 khi chỉ đi sát vật cản.
- [ ] Trúng Phun Trân Châu: màn hình bị che ~60% trong 5s, không nhặt được gì trong 5s; HUD và minimap không bị che.
- [ ] Shipper dùng Giao Hỏa Tốc đi xuyên vách đá được nhưng không ra khỏi biên khu; hết giờ trong vật cản thì được đẩy ra chỗ trống.

- [ ] Hai Kim Đan 60 kiếm mài nhau 1 giây → mỗi bên còn ~30 kiếm; đi xuyên qua nhau được; chỉ người còn 1 vòng mới bị chém thân.
- [ ] Tốc Kiếm chỉ làm kiếm quay nhanh hơn, tầm chém không đổi.
- [ ] Luyện Khí có Cương Kiếm va Trúc Cơ không buff → Trúc Cơ mất 1 mỗi lần va, Luyện Khí nứt 1/1.25; phá rào tre nhanh gấp đôi.
- [ ] `cmd/loadtest` 40 nhân vật, 10 người Đạo Tổ đủ kiếm: tick < 4 ms, snapshot < 8 KB, FPS ≥ 60 trên PC tầm trung, ping tăng < 20 ms.
- [ ] Không còn vật phẩm Trường Kiếm trên bản đồ, HUD, config, debug (phím 1–7 bỏ buff này).
- [ ] Giữ Shift: chạy ×1.8, mất 8% EXP tầng/giây, không bao giờ rớt cảnh giới.
- [ ] Đài nhỏ phun mỗi 5s, đợt lớn mỗi 30s; Thiên Tiên Đài báo trước 10s và phun viên 150 điểm mỗi 150s.
- [ ] mapgen: 10 seed liên tiếp đều qua kiểm tra liên thông, có Thiên Tiên Đài, 4 sân môn phái, đường chính không vật cản.
- [ ] `?debug=true` mới dùng được phím debug; `B` thả bot.

- [ ] Luyện Khí tầng 1 nhỏ hơn rõ ràng so với Luyện Khí tầng 10; Đại Thừa tầng 10 scale = 5.
- [ ] Kiếm dài ra và vòng rộng ra theo scale.
- [ ] Cùng 1 tầng: 1 → 10 kiếm, bán kính vòng không đổi; kiếm thứ 11 mới mở vòng 2 rộng hơn.
- [ ] Luyện Khí tối đa 10 kiếm = 1 vòng; Trúc Cơ 30 = 2 vòng (10 + 20); Kim Đan 60 = 3 vòng; các vòng so le, mỗi vòng một màu, đều nhìn thấy; lưỡi nghiêng về phía sau và lật khi đổi chiều.
- [ ] Mũi kiếm vòng 3 xa tâm hơn vòng 1; hitbox server khớp với hình vẽ (bật chế độ debug vẽ capsule để so).
- [ ] Kiếm quay đúng 1 vòng / 1 giây (0.625 giây khi có Tốc Kiếm), nhìn rõ từng thanh, không bị hiệu ứng quay ngược.
- [ ] Nhân vật đi sát mép trên tảng đá được (chỉ bị chặn ở chân đế), đứng sau đá thì bị che.
- [ ] Farm không đánh nhau: Luyện Khí → Trúc Cơ trong 30–60s (đo bằng cmd/sim).
- [ ] 2 người cùng cảnh giới va kiếm → mỗi bên mất 1 kiếm, cả hai đổi chiều quay.
- [ ] Trúc Cơ va Luyện Khí: sau 12 lần va, Luyện Khí mất 12 kiếm, Trúc Cơ mất 10; **không ai đổi chiều**.
- [ ] Nguyên Anh va Luyện Khí: Luyện Khí mất 1 mỗi lần va, Nguyên Anh mất 1 sau mỗi ≈1.73 lần (crack cộng 1/1.728).
- [ ] Người 0 kiếm vẫn sống; bị chém thân thì rớt cảnh giới (Luyện Khí thì chết).
- [ ] Viên EXP từ quái chỉ người chém nhặt được trong 1.5s đầu.
- [ ] Đi chéo 90° so với gió → tốc độ không đổi.
- [ ] Vật cản bị chém vỡ → đi qua được → 30s sau mọc lại, không kẹt người.
- [ ] Kiếm chém vật cản không bị mất.
- [ ] Luyện Khí đủ kiếm phá rào tre ≈ 15s; Đại Thừa phá đá tảng ≈ 9s.
- [ ] Vào cổng → ra ở cổng khác, cổng đích lóe sáng trước, không có bất tử khi ra.
- [ ] Người scale lớn không chui lọt hẻm 150.
- [ ] Buff hết hạn đúng giờ, nhặt lại thì làm mới thời gian.
- [ ] Ẩn Thân: người khác thấy mờ 20%, không tên; tối đa 5s; chém trúng người là hiện ngay.
- [ ] Trúc Cơ chém trúng Nguyên Anh → +850 Uy Danh; chém lại cùng người trong 60s → +0.
- [ ] Bảng Thần Tốc chỉ có những mạng đã đạt Đại Thừa.
- [ ] Đại Thừa tầng 10 chưa đầy thanh: nhặt Linh khí bình thường, không nhặt được Tiên khí. Đầy 694 960: nhặt 1 Tiên khí → Chân Tiên; 1 001 → Kim Tiên; 10 001 → Đạo Tổ. Người dưới Đại Thừa 10 đi qua viên Tiên khí không nhặt được (viên mờ 30%); người Tiên giới đi qua viên Linh khí không nhặt được.
- [ ] Đột phá Kim Tiên: rồng vàng chui lên từ chân, quấn quanh người 1.5 vòng (nửa sau khuất sau nhân vật), bay vút lên, 2.5s; sau đó rồng nhỏ (60%) luôn lượn lờ quanh người, mỗi 8–15s làm 1 động tác ngẫu nhiên, khi chạy thì bị kéo dài phía sau.
- [ ] Chém người Tiên giới (người chém là phàm, Tiên hoặc bot đều được): luôn có viên Tiên khí vàng văng ra quanh xác, tổng = 30% Tiên khí nạn nhân mất (Chân Tiên 800 Tiên khí bị chém → rơi 240).
- [ ] Đạo Tổ: nhặt Tiên khí khi chưa đủ 15 000 → nạp thanh, không cộng Uy Danh; đủ 15 000 → +1 Uy Danh mỗi 10 Tiên khí; Thiêu Đốt kéo thanh xuống (40/giây), phải nạp lại 15 000 mới đổi tiếp.
- [ ] Không còn vùng nền/bùn/bụi/đường nào là hình chữ nhật thẳng cạnh hay hình tròn hoàn hảo; ranh giới biome loang tự nhiên; đứng sát mép bùn thì hiệu ứng chậm khớp đúng mép đa giác.
- [ ] HUD góc trên trái là chân dung nhân vật, viền màu cảnh giới, biểu cảm đổi theo trạng thái.
- [ ] Bảng tên đọc rõ trên nền vòng kiếm nhiều màu; tên mình màu vàng có ▼; người cao cảnh giới hơn mình có ▲ đỏ; Top 1 có ruy băng đỏ-vàng + vương miện + vệt sáng lướt.
- [ ] Top 1/2/3 phòng hiện đúng danh hiệu trên đầu và trong bảng; đổi hạng phải giữ 3s mới đổi danh hiệu.
- [ ] Minimap thấy mình, khung camera, cổng, Top 1, Rương Báu; không thấy người khác; Top 1 vào bụi thì biến khỏi minimap.
- [ ] Phòng 5 người → 15 bot; thêm người không xóa bot; người thứ 21 sang phòng mới.
- [ ] Leaderboard ghi đúng 3 bảng sau khi chết, bot không có trong bảng toàn cục.
