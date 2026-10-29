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

## 10. `tile_grass` · ưu tiên 1

- **Sinh lại**: spec v3: nền nhạt, ít chi tiết (ART_PROMPTS §9.0)
- **Lưu thành**: `assets/_raw/tile_grass.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/grass.png

```text
STYLE: 2D mobile game art, cartoon, flat colors with simple cel shading, vibrant saturated palette, clean vector-like look, no text, no watermark. Seamless tileable top-down ground texture, soft muted sage green grass, very low contrast, subtle large brush strokes, no flowers, no small colorful dots, calm and plain, square image filling the whole frame, cartoon game style, no objects, no characters, even lighting, edges must tile perfectly. Desaturated, low detail, designed as a background that does not distract from gameplay objects.
```

## 11. `tile_concrete` · ưu tiên 1

- **Sinh lại**: spec v3: nền nhạt, ít chi tiết (ART_PROMPTS §9.0)
- **Lưu thành**: `assets/_raw/tile_concrete.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/concrete.png

```text
STYLE: 2D mobile game art, cartoon, flat colors with simple cel shading, vibrant saturated palette, clean vector-like look, no text, no watermark. Seamless tileable top-down ground texture, pale warm grey concrete, very low contrast, few faint cracks, plain and calm, square image filling the whole frame, cartoon game style, no objects, no characters, even lighting, edges must tile perfectly. Desaturated, low detail, designed as a background that does not distract from gameplay objects.
```

## 12. `tile_dirt` · ưu tiên 1

- **Sinh lại**: spec v3: nền nhạt, ít chi tiết (ART_PROMPTS §9.0)
- **Lưu thành**: `assets/_raw/tile_dirt.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/dirt.png

```text
STYLE: 2D mobile game art, cartoon, flat colors with simple cel shading, vibrant saturated palette, clean vector-like look, no text, no watermark. Seamless tileable top-down ground texture, light muted sandy brown dirt, very low contrast, only a few faint large pebbles, plain and calm, square image filling the whole frame, cartoon game style, no objects, no characters, even lighting, edges must tile perfectly. Desaturated, low detail, designed as a background that does not distract from gameplay objects.
```

## 13. `char_ba_tam_loa_hit` · ưu tiên 2

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/char_ba_tam_loa_hit.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/ba_tam_loa/frames (hit ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a stout loud fish market lady holding a dented megaphone, bright red lipstick smeared far outside her lips and stuck on her teeth, two perfectly round blotchy pink blush circles, eyebrows drawn way too high in pencil, a tiny pencil stuck in her hair bun, floral blouse with fish scales stuck on it, wet rubber apron, tall green rubber boots, color palette #2A9D8F #E9C46A #F28482 #264653 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [hit, 4 frames] 1) face squashed flat to the left 2) body knocked backwards 3) spiral eyes with stars around the head 4) shaking head recovering; comedy gag of this character in this animation: her lipstick mark gets printed on the air in front of her. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 14. `char_dai_ca_dep_hit` · ưu tiên 2

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/char_dai_ca_dep_hit.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/dai_ca_dep/frames (hit ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a pompous short gang boss in an oversized shiny purple suit with shoulder pads, tiny round black sunglasses, a gold chain as thick as a dog leash, a thin pencil mustache, slicked-back hair so greasy that flies buzz around it, a gold front tooth, cheap honeycomb plastic sandals with socks, color palette #7209B7 #FFD60A #000000 #F72585 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [hit, 4 frames] 1) face squashed flat to the left 2) body knocked backwards 3) spiral eyes with stars around the head 4) shaking head recovering; comedy gag of this character in this animation: his sunglasses fall off revealing cross-eyes. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 15. `tile_rocky_ground` · ưu tiên 2

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/tile_rocky_ground.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/rocky_ground.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. grey rocky ground, seamless tileable top-down, desaturated, low contrast, low detail, no objects
```

## 16. `tile_fine_gravel` · ưu tiên 2

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/tile_fine_gravel.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/fine_gravel.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. fine gravel, seamless tileable top-down, desaturated, low contrast, low detail, no objects
```

## 17. `tile_mossy_stone` · ưu tiên 2

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/tile_mossy_stone.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/mossy_stone.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. patchy moss on stone, seamless tileable top-down, desaturated, low contrast, low detail, no objects
```

## 18. `tile_soft_grass` · ưu tiên 2

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/tile_soft_grass.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/soft_grass.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. soft muted grass, seamless tileable top-down, desaturated, low contrast, low detail, no objects
```

## 19. `tile_bamboo_litter` · ưu tiên 2

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/tile_bamboo_litter.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/bamboo_litter.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. dry bamboo leaves litter on soil, seamless tileable top-down, desaturated, low contrast, low detail, no objects
```

## 20. `tile_damp_earth` · ưu tiên 2

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/tile_damp_earth.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/damp_earth.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. damp dark earth, seamless tileable top-down, desaturated, low contrast, low detail, no objects
```

## 21. `tile_dirt_path` · ưu tiên 2

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/tile_dirt_path.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/dirt_path.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. trampled dirt path ground, seamless tileable top-down, desaturated, low contrast, low detail, no objects
```

## 22. `tile_village_grass` · ưu tiên 2

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/tile_village_grass.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/village_grass.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. short village grass, seamless tileable top-down, desaturated, low contrast, low detail, no objects
```

## 23. `tile_dusty_earth` · ưu tiên 2

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/tile_dusty_earth.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/dusty_earth.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. dusty packed earth, seamless tileable top-down, desaturated, low contrast, low detail, no objects
```

## 24. `plate_shine_streak` · ưu tiên 3

- **Lưu thành**: `assets/_raw/plate_shine_streak.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: ui/shine_streak.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. soft diagonal white light streak Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 25. `zone_bamboo_wall` · ưu tiên 4

- **Lưu thành**: `assets/_raw/zone_bamboo_wall.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: obstacles/bamboo_wall.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. A very dense bamboo grove wall segment seen from above, tileable left to right Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 26. `zone_bamboo_corner` · ưu tiên 4

- **Lưu thành**: `assets/_raw/zone_bamboo_corner.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: obstacles/bamboo_corner.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. A very dense bamboo grove corner piece seen from above, L-shaped Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 27. `zone_village_roof` · ưu tiên 4

- **Lưu thành**: `assets/_raw/zone_village_roof.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: obstacles/village_roof.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. A small village house roof of grey tiles seen from directly above Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 28. `zone_lantern_string` · ưu tiên 4

- **Lưu thành**: `assets/_raw/zone_lantern_string.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: obstacles/lantern_string.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. A string of paper lanterns seen from above Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 29. `zone_ice_island_edge` · ưu tiên 4

- **Lưu thành**: `assets/_raw/zone_ice_island_edge.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: obstacles/ice_island_edge.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. A small ice island edge segment seen from above Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 30. `tile_mountain_rock` · ưu tiên 4

- **Lưu thành**: `assets/_raw/tile_mountain_rock.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/mountain_rock.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Seamless tileable top-down ground texture: grey mountain rock with patches of moss, flat even lighting, no shadows Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 31. `tile_bamboo_floor` · ưu tiên 4

- **Lưu thành**: `assets/_raw/tile_bamboo_floor.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/bamboo_floor.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Seamless tileable top-down ground texture: damp forest floor covered with bamboo leaves, flat even lighting, no shadows Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 32. `tile_cobblestone` · ưu tiên 4

- **Lưu thành**: `assets/_raw/tile_cobblestone.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/cobblestone.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Seamless tileable top-down ground texture: old cobblestone street, flat even lighting, no shadows Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 33. `tile_ice` · ưu tiên 4

- **Lưu thành**: `assets/_raw/tile_ice.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/ice.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Seamless tileable top-down ground texture: smooth glossy pale blue ice with subtle cracks and reflections, flat even lighting, no shadows Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 34. `tile_snow` · ưu tiên 4

- **Lưu thành**: `assets/_raw/tile_snow.png`
- **Bố cục**: texture vuông lặp liền mạch, phủ kín ảnh
- **Ra file game**: tiles/snow.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Seamless tileable top-down ground texture: soft deep snow ground, flat even lighting, no shadows Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 35. `char_map_tra_sua_teleport_out` · ưu tiên 4

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/char_map_tra_sua_teleport_out.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/map_tra_sua/frames (teleport_out ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a round chubby teen with huge jiggly cheeks and a double chin, tapioca pearls stuck around his mouth, clutching a giant bubble tea cup with a fat straw, oversized hoodie with sweat-soaked armpits, tiny shorts, untied sneakers, color palette #CDB4DB #A47148 #FFC8DD #3D405B with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [teleport_out, 4 frames] 1) startled 2) body twisting like being sucked into a vortex 3) only the sobbing head left 4) only a pair of flip-flops left. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 36. `char_be_ti_teleport_in` · ưu tiên 4

- **Sinh lại**: sheet bị cắt sai số khung (xem cảnh báo của `artgen prep`)
- **Lưu thành**: `assets/_raw/char_be_ti_teleport_in.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: characters/be_ti/frames (teleport_in ×4)

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Sprite sheet of a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing, color palette #E76F51 #F4A261 #2A9D8F #1D3557 with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [teleport_in, 4 frames] 1) spat out in mid-air 2) flailing 3) landing staggering 4) spiral eyes dizzy. Identical character in every frame: same outfit, colors, proportions, every frame facing right in 3/4 view, character the same height in every cell, feet on the same baseline. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 37. `vfx_run_dust` · ưu tiên 4

- **Lưu thành**: `assets/_raw/vfx_run_dust.png`
- **Bố cục**: sprite sheet, hàng 1: 4 hình, nền magenta #FF00FF
- **Ra file game**: vfx/run_dust (run_dust ×4)

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Sprite sheet of a game visual effect with 4 frames in total. Exactly 4 columns and 1 row, equal cell size. Frames are read left-to-right, top-to-bottom: [run_dust, 4 frames] small brown dust puffs behind running feet. Leave plenty of empty background space between frames and between rows so every frame is clearly separated; no grid lines, no frame numbers. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 38. `decal_ice_skid` · ưu tiên 5

- **Lưu thành**: `assets/_raw/decal_ice_skid.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: decals/ice_skid.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Top-down ground decal: a pair of curved ice skid marks Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 39. `decal_snow_pile` · ưu tiên 5

- **Lưu thành**: `assets/_raw/decal_snow_pile.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: decals/snow_pile.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Top-down ground decal: a small pile of snow Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 40. `decal_frozen_reeds` · ưu tiên 5

- **Lưu thành**: `assets/_raw/decal_frozen_reeds.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: decals/frozen_reeds.png

```text
STYLE: 2D game asset, lowbrow slapstick cartoon, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow. Do not use magenta or pink-purple in the subject itself. Top-down ground decal: a few frozen reeds poking through ice Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 41. `portrait_be_ti_evil` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_be_ti_evil.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/be_ti/portrait_evil.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing, head and shoulders portrait, facing the viewer slightly turned right, centered, evil grin, chasing Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 42. `portrait_be_ti_panic` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_be_ti_panic.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/be_ti/portrait_panic.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing, head and shoulders portrait, facing the viewer slightly turned right, centered, panicked, sweating Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 43. `portrait_be_ti_cry` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_be_ti_cry.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/be_ti/portrait_cry.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing, head and shoulders portrait, facing the viewer slightly turned right, centered, crying with streams of tears Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 44. `portrait_shipper_neutral` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_shipper_neutral.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/shipper/portrait_neutral.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand, head and shoulders portrait, facing the viewer slightly turned right, centered, neutral smug face Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 45. `portrait_shipper_evil` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_shipper_evil.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/shipper/portrait_evil.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand, head and shoulders portrait, facing the viewer slightly turned right, centered, evil grin, chasing Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 46. `portrait_shipper_panic` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_shipper_panic.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/shipper/portrait_panic.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand, head and shoulders portrait, facing the viewer slightly turned right, centered, panicked, sweating Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 47. `portrait_shipper_cry` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_shipper_cry.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/shipper/portrait_cry.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand, head and shoulders portrait, facing the viewer slightly turned right, centered, crying with streams of tears Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 48. `portrait_cu_rua_neutral` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_cu_rua_neutral.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/portrait_neutral.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, head and shoulders portrait, facing the viewer slightly turned right, centered, neutral smug face Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 49. `portrait_cu_rua_evil` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_cu_rua_evil.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/portrait_evil.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, head and shoulders portrait, facing the viewer slightly turned right, centered, evil grin, chasing Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 50. `portrait_cu_rua_panic` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_cu_rua_panic.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/portrait_panic.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, head and shoulders portrait, facing the viewer slightly turned right, centered, panicked, sweating Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 51. `portrait_cu_rua_cry` · ưu tiên 5

- **Lưu thành**: `assets/_raw/portrait_cu_rua_cry.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/portrait_cry.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, head and shoulders portrait, facing the viewer slightly turned right, centered, crying with streams of tears Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 52. `char_a_hon_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_a_hon_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/a_hon/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a loud-mouthed scrawny street vendor with a giant steamed bun as a hat that has a big bite missing, two huge yellowish buck teeth with one chipped, a mole on his cheek with one long curly hair, sweat-stained yellowed sleeveless undershirt riding up over a round hairy belly button, baggy red shorts, mismatched flip-flops, color palette #FFFFFF #E63946 #F4A261 #2B2D42. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 53. `char_ba_tam_loa_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_ba_tam_loa_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/ba_tam_loa/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a stout loud fish market lady holding a dented megaphone, bright red lipstick smeared far outside her lips and stuck on her teeth, two perfectly round blotchy pink blush circles, eyebrows drawn way too high in pencil, a tiny pencil stuck in her hair bun, floral blouse with fish scales stuck on it, wet rubber apron, tall green rubber boots, color palette #2A9D8F #E9C46A #F28482 #264653. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 54. `char_ong_ba_vong_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_ong_ba_vong_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/ong_ba_vong/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a skinny lazy old man with only two teeth left, nose hairs so long they look like a mustache, three strands of hair combed over a shiny bald head, half-closed sleepy eyes, toothpick in mouth, folded hammock carried on his back like a cape, saggy white tank top, striped pajama pants slipping down showing flowery underwear waistband, color palette #8D99AE #EDF2F4 #6D597A #B56576. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 55. `char_dai_ca_dep_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_dai_ca_dep_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/dai_ca_dep/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a pompous short gang boss in an oversized shiny purple suit with shoulder pads, tiny round black sunglasses, a gold chain as thick as a dog leash, a thin pencil mustache, slicked-back hair so greasy that flies buzz around it, a gold front tooth, cheap honeycomb plastic sandals with socks, color palette #7209B7 #FFD60A #000000 #F72585. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 56. `char_map_tra_sua_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_map_tra_sua_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/map_tra_sua/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a round chubby teen with huge jiggly cheeks and a double chin, tapioca pearls stuck around his mouth, clutching a giant bubble tea cup with a fat straw, oversized hoodie with sweat-soaked armpits, tiny shorts, untied sneakers, color palette #CDB4DB #A47148 #FFC8DD #3D405B. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 57. `char_co_ba_ganh_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_co_ba_ganh_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/co_ba_ganh/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a hyperactive young fruit seller with one thick unibrow, a huge gap-toothed grin, two stiff braids sticking straight out sideways, a crooked conical hat, balancing a bamboo shoulder pole with two overflowing fruit baskets, patched brown blouse, mismatched plastic slippers, color palette #BC6C25 #DDA15E #606C38 #FEFAE0. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 58. `char_su_phu_mi_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_su_phu_mi_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/su_phu_mi/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of an eccentric old kung fu master whose long beard and eyebrows are made of curly instant noodles, cross-eyed, bright red bulbous nose, one eye squinting, stained orange robe with soup splashes, holding a pair of chopsticks like a weapon, color palette #FB8500 #FFB703 #8ECAE6 #023047. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 59. `char_be_ti_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_be_ti_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/be_ti/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a very tiny man with absurdly huge oiled shiny bodybuilder muscles, a chubby baby face with scruffy stubble, a sweatband, a tiny wrestling singlet, veins popping on his forehead, constantly flexing, color palette #E76F51 #F4A261 #2A9D8F #1D3557. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 60. `char_shipper_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_shipper_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/shipper/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a sunburned delivery rider with a dented helmet, a pale sunglasses tan line across his face, a face mask pulled down under his chin, a fake overly polite grin, reflective vest, a giant delivery box on his back with a leaking sauce stain, phone in hand, color palette #FFBE0B #FB5607 #3A86FF #222222. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

## 61. `char_cu_rua_ref` · ưu tiên 9

- **Lưu thành**: `assets/_raw/char_cu_rua_ref.png`
- **Bố cục**: 1 hình ở giữa, nền magenta #FF00FF
- **Ra file game**: characters/cu_rua/ref_sheet.png

```text
STYLE: 2D game sprite, grotesque comedic caricature, ugly-funny lowbrow slapstick cartoon, huge head on a small scrawny or saggy body, shamelessly exaggerated facial features: crooked or missing teeth, bulging or droopy mismatched eyes, big pores and oily shiny skin, moles with a single long hair, sweat drops, rosy blotchy cheeks, double chin, over-the-top smug or stupid facial expressions exaggerated 200 percent, cheap mismatched clothes with stains, thick wobbly black outline, flat saturated colors with simple cel shading, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy, pure flat magenta #FF00FF background filling everything around the subject, no text, no watermark, no ground shadow, full body visible. Character reference sheet of a tiny ancient grandpa with eyebrows so long they drag on the ground, giant earlobes, a toothless peaceful smile, wearing a turtle shell backpack covered in cheap stickers, walking cane, flip-flops, color palette #52B788 #D8F3DC #6F4E37 #F1FAEE. Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions (neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view, separated panels. Avoid: cute, kawaii, pretty, handsome, beautiful, elegant, clean, polished, heroic, idealized, realistic, photo, 3d render, blurry, extra limbs, cropped body, text, letters, numbers, logo, watermark, background scenery, multiple different characters, inconsistent outfit, nudity, grid lines, frame borders.
```

