# Câu hỏi đang chờ anh trả lời

Mỗi câu: bối cảnh → phương án → **tạm dùng** (đã code theo phương án này, đổi lại dễ).
Câu đã xong ở `docs/QUESTIONS_DONE.md`.

---

*(Không còn câu nào đang chờ — tính tới 31/10/2026.)*

Q20, Q22, Q25 và Q26 anh đã trả lời trong `docs/_raw/ANSWERS.md` ngày 30/10; code đã làm
theo và cả bốn đã chuyển sang `QUESTIONS_DONE.md`. Hai việc "chờ anh quyết" — tối ưu
CHANGES_v4 mục 25–27, và hai thư viện `pixi-filters` / `revolt-fx` — cũng đã được duyệt,
xem cuối `QUESTIONS_DONE.md`.

**Thứ tự việc lớn anh chốt 31/10** (không phải câu hỏi):
1. ~~Q27-1 nền 5 khu~~ · ~~Q27-2 trang trí~~ · ~~trang chủ hiện ngay~~ — **xong 31/10**.
2. **Phần B của CHANGES_v6_4** — Tiên khí ×20 (§3.4, §4.1c): `tien_khi_ceilings`
   [2000, 6000, 12000, 20000, 30000], Thiêu Đốt Tiên giới theo tỉ lệ
   (`burn_tien_frac_per_sec` 0.004), viên Tiên khí mệnh giá 500/100/10/2/1 tối đa 40 viên,
   sửa test theo checklist §17.
3. **CHANGES_v6_5** — kỹ năng theo U, 3 lượt ở Tiên giới. ⚠ **chưa có `docs/_raw/CHANGES_v6_5.md`
   trong repo** — anh dán vào rồi tôi làm.
4. **NETCODE bước 3 — Protocol v2** (NETCODE_SPEC v1.2): giữ v1 chạy song song bằng cờ
   (server config + `?proto=1|2`), thoả thuận phiên bản lúc handshake; chia bước nhỏ, mỗi
   bước game chơi được + test + golden pass + ghi PERF_LOG. Thứ tự: ~~schema + generator +
   golden/fuzz~~ (xong) → `Spawn/Desc/Update/Despawn` tường minh → delta theo ack →
   keyframe 2s → lượng tử hoá §4.6.
   **Mục tiêu dừng**: respawn = 0 (bench có bộ đếm "quay lại" > 0 thật) và gói < 6 KB. Đạt
   thì dừng, không tối ưu thêm.

Tối ưu CHANGES_v4 mục 25–27 **đã xong** (mốc chuẩn #4 trong PERF_LOG). Hướng A (một thân
nhiều mặt) để sau cùng — art thử đã sinh cho 3 nhân vật, chưa code.
