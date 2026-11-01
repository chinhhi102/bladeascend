package admin

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/png"
	"testing"
)

// Q39 (GAME_SPEC v6.34): ảnh góp ý được GIẢI MÃ rồi MÃ HOÁ LẠI, không chỉ gỡ khối biết tên.

// Khối APP lạ (ví dụ APP9, không thuộc danh sách gỡ nào) vẫn biến mất sau khi mã hoá lại;
// ảnh ra vẫn đọc được, cùng kích thước.
func TestReencodeJPEGDropsEverythingButPixels(t *testing.T) {
	secret := []byte("NHA-NGUOI-CHOI-21.0285,105.8542")
	in := insertAfterSOI(testJPEG(t), jpegSeg(0xE9, secret), exifGPSBlock())
	out, ext, err := ReencodeImage(in)
	if err != nil || ext != ".jpg" {
		t.Fatalf("ReencodeImage: %v %q", err, ext)
	}
	if bytes.Contains(out, secret) || bytes.Contains(out, []byte("Exif")) {
		t.Fatal("dữ liệu ngoài điểm ảnh còn sót sau khi mã hoá lại")
	}
	img, _, err := image.Decode(bytes.NewReader(out))
	if err != nil || img.Bounds() != testImage().Bounds() {
		t.Fatalf("ảnh ra hỏng: %v", err)
	}
}

// PNG giữ PNG (ảnh chụp màn hình không nhoè), khối chữ tEXt mất.
func TestReencodePNGDropsText(t *testing.T) {
	src := testPNG(t)
	chunk := pngChunk("tEXt", []byte("Comment\x00guest-123456"))
	in := append(append(append([]byte{}, src[:33]...), chunk...), src[33:]...)
	out, ext, err := ReencodeImage(in)
	if err != nil || ext != ".png" {
		t.Fatalf("ReencodeImage: %v %q", err, ext)
	}
	if bytes.Contains(out, []byte("guest-123456")) {
		t.Fatal("khối tEXt còn sót")
	}
}

// WebP không mã hoá lại được bằng thư viện chuẩn → từ chối, không cho qua nửa vời.
func TestReencodeRejectsWebP(t *testing.T) {
	webp := append([]byte("RIFF\x10\x00\x00\x00WEBPVP8L"), make([]byte, 8)...)
	if _, _, err := ReencodeImage(webp); !errors.Is(err, ErrNotAnImage) {
		t.Fatalf("WebP phải bị từ chối, ra %v", err)
	}
}

// Ảnh khai kích thước khổng lồ bị chặn TRƯỚC khi giải mã (bom giải nén).
func TestReencodeRejectsHugeDimensions(t *testing.T) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewGray(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	b := buf.Bytes()
	binary.BigEndian.PutUint32(b[16:], 30000) // IHDR width
	binary.BigEndian.PutUint32(b[20:], 30000) // IHDR height
	binary.BigEndian.PutUint32(b[29:], crc32.ChecksumIEEE(b[12:29]))
	if _, _, err := ReencodeImage(b); !errors.Is(err, ErrImageTooLarge) {
		t.Fatalf("cần ErrImageTooLarge, ra %v", err)
	}
}
