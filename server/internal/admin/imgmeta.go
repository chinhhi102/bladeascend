package admin

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// Gỡ siêu dữ liệu khỏi ảnh góp ý — ADMIN_SPEC §7.1 ("xoá toàn bộ EXIF").
//
// Ảnh chụp bằng điện thoại thường mang toạ độ GPS của nơi chụp — tức rất hay là NHÀ người
// chơi. Client đã vẽ lại ảnh qua canvas (thao tác đó vốn bỏ EXIF), nhưng server không bao
// giờ tin client: gói tin có thể gửi thẳng bằng curl. Nên gỡ lại ở đây, bằng thư viện chuẩn.
//
// Cách làm là ĐI QUA CẤU TRÚC tệp và bỏ các khối siêu dữ liệu, không giải mã rồi mã hoá
// lại điểm ảnh: Go chuẩn không mã hoá được WebP, và mã hoá lại JPEG làm ảnh xấu thêm một
// lượt. Đi qua cấu trúc còn là một lớp kiểm định dạng chặt hơn chữ ký đầu tệp: tệp nào
// không đi hết được cấu trúc thì bị từ chối.

// ErrBadImage: tệp có chữ ký ảnh nhưng cấu trúc bên trong hỏng.
var ErrBadImage = errors.New("ảnh hỏng hoặc không đúng định dạng")

// StripImageMetadata trả về bản sao của ảnh đã gỡ siêu dữ liệu, cùng đuôi tệp.
func StripImageMetadata(b []byte) ([]byte, string, error) {
	ext := imageExt(b)
	var out []byte
	var err error
	switch ext {
	case ".jpg":
		out, err = stripJPEG(b)
	case ".png":
		out, err = stripPNG(b)
	case ".webp":
		out, err = stripWebP(b)
	default:
		return nil, "", ErrNotAnImage
	}
	if err != nil {
		return nil, "", err
	}
	return out, ext, nil
}

// ── JPEG ────────────────────────────────────────────────────────────────────────────────

// jpegDrop: khối nào của JPEG phải bỏ.
//
// APP1 (EXIF, XMP), APP2..APP15 (ICC, MPF, Photoshop/IPTC...) và COM. Ngoại lệ duy nhất là
// APP14 "Adobe": nó chỉ ghi một cờ biến đổi màu, không có gì về người chụp, mà thiếu nó thì
// ảnh RGB/CMYK của Adobe giải mã ra sai màu.
func jpegDrop(marker byte, payload []byte) bool {
	if marker == 0xEE && bytes.HasPrefix(payload, []byte("Adobe")) {
		return false
	}
	return (marker >= 0xE1 && marker <= 0xEF) || marker == 0xFE
}

// stripJPEG đi qua các khối của JPEG, kể cả các khối nằm GIỮA các lượt quét của ảnh
// progressive, và cắt mọi thứ sau EOI.
//
// Cắt sau EOI là cố ý: chuẩn MPF của máy ảnh/điện thoại nối thêm ảnh thứ hai (ảnh xem
// trước) SAU EOI, và ảnh đó mang một khối EXIF riêng — có cả GPS.
func stripJPEG(b []byte) ([]byte, error) {
	if len(b) < 4 || b[0] != 0xFF || b[1] != 0xD8 {
		return nil, ErrBadImage
	}
	out := make([]byte, 0, len(b))
	out = append(out, 0xFF, 0xD8)
	for i := 2; i < len(b); {
		if b[i] != 0xFF {
			return nil, ErrBadImage
		}
		// Byte 0xFF lặp lại là byte đệm hợp lệ trước một marker.
		for i < len(b) && b[i] == 0xFF {
			i++
		}
		if i >= len(b) {
			return nil, ErrBadImage
		}
		m := b[i]
		i++
		if m == 0xD9 { // EOI
			return append(out, 0xFF, 0xD9), nil
		}
		next, err := jpegSegment(b, i, m, &out)
		if err != nil {
			return nil, err
		}
		i = next
	}
	return nil, ErrBadImage // hết tệp mà chưa gặp EOI
}

// jpegSegment chép (hoặc bỏ) khối có marker `m` bắt đầu ở `i` (ngay sau byte marker), và
// nếu đó là SOS thì chép luôn dữ liệu nén theo sau. Trả về vị trí đọc tiếp.
func jpegSegment(b []byte, i int, m byte, out *[]byte) (int, error) {
	if m == 0x01 || (m >= 0xD0 && m <= 0xD7) { // marker đứng một mình
		*out = append(*out, 0xFF, m)
		return i, nil
	}
	if m == 0x00 || m == 0xD8 || i+2 > len(b) {
		return 0, ErrBadImage
	}
	n := int(binary.BigEndian.Uint16(b[i:]))
	if n < 2 || i+n > len(b) {
		return 0, ErrBadImage
	}
	seg := b[i : i+n]
	i += n
	if !jpegDrop(m, seg[2:]) {
		*out = append(*out, 0xFF, m)
		*out = append(*out, seg...)
	}
	if m != 0xDA { // chỉ SOS mới có dữ liệu nén theo sau
		return i, nil
	}
	j, ok := skipEntropy(b, i)
	if !ok {
		return 0, ErrBadImage
	}
	*out = append(*out, b[i:j]...)
	return j, nil
}

// skipEntropy tìm chỗ kết thúc của dữ liệu nén bắt đầu từ `i`: vị trí của byte 0xFF mở
// đầu marker kế tiếp (không phải byte nhồi FF00, không phải RSTn).
func skipEntropy(b []byte, i int) (int, bool) {
	for i < len(b) {
		if b[i] != 0xFF {
			i++
			continue
		}
		if i+1 >= len(b) {
			return 0, false
		}
		nx := b[i+1]
		if nx == 0x00 || (nx >= 0xD0 && nx <= 0xD7) {
			i += 2
			continue
		}
		return i, true
	}
	return 0, false
}

// ── PNG ─────────────────────────────────────────────────────────────────────────────────

// pngDrop: các chunk chữ và EXIF. tEXt/zTXt/iTXt là chỗ phần mềm ghi tác giả, phần mềm, có
// khi cả XMP (iTXt "XML:com.adobe.xmp"); eXIf là EXIF; tIME là thời điểm sửa ảnh.
var pngDrop = map[string]bool{"eXIf": true, "tEXt": true, "zTXt": true, "iTXt": true, "tIME": true}

func stripPNG(b []byte) ([]byte, error) {
	const sigLen = 8
	out := make([]byte, 0, len(b))
	out = append(out, b[:sigLen]...)
	i := sigLen
	first := true
	for i+12 <= len(b) {
		n := int(binary.BigEndian.Uint32(b[i:]))
		typ := string(b[i+4 : i+8])
		end := i + 12 + n
		if n < 0 || n > len(b) || end > len(b) {
			return nil, ErrBadImage
		}
		if first && typ != "IHDR" {
			return nil, ErrBadImage
		}
		first = false
		if !pngDrop[typ] {
			out = append(out, b[i:end]...)
		}
		i = end
		if typ == "IEND" {
			return out, nil // bỏ mọi thứ nối sau IEND
		}
	}
	return nil, ErrBadImage
}

// ── WebP ────────────────────────────────────────────────────────────────────────────────

// Cờ trong byte đầu của chunk VP8X (đặc tả WebP container).
const (
	vp8xFlagXMP  = 0x04
	vp8xFlagEXIF = 0x08
)

// stripWebP bỏ chunk "EXIF" và "XMP ", tắt cờ tương ứng trong VP8X và tính lại kích thước
// RIFF. Quên tắt cờ thì một số bộ đọc đi tìm chunk không còn và báo ảnh hỏng.
func stripWebP(b []byte) ([]byte, error) {
	riff := int(binary.LittleEndian.Uint32(b[4:8]))
	if riff < 4 || riff+8 > len(b) {
		return nil, ErrBadImage
	}
	b = b[:riff+8] // bỏ phần nối sau RIFF
	out := make([]byte, 12, len(b))
	copy(out, b[:12])
	vp8x := -1
	hasImage := false
	for i := 12; i < len(b); {
		if i+8 > len(b) {
			return nil, ErrBadImage
		}
		fourcc := string(b[i : i+4])
		n := int(binary.LittleEndian.Uint32(b[i+4:]))
		dataEnd := i + 8 + n
		if n > len(b) || dataEnd > len(b) {
			return nil, ErrBadImage
		}
		next := min(dataEnd+n%2, len(b)) // chunk lẻ có 1 byte đệm; thiếu ở cuối thì bỏ qua
		switch fourcc {
		case "EXIF", "XMP ":
			i = next
			continue
		case "VP8X":
			if n < 10 {
				return nil, ErrBadImage
			}
			vp8x = len(out) + 8
		case "VP8 ", "VP8L", "ANIM":
			hasImage = true
		}
		out = append(out, b[i:dataEnd]...)
		if n%2 == 1 {
			out = append(out, 0)
		}
		i = next
	}
	if !hasImage {
		return nil, ErrBadImage
	}
	if vp8x >= 0 {
		out[vp8x] &^= vp8xFlagEXIF | vp8xFlagXMP
	}
	binary.LittleEndian.PutUint32(out[4:8], uint32(len(out)-8))
	return out, nil
}
