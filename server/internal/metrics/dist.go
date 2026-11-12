// Package metrics gom số liệu hiệu năng theo phút / giờ / ngày (ADMIN_SPEC §12).
//
// Ý tưởng cả gói gói trong một câu: KHÔNG lưu từng mẫu. Mỗi chỉ số, mỗi phạm vi, mỗi phút
// chỉ còn một dòng `count · min · max · sum · p95 · histogram`. Histogram là thứ cho phép
// gộp phút → giờ → ngày mà p95 vẫn đúng; lưu từng mẫu thì một ngày đã là hàng triệu dòng và
// p95 của giờ không tính lại được từ p95 của các phút.
package metrics

import (
	"encoding/binary"
	"math"
)

// Histogram log: ô thứ i chứa các giá trị quanh `histMin * histBase^i`.
//
// Vì sao log chứ không chia đều: cùng một cơ chế phải đo tick 0,8 ms lẫn RAM 2 000 MB lẫn
// load 25,1. Chia đều thì hoặc mất hết độ phân giải ở đầu nhỏ, hoặc cần hàng vạn ô. Log cho
// SAI SỐ TƯƠNG ĐỐI cố định 4% ở mọi mức — p95 tick 4 ms sai tối đa 0,16 ms, đủ dùng để
// theo dõi xu hướng, và con số min/max/avg thì vẫn chính xác tuyệt đối vì lưu riêng.
const (
	histBuckets = 512
	histBase    = 1.04
	histMin     = 1e-4
)

var histLogBase = math.Log(histBase)

// Hist là histogram thưa: phần lớn phút chỉ chạm vài ô, nên lưu map thay vì mảng 512 số.
type Hist map[uint16]uint32

func bucketOf(v float64) uint16 {
	if v <= histMin {
		return 0
	}
	i := int(math.Log(v/histMin)/histLogBase) + 1
	if i >= histBuckets {
		i = histBuckets - 1
	}
	return uint16(i)
}

// bucketValue: giá trị ĐẠI DIỆN của một ô (mép trên) — dùng khi đọc p95 ra.
func bucketValue(i uint16) float64 {
	if i == 0 {
		return histMin
	}
	return histMin * math.Pow(histBase, float64(i))
}

// Dist là phân phối của một chỉ số trong một khoảng thời gian.
type Dist struct {
	Count         int64
	Min, Max, Sum float64
	Hist          Hist
}

// Add ghi một mẫu.
func (d *Dist) Add(v float64) {
	if d.Count == 0 || v < d.Min {
		d.Min = v
	}
	if d.Count == 0 || v > d.Max {
		d.Max = v
	}
	d.Count++
	d.Sum += v
	if d.Hist == nil {
		d.Hist = Hist{}
	}
	d.Hist[bucketOf(v)]++
}

// Merge gộp một phân phối khác vào — phép toán duy nhất cần cho cuộn phút → giờ → ngày.
func (d *Dist) Merge(o Dist) {
	if o.Count == 0 {
		return
	}
	if d.Count == 0 || o.Min < d.Min {
		d.Min = o.Min
	}
	if d.Count == 0 || o.Max > d.Max {
		d.Max = o.Max
	}
	d.Count += o.Count
	d.Sum += o.Sum
	if d.Hist == nil {
		d.Hist = Hist{}
	}
	for b, n := range o.Hist {
		d.Hist[b] += n
	}
}

// Avg trả trung bình; 0 mẫu thì 0.
func (d Dist) Avg() float64 {
	if d.Count == 0 {
		return 0
	}
	return d.Sum / float64(d.Count)
}

// P95 đọc từ histogram. Kẹp vào [Min, Max] vì ô là mép trên: một khoảng chỉ có đúng một mẫu
// 3,0 ms không được phép hiện "p95 3,12".
func (d Dist) P95() float64 {
	if d.Count == 0 {
		return 0
	}
	target := float64(d.Count) * 0.95
	var seen float64
	for i := uint16(0); i < histBuckets; i++ {
		seen += float64(d.Hist[i])
		if seen >= target {
			return math.Min(math.Max(bucketValue(i), d.Min), d.Max)
		}
	}
	return d.Max
}

// Counter là bản dùng TRONG VÒNG TICK của Dist: mảng cố định, không map, không cấp phát.
//
// Vì sao không dùng thẳng Dist: `Dist.Add` ghi vào map, mà map trong vòng tick nghĩa là
// băm + có thể cấp phát lại — spec §13.4 cấm cấp phát ở đường nóng, và trần của §12 là
// "thu gom không làm tick tăng quá 0,1 ms". Mảng 512 ô = 2 KB mỗi bộ đếm, cộng vào tick
// đúng một phép so sánh và một phép tăng.
type Counter struct {
	count         int64
	min, max, sum float64
	buckets       [histBuckets]uint32
}

// Add ghi một mẫu. Đây là hàm DUY NHẤT của gói này chạy trong vòng tick.
func (c *Counter) Add(v float64) {
	if c.count == 0 || v < c.min {
		c.min = v
	}
	if c.count == 0 || v > c.max {
		c.max = v
	}
	c.count++
	c.sum += v
	c.buckets[bucketOf(v)]++
}

// Drain trả về phân phối kể từ lần vét trước rồi đặt lại bộ đếm.
func (c *Counter) Drain() Dist {
	d := Dist{Count: c.count, Min: c.min, Max: c.max, Sum: c.sum}
	if c.count > 0 {
		d.Hist = Hist{}
		for i, n := range c.buckets {
			if n > 0 {
				d.Hist[uint16(i)] = n
			}
		}
	}
	*c = Counter{}
	return d
}

// EncodeHist đóng gói histogram thưa thành cặp (ô, số lần) mã varint, để nằm vừa một cột
// BLOB. Một phút tick điển hình chạm ~30 ô → khoảng 90 byte.
func EncodeHist(h Hist) []byte {
	if len(h) == 0 {
		return nil
	}
	keys := make([]uint16, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	// Sắp xếp để byte ra ổn định (test so sánh được, và nén tốt hơn).
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	buf := make([]byte, 0, len(keys)*4)
	var tmp [binary.MaxVarintLen64]byte
	for _, k := range keys {
		n := binary.PutUvarint(tmp[:], uint64(k))
		buf = append(buf, tmp[:n]...)
		n = binary.PutUvarint(tmp[:], uint64(h[k]))
		buf = append(buf, tmp[:n]...)
	}
	return buf
}

// DecodeHist đọc ngược EncodeHist. Dữ liệu hỏng → trả về phần đọc được, không panic: một
// dòng số liệu hỏng không đáng làm sập trang quản trị.
func DecodeHist(b []byte) Hist {
	h := Hist{}
	for len(b) > 0 {
		k, n := binary.Uvarint(b)
		if n <= 0 {
			break
		}
		b = b[n:]
		c, n := binary.Uvarint(b)
		if n <= 0 {
			break
		}
		b = b[n:]
		h[uint16(k)] += uint32(c)
	}
	return h
}
