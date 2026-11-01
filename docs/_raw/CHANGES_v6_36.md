# Việc cần làm — GAME_SPEC v6.36 (hiển thị kỹ năng) + các quyết định

## Quyết định
1. **Duyệt OG có logo** → `SHARE_CARD_APPROVED = True`, dựng lại, deploy.
2. **Lịch sử git có dữ liệu người chơi**: repo hiện 0 fork / 0 sao → đây là lúc rẻ nhất để làm sạch. Làm theo thứ tự:
   a. Sao lưu repo đầy đủ (`git clone --mirror` ra ~/backup/bladeascend-mirror-<ngày>).
   b. `git filter-repo` xoá/thay các tệp chứa tên, UUID đầy đủ, IP người chơi trong lịch sử (liệt kê tệp trước, gửi tôi danh sách rồi mới chạy).
   c. Force push main; báo tôi để tôi kiểm trên GitHub. Ghi vào PROGRESS cách khôi phục từ bản mirror.
3. **SECURITY.md**: không dùng email — bật GitHub **Private vulnerability reporting** (Settings → Code security) và ghi trong SECURITY.md: "báo qua tab Security → Report a vulnerability".
4. IP LAN <LAN_IP> trong docs/_raw: thay bằng `<LAN_IP>` khi tôi gửi spec bản sau (tôi sẽ sửa bản _raw). Agent sửa các tệp ngoài _raw.
5. B21 (Thiên Đình p95 3,5–3,75 ms, sát 4 ms): giữ sàn mới. Thêm cảnh báo trên dashboard admin khi tick p95 khu > 3,5 ms trong 5 phút.

## Hiển thị kỹ năng (§6.5)
6. Rà 10 kỹ năng: bảng hiện trạng từng lớp ①–⑤ (có/thiếu) → docs/SKILL_VFX_AUDIT.md trước khi làm.
7. Làm đủ 5 lớp theo bảng §6.5 cho cả 10 kỹ năng; event `StatusStart/StatusEnd` từ server; icon trạng thái trên HUD người bị trúng + mũi tên mép màn hình khi bị trúng từ ngoài khung nhìn.
8. Icon trạng thái 64 px (8 loại): vector trước, thêm job artgen sau (không chờ art).
9. Chụp mỗi kỹ năng 1 ảnh (góc nhìn người đứng xem) + 1 ảnh màn hình người bị trúng → docs/review/skills/.
10. Mỗi mục: commit + deploy + push, báo commit hash.
