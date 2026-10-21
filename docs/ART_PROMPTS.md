# KIẾM XOAY TU TIÊN — Art, Animation & Audio Prompt Pack

> Dùng cho tool sinh ảnh (minigen). Mọi asset 2D, nền trong suốt (PNG RGBA).
> Prompt viết tiếng Anh vì đa số model ảnh hiểu tốt hơn; mô tả tiếng Việt kèm theo để bạn chỉnh.

---

## 1. Quy chuẩn kỹ thuật (BẮT BUỘC cho mọi frame)

| Thuộc tính | Giá trị |
|---|---|
| Canvas mỗi frame | 512×512 px, nền trong suốt |
| Góc nhìn | 3/4 top-down, nhân vật hơi nghiêng **quay sang PHẢI** (client tự lật trái) |
| Tỉ lệ | Chibi, đầu to ≈ 40% chiều cao cơ thể |
| Vị trí | Nhân vật cao ~70% canvas, **chân chạm y = 460px, tâm x = 256px** (pivot cố định) |
| Viền | Viền đen dày 4–6px, tô màu phẳng + 1 lớp bóng (cel shading) |
| Không có | Chữ, watermark, nền, bóng đổ dưới chân (game tự vẽ bóng) |

Kích thước nhỏ hơn được: item/kiếm 256×256, VFX 256×256 hoặc 512×512.

Lưu ý về scale: nhân vật to tới 5× trong thế giới nhưng camera lùi ra tương ứng, nên trên màn hình chỉ to tối đa ~1.3× → sprite 512px là đủ nét, không cần vẽ bản to riêng. Vật cản lớn (đá tảng, tường) vẽ ở 1024×1024.

## 2. Style lock (dán vào ĐẦU mọi prompt)

```
STYLE: 2D mobile game sprite, chibi cartoon caricature, exaggerated comedic facial expressions,
big head small body, thick black outline, flat colors with simple cel shading, vibrant saturated palette,
clean vector-like look, Vietnamese street comedy vibe mixed with wuxia cultivation fantasy,
transparent background, no text, no watermark, no ground shadow, centered, full body visible
```

Negative prompt (nếu tool hỗ trợ):
```
realistic, photo, 3d render, blurry, extra limbs, cropped body, text, logo, watermark, background scenery,
multiple characters, inconsistent outfit
```

## 3. Quy trình giữ nhất quán khi sinh keyframe

1. **Bước 1 — Character reference sheet**: sinh 1 ảnh tham chiếu cho mỗi nhân vật (prompt mục 4.1). Chọn bản đẹp nhất, khóa seed.
2. **Bước 2 — Sinh frame**: mỗi frame dùng ảnh tham chiếu làm image reference + cùng seed + prompt frame (mục 4.2). Nếu tool hỗ trợ, sinh cả **sprite sheet dạng lưới** trong 1 ảnh (mục 4.3) — thường đồng đều hơn sinh lẻ.
3. **Bước 3 — Hậu kỳ** (tự động hóa bằng script):
   - Xóa nền (nếu model không ra nền trong suốt).
   - Căn pivot: tìm bounding box, đặt chân về y=460, x=256.
   - Chuẩn hóa màu theo palette của nhân vật (mục 5).
   - Xuất GIF preview để duyệt: `ffmpeg` hoặc `gifski`.
   - Ghép atlas: `tools/atlas-pack` → `character_atlas.png` + `character_atlas.json`.
4. **Bước 4 — Duyệt**: xem GIF loop, frame nào lệch trang phục/mặt thì sinh lại riêng frame đó.

## 4. Template prompt nhân vật

### 4.1 Reference sheet
```
{STYLE}
Character reference sheet of {CHARACTER_DESCRIPTION}.
Show 3 views in a row: front, 3/4 facing right, side facing right. Below: 6 facial expressions
(neutral, screaming, evil grin, panicked, crying, shocked). Same outfit and colors in every view.
Color palette: {PALETTE}. White background separated panels.
```

### 4.2 Một keyframe
```
{STYLE}
{CHARACTER_DESCRIPTION}, 3/4 view facing right.
Animation "{ANIM_NAME}", frame {i} of {N}: {POSE_DESCRIPTION_FOR_FRAME_i}.
Expression: {EXPRESSION}. Keep exact same outfit, colors, proportions as reference.
512x512, feet at bottom center, transparent background.
```

### 4.3 Sprite sheet dạng lưới
```
{STYLE}
Sprite sheet, {COLS}x{ROWS} grid, each cell 512x512, of {CHARACTER_DESCRIPTION} facing right,
animation "{ANIM_NAME}" in {N} sequential frames read left-to-right, top-to-bottom:
1) {pose1} 2) {pose2} ... {N}) {poseN}.
Identical character in every cell, same size, feet aligned on same baseline, transparent background.
```

### 4.4 Danh sách animation & keyframe

| Anim | Frames | Loop | Mô tả từng frame (dán vào POSE_DESCRIPTION) |
|---|---|---|---|
| `idle` | 4 | ✔ | 1 đứng thở ra · 2 hít vào vai nhô · 3 ngáp há to · 4 gãi đầu |
| `run` | 8 | ✔ | Chạy guồng chân cartoon, tay vung loạn xạ, **miệng há la hét**, tóc bay; 8 pha chu kỳ chạy (contact, down, pass, up × 2 chân) |
| `chase` | 8 | ✔ | Chạy rón rén nghiêng người tới trước, **cười nham hiểm** mắt híp, xoa hai tay vào nhau; 8 pha chạy |
| `flee` | 8 | ✔ | Chạy hoảng loạn, **hai tay giơ lên trời**, mồ hôi văng, mắt trợn, miệng méo; 8 pha chạy |
| `hit` | 4 | ✘ | 1 mặt bị ép bẹp sang trái · 2 người văng ra sau · 3 mắt xoắn ốc, sao quanh đầu · 4 lắc đầu tỉnh lại |
| `levelup` | 6 | ✘ | 1 thủ thế · 2–4 hào quang bùng lên, tạo dáng siêu anh hùng · 5 giơ ngón cái · 6 về thủ thế |
| `breakthrough` | 10 | ✘ | 1 ngồi thiền · 2–3 run rẩy · 4 tóc dựng đứng · 5–6 ngửa mặt hét lên trời · 7–8 sét đánh trúng người · 9 cháy xém tóc xù khói · 10 cười đắc thắng |
| `realm_drop` | 6 | ✘ | 1 sững sờ · 2 đầu gối khuỵu · 3 ngồi bệt · 4–5 khóc thành suối nước mắt · 6 lau mũi đứng dậy |
| `death` | 8 | ✘ | 1–2 bị đập · 3–4 dẹp lép như bánh tráng · 5 hồn trắng bay ra · 6–8 hồn vẫy tay bay lên |
| `respawn` | 6 | ✘ | 1 rơi từ trên cao · 2 sắp chạm đất · 3 đáp cắm mặt · 4 lộn nhào · 5–6 đứng dậy phủi bụi tạo dáng |
| `teleport_out` | 4 | ✘ | 1 giật mình · 2 người xoắn lại như bị hút · 3 chỉ còn cái đầu mếu máo · 4 còn lại đôi dép |
| `teleport_in` | 4 | ✘ | 1 bị phun ra giữa không trung · 2 chới với · 3 đáp đất loạng choạng · 4 mắt xoắn ốc chóng mặt |

Tổng: 76 frame / nhân vật.

## 5. Dàn nhân vật (nguyên bản)

> Nhân vật trong game nên là thiết kế của riêng bạn để tránh rủi ro bản quyền và hình ảnh diễn viên thật khi phát hành. Dưới đây là 10 nhân vật gốc theo tinh thần "dân phố võ lâm tấu hài".

| # | Tên | CHARACTER_DESCRIPTION (prompt) | Palette |
|---|---|---|---|
| 1 | A Hớn Bánh Bao | `a loud-mouthed young street vendor wearing a giant steamed bun as a hat, sleeveless white undershirt, baggy red shorts, flip-flops, huge buck teeth` | #FFFFFF #E63946 #F4A261 #2B2D42 |
| 2 | Bà Tám Loa | `a sturdy middle-aged fish market lady holding a megaphone, floral blouse, rubber apron, tall green rubber boots, bun hair with a pencil stuck in it` | #2A9D8F #E9C46A #F28482 #264653 |
| 3 | Ông Ba Võng | `a sleepy skinny old man carrying a folded hammock on his back like a cape, white tank top, striped pajama pants, half-closed eyes, toothpick in mouth` | #8D99AE #EDF2F4 #6D597A #B56576 |
| 4 | Đại Ca Dép Tổ Ong | `a pompous gang boss in an oversized shiny purple suit, tiny black sunglasses, gold chain, wearing cheap honeycomb plastic sandals, slicked-back hair` | #7209B7 #FFD60A #000000 #F72585 |
| 5 | Thằng Mập Trà Sữa | `a round chubby teen always clutching a giant bubble tea cup with a fat straw, oversized hoodie, cheeks puffed, tapioca pearls flying around` | #CDB4DB #A47148 #FFC8DD #3D405B |
| 6 | Cô Ba Gánh Hàng | `a fast nimble young woman balancing a bamboo shoulder pole with two baskets of fruit, conical hat, simple brown blouse, determined eyebrows` | #BC6C25 #DDA15E #606C38 #FEFAE0 |
| 7 | Sư Phụ Mì Gói | `an eccentric old kung fu master whose long beard and eyebrows are made of curly instant noodles, orange robe, holding chopsticks like weapons` | #FB8500 #FFB703 #8ECAE6 #023047 |
| 8 | Bé Tí Cơ Bắp | `a very tiny man with absurdly huge bodybuilder muscles, baby face, sweatband, tiny shorts, flexing constantly` | #E76F51 #F4A261 #2A9D8F #1D3557 |
| 9 | Anh Shipper Tia Chớp | `a delivery rider in a bright helmet and reflective vest carrying a big cube delivery box on his back, lightning bolt stickers, phone in hand` | #FFBE0B #FB5607 #3A86FF #222222 |
| 10 | Cụ Rùa Thiền | `a slow ancient grandpa wearing a turtle shell backpack, long white eyebrows covering eyes, walking cane, peaceful smile` | #52B788 #D8F3DC #6F4E37 #F1FAEE |

Khi cần thêm nhân vật: giữ công thức **1 nghề/đặc điểm đời thường + 1 đạo cụ biểu tượng + 1 nét mặt cường điệu**.

## 6. Quái (PvE)

Mỗi quái: `idle` 4 frame (lắc lư), `hurt` 2 frame, `death` 6 frame (vỡ tung thành viên EXP). Vẽ chừa vùng giữa bụng tròn trống để game in số HP lên.

| Loại | Prompt |
|---|---|
| Nhỏ — Bánh Bao Cáu | `{STYLE} an angry living steamed bun creature with tiny legs, steam puffing from head, round belly area left blank in the center` |
| Nhỏ — Dép Lào Biết Đi | `{STYLE} a sentient flip-flop sandal creature hopping, googly eyes, sticking tongue out, blank round circle on its sole` |
| Vừa — Nồi Cơm Điện Cắn | `{STYLE} a rice cooker monster with its lid as a biting mouth full of rice teeth, blank round panel on front` |
| Vừa — Gà Hoảng Loạn | `{STYLE} a panicking chubby chicken monster, feathers flying, bulging eyes, blank round badge on chest` |
| Lớn — Muỗi Khổng Lồ | `{STYLE} a giant fat mosquito monster with a straw-like nose, tiny wings, smug face, blank round area on belly` |
| Boss — Vua Ghế Nhựa | `{STYLE} a towering boss made of stacked red plastic stools, crown of a plastic basin, angry eyes, blank big circle in the middle` |

## 7. Kiếm theo cảnh giới (skin đổi khi đột phá)

Canvas 256×256, kiếm **thẳng đứng, mũi hướng LÊN**, cán ở dưới. Thêm bản `glow` 4 frame loop nếu muốn.

```
{STYLE} single game weapon icon, {SWORD}, vertical, tip pointing up, centered, transparent background
```

| Cảnh giới | {SWORD} |
|---|---|
| Luyện Khí | `a pair of bamboo chopsticks tied together as a sword` |
| Trúc Cơ | `a kitchen cleaver with a wooden handle` |
| Kim Đan | `a peach wood taoist sword with red tassel` |
| Nguyên Anh | `a polished steel jian sword` |
| Hóa Thần | `a flaming red-orange sword with fire trail` |
| Luyện Hư | `an ice crystal blue sword with frost sparkles` |
| Hợp Thể | `a crackling lightning yellow sword` |
| Độ Kiếp | `a dark purple cursed sword with smoky aura` |
| Đại Thừa | `a radiant jade and gold celestial sword glowing white` |

## 8. Vật phẩm

| Asset | Prompt |
|---|---|
| Bong bóng kiếm | `{STYLE} a glossy transparent soap bubble containing a small glowing sword, shiny highlight, 256x256` |
| EXP nhỏ/vừa/lớn | `{STYLE} a glowing {green|blue|gold} spirit orb gem, sparkly, 128x128` |

Vật phẩm buff: mỗi cái là 1 bong bóng chứa icon, cùng 1 template:
```
{STYLE} a glossy transparent soap bubble containing {ICON}, tinted {COLOR}, shiny highlight, 256x256
```

| Item | {ICON} | {COLOR} | Aura trên nhân vật (VFX loop 4 frame) |
|---|---|---|---|
| Tốc Kiếm | `a cyan spinning whirlwind around a small sword` | cyan | vòng gió xanh quanh vòng kiếm |
| Thần Hành | `a winged sandal` | orange | vệt khói chạy sau chân |
| Nam Châm | `a red horseshoe magnet` | red | sóng tròn hút vào |
| Trường Kiếm | `a very long thin sword with an up arrow` | purple | kiếm phát sáng tím |
| Kim Cang | `a golden buddhist-style round shield` | gold | quả cầu vàng trong suốt bao quanh người |
| Song Bội Đan | `a glowing pill with "x2" shape made of light` | green | đốm sáng xanh bay lên |
| Ẩn Thân Phù | `a paper talisman with swirling ink` | grey | (chỉ chủ nhân thấy mình mờ 40%) |

Rương Báu: `{STYLE} an ornate wooden treasure chest wrapped in red ribbons, glowing golden light leaking out, 512x512` — anim `idle` 4 frame rung lắc, `open` 6 frame bung nắp, đồ bay ra.

Icon buff cho HUD: dùng lại icon ở trên, nền tròn, 128×128.

Bong bóng: thêm anim `float` 4 frame (nhấp nhô) và `pop` 4 frame.

## 9. Bản đồ

Nền: tile vuông **seamless/tileable** 512×512, nhìn top-down.

```
{STYLE without "transparent background"} seamless tileable top-down ground texture, {BIOME}, cartoon game style,
no objects, no characters, even lighting, edges must tile perfectly
```

| Biome | {BIOME} |
|---|---|
| Sân làng | `soft green grass with small clover patches and tiny flowers` |
| Khu chợ | `cracked old concrete street with faded painted lines` |
| Bãi đất | `dry dusty brown dirt with pebbles` |

Decor (nền trong suốt, 256×256 hoặc 512×512, chỉ để trang trí, không chặn đường):
`a red plastic stool`, `an old bicycle leaning`, `a pile of watermelons`, `a street food cart`, `a banana tree`,
`a rusty water jar`, `a broken TV`, `a lazy sleeping dog`, `a bush with flowers`, `a stone lantern`.

Tường biên: `{STYLE} a seamless horizontal stone and bamboo fence segment, top-down 3/4 view, tileable left-right`.

### 9.1 Vật cản phá được (mỗi loại 4 trạng thái)

Mỗi vật cản cần 4 ảnh cùng góc nhìn, cùng kích thước canvas: `intact` · `cracked` · `nearly_broken` · `rubble` (đống đổ nát, nhìn là biết đi qua được). Thêm 1 anim `break` 6 frame (nổ vụn) và `regrow` 6 frame (mọc lại từ đất, kèm ánh sáng).

```
{STYLE} top-down 3/4 view game obstacle, {OBSTACLE}, state: {STATE}, transparent background, {SIZE}
STATE: "intact and solid" | "with visible cracks" | "heavily damaged, crumbling, pieces missing" | "flat pile of rubble and debris on the ground"
```

| Loại | {OBSTACLE} | {SIZE} |
|---|---|---|
| Hàng rào tre | `a straight bamboo fence segment tied with rope` | 1024x256 |
| Sạp hàng | `a street market stall with a striped canopy and fruit baskets` | 512x512 |
| Xe đẩy | `a wooden street food cart with a small stove` | 512x512 |
| Tường gạch | `a short red brick wall segment with moss` | 1024x512 |
| Đá tảng | `a huge round grey boulder with cartoon cracks lines` | 1024x1024 |
| Tường đá cổ (không phá) | `an ancient mossy stone wall with carved patterns` | 1024x512 |

### 9.2 Vùng địa hình

| Vùng | Prompt | Ghi chú |
|---|---|---|
| Bụi rậm | `{STYLE} top-down dense leafy bush cluster, round, lush green, 512x512` | Vẽ 2 lớp: thân bụi + lớp lá trên cùng bán trong suốt để thấy mờ mờ người bên trong (chỉ với chính mình) |
| Bùn lầy | `seamless tileable top-down muddy brown puddle texture with bubbles` | Tile 512, viền bằng mask mềm trong code |
| Đường gió | `{STYLE} a white cartoon wind gust arrow with swirl lines` | Anim 4 frame chạy dọc làn gió |

### 9.3 Cổng dịch chuyển

```
{STYLE} top-down magical teleport portal on the ground, swirling {COLOR} vortex inside a carved stone ring
with taoist symbols, glowing, 512x512, frame {i} of {N}
```

| Anim | Frames | Mô tả |
|---|---|---|
| `idle` | 6 loop | Xoáy chậm |
| `warn` | 4 | Lóe sáng mạnh trước khi có người chui ra (0.5s) |
| `burst` | 6 | Bùng sáng khi người chui vào/ra |

Mỗi cặp/mạng cổng một màu (tím, xanh lam, cam) để người chơi đoán được cổng nối với nhau.

## 10. VFX

| VFX | Frames | Prompt |
|---|---|---|
| Tia lửa va kiếm | 6 | `{STYLE} comic sword clash spark burst, white and yellow star shapes, frame {i} of 6 expanding then fading` |
| Sao choáng | 4 loop | `{STYLE} ring of small yellow cartoon stars orbiting, frame {i} of 4` |
| Hào quang lên cấp | 6 | `{STYLE} rising golden aura ring with sparkles, frame {i} of 6` |
| Sét đột phá | 6 | `{STYLE} thick cartoon lightning bolt striking down from the top, frame {i} of 6` |
| Khói chết | 6 | `{STYLE} comic white smoke puff cloud, frame {i} of 6 expanding then dissipating` |
| Bụi chạy | 4 | `{STYLE} small dust puffs behind running feet, frame {i} of 4` |

| Kiếm bể | 6 | `{STYLE} a cartoon sword shattering into sharp shards flying outward, frame {i} of 6` |
| Vết nứt vòng kiếm | 1 | `{STYLE} thin white crack lines overlay, transparent, 256x256` (chồng lên kiếm khi bị nứt) |
| Khiên vỡ | 6 | `{STYLE} golden bubble shield cracking and bursting into glass shards, frame {i} of 6` |

Bong bóng thoại (UI): sinh khung trống `{STYLE} empty white comic speech bubble with black outline`; chữ do game render (dùng font hỗ trợ tiếng Việt).
Gợi ý câu thoại: `run`: "AAAAA!", "Tránh ra!"; `chase`: "Hehehe...", "Đứng lại đó!"; `flee`: "Cứu tôi với!", "Mẹ ơi!"; `hit`: "Ối giời!", "Đau quá!"; `breakthrough`: "Ta đã ngộ rồi!!!"; mất kiếm: "Kiếm của taaa!"; ra cổng: "Đây là đâu???"; nấp bụi: "Suỵt...".

## 11. UI

- 9 huy hiệu cảnh giới: `{STYLE} round game badge emblem for cultivation rank "{tier theme}", 256x256` (theme theo màu skin kiếm mục 7, KHÔNG để model viết chữ — chữ do game render).
- Joystick ảo: đế tròn mờ + núm tròn.
- Thanh EXP, khung leaderboard, khung kill feed: vẽ đơn giản bằng code hoặc 9-slice.
- Trang chủ: logo game (chữ do bạn tự thiết kế hoặc render bằng font), ảnh nền: `{STYLE without transparent} wide illustration of a chaotic cartoon village square with many chibi characters spinning swords around them, sunny, no text`.
- Màn hình Tổng kết: khung giấy cuộn cổ kính `{STYLE} an empty unrolled ancient paper scroll with wooden handles, transparent background`; ảnh "hồn bay" của nhân vật dùng lại frame cuối anim `death`.
- Bảng xếp hạng: 3 huy chương top 1–3 `{STYLE} {gold|silver|bronze} medal with a tiny sword emblem, 128x128`.

## 12. Âm thanh (SFX / BGM) — nguồn miễn phí

| Nguồn | Ghi chú giấy phép |
|---|---|
| kenney.nl | Đa số CC0, dùng thương mại thoải mái (có cả pack VFX/UI) |
| opengameart.org | Lọc theo CC0 hoặc CC-BY (CC-BY cần ghi credit) |
| freesound.org | Lọc theo license CC0; CC-BY cần credit; tránh CC-BY-NC nếu thương mại |
| pixabay.com (sound effects & music) | Giấy phép Pixabay, cho phép dùng thương mại |

**Luôn lưu file `CREDITS.md` ghi nguồn + license từng file.**

Danh sách SFX cần:
- `sword_spin_loop` (vù vù nhẹ, pitch tăng theo ω), `sword_clash` (keng ×3 biến thể), `sword_hit_body` (bốp, chát).
- `pickup_sword`, `pickup_buff`, `pickup_exp` (ting, pitch tăng nếu lượm liên tục), `buff_end` (xì hơi).
- `sword_break` (vỡ kính), `shield_break`, `chest_incoming` (trống báo), `chest_open`.
- `obstacle_hit` (cộc gỗ / đá), `obstacle_break` (đổ sập), `obstacle_regrow` (mọc lên "pựt").
- `gate_in` (hút "xoẹt"), `gate_out` (phụt "bùm"), `gate_warn`.
- `mud_step` (lép nhép), `bush_rustle` (sột soạt), `wind_loop`.
- `monster_hurt`, `monster_pop`.
- `levelup` (ding), `breakthrough` (sấm + "hoàaa"), `realm_drop` (kèn "wah wah" thất bại).
- `death` (tiếng "bẹp" + còi hoạt hình), `respawn` (rơi "vèoo" + đáp "bịch").
- Voice tấu hài (có thể tự thu bằng giọng team hoặc dùng TTS): la hét khi chạy, cười nham hiểm, kêu cứu, ối giời — mỗi loại 3–5 biến thể.
- BGM: 1 bài loop vui nhộn kiểu nhạc cổ phong/ngũ cung tốc độ nhanh.

## 13. Cấu trúc thư mục & đặt tên

```
assets/
├── characters/{char_id}/
│   ├── ref_sheet.png
│   ├── frames/{anim}_{00..N}.png
│   ├── atlas.png
│   └── atlas.json          # { "anims": { "run": { "frames": [...], "fps": 12, "loop": true } } }
├── monsters/{monster_id}/...
├── swords/realm_{0..8}.png
├── items/  buffs/  vfx/  tiles/  decor/  ui/
├── obstacles/{type}/{intact|cracked|nearly_broken|rubble}.png + break/ + regrow/
├── zones/  gates/{color}/
├── audio/sfx/  audio/voice/{char_id}/  audio/bgm/
└── CREDITS.md
```

FPS gợi ý: `idle` 6, `run`/`chase`/`flee` 12, các anim một lần 12–15.
`char_id` dạng snake_case không dấu: `a_hon`, `ba_tam_loa`, `ong_ba_vong`, `dai_ca_dep`, `map_tra_sua`, `co_ba_ganh`, `su_phu_mi`, `be_ti`, `shipper`, `cu_rua`.
