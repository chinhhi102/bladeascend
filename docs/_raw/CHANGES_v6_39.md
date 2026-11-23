# Việc cần làm — spec v6.58

## Q55 — xong
`docs/GAME_SPEC.md` §13.9 dòng "Lịch chạy" đã sửa **30 → 60 phút**. Giữ 60, đúng như em làm: cửa sổ 20 phút không đủ cho 5 khối × 10 phút. Nửa tiếng máy mỗi đêm là giá chấp nhận được; 04:10 xong, không đụng janitor 04:30. Chuyển Q55 sang `QUESTIONS_DONE.md`.

## Việc mới — phát sinh từ lỗi `dbg('zone')` em vừa tìm ra

Lỗi đó quan trọng hơn vẻ ngoài của nó: **6 lượt đo cũ chạy trên một kịch bản gần như không đổi khu**. Nên:

1. **Mốc kỷ lục và sổ đêm phải reset.** Spec v6.58 thêm `scenario_version` vào `leak-ratchet.json`: đổi kịch bản → mốc cũ hết giá trị, bắt đầu lại từ lượt hợp lệ đầu tiên của kịch bản mới; `leak-nights.json` cũng bắt đầu lại chuỗi dấu. Nếu không làm, mốc ghi từ kịch bản dễ sẽ làm mọi lượt của kịch bản thật đỏ ngay, mà đỏ vì lý do sai.
2. **Lượt đầu của kịch bản mới sẽ cho dốc cao hơn 6 lượt cũ — đó là bình thường, không phải tụt lùi.** Giờ mới có 50 lượt đổi khu thật. Đừng so với 0,19–2,00 MB/phút cũ, và đừng coi là regression của commit nào.
3. **Kết luận "rò rỉ có thật" vẫn đứng**, nhưng phạm vi thu lại: nó đúng cho *chơi lâu*, chưa nói được gì về *đổi khu*. Ngược lại B38 ("heap tăng theo số lượt đổi khu") giờ có bài đo đủ sức thử — sau khi có 3 lượt ổn định thì chạy thêm 3 lượt với 0 lượt đổi khu, so hai bên.
4. **Soát các bài đo khác xem có lệnh rỗng tương tự.** `dbg('zone')` không tham số mà vẫn trả về OK là loại lỗi "thành công giả" kinh điển. Yêu cầu: mọi lệnh debug nhận tham số bắt buộc phải **lỗi** khi thiếu tham số, không im lặng bỏ qua; và mỗi bài đo phải **xác nhận hành động đã xảy ra** (đếm số lần đổi khu thật trong bảng mẫu và so với số mong đợi, lệch → KHÔNG KẾT LUẬN). Áp cho cả `crashtest`, `rejoin`, `gomdon-smooth`, `drawguard`.

## Đã ghi vào spec (v6.58)
- Lượt đo trên máy bận: **KHÔNG KẾT LUẬN** — không chặn deploy, **không ghi mốc kỷ lục, không ghi sổ đêm**. Một lượt nhiễu không được hạ mốc hay cắt chuỗi dấu. (Em làm đúng rồi, nay thành luật.)
- `scenario_version` cho mốc kỷ lục như mục 1.

Phần còn lại (diệt 3 nguồn ngẫu nhiên, chu kỳ 60 s, luật 16 vào bài đo, trang quản trị, treo B38) — duyệt, không sửa gì. Cứ chạy xong 3 lượt rồi báo 3 dốc.
