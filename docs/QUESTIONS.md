# Câu hỏi đang chờ anh trả lời

Mỗi câu: bối cảnh → phương án → **tạm dùng** (đã code theo phương án này, đổi lại dễ).
Câu đã xong ở `docs/QUESTIONS_DONE.md`.

---

## Q27 — Tên vật trang trí: đổi mapgen hay vẽ thêm art?

**Bối cảnh**: làm spec v6.4 phần A (§13.7), lúc phải trả lời "tệp art này thuộc khu nào" thì
lộ ra `mapview.buildDecor()` **chưa bao giờ vẽ được một vật trang trí nào**. Nó tìm
`decor/<tên>.png` với tên lấy nguyên từ `type` trong Tiled, mà hai bộ tên không trùng chữ nào:

| mapgen đặt vào `.tmj` (299 vật trong 5 bản đồ) | `assets/src/decor/` đang có |
|---|---|
| `rock_small`, `pine`, `bamboo`, `bush_small`, `lantern`, `well`, `banner` | `red_plastic_stool`, `old_bicycle_leaning`, `pile_of_watermelons`, `street_food_cart`, `banana_tree`, `rusty_water_jar`, `broken_tv`, `lazy_sleeping_dog`, `bush_with_flowers`, `stone_lantern`, `wall_border` |

Bộ bên phải là bộ "tấu hài phố" trong ART_PROMPTS §9; bộ bên trái là cây cỏ/đá/giếng mapgen
sinh ra theo địa hình từng khu. Hai bộ sinh ra ở hai thời điểm khác nhau và chưa ai nối lại.

**Phương án**:
- **(a)** Đổi `cmd/mapgen` để đặt đúng 11 tên art đã có. Không cần vẽ gì. Nhưng Linh Sơn Cốc
  sẽ rải ghế nhựa và xe đạp cũ giữa hẻm núi — hợp Phàm Trấn, không hợp núi đá.
- **(b)** Vẽ thêm 7 ảnh theo tên mapgen đang dùng (đá nhỏ, thông, tre, bụi nhỏ, đèn lồng,
  giếng, băng-rôn). Mỗi khu được bộ trang trí đúng chất của nó.
- **(c)** Cả hai: đổi mapgen cho Phàm Trấn / Thiên Đình dùng bộ phố đã có, và vẽ thêm cho
  Linh Sơn Cốc / Trúc Lâm / Băng Nguyệt Hồ.

**Tạm dùng**: **chưa code gì** — hiện trạng là không vẽ trang trí, và nó đã như vậy từ trước
phiên này. Tôi không tự chọn vì cả ba phương án đều **đổi hình ảnh trong game**, là việc của
anh. Bộ `decor/` (0,13 MB) vẫn phát hành trong nhóm `boot`, nên chọn phương án nào cũng không
phải dựng lại asset.

Kèm theo, cùng gốc nhưng **không phải câu hỏi** (là lỗi rõ, chỉ chờ anh ừ): nền của cả 5 khu
đang đều là cỏ, vì `BIOME_TILES` đánh khoá bằng tên khu (`'Linh Sơn'`) còn `map.biome` là tên
biome (`'dirt'`). 14 trong 19 ảnh nền chưa từng hiện ra, mà `tiles/snow.png` / `dirt.png` /
`concrete.png` thì có sẵn và tên khớp đúng giá trị `biome`. Anh ừ thì tôi đổi khoá.

---

*(Ngoài Q27, không còn câu nào đang chờ — tính tới 30/10/2026.)*

Q20, Q22, Q25 và Q26 anh đã trả lời trong `docs/_raw/ANSWERS.md` ngày 30/10; code đã làm
theo và cả bốn đã chuyển sang `QUESTIONS_DONE.md`. Hai việc "chờ anh quyết" — tối ưu
CHANGES_v4 mục 25–27, và hai thư viện `pixi-filters` / `revolt-fx` — cũng đã được duyệt,
xem cuối `QUESTIONS_DONE.md`.

**Việc lớn tiếp theo** (không phải câu hỏi), cập nhật 30/10 chiều:
**NETCODE bước 3 — Protocol v2** (NETCODE_SPEC v1.2): schema chung + generator + bài đối
chiếu golden/fuzz TRƯỚC, rồi vòng đời `Spawn/Desc/Update/Despawn` tường minh, delta theo ack,
keyframe 2s, lượng tử hoá §4.6. Mục tiêu: đóng nốt respawn (về 0) và gói < 6 KB.

Tối ưu CHANGES_v4 mục 25–27 **đã xong** (mốc chuẩn #4 trong PERF_LOG). Hướng A (một thân
nhiều mặt) để sau cùng — art thử đã sinh cho 3 nhân vật, chưa code.
