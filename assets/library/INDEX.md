# Thư viện art — `assets/library/`

Luật 13 CLAUDE.md · GAME_SPEC §13.7-3b: art do artgen sinh là **tài nguyên trả tiền, không bao giờ xoá**.
Art không dùng nữa nằm ở đây, giữ nguyên tên.

- Cấu trúc: `assets/library/<nhóm>/<đường dẫn y như trong assets/src>`.
- **Không** build vào `assets/dist`, không vào manifest, không tải cho người chơi, không tính ngân sách
  (pipeline và mọi `check:*` chỉ đọc `assets/src` / `assets/dist`).
- `artgen` coi art ở đây là ĐÃ CÓ → không bao giờ mua lại (`server/cmd/artgen/library.go`).
- Dùng lại: `git mv assets/library/<nhóm>/<đường dẫn> assets/src/<đường dẫn>`, thêm dòng vào
  `assets/src/manifest.json`, rồi build như bình thường. Xoá dòng tương ứng ở đây.
- Chuyển art mới vào: `git mv`, thêm một dòng vào bảng của nhóm (tạo nhóm mới nếu cần).

**Tổng: 230 tệp** (cập nhật 2026-11-01).

| Nhóm | Số tệp | Lý do ngừng dùng | Spec |
|---|---|---|---|
| `obstacle_states_v6.7` | 78 | Bộ trạng thái nứt/sắp vỡ/đổ nát + hoạt ảnh `break` của vật cản. Spec v6.7 bỏ: vật cản biến mất ngay, không hoạt ảnh vỡ, chỉ còn trạng thái `intact`. | v6.7 §8.2 |
| `obstacle_regrow` | 54 | Hoạt ảnh mọc lại `regrow/`. Client không nạp (QUESTIONS Q31 nhóm 1). | §8.2 · Q31 |
| `obstacle_states_04` | 3 | Khung trạng thái 04 cũ; v6.7 rút còn một trạng thái (`OBSTACLE_STATES = ['intact']`). | v6.7 §8.2 · Q31 |
| `vfx_boba` | 7 | Viên trân châu: `boba.ts` vẽ bằng canvas, không dùng ảnh (Q31 nhóm 1). | Q31 |
| `buffs_removed` | 1 | Icon buff Trường Kiếm — buff đã bỏ từ spec v4.3. | v4.3 · Q31 |
| `extra_frames` | 6 | Khung dư ngoài số khung danh mục khai (burst/open 6 khung khai tới 05, run_dust khai 4 khung). | ART_PROMPTS · Q31 nhóm 3 |
| `char_frames_extra_2026-10-22` | 4 | Khung dư sau khi sinh lại sheet idle/run (sheet mới ít khung hơn). | ART_PROMPTS §12 |
| `char_frames_extra_2026-10-29` | 35 | Khung dư sau khi sinh lại idle/hit bị cắt hỏng (sheet mới ít khung hơn). | ART_PROMPTS §12 |
| `char_frames_extra_2026-10-30` | 17 | Khung dư bị loại khi chuẩn hoá frame tự động (§12.5 B). | v6.3 §12.5 |
| `q31_chua_khu_nao_dung` | 13 | Art có sẵn nhưng chưa khu nào dùng (Q31 nhóm 2). Trước nằm trong `assets/src` + NEVER_SHIP; nay chuyển hẳn vào thư viện. | §13.7-3b · Q31 |
| `q31_dan_em` | 10 | Atlas riêng của Đàn Em — không nằm trong 10 nhân vật; đàn em vẽ bằng atlas Đại Ca thu nhỏ (spec v5.4 mục 19). | v5.4 · Q31 |
| `bi_kip_v6.25` | 2 | Vật phẩm Bí Kíp + icon cuộn cạnh nút kỹ năng. Spec v6.25 bỏ Bí Kíp: kỹ năng tự hồi chiêu + tích lượt. | v6.25 §6.3 |

## obstacle_states_v6.7

Bộ trạng thái nứt/sắp vỡ/đổ nát + hoạt ảnh `break` của vật cản. Spec v6.7 bỏ: vật cản biến mất ngay, không hoạt ảnh vỡ, chỉ còn trạng thái `intact`.

| Tệp | Ngày | Job / prompt gốc | Ghi chú |
|---|---|---|---|
| `obstacles/bamboo_wall/break/break_00.png` | 2026-10-31 | `obs_bamboo_wall_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/break/break_01.png` | 2026-10-31 | `obs_bamboo_wall_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/break/break_02.png` | 2026-10-31 | `obs_bamboo_wall_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/break/break_03.png` | 2026-10-31 | `obs_bamboo_wall_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/break/break_04.png` | 2026-10-31 | `obs_bamboo_wall_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/break/break_05.png` | 2026-10-31 | `obs_bamboo_wall_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/cracked.png` | 2026-10-31 | `obs_bamboo_wall_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/cracked_vertical.png` | 2026-10-31 | `obs_bamboo_wall_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/nearly_broken.png` | 2026-10-31 | `obs_bamboo_wall_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/nearly_broken_vertical.png` | 2026-10-31 | `obs_bamboo_wall_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/rubble.png` | 2026-10-31 | `obs_bamboo_wall_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/rubble_vertical.png` | 2026-10-31 | `obs_bamboo_wall_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/boulder/break/break_00.png` | 2026-10-31 | `obs_boulder_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/boulder/break/break_01.png` | 2026-10-31 | `obs_boulder_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/boulder/break/break_02.png` | 2026-10-31 | `obs_boulder_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/boulder/break/break_03.png` | 2026-10-31 | `obs_boulder_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/boulder/break/break_04.png` | 2026-10-31 | `obs_boulder_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/boulder/break/break_05.png` | 2026-10-31 | `obs_boulder_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/boulder/cracked.png` | 2026-10-31 | `obs_boulder_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/boulder/cracked_vertical.png` | 2026-10-31 | `obs_boulder_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/boulder/nearly_broken.png` | 2026-10-31 | `obs_boulder_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/boulder/nearly_broken_vertical.png` | 2026-10-31 | `obs_boulder_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/boulder/rubble.png` | 2026-10-31 | `obs_boulder_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/boulder/rubble_vertical.png` | 2026-10-31 | `obs_boulder_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/brick/break/break_00.png` | 2026-10-31 | `obs_brick_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/brick/break/break_01.png` | 2026-10-31 | `obs_brick_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/brick/break/break_02.png` | 2026-10-31 | `obs_brick_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/brick/break/break_03.png` | 2026-10-31 | `obs_brick_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/brick/break/break_04.png` | 2026-10-31 | `obs_brick_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/brick/cracked.png` | 2026-10-31 | `obs_brick_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/brick/cracked_vertical.png` | 2026-10-31 | `obs_brick_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/brick/nearly_broken.png` | 2026-10-31 | `obs_brick_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/brick/nearly_broken_vertical.png` | 2026-10-31 | `obs_brick_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/brick/rubble.png` | 2026-10-31 | `obs_brick_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/brick/rubble_vertical.png` | 2026-10-31 | `obs_brick_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/cart/break/break_00.png` | 2026-10-31 | `obs_cart_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/cart/break/break_01.png` | 2026-10-31 | `obs_cart_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/cart/break/break_02.png` | 2026-10-31 | `obs_cart_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/cart/break/break_03.png` | 2026-10-31 | `obs_cart_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/cart/break/break_04.png` | 2026-10-31 | `obs_cart_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/cart/break/break_05.png` | 2026-10-31 | `obs_cart_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/cart/cracked.png` | 2026-10-31 | `obs_cart_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/cart/cracked_vertical.png` | 2026-10-31 | `obs_cart_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/cart/nearly_broken.png` | 2026-10-31 | `obs_cart_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/cart/nearly_broken_vertical.png` | 2026-10-31 | `obs_cart_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/cart/rubble.png` | 2026-10-31 | `obs_cart_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/cart/rubble_vertical.png` | 2026-10-31 | `obs_cart_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/fence/break/break_00.png` | 2026-10-31 | `obs_fence_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/fence/break/break_01.png` | 2026-10-31 | `obs_fence_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/fence/break/break_02.png` | 2026-10-31 | `obs_fence_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/fence/break/break_03.png` | 2026-10-31 | `obs_fence_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/fence/break/break_04.png` | 2026-10-31 | `obs_fence_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/fence/break/break_05.png` | 2026-10-31 | `obs_fence_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/fence/break/break_06.png` | 2026-10-31 | `obs_fence_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/fence/cracked.png` | 2026-10-31 | `obs_fence_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/fence/cracked_vertical.png` | 2026-10-31 | `obs_fence_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/fence/nearly_broken.png` | 2026-10-31 | `obs_fence_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/fence/nearly_broken_vertical.png` | 2026-10-31 | `obs_fence_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/fence/rubble.png` | 2026-10-31 | `obs_fence_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/fence/rubble_vertical.png` | 2026-10-31 | `obs_fence_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stall/break/break_00.png` | 2026-10-31 | `obs_stall_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stall/break/break_01.png` | 2026-10-31 | `obs_stall_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stall/break/break_02.png` | 2026-10-31 | `obs_stall_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stall/break/break_03.png` | 2026-10-31 | `obs_stall_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stall/break/break_04.png` | 2026-10-31 | `obs_stall_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stall/break/break_05.png` | 2026-10-31 | `obs_stall_break` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stall/cracked.png` | 2026-10-31 | `obs_stall_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stall/cracked_vertical.png` | 2026-10-31 | `obs_stall_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stall/nearly_broken.png` | 2026-10-31 | `obs_stall_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stall/nearly_broken_vertical.png` | 2026-10-31 | `obs_stall_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stall/rubble.png` | 2026-10-31 | `obs_stall_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stall/rubble_vertical.png` | 2026-10-31 | `obs_stall_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stone_wall/cracked.png` | 2026-10-31 | `obs_stone_wall_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stone_wall/cracked_vertical.png` | 2026-10-31 | `obs_stone_wall_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stone_wall/nearly_broken.png` | 2026-10-31 | `obs_stone_wall_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stone_wall/nearly_broken_vertical.png` | 2026-10-31 | `obs_stone_wall_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stone_wall/rubble.png` | 2026-10-31 | `obs_stone_wall_states` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |
| `obstacles/stone_wall/rubble_vertical.png` | 2026-10-31 | `obs_stone_wall_states_vertical` — prompt: `git show 79a2d04^:server/cmd/artgen/catalog.go` | xoá ở 79a2d04, cứu lại 2026-11-01 |

## obstacle_regrow

Hoạt ảnh mọc lại `regrow/`. Client không nạp (QUESTIONS Q31 nhóm 1).

| Tệp | Ngày | Job / prompt gốc | Ghi chú |
|---|---|---|---|
| `obstacles/bamboo_ancient/regrow/regrow_00.png` | 2026-10-31 | `obs_bamboo_ancient_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/bamboo_ancient/regrow/regrow_01.png` | 2026-10-31 | `obs_bamboo_ancient_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/bamboo_ancient/regrow/regrow_02.png` | 2026-10-31 | `obs_bamboo_ancient_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/bamboo_ancient/regrow/regrow_03.png` | 2026-10-31 | `obs_bamboo_ancient_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/bamboo_ancient/regrow/regrow_04.png` | 2026-10-31 | `obs_bamboo_ancient_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/bamboo_ancient/regrow/regrow_05.png` | 2026-10-31 | `obs_bamboo_ancient_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/bamboo_thin/regrow/regrow_00.png` | 2026-10-31 | `obs_bamboo_thin_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/bamboo_thin/regrow/regrow_01.png` | 2026-10-31 | `obs_bamboo_thin_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/bamboo_thin/regrow/regrow_02.png` | 2026-10-31 | `obs_bamboo_thin_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/bamboo_thin/regrow/regrow_03.png` | 2026-10-31 | `obs_bamboo_thin_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/bamboo_thin/regrow/regrow_04.png` | 2026-10-31 | `obs_bamboo_thin_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/regrow/regrow_00.png` | 2026-10-31 | `obs_bamboo_wall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/regrow/regrow_01.png` | 2026-10-31 | `obs_bamboo_wall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/regrow/regrow_02.png` | 2026-10-31 | `obs_bamboo_wall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/regrow/regrow_03.png` | 2026-10-31 | `obs_bamboo_wall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/regrow/regrow_04.png` | 2026-10-31 | `obs_bamboo_wall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/bamboo_wall/regrow/regrow_05.png` | 2026-10-31 | `obs_bamboo_wall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/boulder/regrow/regrow_00.png` | 2026-10-31 | `obs_boulder_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/boulder/regrow/regrow_01.png` | 2026-10-31 | `obs_boulder_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/boulder/regrow/regrow_02.png` | 2026-10-31 | `obs_boulder_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/boulder/regrow/regrow_03.png` | 2026-10-31 | `obs_boulder_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/boulder/regrow/regrow_04.png` | 2026-10-31 | `obs_boulder_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/boulder/regrow/regrow_05.png` | 2026-10-31 | `obs_boulder_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/brick/regrow/regrow_00.png` | 2026-10-31 | `obs_brick_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/brick/regrow/regrow_01.png` | 2026-10-31 | `obs_brick_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/brick/regrow/regrow_02.png` | 2026-10-31 | `obs_brick_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/brick/regrow/regrow_03.png` | 2026-10-31 | `obs_brick_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/brick/regrow/regrow_04.png` | 2026-10-31 | `obs_brick_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/brick/regrow/regrow_05.png` | 2026-10-31 | `obs_brick_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/cart/regrow/regrow_00.png` | 2026-10-31 | `obs_cart_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/cart/regrow/regrow_01.png` | 2026-10-31 | `obs_cart_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/cart/regrow/regrow_02.png` | 2026-10-31 | `obs_cart_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/cart/regrow/regrow_03.png` | 2026-10-31 | `obs_cart_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/cart/regrow/regrow_04.png` | 2026-10-31 | `obs_cart_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/cart/regrow/regrow_05.png` | 2026-10-31 | `obs_cart_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/fence/regrow/regrow_00.png` | 2026-10-31 | `obs_fence_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/fence/regrow/regrow_01.png` | 2026-10-31 | `obs_fence_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/fence/regrow/regrow_02.png` | 2026-10-31 | `obs_fence_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/fence/regrow/regrow_03.png` | 2026-10-31 | `obs_fence_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/fence/regrow/regrow_04.png` | 2026-10-31 | `obs_fence_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/fence/regrow/regrow_05.png` | 2026-10-31 | `obs_fence_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/stall/regrow/regrow_00.png` | 2026-10-31 | `obs_stall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/stall/regrow/regrow_01.png` | 2026-10-31 | `obs_stall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/stall/regrow/regrow_02.png` | 2026-10-31 | `obs_stall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/stall/regrow/regrow_03.png` | 2026-10-31 | `obs_stall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/stall/regrow/regrow_04.png` | 2026-10-31 | `obs_stall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/stall/regrow/regrow_05.png` | 2026-10-31 | `obs_stall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/stall/regrow/regrow_06.png` | 2026-10-31 | `obs_stall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/stone_wall/regrow/regrow_00.png` | 2026-10-31 | `obs_stone_wall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/stone_wall/regrow/regrow_01.png` | 2026-10-31 | `obs_stone_wall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/stone_wall/regrow/regrow_02.png` | 2026-10-31 | `obs_stone_wall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/stone_wall/regrow/regrow_03.png` | 2026-10-31 | `obs_stone_wall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/stone_wall/regrow/regrow_04.png` | 2026-10-31 | `obs_stone_wall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/stone_wall/regrow/regrow_05.png` | 2026-10-31 | `obs_stone_wall_regrow` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |

## obstacle_states_04

Khung trạng thái 04 cũ; v6.7 rút còn một trạng thái (`OBSTACLE_STATES = ['intact']`).

| Tệp | Ngày | Job / prompt gốc | Ghi chú |
|---|---|---|---|
| `obstacles/bamboo_wall/states_04.png` | 2026-10-31 | `obs_bamboo_wall_states` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/brick/states_04.png` | 2026-10-31 | `obs_brick_states` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `obstacles/fence/states_04.png` | 2026-10-31 | `obs_fence_states` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |

## vfx_boba

Viên trân châu: `boba.ts` vẽ bằng canvas, không dùng ảnh (Q31 nhóm 1).

| Tệp | Ngày | Job / prompt gốc | Ghi chú |
|---|---|---|---|
| `vfx/boba_1.png` | 2026-10-31 | `vfx_boba_1` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `vfx/boba_2.png` | 2026-10-31 | `vfx_boba_2` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `vfx/boba_3.png` | 2026-10-31 | `vfx_boba_3` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `vfx/boba_4.png` | 2026-10-31 | `vfx_boba_4` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `vfx/boba_5.png` | 2026-10-31 | `vfx_boba_5` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `vfx/boba_6.png` | 2026-10-31 | `vfx_boba_6` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `vfx/boba_drip.png` | 2026-10-31 | `vfx_boba_drip` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |

## buffs_removed

Icon buff Trường Kiếm — buff đã bỏ từ spec v4.3.

| Tệp | Ngày | Job / prompt gốc | Ghi chú |
|---|---|---|---|
| `buffs/long.png` | 2026-10-31 | `buff_long` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |

## extra_frames

Khung dư ngoài số khung danh mục khai (burst/open 6 khung khai tới 05, run_dust khai 4 khung).

| Tệp | Ngày | Job / prompt gốc | Ghi chú |
|---|---|---|---|
| `gates/blue/burst_06.png` | 2026-10-31 | `gate_blue` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `gates/orange/burst_06.png` | 2026-10-31 | `gate_orange` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `gates/purple/burst_06.png` | 2026-10-31 | `gate_purple` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `items/chest/open_06.png` | 2026-10-31 | `chest` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `vfx/run_dust/run_dust_04.png` | 2026-10-31 | `vfx_run_dust` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |
| `vfx/run_dust/run_dust_05.png` | 2026-10-31 | `vfx_run_dust` — prompt: `git show e8717f1^:server/cmd/artgen/catalog.go` | xoá ở e8717f1, cứu lại 2026-11-01 |

## char_frames_extra_2026-10-22

Khung dư sau khi sinh lại sheet idle/run (sheet mới ít khung hơn).

| Tệp | Ngày | Job / prompt gốc | Ghi chú |
|---|---|---|---|
| `characters/co_ba_ganh/frames/run_08.png` | 2026-10-22 | `char_co_ba_ganh_run` — prompt: `git show 72648c3^:server/cmd/artgen/catalog.go` | xoá ở 72648c3, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_04.png` | 2026-10-22 | `char_dai_ca_dep_idle` — prompt: `git show f954695^:server/cmd/artgen/catalog.go` | xoá ở f954695, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_05.png` | 2026-10-22 | `char_dai_ca_dep_idle` — prompt: `git show f954695^:server/cmd/artgen/catalog.go` | xoá ở f954695, cứu lại 2026-11-01 |
| `characters/ong_ba_vong/frames/idle_04.png` | 2026-10-22 | `char_ong_ba_vong_idle` — prompt: `git show f954695^:server/cmd/artgen/catalog.go` | xoá ở f954695, cứu lại 2026-11-01 |

## char_frames_extra_2026-10-29

Khung dư sau khi sinh lại idle/hit bị cắt hỏng (sheet mới ít khung hơn).

| Tệp | Ngày | Job / prompt gốc | Ghi chú |
|---|---|---|---|
| `characters/a_hon/frames/teleport_out_03.png` | 2026-10-29 | `char_a_hon_teleport_out` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/a_hon/frames/teleport_out_04.png` | 2026-10-29 | `char_a_hon_teleport_out` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/a_hon/frames/teleport_out_05.png` | 2026-10-29 | `char_a_hon_teleport_out` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/co_ba_ganh/frames/hit_04.png` | 2026-10-29 | `char_co_ba_ganh_hit` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/co_ba_ganh/frames/hit_05.png` | 2026-10-29 | `char_co_ba_ganh_hit` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/co_ba_ganh/frames/hit_06.png` | 2026-10-29 | `char_co_ba_ganh_hit` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/co_ba_ganh/frames/hit_07.png` | 2026-10-29 | `char_co_ba_ganh_hit` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/co_ba_ganh/frames/idle_04.png` | 2026-10-29 | `char_co_ba_ganh_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/co_ba_ganh/frames/idle_05.png` | 2026-10-29 | `char_co_ba_ganh_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/co_ba_ganh/frames/idle_06.png` | 2026-10-29 | `char_co_ba_ganh_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/co_ba_ganh/frames/idle_07.png` | 2026-10-29 | `char_co_ba_ganh_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/cu_rua/frames/idle_04.png` | 2026-10-29 | `char_cu_rua_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/cu_rua/frames/idle_05.png` | 2026-10-29 | `char_cu_rua_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/cu_rua/frames/idle_06.png` | 2026-10-29 | `char_cu_rua_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/cu_rua/frames/idle_07.png` | 2026-10-29 | `char_cu_rua_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_04.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_05.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_06.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_07.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_08.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_09.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_10.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_11.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_12.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_13.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_14.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_15.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_16.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_17.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_18.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_19.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_20.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/idle_21.png` | 2026-10-29 | `char_dai_ca_dep_idle` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/su_phu_mi/frames/death_08.png` | 2026-10-29 | `char_su_phu_mi_death` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |
| `characters/su_phu_mi/frames/death_09.png` | 2026-10-29 | `char_su_phu_mi_death` — prompt: `git show 605f896^:server/cmd/artgen/catalog.go` | xoá ở 605f896, cứu lại 2026-11-01 |

## char_frames_extra_2026-10-30

Khung dư bị loại khi chuẩn hoá frame tự động (§12.5 B).

| Tệp | Ngày | Job / prompt gốc | Ghi chú |
|---|---|---|---|
| `characters/ba_tam_loa/frames/hit_04.png` | 2026-10-30 | `char_ba_tam_loa_hit` — prompt: `git show 9204be0^:server/cmd/artgen/catalog.go` | xoá ở 9204be0, cứu lại 2026-11-01 |
| `characters/ba_tam_loa/frames/hit_05.png` | 2026-10-30 | `char_ba_tam_loa_hit` — prompt: `git show 9204be0^:server/cmd/artgen/catalog.go` | xoá ở 9204be0, cứu lại 2026-11-01 |
| `characters/ba_tam_loa/frames/hit_06.png` | 2026-10-30 | `char_ba_tam_loa_hit` — prompt: `git show 9204be0^:server/cmd/artgen/catalog.go` | xoá ở 9204be0, cứu lại 2026-11-01 |
| `characters/ba_tam_loa/frames/hit_07.png` | 2026-10-30 | `char_ba_tam_loa_hit` — prompt: `git show 9204be0^:server/cmd/artgen/catalog.go` | xoá ở 9204be0, cứu lại 2026-11-01 |
| `characters/be_ti/frames/teleport_in_04.png` | 2026-10-30 | `char_be_ti_teleport_in` — prompt: `git show 9204be0^:server/cmd/artgen/catalog.go` | xoá ở 9204be0, cứu lại 2026-11-01 |
| `characters/be_ti/frames/teleport_in_05.png` | 2026-10-30 | `char_be_ti_teleport_in` — prompt: `git show 9204be0^:server/cmd/artgen/catalog.go` | xoá ở 9204be0, cứu lại 2026-11-01 |
| `characters/be_ti/frames/teleport_in_06.png` | 2026-10-30 | `char_be_ti_teleport_in` — prompt: `git show 9204be0^:server/cmd/artgen/catalog.go` | xoá ở 9204be0, cứu lại 2026-11-01 |
| `characters/be_ti/frames/teleport_in_07.png` | 2026-10-30 | `char_be_ti_teleport_in` — prompt: `git show 9204be0^:server/cmd/artgen/catalog.go` | xoá ở 9204be0, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/hit_04.png` | 2026-10-30 | `char_dai_ca_dep_hit` — prompt: `git show 9204be0^:server/cmd/artgen/catalog.go` | xoá ở 9204be0, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/hit_05.png` | 2026-10-30 | `char_dai_ca_dep_hit` — prompt: `git show 9204be0^:server/cmd/artgen/catalog.go` | xoá ở 9204be0, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/hit_06.png` | 2026-10-30 | `char_dai_ca_dep_hit` — prompt: `git show 9204be0^:server/cmd/artgen/catalog.go` | xoá ở 9204be0, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/hit_07.png` | 2026-10-30 | `char_dai_ca_dep_hit` — prompt: `git show 9204be0^:server/cmd/artgen/catalog.go` | xoá ở 9204be0, cứu lại 2026-11-01 |
| `characters/dai_ca_dep/frames/teleport_out_03.png` | 2026-10-30 | `char_dai_ca_dep_teleport_out` — prompt: `git show 9cff870^:server/cmd/artgen/catalog.go` | xoá ở 9cff870, cứu lại 2026-11-01 |
| `characters/map_tra_sua/frames/teleport_out_03.png` | 2026-10-30 | `char_map_tra_sua_teleport_out` — prompt: `git show 9204be0^:server/cmd/artgen/catalog.go` | xoá ở 9204be0, cứu lại 2026-11-01 |
| `characters/map_tra_sua/frames/teleport_out_04.png` | 2026-10-30 | `char_map_tra_sua_teleport_out` — prompt: `git show 9204be0^:server/cmd/artgen/catalog.go` | xoá ở 9204be0, cứu lại 2026-11-01 |
| `characters/map_tra_sua/frames/teleport_out_05.png` | 2026-10-30 | `char_map_tra_sua_teleport_out` — prompt: `git show 9204be0^:server/cmd/artgen/catalog.go` | xoá ở 9204be0, cứu lại 2026-11-01 |
| `characters/map_tra_sua/frames/teleport_out_06.png` | 2026-10-30 | `char_map_tra_sua_teleport_out` — prompt: `git show 9204be0^:server/cmd/artgen/catalog.go` | xoá ở 9204be0, cứu lại 2026-11-01 |

## q31_chua_khu_nao_dung

Art có sẵn nhưng chưa khu nào dùng (Q31 nhóm 2). Trước nằm trong `assets/src` + NEVER_SHIP; nay chuyển hẳn vào thư viện.

| Tệp | Ngày | Job / prompt gốc | Ghi chú |
|---|---|---|---|
| `decals/frozen_reeds.png` | 2026-11-01 | `decal_frozen_reeds` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `decals/ice_skid.png` | 2026-11-01 | `decal_ice_skid` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `decals/snow_pile.png` | 2026-11-01 | `decal_snow_pile` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `obstacles/bamboo_corner.png` | 2026-11-01 | `zone_bamboo_corner` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `obstacles/cliff_wall.png` | 2026-11-01 | `zone_cliff_wall` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `obstacles/ice_island_edge.png` | 2026-11-01 | `zone_ice_island_edge` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `obstacles/lantern_string.png` | 2026-11-01 | `zone_lantern_string` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `obstacles/rope_bridge.png` | 2026-11-01 | `zone_rope_bridge` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `obstacles/village_roof.png` | 2026-11-01 | `zone_village_roof` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `tiles/cobblestone.png` | 2026-11-01 | `tile_cobblestone` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `tiles/damp_earth.png` | 2026-11-01 | `tile_damp_earth` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `tiles/mossy_stone.png` | 2026-11-01 | `tile_mossy_stone` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `tiles/mountain_rock.png` | 2026-11-01 | `tile_mountain_rock` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |

## q31_dan_em

Atlas riêng của Đàn Em — không nằm trong 10 nhân vật; đàn em vẽ bằng atlas Đại Ca thu nhỏ (spec v5.4 mục 19).

| Tệp | Ngày | Job / prompt gốc | Ghi chú |
|---|---|---|---|
| `characters/dan_em/atlas.json` | 2026-11-01 | `char_dan_em_run` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `characters/dan_em/atlas.png` | 2026-11-01 | `char_dan_em_run` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `characters/dan_em/frames/run_00.png` | 2026-11-01 | `char_dan_em_run` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `characters/dan_em/frames/run_01.png` | 2026-11-01 | `char_dan_em_run` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `characters/dan_em/frames/run_02.png` | 2026-11-01 | `char_dan_em_run` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `characters/dan_em/frames/run_03.png` | 2026-11-01 | `char_dan_em_run` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `characters/dan_em/frames/run_04.png` | 2026-11-01 | `char_dan_em_run` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `characters/dan_em/frames/run_05.png` | 2026-11-01 | `char_dan_em_run` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `characters/dan_em/frames/run_06.png` | 2026-11-01 | `char_dan_em_run` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |
| `characters/dan_em/frames/run_07.png` | 2026-11-01 | `char_dan_em_run` — prompt trong `server/cmd/artgen/catalog.go` hiện tại | chuyển từ assets/src (git mv) |

## bi_kip_v6.25

Spec v6.25 §6.3 bỏ vật phẩm Bí Kíp (kỹ năng tự hồi chiêu + tích lượt theo nhân vật).

| Tệp | Ngày | Job / prompt gốc | Ghi chú |
|---|---|---|---|
| `items/skill_scroll.png` | 2026-11-01 | `item_skill_scroll` — prompt trong `server/cmd/artgen/catalog.go` | chuyển từ assets/src (git mv) |
| `ui/icon_skill_scroll.png` | 2026-11-01 | `icon_skill_scroll` — prompt trong `server/cmd/artgen/catalog.go` | chuyển từ assets/src (git mv); trước nằm cạnh số lượt trên nút kỹ năng |

## Bản cũ của ảnh bị ghi đè — KHÔNG chuyển vào thư viện, lấy lại bằng git

Lịch sử git còn **396 bản cũ** của **196 ảnh**. Mỗi bản là ảnh trước khi artgen sinh lại
hoặc cắt lại. Chủ dự án chốt 1/11 (Q33-c): không chép ra đây, vì git đã giữ trọn và chép ra
thì repo phình thêm mà không thêm thông tin gì.

**Lấy một ảnh** ở trạng thái NGAY TRƯỚC một đợt ghi đè:

```bash
git show <commit>^:<đường dẫn> > /tmp/anh_cu.png
# ví dụ: git show 605f896^:assets/characters/a_hon/frames/idle_00.png > /tmp/a.png
```

**Lấy cả một đợt** về một thư mục tạm, để xem cạnh nhau:

```bash
C=<commit>; mkdir -p /tmp/dot_$C
git diff-tree --no-commit-id --name-only --diff-filter=M -r $C -- "assets/*.png" |
  while read f; do mkdir -p "/tmp/dot_$C/$(dirname "$f")"; git show "$C^:$f" > "/tmp/dot_$C/$f"; done
```

**Xem mọi bản** của một ảnh: `git log --format="%h %ad %s" --date=short -- <đường dẫn>`.

Đường dẫn trước ngày 29/10 là `assets/<…>`, sau đó là `assets/src/<…>` (git tự theo dấu đổi tên
khi dùng `git log --follow`).

| Commit (đợt) | Ngày | Số ảnh bị ghi đè | Nội dung đợt |
|---|---|---|---|
| `9b0e7a8` | 2026-11-01 | 1 | Vương miện: artgen vẽ lại crown_small (chỉ vương miện) + kiểm icon UI không chứa người |
| `611558a` | 2026-10-31 | 8 | Lượt 3.1 — phá cảm giác bàn cờ, vòng tre quanh Thiên Môn + ART THẬT 16 loại |
| `8029741` | 2026-10-30 | 17 | dep_lao idle hết đỏ (56.9 → 44.1); bắt "sai dáng"; art thử hướng A |
| `9cff870` | 2026-10-30 | 16 | C: bộ điều khiển hoạt ảnh; B hoàn tất — giật lặp -24%, chỉ còn 1 anim đỏ |
| `9204be0` | 2026-10-30 | 42 | B: chuẩn hoá frame tự động (GAME_SPEC v6.3 §12.5 B) — giật lặp -16%, một lần -57% |
| `ab0f706` | 2026-10-30 | 34 | Sửa lỗi mất hình trên laptop: bộ lọc nền cấp phát texture 1 GB; art thừa 3-8 lần |
| `c52bd11` | 2026-10-30 | 10 | Art: 32 ảnh mới (Bí Kíp, icon 10 kỹ năng, VFX trân châu, chân dung, decor khu) |
| `b86ff87` | 2026-10-29 | 11 | HUD trái xếp dọc bằng flex nên không tự đè nhau; vá magenta sót trong 7 ảnh nền |
| `605f896` | 2026-10-29 | 50 | Art: thêm 23 ảnh (sinh lại idle/hit bị cắt hỏng + hiệu ứng dịch chuyển, lên cấp); ART_JOBS.md còn 94 ảnh |
| `2aada2d` | 2026-10-28 | 6 | Art: thêm 62 ảnh từ hai đợt sinh (thần thú, hiệu ứng nhân vật, huy hiệu, bảng tên); ART_JOBS.md còn 113 ảnh |
| `e436652` | 2026-10-28 | 1 | Sửa Đại Ca Dép idle thành cái áo tím nhấp nháy; chặn sheet cắt sai ở cả hai đầu |
| `d4b9dfa` | 2026-10-28 | 1 | Sửa lag: bể kiếm bắn 1 867 sự kiện/giây; viên EXP vẽ bằng ParticleContainer |
| `4e92227` | 2026-10-22 | 12 | Sửa lỗi nhân vật mình luôn bị vẽ mờ; bảng xếp hạng + danh hiệu to rõ hơn; kill feed tô màu theo vai; hướng dẫn |
| `9d7870b` | 2026-10-22 | 14 | Art: chase/flee 10 nhân vật, levelup/đột phá/rớt cảnh giới/hồi sinh/vào cổng của A Hớn, hit/death còn lại, VFX |
| `72648c3` | 2026-10-22 | 12 | Art: khung thoại, rương, hit/death Ông Ba Vòng, death Đại Ca, hit Thằng Mập; artgen cắt theo lưới khi sheet nh |
| `64a9b7f` | 2026-10-22 | 87 | Art ưu tiên 1: quái, nền, bụi, vật cản, cổng, Shipper, Cụ Rùa; artgen cắt theo đường kẻ lưới/ô, nhận nền magen |
| `a6fcda9` | 2026-10-22 | 17 | Art: idle/run Bé Tí, Sư Phụ Mì, Cô Ba Gánh, run Thằng Mập; artgen xóa đường kẻ lưới, bỏ chấm watermark xa hình |
| `f954695` | 2026-10-22 | 55 | Art bựa mới: idle/run của A Hớn, Bà Tám, Ông Ba, Đại Ca + idle Thằng Mập; hệ số chiều cao theo anim để nhân vậ |
| `aa95fbb` | 2026-10-21 | 2 | M0–M3: client Vite+PixiJS (nội suy, dự đoán, HUD, trang chủ, tổng kết, bảng xếp hạng), cmd/sim, test phòng/bot |
