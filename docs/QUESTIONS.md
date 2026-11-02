# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026; cập nhật 02/11: Q37–Q42 đã xong (spec v6.35). Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.

## Q43 — `minimap.size_mobile`: spec ghi hai số khác nhau

Tìm ra 02/11 khi rà lại spec so với code.

`docs/_raw/GAME_SPEC.md` ghi hai giá trị mâu thuẫn cho cùng một khoá:

| Chỗ trong spec | Giá trị |
|---|---|
| Dòng 534 — bảng config §16 | `minimap.size_mobile` → **96 px** |
| Dòng 1585 — khối JSON mẫu §16 | `"size_mobile": 120` |

**Phương án đang dùng:** giữ **96** (`config/game.json`), tức theo dòng 534. Không đổi gì.

Đáng nói là bản rà cũ (`SPEC_AUDIT_v630.md`, mục 9 của Top 10) chấm đây là lỗi cài đặt —
"minimap điện thoại 96 px thay vì 120". Thật ra **code không sai**: spec có hai con số và
code theo một trong hai. Lần rà nào cũng sẽ chấm sai như vậy cho tới khi spec thống nhất.

Hỏi: giữ **96** hay đổi sang **120**? Chọn xong anh sửa giúp dòng còn lại trong spec cho khớp.

Để anh cân: minimap to hơn thì dễ nhìn trên điện thoại, nhưng nó là khối HUD chiếm chỗ nhiều
thứ hai sau bảng xếp hạng, và §13.6 có giới hạn tổng diện tích HUD che màn hình.
`tools/browsertest/uioverlap.mjs` đo đúng con số đó ở 7 cỡ màn hình — đổi sang 120 thì chạy
lại bài đó là biết ngay có vượt luật không.
