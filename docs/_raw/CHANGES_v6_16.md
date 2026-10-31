# Việc cần làm — GAME_SPEC v6.16 + ADMIN_SPEC v1.2

Làm sau khi xong lượt Protocol v2 đang chạy. Luật chung như các lượt trước (commit riêng, test/build/check pass, deploy, câu mới ghi QUESTIONS.md).

## A. Góp ý: tải ảnh lên (ADMIN_SPEC §7.1)
1. Nút "Thêm ảnh" (file input image/*, multiple), kéo-thả, dán Ctrl+V; tối đa 3 ảnh (tính cả ảnh chụp màn hình); xem trước + xoá.
2. Client thu nhỏ cạnh dài ≤ 1 600 px, nén WebP ≤ 300 KB/ảnh. HEIC không đọc được → báo rõ.
3. Server kiểm định dạng theo nội dung, **mã hoá lại ảnh để xoá EXIF/GPS**, 3 ảnh × 300 KB tối đa. Admin xem dạng PreviewGroup.
4. Kiểm bằng Playwright WebKit (iPhone) + Chromium Android.

## B. Admin: Ant Design + bản đồ toàn cảnh (ADMIN_SPEC §2.1, §5)
5. Chuyển trang admin sang React + antd v5 (bundle riêng, không ảnh hưởng game — đã duyệt). Layout Sider/Header/Content, sáng/tối.
6. Dashboard: Card + Statistic + ngưỡng màu. Mọi bảng: antd Table (header cố định, sort, filter, phân trang, tìm kiếm, Tag trạng thái, thời gian tương đối, Drawer chi tiết). Thao tác nguy hiểm: Modal.confirm + lý do.
7. Bản đồ: bỏ luồng vị trí người chơi 5 Hz; vẽ toàn cảnh khu bằng chính bộ vẽ bản đồ của game (art thật), phóng to/kéo mượt, xuất PNG; lớp phủ tuỳ chọn (bậc vật cản, lưới, --check theo r, vùng); thanh số liệu live của khu (2s, chỉ số).

## C. Đa ngôn ngữ (GAME_SPEC §19)
8. Gom mọi chuỗi vào `client/src/i18n/vi.json`; CI báo chuỗi cứng còn sót trong code.
9. Server gửi mã sự kiện + tham số, client dịch (kill feed, tin loan, danh hiệu).
10. Soạn en, pt-BR, es, id, th theo bảng thuật ngữ §19; câu tấu hài viết lại theo từng ngôn ngữ; ghi ngữ cảnh từng câu vào `i18n/NOTES.md` để người bản ngữ duyệt. Đánh dấu bản nháp chưa duyệt bằng `"_draft": true`.
11. Chọn ngôn ngữ (tự động + nút cờ + ?lang=), Noto Sans Thai cho th, chỉ tải font ngôn ngữ đang chọn.
12. SEO: trang chủ tĩnh theo ngôn ngữ (/, /en/, /pt/, /es/, /id/, /th/), title/description/OG, hreflang, sitemap.xml, đoạn giới thiệu + 3 ảnh chụp dưới khung game.

## D. Tên game (GAME_SPEC §20)
13. **Tên đã chốt** (GAME_SPEC v6.18 §20): quốc tế **BladeAscend**, Tiếng Việt **"Đạo Hữu Xin Dừng Bước"**. `config.brand = { name: "BladeAscend", domain: "bladeascend.io", name_vi: "Đạo Hữu Xin Dừng Bước", tagline_vi: "Kiếm bay vòng vòng, đạo hữu chạy vòng vòng." }`. Dùng đúng `<title>`, mô tả, khẩu hiệu, đoạn giới thiệu trong §20. Thay mọi chỗ hiển thị tên game (trang chủ, màn hình tải, tiêu đề tab, OG, manifest PWA, admin, ảnh chia sẻ). Khẩu hiệu ngôn ngữ khác trong i18n. Logo chữ: tạm render bằng font, thêm job artgen logo.
14. **Bỏ hẳn "Kiếm Xoay Tu Tiên"** ở mọi nơi người chơi thấy. CI: grep chuỗi "Kiếm Xoay" / "kiem-xoay" trong client, server, assets → fail (trừ CHANGELOG/lịch sử).
15. **Đổi tên project**: repo + thư mục gốc → `bladeascend`, module Go, `package.json` name, tiêu đề tài liệu trong `docs/` (GAME_SPEC, ADMIN_SPEC, CLAUDE.md…), service/systemd/Caddy nếu có tên cũ. Đổi tên repo GitHub (giữ redirect tự động của GitHub), cập nhật remote + `tools/deploy.sh`. Làm riêng 1 commit cuối lượt, deploy + self-check sau khi đổi.
