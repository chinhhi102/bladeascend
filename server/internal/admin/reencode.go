package admin

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
)

// Mã hoá lại ảnh góp ý — ADMIN_SPEC §7.1, chủ dự án chốt Q39 (GAME_SPEC v6.34).
//
// Gỡ khối EXIF theo cấu trúc tệp (imgmeta.go) chỉ bỏ được những gì mình BIẾT là siêu dữ
// liệu. Giải mã ra điểm ảnh rồi mã hoá lại thì tệp ra chỉ còn điểm ảnh: mọi khối lạ, dữ liệu
// giấu sau EOI, ảnh nhúng… đều mất, kể cả loại chưa ai nghĩ tới.
//
// Chỉ dùng thư viện chuẩn: Go đọc/ghi được JPEG và PNG, KHÔNG đọc được WebP. Vì vậy client
// luôn gửi JPEG (imagecompress.ts), và server từ chối WebP thay vì cho qua mà không mã hoá.

// reencodeMaxSide: cạnh dài tối đa server chịu giải mã. Client đã thu về ≤ 1 600 px; trần
// này chặn "bom giải nén" — tệp 300 KB khai 30 000 × 30 000 px là 3.6 GB RAM khi giải mã.
const reencodeMaxSide = 4096

// reencodeJPEGQuality: đủ nét cho ảnh chụp màn hình góp ý; ảnh vào đã qua nén ở client.
const reencodeJPEGQuality = 85

// ErrImageTooLarge: kích thước điểm ảnh vượt trần giải mã.
var ErrImageTooLarge = errors.New("ảnh quá nhiều điểm ảnh")

// ReencodeImage giải mã JPEG/PNG rồi mã hoá lại cùng định dạng. Trả về byte mới và đuôi tệp.
func ReencodeImage(b []byte) ([]byte, string, error) {
	ext := imageExt(b)
	if ext != ".jpg" && ext != ".png" {
		return nil, "", ErrNotAnImage
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return nil, "", ErrBadImage
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > reencodeMaxSide || cfg.Height > reencodeMaxSide {
		return nil, "", ErrImageTooLarge
	}
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, "", ErrBadImage
	}
	var out bytes.Buffer
	if ext == ".png" {
		err = png.Encode(&out, img)
	} else {
		err = jpeg.Encode(&out, img, &jpeg.Options{Quality: reencodeJPEGQuality})
	}
	if err != nil {
		return nil, "", err
	}
	return out.Bytes(), ext, nil
}
