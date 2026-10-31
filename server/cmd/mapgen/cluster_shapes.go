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

// rockField — Linh Sơn Cốc và bờ Băng Nguyệt Hồ: bãi đá ÍT MẢNH, cỡ lệch hẳn nhau.
//
// Bản lượt 3 rải 2–3 vòng đồng tâm, mỗi vòng 3–9 tảng — ra tới 20 tảng một bãi, và phần lớn
// là mảnh li ti. Nhìn thành một đám sỏi đều tăm tắp chứ không ra bãi đá.
//
// Giờ: 1–2 KHỐI LỚN làm lõi, cộng 2–4 khối nhỏ nép quanh. Chênh lệch cỡ giữa lớn và nhỏ ít
// nhất gấp đôi, nên mắt đọc ra "một tảng to với mấy hòn kê" chứ không ra "nhiều hòn bằng
// nhau". Không mảnh nào nhỏ hơn `minRockSize` — dưới cỡ đó nó chỉ là nhiễu thị giác, vướng
// chân mà không che được ai.
func (g *gen) rockField(k clusterKit) {
	// Khối lớn: 1 hoặc 2, đặt lệch tâm một chút để bãi không đối xứng.
	bigN := 1 + g.b.rng.Intn(2)
	placed := make([][3]float64, 0, 6) // x, y, bán kính thô — để khối nhỏ nép vào
	for i := 0; i < bigN; i++ {
		a := g.b.rnd(0, 2*math.Pi)
		d := k.r * g.b.rnd(0, 0.22)
		size := k.r * g.b.rnd(0.70, 1.05)
		squash := g.b.rnd(0.72, 1.38) // méo theo một trục → không tảng nào tròn đều
		w, h := size, size*squash
		x, y := k.cx+math.Cos(a)*d, k.cy+math.Sin(a)*d
		g.b.obstacleForce(k.core, x-w/2, y-h/2, w, h)
		placed = append(placed, [3]float64{x, y, math.Max(w, h) / 2})
	}
	// Khối nhỏ: nép vào SƯỜN khối lớn, cỡ bằng 0,3–0,5 khối lớn.
	for i := 0; i < 2+g.b.rng.Intn(3); i++ {
		host := placed[g.b.rng.Intn(len(placed))]
		a := g.b.rnd(0, 2*math.Pi)
		size := math.Max(k.r*g.b.rnd(0.22, 0.38), minRockSize)
		squash := g.b.rnd(0.7, 1.4)
		w, h := size, size*squash
		// Đặt ở mép khối lớn, chồng vào khoảng 1/4 để trông như kê vào nhau, không rời rạc.
		d := host[2] + size*0.3
		g.b.obstacleForce(k.rim, host[0]+math.Cos(a)*d-w/2, host[1]+math.Sin(a)*d-h/2, w, h)
	}
}

// minRockSize: mảnh đá nhỏ hơn thế này chỉ là nhiễu — vướng chân mà không che được ai.
const minRockSize = 150.0

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
//
// Hai cạnh DỌC thụt vào `wall` mỗi đầu để góc nhà **chạm chứ không chồng** lên cạnh ngang.
//
// Chồng mép ở góc nghe thì vô hại, nhưng nó phá phép đẩy ra khỏi vật cản: `resolveBlocking`
// lặp 3 lượt, và với các khối RỜI nhau thì nó hội tụ về đúng một kết quả, còn với các khối
// CHỒNG nhau thì kết quả phụ thuộc số lượt. Gói nhiều nhà vào một chỗ là một điểm có cả
// chục cặp chồng mép, 3 lượt không đủ, và client tính ra khác server — bài đối chiếu bắt
// được **lệch 12,98 đơn vị** ở Phàm Trấn, đủ để người chơi thấy giật nảy.
func (g *gen) houseWallAt(cx, cy, w, h, wall float64, side int, horiz bool, long, off float64, kind string) {
	switch side {
	case 0: // trên
		g.b.obstacleForce(kind, cx-long/2+off, cy-h/2, long, wall)
	case 2: // dưới
		g.b.obstacleForce(kind, cx-long/2+off, cy+h/2-wall, long, wall)
	case 1: // phải
		if ln := long - 2*wall; ln > 40 {
			g.b.obstacleForce(kind, cx+w/2-wall, cy-ln/2+off, wall, ln)
		}
	default: // trái
		if ln := long - 2*wall; ln > 40 {
			g.b.obstacleForce(kind, cx-w/2, cy-ln/2+off, wall, ln)
		}
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
