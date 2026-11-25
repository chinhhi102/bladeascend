# Việc cần làm — spec v6.59 (trả lời Q56, Q57)

## Q56 — mốc kỷ lục trên một mức nền chênh 31%

Đề xuất "vượt mốc cũng cần 2 đêm xác nhận" — **duyệt**, nhưng chỉ nó thì chưa đủ: chênh 31% mà ngưỡng 10% thì hai đêm liên tiếp đỏ oan vẫn xảy ra thường. Phải sửa cả phép đo và cả định nghĩa mốc. Spec §13.9 đã viết lại ba chỗ:

1. **Mức nền không còn là một mẫu ở phút 10.** Heap răng cưa theo nhịp GC, nên một mẫu chủ yếu đo "lúc đó GC vừa chạy hay chưa" — đây là nguồn chênh 31% nhiều khả năng nhất.
   - **Gọi GC cưỡng bức ngay trước mỗi mẫu** (`--js-flags=--expose-gc`, hoặc CDP `HeapProfiler.collectGarbage`). Đáy sau GC mới là phần thật sự bị giữ lại.
   - **Mức nền = phân vị 10 của toàn bộ mẫu trong khối 1** (phút 10–20), không phải 1 mẫu.
   - Làm xong hai việc này thì **đo lại độ chênh của 3 lượt cùng commit** và báo tôi con số mới. Tôi đoán nó tụt xuống dưới 10%; nếu không thì còn một nguồn nhiễu khác chưa tìm ra.

2. **Mốc = trung vị mức nền của 3 đêm hợp lệ gần nhất**, không phải "mức thấp nhất từng đo". Chỗ này là lỗi thiết kế của tôi ở v6.57: trên một phép đo còn nhiễu, "thấp nhất từng đo" sẽ bám vào lượt may mắn nhất rồi khiến mọi lượt sau đỏ oan. Mốc vẫn chỉ đi xuống. Hạ mốc cần 2 đêm, **vượt mốc cũng cần 2 đêm** — đối xứng, đúng như em đề xuất.

3. **Ngưỡng không được thấp hơn sàn nhiễu của chính nó.** Ngưỡng vượt mốc = **≥ 2× độ chênh 3 lượt cùng commit**, tối thiểu 10%. Trong khi độ chênh còn > 10% thì cổng mức nền **chỉ cảnh báo vàng, không chặn**. Cổng **dốc** (KTC) vẫn chặn bình thường — nó không dùng mức nền nên không bị ảnh hưởng.

Hệ quả: hiện tại cổng mức nền ở chế độ cảnh báo, deploy vẫn bị chặn bởi cổng dốc (0,69 MB/phút, KTC chưa cắt 0). Không đổi kết luận, chỉ đổi lý do chặn cho đúng.

Về "nền ~88 MB": nó **chưa** đủ để bật 80 MB thành cổng cứng. Vừa đúng mép 88, lại đo bằng cách cũ. §13.10 nay ghi rõ điều kiện 3 đêm < 88 MB phải đo bằng cách của v6.59.

## Q57 — một lượt 2 giờ cho nguyên nhân 4: duyệt

Chạy **đêm nay trong khung 03:00**, không chiếm máy giờ làm việc. Trước khi chạy ghi rõ vào báo cáo **cái gì sẽ phân định**, để lượt đo không kết thúc bằng "có vẻ như":

- **Bão hoà** = sau một mốc nào đó buffer gom lô đứng yên trong ≥ 40 phút cuối, kể cả khi vẫn đổi khu và vẫn chết. Nếu vậy thì **đây không phải rò rỉ**, mà là một khoản đỉnh một lần (~35,6 MB) và phải vào **bảng ngân sách §13.10 thành một dòng riêng**; cách sửa là **chặn trần kích thước lô**, không phải đụng VFX/decor.
- **Không bão hoà** = còn dốc dương sau 2 giờ → là rò rỉ thật, lúc đó mới bàn tới cách vẽ.

Dự đoán của tôi là nó bão hoà: số vật thể trên màn hình có trần, nên lô cũng có trần — 87 → 274 KB/cái nghe như lô phình tới mức đỉnh rồi `shrinkToFit: false` giữ nguyên, chứ không phải tăng vô hạn. Nhưng đó chỉ là dự đoán, cứ đo.

**Không đụng VFX/decor trước khi có kết quả 2 giờ.**

## Push: push đi

20 commit nằm trên máy em là rủi ro lớn hơn việc repo công khai thấy code chưa deploy. Commit không phải deploy — cổng đỏ chặn deploy, không chặn push. Push, rồi báo tôi hash.

## Hai chỗ em rút lại
Cả hai đều đúng và đáng ghi lại thành luật làm việc:
- **"Sửa nguyên nhân 1 là xong"** → nguyên nhân sau có thể bị nguyên nhân trước che. Sau mỗi lần sửa rò rỉ phải đo lại toàn bộ bộ đếm, không chỉ bộ đếm vừa sửa.
- **"200 Buffer mồ côi"** → ảnh heap trả lời *cái gì vừa sinh ra*; *cái gì đang tích lại* phải đọc từ một con số đếm theo thời gian. Câu này nên thành một dòng trong `CLAUDE.md`: **kết luận rò rỉ chỉ được rút ra từ dốc theo thời gian, không từ một ảnh heap**.

Và hướng sai `context: true` cho cả 24 chỗ — phát hiện trong 15 giây nhờ có `check:destroy` + FPS/texture trong bài đo. Đúng kiểu lỗi mà cổng tự động phải bắt, chứ không phải người đọc code bắt.
