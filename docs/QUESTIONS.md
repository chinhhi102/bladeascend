# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026; cập nhật 02/11 sau khi anh trả lời Q37–Q42. Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.

## Q37 — Ghi số config đã chỉnh theo sim vào spec §6.0 · CHỜ ANH SỬA SPEC
Anh chốt **giữ config**. Bảng để anh chép vào spec (đo `cmd/sim --farm`, 8 khu × 12 bot, seed 1, bản v6.33):

| Khoá | Giá trị đang chạy | Ghi chú |
|---|---|---|
| `gates.transit_secs` | 2.0 | qua cổng (spec cũ 0.8) |
| `minimap.size_mobile` | 96 | px (spec cũ 120) |
| `buffs.magnet.mult` | 1.2 | theo §4.0 = **+0.2** bán kính nhặt (spec cũ 1.5) |
| `exp_orbs.cell` | 160 | (spec cũ 140) |
| `room.monster_count` | 150 | **mỗi khu** (750/phòng 5 khu) |
| `exp_orbs.values` / `weights` | [5, 20, 60] / [80, 18, 2] | |
| `exp_economy.orb_value_mult` | 1.5 | |
| `exp_economy.pickup_scale_exp` | 0.5 | bán kính nhặt theo √scale |
| `exp_economy.bot_steal_share` | 0.15 | |
| `cultivation.exp_base` / `realm_mult` / `tier_step` | 100 / 2.0 / 0.08 | |

| Cảnh giới | Farm đo (phút) | Bảng §6.0 | Cộng dồn | Có Nam Châm (phút) |
|---|---|---|---|---|
| Luyện Khí | 1.19 | 1.3 | 1.2 | – |
| Trúc Cơ | 1.31 | 1.5 | 2.5 | – |
| Kim Đan | 1.55 | 1.8 | 4.1 | – |
| Nguyên Anh | 1.94 | 2.1 | 6.0 | 1.65 |
| Hóa Thần | 2.46 | 2.6 | 8.5 | 2.21 |
| Luyện Hư | 3.13 | 3.3 | 11.6 | 2.84 |
| Hợp Thể | 3.89 | 4.1 | 15.5 | 3.53 |
| Độ Kiếp | 4.59 | 5.2 | 20.1 | 4.46 |
| Đại Thừa | 5.64 | 6.6 | **25.7** (bảng 29.1) | 5.93 |

Luyện Hư 1 → Chân Tiên khi có Nam Châm suốt: 23.1 phút (cần ≥ 15). Xong khi spec có bảng này.

## Q38 — Những dòng spec còn luật "chết rớt cảnh giới" · CHỜ ANH SỬA SPEC
Số dòng theo `docs/_raw/GAME_SPEC.md` bản v6.34 (code đã theo v6.26: chém trúng thân = chết):

| Dòng | Mục | Nội dung còn sót |
|---|---|---|
| 145 | §1 vòng chơi | "bị chém thì rớt tầng, mất kiếm" |
| 299 | §3.2 hiệu ứng | "Bị chém rớt cảnh giới: thần thú gãy vụn" |
| 461 | §5.2 | "Bó Kiếm khi rớt cảnh giới gần như không xảy ra" |
| 483 | §5.2 | "Trúc Cơ trở lên rớt 1 cảnh giới, Luyện Khí chết" |
| 557 | bảng vật phẩm | Bó Kiếm "rơi ra khi rớt cảnh giới/chết" |
| 580 | §6.3 | "Qua cổng, rớt cảnh giới không mất" |
| 674 | kill feed | mẫu `"{B} bị {A} chém rớt xuống {realm}"` |
| 1098 | §11.4 tin loan | "bị chém trúng (rớt cảnh giới hoặc chết)" |
| 1197 | §12 chân dung | "khóc khi rớt cảnh giới" |
| 1444 | §16 M2 | "chém thân, rớt tầng" |
| 1461 | §16 config mẫu | `"tien_rob_share": 0.5` (config thật không còn khoá này) |
| 1486 | §16 config mẫu | `"hit_drops_realm": true` (config thật không còn khoá này) |
| 1614 | §20 giới thiệu | "mười bí kíp riêng" (kỹ năng giờ tự hồi) |
| 1705 | §17 checklist | "bị chém thân thì rớt cảnh giới (Luyện Khí thì chết)" |

Dòng 434 và 1689 ("không bao giờ rớt cảnh giới" khi chạy nhanh) vẫn đúng, không cần sửa.
