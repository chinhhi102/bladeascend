# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026; cập nhật 02/11: Q37–Q42 đã xong (spec v6.35). Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.


## Q56 — mốc kỷ lục gác MỨC NỀN, mà mức nền chênh 31% giữa ba lượt giống hệt nhau

**Bối cảnh.** Ba lượt nghiệm thu 60 phút, cùng commit, cùng kịch bản cố định, máy rảnh
(`tools/leakrepeat.sh 3 60`, PERF_LOG #55):

| | dốc heap | mức nền (phút 10) |
|---|---:|---:|
| lượt 1 | +1,565 MB/phút | **88,0 MB** |
| lượt 2 | +1,317 | **112,2** |
| lượt 3 | +1,424 | **115,1** |

**Dốc đã tái lập được** (chênh 1,2 lần — phép nghiệm thu của v6.57 ĐẠT). **Mức nền thì
chưa**: lượt 2 và 3 khớp nhau ~3%, riêng lượt 1 thấp hơn 24% và thấp đều từ phút 4 trở đi.

**Vì sao nó đụng vào luật.** Mốc kỷ lục §13.9 lấy **mức nền thấp nhất từng đo** rồi chặn
build vượt > 10%. Với khoảng 31% này: nếu lượt 1 đặt mốc 88 thì lượt 2 và lượt 3 — cùng
build, cùng kịch bản, máy rảnh, không rò rỉ gì thêm — đều ĐỎ vì "vượt mốc 28%". Cổng sẽ đỏ
vì nhiễu, và một cổng đỏ vì nhiễu thì sớm muộn bị bỏ qua.

**Phương án đang dùng (không đổi luật):** giữ nguyên v6.58 — mốc vẫn gác, vẫn 10%.

**Đề xuất:** làm cho nó **đối xứng**. Spec đã bắt *hạ* mốc phải có 2 đêm xác nhận; đề nghị
*vượt* mốc cũng cần **2 đêm liên tiếp** mới ĐỎ (đêm đầu là thẻ vàng). Một dòng luật, bịt
đúng chỗ nhiễu, và không nới ngưỡng 10%.

Hai cách khác nếu anh thích hơn: (b) mốc so với **trung vị 3 đêm hợp lệ gần nhất** thay vì
giá trị thấp nhất từng đo; (c) nới ngưỡng 10% → 35%, nhưng cách này làm cổng gần như mất
tác dụng với một thoái lui thật 20–30%.

## Q57 — phần rò rỉ còn lại là buffer gom lô của Pixi: sửa cách vẽ, hay chấp nhận mức nền?

**Đã khoanh được bằng ảnh heap cửa sổ 50 phút (phút 10 → 60, máy rảnh):**

| nhãn buffer | số cái | MB |
|---|---|---|
| `attribute-batch-buffer` | 105 → 141 | **9,18 → 38,60 (+29,4)** |
| `index-batch-buffer` | 105 → 141 | 2,11 → 8,33 (+6,2) |
| mọi nhãn còn lại (mesh, particle) | ~đứng | ~0,00 |

**35,6 trên 35,8 MB** phần heap tăng thêm nằm ở đúng hai dòng này. Chúng là buffer hình học
của **bộ gom lô (Batcher) của PixiJS**, tạo với `shrinkToFit: false` — mỗi batcher giữ lô lớn
nhất nó từng gặp và **không bao giờ nhả bớt**. Trung bình 87 KB → **274 KB mỗi cái** sau 50
phút. Bộ dọn rác của Pixi không giúp được: chúng được dùng mỗi khung nên không bao giờ "rỗi".

Số batcher tăng theo số **GraphicsContext còn sống**, mà cảnh của mình đang giữ **330–400**
context (riêng `mapView.under` 157 cái: bùn, đường, nước — mỗi hình một `Graphics`).

**Đây có thể KHÔNG phải rò rỉ.** Dốc giảm dần theo từng khối (lượt 2: 1,60 · 0,77 · 0,60 ·
0,43 · 0,34) — dạng của một đại lượng **đang bão hoà**. Nếu nó dừng quanh ~40 MB thì đó là
**mức nền cao**, không phải rò rỉ vô hạn.

**Khuyến nghị của tôi: ĐO TRƯỚC, đừng sửa cách vẽ.** Một lượt **2 giờ** (cửa sổ phút 10–120)
trả lời dứt khoát "bão hoà hay không" — và câu trả lời quyết định làm gì tiếp:

- **Nếu bão hoà** → không sửa gì trong VFX/decor. Chỉ cần ghi nhận mức nền mới, và cân nhắc
  hạ `gcMaxUnusedTime` của Pixi để batcher rỗi được nhả sớm hơn. Cổng §13.9 nên tính dốc trên
  cửa sổ SAU khi bão hoà, không thì mọi đêm đều đỏ vì một thứ không tăng mãi.
- **Nếu không bão hoà** → mới đụng tới cách vẽ, và lúc đó có hai hướng, cả hai đều cần anh
  duyệt vì chúng đổi cách vẽ chứ không chỉ đổi code: (a) gom decor tĩnh (`mapView.under`)
  thành texture/sprite hoặc dùng **context dùng chung** như `entities.ts` đã làm cho viên
  Linh khí — bớt vài trăm context; (b) ép `batchMode` cho từng nhóm Graphics để chúng không
  mỗi cái một batcher.

**Phương án đang dùng:** chưa sửa gì, chưa đụng VFX. Bài 2 giờ cần một máy rảnh liền 2 tiếng
nên tôi chờ anh bật đèn xanh (hoặc để cron 03:00 chạy, nhưng bài hằng đêm hiện là 60 phút).
