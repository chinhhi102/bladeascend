# Chọn giấy phép cho repo công khai — các phương án

> Soạn 02/11/2026 khi chuẩn bị mở công khai repo.
>
> **ĐÃ CHỐT 02/11/2026:** mã nguồn **MIT** (`LICENSE`), art **CC BY-NC-SA 4.0**
> (`assets/LICENSE`), mục "Tên và logo" đã thêm vào `README.md`. Không cần CLA (chỉ cần khi
> chọn AGPL). Phần dưới giữ lại làm bối cảnh.
> Chưa có tệp `LICENSE` thì theo luật, mọi quyền vẫn thuộc anh: người khác **xem** được mã
> nguồn trên GitHub nhưng **không được phép** chép, sửa, chạy lại. Mở công khai mà chưa có
> giấy phép là an toàn, chỉ là chưa ai đóng góp hợp pháp được.

Đề xuất chia làm **ba phần**, mỗi phần một giấy phép riêng:

| Phần | Gồm | Phương án |
|---|---|---|
| Mã nguồn | `server/`, `client/`, `tools/`, `proto/`, `config/`, `maps/` | **AGPL-3.0** *hoặc* **MIT** (anh chọn, mục 1) |
| Art | `assets/` (ảnh, hoạt ảnh, âm thanh do dự án tự sinh) | **CC BY-NC-SA 4.0** (mục 2) |
| Tên + logo | "Đạo Hữu Xin Dừng Bước", "BladeAscend", logo, tên miền | **Giữ bản quyền / thương hiệu** — bản fork phải đổi tên (mục 3) |

Font, thư viện, hạt VFX của người khác giữ nguyên giấy phép gốc của họ — xem
`assets/src/CREDITS.md`.

---

## 1. Mã nguồn: AGPL-3.0 hay MIT?

Nói ngắn: **AGPL** = "ai lấy code về chạy server cho người khác chơi thì phải công khai code
của họ". **MIT** = "ai lấy cũng được, làm gì cũng được, chỉ cần giữ dòng ghi tên anh".

### Phương án A — AGPL-3.0

Vì sao hợp với game nhiều người chơi: giấy phép GPL thường chỉ bắt công khai code khi
**phát hành phần mềm** (đưa tệp cho người khác). Server game thì không ai phát hành cả — người
ta chỉ chạy nó và cho người chơi kết nối vào. AGPL bịt đúng lỗ hổng đó: **cho người khác dùng
qua mạng cũng tính là phát hành**.

| Được | Mất |
|---|---|
| Ai đó lấy code dựng một server "lậu" có sửa đổi thì **phải công khai** phần đã sửa — anh xem được, học lại được những gì họ làm tốt hơn. | Công ty, nhà phát hành lớn thường **tránh** AGPL. Nếu sau này có người muốn mua bản quyền / hợp tác thương mại, họ sẽ đòi giấy phép riêng. |
| Không ai lấy code của anh làm game đóng (giấu code) để cạnh tranh được. | Người đóng góp nhỏ đôi khi ngại AGPL vì nghe "khó". |
| Anh vẫn giữ toàn quyền: là chủ sở hữu, anh luôn có thể **bán thêm giấy phép riêng** (dual-license) cho ai cần, *nếu* người đóng góp đồng ý chuyển quyền (xem lưu ý dưới). | Phải ghi rõ trong game có link tới mã nguồn (AGPL §13) — ví dụ một dòng ở chân trang. |

### Phương án B — MIT

| Được | Mất |
|---|---|
| Đơn giản nhất, ai cũng hiểu, nhiều người sẵn lòng đóng góp nhất. | Ai cũng có thể lấy code, sửa, dựng server riêng, **giấu toàn bộ phần sửa**, thậm chí bán — chỉ cần giữ dòng ghi tên anh. |
| Hợp nếu mục tiêu là *chia sẻ kiến thức* (người khác học cách làm game .io). | Một bản sao có tiền quảng cáo mạnh hơn có thể vượt bản gốc mà anh không làm gì được về mặt mã nguồn (tên + logo vẫn được bảo vệ ở mục 3). |
| Không cần thêm gì vào game. | Không lấy lại được phần cải tiến người khác làm trên code của anh. |

### Gợi ý cách nghĩ

- Anh muốn **game này là của anh**, server chính thức là nơi duy nhất người ta chơi, và lo
  bản sao → nghiêng về **AGPL-3.0**.
- Anh muốn **chia sẻ, càng nhiều người dùng lại càng tốt**, không quan tâm bản sao → **MIT**.
- Trong cả hai trường hợp, tên và logo (mục 3) mới là thứ thật sự ngăn người khác "giả danh"
  game của anh.

**Lưu ý về người đóng góp:** khi có người gửi PR, phần code của họ thuộc về họ nhưng được cấp
phép theo giấy phép của repo. Nếu anh nghĩ tới chuyện sau này bán giấy phép riêng (với AGPL),
nên thêm một dòng vào `CONTRIBUTING.md`: *"Gửi PR là đồng ý cấp cho chủ dự án quyền cấp phép
lại phần đóng góp theo giấy phép khác"* (gọi là CLA đơn giản). Việc này cũng **chưa làm** —
chờ anh quyết.

Khi chọn xong: tạo tệp `LICENSE` ở gốc repo bằng văn bản chuẩn (GitHub → *Add file → Create
new file → tên `LICENSE` → Choose a license template*), rồi sửa mục "Giấy phép" trong
`README.md`.

---

## 2. Art: CC BY-NC-SA 4.0

Art do dự án sinh ra là tài nguyên **trả tiền** (CLAUDE.md quy tắc 13). CC BY-NC-SA 4.0 nghĩa
là:

- **BY** — dùng lại phải ghi tên dự án.
- **NC** — **không được dùng để kiếm tiền** (bán, đưa vào game có thu phí / quảng cáo…).
- **SA** — sửa lại thì phải chia sẻ bản sửa theo đúng giấy phép này.

Người ta vẫn chạy được game từ mã nguồn để học, để chơi với bạn bè; nhưng không ai lấy art của
anh làm game thương mại được. Muốn dùng thương mại thì phải xin anh.

Lưu ý: các tệp của người khác trong `assets/` (font, hạt Kenney CC0…) **không** thuộc giấy phép
này — chúng giữ giấy phép gốc, ghi trong `assets/src/CREDITS.md`. Bản nháp dưới đã ghi rõ
ngoại lệ đó.

Một điểm cần anh biết: tình trạng bản quyền của ảnh sinh bằng AI còn chưa rõ ở nhiều nước
(có nơi không công nhận bản quyền cho ảnh AI thuần). Giấy phép vẫn nên có — nó nói rõ ý anh
muốn — nhưng đừng coi nó là lá chắn tuyệt đối. Tên + logo (mục 3) chắc chắn hơn.

### Bản nháp `assets/LICENSE` (CHƯA tạo tệp — chép ra khi anh duyệt)

```text
Art của "Đạo Hữu Xin Dừng Bước" (BladeAscend)
Bản quyền (c) 2026 chủ dự án BladeAscend.

Trừ khi có ghi chú khác, mọi tệp ảnh, hoạt ảnh và âm thanh trong thư mục `assets/` do dự án
tạo ra được cấp phép theo Creative Commons Ghi công – Phi thương mại – Chia sẻ tương tự 4.0
Quốc tế (CC BY-NC-SA 4.0): https://creativecommons.org/licenses/by-nc-sa/4.0/deed.vi

Ghi công khi dùng lại: "Art từ Đạo Hữu Xin Dừng Bước (BladeAscend) —
https://github.com/chinhhi102/bladeascend — CC BY-NC-SA 4.0".

NGOẠI LỆ — các tệp sau thuộc bên thứ ba và giữ giấy phép gốc của họ (xem
assets/src/CREDITS.md):
  - assets/src/fonts/*            SIL Open Font License 1.1
  - assets/src/vfx/particles/*    Kenney Particle Pack, CC0 1.0 (LICENSE-kenney.txt)

Logo, tên "Đạo Hữu Xin Dừng Bước" và "BladeAscend" KHÔNG thuộc giấy phép này — xem mục
"Tên và logo" trong README.md.

Muốn dùng art cho mục đích thương mại: liên hệ chủ dự án qua GitHub.

---

Art of "Đạo Hữu Xin Dừng Bước" (BladeAscend)
Copyright (c) 2026 the BladeAscend project owner.

Unless otherwise noted, all image, animation and audio files in `assets/` created by this
project are licensed under the Creative Commons Attribution-NonCommercial-ShareAlike 4.0
International License (CC BY-NC-SA 4.0): https://creativecommons.org/licenses/by-nc-sa/4.0/

Third-party files listed in assets/src/CREDITS.md keep their original licenses. The game
name, "BladeAscend" and the logo are NOT covered by this license.
```

---

## 3. Tên và logo: giữ bản quyền, fork phải đổi tên

Dù chọn AGPL hay MIT cho code, **giấy phép mã nguồn không cấp quyền dùng tên và logo** —
nhưng phải nói rõ ra, để không ai hiểu lầm. Đề xuất thêm vào `README.md` (khi chọn xong
giấy phép):

```text
## Tên và logo

"Đạo Hữu Xin Dừng Bước", "BladeAscend", logo và tên miền bladeascend.io là tên riêng của
dự án, không thuộc giấy phép mã nguồn hay giấy phép art. Bạn được fork và chạy mã nguồn theo
giấy phép, nhưng bản fork công khai (nhất là server cho người khác chơi) phải ĐỔI TÊN và
KHÔNG dùng logo, để người chơi không nhầm với bản chính thức. Nhắc tới dự án gốc để ghi
nguồn ("dựa trên BladeAscend") thì được và rất hoan nghênh.

"Đạo Hữu Xin Dừng Bước", "BladeAscend", the logo and the bladeascend.io domain are not
licensed under the code or art licenses. Public forks — especially servers others can play
on — must use a different name and must not use the logo.
```

Muốn bảo vệ chắc hơn về pháp lý thì đăng ký nhãn hiệu tại Cục Sở hữu trí tuệ — việc đó ngoài
phạm vi repo, anh tự cân nhắc.

---

## Việc còn chờ anh quyết

1. Code: AGPL-3.0 hay MIT? → tạo `LICENSE` ở gốc.
2. Art: đồng ý CC BY-NC-SA 4.0? → chép bản nháp mục 2 thành `assets/LICENSE`.
3. Thêm mục "Tên và logo" (mục 3) vào `README.md`.
4. (Tuỳ) thêm dòng CLA đơn giản vào `CONTRIBUTING.md` nếu chọn AGPL và muốn giữ quyền bán
   giấy phép riêng.
