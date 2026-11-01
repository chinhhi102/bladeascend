# Câu hỏi chờ chủ dự án quyết

> Claude ghi vào đây khi gặp chỗ cần quyết mà không muốn tự đoán (CLAUDE.md luật 10).
> Câu nào trả lời xong thì chuyển sang `docs/QUESTIONS_DONE.md`.
## Q32 — Spec §13.6 mục 8 chưa ghi luật chất lượng anh chốt 1/11 (chỉ cần anh sửa chữ)

Bối cảnh: lượt 1/11, việc 4 (chất lượng mặc định). Tệp: `client/src/render/quality.ts`.

Code đã làm đúng lời anh chốt trong chat:
- Mặc định: PC → **Vừa**; điện thoại / GPU tích hợp → **Thấp**. Không ai bắt đầu ở Cao.
- Vừa → Cao chỉ khi ≥ 55 FPS suốt **20 giây**.

Spec §13.6 mục 8 hiện ghi: "FPS ổn lại **10s** → bật dần" và không nói mức mặc định. Tôi giữ
**Thấp → Vừa = 10 giây** theo spec, chỉ bậc lên Cao là 20 giây. Không sửa `docs/_raw/`.

Anh chỉ cần thêm vào §13.6 mục 8 (đề xuất câu chữ):
> Mặc định: PC Vừa, điện thoại/GPU tích hợp Thấp. Thấp→Vừa sau 10s ổn ≥ 55 FPS; Vừa→Cao sau 20s.
> Chỉ lựa chọn tay của người chơi được lưu; máy tự đổi không lưu.

Nếu anh muốn Thấp → Vừa cũng 20 giây thì báo, đổi một hằng số.
