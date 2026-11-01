package admin

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

// Gỡ siêu dữ liệu ảnh góp ý — ADMIN_SPEC §7.1.
//
// Fixture dựng bằng tay từng byte, vì điều cần kiểm là ĐÚNG khối EXIF/GPS biến mất mà ảnh
// vẫn giải mã được. JPEG và PNG được giải mã lại bằng thư viện chuẩn để chắc là không hỏng.

// gpsMarker là chuỗi nhận diện nằm trong khối GPS giả — sau khi gỡ, nó không được còn.
const gpsMarker = "GPS-21.0285N-105.8542E"

func testImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 16, 12))
	for y := 0; y < 12; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 16), uint8(y * 20), 128, 255})
		}
	}
	return img
}

func testJPEG(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, testImage(), &jpeg.Options{Quality: 80}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func testPNG(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, testImage()); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// exifGPSBlock dựng nội dung APP1 "Exif" kiểu big-endian: IFD0 có một mục GPSInfo trỏ tới
// GPS IFD, trong đó có GPSLatitudeRef = 'N' và một chuỗi chứa gpsMarker để dễ dò.
func exifGPSBlock() []byte {
	var t bytes.Buffer
	w := func(v any) { _ = binary.Write(&t, binary.BigEndian, v) }
	t.WriteString("MM\x00\x2A")
	w(uint32(8)) // IFD0 ở offset 8
	w(uint16(1))
	w([]uint16{0x8825, 4}) // GPSInfo, LONG
	w([]uint32{1, 26})     // GPS IFD ở offset 26
	w(uint32(0))
	w(uint16(2))
	w([]uint16{0x0001, 2}) // GPSLatitudeRef, ASCII
	w([]uint32{2, 0x4E000000})
	w([]uint16{0x001B, 7}) // GPSProcessingMethod, UNDEFINED
	w([]uint32{uint32(len(gpsMarker)), 56})
	w(uint32(0))
	t.WriteString(gpsMarker)
	return append([]byte("Exif\x00\x00"), t.Bytes()...)
}

func jpegSeg(marker byte, payload []byte) []byte {
	s := []byte{0xFF, marker, 0, 0}
	binary.BigEndian.PutUint16(s[2:], uint16(len(payload)+2))
	return append(s, payload...)
}

// insertAfterSOI chèn các khối ngay sau SOI (chỗ điện thoại đặt EXIF).
func insertAfterSOI(j []byte, segs ...[]byte) []byte {
	out := append([]byte{}, j[:2]...)
	for _, s := range segs {
		out = append(out, s...)
	}
	return append(out, j[2:]...)
}

func TestStripJPEGRemovesGPSExif(t *testing.T) {
	src := insertAfterSOI(testJPEG(t),
		jpegSeg(0xE1, exifGPSBlock()),
		jpegSeg(0xE1, []byte("http://ns.adobe.com/xap/1.0/\x00<x:xmpmeta>XMP-SECRET</x:xmpmeta>")),
		jpegSeg(0xE2, []byte("ICC_PROFILE\x00ICC-SECRET")),
		jpegSeg(0xED, []byte("Photoshop 3.0\x00IPTC-SECRET")),
		jpegSeg(0xFE, []byte("COMMENT-SECRET")),
		jpegSeg(0xEE, []byte("Adobe\x00\x64\x00\x00\x00\x00\x01")),
	)
	// MPF: điện thoại nối ảnh xem trước (có EXIF riêng) SAU EOI.
	src = append(src, insertAfterSOI(testJPEG(t), jpegSeg(0xE1, exifGPSBlock()))...)
	if !bytes.Contains(src, []byte(gpsMarker)) {
		t.Fatal("fixture không có GPS — bài kiểm không kiểm gì cả")
	}

	out, ext, err := StripImageMetadata(src)
	if err != nil {
		t.Fatal(err)
	}
	if ext != ".jpg" {
		t.Fatalf("ext = %q", ext)
	}
	for _, secret := range []string{gpsMarker, "Exif", "XMP-SECRET", "ICC-SECRET", "IPTC-SECRET", "COMMENT-SECRET"} {
		if bytes.Contains(out, []byte(secret)) {
			t.Errorf("còn sót %q sau khi gỡ", secret)
		}
	}
	if !bytes.Contains(out, []byte("Adobe")) {
		t.Error("APP14 Adobe (cờ màu, không có thông tin cá nhân) không được bỏ")
	}
	if !bytes.HasSuffix(out, []byte{0xFF, 0xD9}) {
		t.Error("phải kết thúc đúng ở EOI, cắt mọi thứ nối sau")
	}
	if _, err := jpeg.Decode(bytes.NewReader(out)); err != nil {
		t.Fatalf("ảnh sau khi gỡ không giải mã được: %v", err)
	}
}

// Dữ liệu nén có byte nhồi FF00 và marker RSTn — không được hiểu nhầm là khối mới.
func TestStripJPEGEntropyStuffingAndRestart(t *testing.T) {
	scan := []byte{0x12, 0xFF, 0x00, 0x34, 0xFF, 0xD0, 0x56, 0xFF, 0xD1, 0x78}
	src := []byte{0xFF, 0xD8}
	src = append(src, jpegSeg(0xDB, make([]byte, 65))...)
	src = append(src, jpegSeg(0xDA, []byte{1, 1, 0, 0, 63, 0})...)
	src = append(src, scan...)
	// Khối giữa hai lượt quét (ảnh progressive), kể cả một COM phải bị bỏ.
	src = append(src, jpegSeg(0xFE, []byte("MID-SECRET"))...)
	src = append(src, jpegSeg(0xDA, []byte{1, 1, 0, 0, 63, 0})...)
	src = append(src, 0x9A, 0xFF, 0xD9)

	out, _, err := StripImageMetadata(src)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(out, []byte("MID-SECRET")) {
		t.Error("COM giữa hai lượt quét còn sót")
	}
	if !bytes.Contains(out, scan) {
		t.Error("dữ liệu nén bị cắt sai ở byte nhồi/RST")
	}
	want := len(src) - len(jpegSeg(0xFE, []byte("MID-SECRET")))
	if len(out) != want {
		t.Errorf("dài %d, muốn %d", len(out), want)
	}
}

func TestStripJPEGRejectsBroken(t *testing.T) {
	good := testJPEG(t)
	cases := map[string][]byte{
		"chỉ có chữ ký":   {0xFF, 0xD8, 0xFF},
		"thiếu EOI":       good[:len(good)-2],
		"độ dài khối láo": {0xFF, 0xD8, 0xFF, 0xE1, 0xFF, 0xFF, 1, 2},
		"rác sau khối":    {0xFF, 0xD8, 0xFF, 0xE0, 0, 4, 0, 0, 'h', 't', 'm', 'l'},
	}
	for name, b := range cases {
		if _, _, err := StripImageMetadata(b); err == nil {
			t.Errorf("%s: phải bị từ chối", name)
		}
	}
}

func pngChunk(typ string, data []byte) []byte {
	c := make([]byte, 8, 12+len(data))
	binary.BigEndian.PutUint32(c, uint32(len(data)))
	copy(c[4:], typ)
	c = append(c, data...)
	return binary.BigEndian.AppendUint32(c, crc32.ChecksumIEEE(c[4:]))
}

func TestStripPNGRemovesTextAndExif(t *testing.T) {
	p := testPNG(t)
	at := 8 + 25 // sau chữ ký và chunk IHDR
	extra := bytes.Join([][]byte{
		pngChunk("eXIf", exifGPSBlock()[6:]),
		pngChunk("tEXt", []byte("Author\x00TEXT-SECRET")),
		pngChunk("zTXt", []byte("Comment\x00\x00ZTXT-SECRET")),
		pngChunk("iTXt", []byte("XML:com.adobe.xmp\x00\x00\x00\x00\x00ITXT-SECRET")),
		pngChunk("tIME", []byte{0x07, 0xEA, 11, 1, 12, 0, 0}),
	}, nil)
	src := append(append(append([]byte{}, p[:at]...), extra...), p[at:]...)
	src = append(src, []byte("TRAILER-SECRET")...)

	out, ext, err := StripImageMetadata(src)
	if err != nil {
		t.Fatal(err)
	}
	if ext != ".png" {
		t.Fatalf("ext = %q", ext)
	}
	for _, secret := range []string{gpsMarker, "TEXT-SECRET", "ZTXT-SECRET", "ITXT-SECRET", "tIME", "TRAILER-SECRET"} {
		if bytes.Contains(out, []byte(secret)) {
			t.Errorf("còn sót %q", secret)
		}
	}
	if !bytes.Equal(out, p) {
		t.Error("gỡ xong phải ra đúng PNG gốc")
	}
	if _, err := png.Decode(bytes.NewReader(out)); err != nil {
		t.Fatalf("PNG sau khi gỡ không giải mã được: %v", err)
	}
}

func TestStripPNGRejectsBroken(t *testing.T) {
	sig := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	good := testPNG(t)
	cases := map[string][]byte{
		"chỉ chữ ký":           sig,
		"chunk đầu không IHDR": append(append([]byte{}, sig...), pngChunk("tEXt", []byte("a"))...),
		"thiếu IEND":           good[:len(good)-12],
		"chữ ký + số 0":        append(append([]byte{}, sig...), make([]byte, 64)...),
	}
	for name, b := range cases {
		if _, _, err := StripImageMetadata(b); err == nil {
			t.Errorf("%s: phải bị từ chối", name)
		}
	}
}

func riffChunk(fourcc string, data []byte) []byte {
	c := make([]byte, 8, 9+len(data))
	copy(c, fourcc)
	binary.LittleEndian.PutUint32(c[4:], uint32(len(data)))
	c = append(c, data...)
	if len(data)%2 == 1 {
		c = append(c, 0)
	}
	return c
}

func webpFile(chunks ...[]byte) []byte {
	body := append([]byte("WEBP"), bytes.Join(chunks, nil)...)
	out := []byte("RIFF\x00\x00\x00\x00")
	binary.LittleEndian.PutUint32(out[4:], uint32(len(body)))
	return append(out, body...)
}

func TestStripWebPRemovesExifXMP(t *testing.T) {
	vp8x := make([]byte, 10)
	vp8x[0] = 0x20 | 0x10 | vp8xFlagEXIF | vp8xFlagXMP       // ICC + alpha + EXIF + XMP
	vp8l := []byte{0x2F, 0x00, 0x00, 0x00, 0x00, 0x01, 0x02} // độ dài lẻ → có byte đệm
	src := webpFile(
		riffChunk("VP8X", vp8x),
		riffChunk("ICCP", []byte("icc")),
		riffChunk("VP8L", vp8l),
		riffChunk("EXIF", exifGPSBlock()),
		riffChunk("XMP ", []byte("<x:xmpmeta>XMP-SECRET</x:xmpmeta>")),
	)
	src = append(src, []byte("TRAILER-SECRET")...)

	out, ext, err := StripImageMetadata(src)
	if err != nil {
		t.Fatal(err)
	}
	if ext != ".webp" {
		t.Fatalf("ext = %q", ext)
	}
	for _, secret := range []string{gpsMarker, "XMP-SECRET", "EXIF", "XMP ", "TRAILER-SECRET"} {
		if bytes.Contains(out, []byte(secret)) {
			t.Errorf("còn sót %q", secret)
		}
	}
	want := webpFile(riffChunk("VP8X", append([]byte{0x30}, vp8x[1:]...)),
		riffChunk("ICCP", []byte("icc")), riffChunk("VP8L", vp8l))
	if !bytes.Equal(out, want) {
		t.Errorf("kết quả sai:\n got %x\nwant %x", out, want)
	}
	if got := binary.LittleEndian.Uint32(out[4:8]); int(got) != len(out)-8 {
		t.Errorf("kích thước RIFF %d, tệp thật %d", got, len(out)-8)
	}
}

// WebP đơn giản (không VP8X) không có siêu dữ liệu — gỡ xong phải nguyên vẹn.
func TestStripWebPSimpleUnchanged(t *testing.T) {
	src := webpFile(riffChunk("VP8 ", []byte{0x10, 0x02, 0x00, 0x9D, 0x01, 0x2A}))
	out, _, err := StripImageMetadata(src)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, src) {
		t.Error("WebP không có siêu dữ liệu bị đổi")
	}
}

func TestStripWebPRejectsBroken(t *testing.T) {
	cases := map[string][]byte{
		"RIFF dài hơn tệp":  append([]byte("RIFF\xff\x00\x00\x00WEBP"), riffChunk("VP8 ", []byte{1, 2})...),
		"chunk dài hơn tệp": webpFile([]byte("VP8 \xff\x00\x00\x00")),
		"không có ảnh":      webpFile(riffChunk("EXIF", exifGPSBlock())),
		"VP8X cụt":          webpFile(riffChunk("VP8X", []byte{0}), riffChunk("VP8L", []byte{1})),
	}
	for name, b := range cases {
		if _, _, err := StripImageMetadata(b); err == nil {
			t.Errorf("%s: phải bị từ chối", name)
		}
	}
}

func TestStripRejectsNonImage(t *testing.T) {
	if _, _, err := StripImageMetadata([]byte("<html><script>alert(1)</script>")); err != ErrNotAnImage {
		t.Fatalf("muốn ErrNotAnImage, được %v", err)
	}
}
