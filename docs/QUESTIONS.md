# Câu hỏi đang chờ anh trả lời

Mỗi câu: bối cảnh → phương án → **tạm dùng** (đã code theo phương án này, đổi lại dễ).
Câu đã xong ở `docs/QUESTIONS_DONE.md`.

---

## Q16. Vật phẩm Trường Kiếm — spec có, anh nói không có

**Bối cảnh:** anh nhắn "không có item trường kiếm, có item tăng tốc độ quay của kiếm chứ
không có trường kiếm". Nhưng `docs/_raw/GAME_SPEC.md` §6.2 đang ghi rõ nó:

> | Trường Kiếm | Chiều dài mọi vòng kiếm × 1.3 | 12s | 7 |

Và nó **đang có thật trong game**: đếm thử 1 phòng 200 vật phẩm ra 13 viên Trường Kiếm
(ngang với Hộ Thuẫn 15, Song Tu 9), ảnh `assets/buffs/long.png` cũng đã vẽ xong — bong bóng
tím, thanh kiếm kèm mũi tên lên, khác hẳn viên Tốc Kiếm (bong bóng xanh ngọc hình lốc xoáy).

**Phương án:**
1. **Giữ nguyên** — Trường Kiếm có thật, chỉ là anh chưa nhặt trúng (7% số vật phẩm).
2. **Bỏ hẳn** — xoá `long` khỏi `buffs` + `items.weights`, chia lại trọng số. Việc này **sửa
   §6.2 và §4.1a của spec**, mà spec chỉ có anh sửa.

**Tạm dùng: phương án 1** (code không đổi) vì spec đang ghi là có.

Nếu anh muốn bỏ, nhắn một câu là xoá — nhưng anh cần sửa §6.2 trong `docs/_raw/` trước, vì
§4.1a cũng đang dùng `buff_long` trong công thức tầm chém.

---

## Q17. Bỏ chặn đường đi làm "Bó Kiếm khi rớt cảnh giới" gần như không xảy ra

**Bối cảnh:** sau khi chuyển lá chắn thành luật chém (xem Q15 ở file DONE), nạn nhân chỉ ăn
được cú chém thân khi **còn đúng 1 vòng, tức ≤ 10 kiếm**.

Mà §5.1 nói rớt cảnh giới thì kiếm thừa so với trần cảnh giới mới rơi xuống thành **Bó Kiếm**.
Trần thấp nhất là Luyện Khí 10 kiếm — nạn nhân lúc bị chém đã ≤ 10 kiếm rồi, nên **không bao
giờ còn kiếm thừa để rơi**. Nhánh Bó Kiếm chỉ còn chạy lúc chết hẳn (rơi 50% số kiếm).

**Phương án:**
1. **Chấp nhận** — Bó Kiếm chủ yếu đến từ xác người chết và từ vật phẩm rơi sẵn trên bản đồ.
2. **Cho chém thân sớm hơn**: lá chắn chỉ tính từ vòng thứ 3 trở lên (còn ≤ 2 vòng = ≤ 30
   kiếm là chém thân được). Lúc đó rớt từ Kim Đan xuống Trúc Cơ mới có kiếm thừa rơi ra.

**Tạm dùng: phương án 1** — đúng nguyên văn §5.2 ("còn 1 vòng mới chạm được tới thân").

---

## Q18. Linh khí và Tiên khí giờ tách hẳn — §3.4 vẫn ghi "đổi qua lại"

**Bối cảnh:** anh nhắn 2026-10-28: *"cảnh giới tiên không thu thập exp linh khí nữa, chỉ hấp
thu tiên khí, ngược lại đối với dưới cảnh giới tiên cũng không hấp thụ exp tiên khí, chỉ hấp
thu linh khí"*. Spec §3.4 đang ghi ngược lại:

> Dưới mức đó: vẫn nhặt được, mỗi điểm Tiên khí **đổi thành `50 × 2^r` Linh khí**

**Đã làm theo lời anh:** ranh giới là Đại Thừa tầng 10 (`BanksTienKhi`).
- Từ Đại Thừa tầng 10 trở lên: viên Linh khí **đi xuyên qua**, `GainExp` không vào.
- Dưới đó: viên Tiên khí **đi xuyên qua**, `GainTienKhi` trả 0.
- Viên không hấp thụ được vẫn nằm đó cho người dùng được tới nhặt, và client **làm mờ 30%**
  để không chạy theo vô ích.

**Một hệ quả anh cần biết:** người phàm chém chết người Tiên giới giờ **không nhận Tiên khí**
nữa (trước là đổi ra Linh khí). Phần thưởng của họ chỉ còn **Uy Danh** — mà chém trúng Đạo Tổ
là 1 400 điểm × lật kèo (§11.1), vẫn là món hời lớn nhất bản đồ. Nếu anh thấy vậy là bạc bẽo
quá, nói một câu em cho riêng đường cướp đoạt được đổi ra Linh khí.

Cần anh sửa §3.4 trong `docs/_raw/` cho khớp.

---

## Q19. Tốc độ bể kiếm: bỏ cooldown 0.3s của §5.2

**Bối cảnh:** anh nhắn *"tốc độ vỡ của kiếm khá chậm... vòng ngoài cùng có 30 kiếm, tốc độ
quay 1000ms/vòng thì sau 1 giây nếu cả 2 ngang nhau thì cả 2 đã mất hết kiếm ở vòng ngoài"*.
Spec §5.2 ghi: *"Cooldown tính bể kiếm mỗi cặp (A, B) 0.3s"* → chỉ 3.3 kiếm/giây, mài hết
vòng 30 kiếm mất 9 giây.

**Đã làm:** tách làm hai cooldown riêng.
- `clash_cooldown: 0` — bể kiếm tính **mỗi tick**, tức 30 kiếm/giây, đúng con số anh đưa.
- `flip_cooldown: 0.3` — giữ nguyên, **chỉ** giới hạn việc đổi chiều quay, nếu không vòng
  kiếm sẽ giật 30 lần/giây.

Đo thật: hai người Kim Đan 60 kiếm (3 vòng 10+20+30) đứng mài — sau đúng 1,0 giây cả hai còn
30 kiếm, tức rụng trọn vòng ngoài. Có test giữ (`TestClashGrindsOneOuterRingPerSecond`).

Cần anh sửa câu "cooldown 0.3s" ở §5.2 cho khớp.
