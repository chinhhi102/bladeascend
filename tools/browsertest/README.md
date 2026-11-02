# Bài đo chạy bằng trình duyệt

Mỗi tệp `.mjs` ở đây mở game thật bằng Playwright rồi đo một thứ. Chạy từ thư mục này:

```bash
cd tools/browsertest
node uioverlap.mjs http://127.0.0.1:9999
```

Lần đầu: `npm install && npm run install-browser`.

## Luật: KHÔNG tìm phần tử bằng chữ hiển thị

Dùng **class**, không dùng `getByText('Vào chơi')`.

| Việc | Selector |
|---|---|
| Nút "Vào chơi" ở trang chủ | `.home .btn.primary` |
| Nút "Chơi tiếp" ở màn tổng kết | `.summary .btn.primary` |

Vì sao có luật này, bằng một lần đã trả giá:

Từ khi có i18n (D13), chữ trên nút theo `navigator.language`. Chromium headless báo locale
**Anh**, nên nút thành "Play" — và **15 bài đo trong thư mục này chết cùng lúc** với
`TimeoutError: waiting for getByText('Vào chơi')`. Ngày 2/11 `uioverlap.mjs` chết ngay ở cỡ
màn hình đầu tiên, bỏ qua 6/7 cỡ còn lại; trong quãng đó trang chủ vỡ tới mức **nút "Vào
chơi" không nhìn thấy ở 4/7 cỡ** mà không bài đo nào kêu.

Chữ hiển thị là thứ người ta đổi thường xuyên và đổi mà không nghĩ tới bài đo. Class thì
không.

## Bài nào cần ép ngôn ngữ

Bài nào **đo chữ** (bề ngang, xuống dòng, tương phản) thì mở bằng `?lang=vi` — tiếng Việt
là bản chữ dài nhất, tức trường hợp khó nhất cho mọi luật bề ngang. Hiện chỉ `uioverlap.mjs`
cần. Bài chỉ bấm nút rồi đo số thì không cần, vì selector theo class đã không phụ thuộc
ngôn ngữ.

## Bài nào cần cửa debug

`shots.mjs`, `skillshots.mjs` và các bài dựng sẵn trạng thái cần phím test. `?debug=true`
trên URL **không còn cấp quyền gì** (ADMIN_SPEC §1.1) — phải chạy server với:

```bash
cd server && APP_ENV=dev go run ./cmd/server -debug=true
```

Thiếu `APP_ENV=dev` thì phím test im lặng không làm gì, bài đo vẫn chạy tới cuối và cho ra
kết quả sai mà không có một dòng lỗi nào. `shots.mjs` tự kiểm bằng cách đếm kiếm trước/sau;
bài mới nên làm tương tự.
