# Ghi chú bản dịch — BẢN NHÁP cần người bản ngữ duyệt

> **TẤT CẢ các file `en.json`, `pt-BR.json`, `es.json`, `id.json`, `th.json` đều là BẢN NHÁP**
> (có `"_draft": true`), do agent soạn ngày 01/11/2026. Theo GAME_SPEC §19, **mỗi ngôn ngữ
> phải có 1 người bản ngữ đọc duyệt trước khi phát hành**. Sau khi duyệt xong một ngôn ngữ:
> xoá dòng `"_draft": true` trong file đó (lệnh `npm run check:i18n` sẽ thôi ghi "bản nháp").

## Cách đọc file này

- Phần đầu: các quyết định chung + những chỗ agent **không chắc** (⚠) cho từng ngôn ngữ.
- Phần sau: bảng **mọi câu tấu hài** của từng ngôn ngữ, đặt cạnh câu gốc tiếng Việt và
  ngữ cảnh (hiện ở đâu, giới hạn độ dài). Người duyệt chỉ cần đọc bảng của ngôn ngữ mình.
- Câu tấu hài **được viết lại cho buồn cười theo từng ngôn ngữ**, không dịch sát chữ
  (spec §19). Vì vậy nghĩa có thể lệch khỏi câu Việt — đó là chủ ý. Người duyệt cứ sửa
  thoải mái miễn giữ: đúng số câu trong mỗi mảng, giữ nguyên `{a}` `{b}` `{realm}` `{title}`…,
  emote ≤ 22 ký tự, danh hiệu ≤ 32 ký tự, biệt danh ≤ 18 ký tự.

## Quyết định chung (áp dụng mọi ngôn ngữ)

| Mục | Cách làm |
|---|---|
| Thuật ngữ tu tiên | Theo bảng §19: Qi Refining … Dao Ancestor, Spirit Qi, Immortal Qi, Renown. Ngôn ngữ khác dịch nghĩa theo bản English. |
| `realmNames` | Điền đủ 14 tên theo đúng thứ tự `config/game.json` (9 cảnh giới phàm `cultivation.realms` + 5 Tiên `tien_realms`). |
| `characters`, `charactersShort` | Giữ nguyên tên Việt có dấu (bản sắc, spec §19). Biệt danh dịch nằm ở `charNick`. |
| `zoneNames`, `buffs`, `skillNames` | Dịch nghĩa (Heavenly Court, Bamboo Maze…), không giữ phiên âm Việt. |
| "Thiêu Đốt Tinh Huyết" | EN *Blood Burn* · PT *Queima de Sangue* · ES *Quema de Sangre* · ID *Bakar Darah* · TH *เผาโลหิต*. |
| `skillCooldown` (thẻ ~100px) | EN `CD {secs}s · {n} charges` · PT `Recarga {secs}s · {n} cargas` · ES `Rec. {secs}s · {n} cargas` · ID `CD {secs}d · {n} muatan` · TH `คูล {secs}วิ · สะสม {n}`. Cần xem thực tế trên thẻ có tràn không (PT dài nhất). |
| `titles[0]` ↔ `feed.title_gain` | Tin loan Top 1 trích nguyên văn danh hiệu Top 1 — sửa một chỗ thì phải sửa chỗ kia. `title_lose` chơi lại chữ của danh hiệu này. |
| `loadTips[10]`, `loadTips[11]` | Nhắc tới Thằng Mập / Đại Ca Dép bằng **biệt danh** của ngôn ngữ đó — sửa `charNick` thì sửa luôn 2 mẹo này. |
| Chữ "Daoist" trong emote `chase[6]` và SEO | Là câu tên game "Đạo hữu xin dừng bước!" — mỗi ngôn ngữ có bản tương đương. |
| `seo.title` | Dùng `{name}` = "BladeAscend". Spec §20 gợi ý khẩu hiệu EN *"Spin your swords, ascend to immortality"* — agent đặt `tagline` khác (giữ trò chơi chữ lặp của bản Việt); chủ dự án chọn. |

## Chỗ agent không chắc (⚠ cần người bản ngữ quyết)

### English
- ⚠ `tagline` "Swords go round and round, Daoists run around and around." — giữ trò lặp chữ của bản Việt; phương án khác theo spec §20: "Spin your swords, ascend to immortality".
- ⚠ `deathJokes[17]` gốc là meme "Có cái nịt" (không có gì cả) — viết thành "Got nothing. Not even a pulse." Mất meme, cần câu hay hơn.
- ⚠ `feed.gate_kill[0]` gốc có thành ngữ cờ bạc "đen thôi đỏ quên đi" — chỉ còn "Unlucky."
- ⚠ `emotes.hit[2]` "What the—!" thay cho "Á đù" (văng tục nhẹ) — giữ mức lịch sự.
- `realmTier` dùng "lv" ("Golden Core lv 3") cho ngắn.

### Português (Brasil)
- ⚠ `deathJokes[17]` "Ficou sem nada. Nem com o CPF." — thử bắt chước meme "có cái nịt"; không chắc người Brasil thấy buồn cười. Cần thay.
- ⚠ `emotes.sword_lost[4]` "Achado não é roubado?!" — đảo thành ngữ "nhặt được không phải ăn cắp" (gốc Việt: "Ai nhặt là mất dạy đó!"). Kiểm tra giọng có đúng ý "đừng có lụm" không.
- ⚠ `feed.realm_drop[4]` "foi pro saco" hơi suồng sã/tục nhẹ.
- `deathJokes[0]` đổi "nuôi cá" → "criar galinha", `[2]` "bánh bao" → "pastel na feira" (địa phương hoá).
- `skillLines.ong_ba_vong` dùng bài ru Brasil "Nana, neném… que a cuca vem pegar".
- `charNick.a_hon` "o Cara do Pãozinho" đúng 18 ký tự — tối đa.

### Español (Tây Ban Nha)
- ⚠ Viết theo giọng **Tây Ban Nha** (locale es-ES, dùng "vosotros": `skillLines.dai_ca_dep`, `cu_rua`, FAQ "entráis/pulsad"). Nếu thị trường chính là Mỹ Latinh thì nên đổi sang "ustedes" và từ vựng Latinh ("celular" thay "móvil", "computadora"…).
- ⚠ `emotes.sword_lost[4]` "¡El que la coja, caca!" — kiểu nói trẻ con, không chắc tự nhiên.
- ⚠ `deathJokes[17]` "Te quedaste sin nada. Ni con lo puesto." — chơi thành ngữ "con lo puesto"; cần người bản ngữ xác nhận.
- `emotes.chest[2]` "¡Me tocó el Gordo!" = giải độc đắc xổ số Giáng sinh Tây Ban Nha (người Latinh có thể không quen).
- `deathJokes` dùng churros, polvorón (địa phương hoá bánh bao, bánh tráng).

### Bahasa Indonesia
- ⚠ `realmNames`: cộng đồng truyện tu tiên Indonesia hay dùng luôn tên tiếng Anh (Qi Refining, Foundation Establishment…) hoặc Hán-Việt kiểu "Lianqi". Agent dịch nghĩa (Pemurnian Qi, Inti Emas, Jiwa Baru Lahir…) theo ghi chú §19 — cần người bản ngữ chọn.
- ⚠ `emotes.hit[2]` "Anjir!" — văng tục nhẹ, ngang mức "Á đù". Nếu muốn sạch hơn: "Astaga!".
- ⚠ `skillCooldown`/`skillWait` dùng "d" (detik) cho ngắn — có thể khó hiểu, cân nhắc "dtk".
- `emotes.chest[2]` "Sultan, bestie!" — tiếng lóng giới trẻ, có thể lỗi thời nhanh.
- `deathJokes` địa phương hoá: ternak lele, warkop, kerupuk.

### ไทย (Thái)
- ⚠ `realmNames` là chỗ **không chắc nhất**: agent dịch nghĩa (หลอมปราณ, สร้างรากฐาน, แก่นทอง, วิญญาณแรกกำเนิด, แปลงกายเทพ…). Truyện tu tiên dịch Thái có thể dùng tên khác (vd "ขั้นสร้างฐาน", "ลมปราณ"…). Cần người đọc truyện tu tiên Thái chọn tên quen thuộc.
- ⚠ `charNick.su_phu_mi` "อาจารย์มาม่า" — "มาม่า" (Mama) là **tên thương hiệu** mì gói ở Thái (dù người Thái dùng như danh từ chung). Spec tránh đụng thương hiệu → nếu ngại thì đổi "อาจารย์บะหมี่".
- ⚠ `Renown` dịch "ชื่อเสียง" (danh tiếng); phương án khác "บารมี" (uy danh, sắc thái tôn kính hơn).
- Tiếng Thái không dùng dấu chấm cuối câu → agent bỏ dấu chấm. Cần font **Noto Sans Thai** (spec §19) để hiển thị đúng.
- `titles[0]` "อาตมาไม่เคยกลัวใคร" dùng đại từ nhà sư "อาตมา" — sát nghĩa "bần tăng".
- `emotes.chest[0]` "ถูกหวย!" = trúng xổ số.

## Khoá không dịch được

Không có. `npm run check:i18n` báo 253/253 khoá cho cả 5 ngôn ngữ. Các khoá giữ nguyên
có chủ đích: `characters.*`, `charactersShort.*` (tên riêng Việt), `zoneEntered`,
`scorePop`, `loadMB`, `loadBg`, `skillWait`, `board.zoneHidden` (chỉ là placeholder/ký hiệu).

---

# Bảng câu tấu hài theo từng ngôn ngữ

## English (en-US) — `en.json`

**Trạng thái: BẢN NHÁP (`_draft: true`) — cần người bản ngữ duyệt trước khi phát hành.**


### `tagline`
Khẩu hiệu dưới logo trang chủ. Gốc: chơi chữ "vòng vòng" lặp đôi, có vần.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `tagline` |  | Swords go round and round, Daoists run around and around. |

### `charNick`
Biệt danh ở thẻ chọn nhân vật, dạng "A Hớn — *biệt danh*". ≤ 18 ký tự.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `charNick.a_hon` | A Hớn Bánh Bao | the Bun Guy |
| `charNick.ba_tam_loa` | Bà Tám Loa | Megaphone Auntie |
| `charNick.ong_ba_vong` | Ông Ba Võng | Hammock Grandpa |
| `charNick.dai_ca_dep` | Đại Ca Dép Tổ Ong | Flip-Flop Boss |
| `charNick.map_tra_sua` | Thằng Mập Trà Sữa | Boba Big Boy |
| `charNick.co_ba_ganh` | Cô Ba Gánh Hàng | the Street Vendor |
| `charNick.su_phu_mi` | Sư Phụ Mì Gói | Noodle Master |
| `charNick.be_ti` | Bé Tí Cơ Bắp | Tiny Muscles |
| `charNick.shipper` | Anh Shipper Tia Chớp | Lightning Courier |
| `charNick.cu_rua` | Cụ Rùa Thiền | Zen Turtle |

### `skillLines`
Câu thoại nhân vật hét khi dùng kỹ năng (bong bóng chữ trên đầu). Nên ngắn, đúng tính cách.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `skillLines.a_hon` | Ăn miếng bánh bao, kiếm về như mới! | One bite and my swords are brand new! |
| `skillLines.ba_tam_loa` | A lô a lô, tránh ra cho bà đi! | TESTING, TESTING! Out of my way! |
| `skillLines.ong_ba_vong` | À ơi… ngủ đi con… | Hush now… go to sleep… |
| `skillLines.dai_ca_dep` | Tụi bay đâu, lên đây! | Boys! Get over here! |
| `skillLines.map_tra_sua` | Nếm thử trân châu nhà tao! | Taste my house special boba! |
| `skillLines.co_ba_ganh` | Hàng về! Tránh ra cho cô bày! | Fresh stock! Make room, make room! |
| `skillLines.su_phu_mi` | Sùn sụt… kiếm mày ngon đấy! | *sluuurp*… tasty swords you got! |
| `skillLines.be_ti` | Gồng! Không đứa nào đẩy nổi! | FLEX! Nobody moves me! |
| `skillLines.shipper` | Đơn gấp, xin phép đi xuyên! | Urgent delivery, coming through walls! |
| `skillLines.cu_rua` | Thiền… các ngươi cứ chém đi. | Om… go ahead, swing away. |

### `titles`
Danh hiệu Top 1→Top 10 của phòng, uy giảm dần. ≤ 32 ký tự. Top 1 = sư khoe khoang "chưa ngán ai"; Top 10 = sống được nhờ núp.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `titles[0]` | Bần tăng chưa ngán ai bao giờ | This Monk Fears No One |
| `titles[1]` | Ngon thì nhào vô | Come at me, bro |
| `titles[2]` | Chỗ này nhiều máu nè | Plenty of HP over here |
| `titles[3]` | Chưa gặp đối thủ xứng tầm | Still looking for a worthy rival |
| `titles[4]` | Đi nhẹ nói khẽ, chém hơi mạnh | Walks softly, swings hard |
| `titles[5]` | Cũng có số má đấy chứ | Kind of a big deal |
| `titles[6]` | Tạm gọi là biết múa kiếm | Can wave a sword, sort of |
| `titles[7]` | Đang cố gắng từng ngày | Getting better every day |
| `titles[8]` | Thấy ai cũng chào trước | Says hi to everyone first |
| `titles[9]` | Nấp giỏi nên còn sống | Alive thanks to hiding skills |

### `deathJokes`
Câu đùa ngẫu nhiên ở màn hình Tổng kết khi chết. Không giới hạn cứng nhưng nên ≤ 1 dòng (~60 ký tự).

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `deathJokes[0]` | Bạn đã được tiễn về quê nuôi cá. | You've been sent back to the village to farm fish. |
| `deathJokes[1]` | Kiếm gãy, người bay, mộng tu tiên tạm hoãn. | Swords broken, body airborne, immortality postponed. |
| `deathJokes[2]` | Tu tiên không thành, về bán bánh bao tiếp. | Didn't ascend. Back to selling buns. |
| `deathJokes[3]` | Ăn một nhát kiếm, ngộ ra chân lý: chạy nhanh hơn. | Took one sword to the face, learned the Dao: run faster. |
| `deathJokes[4]` | Đạo tâm vỡ vụn như bánh tráng. | Your Dao heart shattered like a cheap rice cracker. |
| `deathJokes[5]` | Thôi thì kiếp sau làm quái cho đỡ khổ. | Maybe respawn as a monster next life. Less stress. |
| `deathJokes[6]` | Tu được ba hôm, nằm viện ba tháng. | Cultivated for three days, hospitalized for three months. |
| `deathJokes[7]` | Cảnh giới cao nhất bạn đạt được: cảnh giới nằm. | Highest realm achieved: Horizontal. |
| `deathJokes[8]` | Người ta phi thăng, bạn phi ra khỏi bản đồ. | Others ascend to the heavens. You ascended off the map. |
| `deathJokes[9]` | Sư phụ dặn né, bạn đứng lại ngắm. | Master said dodge. You stopped to admire the view. |
| `deathJokes[10]` | Độ kiếp thất bại ngay tại tầng gửi xe. | Tribulation failed. In the parking garage. |
| `deathJokes[11]` | Tu vi của bạn: đang trong quá trình tích lũy lại từ đầu. | Your cultivation is now being re-accumulated from zero. |
| `deathJokes[12]` | Kiếm người ta xoay, kiếm bạn xoay xuống đất. | Their swords spin. Yours spin into the ground. |
| `deathJokes[13]` | Linh khí trời đất không cứu nổi tốc độ phản xạ của bạn. | All the Spirit Qi in the world can't fix your reaction time. |
| `deathJokes[14]` | Chốt đơn: một vé về quê, không hoàn không huỷ. | Order confirmed: one ticket home. No refunds, no cancellations. |
| `deathJokes[15]` | Bạn đi rồi, bản đồ thoáng hẳn ra. | You're gone, and the map feels so roomy now. |
| `deathJokes[16]` | Thọ nguyên còn dài, nhưng kiếm người ta dài hơn. | Your lifespan was long. Their sword was longer. |
| `deathJokes[17]` | Có cái nịt, không có cái mạng. | Got nothing. Not even a pulse. |
| `deathJokes[18]` | Đạo hữu nằm xuống, nhạc nền nổi lên. | A Daoist falls. The sad violin starts playing. |
| `deathJokes[19]` | Hẹn gặp lại ở quán trà đá cuối thôn. | See you at the tea stall at the end of the village. |
| `deathJokes[20]` | Thiên đạo nhắn: "Thử lại sau nhé." | The Heavenly Dao says: "Please try again later." |
| `deathJokes[21]` | Bế quan mãi mãi, không ai làm phiền nữa. | Eternal closed-door meditation. Do not disturb. |

### `feed.kill`
Kill feed: {a} chém chết {b}. Hiện góc màn hình, mọi người trong phòng thấy.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.kill[0]` | {a} vừa tiễn {b} về quê nuôi cá | {a} sent {b} back to the village to farm fish |
| `feed.kill[1]` | {a} tặng {b} một vé về quê, không hoàn không huỷ | {a} gave {b} a one-way ticket home, no refunds |
| `feed.kill[2]` | {b} bị {a} chém bay cả giấc mộng tu tiên | {a} sliced {b}'s immortal dreams clean in half |
| `feed.kill[3]` | {a} dọn dẹp {b} gọn gàng sạch sẽ | {a} tidied {b} up nice and neat |
| `feed.kill[4]` | {b} nằm xuống, {a} phủi tay đi tiếp | {b} goes down, {a} dusts off and moves on |
| `feed.kill[5]` | {a} vừa cho {b} bế quan vĩnh viễn | {a} put {b} into permanent meditation |

### `feed.realm_drop`
Kill feed: {a} chém {b} rớt xuống cảnh giới {realm}.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.realm_drop[0]` | {b} bị {a} chém rớt xuống {realm} | {a} knocked {b} down to {realm} |
| `feed.realm_drop[1]` | {a} đạp {b} về lại {realm} | {a} kicked {b} back to {realm} |
| `feed.realm_drop[2]` | {b} tu bao ngày, về {realm} trong một nhát của {a} | {b} grinded for ages, back to {realm} after one swing from {a} |
| `feed.realm_drop[3]` | {a} dạy {b} một bài: {realm} thôi nhé | {a} taught {b} a lesson: {realm}, that's all |
| `feed.realm_drop[4]` | {b} rớt xuống {realm}, công sức bay theo gió | {b} dropped to {realm}, all that effort gone with the wind |

### `feed.breakthrough`
Kill feed: {a} đột phá lên {realm}. Câu 5 nhại meme "tuyệt đối điện ảnh" (absolute cinema).

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.breakthrough[0]` | {a} đột phá {realm}, cả phòng run rẩy | {a} broke through to {realm}, the whole room trembles |
| `feed.breakthrough[1]` | {a} lên {realm} rồi, ai nhanh chân thì chạy | {a} hit {realm} — run while you can |
| `feed.breakthrough[2]` | Linh khí dồn về một chỗ: {a} chạm {realm} | Spirit Qi gathers in one spot: {a} reaches {realm} |
| `feed.breakthrough[3]` | {a} bước vào {realm}, trời đất đổi màu | {a} enters {realm}, the sky changes color |
| `feed.breakthrough[4]` | {a} vừa lên {realm} — tuyệt đối điện ảnh | {a} just reached {realm} — absolute cinema |

### `feed.gate_kill`
Kill feed: {b} vừa ra khỏi cổng dịch chuyển thì bị {a} chém. Câu 1 gốc: "đen thôi đỏ quên đi" = thành ngữ cờ bạc, ý "xui thì chịu".

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.gate_kill[0]` | {b} chui cổng ra liền bị {a} chém, đen thôi đỏ quên đi | {b} popped out of a portal straight into {a}'s swords. Unlucky. |
| `feed.gate_kill[1]` | {b} vừa ló đầu khỏi cổng đã ăn kiếm của {a} | {b} barely poked their head out of the portal and ate {a}'s sword |
| `feed.gate_kill[2]` | {a} ngồi canh cổng, {b} xui xẻo bước ra | {a} camped the portal, {b} walked right out |
| `feed.gate_kill[3]` | {b} dịch chuyển hết một vòng để tới chỗ {a} chết | {b} teleported across the world just to die to {a} |

### `feed.title_gain`
Tin loan khi ai đó lên Top 1 — PHẢI trích đúng nguyên văn titles[0].

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.title_gain` | {a} đã lên ngôi "Bần tăng chưa ngán ai bao giờ"! | {a} has claimed the title "This Monk Fears No One"! |

### `feed.title_lose`
Tin loan khi Top 1 bị hạ — chơi lại chữ của titles[0] ("giờ đã ngán").

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.title_lose` | {a} bị hạ bệ, bần tăng giờ đã ngán | {a} was dethroned — this monk now fears someone |

### `loadTips.10`
Mẹo màn hình tải có nhắc biệt danh nhân vật (Thằng Mập) — phải khớp charNick.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `loadTips[10]` | Thằng Mập phun trân châu kín màn hình. 5 giây đó bạn cũng không nhặt được gì. | Boba Big Boy covers your whole screen in boba. For those 5 seconds you can't pick anything up either. |

### `loadTips.11`
Mẹo màn hình tải nhắc Đại Ca Dép — phải khớp charNick.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `loadTips[11]` | Đại Ca Dép gọi hai đàn em mang nửa số kiếm của hắn. Đừng xông vào giữa. | Flip-Flop Boss calls two minions carrying half his swords. Don't charge into the middle. |

### `emotes.*`
Emote: bong bóng chữ trên đầu nhân vật (người và bot), ≤ 22 ký tự.

| Khoá | Ngữ cảnh | Gốc tiếng Việt | Bản nháp |
|---|---|---|---|
| `sword_lost[0]` | khi mất kiếm | Kiếm của taaa! | My swooord! |
| `sword_lost[1]` | ″ | Trả kiếm đây! | Give it back! |
| `sword_lost[2]` | ″ | Kiếm mới mua mà! | I just bought that! |
| `sword_lost[3]` | ″ | Rụng mất một cây rồi! | One fell off! |
| `sword_lost[4]` | ″ | Ai nhặt là mất dạy đó! | Finders NOT keepers! |
| `shield_break[0]` | khi khiên Kim Cang vỡ | Khiên ơi! | My shield! |
| `shield_break[1]` | ″ | Vỡ rồi! | It broke! |
| `shield_break[2]` | ″ | Khiên giấy à trời! | Paper shield?! |
| `shield_break[3]` | ″ | Mới đeo được 3 giây! | Lasted 3 seconds! |
| `shield_break[4]` | ″ | Bảo hành đâu?! | Where's the warranty?! |
| `chest[0]` | khi mở Rương Báu ("Ét ô ét" = SOS kiểu Việt; "Đổi đời rồi bà con") | Trúng mánh! | Jackpot! |
| `chest[1]` | ″ | Giàu rồi! | I'm rich! |
| `chest[2]` | ″ | Đổi đời rồi bà con! | Life changed, folks! |
| `chest[3]` | ″ | Hôm nay số đỏ! | Lucky day! |
| `chest[4]` | ″ | Toàn đồ xịn! | All the good loot! |
| `chest[5]` | ″ | Ét ô ét, giàu quá! | Money money money! |
| `revealed[0]` | khi bị lộ khỏi Ẩn Thân | Lộ rồi! | Spotted! |
| `revealed[1]` | ″ | Ối! | Oops! |
| `revealed[2]` | ″ | Sao thấy tui hay vậy? | How'd you see me? |
| `revealed[3]` | ″ | Chưa núp xong mà! | I wasn't done hiding! |
| `revealed[4]` | ″ | Thôi xong, lộ hàng! | Welp, busted! |
| `run[0]` | khi đang chạy (Thiêu Đốt) | AAAAA! | AAAAA! |
| `run[1]` | ″ | Tránh ra! | Move! |
| `run[2]` | ″ | Xe tui không phanh! | No brakes! |
| `run[3]` | ″ | Dô dô dô! | Zoom zoom zoom! |
| `run[4]` | ″ | Chân ơi cố lên! | Legs, don't fail me! |
| `run[5]` | ″ | Nhanh nhanh nhanh! | Go go go! |
| `chase[0]` | khi đuổi người yếu hơn; câu cuối là tên game | Hehehe... | Hehehe... |
| `chase[1]` | ″ | Đứng lại đó! | Stop right there! |
| `chase[2]` | ″ | Chạy đâu cho thoát! | You can't escape! |
| `chase[3]` | ″ | Lại đây cưng! | C'mere, sweetie! |
| `chase[4]` | ″ | Đứng im một nhát thôi! | Hold still, just one! |
| `chase[5]` | ″ | Ngon, con mồi kìa! | Ooh, fresh prey! |
| `chase[6]` | ″ | Đạo hữu, dừng bước! | Daoist, please stop! |
| `flee[0]` | khi chạy trốn người mạnh hơn | Cứu tôi với! | Help me! |
| `flee[1]` | ″ | Mẹ ơi! | Mommy! |
| `flee[2]` | ″ | Tui còn nhỏ lắm! | I'm just a baby! |
| `flee[3]` | ″ | Nhầm người rồi! | Wrong guy! |
| `flee[4]` | ″ | Tha cho đệ! | Spare me! |
| `flee[5]` | ″ | Đệ xin hàng! | I surrender! |
| `flee[6]` | ″ | Không đánh nhau, hoà bình! | Peace, not war! |
| `hit[0]` | khi bị chém trúng ("Á đù" = văng tục nhẹ) | Ối giời! | Ouch! |
| `hit[1]` | ″ | Đau quá! | That hurts! |
| `hit[2]` | ″ | Á đù! | What the—! |
| `hit[3]` | ″ | Chơi bẩn! | Cheap shot! |
| `hit[4]` | ″ | Hự! | Oof! |
| `hit[5]` | ″ | Ai đánh lén đó?! | Who did that?! |
| `hit[6]` | ″ | Nhẹ tay coi! | Easy there! |
| `breakthrough[0]` | khi lên đại cảnh giới | Ta đã ngộ rồi!!! | I'm enlightened!!! |
| `breakthrough[1]` | ″ | Thiên địa chứng giám! | Heaven be my witness! |
| `breakthrough[2]` | ″ | Lên cảnh giới rồi nha! | New realm, baby! |
| `breakthrough[3]` | ″ | Nhìn ta đây này! | Look at me now! |
| `breakthrough[4]` | ″ | Cảm tạ trời đất! | Thank the heavens! |
| `levelup[0]` | khi lên tầng nhỏ | Lên tầng! | Level up! |
| `levelup[1]` | ″ | Thơm quá! | Smells good! |
| `levelup[2]` | ″ | Tiến bộ rồi nha! | Getting stronger! |
| `levelup[3]` | ″ | Thêm tí tu vi! | A bit more power! |
| `levelup[4]` | ″ | Mượt! | Smooth! |
| `realm_drop[0]` | khi bị rớt cảnh giới | Mất cảnh giới rồi! | Lost my realm! |
| `realm_drop[1]` | ″ | Công sức bao ngày... | All that grinding... |
| `realm_drop[2]` | ″ | Tụt hạng đau quá! | Demoted, ouch! |
| `realm_drop[3]` | ″ | Làm lại từ đầu vậy! | Back to square one! |
| `realm_drop[4]` | ″ | Trời không thương tui! | Heaven hates me! |
| `teleport_in[0]` | khi vừa ra khỏi cổng | Đây là đâu??? | Where am I??? |
| `teleport_in[1]` | ″ | Ủa? Lạc rồi! | Huh? I'm lost! |
| `teleport_in[2]` | ″ | Cổng gì chóng mặt vậy! | Dizzy portal! |
| `teleport_in[3]` | ″ | Ai đưa tui tới đây? | Who brought me here? |
| `teleport_in[4]` | ″ | Về nhầm nhà rồi! | Wrong house! |
| `teleport_out[0]` | khi chui vào cổng | Tui biến đây! | I'm outta here! |
| `teleport_out[1]` | ″ | Bái bai nha! | Bye bye! |
| `teleport_out[2]` | ″ | Hẹn gặp bên kia! | See ya over there! |
| `teleport_out[3]` | ″ | Chui cổng cái đã! | Portal time! |
| `death[0]` | khi chết | Thôi xong... | Welp... |
| `death[1]` | ″ | Tui đi trước nha! | I'll go first! |
| `death[2]` | ″ | Hẹn kiếp sau! | See you next life! |
| `death[3]` | ″ | Tin được không trời! | Unbelievable! |
| `respawn[0]` | khi bot hồi sinh | Tui lại về rồi! | I'm back! |
| `respawn[1]` | ″ | Chưa hết đâu nha! | Not over yet! |
| `respawn[2]` | ″ | Hồi sinh, chém tiếp! | Respawn, slash again! |
| `respawn[3]` | ″ | Nghỉ tí rồi tu lại! | Quick nap, back to it! |
| `no_swords[0]` | khi hết kiếm | Hết kiếm rồi! | Out of swords! |
| `no_swords[1]` | ″ | Tui tay không nè! | Bare hands here! |
| `no_swords[2]` | ″ | Ai cho mượn cây kiếm... | Anyone got a sword? |
| `no_swords[3]` | ″ | Thôi chạy đã! | Run first! |
| `bush[0]` | khi núp bụi cỏ | Suỵt... | Shhh... |
| `bush[1]` | ″ | Không có ai ở đây đâu. | Nobody here. |
| `bush[2]` | ″ | Tui là cái bụi cỏ. | I am a bush. |
| `bush[3]` | ″ | Đừng nhìn, đừng nhìn... | Don't look at me... |
| `shocked[0]` | khi thấy người khác đột phá / chém đẹp; câu cuối = meme "tuyệt đối điện ảnh" | Hả??? | Huh??? |
| `shocked[1]` | ″ | Trời đất ơi! | Oh my heavens! |
| `shocked[2]` | ″ | Ghê vậy! | Whoa! |
| `shocked[3]` | ″ | Gì cơ?! | Say what?! |
| `shocked[4]` | ″ | Ông này ghê quá! | This guy is nuts! |
| `shocked[5]` | ″ | Tuyệt đối điện ảnh! | Absolute cinema! |



## Português Brasil (pt-BR) — `pt-BR.json`

**Trạng thái: BẢN NHÁP (`_draft: true`) — cần người bản ngữ duyệt trước khi phát hành.**


### `tagline`
Khẩu hiệu dưới logo trang chủ. Gốc: chơi chữ "vòng vòng" lặp đôi, có vần.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `tagline` |  | As espadas giram, giram, e os taoístas correm, correm. |

### `charNick`
Biệt danh ở thẻ chọn nhân vật, dạng "A Hớn — *biệt danh*". ≤ 18 ký tự.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `charNick.a_hon` | A Hớn Bánh Bao | o Cara do Pãozinho |
| `charNick.ba_tam_loa` | Bà Tám Loa | Tia do Megafone |
| `charNick.ong_ba_vong` | Ông Ba Võng | Vovô da Rede |
| `charNick.dai_ca_dep` | Đại Ca Dép Tổ Ong | Chefão do Chinelo |
| `charNick.map_tra_sua` | Thằng Mập Trà Sữa | Gordinho do Chá |
| `charNick.co_ba_ganh` | Cô Ba Gánh Hàng | a Camelô |
| `charNick.su_phu_mi` | Sư Phụ Mì Gói | Mestre do Miojo |
| `charNick.be_ti` | Bé Tí Cơ Bắp | Pequeno Marombeiro |
| `charNick.shipper` | Anh Shipper Tia Chớp | Motoboy Relâmpago |
| `charNick.cu_rua` | Cụ Rùa Thiền | Tartaruga Zen |

### `skillLines`
Câu thoại nhân vật hét khi dùng kỹ năng (bong bóng chữ trên đầu). Nên ngắn, đúng tính cách.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `skillLines.a_hon` | Ăn miếng bánh bao, kiếm về như mới! | Uma mordida e minhas espadas voltam novinhas! |
| `skillLines.ba_tam_loa` | A lô a lô, tránh ra cho bà đi! | ALÔ, ALÔ! Sai da frente que lá vem a tia! |
| `skillLines.ong_ba_vong` | À ơi… ngủ đi con… | Nana, neném… que a cuca vem pegar… |
| `skillLines.dai_ca_dep` | Tụi bay đâu, lên đây! | Cadê os parças? Chega mais! |
| `skillLines.map_tra_sua` | Nếm thử trân châu nhà tao! | Prova a bolinha da casa! |
| `skillLines.co_ba_ganh` | Hàng về! Tránh ra cho cô bày! | Chegou mercadoria! Abre espaço aí! |
| `skillLines.su_phu_mi` | Sùn sụt… kiếm mày ngon đấy! | *sluuurp*… que espada gostosa, hein! |
| `skillLines.be_ti` | Gồng! Không đứa nào đẩy nổi! | É MAROMBA! Ninguém me empurra! |
| `skillLines.shipper` | Đơn gấp, xin phép đi xuyên! | Entrega urgente, licença que eu atravesso! |
| `skillLines.cu_rua` | Thiền… các ngươi cứ chém đi. | Ommm… pode bater à vontade. |

### `titles`
Danh hiệu Top 1→Top 10 của phòng, uy giảm dần. ≤ 32 ký tự. Top 1 = sư khoe khoang "chưa ngán ai"; Top 10 = sống được nhờ núp.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `titles[0]` | Bần tăng chưa ngán ai bao giờ | Este Monge Não Teme Ninguém |
| `titles[1]` | Ngon thì nhào vô | Vem pro x1, vem |
| `titles[2]` | Chỗ này nhiều máu nè | Aqui tem muita vida, ó |
| `titles[3]` | Chưa gặp đối thủ xứng tầm | Procurando um rival à altura |
| `titles[4]` | Đi nhẹ nói khẽ, chém hơi mạnh | Fala baixo, bate forte |
| `titles[5]` | Cũng có số má đấy chứ | Tem seu nome na praça |
| `titles[6]` | Tạm gọi là biết múa kiếm | Sabe balançar uma espada |
| `titles[7]` | Đang cố gắng từng ngày | Melhorando a cada dia |
| `titles[8]` | Thấy ai cũng chào trước | Dá bom dia pra todo mundo |
| `titles[9]` | Nấp giỏi nên còn sống | Vivo de tanto se esconder |

### `deathJokes`
Câu đùa ngẫu nhiên ở màn hình Tổng kết khi chết. Không giới hạn cứng nhưng nên ≤ 1 dòng (~60 ký tự).

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `deathJokes[0]` | Bạn đã được tiễn về quê nuôi cá. | Você foi mandado de volta pro interior criar galinha. |
| `deathJokes[1]` | Kiếm gãy, người bay, mộng tu tiên tạm hoãn. | Espada quebrada, corpo voando, imortalidade adiada. |
| `deathJokes[2]` | Tu tiên không thành, về bán bánh bao tiếp. | Não virou imortal. Volta a vender pastel na feira. |
| `deathJokes[3]` | Ăn một nhát kiếm, ngộ ra chân lý: chạy nhanh hơn. | Levou uma espadada e aprendeu o Dao: corra mais rápido. |
| `deathJokes[4]` | Đạo tâm vỡ vụn như bánh tráng. | Seu coração do Dao se partiu igual biscoito de polvilho. |
| `deathJokes[5]` | Thôi thì kiếp sau làm quái cho đỡ khổ. | Na próxima vida renasce como monstro. Menos estresse. |
| `deathJokes[6]` | Tu được ba hôm, nằm viện ba tháng. | Cultivou três dias, ficou três meses no hospital. |
| `deathJokes[7]` | Cảnh giới cao nhất bạn đạt được: cảnh giới nằm. | Maior reino alcançado: Deitado. |
| `deathJokes[8]` | Người ta phi thăng, bạn phi ra khỏi bản đồ. | Os outros ascendem aos céus. Você ascendeu pra fora do mapa. |
| `deathJokes[9]` | Sư phụ dặn né, bạn đứng lại ngắm. | O mestre mandou desviar. Você parou pra apreciar a vista. |
| `deathJokes[10]` | Độ kiếp thất bại ngay tại tầng gửi xe. | Tribulação fracassada. No estacionamento do shopping. |
| `deathJokes[11]` | Tu vi của bạn: đang trong quá trình tích lũy lại từ đầu. | Seu cultivo está sendo reacumulado do zero. Aguarde. |
| `deathJokes[12]` | Kiếm người ta xoay, kiếm bạn xoay xuống đất. | As espadas dos outros giram. As suas giram no chão. |
| `deathJokes[13]` | Linh khí trời đất không cứu nổi tốc độ phản xạ của bạn. | Nem todo o Qi do universo salva seu reflexo. |
| `deathJokes[14]` | Chốt đơn: một vé về quê, không hoàn không huỷ. | Pedido confirmado: uma passagem pra casa. Sem reembolso. |
| `deathJokes[15]` | Bạn đi rồi, bản đồ thoáng hẳn ra. | Você saiu e o mapa ficou até mais espaçoso. |
| `deathJokes[16]` | Thọ nguyên còn dài, nhưng kiếm người ta dài hơn. | Sua vida era longa. A espada dele era mais. |
| `deathJokes[17]` | Có cái nịt, không có cái mạng. | Ficou sem nada. Nem com o CPF. |
| `deathJokes[18]` | Đạo hữu nằm xuống, nhạc nền nổi lên. | Um taoísta cai. Começa a tocar música triste de novela. |
| `deathJokes[19]` | Hẹn gặp lại ở quán trà đá cuối thôn. | Te vejo no boteco no fim da vila. |
| `deathJokes[20]` | Thiên đạo nhắn: "Thử lại sau nhé." | O Dao Celestial avisa: "Tente novamente mais tarde." |
| `deathJokes[21]` | Bế quan mãi mãi, không ai làm phiền nữa. | Meditação a portas fechadas eterna. Não perturbe. |

### `feed.kill`
Kill feed: {a} chém chết {b}. Hiện góc màn hình, mọi người trong phòng thấy.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.kill[0]` | {a} vừa tiễn {b} về quê nuôi cá | {a} mandou {b} de volta pro interior criar galinha |
| `feed.kill[1]` | {a} tặng {b} một vé về quê, không hoàn không huỷ | {a} deu a {b} uma passagem só de ida pra casa |
| `feed.kill[2]` | {b} bị {a} chém bay cả giấc mộng tu tiên | {a} cortou ao meio o sonho de imortalidade de {b} |
| `feed.kill[3]` | {a} dọn dẹp {b} gọn gàng sạch sẽ | {a} fez a faxina em {b}, limpinho |
| `feed.kill[4]` | {b} nằm xuống, {a} phủi tay đi tiếp | {b} caiu, {a} bateu a poeira e seguiu |
| `feed.kill[5]` | {a} vừa cho {b} bế quan vĩnh viễn | {a} colocou {b} em meditação eterna |

### `feed.realm_drop`
Kill feed: {a} chém {b} rớt xuống cảnh giới {realm}.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.realm_drop[0]` | {b} bị {a} chém rớt xuống {realm} | {a} derrubou {b} para {realm} |
| `feed.realm_drop[1]` | {a} đạp {b} về lại {realm} | {a} chutou {b} de volta para {realm} |
| `feed.realm_drop[2]` | {b} tu bao ngày, về {realm} trong một nhát của {a} | {b} ralou por dias e voltou a {realm} num golpe de {a} |
| `feed.realm_drop[3]` | {a} dạy {b} một bài: {realm} thôi nhé | {a} deu uma aula em {b}: {realm} e olhe lá |
| `feed.realm_drop[4]` | {b} rớt xuống {realm}, công sức bay theo gió | {b} caiu para {realm}, todo o esforço foi pro saco |

### `feed.breakthrough`
Kill feed: {a} đột phá lên {realm}. Câu 5 nhại meme "tuyệt đối điện ảnh" (absolute cinema).

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.breakthrough[0]` | {a} đột phá {realm}, cả phòng run rẩy | {a} avançou para {realm}, a sala inteira treme |
| `feed.breakthrough[1]` | {a} lên {realm} rồi, ai nhanh chân thì chạy | {a} chegou a {realm} — corre quem puder |
| `feed.breakthrough[2]` | Linh khí dồn về một chỗ: {a} chạm {realm} | O Qi se concentra num ponto: {a} alcança {realm} |
| `feed.breakthrough[3]` | {a} bước vào {realm}, trời đất đổi màu | {a} entra em {realm}, o céu muda de cor |
| `feed.breakthrough[4]` | {a} vừa lên {realm} — tuyệt đối điện ảnh | {a} acabou de chegar a {realm} — cinema absoluto |

### `feed.gate_kill`
Kill feed: {b} vừa ra khỏi cổng dịch chuyển thì bị {a} chém. Câu 1 gốc: "đen thôi đỏ quên đi" = thành ngữ cờ bạc, ý "xui thì chịu".

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.gate_kill[0]` | {b} chui cổng ra liền bị {a} chém, đen thôi đỏ quên đi | {b} saiu do portal direto na espada de {a}. Azar define |
| `feed.gate_kill[1]` | {b} vừa ló đầu khỏi cổng đã ăn kiếm của {a} | {b} mal botou a cara pra fora do portal e já levou espadada de {a} |
| `feed.gate_kill[2]` | {a} ngồi canh cổng, {b} xui xẻo bước ra | {a} campou o portal e {b} saiu bem na frente |
| `feed.gate_kill[3]` | {b} dịch chuyển hết một vòng để tới chỗ {a} chết | {b} se teleportou pro outro lado do mundo só pra morrer pra {a} |

### `feed.title_gain`
Tin loan khi ai đó lên Top 1 — PHẢI trích đúng nguyên văn titles[0].

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.title_gain` | {a} đã lên ngôi "Bần tăng chưa ngán ai bao giờ"! | {a} conquistou o título "Este Monge Não Teme Ninguém"! |

### `feed.title_lose`
Tin loan khi Top 1 bị hạ — chơi lại chữ của titles[0] ("giờ đã ngán").

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.title_lose` | {a} bị hạ bệ, bần tăng giờ đã ngán | {a} foi destronado — agora o monge tem medo |

### `loadTips.10`
Mẹo màn hình tải có nhắc biệt danh nhân vật (Thằng Mập) — phải khớp charNick.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `loadTips[10]` | Thằng Mập phun trân châu kín màn hình. 5 giây đó bạn cũng không nhặt được gì. | O Gordinho do Chá cobre sua tela de bolinhas. Nesses 5 segundos você também não pega nada. |

### `loadTips.11`
Mẹo màn hình tải nhắc Đại Ca Dép — phải khớp charNick.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `loadTips[11]` | Đại Ca Dép gọi hai đàn em mang nửa số kiếm của hắn. Đừng xông vào giữa. | O Chefão do Chinelo chama dois capangas com metade das espadas dele. Não entre no meio. |

### `emotes.*`
Emote: bong bóng chữ trên đầu nhân vật (người và bot), ≤ 22 ký tự.

| Khoá | Ngữ cảnh | Gốc tiếng Việt | Bản nháp |
|---|---|---|---|
| `sword_lost[0]` | khi mất kiếm | Kiếm của taaa! | Minha espadaaa! |
| `sword_lost[1]` | ″ | Trả kiếm đây! | Devolve aí! |
| `sword_lost[2]` | ″ | Kiếm mới mua mà! | Acabei de comprar! |
| `sword_lost[3]` | ″ | Rụng mất một cây rồi! | Caiu uma! |
| `sword_lost[4]` | ″ | Ai nhặt là mất dạy đó! | Achado não é roubado?! |
| `shield_break[0]` | khi khiên Kim Cang vỡ | Khiên ơi! | Meu escudo! |
| `shield_break[1]` | ″ | Vỡ rồi! | Quebrou! |
| `shield_break[2]` | ″ | Khiên giấy à trời! | Escudo de papel?! |
| `shield_break[3]` | ″ | Mới đeo được 3 giây! | Durou 3 segundos! |
| `shield_break[4]` | ″ | Bảo hành đâu?! | Cadê a garantia?! |
| `chest[0]` | khi mở Rương Báu ("Ét ô ét" = SOS kiểu Việt; "Đổi đời rồi bà con") | Trúng mánh! | Na mosca! |
| `chest[1]` | ″ | Giàu rồi! | Tô rico! |
| `chest[2]` | ″ | Đổi đời rồi bà con! | Mudei de vida, galera! |
| `chest[3]` | ″ | Hôm nay số đỏ! | Dia de sorte! |
| `chest[4]` | ″ | Toàn đồ xịn! | Só item top! |
| `chest[5]` | ″ | Ét ô ét, giàu quá! | Ganhei na Mega-Sena! |
| `revealed[0]` | khi bị lộ khỏi Ẩn Thân | Lộ rồi! | Me acharam! |
| `revealed[1]` | ″ | Ối! | Opa! |
| `revealed[2]` | ″ | Sao thấy tui hay vậy? | Como me viu? |
| `revealed[3]` | ″ | Chưa núp xong mà! | Nem tinha escondido! |
| `revealed[4]` | ″ | Thôi xong, lộ hàng! | Ih, fui pego! |
| `run[0]` | khi đang chạy (Thiêu Đốt) | AAAAA! | AAAAA! |
| `run[1]` | ″ | Tránh ra! | Sai da frente! |
| `run[2]` | ″ | Xe tui không phanh! | Tô sem freio! |
| `run[3]` | ″ | Dô dô dô! | Vruuum! |
| `run[4]` | ″ | Chân ơi cố lên! | Pernas, não falhem! |
| `run[5]` | ″ | Nhanh nhanh nhanh! | Vai vai vai! |
| `chase[0]` | khi đuổi người yếu hơn; câu cuối là tên game | Hehehe... | Hehehe... |
| `chase[1]` | ″ | Đứng lại đó! | Para aí! |
| `chase[2]` | ″ | Chạy đâu cho thoát! | Não tem pra onde ir! |
| `chase[3]` | ″ | Lại đây cưng! | Vem cá, fofinho! |
| `chase[4]` | ″ | Đứng im một nhát thôi! | Fica parado, só uma! |
| `chase[5]` | ″ | Ngon, con mồi kìa! | Opa, presa fresca! |
| `chase[6]` | ″ | Đạo hữu, dừng bước! | Taoísta, pare aí! |
| `flee[0]` | khi chạy trốn người mạnh hơn | Cứu tôi với! | Socorro! |
| `flee[1]` | ″ | Mẹ ơi! | Mãããe! |
| `flee[2]` | ″ | Tui còn nhỏ lắm! | Sou só um bebê! |
| `flee[3]` | ″ | Nhầm người rồi! | Pessoa errada! |
| `flee[4]` | ″ | Tha cho đệ! | Me poupa! |
| `flee[5]` | ″ | Đệ xin hàng! | Eu me rendo! |
| `flee[6]` | ″ | Không đánh nhau, hoà bình! | Paz e amor! |
| `hit[0]` | khi bị chém trúng ("Á đù" = văng tục nhẹ) | Ối giời! | Ai! |
| `hit[1]` | ″ | Đau quá! | Doeu! |
| `hit[2]` | ″ | Á đù! | Mas que—! |
| `hit[3]` | ″ | Chơi bẩn! | Jogo sujo! |
| `hit[4]` | ″ | Hự! | Uff! |
| `hit[5]` | ″ | Ai đánh lén đó?! | Quem foi?! |
| `hit[6]` | ″ | Nhẹ tay coi! | Pega leve! |
| `breakthrough[0]` | khi lên đại cảnh giới | Ta đã ngộ rồi!!! | Fui iluminado!!! |
| `breakthrough[1]` | ″ | Thiên địa chứng giám! | Os céus que me vejam! |
| `breakthrough[2]` | ″ | Lên cảnh giới rồi nha! | Reino novo, bebê! |
| `breakthrough[3]` | ″ | Nhìn ta đây này! | Olha eu aqui! |
| `breakthrough[4]` | ″ | Cảm tạ trời đất! | Graças aos céus! |
| `levelup[0]` | khi lên tầng nhỏ | Lên tầng! | Subi de nível! |
| `levelup[1]` | ″ | Thơm quá! | Que delícia! |
| `levelup[2]` | ″ | Tiến bộ rồi nha! | Tô evoluindo! |
| `levelup[3]` | ″ | Thêm tí tu vi! | Mais um tiquinho! |
| `levelup[4]` | ″ | Mượt! | Liso! |
| `realm_drop[0]` | khi bị rớt cảnh giới | Mất cảnh giới rồi! | Perdi o reino! |
| `realm_drop[1]` | ″ | Công sức bao ngày... | Tanto esforço... |
| `realm_drop[2]` | ″ | Tụt hạng đau quá! | Rebaixado, ai! |
| `realm_drop[3]` | ″ | Làm lại từ đầu vậy! | Começar do zero! |
| `realm_drop[4]` | ″ | Trời không thương tui! | O céu me odeia! |
| `teleport_in[0]` | khi vừa ra khỏi cổng | Đây là đâu??? | Onde eu tô??? |
| `teleport_in[1]` | ″ | Ủa? Lạc rồi! | Ué? Me perdi! |
| `teleport_in[2]` | ″ | Cổng gì chóng mặt vậy! | Portal tonto! |
| `teleport_in[3]` | ″ | Ai đưa tui tới đây? | Quem me trouxe aqui? |
| `teleport_in[4]` | ″ | Về nhầm nhà rồi! | Casa errada! |
| `teleport_out[0]` | khi chui vào cổng | Tui biến đây! | Fui! |
| `teleport_out[1]` | ″ | Bái bai nha! | Tchau tchau! |
| `teleport_out[2]` | ″ | Hẹn gặp bên kia! | Te vejo lá! |
| `teleport_out[3]` | ″ | Chui cổng cái đã! | Bora de portal! |
| `death[0]` | khi chết | Thôi xong... | Já era... |
| `death[1]` | ″ | Tui đi trước nha! | Vou na frente! |
| `death[2]` | ″ | Hẹn kiếp sau! | Até a próxima vida! |
| `death[3]` | ″ | Tin được không trời! | Não acredito! |
| `respawn[0]` | khi bot hồi sinh | Tui lại về rồi! | Voltei! |
| `respawn[1]` | ″ | Chưa hết đâu nha! | Ainda não acabou! |
| `respawn[2]` | ″ | Hồi sinh, chém tiếp! | Renasci, bora cortar! |
| `respawn[3]` | ″ | Nghỉ tí rồi tu lại! | Cochilei, bora lá! |
| `no_swords[0]` | khi hết kiếm | Hết kiếm rồi! | Sem espadas! |
| `no_swords[1]` | ″ | Tui tay không nè! | Tô de mão vazia! |
| `no_swords[2]` | ″ | Ai cho mượn cây kiếm... | Me empresta uma? |
| `no_swords[3]` | ″ | Thôi chạy đã! | Primeiro, correr! |
| `bush[0]` | khi núp bụi cỏ | Suỵt... | Psiu... |
| `bush[1]` | ″ | Không có ai ở đây đâu. | Ninguém aqui. |
| `bush[2]` | ″ | Tui là cái bụi cỏ. | Eu sou uma moita. |
| `bush[3]` | ″ | Đừng nhìn, đừng nhìn... | Não olha, não olha... |
| `shocked[0]` | khi thấy người khác đột phá / chém đẹp; câu cuối = meme "tuyệt đối điện ảnh" | Hả??? | Hã??? |
| `shocked[1]` | ″ | Trời đất ơi! | Meu Deus do céu! |
| `shocked[2]` | ″ | Ghê vậy! | Caramba! |
| `shocked[3]` | ″ | Gì cơ?! | Como é que é?! |
| `shocked[4]` | ″ | Ông này ghê quá! | Esse cara é louco! |
| `shocked[5]` | ″ | Tuyệt đối điện ảnh! | Cinema absoluto! |



## Español (es-ES) — `es.json`

**Trạng thái: BẢN NHÁP (`_draft: true`) — cần người bản ngữ duyệt trước khi phát hành.**


### `tagline`
Khẩu hiệu dưới logo trang chủ. Gốc: chơi chữ "vòng vòng" lặp đôi, có vần.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `tagline` |  | Las espadas dan vueltas, y los taoístas, vueltas y vueltas. |

### `charNick`
Biệt danh ở thẻ chọn nhân vật, dạng "A Hớn — *biệt danh*". ≤ 18 ký tự.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `charNick.a_hon` | A Hớn Bánh Bao | el del Bollo |
| `charNick.ba_tam_loa` | Bà Tám Loa | Tía Megáfono |
| `charNick.ong_ba_vong` | Ông Ba Võng | Abuelo Hamaca |
| `charNick.dai_ca_dep` | Đại Ca Dép Tổ Ong | Jefe de la Chancla |
| `charNick.map_tra_sua` | Thằng Mập Trà Sữa | el Gordo del Té |
| `charNick.co_ba_ganh` | Cô Ba Gánh Hàng | la Vendedora |
| `charNick.su_phu_mi` | Sư Phụ Mì Gói | Maestro Fideo |
| `charNick.be_ti` | Bé Tí Cơ Bắp | Mini Musculitos |
| `charNick.shipper` | Anh Shipper Tia Chớp | Repartidor Rayo |
| `charNick.cu_rua` | Cụ Rùa Thiền | Tortuga Zen |

### `skillLines`
Câu thoại nhân vật hét khi dùng kỹ năng (bong bóng chữ trên đầu). Nên ngắn, đúng tính cách.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `skillLines.a_hon` | Ăn miếng bánh bao, kiếm về như mới! | ¡Un mordisco y mis espadas como nuevas! |
| `skillLines.ba_tam_loa` | A lô a lô, tránh ra cho bà đi! | ¡PROBANDO, PROBANDO! ¡Apártense, que paso yo! |
| `skillLines.ong_ba_vong` | À ơi… ngủ đi con… | A la nanita nana… a dormir… |
| `skillLines.dai_ca_dep` | Tụi bay đâu, lên đây! | ¡Chavales! ¡Venid aquí! |
| `skillLines.map_tra_sua` | Nếm thử trân châu nhà tao! | ¡Prueba las perlas de la casa! |
| `skillLines.co_ba_ganh` | Hàng về! Tránh ra cho cô bày! | ¡Mercancía fresca! ¡Hagan sitio! |
| `skillLines.su_phu_mi` | Sùn sụt… kiếm mày ngon đấy! | *sluuurp*… ¡qué rica tu espada! |
| `skillLines.be_ti` | Gồng! Không đứa nào đẩy nổi! | ¡MÚSCULO! ¡Nadie me mueve! |
| `skillLines.shipper` | Đơn gấp, xin phép đi xuyên! | ¡Pedido urgente, paso por la pared! |
| `skillLines.cu_rua` | Thiền… các ngươi cứ chém đi. | Ommm… vosotros cortad, tranquilos. |

### `titles`
Danh hiệu Top 1→Top 10 của phòng, uy giảm dần. ≤ 32 ký tự. Top 1 = sư khoe khoang "chưa ngán ai"; Top 10 = sống được nhờ núp.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `titles[0]` | Bần tăng chưa ngán ai bao giờ | Este Monje No Teme a Nadie |
| `titles[1]` | Ngon thì nhào vô | Ven si te atreves |
| `titles[2]` | Chỗ này nhiều máu nè | Aquí hay mucha vida, eh |
| `titles[3]` | Chưa gặp đối thủ xứng tầm | Buscando un rival digno |
| `titles[4]` | Đi nhẹ nói khẽ, chém hơi mạnh | Habla bajito, pega fuerte |
| `titles[5]` | Cũng có số má đấy chứ | Tiene su fama, ojo |
| `titles[6]` | Tạm gọi là biết múa kiếm | Sabe menear la espada |
| `titles[7]` | Đang cố gắng từng ngày | Mejorando cada día |
| `titles[8]` | Thấy ai cũng chào trước | Saluda a todos primero |
| `titles[9]` | Nấp giỏi nên còn sống | Vivo de tanto esconderse |

### `deathJokes`
Câu đùa ngẫu nhiên ở màn hình Tổng kết khi chết. Không giới hạn cứng nhưng nên ≤ 1 dòng (~60 ký tự).

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `deathJokes[0]` | Bạn đã được tiễn về quê nuôi cá. | Te han mandado de vuelta al pueblo a criar gallinas. |
| `deathJokes[1]` | Kiếm gãy, người bay, mộng tu tiên tạm hoãn. | Espada rota, cuerpo volando, inmortalidad aplazada. |
| `deathJokes[2]` | Tu tiên không thành, về bán bánh bao tiếp. | No has ascendido. Vuelve a vender churros. |
| `deathJokes[3]` | Ăn một nhát kiếm, ngộ ra chân lý: chạy nhanh hơn. | Te comiste una espada y aprendiste el Dao: corre más. |
| `deathJokes[4]` | Đạo tâm vỡ vụn như bánh tráng. | Tu corazón del Dao se rompió como un polvorón. |
| `deathJokes[5]` | Thôi thì kiếp sau làm quái cho đỡ khổ. | En la próxima vida, renace de monstruo. Menos estrés. |
| `deathJokes[6]` | Tu được ba hôm, nằm viện ba tháng. | Tres días cultivando, tres meses en el hospital. |
| `deathJokes[7]` | Cảnh giới cao nhất bạn đạt được: cảnh giới nằm. | Reino más alto alcanzado: Horizontal. |
| `deathJokes[8]` | Người ta phi thăng, bạn phi ra khỏi bản đồ. | Otros ascienden al cielo. Tú ascendiste fuera del mapa. |
| `deathJokes[9]` | Sư phụ dặn né, bạn đứng lại ngắm. | El maestro dijo esquiva. Tú te paraste a mirar el paisaje. |
| `deathJokes[10]` | Độ kiếp thất bại ngay tại tầng gửi xe. | Tribulación fallida. En el parking del súper. |
| `deathJokes[11]` | Tu vi của bạn: đang trong quá trình tích lũy lại từ đầu. | Tu cultivo se está volviendo a acumular desde cero. |
| `deathJokes[12]` | Kiếm người ta xoay, kiếm bạn xoay xuống đất. | Sus espadas giran. Las tuyas giran por el suelo. |
| `deathJokes[13]` | Linh khí trời đất không cứu nổi tốc độ phản xạ của bạn. | Ni todo el Qi del universo arregla tus reflejos. |
| `deathJokes[14]` | Chốt đơn: một vé về quê, không hoàn không huỷ. | Pedido confirmado: un billete a casa. Sin devoluciones. |
| `deathJokes[15]` | Bạn đi rồi, bản đồ thoáng hẳn ra. | Te has ido y el mapa se ve más amplio. |
| `deathJokes[16]` | Thọ nguyên còn dài, nhưng kiếm người ta dài hơn. | Tu vida era larga. Su espada, más. |
| `deathJokes[17]` | Có cái nịt, không có cái mạng. | Te quedaste sin nada. Ni con lo puesto. |
| `deathJokes[18]` | Đạo hữu nằm xuống, nhạc nền nổi lên. | Cae un taoísta. Empieza a sonar el violín triste. |
| `deathJokes[19]` | Hẹn gặp lại ở quán trà đá cuối thôn. | Nos vemos en el bar del final del pueblo. |
| `deathJokes[20]` | Thiên đạo nhắn: "Thử lại sau nhé." | El Dao Celestial dice: "Inténtalo más tarde." |
| `deathJokes[21]` | Bế quan mãi mãi, không ai làm phiền nữa. | Meditación a puerta cerrada eterna. No molestar. |

### `feed.kill`
Kill feed: {a} chém chết {b}. Hiện góc màn hình, mọi người trong phòng thấy.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.kill[0]` | {a} vừa tiễn {b} về quê nuôi cá | {a} mandó a {b} de vuelta al pueblo a criar gallinas |
| `feed.kill[1]` | {a} tặng {b} một vé về quê, không hoàn không huỷ | {a} le regaló a {b} un billete solo de ida a casa |
| `feed.kill[2]` | {b} bị {a} chém bay cả giấc mộng tu tiên | {a} partió en dos el sueño inmortal de {b} |
| `feed.kill[3]` | {a} dọn dẹp {b} gọn gàng sạch sẽ | {a} dejó a {b} limpito y ordenado |
| `feed.kill[4]` | {b} nằm xuống, {a} phủi tay đi tiếp | {b} cae, {a} se sacude el polvo y sigue |
| `feed.kill[5]` | {a} vừa cho {b} bế quan vĩnh viễn | {a} mandó a {b} a meditar para siempre |

### `feed.realm_drop`
Kill feed: {a} chém {b} rớt xuống cảnh giới {realm}.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.realm_drop[0]` | {b} bị {a} chém rớt xuống {realm} | {a} tumbó a {b} hasta {realm} |
| `feed.realm_drop[1]` | {a} đạp {b} về lại {realm} | {a} devolvió a {b} de una patada a {realm} |
| `feed.realm_drop[2]` | {b} tu bao ngày, về {realm} trong một nhát của {a} | {b} sudó días para volver a {realm} de un tajo de {a} |
| `feed.realm_drop[3]` | {a} dạy {b} một bài: {realm} thôi nhé | {a} le dio una lección a {b}: {realm} y gracias |
| `feed.realm_drop[4]` | {b} rớt xuống {realm}, công sức bay theo gió | {b} bajó a {realm}, todo el esfuerzo se lo llevó el viento |

### `feed.breakthrough`
Kill feed: {a} đột phá lên {realm}. Câu 5 nhại meme "tuyệt đối điện ảnh" (absolute cinema).

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.breakthrough[0]` | {a} đột phá {realm}, cả phòng run rẩy | {a} avanzó a {realm}, toda la sala tiembla |
| `feed.breakthrough[1]` | {a} lên {realm} rồi, ai nhanh chân thì chạy | {a} llegó a {realm} — sálvese quien pueda |
| `feed.breakthrough[2]` | Linh khí dồn về một chỗ: {a} chạm {realm} | El Qi se concentra en un punto: {a} alcanza {realm} |
| `feed.breakthrough[3]` | {a} bước vào {realm}, trời đất đổi màu | {a} entra en {realm}, el cielo cambia de color |
| `feed.breakthrough[4]` | {a} vừa lên {realm} — tuyệt đối điện ảnh | {a} acaba de llegar a {realm} — cine absoluto |

### `feed.gate_kill`
Kill feed: {b} vừa ra khỏi cổng dịch chuyển thì bị {a} chém. Câu 1 gốc: "đen thôi đỏ quên đi" = thành ngữ cờ bạc, ý "xui thì chịu".

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.gate_kill[0]` | {b} chui cổng ra liền bị {a} chém, đen thôi đỏ quên đi | {b} salió del portal directo a la espada de {a}. Mala suerte |
| `feed.gate_kill[1]` | {b} vừa ló đầu khỏi cổng đã ăn kiếm của {a} | {b} apenas asomó la cabeza del portal y se comió la espada de {a} |
| `feed.gate_kill[2]` | {a} ngồi canh cổng, {b} xui xẻo bước ra | {a} campeaba el portal y {b} salió justo delante |
| `feed.gate_kill[3]` | {b} dịch chuyển hết một vòng để tới chỗ {a} chết | {b} se teletransportó al otro lado del mundo solo para morir ante {a} |

### `feed.title_gain`
Tin loan khi ai đó lên Top 1 — PHẢI trích đúng nguyên văn titles[0].

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.title_gain` | {a} đã lên ngôi "Bần tăng chưa ngán ai bao giờ"! | ¡{a} se ha coronado "Este Monje No Teme a Nadie"! |

### `feed.title_lose`
Tin loan khi Top 1 bị hạ — chơi lại chữ của titles[0] ("giờ đã ngán").

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.title_lose` | {a} bị hạ bệ, bần tăng giờ đã ngán | {a} ha sido destronado — ahora el monje sí tiene miedo |

### `loadTips.10`
Mẹo màn hình tải có nhắc biệt danh nhân vật (Thằng Mập) — phải khớp charNick.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `loadTips[10]` | Thằng Mập phun trân châu kín màn hình. 5 giây đó bạn cũng không nhặt được gì. | El Gordo del Té te llena la pantalla de perlas. Esos 5 segundos tampoco puedes recoger nada. |

### `loadTips.11`
Mẹo màn hình tải nhắc Đại Ca Dép — phải khớp charNick.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `loadTips[11]` | Đại Ca Dép gọi hai đàn em mang nửa số kiếm của hắn. Đừng xông vào giữa. | El Jefe de la Chancla llama a dos secuaces con la mitad de sus espadas. No te metas en medio. |

### `emotes.*`
Emote: bong bóng chữ trên đầu nhân vật (người và bot), ≤ 22 ký tự.

| Khoá | Ngữ cảnh | Gốc tiếng Việt | Bản nháp |
|---|---|---|---|
| `sword_lost[0]` | khi mất kiếm | Kiếm của taaa! | ¡Mi espadaaa! |
| `sword_lost[1]` | ″ | Trả kiếm đây! | ¡Devuélvemela! |
| `sword_lost[2]` | ″ | Kiếm mới mua mà! | ¡Si era nueva! |
| `sword_lost[3]` | ″ | Rụng mất một cây rồi! | ¡Se me cayó una! |
| `sword_lost[4]` | ″ | Ai nhặt là mất dạy đó! | ¡El que la coja, caca! |
| `shield_break[0]` | khi khiên Kim Cang vỡ | Khiên ơi! | ¡Mi escudo! |
| `shield_break[1]` | ″ | Vỡ rồi! | ¡Se rompió! |
| `shield_break[2]` | ″ | Khiên giấy à trời! | ¿Escudo de papel? |
| `shield_break[3]` | ″ | Mới đeo được 3 giây! | ¡Duró 3 segundos! |
| `shield_break[4]` | ″ | Bảo hành đâu?! | ¿Y la garantía? |
| `chest[0]` | khi mở Rương Báu ("Ét ô ét" = SOS kiểu Việt; "Đổi đời rồi bà con") | Trúng mánh! | ¡Premio! |
| `chest[1]` | ″ | Giàu rồi! | ¡Soy rico! |
| `chest[2]` | ″ | Đổi đời rồi bà con! | ¡Me tocó el Gordo! |
| `chest[3]` | ″ | Hôm nay số đỏ! | ¡Día de suerte! |
| `chest[4]` | ″ | Toàn đồ xịn! | ¡Todo de calidad! |
| `chest[5]` | ″ | Ét ô ét, giàu quá! | ¡Olé, qué riqueza! |
| `revealed[0]` | khi bị lộ khỏi Ẩn Thân | Lộ rồi! | ¡Me pillaron! |
| `revealed[1]` | ″ | Ối! | ¡Uy! |
| `revealed[2]` | ″ | Sao thấy tui hay vậy? | ¿Cómo me ves? |
| `revealed[3]` | ″ | Chưa núp xong mà! | ¡Aún no me escondí! |
| `revealed[4]` | ″ | Thôi xong, lộ hàng! | ¡Ups, cazado! |
| `run[0]` | khi đang chạy (Thiêu Đốt) | AAAAA! | ¡AAAAA! |
| `run[1]` | ″ | Tránh ra! | ¡Apartad! |
| `run[2]` | ″ | Xe tui không phanh! | ¡Sin frenos! |
| `run[3]` | ″ | Dô dô dô! | ¡Fiuuum! |
| `run[4]` | ″ | Chân ơi cố lên! | ¡Piernas, no falléis! |
| `run[5]` | ″ | Nhanh nhanh nhanh! | ¡Vamos vamos! |
| `chase[0]` | khi đuổi người yếu hơn; câu cuối là tên game | Hehehe... | Jejeje... |
| `chase[1]` | ″ | Đứng lại đó! | ¡Quieto ahí! |
| `chase[2]` | ″ | Chạy đâu cho thoát! | ¡No te escapas! |
| `chase[3]` | ″ | Lại đây cưng! | ¡Ven aquí, guapo! |
| `chase[4]` | ″ | Đứng im một nhát thôi! | ¡Quieto, solo uno! |
| `chase[5]` | ″ | Ngon, con mồi kìa! | ¡Uh, presa fresca! |
| `chase[6]` | ″ | Đạo hữu, dừng bước! | ¡Taoísta, detente! |
| `flee[0]` | khi chạy trốn người mạnh hơn | Cứu tôi với! | ¡Socorro! |
| `flee[1]` | ″ | Mẹ ơi! | ¡Mamáaa! |
| `flee[2]` | ″ | Tui còn nhỏ lắm! | ¡Soy solo un bebé! |
| `flee[3]` | ″ | Nhầm người rồi! | ¡Te equivocas! |
| `flee[4]` | ″ | Tha cho đệ! | ¡Perdóname! |
| `flee[5]` | ″ | Đệ xin hàng! | ¡Me rindo! |
| `flee[6]` | ″ | Không đánh nhau, hoà bình! | ¡Paz y amor! |
| `hit[0]` | khi bị chém trúng ("Á đù" = văng tục nhẹ) | Ối giời! | ¡Ay! |
| `hit[1]` | ″ | Đau quá! | ¡Duele! |
| `hit[2]` | ″ | Á đù! | ¡Pero qué—! |
| `hit[3]` | ″ | Chơi bẩn! | ¡Juego sucio! |
| `hit[4]` | ″ | Hự! | ¡Uf! |
| `hit[5]` | ″ | Ai đánh lén đó?! | ¿Quién ha sido? |
| `hit[6]` | ″ | Nhẹ tay coi! | ¡Con cuidado! |
| `breakthrough[0]` | khi lên đại cảnh giới | Ta đã ngộ rồi!!! | ¡¡¡Me iluminé!!! |
| `breakthrough[1]` | ″ | Thiên địa chứng giám! | ¡Que el cielo lo vea! |
| `breakthrough[2]` | ″ | Lên cảnh giới rồi nha! | ¡Reino nuevo! |
| `breakthrough[3]` | ″ | Nhìn ta đây này! | ¡Miradme! |
| `breakthrough[4]` | ″ | Cảm tạ trời đất! | ¡Gracias, cielo! |
| `levelup[0]` | khi lên tầng nhỏ | Lên tầng! | ¡Subo de nivel! |
| `levelup[1]` | ″ | Thơm quá! | ¡Qué rico! |
| `levelup[2]` | ″ | Tiến bộ rồi nha! | ¡Voy mejorando! |
| `levelup[3]` | ″ | Thêm tí tu vi! | ¡Un poquito más! |
| `levelup[4]` | ″ | Mượt! | ¡Fino! |
| `realm_drop[0]` | khi bị rớt cảnh giới | Mất cảnh giới rồi! | ¡Perdí el reino! |
| `realm_drop[1]` | ″ | Công sức bao ngày... | Tanto esfuerzo... |
| `realm_drop[2]` | ″ | Tụt hạng đau quá! | ¡Descenso, ay! |
| `realm_drop[3]` | ″ | Làm lại từ đầu vậy! | ¡A empezar de cero! |
| `realm_drop[4]` | ″ | Trời không thương tui! | ¡El cielo me odia! |
| `teleport_in[0]` | khi vừa ra khỏi cổng | Đây là đâu??? | ¿¿¿Dónde estoy??? |
| `teleport_in[1]` | ″ | Ủa? Lạc rồi! | ¿Eh? ¡Me perdí! |
| `teleport_in[2]` | ″ | Cổng gì chóng mặt vậy! | ¡Qué portal más loco! |
| `teleport_in[3]` | ″ | Ai đưa tui tới đây? | ¿Quién me trajo aquí? |
| `teleport_in[4]` | ″ | Về nhầm nhà rồi! | ¡Casa equivocada! |
| `teleport_out[0]` | khi chui vào cổng | Tui biến đây! | ¡Me piro! |
| `teleport_out[1]` | ″ | Bái bai nha! | ¡Chao chao! |
| `teleport_out[2]` | ″ | Hẹn gặp bên kia! | ¡Nos vemos allá! |
| `teleport_out[3]` | ″ | Chui cổng cái đã! | ¡Al portal! |
| `death[0]` | khi chết | Thôi xong... | Pues nada... |
| `death[1]` | ″ | Tui đi trước nha! | ¡Me adelanto! |
| `death[2]` | ″ | Hẹn kiếp sau! | ¡Hasta otra vida! |
| `death[3]` | ″ | Tin được không trời! | ¡No me lo creo! |
| `respawn[0]` | khi bot hồi sinh | Tui lại về rồi! | ¡He vuelto! |
| `respawn[1]` | ″ | Chưa hết đâu nha! | ¡Esto no acaba! |
| `respawn[2]` | ″ | Hồi sinh, chém tiếp! | ¡Revivo, a cortar! |
| `respawn[3]` | ″ | Nghỉ tí rồi tu lại! | ¡Siestita y a seguir! |
| `no_swords[0]` | khi hết kiếm | Hết kiếm rồi! | ¡Sin espadas! |
| `no_swords[1]` | ″ | Tui tay không nè! | ¡A mano limpia! |
| `no_swords[2]` | ″ | Ai cho mượn cây kiếm... | ¿Alguien me presta? |
| `no_swords[3]` | ″ | Thôi chạy đã! | ¡Primero, correr! |
| `bush[0]` | khi núp bụi cỏ | Suỵt... | Chsss... |
| `bush[1]` | ″ | Không có ai ở đây đâu. | Aquí no hay nadie. |
| `bush[2]` | ″ | Tui là cái bụi cỏ. | Soy un arbusto. |
| `bush[3]` | ″ | Đừng nhìn, đừng nhìn... | No mires, no mires... |
| `shocked[0]` | khi thấy người khác đột phá / chém đẹp; câu cuối = meme "tuyệt đối điện ảnh" | Hả??? | ¿¿Qué?? |
| `shocked[1]` | ″ | Trời đất ơi! | ¡Madre mía! |
| `shocked[2]` | ″ | Ghê vậy! | ¡Hala! |
| `shocked[3]` | ″ | Gì cơ?! | ¿Cómo dices? |
| `shocked[4]` | ″ | Ông này ghê quá! | ¡Este tío está loco! |
| `shocked[5]` | ″ | Tuyệt đối điện ảnh! | ¡Cine absoluto! |



## Bahasa Indonesia (id-ID) — `id.json`

**Trạng thái: BẢN NHÁP (`_draft: true`) — cần người bản ngữ duyệt trước khi phát hành.**


### `tagline`
Khẩu hiệu dưới logo trang chủ. Gốc: chơi chữ "vòng vòng" lặp đôi, có vần.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `tagline` |  | Pedang muter-muter, Daois lari muter-muter. |

### `charNick`
Biệt danh ở thẻ chọn nhân vật, dạng "A Hớn — *biệt danh*". ≤ 18 ký tự.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `charNick.a_hon` | A Hớn Bánh Bao | si Tukang Bakpao |
| `charNick.ba_tam_loa` | Bà Tám Loa | Emak Toa |
| `charNick.ong_ba_vong` | Ông Ba Võng | Kakek Hammock |
| `charNick.dai_ca_dep` | Đại Ca Dép Tổ Ong | Bos Sandal Jepit |
| `charNick.map_tra_sua` | Thằng Mập Trà Sữa | si Gendut Boba |
| `charNick.co_ba_ganh` | Cô Ba Gánh Hàng | Mbak Pedagang |
| `charNick.su_phu_mi` | Sư Phụ Mì Gói | Suhu Mi Instan |
| `charNick.be_ti` | Bé Tí Cơ Bắp | Bocil Kekar |
| `charNick.shipper` | Anh Shipper Tia Chớp | Kurir Kilat |
| `charNick.cu_rua` | Cụ Rùa Thiền | Kakek Kura Zen |

### `skillLines`
Câu thoại nhân vật hét khi dùng kỹ năng (bong bóng chữ trên đầu). Nên ngắn, đúng tính cách.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `skillLines.a_hon` | Ăn miếng bánh bao, kiếm về như mới! | Sekali gigit, pedang balik kayak baru! |
| `skillLines.ba_tam_loa` | A lô a lô, tránh ra cho bà đi! | TES, TES! Minggir, emak mau lewat! |
| `skillLines.ong_ba_vong` | À ơi… ngủ đi con… | Nina bobo… oh nina bobo… |
| `skillLines.dai_ca_dep` | Tụi bay đâu, lên đây! | Anak-anak, sini semua! |
| `skillLines.map_tra_sua` | Nếm thử trân châu nhà tao! | Cobain boba spesial gue! |
| `skillLines.co_ba_ganh` | Hàng về! Tránh ra cho cô bày! | Barang baru! Minggir, mau gelar lapak! |
| `skillLines.su_phu_mi` | Sùn sụt… kiếm mày ngon đấy! | *sluuurp*… pedangmu enak juga! |
| `skillLines.be_ti` | Gồng! Không đứa nào đẩy nổi! | OTOT! Nggak ada yang bisa dorong! |
| `skillLines.shipper` | Đơn gấp, xin phép đi xuyên! | Paket urgent, permisi tembus tembok! |
| `skillLines.cu_rua` | Thiền… các ngươi cứ chém đi. | Ommm… silakan tebas saja. |

### `titles`
Danh hiệu Top 1→Top 10 của phòng, uy giảm dần. ≤ 32 ký tự. Top 1 = sư khoe khoang "chưa ngán ai"; Top 10 = sống được nhờ núp.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `titles[0]` | Bần tăng chưa ngán ai bao giờ | Biksu Ini Tak Takut Siapa Pun |
| `titles[1]` | Ngon thì nhào vô | Sini maju kalau berani |
| `titles[2]` | Chỗ này nhiều máu nè | Darahku masih banyak nih |
| `titles[3]` | Chưa gặp đối thủ xứng tầm | Belum ketemu lawan sepadan |
| `titles[4]` | Đi nhẹ nói khẽ, chém hơi mạnh | Bicara pelan, tebas keras |
| `titles[5]` | Cũng có số má đấy chứ | Lumayan punya nama |
| `titles[6]` | Tạm gọi là biết múa kiếm | Bisa lah ayun pedang |
| `titles[7]` | Đang cố gắng từng ngày | Berusaha tiap hari |
| `titles[8]` | Thấy ai cũng chào trước | Selalu nyapa duluan |
| `titles[9]` | Nấp giỏi nên còn sống | Hidup karena jago ngumpet |

### `deathJokes`
Câu đùa ngẫu nhiên ở màn hình Tổng kết khi chết. Không giới hạn cứng nhưng nên ≤ 1 dòng (~60 ký tự).

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `deathJokes[0]` | Bạn đã được tiễn về quê nuôi cá. | Kamu dipulangkan ke kampung buat ternak lele. |
| `deathJokes[1]` | Kiếm gãy, người bay, mộng tu tiên tạm hoãn. | Pedang patah, badan terbang, mimpi jadi dewa ditunda. |
| `deathJokes[2]` | Tu tiên không thành, về bán bánh bao tiếp. | Gagal jadi abadi. Balik jualan bakpao lagi. |
| `deathJokes[3]` | Ăn một nhát kiếm, ngộ ra chân lý: chạy nhanh hơn. | Kena satu tebasan, dapat pencerahan: larilah lebih cepat. |
| `deathJokes[4]` | Đạo tâm vỡ vụn như bánh tráng. | Hati Dao-mu remuk kayak kerupuk kena angin. |
| `deathJokes[5]` | Thôi thì kiếp sau làm quái cho đỡ khổ. | Kehidupan berikutnya jadi monster aja, lebih santai. |
| `deathJokes[6]` | Tu được ba hôm, nằm viện ba tháng. | Kultivasi tiga hari, opname tiga bulan. |
| `deathJokes[7]` | Cảnh giới cao nhất bạn đạt được: cảnh giới nằm. | Alam tertinggi yang dicapai: Alam Rebahan. |
| `deathJokes[8]` | Người ta phi thăng, bạn phi ra khỏi bản đồ. | Orang lain naik ke langit. Kamu naik keluar peta. |
| `deathJokes[9]` | Sư phụ dặn né, bạn đứng lại ngắm. | Guru bilang menghindar. Kamu malah berhenti foto-foto. |
| `deathJokes[10]` | Độ kiếp thất bại ngay tại tầng gửi xe. | Gagal lewati kesengsaraan. Di parkiran mal. |
| `deathJokes[11]` | Tu vi của bạn: đang trong quá trình tích lũy lại từ đầu. | Kultivasimu sedang dikumpulkan ulang dari nol. |
| `deathJokes[12]` | Kiếm người ta xoay, kiếm bạn xoay xuống đất. | Pedang orang berputar. Pedangmu berputar ke tanah. |
| `deathJokes[13]` | Linh khí trời đất không cứu nổi tốc độ phản xạ của bạn. | Seluruh Qi di dunia nggak bisa nolong refleksmu. |
| `deathJokes[14]` | Chốt đơn: một vé về quê, không hoàn không huỷ. | Pesanan dikonfirmasi: satu tiket pulang. Tanpa refund. |
| `deathJokes[15]` | Bạn đi rồi, bản đồ thoáng hẳn ra. | Kamu pergi, petanya jadi lega banget. |
| `deathJokes[16]` | Thọ nguyên còn dài, nhưng kiếm người ta dài hơn. | Umurmu panjang. Pedang dia lebih panjang. |
| `deathJokes[17]` | Có cái nịt, không có cái mạng. | Nggak dapat apa-apa. Nyawa pun nggak. |
| `deathJokes[18]` | Đạo hữu nằm xuống, nhạc nền nổi lên. | Seorang Daois tumbang. Lagu galau mulai diputar. |
| `deathJokes[19]` | Hẹn gặp lại ở quán trà đá cuối thôn. | Sampai jumpa di warkop ujung kampung. |
| `deathJokes[20]` | Thiên đạo nhắn: "Thử lại sau nhé." | Langit berpesan: "Silakan coba lagi nanti." |
| `deathJokes[21]` | Bế quan mãi mãi, không ai làm phiền nữa. | Bertapa selamanya. Jangan diganggu. |

### `feed.kill`
Kill feed: {a} chém chết {b}. Hiện góc màn hình, mọi người trong phòng thấy.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.kill[0]` | {a} vừa tiễn {b} về quê nuôi cá | {a} memulangkan {b} ke kampung buat ternak lele |
| `feed.kill[1]` | {a} tặng {b} một vé về quê, không hoàn không huỷ | {a} kasih {b} tiket pulang sekali jalan, tanpa refund |
| `feed.kill[2]` | {b} bị {a} chém bay cả giấc mộng tu tiên | {a} menebas mimpi kultivasi {b} jadi dua |
| `feed.kill[3]` | {a} dọn dẹp {b} gọn gàng sạch sẽ | {a} membereskan {b} sampai bersih |
| `feed.kill[4]` | {b} nằm xuống, {a} phủi tay đi tiếp | {b} tumbang, {a} tepuk-tepuk tangan lalu lanjut |
| `feed.kill[5]` | {a} vừa cho {b} bế quan vĩnh viễn | {a} bikin {b} bertapa selamanya |

### `feed.realm_drop`
Kill feed: {a} chém {b} rớt xuống cảnh giới {realm}.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.realm_drop[0]` | {b} bị {a} chém rớt xuống {realm} | {a} menebas {b} sampai turun ke {realm} |
| `feed.realm_drop[1]` | {a} đạp {b} về lại {realm} | {a} menendang {b} balik ke {realm} |
| `feed.realm_drop[2]` | {b} tu bao ngày, về {realm} trong một nhát của {a} | {b} kultivasi berhari-hari, balik ke {realm} cuma sekali tebas {a} |
| `feed.realm_drop[3]` | {a} dạy {b} một bài: {realm} thôi nhé | {a} kasih pelajaran ke {b}: {realm} aja ya |
| `feed.realm_drop[4]` | {b} rớt xuống {realm}, công sức bay theo gió | {b} turun ke {realm}, kerja kerasnya terbawa angin |

### `feed.breakthrough`
Kill feed: {a} đột phá lên {realm}. Câu 5 nhại meme "tuyệt đối điện ảnh" (absolute cinema).

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.breakthrough[0]` | {a} đột phá {realm}, cả phòng run rẩy | {a} menembus {realm}, seisi room gemetar |
| `feed.breakthrough[1]` | {a} lên {realm} rồi, ai nhanh chân thì chạy | {a} sudah {realm} — yang bisa lari, larilah |
| `feed.breakthrough[2]` | Linh khí dồn về một chỗ: {a} chạm {realm} | Qi berkumpul di satu titik: {a} mencapai {realm} |
| `feed.breakthrough[3]` | {a} bước vào {realm}, trời đất đổi màu | {a} memasuki {realm}, langit berubah warna |
| `feed.breakthrough[4]` | {a} vừa lên {realm} — tuyệt đối điện ảnh | {a} baru saja naik ke {realm} — sinematik parah |

### `feed.gate_kill`
Kill feed: {b} vừa ra khỏi cổng dịch chuyển thì bị {a} chém. Câu 1 gốc: "đen thôi đỏ quên đi" = thành ngữ cờ bạc, ý "xui thì chịu".

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.gate_kill[0]` | {b} chui cổng ra liền bị {a} chém, đen thôi đỏ quên đi | {b} keluar portal langsung kena pedang {a}. Apes |
| `feed.gate_kill[1]` | {b} vừa ló đầu khỏi cổng đã ăn kiếm của {a} | {b} baru nongol dari portal sudah dimakan pedang {a} |
| `feed.gate_kill[2]` | {a} ngồi canh cổng, {b} xui xẻo bước ra | {a} jaga portal, {b} sial keluar pas di depannya |
| `feed.gate_kill[3]` | {b} dịch chuyển hết một vòng để tới chỗ {a} chết | {b} teleport keliling dunia cuma buat mati di tangan {a} |

### `feed.title_gain`
Tin loan khi ai đó lên Top 1 — PHẢI trích đúng nguyên văn titles[0].

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.title_gain` | {a} đã lên ngôi "Bần tăng chưa ngán ai bao giờ"! | {a} kini menyandang gelar "Biksu Ini Tak Takut Siapa Pun"! |

### `feed.title_lose`
Tin loan khi Top 1 bị hạ — chơi lại chữ của titles[0] ("giờ đã ngán").

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.title_lose` | {a} bị hạ bệ, bần tăng giờ đã ngán | {a} dilengserkan — biksunya sekarang takut |

### `loadTips.10`
Mẹo màn hình tải có nhắc biệt danh nhân vật (Thằng Mập) — phải khớp charNick.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `loadTips[10]` | Thằng Mập phun trân châu kín màn hình. 5 giây đó bạn cũng không nhặt được gì. | Si Gendut Boba menutupi layarmu dengan boba. Selama 5 detik itu kamu juga nggak bisa mungut apa-apa. |

### `loadTips.11`
Mẹo màn hình tải nhắc Đại Ca Dép — phải khớp charNick.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `loadTips[11]` | Đại Ca Dép gọi hai đàn em mang nửa số kiếm của hắn. Đừng xông vào giữa. | Bos Sandal Jepit memanggil dua anak buah yang bawa setengah pedangnya. Jangan nyelonong ke tengah. |

### `emotes.*`
Emote: bong bóng chữ trên đầu nhân vật (người và bot), ≤ 22 ký tự.

| Khoá | Ngữ cảnh | Gốc tiếng Việt | Bản nháp |
|---|---|---|---|
| `sword_lost[0]` | khi mất kiếm | Kiếm của taaa! | Pedangkuuu! |
| `sword_lost[1]` | ″ | Trả kiếm đây! | Balikin pedangku! |
| `sword_lost[2]` | ″ | Kiếm mới mua mà! | Baru beli itu! |
| `sword_lost[3]` | ″ | Rụng mất một cây rồi! | Copot satu! |
| `sword_lost[4]` | ″ | Ai nhặt là mất dạy đó! | Yang mungut, dosa! |
| `shield_break[0]` | khi khiên Kim Cang vỡ | Khiên ơi! | Perisaiku! |
| `shield_break[1]` | ″ | Vỡ rồi! | Pecah! |
| `shield_break[2]` | ″ | Khiên giấy à trời! | Perisai kertas?! |
| `shield_break[3]` | ″ | Mới đeo được 3 giây! | Baru 3 detik! |
| `shield_break[4]` | ″ | Bảo hành đâu?! | Garansinya mana?! |
| `chest[0]` | khi mở Rương Báu ("Ét ô ét" = SOS kiểu Việt; "Đổi đời rồi bà con") | Trúng mánh! | Jackpot! |
| `chest[1]` | ″ | Giàu rồi! | Kaya mendadak! |
| `chest[2]` | ″ | Đổi đời rồi bà con! | Sultan, bestie! |
| `chest[3]` | ″ | Hôm nay số đỏ! | Lagi hoki! |
| `chest[4]` | ″ | Toàn đồ xịn! | Isinya mantap! |
| `chest[5]` | ″ | Ét ô ét, giàu quá! | Cuan cuan cuan! |
| `revealed[0]` | khi bị lộ khỏi Ẩn Thân | Lộ rồi! | Ketahuan! |
| `revealed[1]` | ″ | Ối! | Waduh! |
| `revealed[2]` | ″ | Sao thấy tui hay vậy? | Kok bisa lihat aku? |
| `revealed[3]` | ″ | Chưa núp xong mà! | Belum selesai ngumpet! |
| `revealed[4]` | ″ | Thôi xong, lộ hàng! | Yah, ketahuan! |
| `run[0]` | khi đang chạy (Thiêu Đốt) | AAAAA! | AAAAA! |
| `run[1]` | ″ | Tránh ra! | Minggir! |
| `run[2]` | ″ | Xe tui không phanh! | Rem blong! |
| `run[3]` | ″ | Dô dô dô! | Wusss! |
| `run[4]` | ″ | Chân ơi cố lên! | Kaki, kuatlah! |
| `run[5]` | ″ | Nhanh nhanh nhanh! | Cepet cepet cepet! |
| `chase[0]` | khi đuổi người yếu hơn; câu cuối là tên game | Hehehe... | Hehehe... |
| `chase[1]` | ″ | Đứng lại đó! | Berhenti di situ! |
| `chase[2]` | ″ | Chạy đâu cho thoát! | Mau lari ke mana? |
| `chase[3]` | ″ | Lại đây cưng! | Sini, sayang! |
| `chase[4]` | ″ | Đứng im một nhát thôi! | Diam, sekali aja! |
| `chase[5]` | ″ | Ngon, con mồi kìa! | Wah, mangsa segar! |
| `chase[6]` | ″ | Đạo hữu, dừng bước! | Daois, berhenti! |
| `flee[0]` | khi chạy trốn người mạnh hơn | Cứu tôi với! | Tolong! |
| `flee[1]` | ″ | Mẹ ơi! | Mamaaa! |
| `flee[2]` | ″ | Tui còn nhỏ lắm! | Aku masih kecil! |
| `flee[3]` | ″ | Nhầm người rồi! | Salah orang! |
| `flee[4]` | ″ | Tha cho đệ! | Ampun, bang! |
| `flee[5]` | ″ | Đệ xin hàng! | Aku nyerah! |
| `flee[6]` | ″ | Không đánh nhau, hoà bình! | Damai aja, damai! |
| `hit[0]` | khi bị chém trúng ("Á đù" = văng tục nhẹ) | Ối giời! | Aduh! |
| `hit[1]` | ″ | Đau quá! | Sakit! |
| `hit[2]` | ″ | Á đù! | Anjir! |
| `hit[3]` | ″ | Chơi bẩn! | Curang! |
| `hit[4]` | ″ | Hự! | Ugh! |
| `hit[5]` | ″ | Ai đánh lén đó?! | Siapa yang nyerang?! |
| `hit[6]` | ″ | Nhẹ tay coi! | Pelan-pelan dong! |
| `breakthrough[0]` | khi lên đại cảnh giới | Ta đã ngộ rồi!!! | Aku tercerahkan!!! |
| `breakthrough[1]` | ″ | Thiên địa chứng giám! | Langit jadi saksi! |
| `breakthrough[2]` | ″ | Lên cảnh giới rồi nha! | Naik alam nih! |
| `breakthrough[3]` | ″ | Nhìn ta đây này! | Lihat aku! |
| `breakthrough[4]` | ″ | Cảm tạ trời đất! | Terima kasih, langit! |
| `levelup[0]` | khi lên tầng nhỏ | Lên tầng! | Naik tingkat! |
| `levelup[1]` | ″ | Thơm quá! | Mantap! |
| `levelup[2]` | ″ | Tiến bộ rồi nha! | Makin kuat! |
| `levelup[3]` | ″ | Thêm tí tu vi! | Tambah dikit! |
| `levelup[4]` | ″ | Mượt! | Mulus! |
| `realm_drop[0]` | khi bị rớt cảnh giới | Mất cảnh giới rồi! | Alamku turun! |
| `realm_drop[1]` | ″ | Công sức bao ngày... | Kerja keras... |
| `realm_drop[2]` | ″ | Tụt hạng đau quá! | Turun kasta! |
| `realm_drop[3]` | ″ | Làm lại từ đầu vậy! | Mulai dari nol! |
| `realm_drop[4]` | ″ | Trời không thương tui! | Langit benci aku! |
| `teleport_in[0]` | khi vừa ra khỏi cổng | Đây là đâu??? | Ini di mana??? |
| `teleport_in[1]` | ″ | Ủa? Lạc rồi! | Lho? Nyasar! |
| `teleport_in[2]` | ″ | Cổng gì chóng mặt vậy! | Portal bikin pusing! |
| `teleport_in[3]` | ″ | Ai đưa tui tới đây? | Kok aku di sini? |
| `teleport_in[4]` | ″ | Về nhầm nhà rồi! | Salah rumah! |
| `teleport_out[0]` | khi chui vào cổng | Tui biến đây! | Cabut dulu! |
| `teleport_out[1]` | ″ | Bái bai nha! | Dadah! |
| `teleport_out[2]` | ″ | Hẹn gặp bên kia! | Ketemu di sana! |
| `teleport_out[3]` | ″ | Chui cổng cái đã! | Masuk portal dulu! |
| `death[0]` | khi chết | Thôi xong... | Tamat... |
| `death[1]` | ″ | Tui đi trước nha! | Duluan ya! |
| `death[2]` | ″ | Hẹn kiếp sau! | Sampai jumpa! |
| `death[3]` | ″ | Tin được không trời! | Nggak percaya! |
| `respawn[0]` | khi bot hồi sinh | Tui lại về rồi! | Aku kembali! |
| `respawn[1]` | ″ | Chưa hết đâu nha! | Belum selesai! |
| `respawn[2]` | ″ | Hồi sinh, chém tiếp! | Hidup lagi, tebas! |
| `respawn[3]` | ″ | Nghỉ tí rồi tu lại! | Rehat bentar, lanjut! |
| `no_swords[0]` | khi hết kiếm | Hết kiếm rồi! | Pedang habis! |
| `no_swords[1]` | ″ | Tui tay không nè! | Tangan kosong nih! |
| `no_swords[2]` | ″ | Ai cho mượn cây kiếm... | Pinjam pedang dong... |
| `no_swords[3]` | ″ | Thôi chạy đã! | Lari dulu! |
| `bush[0]` | khi núp bụi cỏ | Suỵt... | Ssst... |
| `bush[1]` | ″ | Không có ai ở đây đâu. | Nggak ada orang. |
| `bush[2]` | ″ | Tui là cái bụi cỏ. | Aku semak-semak. |
| `bush[3]` | ″ | Đừng nhìn, đừng nhìn... | Jangan lihat... |
| `shocked[0]` | khi thấy người khác đột phá / chém đẹp; câu cuối = meme "tuyệt đối điện ảnh" | Hả??? | Hah??? |
| `shocked[1]` | ″ | Trời đất ơi! | Ya ampun! |
| `shocked[2]` | ″ | Ghê vậy! | Gila! |
| `shocked[3]` | ″ | Gì cơ?! | Apa?! |
| `shocked[4]` | ″ | Ông này ghê quá! | Orang ini gila! |
| `shocked[5]` | ″ | Tuyệt đối điện ảnh! | Sinematik parah! |



## ไทย (th-TH) — `th.json`

**Trạng thái: BẢN NHÁP (`_draft: true`) — cần người bản ngữ duyệt trước khi phát hành.**


### `tagline`
Khẩu hiệu dưới logo trang chủ. Gốc: chơi chữ "vòng vòng" lặp đôi, có vần.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `tagline` |  | กระบี่หมุนติ้ว ๆ สหายเต๋าก็วิ่งหนีติ้ว ๆ |

### `charNick`
Biệt danh ở thẻ chọn nhân vật, dạng "A Hớn — *biệt danh*". ≤ 18 ký tự.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `charNick.a_hon` | A Hớn Bánh Bao | หนุ่มซาลาเปา |
| `charNick.ba_tam_loa` | Bà Tám Loa | ป้าโทรโข่ง |
| `charNick.ong_ba_vong` | Ông Ba Võng | ลุงเปลญวน |
| `charNick.dai_ca_dep` | Đại Ca Dép Tổ Ong | ลูกพี่รองเท้าแตะ |
| `charNick.map_tra_sua` | Thằng Mập Trà Sữa | อ้วนชานม |
| `charNick.co_ba_ganh` | Cô Ba Gánh Hàng | แม่ค้าหาบเร่ |
| `charNick.su_phu_mi` | Sư Phụ Mì Gói | อาจารย์มาม่า |
| `charNick.be_ti` | Bé Tí Cơ Bắp | จิ๋วกล้ามโต |
| `charNick.shipper` | Anh Shipper Tia Chớp | ไรเดอร์สายฟ้า |
| `charNick.cu_rua` | Cụ Rùa Thiền | ปู่เต่าเซน |

### `skillLines`
Câu thoại nhân vật hét khi dùng kỹ năng (bong bóng chữ trên đầu). Nên ngắn, đúng tính cách.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `skillLines.a_hon` | Ăn miếng bánh bao, kiếm về như mới! | งับซาลาเปาคำเดียว กระบี่กลับมาเหมือนใหม่! |
| `skillLines.ba_tam_loa` | A lô a lô, tránh ra cho bà đi! | ฮัลโหล ๆ หลีกไป ป้าจะผ่าน! |
| `skillLines.ong_ba_vong` | À ơi… ngủ đi con… | นอนเสียเถิด… หลับไปเถิดลูก… |
| `skillLines.dai_ca_dep` | Tụi bay đâu, lên đây! | ไอ้พวกนั้นอยู่ไหน มานี่! |
| `skillLines.map_tra_sua` | Nếm thử trân châu nhà tao! | ลองไข่มุกสูตรเด็ดของข้าซะ! |
| `skillLines.co_ba_ganh` | Hàng về! Tránh ra cho cô bày! | ของมาแล้ว! หลีกหน่อย ป้าจะตั้งแผง! |
| `skillLines.su_phu_mi` | Sùn sụt… kiếm mày ngon đấy! | ซู้ดดด… กระบี่เจ้าอร่อยดีนะ! |
| `skillLines.be_ti` | Gồng! Không đứa nào đẩy nổi! | เบ่ง! ไม่มีใครผลักข้าได้! |
| `skillLines.shipper` | Đơn gấp, xin phép đi xuyên! | ของด่วน ขอทะลุหน่อยครับ! |
| `skillLines.cu_rua` | Thiền… các ngươi cứ chém đi. | สมาธิ… ฟันไปเถอะ ตามสบาย |

### `titles`
Danh hiệu Top 1→Top 10 của phòng, uy giảm dần. ≤ 32 ký tự. Top 1 = sư khoe khoang "chưa ngán ai"; Top 10 = sống được nhờ núp.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `titles[0]` | Bần tăng chưa ngán ai bao giờ | อาตมาไม่เคยกลัวใคร |
| `titles[1]` | Ngon thì nhào vô | แน่จริงเข้ามาเลย |
| `titles[2]` | Chỗ này nhiều máu nè | เลือดเยอะนะจะบอกให้ |
| `titles[3]` | Chưa gặp đối thủ xứng tầm | ยังไม่เจอคู่ต่อสู้ที่คู่ควร |
| `titles[4]` | Đi nhẹ nói khẽ, chém hơi mạnh | เดินเบา พูดเบา ฟันแรง |
| `titles[5]` | Cũng có số má đấy chứ | ก็พอมีชื่อเสียงอยู่นะ |
| `titles[6]` | Tạm gọi là biết múa kiếm | พอจะรำกระบี่เป็น |
| `titles[7]` | Đang cố gắng từng ngày | พยายามทุกวัน |
| `titles[8]` | Thấy ai cũng chào trước | เจอใครก็ไหว้ก่อน |
| `titles[9]` | Nấp giỏi nên còn sống | รอดเพราะซ่อนเก่ง |

### `deathJokes`
Câu đùa ngẫu nhiên ở màn hình Tổng kết khi chết. Không giới hạn cứng nhưng nên ≤ 1 dòng (~60 ký tự).

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `deathJokes[0]` | Bạn đã được tiễn về quê nuôi cá. | คุณถูกส่งกลับบ้านนอกไปเลี้ยงปลาแล้ว |
| `deathJokes[1]` | Kiếm gãy, người bay, mộng tu tiên tạm hoãn. | กระบี่หัก ตัวปลิว ความฝันเป็นเซียนพักไว้ก่อน |
| `deathJokes[2]` | Tu tiên không thành, về bán bánh bao tiếp. | บำเพ็ญไม่สำเร็จ กลับไปขายซาลาเปาต่อ |
| `deathJokes[3]` | Ăn một nhát kiếm, ngộ ra chân lý: chạy nhanh hơn. | โดนฟันไปหนึ่งที บรรลุสัจธรรม: วิ่งให้เร็วกว่านี้ |
| `deathJokes[4]` | Đạo tâm vỡ vụn như bánh tráng. | จิตเต๋าแตกละเอียดเหมือนข้าวเกรียบ |
| `deathJokes[5]` | Thôi thì kiếp sau làm quái cho đỡ khổ. | ชาติหน้าเกิดเป็นมอนสเตอร์ดีกว่า เครียดน้อยกว่า |
| `deathJokes[6]` | Tu được ba hôm, nằm viện ba tháng. | บำเพ็ญสามวัน นอนโรงพยาบาลสามเดือน |
| `deathJokes[7]` | Cảnh giới cao nhất bạn đạt được: cảnh giới nằm. | ขอบเขตสูงสุดที่ทำได้: ขอบเขตนอนราบ |
| `deathJokes[8]` | Người ta phi thăng, bạn phi ra khỏi bản đồ. | คนอื่นเหาะขึ้นสวรรค์ คุณเหาะออกนอกแผนที่ |
| `deathJokes[9]` | Sư phụ dặn né, bạn đứng lại ngắm. | อาจารย์บอกให้หลบ คุณดันหยุดชมวิว |
| `deathJokes[10]` | Độ kiếp thất bại ngay tại tầng gửi xe. | ข้ามทัณฑ์สวรรค์ล้มเหลว ที่ลานจอดรถห้าง |
| `deathJokes[11]` | Tu vi của bạn: đang trong quá trình tích lũy lại từ đầu. | พลังฝึกของคุณ: กำลังสะสมใหม่ตั้งแต่ศูนย์ |
| `deathJokes[12]` | Kiếm người ta xoay, kiếm bạn xoay xuống đất. | กระบี่คนอื่นหมุนกลางอากาศ กระบี่คุณหมุนลงพื้น |
| `deathJokes[13]` | Linh khí trời đất không cứu nổi tốc độ phản xạ của bạn. | ปราณทั้งฟ้าดินก็ช่วยความไวของคุณไม่ได้ |
| `deathJokes[14]` | Chốt đơn: một vé về quê, không hoàn không huỷ. | ยืนยันออเดอร์: ตั๋วกลับบ้านหนึ่งใบ ไม่คืนเงิน |
| `deathJokes[15]` | Bạn đi rồi, bản đồ thoáng hẳn ra. | คุณไปแล้ว แผนที่โล่งขึ้นเยอะเลย |
| `deathJokes[16]` | Thọ nguyên còn dài, nhưng kiếm người ta dài hơn. | อายุขัยยังยาว แต่กระบี่เขายาวกว่า |
| `deathJokes[17]` | Có cái nịt, không có cái mạng. | ไม่เหลืออะไรเลย แม้แต่ลมหายใจ |
| `deathJokes[18]` | Đạo hữu nằm xuống, nhạc nền nổi lên. | สหายเต๋าล้มลง เพลงเศร้าก็ดังขึ้น |
| `deathJokes[19]` | Hẹn gặp lại ở quán trà đá cuối thôn. | แล้วเจอกันที่ร้านชาท้ายหมู่บ้าน |
| `deathJokes[20]` | Thiên đạo nhắn: "Thử lại sau nhé." | ฟ้าฝากบอก: "โปรดลองใหม่ภายหลัง" |
| `deathJokes[21]` | Bế quan mãi mãi, không ai làm phiền nữa. | ปิดด่านบำเพ็ญตลอดกาล ไม่มีใครมากวนอีกแล้ว |

### `feed.kill`
Kill feed: {a} chém chết {b}. Hiện góc màn hình, mọi người trong phòng thấy.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.kill[0]` | {a} vừa tiễn {b} về quê nuôi cá | {a} ส่ง {b} กลับบ้านนอกไปเลี้ยงปลา |
| `feed.kill[1]` | {a} tặng {b} một vé về quê, không hoàn không huỷ | {a} มอบตั๋วกลับบ้านให้ {b} ไม่คืนเงิน |
| `feed.kill[2]` | {b} bị {a} chém bay cả giấc mộng tu tiên | {a} ฟันความฝันเป็นเซียนของ {b} ขาดสองท่อน |
| `feed.kill[3]` | {a} dọn dẹp {b} gọn gàng sạch sẽ | {a} เก็บกวาด {b} เรียบร้อยหมดจด |
| `feed.kill[4]` | {b} nằm xuống, {a} phủi tay đi tiếp | {b} ล้มลง {a} ปัดฝุ่นแล้วเดินต่อ |
| `feed.kill[5]` | {a} vừa cho {b} bế quan vĩnh viễn | {a} ส่ง {b} ไปปิดด่านบำเพ็ญตลอดกาล |

### `feed.realm_drop`
Kill feed: {a} chém {b} rớt xuống cảnh giới {realm}.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.realm_drop[0]` | {b} bị {a} chém rớt xuống {realm} | {a} ฟัน {b} ร่วงลงไป {realm} |
| `feed.realm_drop[1]` | {a} đạp {b} về lại {realm} | {a} ถีบ {b} กลับไป {realm} |
| `feed.realm_drop[2]` | {b} tu bao ngày, về {realm} trong một nhát của {a} | {b} บำเพ็ญมาหลายวัน กลับไป {realm} ในดาบเดียวของ {a} |
| `feed.realm_drop[3]` | {a} dạy {b} một bài: {realm} thôi nhé | {a} สอนบทเรียนให้ {b}: {realm} พอนะ |
| `feed.realm_drop[4]` | {b} rớt xuống {realm}, công sức bay theo gió | {b} ร่วงลง {realm} ความพยายามปลิวไปกับลม |

### `feed.breakthrough`
Kill feed: {a} đột phá lên {realm}. Câu 5 nhại meme "tuyệt đối điện ảnh" (absolute cinema).

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.breakthrough[0]` | {a} đột phá {realm}, cả phòng run rẩy | {a} ทะลวงสู่ {realm} ทั้งห้องสั่นสะท้าน |
| `feed.breakthrough[1]` | {a} lên {realm} rồi, ai nhanh chân thì chạy | {a} ขึ้น {realm} แล้ว ใครขาไวก็วิ่งเลย |
| `feed.breakthrough[2]` | Linh khí dồn về một chỗ: {a} chạm {realm} | ปราณรวมเป็นหนึ่ง: {a} แตะ {realm} |
| `feed.breakthrough[3]` | {a} bước vào {realm}, trời đất đổi màu | {a} ก้าวเข้าสู่ {realm} ฟ้าดินเปลี่ยนสี |
| `feed.breakthrough[4]` | {a} vừa lên {realm} — tuyệt đối điện ảnh | {a} เพิ่งขึ้น {realm} — ระดับภาพยนตร์ |

### `feed.gate_kill`
Kill feed: {b} vừa ra khỏi cổng dịch chuyển thì bị {a} chém. Câu 1 gốc: "đen thôi đỏ quên đi" = thành ngữ cờ bạc, ý "xui thì chịu".

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.gate_kill[0]` | {b} chui cổng ra liền bị {a} chém, đen thôi đỏ quên đi | {b} โผล่จากประตูมิติปุ๊บโดน {a} ฟันปั๊บ ดวงซวยจริง |
| `feed.gate_kill[1]` | {b} vừa ló đầu khỏi cổng đã ăn kiếm của {a} | {b} เพิ่งโผล่หัวออกจากประตูก็โดนกระบี่ {a} |
| `feed.gate_kill[2]` | {a} ngồi canh cổng, {b} xui xẻo bước ra | {a} นั่งเฝ้าประตู {b} ดวงซวยเดินออกมาพอดี |
| `feed.gate_kill[3]` | {b} dịch chuyển hết một vòng để tới chỗ {a} chết | {b} วาร์ปไปไกลครึ่งโลกเพื่อมาตายให้ {a} |

### `feed.title_gain`
Tin loan khi ai đó lên Top 1 — PHẢI trích đúng nguyên văn titles[0].

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.title_gain` | {a} đã lên ngôi "Bần tăng chưa ngán ai bao giờ"! | {a} ขึ้นครองตำแหน่ง "อาตมาไม่เคยกลัวใคร" แล้ว! |

### `feed.title_lose`
Tin loan khi Top 1 bị hạ — chơi lại chữ của titles[0] ("giờ đã ngán").

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `feed.title_lose` | {a} bị hạ bệ, bần tăng giờ đã ngán | {a} ถูกโค่นแล้ว ตอนนี้อาตมากลัวแล้ว |

### `loadTips.10`
Mẹo màn hình tải có nhắc biệt danh nhân vật (Thằng Mập) — phải khớp charNick.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `loadTips[10]` | Thằng Mập phun trân châu kín màn hình. 5 giây đó bạn cũng không nhặt được gì. | อ้วนชานมพ่นไข่มุกเต็มจอ 5 วินาทีนั้นคุณเก็บอะไรไม่ได้ด้วย |

### `loadTips.11`
Mẹo màn hình tải nhắc Đại Ca Dép — phải khớp charNick.

| Khoá | Gốc tiếng Việt | Bản nháp |
|---|---|---|
| `loadTips[11]` | Đại Ca Dép gọi hai đàn em mang nửa số kiếm của hắn. Đừng xông vào giữa. | ลูกพี่รองเท้าแตะเรียกลูกน้องสองคนถือกระบี่ครึ่งหนึ่งของเขา อย่าบุกเข้าไปตรงกลาง |

### `emotes.*`
Emote: bong bóng chữ trên đầu nhân vật (người và bot), ≤ 22 ký tự.

| Khoá | Ngữ cảnh | Gốc tiếng Việt | Bản nháp |
|---|---|---|---|
| `sword_lost[0]` | khi mất kiếm | Kiếm của taaa! | กระบี่ข้าาา! |
| `sword_lost[1]` | ″ | Trả kiếm đây! | คืนกระบี่มา! |
| `sword_lost[2]` | ″ | Kiếm mới mua mà! | เพิ่งซื้อมานะ! |
| `sword_lost[3]` | ″ | Rụng mất một cây rồi! | หลุดไปเล่มนึง! |
| `sword_lost[4]` | ″ | Ai nhặt là mất dạy đó! | ใครเก็บไปใจร้ายนะ! |
| `shield_break[0]` | khi khiên Kim Cang vỡ | Khiên ơi! | โล่ข้า! |
| `shield_break[1]` | ″ | Vỡ rồi! | แตกแล้ว! |
| `shield_break[2]` | ″ | Khiên giấy à trời! | โล่กระดาษเหรอ! |
| `shield_break[3]` | ″ | Mới đeo được 3 giây! | ใส่ได้ 3 วิเอง! |
| `shield_break[4]` | ″ | Bảo hành đâu?! | ประกันอยู่ไหน?! |
| `chest[0]` | khi mở Rương Báu ("Ét ô ét" = SOS kiểu Việt; "Đổi đời rồi bà con") | Trúng mánh! | ถูกหวย! |
| `chest[1]` | ″ | Giàu rồi! | รวยแล้ว! |
| `chest[2]` | ″ | Đổi đời rồi bà con! | ชีวิตเปลี่ยนแล้ว! |
| `chest[3]` | ″ | Hôm nay số đỏ! | วันนี้ดวงดี! |
| `chest[4]` | ″ | Toàn đồ xịn! | ของดีทั้งนั้น! |
| `chest[5]` | ″ | Ét ô ét, giàu quá! | รวยเละ! |
| `revealed[0]` | khi bị lộ khỏi Ẩn Thân | Lộ rồi! | โดนจับได้! |
| `revealed[1]` | ″ | Ối! | อุ๊ย! |
| `revealed[2]` | ″ | Sao thấy tui hay vậy? | เห็นได้ไงเนี่ย? |
| `revealed[3]` | ″ | Chưa núp xong mà! | ยังซ่อนไม่เสร็จ! |
| `revealed[4]` | ″ | Thôi xong, lộ hàng! | จบเห่ โดนเห็น! |
| `run[0]` | khi đang chạy (Thiêu Đốt) | AAAAA! | อ๊ากกก! |
| `run[1]` | ″ | Tránh ra! | หลบไป! |
| `run[2]` | ″ | Xe tui không phanh! | เบรกแตก! |
| `run[3]` | ″ | Dô dô dô! | ซิ่งงง! |
| `run[4]` | ″ | Chân ơi cố lên! | ขาเอ๋ย สู้ ๆ! |
| `run[5]` | ″ | Nhanh nhanh nhanh! | เร็ว ๆ ๆ! |
| `chase[0]` | khi đuổi người yếu hơn; câu cuối là tên game | Hehehe... | เหะ ๆ ๆ... |
| `chase[1]` | ″ | Đứng lại đó! | หยุดตรงนั้น! |
| `chase[2]` | ″ | Chạy đâu cho thoát! | หนีไม่พ้นหรอก! |
| `chase[3]` | ″ | Lại đây cưng! | มานี่มาจ๊ะ! |
| `chase[4]` | ″ | Đứng im một nhát thôi! | อยู่นิ่ง ๆ ทีเดียว! |
| `chase[5]` | ″ | Ngon, con mồi kìa! | เหยื่อมาแล้ว! |
| `chase[6]` | ″ | Đạo hữu, dừng bước! | สหายเต๋า หยุดก่อน! |
| `flee[0]` | khi chạy trốn người mạnh hơn | Cứu tôi với! | ช่วยด้วย! |
| `flee[1]` | ″ | Mẹ ơi! | แม่จ๋า! |
| `flee[2]` | ″ | Tui còn nhỏ lắm! | ผมยังเด็กอยู่! |
| `flee[3]` | ″ | Nhầm người rồi! | จำคนผิดแล้ว! |
| `flee[4]` | ″ | Tha cho đệ! | ไว้ชีวิตข้าเถอะ! |
| `flee[5]` | ″ | Đệ xin hàng! | ยอมแล้ว ๆ! |
| `flee[6]` | ″ | Không đánh nhau, hoà bình! | รักสงบ ไม่สู้! |
| `hit[0]` | khi bị chém trúng ("Á đù" = văng tục nhẹ) | Ối giời! | โอ๊ย! |
| `hit[1]` | ″ | Đau quá! | เจ็บ! |
| `hit[2]` | ″ | Á đù! | เฮ้ย! |
| `hit[3]` | ″ | Chơi bẩn! | เล่นสกปรก! |
| `hit[4]` | ″ | Hự! | อึ๊ก! |
| `hit[5]` | ″ | Ai đánh lén đó?! | ใครลอบกัด?! |
| `hit[6]` | ″ | Nhẹ tay coi! | เบา ๆ หน่อย! |
| `breakthrough[0]` | khi lên đại cảnh giới | Ta đã ngộ rồi!!! | ข้าบรรลุแล้ว!!! |
| `breakthrough[1]` | ″ | Thiên địa chứng giám! | ฟ้าดินเป็นพยาน! |
| `breakthrough[2]` | ″ | Lên cảnh giới rồi nha! | ขึ้นขอบเขตแล้ว! |
| `breakthrough[3]` | ″ | Nhìn ta đây này! | ดูข้านี่! |
| `breakthrough[4]` | ″ | Cảm tạ trời đất! | ขอบคุณฟ้าดิน! |
| `levelup[0]` | khi lên tầng nhỏ | Lên tầng! | ขึ้นขั้น! |
| `levelup[1]` | ″ | Thơm quá! | หอมหวาน! |
| `levelup[2]` | ″ | Tiến bộ rồi nha! | เก่งขึ้นแล้ว! |
| `levelup[3]` | ″ | Thêm tí tu vi! | พลังเพิ่มอีกนิด! |
| `levelup[4]` | ″ | Mượt! | ลื่นไหล! |
| `realm_drop[0]` | khi bị rớt cảnh giới | Mất cảnh giới rồi! | ขอบเขตร่วง! |
| `realm_drop[1]` | ″ | Công sức bao ngày... | อุตส่าห์ฝึกมา... |
| `realm_drop[2]` | ″ | Tụt hạng đau quá! | ตกชั้น เจ็บ! |
| `realm_drop[3]` | ″ | Làm lại từ đầu vậy! | เริ่มใหม่ก็ได้! |
| `realm_drop[4]` | ″ | Trời không thương tui! | ฟ้าไม่เข้าข้าง! |
| `teleport_in[0]` | khi vừa ra khỏi cổng | Đây là đâu??? | ที่นี่ที่ไหน??? |
| `teleport_in[1]` | ″ | Ủa? Lạc rồi! | อ้าว? หลงแล้ว! |
| `teleport_in[2]` | ″ | Cổng gì chóng mặt vậy! | ประตูอะไรเวียนหัว! |
| `teleport_in[3]` | ″ | Ai đưa tui tới đây? | ใครพามาเนี่ย? |
| `teleport_in[4]` | ″ | Về nhầm nhà rồi! | เข้าบ้านผิด! |
| `teleport_out[0]` | khi chui vào cổng | Tui biến đây! | ไปละ! |
| `teleport_out[1]` | ″ | Bái bai nha! | บ๊ายบาย! |
| `teleport_out[2]` | ″ | Hẹn gặp bên kia! | เจอกันฝั่งโน้น! |
| `teleport_out[3]` | ″ | Chui cổng cái đã! | ลอดประตูก่อน! |
| `death[0]` | khi chết | Thôi xong... | จบกัน... |
| `death[1]` | ″ | Tui đi trước nha! | ไปก่อนนะ! |
| `death[2]` | ″ | Hẹn kiếp sau! | เจอกันชาติหน้า! |
| `death[3]` | ″ | Tin được không trời! | ไม่อยากเชื่อ! |
| `respawn[0]` | khi bot hồi sinh | Tui lại về rồi! | ข้ากลับมาแล้ว! |
| `respawn[1]` | ″ | Chưa hết đâu nha! | ยังไม่จบหรอก! |
| `respawn[2]` | ″ | Hồi sinh, chém tiếp! | เกิดใหม่ ฟันต่อ! |
| `respawn[3]` | ″ | Nghỉ tí rồi tu lại! | พักแป๊บ ฝึกต่อ! |
| `no_swords[0]` | khi hết kiếm | Hết kiếm rồi! | กระบี่หมด! |
| `no_swords[1]` | ″ | Tui tay không nè! | มือเปล่าเลย! |
| `no_swords[2]` | ″ | Ai cho mượn cây kiếm... | ใครให้ยืมกระบี่ที... |
| `no_swords[3]` | ″ | Thôi chạy đã! | วิ่งก่อน! |
| `bush[0]` | khi núp bụi cỏ | Suỵt... | ชู่ว... |
| `bush[1]` | ″ | Không có ai ở đây đâu. | ไม่มีใครอยู่ตรงนี้ |
| `bush[2]` | ″ | Tui là cái bụi cỏ. | ข้าคือพุ่มไม้ |
| `bush[3]` | ″ | Đừng nhìn, đừng nhìn... | อย่ามอง อย่ามอง... |
| `shocked[0]` | khi thấy người khác đột phá / chém đẹp; câu cuối = meme "tuyệt đối điện ảnh" | Hả??? | หา??? |
| `shocked[1]` | ″ | Trời đất ơi! | ตายแล้ว! |
| `shocked[2]` | ″ | Ghê vậy! | โหดอ่ะ! |
| `shocked[3]` | ″ | Gì cơ?! | อะไรนะ?! |
| `shocked[4]` | ″ | Ông này ghê quá! | คนนี้โหดจัด! |
| `shocked[5]` | ″ | Tuyệt đối điện ảnh! | ระดับภาพยนตร์! |



---

# Tên bot (v6.34 §9.4)

Agent soạn ngày 02/11/2026. Mỗi ngôn ngữ có 1 tệp tên bot `config/bot_names/<lang>.txt`
(vi = 1 000 tên cũ chép nguyên; en, pt, es, id, th mỗi tệp 440 tên; `intl.txt` 440 gamer tag
trung tính) và 1 tệp từ tục `config/bad_words/<lang>.txt` (intl dùng chung `en.txt`).
Người bản ngữ duyệt **cùng đợt** với bản dịch json của ngôn ngữ mình.

**Luật của tệp tên (CI sẽ kiểm):** ≥ 400 tên (intl ≥ 300), 2–16 ký tự (đếm code point —
tiếng Thái tính cả dấu), chỉ chữ Latin (có dấu), chữ Thái, chữ số và `_ . ~ ^ -` dấu cách;
không ★ ♥ ✿ emoji; không trùng; không chứa từ nào trong `bad_words/<lang>.txt` hoặc
`bad_words/en.txt` (so chuỗi con, không phân biệt hoa thường). Dòng `#` là chú thích.

## Người duyệt cần xem gì trong `bot_names/<lang>.txt`
1. **Có giống nickname người thật nước mình không?** Tên ghép tự động (tên + số/năm, tên +
   mã vùng như `_BR` `_MX` `.BDG` `_CNX`, ghép tính từ + con vật) — tên nào nghe như máy
   sinh thì xoá hoặc sửa. Giữ tỉ lệ pha trộn: tên thật, tên ở nhà, câu đùa, teencode chữ thường.
2. **Không có người nổi tiếng / streamer / thương hiệu / nhân vật có bản quyền** của nước mình
   mà agent không biết (vd ca sĩ, diễn viên, YouTuber địa phương trùng nickname).
3. **Không có nghĩa xấu, tục, phân biệt** theo tiếng lóng địa phương (agent không bắt hết được).
4. Câu đùa (vd `lag is real`, `não me mata`, `numpang lewat`, `ขอเกิดใหม่`) có tự nhiên không.
5. Sửa xong: số tên vẫn ≥ 400, không thêm ký tự ngoài danh sách trên.

## Người duyệt cần xem gì trong `bad_words/<lang>.txt`
- Thiếu từ tục/miệt thị phổ biến nào thì thêm (viết thường, 1 từ/dòng).
- **Tránh từ quá ngắn** vì lọc theo chuỗi con sẽ che nhầm tên thường (vd en `ass` dính
  `Cassie`, es `culo` dính `vehículo`, th `หี` dính `หีบ`, `สัด` dính `สัดส่วน`). Agent đã cố ý bỏ các từ đó.
- en đã cố ý **không** đưa `kike` (miệt thị) vì trùng biệt danh Tây Ban Nha rất phổ biến `Kike`
  (Enrique) — nếu server gộp mọi tệp từ tục để lọc tên người chơi thì sẽ che nhầm.

## ⚠ Tên agent không chắc
- **en**: `ok boomer jk` (có thể bị coi là khiếm nhã nhẹ), `free real estate` (meme — ổn không?).
- **pt**: ghép tính từ + con vật (`GatoDoido`, `QuatiSortudo`, `BodeRápido`) — đã chỉ dùng
  danh từ giống đực để khỏi lệch giống; xem có tự nhiên không. Đã bỏ `Babi` (trùng từ tục id `babi`).
- **es**: `Güero`, `Flaco`, `Gordi`, `Chula` (biệt danh thân mật kiểu Mexico/Tây Ban Nha — có bị
  coi là chê ngoại hình không?), `Kike`. Pha trộn từ lóng nhiều nước (`qué onda`, `bacán`,
  `mola mucho`, `re piola`) — chấp nhận được vì bot là "người nhiều nước". Đã bỏ `Begoña`
  (dính từ tục id `bego` nếu gộp bộ lọc).
- **id**: `Gendut`, `Ndut` (béo — thân mật hay chê?), `Bocil` (bé con — có bị coi là chế giễu?),
  `Nyai` (có thể mang nghĩa cổ/khó chịu), `sobat misqueen`, `jomblo abadi`.
- **th**: `Earth`, `New`, `Ohm` (nickname phổ biến nhưng trùng tên diễn viên nổi tiếng — agent đã
  bỏ `Mew`, `Bright`, `Win`, `Gulf`, `Tay`, `Off`, `Toon`, `Lisa`, `BamBam`); tiền tố `ไอ้`
  (`ไอ้ตาล`, `ไอ้ทองหยอด`) có thể nghe suồng sã; tên Latin kiểu `Isan Girl`, `Sao Chiang Mai`,
  `Bkk Boy` có tự nhiên không.
- **intl**: một số từ chung trùng tên nhân vật game (`Echo`, `Rogue`, `VoidWalker`, `Nova`) —
  agent coi là từ thông thường, đã bỏ `Apex`, `Jinx`, `Taz`, `FirstBlood`, `BladeRunner`.
- **vi.txt**: chép nguyên `config/bot_names.txt` (tệp gốc không có dòng chú thích đầu tệp).
