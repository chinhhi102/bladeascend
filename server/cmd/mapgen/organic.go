package main

import "math"

// Vùng địa hình hình hữu cơ (spec §8.6 mục 2).
//
// Luật cứng của spec: bản đồ hiển thị **không được có hình chữ nhật thẳng cạnh hay hình
// tròn hoàn hảo nào lộ ra**. Hình học đơn giản chỉ dành cho va chạm và chế độ debug.
//
// Cách làm: lấy bán kính thay đổi theo góc `r(θ) = r0 × (1 + 0.25 × noise(θ))` rồi làm mượt
// bằng Chaikin hai lần. Quan trọng là **server dùng đúng đa giác này** cho logic (§8.6), nên
// nó phải nằm trong file bản đồ chứ không phải do client tự bịa ra lúc vẽ — nếu không, chỗ
// nhìn như bùn lại không làm chậm chân, hoặc ngược lại.

// wobble là nhiễu tuần hoàn theo góc: cộng ba sóng sin lệch pha, chu kỳ khác nhau. Dùng sóng
// thay vì Perlin vì ở đây chỉ cần méo theo MỘT chiều (góc) và phải khép kín vòng tròn —
// Perlin 1D tuần hoàn thì phải tự lo chỗ nối, còn tổng các sóng nguyên chu kỳ thì khép kín
// sẵn.
func wobble(theta float64, seed int64) float64 {
	s := float64(seed%97) * 0.7
	return 0.55*math.Sin(3*theta+s) + 0.30*math.Sin(5*theta+s*1.7) + 0.15*math.Sin(8*theta+s*2.3)
}

// organicPoly dựng một đa giác méo quanh (cx, cy) bán kính trung bình r0.
// `verts` là số đỉnh trước khi làm mượt (spec: 24–40 sau khi mượt).
// `squash` ép dẹt theo trục y (1 = tròn đều, 0.6 = bẹt).
func organicPoly(cx, cy, r0 float64, verts int, squash float64, seed int64) []float64 {
	if verts < 8 {
		verts = 8
	}
	pts := make([]float64, 0, verts*2)
	for i := 0; i < verts; i++ {
		th := (float64(i) / float64(verts)) * 2 * math.Pi
		r := r0 * (1 + 0.25*wobble(th, seed))
		pts = append(pts, cx+math.Cos(th)*r, cy+math.Sin(th)*r*squash)
	}
	return chaikin(chaikin(pts))
}

// chaikin cắt góc: mỗi cạnh thay bằng hai điểm ở 1/4 và 3/4. Hai lần là đủ mượt mà vẫn giữ
// được dáng méo; nhiều hơn thì đa giác bắt đầu co về hình tròn — đúng thứ spec cấm.
func chaikin(p []float64) []float64 {
	n := len(p) / 2
	if n < 3 {
		return p
	}
	out := make([]float64, 0, n*4)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		ax, ay := p[i*2], p[i*2+1]
		bx, by := p[j*2], p[j*2+1]
		out = append(out, ax+(bx-ax)*0.25, ay+(by-ay)*0.25, ax+(bx-ax)*0.75, ay+(by-ay)*0.75)
	}
	return out
}

// polyBounds trả hộp bao của đa giác, để kiểm tra nằm trong lề bản đồ.
func polyBounds(p []float64) rect {
	r := rect{p[0], p[1], p[0], p[1]}
	for i := 2; i < len(p); i += 2 {
		r.x0 = math.Min(r.x0, p[i])
		r.x1 = math.Max(r.x1, p[i])
		r.y0 = math.Min(r.y0, p[i+1])
		r.y1 = math.Max(r.y1, p[i+1])
	}
	return r
}
