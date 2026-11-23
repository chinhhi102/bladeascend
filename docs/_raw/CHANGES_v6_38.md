# Việc cần làm — spec v6.57 (chốt cách đo rò rỉ client)

Trả lời câu hỏi trong `docs/QUESTIONS.md`. Spec đã sửa: `docs/GAME_SPEC.md` §13.9 (viết lại bài kiểm rò rỉ client) và §13.10 (ghi rõ 80 MB là đích, không phải cổng). Dán file mới vào `docs/_raw/`.

## Chốt: hướng (1) — kéo dài cửa sổ. Nhưng việc quan trọng nhất là việc số 1 dưới đây.

Dốc 6 lượt chênh nhau **hơn 10 lần** (0,19 → 2,00 MB/phút). Nhiễu máy ±13–15 MB không giải thích được khoảng cách đó. Nguyên nhân nhiều khả năng hơn là **mỗi lượt chơi một kiểu** — khác số lần đổi khu, khác khu kết thúc, khác mật độ. Kéo dài cửa sổ mà kịch bản vẫn trôi nổi thì chỉ được một con số nhiễu-ít-hơn của một thứ đang đo sai.

### 1. Kịch bản cố định (làm trước, không có cái này thì cả (1) lẫn (2) đều vô nghĩa)
- Cùng seed, cùng chuỗi hành động, cùng số lần đổi khu (50) và chết/vào lại (10), **cùng khu kết thúc**, cùng thứ tự.
- Chạy 3 lượt liên tiếp trên **cùng một commit** và báo 3 dốc. Nếu 3 dốc còn chênh > 3 lần thì kịch bản chưa cố định thật — báo tôi trước khi đi tiếp.

### 2. Cửa sổ 60 phút, xét dốc theo khối
- Phút 0–10 khởi động (bỏ), **cửa sổ đo phút 10–60**, lấy mẫu **mỗi 15 s**.
- Chia 50 phút thành **5 khối 10 phút**, mỗi khối fit 1 dốc OLS → 5 dốc → KTC 95% của dốc trung bình (t, df = 4).
- Bỏ hẳn cách "so 2 mẫu đầu/cuối" và bỏ ngưỡng "tăng > 5%" cũ.

### 3. Luật đỏ/vàng
- **ĐỎ**: cận dưới KTC 95% > 0 **và** dốc × 60 phút > 10% mức nền; hoặc mức nền cuối cửa sổ vượt **mốc kỷ lục**; hoặc GPU MB vượt ngân sách.
- **VÀNG**: cận dưới KTC ≤ 0 nhưng **3 đêm liên tiếp cùng dấu dương** → thẻ vàng + mở việc điều tra; đêm thứ 4 vẫn dương → ĐỎ. (Đây là hướng (2) của em, giữ lại làm lưới bắt rò rỉ chậm chứ không làm luật chính.)
- Dashboard admin hiện: 5 dốc khối, dốc trung bình, KTC, mốc kỷ lục, chuỗi dấu của 7 đêm gần nhất.

### 4. Mốc kỷ lục cho mức nền (ratchet)
- Lưu **mức nền ổn định thấp nhất từng đo** (giá trị tại phút 10). Build vượt mốc > 10% → ĐỎ.
- Mốc chỉ đi xuống; hạ mốc cần 2 đêm xác nhận.
- **Không** biến 80 MB thành cổng lúc này: nền đang 137–152 MB, bật lên thì mọi build đều đỏ và cổng hết tác dụng báo hiệu. Khi nền ổn định < 88 MB trong 3 đêm thì mới chuyển 80 MB thành cổng cứng.

### 5. Sổ tài nguyên
- Xét theo **dốc** như §13.9, lấy mẫu tại **cùng một điểm trong chu kỳ kịch bản** (vd luôn ngay sau khi vào khu thứ k).
- Bảng "loại nào phình" chỉ được in dốc + KTC, không in chênh lệch 2 mẫu tức thời. Con số "n_sprite +1 094 (+31%)" hiện tại ghi là **chưa kết luận**.

### 6. Giả thuyết "heap tăng theo số lần đổi khu"
- Để treo ở trạng thái **chưa xác nhận**. Chỉ thử lại sau khi có kịch bản cố định, và cần **≥ 3 lượt mỗi bên**.

### 7. Rò rỉ thì vẫn phải tìm
Cách đo mới là để **biết chắc**, không phải để mở cổng. 6/6 lượt dốc dương (p ≈ 0,016) là rò rỉ thật — song song với việc sửa bài đo, vẫn truy nguyên nhân: so snapshot heap đầu/cuối cửa sổ, xem retainer của nhóm lớn nhất, soi riêng đường đổi khu (ground chunk, atlas, listener, texture chữ).

**Deploy vẫn chặn** cho tới khi cổng xanh theo luật mới — đúng như em làm. Không chạy lại bài để "mong ra xanh"; chỉ chạy lại sau khi có thay đổi thật.

## Hai chỗ em rút lại: ghi nhận, đúng.
Chẩn đoán "cổng đỏ do phép đo yếu" và kết luận 2 lượt — cả hai đều nên rút. Cách xử lý (ghi rõ là chưa xác nhận thay vì xóa) là đúng.
