# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.

Bối cảnh chung: lượt cuối 01/11/2026; cập nhật 02/11: Q37–Q42 đã xong (spec v6.35). Chi tiết từng dòng: `docs/SPEC_AUDIT_v630.md`.
Mỗi câu ghi **phương án đang dùng** (an toàn, không đổi luật) — anh chỉ cần trả lời nếu muốn khác.

## Q58 — chiêu "phải chờ" bị ĐẨY thì có mất hay không? (spec chỉ nói "bị chém chết")

**Bối cảnh.** Đi tìm lỗi Bà Tám (28/11) thì thấy `updateChannel` huỷ chiêu khi người dùng
**dịch chuyển** quá `BodyR` khỏi chỗ bấm — một luật KHÔNG có trong spec. §6.3 cho cả ba chiêu
có nhịp chờ chỉ ghi một cách mất:

- Bà Tám: "**Bị chém chết** trong lúc lấy hơi thì mất chiêu"
- Sư Phụ Mì: "**Bị chém chết** trong lúc húp thì mất chiêu"
- Ông Ba Võng: "nếu **bị chém trúng mà có khiên Kim Cang đỡ** thì mất chiêu"

Luật "dịch chuyển thì mất" do một lượt làm trước đặt ra (ghi chú trong `updateChannel`: *"ai
đẩy được Ông Ba Võng ra khỏi chỗ thì cũng phá được câu hát"*). Nó hợp lý với ĐẨY, nhưng nó
cũng chính là đường làm Bà Tám mất chiêu trên mặt băng — chỗ đó tôi đã sửa (băng không còn
đẩy được người đang lấy hơi, vì §6.3 nói "đứng yên").

**Sau bản sửa, còn ba đường huỷ:** chết · qua cổng · **bị ĐẨY** (Loa của Bà Tám khác, Bung Cơ
của Bé Tí, sóng ra mai của Cụ Rùa, sạp Cô Ba, và cú đẩy khi khiên Kim Cang vỡ).

**Phương án đang dùng (không đổi luật):** giữ nguyên — bị đẩy thì mất chiêu. Hành vi đó nay
có **test khoá** (`TestPushCancelsChannel`, 3 lượt con): nếu một lượt làm sau gỡ luật đi thì
bài test đỏ và chỉ thẳng vào câu hỏi này, thay vì luật biến mất trong im lặng. Test cũng
canh hai điều đi kèm: huỷ chiêu **không hoàn lượt**, và luôn phát `<chiêu>_huy` để người chơi
biết vì sao mất lượt.

**Đề xuất:** GIỮ, và ghi vào spec. Nó là "cách chống lại" mà §6.3 mục 3 đòi mỗi kỹ năng phải
có, và nó đối xứng: Bà Tám hét được thì Bà Tám cũng bị hét. Nhưng nó đang là luật ngầm, nên
một lượt làm sau có thể lại gỡ nó đi mà không biết là đang đổi cân bằng.

Nếu anh muốn khác thì hai hướng: (b) **chỉ chết mới mất chiêu** — đúng chữ spec hiện tại,
nhưng ba chiêu đó mạnh lên đáng kể vì gần như không cản được; (c) **đẩy chỉ làm chậm**: chiêu
vẫn phát nhưng nổ ở chỗ mới, không mất lượt.
