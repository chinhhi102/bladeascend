package main

import "math"

// Mẫu cụm RIÊNG cho từng khu — spec §8.1b mục 6, chủ dự án chốt 31/10/2026.
//
// Lượt bố cục 2 dùng MỘT mẫu chung cho cả năm khu: những dải song song cắt theo đường bao
// tròn. Số đo thì đạt (mật độ, khoảng thở, liên thông), nhưng nhìn thì năm khu giống hệt
// nhau — chỉ khác màu nền và màu vật cản. Trúc Lâm Mê Cung không ra mê cung, Phàm Trấn
// không ra thị trấn, Linh Sơn Cốc không ra hẻm núi.
//
// Mỗi khu giờ có hàm dựng riêng, nhận cùng một hộp (tâm + bán kính) và tự quyết hình dáng
// bên trong. Luật chung vẫn áp cho tất cả:
//   - khối trong cùng một cụm ĐƯỢC chạm nhau (hành lang chỉ áp giữa hai cụm);
//   - không khối nào dài quá `maxRunLen` = 800 (§8.1b mục 2);
//   - xen kẽ bậc cao / bậc thấp để khe giữa hai khối kề nhau chỉ cần `corridor_min` của bậc
//     THẤP HƠN — xem `placeCluster` để biết vì sao đó là chỗ quyết định mật độ.

// clusterKit gom những thứ mọi mẫu đều cần, để chữ ký hàm khỏi dài lê thê.
type clusterKit struct {
	cx, cy, r float64
	core, rim string
	// gap: khe tối thiểu giữa hai khối trong cụm, lấy theo bậc THẤP HƠN.
	gap float64
}

// rockField — Linh Sơn Cốc: bãi đá tròn/méo, cỡ lẫn lộn, xoay ngẫu nhiên.
//
// Đá không xếp hàng. Rải theo vòng đồng tâm với bước góc lệch nhau để không lộ ra vòng
// tròn, mỗi tảng một cỡ và một tỉ lệ bẹp riêng.
func (g *gen) rockField(k clusterKit) {
	rings := 2 + g.b.rng.Intn(2)
	for ri := 0; ri < rings; ri++ {
		rr := k.r * (0.3 + 0.6*float64(ri)/float64(max(rings-1, 1)))
		n := 3 + ri*3 + g.b.rng.Intn(3)
		off := g.b.rnd(0, 2*math.Pi)
		for i := 0; i < n; i++ {
			a := off + float64(i)/float64(n)*2*math.Pi + g.b.rnd(-0.25, 0.25)
			// Tảng ở vòng NGOÀI nhỏ hơn: bãi đá có lõi to, rìa vụn dần.
			size := k.r * g.b.rnd(0.18, 0.42) * (1.1 - 0.35*float64(ri)/float64(max(rings-1, 1)))
			squash := g.b.rnd(0.65, 1.45) // méo theo một trục → không tảng nào tròn đều
			w, h := size, size*squash
			kind := k.core
			if ri == rings-1 {
				kind = k.rim
			}
			g.b.obstacleForce(kind,
				k.cx+math.Cos(a)*rr-w/2, k.cy+math.Sin(a)*rr-h/2, w, h)
		}
	}
}

// bambooClumps — Trúc Lâm: lùm tre DÀI và CONG, xếp so le dọc theo lối đi.
//
// Lùm tre là nhiều khúc ngắn đặt dọc một cung tròn, nên nó cong chứ không thẳng; các lùm
// lệch pha nhau nên lối đi giữa chúng ngoằn ngoèo — đúng "lối đi quanh co giữa các lùm tre".
func (g *gen) bambooClumps(k clusterKit) {
	lanes := 2 + g.b.rng.Intn(2)
	// Bán kính cong của lùm: lớn hơn cụm nhiều lần thì cung mới thoải, không cuộn thành vòng.
	bend := k.r * g.b.rnd(2.2, 4.0)
	if g.b.rng.Intn(2) == 0 {
		bend = -bend
	}
	base := g.b.rnd(0, 2*math.Pi)
	thick := g.b.rnd(170, 260)
	for l := 0; l < lanes; l++ {
		// So le: lùm kế tiếp lùi đi một phần ba chiều dài, nên không có đoạn nào hai lùm
		// cùng bắt đầu — người chơi luôn phải lách.
		shift := g.b.rnd(0.25, 0.45) * float64(l%2*2-1)
		dist := k.r * (0.35 + 0.55*float64(l)/float64(max(lanes-1, 1)))
		kind := k.core
		if l%2 == 1 {
			kind = k.rim
		}
		segs := 4 + g.b.rng.Intn(3)
		for i := 0; i < segs; i++ {
			t := (float64(i)/float64(segs-1) - 0.5 + shift) * 1.6
			// Điểm trên cung: lệch ngang `t·r`, lệch dọc theo độ cong.
			dx := t * k.r
			dy := dx * dx / (2 * bend)
			a := base
			x := k.cx + dx*math.Cos(a) - (dist+dy)*math.Sin(a)
			y := k.cy + dx*math.Sin(a) + (dist+dy)*math.Cos(a)
			segLen := math.Min(k.r*0.5, maxRunLen)
			g.b.obstacleForce(kind, x-segLen/2, y-thick/2, segLen, thick)
		}
	}
}

// townBlock — Phàm Trấn: nhà chữ nhật có SÂN ở giữa và CỬA ra một phía.
//
// Bốn bức tường bao quanh một khoảng trống, chừa một đoạn làm cửa. Người chơi chui vào sân
// nấp được, nhưng chỉ có một lối ra — đó là chỗ phục kích, và là thứ một ô vuông đặc không
// bao giờ tạo ra.
func (g *gen) townBlock(k clusterKit) {
	// Hai–ba căn trong một cụm, mỗi căn một cỡ.
	houses := 2 + g.b.rng.Intn(2)
	for h := 0; h < houses; h++ {
		a := float64(h)/float64(houses)*2*math.Pi + g.b.rnd(0, 0.6)
		d := k.r * g.b.rnd(0, 0.45)
		cx, cy := k.cx+math.Cos(a)*d, k.cy+math.Sin(a)*d
		w := math.Min(k.r*g.b.rnd(0.55, 0.95), maxRunLen)
		ht := math.Min(k.r*g.b.rnd(0.45, 0.8), maxRunLen)
		wall := g.b.rnd(130, 190)
		kind := k.core
		if h%2 == 1 {
			kind = k.rim
		}
		door := g.b.rng.Intn(4) // cạnh nào có cửa
		// Cửa rộng hơn hành lang tối thiểu để chắc chắn chui lọt.
		gapW := math.Max(k.gap*1.3, 220)
		for side := 0; side < 4; side++ {
			horiz := side%2 == 0
			long := w
			if !horiz {
				long = ht
			}
			if side != door {
				g.houseWall(cx, cy, w, ht, wall, side, horiz, long, kind)
				continue
			}
			// Cạnh có cửa: hai đoạn ngắn hai bên, chừa khoảng giữa.
			part := (long - gapW) / 2
			if part < 60 {
				continue // căn quá nhỏ: bỏ hẳn cạnh này, coi như nhà hở một mặt
			}
			g.houseWall(cx, cy, w, ht, wall, side, horiz, part, kind)
			g.houseWallAt(cx, cy, w, ht, wall, side, horiz, part, long/2-part/2, kind)
			g.houseWallAt(cx, cy, w, ht, wall, side, horiz, part, -(long/2 - part/2), kind)
		}
	}
}

// houseWall đặt nguyên một cạnh nhà.
func (g *gen) houseWall(cx, cy, w, h, wall float64, side int, horiz bool, long float64, kind string) {
	g.houseWallAt(cx, cy, w, h, wall, side, horiz, long, 0, kind)
}

// houseWallAt đặt một đoạn tường dài `long` trên cạnh `side`, lệch `off` dọc theo cạnh.
func (g *gen) houseWallAt(cx, cy, w, h, wall float64, side int, horiz bool, long, off float64, kind string) {
	switch side {
	case 0: // trên
		g.b.obstacleForce(kind, cx-long/2+off, cy-h/2, long, wall)
	case 2: // dưới
		g.b.obstacleForce(kind, cx-long/2+off, cy+h/2-wall, long, wall)
	case 1: // phải
		g.b.obstacleForce(kind, cx+w/2-wall, cy-long/2+off, wall, long)
	default: // trái
		g.b.obstacleForce(kind, cx-w/2, cy-long/2+off, wall, long)
	}
}

// palaceYard — Thiên Đình: điện có TƯỜNG BAO và sân trong, ba mặt kín một mặt mở.
//
// Khác `townBlock` ở quy mô và ở chỗ có điện ở giữa: tường bao rộng gần hết cụm, bên trong
// là một khối điện đặc nhỏ hơn, giữa hai lớp là sân.
func (g *gen) palaceYard(k clusterKit) {
	w := math.Min(k.r*1.7, maxRunLen)
	h := math.Min(k.r*1.5, maxRunLen)
	wall := g.b.rnd(150, 200)
	gate := math.Max(k.gap*1.4, 260)
	open := g.b.rng.Intn(4)
	for side := 0; side < 4; side++ {
		horiz := side%2 == 0
		long := w
		if !horiz {
			long = h
		}
		if side != open {
			g.houseWall(k.cx, k.cy, w, h, wall, side, horiz, long, k.core)
			continue
		}
		part := (long - gate) / 2
		if part < 60 {
			continue
		}
		g.houseWallAt(k.cx, k.cy, w, h, wall, side, horiz, part, long/2-part/2, k.core)
		g.houseWallAt(k.cx, k.cy, w, h, wall, side, horiz, part, -(long/2 - part/2), k.core)
	}
	// Điện ở giữa sân: nhỏ hơn tường bao, chừa sân vòng quanh ≥ hành lang.
	pw := w - 2*(wall+k.gap)
	ph := h - 2*(wall+k.gap)
	if pw > 160 && ph > 160 {
		g.b.obstacleForce(k.rim, k.cx-pw/2, k.cy-ph/2, pw*0.7, ph*0.7)
	}
}
