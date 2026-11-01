# Rà hiển thị 10 kỹ năng theo spec v6.36 §6.5

Ngày rà: 02/11/2026, trước khi làm (CHANGES_v6_36 mục 6). Đọc từ mã hiện có:
`server/internal/game/skills.go`, `status.go`, `skillstate.go`, `stall.go` ·
`client/src/render/scene.ts` (`onSkillEffect`), `skillfx.ts`, `player.ts`
(`updateStatusIcons`, `drawAura`), `swordring.ts` (`drawPearl`), `stallview.ts`.

Ký hiệu: ✅ có, đúng spec · 🟡 có một phần · ❌ thiếu.

## Phần chung cho cả 10 kỹ năng

| Hạng mục | Hiện trạng |
|---|---|
| ① Anim `skill` | 🟡 cả 10 atlas đã có anim `skill`, nhưng **không bao giờ được phát**: server không có tâm trạng `skill` (`MoodList`), client không tự bật. |
| ① Bong bóng thoại | ✅ `SkillCast` → `showBubble(STR.skillLines[char])`. |
| ① SFX | ❌ `sound.ts` có tên 5 tiếng (`skill_hat_ru`, `skill_sleep`, `skill_hut_mi`, `skill_gom_don`, `skill_bay_hang`) nhưng **không có tệp nào** trong `assets/src/audio/sfx/` → im lặng. 5 kỹ năng còn lại chưa có cả tên. |
| ② Vùng | 🟡 `skillFx.circle/cone/beam` đúng công thức U của server (`skillRange`), nhưng nét liền, mờ 0.45s (spec: nét đứt sáng, 0.3s), màu chưa theo bảng (Ru Ngủ đang xanh lam, spec tím nhạt). |
| Event `StatusStart/StatusEnd` | ❌ chưa có. Client chỉ biết trạng thái qua bit `st` trong snapshot (không có số giây, không biết ai gây ra). Riêng mù có `Blinded` (chỉ gửi nạn nhân). |
| ④ Vòng đếm ngược | ❌ không kỹ năng nào có. |
| Icon trạng thái | 🟡 chỉ có 🚫 cho mù, lại là **emoji hệ thống** (spec cấm). 7 icon còn lại chưa có. |
| HUD người bị trúng (icon + giây dưới chân dung) | ❌ |
| Rung màn hình lúc trúng kỹ năng | ❌ (chỉ có rung khi bị chém thân `Hit`). |
| Mũi tên mép màn hình khi trúng từ ngoài khung nhìn | ❌ |
| Theo chất lượng (Thấp: giữ ②③④ đơn giản) | ❌ chưa phân nhánh. |

## Từng kỹ năng

| Nhân vật — kỹ năng | ② Vùng | ③ Trên mục tiêu / người dùng | ④ Đếm ngược | ⑤ Kết thúc |
|---|---|---|---|---|
| A Hớn — Bánh Bao Thần | ✅ không cần (tự thân) | ❌ không vẽ gì (client không bắt `banh_bao`): thiếu hơi nóng, kiếm bay từ bánh vào vòng, số "+N kiếm" | — (tức thì) | ❌ ợ + khói |
| Bà Tám — Loa Phường | 🟡 vòng tròn 6 U tĩnh, chưa "lan từ tâm ra" | ❌ vệt bụi theo hướng đẩy, icon 📢, loé vòng kiếm khi đổi chiều | ❌ | 🟡 vòng tự mờ (chưa phải sóng) |
| Ông Ba — Ru Ngủ / Trộm Kiếm | 🟡 vòng 4 U lúc hát + lúc ngủ, màu xanh lam (spec tím nhạt), chưa có nốt nhạc | 🟡 người ngủ: chỉ `stun_stars` 1 lần, thiếu "zzz", bong bóng mũi, icon 💤; Ông Ba: chỉ vòng vàng 1 lần, thiếu võng + viền tím 4s; trộm kiếm: chỉ chữ nổi, thiếu vệt tím bay về | ❌ | ❌ võng biến mất + "ngáp" |
| Đại Ca — Gọi Đàn Em | ✅ không cần | 🟡 đàn em hiện ra như người chơi thường (nhỏ 30%), thiếu lối vào, thanh đếm 8s, vòng kiếm tím sẫm | ❌ | ❌ server đã gửi `dan_em_gone` nhưng client bỏ qua (không khói, không thoại) |
| Thằng Mập — Phun Trân Châu | 🟡 nón 45° × 8 U, chưa có hạt bay | 🟡 trân châu dính vòng kiếm ✅ (`drawPearl`), dính kính nạn nhân ✅; icon là emoji 🚫 ❌ | ❌ | ❌ trân châu rơi |
| Cô Ba — Bày Hàng | 🟡 vẽ vòng tròn cách 3 U (không phải 5 khung sạp nhấp nháy) | 🟡 sạp có art/placeholder; thiếu bạt cam, thanh đếm 6s, vỏ chuối không vẽ; người trượt: thiếu vệt trượt + sao choáng | ❌ | ❌ sạp sụp + vỏ chuối mờ |
| Sư Phụ Mì — Hút Mì | 🟡 nón 60° × 5 U **chỉ khi hụt**; trúng thì không loé nón | 🟡 sợi mì ✅ (tia thẳng 0.45s), chữ "−N" trên nạn nhân ✅; thiếu kiếm bay dọc sợi mì, "+N" trên Sư Phụ, mì rơi khi hụt | — (tức thì) | ❌ sợi mì rút về |
| Bé Tí — Gồng Cơ | ✅ không cần | ❌ không vẽ gì (client không bắt `gong_co`, bit `brace` không dùng để vẽ) | ❌ | ❌ xì hơi xẹp |
| Shipper — Gom Đơn | ✅ không cần | ❌ chỉ có tiếng (mà tệp tiếng cũng thiếu); thiếu vệt tia chớp, vật cản mờ khi đi qua, vòng nhặt, icon ⚡ | ❌ | ❌ phanh + khói |
| Cụ Rùa — Rụt Cổ | ✅ không cần | ❌ mai sáng xanh + khiên không có (client không bắt `rut_co`); sóng đẩy 2 U lúc ra mai ✅ (vòng tĩnh) | ❌ | 🟡 sóng có, thiếu "à há" |

## Kế hoạch làm (mục 7–8)

1. **Server**: event `StatusStart {A=nạn nhân, B=người gây (0 = tự thân), N=mã trạng thái,
   F=giây, X/Y=vị trí người gây}` và `StatusEnd {A, N}`; thêm vào CUỐI bảng event hai bên.
   Mã trạng thái = thứ tự `StatusNames` + 3 mã chỉ để hiển thị: `push` (bị đẩy), `steal`
   (Trộm Kiếm), `minion` (đàn em còn bao lâu). Người xem thấy event khi thấy nạn nhân;
   nạn nhân và người gây luôn nhận.
2. **Client**: bảng hẹn giờ trạng thái theo event → icon vector 64 px trên đầu (có vòng
   đếm), HUD dưới chân dung (icon + giây), rung nhẹ, mũi tên mép màn hình, hiệu ứng tan.
3. Lớp ③ riêng từng kỹ năng theo bảng §6.5 (vector + particle trong `vfxLayer`), chất lượng
   Thấp bỏ particle trang trí.
4. ① phát anim `skill` 0.6s phía client khi nhận `SkillCast`; sinh 10 tiếng SFX placeholder
   bằng `tools/make_sfx.py`.

**Không đổi luật** (giữ đúng server): sạp vẫn có va chạm ngay lúc bật — khung nhấp nháy
0.3s chỉ là lớp vẽ chồng lên; đàn em vẫn sinh ở chỗ server đặt — "chạy từ mép vào" vẽ
bằng một vệt lướt ngắn tới chỗ đó.
