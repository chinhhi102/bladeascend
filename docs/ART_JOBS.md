# Danh sách prompt cần sinh ảnh

Tạo lại file này: `cd server && go run ./cmd/artgen prompts` (chỉ liệt kê ảnh còn thiếu, xếp theo ưu tiên).

**Cách dùng**
1. Dán prompt vào công cụ sinh ảnh bất kỳ (Gemini, ChatGPT, Midjourney…). Nếu công cụ trả lời bằng chữ, thêm vào đầu: `Generate an image:`.
2. Ảnh nên khổ ngang 16:9 (hoặc vuông với ảnh 1 món), độ phân giải càng cao càng tốt.
3. Lưu đúng tên ở dòng **Lưu thành** (đuôi `.png`; ảnh JPEG đổi tên thành .png vẫn được).
4. Chạy `cd server && go run ./cmd/artgen prep` → tool tự xóa nền magenta/trắng, cắt frame, căn pivot, ghép atlas.
5. Sheet lệch nhẹ vẫn dùng được, tool cắt theo **vùng pixel liền nhau**.

**Quan trọng với sheet nhân vật** — tool đếm khung bằng cách tìm các mảng pixel rời nhau, nên:
- Mỗi khung phải là MỘT khối liền: tay, chân, vũ khí, nón phải DÍNH vào thân. Một cánh tay bay rời là tool đếm thành một khung nữa.
- Giữa hai khung phải có khoảng trống nền rõ ràng, các khung không được chạm nhau.
- Không vẽ thêm bóng đổ rời, hạt bụi, vệt tốc độ tách khỏi thân.
- Sai số khung quá nhiều thì client **bỏ hẳn animation đó** và dùng tạm `idle`/`run`, nên các mục ghi "sheet bị cắt sai số khung" bên dưới cần sinh lại thật.

## 1. `sword_realm_0` · ưu tiên 0

- **Sinh lại**: spec v3.1: lưỡi kiếm trắng/xám để game nhuộm màu theo vòng (ART_PROMPTS §7)
- **Lưu thành**: `assets/_raw/sword_realm_0.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: swords/realm_0.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Single game weapon icon, a pair of bamboo chopsticks tied together as a sword, perfectly vertical, tip pointing straight up, handle at the bottom, centered. Blade painted in very light grey and white tones so it can be color-tinted in game, colored details only on the handle. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 2. `sword_realm_1` · ưu tiên 0

- **Sinh lại**: spec v3.1: lưỡi kiếm trắng/xám để game nhuộm màu theo vòng (ART_PROMPTS §7)
- **Lưu thành**: `assets/_raw/sword_realm_1.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: swords/realm_1.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Single game weapon icon, a kitchen cleaver with a wooden handle, perfectly vertical, tip pointing straight up, handle at the bottom, centered. Blade painted in very light grey and white tones so it can be color-tinted in game, colored details only on the handle. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 3. `sword_realm_2` · ưu tiên 0

- **Sinh lại**: spec v3.1: lưỡi kiếm trắng/xám để game nhuộm màu theo vòng (ART_PROMPTS §7)
- **Lưu thành**: `assets/_raw/sword_realm_2.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: swords/realm_2.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Single game weapon icon, a peach wood taoist sword with red tassel, perfectly vertical, tip pointing straight up, handle at the bottom, centered. Blade painted in very light grey and white tones so it can be color-tinted in game, colored details only on the handle. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 4. `sword_realm_3` · ưu tiên 0

- **Sinh lại**: spec v3.1: lưỡi kiếm trắng/xám để game nhuộm màu theo vòng (ART_PROMPTS §7)
- **Lưu thành**: `assets/_raw/sword_realm_3.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: swords/realm_3.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Single game weapon icon, a polished steel jian sword, perfectly vertical, tip pointing straight up, handle at the bottom, centered. Blade painted in very light grey and white tones so it can be color-tinted in game, colored details only on the handle. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 5. `sword_realm_4` · ưu tiên 0

- **Sinh lại**: spec v3.1: lưỡi kiếm trắng/xám để game nhuộm màu theo vòng (ART_PROMPTS §7)
- **Lưu thành**: `assets/_raw/sword_realm_4.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: swords/realm_4.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Single game weapon icon, a flaming red-orange sword with fire trail, perfectly vertical, tip pointing straight up, handle at the bottom, centered. Blade painted in very light grey and white tones so it can be color-tinted in game, colored details only on the handle. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 6. `sword_realm_5` · ưu tiên 0

- **Sinh lại**: spec v3.1: lưỡi kiếm trắng/xám để game nhuộm màu theo vòng (ART_PROMPTS §7)
- **Lưu thành**: `assets/_raw/sword_realm_5.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: swords/realm_5.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Single game weapon icon, an ice crystal blue sword with frost sparkles, perfectly vertical, tip pointing straight up, handle at the bottom, centered. Blade painted in very light grey and white tones so it can be color-tinted in game, colored details only on the handle. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 7. `sword_realm_6` · ưu tiên 0

- **Sinh lại**: spec v3.1: lưỡi kiếm trắng/xám để game nhuộm màu theo vòng (ART_PROMPTS §7)
- **Lưu thành**: `assets/_raw/sword_realm_6.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: swords/realm_6.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Single game weapon icon, a crackling lightning yellow sword, perfectly vertical, tip pointing straight up, handle at the bottom, centered. Blade painted in very light grey and white tones so it can be color-tinted in game, colored details only on the handle. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 8. `sword_realm_7` · ưu tiên 0

- **Sinh lại**: spec v3.1: lưỡi kiếm trắng/xám để game nhuộm màu theo vòng (ART_PROMPTS §7)
- **Lưu thành**: `assets/_raw/sword_realm_7.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: swords/realm_7.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Single game weapon icon, a dark purple cursed sword with smoky aura, perfectly vertical, tip pointing straight up, handle at the bottom, centered. Blade painted in very light grey and white tones so it can be color-tinted in game, colored details only on the handle. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 9. `sword_realm_8` · ưu tiên 0

- **Sinh lại**: spec v3.1: lưỡi kiếm trắng/xám để game nhuộm màu theo vòng (ART_PROMPTS §7)
- **Lưu thành**: `assets/_raw/sword_realm_8.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: swords/realm_8.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Single game weapon icon, a radiant jade and gold celestial sword glowing white, perfectly vertical, tip pointing straight up, handle at the bottom, centered. Blade painted in very light grey and white tones so it can be color-tinted in game, colored details only on the handle. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 10. `char_dai_ca_dep_idle` · ưu tiên 1

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/char_dai_ca_dep_idle.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/dai_ca_dep/frames (idle ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a pompous short gang boss in an oversized shiny purple suit with shoulder pads, tiny round black sunglasses, a gold chain as thick as a dog leash, a thin pencil mustache, slicked-back hair so greasy that flies buzz around it, a gold front tooth, cheap honeycomb plastic sandals with socks, color palette #7209B7 #FFD60A #000000 #F72585 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [idle, 4 frames] 1) standing breathing out 2) breathing in, shoulders raised 3) yawning with mouth wide open 4) scratching head, bored sleepy face; comedy gag of this character in this animation: admiring his gold tooth in a tiny hand mirror. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 11. `char_co_ba_ganh_idle` · ưu tiên 1

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/char_co_ba_ganh_idle.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/co_ba_ganh/frames (idle ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a hyperactive young fruit seller with one thick unibrow, a huge gap-toothed grin, two stiff braids sticking straight out sideways, a crooked conical hat, balancing a bamboo shoulder pole with two overflowing fruit baskets, patched brown blouse, mismatched plastic slippers, color palette #BC6C25 #DDA15E #606C38 #FEFAE0 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [idle, 4 frames] 1) standing breathing out 2) breathing in, shoulders raised 3) yawning with mouth wide open 4) scratching head, bored sleepy face. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 12. `char_cu_rua_idle` · ưu tiên 1

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/char_cu_rua_idle.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/frames (idle ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, color palette #52B788 #D8F3DC #6F4E37 #F1FAEE with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [idle, 4 frames] 1) standing breathing out 2) breathing in, shoulders raised 3) yawning with mouth wide open 4) scratching head, bored sleepy face. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 13. `tile_grass` · ưu tiên 1

- **Sinh lại**: spec v3: nền nhạt, ít chi tiết (ART_PROMPTS §9.0)
- **Lưu thành**: `assets/_raw/tile_grass.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/grass.png

```text
STYLE: 2D mobile game art, cartoon, flat colors with simple cel shading, vibrant saturated palette, clean vector-like look, no text, no watermark. Seamless tileable top-down ground texture, soft muted sage green grass, very low contrast, subtle large brush strokes, no flowers, no small colorful dots, calm and plain, square image filling the whole frame, cartoon game style, no objects, no characters, even lighting, edges must tile perfectly. Desaturated, low detail, designed as a background that does not distract from gameplay objects.
```

## 14. `tile_concrete` · ưu tiên 1

- **Sinh lại**: spec v3: nền nhạt, ít chi tiết (ART_PROMPTS §9.0)
- **Lưu thành**: `assets/_raw/tile_concrete.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/concrete.png

```text
STYLE: 2D mobile game art, cartoon, flat colors with simple cel shading, vibrant saturated palette, clean vector-like look, no text, no watermark. Seamless tileable top-down ground texture, pale warm grey concrete, very low contrast, few faint cracks, plain and calm, square image filling the whole frame, cartoon game style, no objects, no characters, even lighting, edges must tile perfectly. Desaturated, low detail, designed as a background that does not distract from gameplay objects.
```

## 15. `tile_dirt` · ưu tiên 1

- **Sinh lại**: spec v3: nền nhạt, ít chi tiết (ART_PROMPTS §9.0)
- **Lưu thành**: `assets/_raw/tile_dirt.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/dirt.png

```text
STYLE: 2D mobile game art, cartoon, flat colors with simple cel shading, vibrant saturated palette, clean vector-like look, no text, no watermark. Seamless tileable top-down ground texture, light muted sandy brown dirt, very low contrast, only a few faint large pebbles, plain and calm, square image filling the whole frame, cartoon game style, no objects, no characters, even lighting, edges must tile perfectly. Desaturated, low detail, designed as a background that does not distract from gameplay objects.
```

## 16. `char_ba_tam_loa_hit` · ưu tiên 2

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/char_ba_tam_loa_hit.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/ba_tam_loa/frames (hit ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a stout loud fish market lady holding a dented megaphone, bright red lipstick smeared far outside her lips and stuck on her teeth, two perfectly round blotchy pink blush circles, eyebrows drawn way too high in pencil, a tiny pencil stuck in her hair bun, floral blouse with fish scales stuck on it, wet rubber apron, tall green rubber boots, color palette #2A9D8F #E9C46A #F28482 #264653 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [hit, 4 frames] 1) face squashed flat to the left 2) body knocked backwards 3) spiral eyes with stars around the head 4) shaking head recovering; comedy gag of this character in this animation: her lipstick mark gets printed on the air in front of her. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 17. `char_dai_ca_dep_hit` · ưu tiên 2

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/char_dai_ca_dep_hit.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/dai_ca_dep/frames (hit ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a pompous short gang boss in an oversized shiny purple suit with shoulder pads, tiny round black sunglasses, a gold chain as thick as a dog leash, a thin pencil mustache, slicked-back hair so greasy that flies buzz around it, a gold front tooth, cheap honeycomb plastic sandals with socks, color palette #7209B7 #FFD60A #000000 #F72585 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [hit, 4 frames] 1) face squashed flat to the left 2) body knocked backwards 3) spiral eyes with stars around the head 4) shaking head recovering; comedy gag of this character in this animation: his sunglasses fall off revealing cross-eyes. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 18. `char_co_ba_ganh_hit` · ưu tiên 2

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/char_co_ba_ganh_hit.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/co_ba_ganh/frames (hit ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a hyperactive young fruit seller with one thick unibrow, a huge gap-toothed grin, two stiff braids sticking straight out sideways, a crooked conical hat, balancing a bamboo shoulder pole with two overflowing fruit baskets, patched brown blouse, mismatched plastic slippers, color palette #BC6C25 #DDA15E #606C38 #FEFAE0 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [hit, 4 frames] 1) face squashed flat to the left 2) body knocked backwards 3) spiral eyes with stars around the head 4) shaking head recovering; comedy gag of this character in this animation: both braids curl up into springs. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 19. `char_su_phu_mi_death` · ưu tiên 2

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/char_su_phu_mi_death.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình · hàng 2: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/su_phu_mi/frames (death ×8)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of an eccentric old kung fu master whose long beard and eyebrows are made of curly instant noodles, cross-eyed, bright red bulbous nose, one eye squinting, stained orange robe with soup splashes, holding a pair of chopsticks like a weapon, color palette #FB8500 #FFB703 #8ECAE6 #023047 with 8 frames in total. Exactly 4 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [death, 8 frames] 1) getting smashed 2) smashed harder 3) flattened like rice paper 4) completely flat 5) white ghost soul rising out 6) ghost waving goodbye floating up 7) ghost floating higher 8) ghost almost gone at the top. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 20. `tile_mossy_stone` · ưu tiên 2

- **Lưu thành**: `assets/_raw/tile_mossy_stone.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/mossy_stone.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. patchy moss on stone, seamless tileable top-down, desaturated, low contrast, low detail, no objects
```

## 21. `plate_shine_streak` · ưu tiên 3

- **Lưu thành**: `assets/_raw/plate_shine_streak.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: ui/shine_streak.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. soft diagonal white light streak Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 22. `char_a_hon_teleport_out` · ưu tiên 4

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/char_a_hon_teleport_out.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/a_hon/frames (teleport_out ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a loud-mouthed scrawny street vendor with a giant steamed bun as a hat that has a big bite missing, two huge yellowish buck teeth with one chipped, a mole on his cheek with one long curly hair, sweat-stained yellowed sleeveless undershirt riding up over a round hairy belly button, baggy red shorts, mismatched flip-flops, color palette #FFFFFF #E63946 #F4A261 #2B2D42 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [teleport_out, 4 frames] 1) startled 2) body twisting like being sucked into a vortex 3) only the sobbing head left 4) only a pair of flip-flops left. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 23. `char_co_ba_ganh_teleport_out` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_co_ba_ganh_teleport_out.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/co_ba_ganh/frames (teleport_out ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a hyperactive young fruit seller with one thick unibrow, a huge gap-toothed grin, two stiff braids sticking straight out sideways, a crooked conical hat, balancing a bamboo shoulder pole with two overflowing fruit baskets, patched brown blouse, mismatched plastic slippers, color palette #BC6C25 #DDA15E #606C38 #FEFAE0 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [teleport_out, 4 frames] 1) startled 2) body twisting like being sucked into a vortex 3) only the sobbing head left 4) only a pair of flip-flops left. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 24. `char_co_ba_ganh_teleport_in` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_co_ba_ganh_teleport_in.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/co_ba_ganh/frames (teleport_in ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a hyperactive young fruit seller with one thick unibrow, a huge gap-toothed grin, two stiff braids sticking straight out sideways, a crooked conical hat, balancing a bamboo shoulder pole with two overflowing fruit baskets, patched brown blouse, mismatched plastic slippers, color palette #BC6C25 #DDA15E #606C38 #FEFAE0 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [teleport_in, 4 frames] 1) spat out in mid-air 2) flailing 3) landing staggering 4) spiral eyes dizzy. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 25. `char_su_phu_mi_levelup` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_su_phu_mi_levelup.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: characters/su_phu_mi/frames (levelup ×6)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of an eccentric old kung fu master whose long beard and eyebrows are made of curly instant noodles, cross-eyed, bright red bulbous nose, one eye squinting, stained orange robe with soup splashes, holding a pair of chopsticks like a weapon, color palette #FB8500 #FFB703 #8ECAE6 #023047 with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [levelup, 6 frames] 1) ready stance 2) golden aura bursting 3) superhero pose with aura 4) bigger aura heroic pose 5) thumbs up 6) back to ready stance. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 26. `char_su_phu_mi_realm_drop` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_su_phu_mi_realm_drop.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: characters/su_phu_mi/frames (realm_drop ×6)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of an eccentric old kung fu master whose long beard and eyebrows are made of curly instant noodles, cross-eyed, bright red bulbous nose, one eye squinting, stained orange robe with soup splashes, holding a pair of chopsticks like a weapon, color palette #FB8500 #FFB703 #8ECAE6 #023047 with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [realm_drop, 6 frames] 1) stunned 2) knees buckling 3) sitting flat on the ground 4) crying rivers of tears 5) crying harder 6) wiping nose and standing up. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 27. `char_su_phu_mi_breakthrough` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_su_phu_mi_breakthrough.png`
- **Bố cục**: sprite sheet, hàng 1: 5 hình · hàng 2: 5 hình, nền magenta #FF00FF
- **Ra file game**: characters/su_phu_mi/frames (breakthrough ×10)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of an eccentric old kung fu master whose long beard and eyebrows are made of curly instant noodles, cross-eyed, bright red bulbous nose, one eye squinting, stained orange robe with soup splashes, holding a pair of chopsticks like a weapon, color palette #FB8500 #FFB703 #8ECAE6 #023047 with 10 frames in total. Exactly 5 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [breakthrough, 10 frames] 1) meditating seated 2) trembling 3) trembling harder 4) hair standing straight up 5) looking up screaming at the sky 6) screaming harder 7) struck by lightning 8) lightning engulfs body 9) singed frizzy smoking hair 10) triumphant laugh. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 28. `char_su_phu_mi_respawn` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_su_phu_mi_respawn.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: characters/su_phu_mi/frames (respawn ×6)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of an eccentric old kung fu master whose long beard and eyebrows are made of curly instant noodles, cross-eyed, bright red bulbous nose, one eye squinting, stained orange robe with soup splashes, holding a pair of chopsticks like a weapon, color palette #FB8500 #FFB703 #8ECAE6 #023047 with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [respawn, 6 frames] 1) falling from high above 2) about to hit the ground 3) landing face first 4) somersault 5) standing up dusting off 6) striking a pose. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 29. `char_su_phu_mi_teleport_out` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_su_phu_mi_teleport_out.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/su_phu_mi/frames (teleport_out ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of an eccentric old kung fu master whose long beard and eyebrows are made of curly instant noodles, cross-eyed, bright red bulbous nose, one eye squinting, stained orange robe with soup splashes, holding a pair of chopsticks like a weapon, color palette #FB8500 #FFB703 #8ECAE6 #023047 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [teleport_out, 4 frames] 1) startled 2) body twisting like being sucked into a vortex 3) only the sobbing head left 4) only a pair of flip-flops left. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 30. `char_su_phu_mi_teleport_in` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_su_phu_mi_teleport_in.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/su_phu_mi/frames (teleport_in ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of an eccentric old kung fu master whose long beard and eyebrows are made of curly instant noodles, cross-eyed, bright red bulbous nose, one eye squinting, stained orange robe with soup splashes, holding a pair of chopsticks like a weapon, color palette #FB8500 #FFB703 #8ECAE6 #023047 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [teleport_in, 4 frames] 1) spat out in mid-air 2) flailing 3) landing staggering 4) spiral eyes dizzy. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 31. `char_be_ti_levelup` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_be_ti_levelup.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: characters/be_ti/frames (levelup ×6)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing, color palette #E76F51 #F4A261 #2A9D8F #1D3557 with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [levelup, 6 frames] 1) ready stance 2) golden aura bursting 3) superhero pose with aura 4) bigger aura heroic pose 5) thumbs up 6) back to ready stance. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 32. `char_be_ti_realm_drop` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_be_ti_realm_drop.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: characters/be_ti/frames (realm_drop ×6)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing, color palette #E76F51 #F4A261 #2A9D8F #1D3557 with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [realm_drop, 6 frames] 1) stunned 2) knees buckling 3) sitting flat on the ground 4) crying rivers of tears 5) crying harder 6) wiping nose and standing up. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 33. `char_be_ti_breakthrough` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_be_ti_breakthrough.png`
- **Bố cục**: sprite sheet, hàng 1: 5 hình · hàng 2: 5 hình, nền magenta #FF00FF
- **Ra file game**: characters/be_ti/frames (breakthrough ×10)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing, color palette #E76F51 #F4A261 #2A9D8F #1D3557 with 10 frames in total. Exactly 5 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [breakthrough, 10 frames] 1) meditating seated 2) trembling 3) trembling harder 4) hair standing straight up 5) looking up screaming at the sky 6) screaming harder 7) struck by lightning 8) lightning engulfs body 9) singed frizzy smoking hair 10) triumphant laugh. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 34. `char_be_ti_respawn` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_be_ti_respawn.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: characters/be_ti/frames (respawn ×6)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing, color palette #E76F51 #F4A261 #2A9D8F #1D3557 with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [respawn, 6 frames] 1) falling from high above 2) about to hit the ground 3) landing face first 4) somersault 5) standing up dusting off 6) striking a pose. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 35. `char_be_ti_teleport_out` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_be_ti_teleport_out.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/be_ti/frames (teleport_out ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing, color palette #E76F51 #F4A261 #2A9D8F #1D3557 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [teleport_out, 4 frames] 1) startled 2) body twisting like being sucked into a vortex 3) only the sobbing head left 4) only a pair of flip-flops left. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 36. `char_be_ti_teleport_in` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_be_ti_teleport_in.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/be_ti/frames (teleport_in ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing, color palette #E76F51 #F4A261 #2A9D8F #1D3557 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [teleport_in, 4 frames] 1) spat out in mid-air 2) flailing 3) landing staggering 4) spiral eyes dizzy. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 37. `char_shipper_levelup` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_shipper_levelup.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: characters/shipper/frames (levelup ×6)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand, color palette #FFBE0B #FB5607 #3A86FF #222222 with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [levelup, 6 frames] 1) ready stance 2) golden aura bursting 3) superhero pose with aura 4) bigger aura heroic pose 5) thumbs up 6) back to ready stance. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 38. `char_shipper_realm_drop` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_shipper_realm_drop.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: characters/shipper/frames (realm_drop ×6)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand, color palette #FFBE0B #FB5607 #3A86FF #222222 with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [realm_drop, 6 frames] 1) stunned 2) knees buckling 3) sitting flat on the ground 4) crying rivers of tears 5) crying harder 6) wiping nose and standing up. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 39. `char_shipper_breakthrough` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_shipper_breakthrough.png`
- **Bố cục**: sprite sheet, hàng 1: 5 hình · hàng 2: 5 hình, nền magenta #FF00FF
- **Ra file game**: characters/shipper/frames (breakthrough ×10)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand, color palette #FFBE0B #FB5607 #3A86FF #222222 with 10 frames in total. Exactly 5 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [breakthrough, 10 frames] 1) meditating seated 2) trembling 3) trembling harder 4) hair standing straight up 5) looking up screaming at the sky 6) screaming harder 7) struck by lightning 8) lightning engulfs body 9) singed frizzy smoking hair 10) triumphant laugh. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 40. `char_shipper_respawn` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_shipper_respawn.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: characters/shipper/frames (respawn ×6)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand, color palette #FFBE0B #FB5607 #3A86FF #222222 with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [respawn, 6 frames] 1) falling from high above 2) about to hit the ground 3) landing face first 4) somersault 5) standing up dusting off 6) striking a pose. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 41. `char_shipper_teleport_out` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_shipper_teleport_out.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/shipper/frames (teleport_out ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand, color palette #FFBE0B #FB5607 #3A86FF #222222 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [teleport_out, 4 frames] 1) startled 2) body twisting like being sucked into a vortex 3) only the sobbing head left 4) only a pair of flip-flops left. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 42. `char_shipper_teleport_in` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_shipper_teleport_in.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/shipper/frames (teleport_in ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand, color palette #FFBE0B #FB5607 #3A86FF #222222 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [teleport_in, 4 frames] 1) spat out in mid-air 2) flailing 3) landing staggering 4) spiral eyes dizzy. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 43. `char_cu_rua_levelup` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_cu_rua_levelup.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/frames (levelup ×6)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, color palette #52B788 #D8F3DC #6F4E37 #F1FAEE with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [levelup, 6 frames] 1) ready stance 2) golden aura bursting 3) superhero pose with aura 4) bigger aura heroic pose 5) thumbs up 6) back to ready stance. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 44. `char_cu_rua_realm_drop` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_cu_rua_realm_drop.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/frames (realm_drop ×6)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, color palette #52B788 #D8F3DC #6F4E37 #F1FAEE with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [realm_drop, 6 frames] 1) stunned 2) knees buckling 3) sitting flat on the ground 4) crying rivers of tears 5) crying harder 6) wiping nose and standing up. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 45. `char_cu_rua_breakthrough` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_cu_rua_breakthrough.png`
- **Bố cục**: sprite sheet, hàng 1: 5 hình · hàng 2: 5 hình, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/frames (breakthrough ×10)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, color palette #52B788 #D8F3DC #6F4E37 #F1FAEE with 10 frames in total. Exactly 5 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [breakthrough, 10 frames] 1) meditating seated 2) trembling 3) trembling harder 4) hair standing straight up 5) looking up screaming at the sky 6) screaming harder 7) struck by lightning 8) lightning engulfs body 9) singed frizzy smoking hair 10) triumphant laugh. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 46. `char_cu_rua_respawn` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_cu_rua_respawn.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/frames (respawn ×6)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, color palette #52B788 #D8F3DC #6F4E37 #F1FAEE with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [respawn, 6 frames] 1) falling from high above 2) about to hit the ground 3) landing face first 4) somersault 5) standing up dusting off 6) striking a pose. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 47. `char_cu_rua_teleport_out` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_cu_rua_teleport_out.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/frames (teleport_out ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, color palette #52B788 #D8F3DC #6F4E37 #F1FAEE with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [teleport_out, 4 frames] 1) startled 2) body twisting like being sucked into a vortex 3) only the sobbing head left 4) only a pair of flip-flops left. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 48. `char_cu_rua_teleport_in` · ưu tiên 4

- **Lưu thành**: `assets/_raw/char_cu_rua_teleport_in.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/frames (teleport_in ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, color palette #52B788 #D8F3DC #6F4E37 #F1FAEE with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [teleport_in, 4 frames] 1) spat out in mid-air 2) flailing 3) landing staggering 4) spiral eyes dizzy. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 49. `obs_fence_break` · ưu tiên 4

- **Lưu thành**: `assets/_raw/obs_fence_break.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: obstacles/fence/break (break ×6)

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Sprite sheet of a top-down 3/4 view game obstacle, a straight horizontal bamboo fence segment tied with rope with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [break, 6 frames] 1) shaking 2) cracking 3) bursting apart 4) debris flying 5) dust cloud 6) rubble pile. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 50. `obs_fence_regrow` · ưu tiên 4

- **Lưu thành**: `assets/_raw/obs_fence_regrow.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: obstacles/fence/regrow (regrow ×6)

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Sprite sheet of a top-down 3/4 view game obstacle, a straight horizontal bamboo fence segment tied with rope with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [regrow, 6 frames] 1) glowing rubble 2) pieces rising with light 3) half rebuilt glowing 4) almost whole 5) whole with sparkles 6) intact and solid. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 51. `obs_stall_break` · ưu tiên 4

- **Lưu thành**: `assets/_raw/obs_stall_break.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: obstacles/stall/break (break ×6)

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Sprite sheet of a top-down 3/4 view game obstacle, a street market stall with a striped canopy and fruit baskets with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [break, 6 frames] 1) shaking 2) cracking 3) bursting apart 4) debris flying 5) dust cloud 6) rubble pile. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 52. `obs_stall_regrow` · ưu tiên 4

- **Lưu thành**: `assets/_raw/obs_stall_regrow.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: obstacles/stall/regrow (regrow ×6)

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Sprite sheet of a top-down 3/4 view game obstacle, a street market stall with a striped canopy and fruit baskets with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [regrow, 6 frames] 1) glowing rubble 2) pieces rising with light 3) half rebuilt glowing 4) almost whole 5) whole with sparkles 6) intact and solid. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 53. `obs_cart_break` · ưu tiên 4

- **Lưu thành**: `assets/_raw/obs_cart_break.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: obstacles/cart/break (break ×6)

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Sprite sheet of a top-down 3/4 view game obstacle, a wooden street food cart with a small stove with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [break, 6 frames] 1) shaking 2) cracking 3) bursting apart 4) debris flying 5) dust cloud 6) rubble pile. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 54. `obs_cart_regrow` · ưu tiên 4

- **Lưu thành**: `assets/_raw/obs_cart_regrow.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: obstacles/cart/regrow (regrow ×6)

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Sprite sheet of a top-down 3/4 view game obstacle, a wooden street food cart with a small stove with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [regrow, 6 frames] 1) glowing rubble 2) pieces rising with light 3) half rebuilt glowing 4) almost whole 5) whole with sparkles 6) intact and solid. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 55. `obs_brick_break` · ưu tiên 4

- **Lưu thành**: `assets/_raw/obs_brick_break.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: obstacles/brick/break (break ×6)

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Sprite sheet of a top-down 3/4 view game obstacle, a short horizontal red brick wall segment with moss with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [break, 6 frames] 1) shaking 2) cracking 3) bursting apart 4) debris flying 5) dust cloud 6) rubble pile. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 56. `obs_brick_regrow` · ưu tiên 4

- **Lưu thành**: `assets/_raw/obs_brick_regrow.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: obstacles/brick/regrow (regrow ×6)

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Sprite sheet of a top-down 3/4 view game obstacle, a short horizontal red brick wall segment with moss with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [regrow, 6 frames] 1) glowing rubble 2) pieces rising with light 3) half rebuilt glowing 4) almost whole 5) whole with sparkles 6) intact and solid. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 57. `obs_boulder_break` · ưu tiên 4

- **Lưu thành**: `assets/_raw/obs_boulder_break.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: obstacles/boulder/break (break ×6)

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Sprite sheet of a top-down 3/4 view game obstacle, a huge round grey boulder with cartoon crack lines with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [break, 6 frames] 1) shaking 2) cracking 3) bursting apart 4) debris flying 5) dust cloud 6) rubble pile. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 58. `obs_boulder_regrow` · ưu tiên 4

- **Lưu thành**: `assets/_raw/obs_boulder_regrow.png`
- **Bố cục**: sprite sheet, hàng 1: 3 hình · hàng 2: 3 hình, nền magenta #FF00FF
- **Ra file game**: obstacles/boulder/regrow (regrow ×6)

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Sprite sheet of a top-down 3/4 view game obstacle, a huge round grey boulder with cartoon crack lines with 6 frames in total. Exactly 3 columns and 2 rows, equal cell size. Frames are read left-to-right, top-to-bottom: [regrow, 6 frames] 1) glowing rubble 2) pieces rising with light 3) half rebuilt glowing 4) almost whole 5) whole with sparkles 6) intact and solid. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 59. `vfx_run_dust` · ưu tiên 4

- **Lưu thành**: `assets/_raw/vfx_run_dust.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: vfx/run_dust (run_dust ×4)

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Sprite sheet of a game visual effect with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [run_dust, 4 frames] small brown dust puffs behind running feet. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 60. `decal_mud_edge` · ưu tiên 4

- **Lưu thành**: `assets/_raw/decal_mud_edge.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: decals/mud_edge.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. top-down zone border strip, muddy edge with dry cracks, tileable left to right, soft muted colors Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 61. `decal_pond_shore` · ưu tiên 4

- **Lưu thành**: `assets/_raw/decal_pond_shore.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: decals/pond_shore.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. top-down zone border strip, pond shore with reeds and pebbles, tileable left to right, soft muted colors Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 62. `decal_stone_path_edge` · ưu tiên 4

- **Lưu thành**: `assets/_raw/decal_stone_path_edge.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: decals/stone_path_edge.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. top-down zone border strip, stone path edge with broken gravel, tileable left to right, soft muted colors Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 63. `decal_bush_edge` · ưu tiên 4

- **Lưu thành**: `assets/_raw/decal_bush_edge.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: decals/bush_edge.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. top-down zone border strip, bush cluster edge with leaves, tileable left to right, soft muted colors Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 64. `portrait_a_hon_neutral` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_a_hon_neutral.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/a_hon/portrait_neutral.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a loud-mouthed scrawny street vendor with a giant steamed bun as a hat that has a big bite missing, two huge yellowish buck teeth with one chipped, a mole on his cheek with one long curly hair, sweat-stained yellowed sleeveless undershirt riding up over a round hairy belly button, baggy red shorts, mismatched flip-flops, head and shoulders portrait, facing the viewer slightly turned right, centered, neutral smug face Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 65. `portrait_a_hon_evil` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_a_hon_evil.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/a_hon/portrait_evil.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a loud-mouthed scrawny street vendor with a giant steamed bun as a hat that has a big bite missing, two huge yellowish buck teeth with one chipped, a mole on his cheek with one long curly hair, sweat-stained yellowed sleeveless undershirt riding up over a round hairy belly button, baggy red shorts, mismatched flip-flops, head and shoulders portrait, facing the viewer slightly turned right, centered, evil grin, chasing Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 66. `portrait_a_hon_panic` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_a_hon_panic.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/a_hon/portrait_panic.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a loud-mouthed scrawny street vendor with a giant steamed bun as a hat that has a big bite missing, two huge yellowish buck teeth with one chipped, a mole on his cheek with one long curly hair, sweat-stained yellowed sleeveless undershirt riding up over a round hairy belly button, baggy red shorts, mismatched flip-flops, head and shoulders portrait, facing the viewer slightly turned right, centered, panicked, sweating Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 67. `portrait_a_hon_cry` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_a_hon_cry.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/a_hon/portrait_cry.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a loud-mouthed scrawny street vendor with a giant steamed bun as a hat that has a big bite missing, two huge yellowish buck teeth with one chipped, a mole on his cheek with one long curly hair, sweat-stained yellowed sleeveless undershirt riding up over a round hairy belly button, baggy red shorts, mismatched flip-flops, head and shoulders portrait, facing the viewer slightly turned right, centered, crying with streams of tears Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 68. `portrait_ba_tam_loa_neutral` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_ba_tam_loa_neutral.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/ba_tam_loa/portrait_neutral.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a stout loud fish market lady holding a dented megaphone, bright red lipstick smeared far outside her lips and stuck on her teeth, two perfectly round blotchy pink blush circles, eyebrows drawn way too high in pencil, a tiny pencil stuck in her hair bun, floral blouse with fish scales stuck on it, wet rubber apron, tall green rubber boots, head and shoulders portrait, facing the viewer slightly turned right, centered, neutral smug face Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 69. `portrait_ba_tam_loa_evil` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_ba_tam_loa_evil.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/ba_tam_loa/portrait_evil.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a stout loud fish market lady holding a dented megaphone, bright red lipstick smeared far outside her lips and stuck on her teeth, two perfectly round blotchy pink blush circles, eyebrows drawn way too high in pencil, a tiny pencil stuck in her hair bun, floral blouse with fish scales stuck on it, wet rubber apron, tall green rubber boots, head and shoulders portrait, facing the viewer slightly turned right, centered, evil grin, chasing Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 70. `portrait_ba_tam_loa_panic` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_ba_tam_loa_panic.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/ba_tam_loa/portrait_panic.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a stout loud fish market lady holding a dented megaphone, bright red lipstick smeared far outside her lips and stuck on her teeth, two perfectly round blotchy pink blush circles, eyebrows drawn way too high in pencil, a tiny pencil stuck in her hair bun, floral blouse with fish scales stuck on it, wet rubber apron, tall green rubber boots, head and shoulders portrait, facing the viewer slightly turned right, centered, panicked, sweating Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 71. `portrait_ba_tam_loa_cry` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_ba_tam_loa_cry.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/ba_tam_loa/portrait_cry.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a stout loud fish market lady holding a dented megaphone, bright red lipstick smeared far outside her lips and stuck on her teeth, two perfectly round blotchy pink blush circles, eyebrows drawn way too high in pencil, a tiny pencil stuck in her hair bun, floral blouse with fish scales stuck on it, wet rubber apron, tall green rubber boots, head and shoulders portrait, facing the viewer slightly turned right, centered, crying with streams of tears Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 72. `portrait_ong_ba_vong_neutral` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_ong_ba_vong_neutral.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/ong_ba_vong/portrait_neutral.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a skinny lazy old man with only two teeth left, nose hairs so long they look like a mustache, three strands of hair combed over a shiny bald head, half-closed sleepy eyes, toothpick in mouth, folded hammock carried on his back like a cape, saggy white tank top, striped pajama pants slipping down showing flowery underwear waistband, head and shoulders portrait, facing the viewer slightly turned right, centered, neutral smug face Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 73. `portrait_ong_ba_vong_evil` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_ong_ba_vong_evil.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/ong_ba_vong/portrait_evil.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a skinny lazy old man with only two teeth left, nose hairs so long they look like a mustache, three strands of hair combed over a shiny bald head, half-closed sleepy eyes, toothpick in mouth, folded hammock carried on his back like a cape, saggy white tank top, striped pajama pants slipping down showing flowery underwear waistband, head and shoulders portrait, facing the viewer slightly turned right, centered, evil grin, chasing Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 74. `portrait_ong_ba_vong_panic` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_ong_ba_vong_panic.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/ong_ba_vong/portrait_panic.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a skinny lazy old man with only two teeth left, nose hairs so long they look like a mustache, three strands of hair combed over a shiny bald head, half-closed sleepy eyes, toothpick in mouth, folded hammock carried on his back like a cape, saggy white tank top, striped pajama pants slipping down showing flowery underwear waistband, head and shoulders portrait, facing the viewer slightly turned right, centered, panicked, sweating Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 75. `portrait_ong_ba_vong_cry` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_ong_ba_vong_cry.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/ong_ba_vong/portrait_cry.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a skinny lazy old man with only two teeth left, nose hairs so long they look like a mustache, three strands of hair combed over a shiny bald head, half-closed sleepy eyes, toothpick in mouth, folded hammock carried on his back like a cape, saggy white tank top, striped pajama pants slipping down showing flowery underwear waistband, head and shoulders portrait, facing the viewer slightly turned right, centered, crying with streams of tears Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 76. `portrait_dai_ca_dep_neutral` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_dai_ca_dep_neutral.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/dai_ca_dep/portrait_neutral.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a pompous short gang boss in an oversized shiny purple suit with shoulder pads, tiny round black sunglasses, a gold chain as thick as a dog leash, a thin pencil mustache, slicked-back hair so greasy that flies buzz around it, a gold front tooth, cheap honeycomb plastic sandals with socks, head and shoulders portrait, facing the viewer slightly turned right, centered, neutral smug face Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 77. `portrait_dai_ca_dep_evil` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_dai_ca_dep_evil.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/dai_ca_dep/portrait_evil.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a pompous short gang boss in an oversized shiny purple suit with shoulder pads, tiny round black sunglasses, a gold chain as thick as a dog leash, a thin pencil mustache, slicked-back hair so greasy that flies buzz around it, a gold front tooth, cheap honeycomb plastic sandals with socks, head and shoulders portrait, facing the viewer slightly turned right, centered, evil grin, chasing Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 78. `portrait_dai_ca_dep_panic` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_dai_ca_dep_panic.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/dai_ca_dep/portrait_panic.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a pompous short gang boss in an oversized shiny purple suit with shoulder pads, tiny round black sunglasses, a gold chain as thick as a dog leash, a thin pencil mustache, slicked-back hair so greasy that flies buzz around it, a gold front tooth, cheap honeycomb plastic sandals with socks, head and shoulders portrait, facing the viewer slightly turned right, centered, panicked, sweating Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 79. `portrait_dai_ca_dep_cry` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_dai_ca_dep_cry.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/dai_ca_dep/portrait_cry.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a pompous short gang boss in an oversized shiny purple suit with shoulder pads, tiny round black sunglasses, a gold chain as thick as a dog leash, a thin pencil mustache, slicked-back hair so greasy that flies buzz around it, a gold front tooth, cheap honeycomb plastic sandals with socks, head and shoulders portrait, facing the viewer slightly turned right, centered, crying with streams of tears Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 80. `portrait_map_tra_sua_neutral` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_map_tra_sua_neutral.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/map_tra_sua/portrait_neutral.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a round chubby teen with huge jiggly cheeks and a double chin, tapioca pearls stuck around his mouth, clutching a giant bubble tea cup with a fat straw, oversized hoodie with sweat-soaked armpits, tiny shorts, untied sneakers, head and shoulders portrait, facing the viewer slightly turned right, centered, neutral smug face Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 81. `portrait_map_tra_sua_evil` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_map_tra_sua_evil.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/map_tra_sua/portrait_evil.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a round chubby teen with huge jiggly cheeks and a double chin, tapioca pearls stuck around his mouth, clutching a giant bubble tea cup with a fat straw, oversized hoodie with sweat-soaked armpits, tiny shorts, untied sneakers, head and shoulders portrait, facing the viewer slightly turned right, centered, evil grin, chasing Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 82. `portrait_map_tra_sua_panic` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_map_tra_sua_panic.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/map_tra_sua/portrait_panic.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a round chubby teen with huge jiggly cheeks and a double chin, tapioca pearls stuck around his mouth, clutching a giant bubble tea cup with a fat straw, oversized hoodie with sweat-soaked armpits, tiny shorts, untied sneakers, head and shoulders portrait, facing the viewer slightly turned right, centered, panicked, sweating Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 83. `portrait_map_tra_sua_cry` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_map_tra_sua_cry.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/map_tra_sua/portrait_cry.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a round chubby teen with huge jiggly cheeks and a double chin, tapioca pearls stuck around his mouth, clutching a giant bubble tea cup with a fat straw, oversized hoodie with sweat-soaked armpits, tiny shorts, untied sneakers, head and shoulders portrait, facing the viewer slightly turned right, centered, crying with streams of tears Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 84. `portrait_co_ba_ganh_neutral` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_co_ba_ganh_neutral.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/co_ba_ganh/portrait_neutral.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a hyperactive young fruit seller with one thick unibrow, a huge gap-toothed grin, two stiff braids sticking straight out sideways, a crooked conical hat, balancing a bamboo shoulder pole with two overflowing fruit baskets, patched brown blouse, mismatched plastic slippers, head and shoulders portrait, facing the viewer slightly turned right, centered, neutral smug face Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 85. `portrait_co_ba_ganh_evil` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_co_ba_ganh_evil.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/co_ba_ganh/portrait_evil.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a hyperactive young fruit seller with one thick unibrow, a huge gap-toothed grin, two stiff braids sticking straight out sideways, a crooked conical hat, balancing a bamboo shoulder pole with two overflowing fruit baskets, patched brown blouse, mismatched plastic slippers, head and shoulders portrait, facing the viewer slightly turned right, centered, evil grin, chasing Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 86. `portrait_co_ba_ganh_panic` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_co_ba_ganh_panic.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/co_ba_ganh/portrait_panic.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a hyperactive young fruit seller with one thick unibrow, a huge gap-toothed grin, two stiff braids sticking straight out sideways, a crooked conical hat, balancing a bamboo shoulder pole with two overflowing fruit baskets, patched brown blouse, mismatched plastic slippers, head and shoulders portrait, facing the viewer slightly turned right, centered, panicked, sweating Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 87. `portrait_co_ba_ganh_cry` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_co_ba_ganh_cry.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/co_ba_ganh/portrait_cry.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a hyperactive young fruit seller with one thick unibrow, a huge gap-toothed grin, two stiff braids sticking straight out sideways, a crooked conical hat, balancing a bamboo shoulder pole with two overflowing fruit baskets, patched brown blouse, mismatched plastic slippers, head and shoulders portrait, facing the viewer slightly turned right, centered, crying with streams of tears Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 88. `portrait_su_phu_mi_neutral` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_su_phu_mi_neutral.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/su_phu_mi/portrait_neutral.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. an eccentric old kung fu master whose long beard and eyebrows are made of curly instant noodles, cross-eyed, bright red bulbous nose, one eye squinting, stained orange robe with soup splashes, holding a pair of chopsticks like a weapon, head and shoulders portrait, facing the viewer slightly turned right, centered, neutral smug face Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 89. `portrait_su_phu_mi_evil` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_su_phu_mi_evil.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/su_phu_mi/portrait_evil.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. an eccentric old kung fu master whose long beard and eyebrows are made of curly instant noodles, cross-eyed, bright red bulbous nose, one eye squinting, stained orange robe with soup splashes, holding a pair of chopsticks like a weapon, head and shoulders portrait, facing the viewer slightly turned right, centered, evil grin, chasing Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 90. `portrait_su_phu_mi_panic` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_su_phu_mi_panic.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/su_phu_mi/portrait_panic.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. an eccentric old kung fu master whose long beard and eyebrows are made of curly instant noodles, cross-eyed, bright red bulbous nose, one eye squinting, stained orange robe with soup splashes, holding a pair of chopsticks like a weapon, head and shoulders portrait, facing the viewer slightly turned right, centered, panicked, sweating Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 91. `portrait_su_phu_mi_cry` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_su_phu_mi_cry.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/su_phu_mi/portrait_cry.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. an eccentric old kung fu master whose long beard and eyebrows are made of curly instant noodles, cross-eyed, bright red bulbous nose, one eye squinting, stained orange robe with soup splashes, holding a pair of chopsticks like a weapon, head and shoulders portrait, facing the viewer slightly turned right, centered, crying with streams of tears Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 92. `portrait_be_ti_neutral` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_be_ti_neutral.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/be_ti/portrait_neutral.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing, head and shoulders portrait, facing the viewer slightly turned right, centered, neutral smug face Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 93. `portrait_be_ti_evil` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_be_ti_evil.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/be_ti/portrait_evil.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing, head and shoulders portrait, facing the viewer slightly turned right, centered, evil grin, chasing Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 94. `portrait_be_ti_panic` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_be_ti_panic.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/be_ti/portrait_panic.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing, head and shoulders portrait, facing the viewer slightly turned right, centered, panicked, sweating Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 95. `portrait_be_ti_cry` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_be_ti_cry.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/be_ti/portrait_cry.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing, head and shoulders portrait, facing the viewer slightly turned right, centered, crying with streams of tears Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 96. `portrait_shipper_neutral` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_shipper_neutral.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/shipper/portrait_neutral.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand, head and shoulders portrait, facing the viewer slightly turned right, centered, neutral smug face Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 97. `portrait_shipper_evil` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_shipper_evil.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/shipper/portrait_evil.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand, head and shoulders portrait, facing the viewer slightly turned right, centered, evil grin, chasing Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 98. `portrait_shipper_panic` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_shipper_panic.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/shipper/portrait_panic.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand, head and shoulders portrait, facing the viewer slightly turned right, centered, panicked, sweating Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 99. `portrait_shipper_cry` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_shipper_cry.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/shipper/portrait_cry.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand, head and shoulders portrait, facing the viewer slightly turned right, centered, crying with streams of tears Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 100. `portrait_cu_rua_neutral` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_cu_rua_neutral.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/portrait_neutral.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, head and shoulders portrait, facing the viewer slightly turned right, centered, neutral smug face Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 101. `portrait_cu_rua_evil` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_cu_rua_evil.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/portrait_evil.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, head and shoulders portrait, facing the viewer slightly turned right, centered, evil grin, chasing Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 102. `portrait_cu_rua_panic` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_cu_rua_panic.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/portrait_panic.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, head and shoulders portrait, facing the viewer slightly turned right, centered, panicked, sweating Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 103. `portrait_cu_rua_cry` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_cu_rua_cry.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/portrait_cry.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, head and shoulders portrait, facing the viewer slightly turned right, centered, crying with streams of tears Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 104. `char_a_hon_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_a_hon_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/a_hon/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a loud-mouthed scrawny street vendor with a giant steamed bun as a hat that has a big bite missing, two huge yellowish buck teeth with one chipped, a mole on his cheek with one long curly hair, sweat-stained yellowed sleeveless undershirt riding up over a round hairy belly button, baggy red shorts, mismatched flip-flops, color palette #FFFFFF #E63946 #F4A261 #2B2D42. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 105. `char_ba_tam_loa_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_ba_tam_loa_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/ba_tam_loa/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a stout loud fish market lady holding a dented megaphone, bright red lipstick smeared far outside her lips and stuck on her teeth, two perfectly round blotchy pink blush circles, eyebrows drawn way too high in pencil, a tiny pencil stuck in her hair bun, floral blouse with fish scales stuck on it, wet rubber apron, tall green rubber boots, color palette #2A9D8F #E9C46A #F28482 #264653. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 106. `char_ong_ba_vong_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_ong_ba_vong_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/ong_ba_vong/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a skinny lazy old man with only two teeth left, nose hairs so long they look like a mustache, three strands of hair combed over a shiny bald head, half-closed sleepy eyes, toothpick in mouth, folded hammock carried on his back like a cape, saggy white tank top, striped pajama pants slipping down showing flowery underwear waistband, color palette #8D99AE #EDF2F4 #6D597A #B56576. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 107. `char_dai_ca_dep_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_dai_ca_dep_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/dai_ca_dep/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a pompous short gang boss in an oversized shiny purple suit with shoulder pads, tiny round black sunglasses, a gold chain as thick as a dog leash, a thin pencil mustache, slicked-back hair so greasy that flies buzz around it, a gold front tooth, cheap honeycomb plastic sandals with socks, color palette #7209B7 #FFD60A #000000 #F72585. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 108. `char_map_tra_sua_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_map_tra_sua_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/map_tra_sua/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a round chubby teen with huge jiggly cheeks and a double chin, tapioca pearls stuck around his mouth, clutching a giant bubble tea cup with a fat straw, oversized hoodie with sweat-soaked armpits, tiny shorts, untied sneakers, color palette #CDB4DB #A47148 #FFC8DD #3D405B. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 109. `char_co_ba_ganh_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_co_ba_ganh_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/co_ba_ganh/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a hyperactive young fruit seller with one thick unibrow, a huge gap-toothed grin, two stiff braids sticking straight out sideways, a crooked conical hat, balancing a bamboo shoulder pole with two overflowing fruit baskets, patched brown blouse, mismatched plastic slippers, color palette #BC6C25 #DDA15E #606C38 #FEFAE0. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 110. `char_su_phu_mi_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_su_phu_mi_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/su_phu_mi/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of an eccentric old kung fu master whose long beard and eyebrows are made of curly instant noodles, cross-eyed, bright red bulbous nose, one eye squinting, stained orange robe with soup splashes, holding a pair of chopsticks like a weapon, color palette #FB8500 #FFB703 #8ECAE6 #023047. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 111. `char_be_ti_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_be_ti_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/be_ti/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing, color palette #E76F51 #F4A261 #2A9D8F #1D3557. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 112. `char_shipper_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_shipper_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/shipper/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand, color palette #FFBE0B #FB5607 #3A86FF #222222. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 113. `char_cu_rua_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_cu_rua_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, color palette #52B788 #D8F3DC #6F4E37 #F1FAEE. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

