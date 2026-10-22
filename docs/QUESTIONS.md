# Câu hỏi đang chờ anh trả lời

Chỉ còn **2 câu**. Câu đã xong chuyển sang `docs/QUESTIONS_DONE.md`.

Mỗi câu: bối cảnh → phương án → **tạm dùng** (đã code theo phương án này, đổi lại dễ).

---

## Q13. Nhịp độ quá chậm sau 3 thay đổi hôm nay

**Bối cảnh:** anh yêu cầu 3 thứ đều làm chậm tiến độ, em đã làm đủ cả ba:
tốc độ chạy còn một nửa · mỗi cảnh giới tốn gấp đôi cảnh giới trước (`realm_mult` 2.0) ·
mật độ EXP và vật phẩm giảm 50% (`cell` 226, `count_per_room` 200).

**Đo được** (`cd server && go run ./cmd/sim --farm --bots 20 --seed 7`):

| Mốc | Trước hôm nay | Bây giờ | Spec gốc §6.1 ghi |
|---|---|---|---|
| Luyện Khí → Trúc Cơ | 40–46s | **165s** | 30–60s ❌ |
| → Kim Đan | ~60s | **321s** | ≈1 phút ❌ |
| → Nguyên Anh | ~72s | **417s** | ≈1.2 phút ❌ |
| Tổng EXP tới Đại Thừa 10 | 28 288 | **694 960** | – |

**Hai điểm cần anh quyết:**

1. **§6.1 đang mâu thuẫn với config.** Spec vẫn ghi mục tiêu "Luyện Khí → Trúc Cơ 30–60s"
   nhưng thực tế 165s. Nếu 165s đúng ý anh thì sửa §6.1 thành "≈2.5–3 phút"; nếu không thì
   em giảm `exp_orbs.cell` lại.
2. **Đại Thừa gần như không tới được.** Tốc độ farm ≈25 EXP/s → 694 960 EXP ≈ **7.7 giờ**
   farm liên tục. Hệ quả: bảng **Thần Tốc** (thời gian lên Đại Thừa) sẽ mãi trống, và
   Tiên giới (§3.4) không ai chạm tới được vì phải qua Đại Thừa tầng 10 trước.
   Muốn giữ EXP gấp đôi mà vẫn tới Đại Thừa trong một ván (~10–20 phút) thì cần tăng nguồn
   EXP ở cảnh giới cao: viên EXP to dần theo cảnh giới, hoặc EXP từ quái / chém người nhân
   theo cảnh giới.

**Tạm dùng:** giữ nguyên như anh nhắn, chưa đổi gì thêm.

---

## Q14. `docs/_raw/GAME_SPEC.md` thiếu 13 thứ anh đã nhờ làm hôm nay

**Bối cảnh:** `docs/_raw/` là spec gốc, chỉ anh sửa — em không ghi vào đó. Bản anh vừa dán có
hệ **Tiên giới** hoàn toàn mới (em đã code xong), nhưng phần còn lại sinh lại từ bản v3.6 nên
mất những thứ anh nhắn trực tiếp sau đó.

Em **giữ code theo lời anh nhắn**, không lùi về `_raw`. Danh sách để anh dán lại vào `_raw`:

### 6 con số `_raw` đang ghi giá trị cũ

| Mục | `_raw` ghi | Code đang chạy | Anh đã nhắn |
|---|---|---|---|
| `player.move_base` | 260 × sqrt(scale) | **130**, bỏ sqrt | "tốc độ chạy giảm 1/2, chỉ item mới tăng tốc" |
| `player.invuln_on_spawn` | 2.0 | **5.0** | "5s đầu tiên vào game là bất tử" |
| `sword.clash_knockback` | 150 | **0** | "va chạm không làm dội lại, mài tới khi hết kiếm" |
| `cultivation.realm_mult` | 1.2 | **2.0** | "cảnh giới mới cần gấp đôi exp cảnh giới cũ" |
| `exp_orbs.cell` | 140 | **226** | "giảm mật độ exp và item 50%" |
| `items.count_per_room` | 400 | **200** | như trên |

### 7 mục `_raw` chưa có

1. **§4.1c Thiêu Đốt Tinh Huyết** — giữ Shift / nút 🔥 chạy nhanh ×1.8, đốt 8% EXP tầng hiện tại mỗi giây, không rớt cảnh giới.
2. **§11.2 bảng phòng xếp theo Uy Danh** — `_raw` vẫn ghi "xếp theo cảnh giới + tầng".
3. **§11.3 danh hiệu Top 1–10** — `_raw` chỉ có Top 3.
4. **§11.4 tin loan toàn server** — hạ người trong Top 5 thì mọi phòng đều thấy.
5. **§9.3 `bots.spawn_rank` = 11** — bot sinh ngang hạng 11, không còn trần cảnh giới.
6. **§18 `?debug=true`** (+ phím `B` thả bot) và **§18.1** bảng điều khiển mờ lúc vào trận.
7. **§12.2 máu quái** — thanh máu nhỏ trên đầu thay cho số to đè lên mặt.

**Riêng `realm_mult` cần anh xác nhận:** `_raw` ghi 1.2 (tổng 28 288 EXP), code đang 2.0
(694 960 EXP). Lệch nhau 25 lần nên em không đoán. Muốn quay về 1.2 thì em đổi 1 dòng.
