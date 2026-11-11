package game

import (
	"math"

	"bladeascend/server/internal/config"
	"bladeascend/server/internal/world"
)

// Bày Hàng Chặn Lối — dãy sạp tạm của Cô Ba Gánh Hàng (spec v6.10/v6.11 §6.3).
//
// Cô Ba là nhân vật **dựng địa hình**: không gây sát thương, không khống chế trực tiếp, chỉ
// đổi hình dạng bản đồ trong 6 giây. Vai đó chỉ có nghĩa nếu dãy sạp là vật cản THẬT — tức
// chặn thân, chịu luật bậc cảnh giới, và y-sort đúng như mọi vật cản khác.
//
// Vì sao một danh sách riêng chứ không nhét vào `Map.Obstacles`:
//   - `Map` là dữ liệu TĨNH dùng chung cho mọi phòng cùng khu. Thêm sạp vào đó là phòng này
//     dựng sạp thì phòng kia cũng thấy.
//   - `world.Grid` băm sẵn ô theo vật cản lúc nạp bản đồ (`indexObstacle`). Chèn thêm phần
//     tử vào giữa là phải băm lại cả lưới 57 600 ô, mỗi lần Cô Ba bấm kỹ năng.
//   - `netproto.Welcome` gửi danh sách vật cản theo CHỈ SỐ. Danh sách co giãn thì mọi id
//     sau nó lệch đi, và client vẽ sai chỗ.
//
// Nên sạp là một tầng va chạm thứ hai, mỏng: `stallBlocking` chạy sau `resolveBlocking`.

// tempStall là một cái sạp tạm.
type tempStall struct {
	// ID: số riêng, không trùng id vật cản tĩnh — client cần phân biệt để biết cái nào biến
	// mất sau 6 giây và cái nào mọc lại sau 30.
	ID    uint32
	Owner uint32
	// Tier: bậc chém được, chốt LÚC BÀY (cảnh giới Cô Ba + 1). Chốt chứ không tính lại mỗi
	// tick: sạp sống 6 giây, trong 6 giây đó Cô Ba có thể lên cảnh giới hoặc bị chém rớt
	// cảnh giới, và một bức tường tự cứng lên hay mềm đi giữa trận là thứ không ai đoán nổi.
	Tier  int
	Shape world.Shape
	Until float64
}

// castBayHang bày 5 sạp thành vòng cung cách `bay_sap_dist` U theo hướng nhắm, dãy dài
// `bay_sap_len` U, sống `bay_sap_secs` giây; dọc chân dãy rải vỏ chuối rộng `vo_chuoi_w` U.
//
// Vòng cung chứ không đường thẳng: đường thẳng vuông góc với hướng nhắm thì hai đầu dãy xa
// Cô Ba hơn khúc giữa, nên con mồi chạy chéo là lách được đúng chỗ dãy mỏng nhất. Cung có
// tâm ở Cô Ba thì mọi điểm trên dãy cách cô đúng `bay_sap_dist`.
func castBayHang(g *Game, p *Player, dir Vec) {
	n := g.skillInt("bay_sap_count")
	if n <= 0 {
		return
	}
	dist := g.skillRange(p, "bay_sap_dist")
	arc := g.skillRange(p, "bay_sap_len")
	secs := g.skillNum("bay_sap_secs")
	tier := g.Prog.Realm(p.Level) + g.skillInt("bay_sap_tier_bonus")
	if tier > config.TierBorder-1 {
		tier = config.TierBorder - 1
	}
	if maxTier := g.Cfg.Obstacles.MaxTier(); tier > maxTier {
		tier = maxTier
	}
	base := math.Atan2(dir.Y, dir.X)
	// `arc` là ĐỘ DÀI dãy (đơn vị), nên góc mở = độ dài / bán kính.
	span := arc / math.Max(dist, 1e-6)
	// Bề rộng mỗi sạp = độ dài dãy / số sạp, nên 5 sạp phủ kín dãy không hở.
	w := arc / float64(n)
	placed := 0
	var mine []Vec
	for i := 0; i < n; i++ {
		// `(i+0.5)/n - 0.5` rải tâm các sạp đều trên cung, đối xứng quanh hướng nhắm.
		a := base + span*((float64(i)+0.5)/float64(n)-0.5)
		c := p.Pos.Add(Vec{X: math.Cos(a), Y: math.Sin(a)}.Mul(dist))
		sh := world.Shape{X: c.X - w/2, Y: c.Y - w/2, W: w, H: w}
		// v6.48: "Lúc sạp bật lên, địch đứng trên đường sạp bị ĐẨY VĂNG ra ngoài (theo
		// hướng ra xa Cô Ba)". Trước đây chỗ có người thì bỏ trống đoạn đó — dãy sạp hở
		// đúng ngay chỗ con mồi đang đứng, tức kỹ năng tự hỏng khi dùng đúng lúc nhất.
		//
		// Quái thì vẫn bỏ trống: đẩy quái không có trong spec, và nhốt quái trong sạp là
		// cách chắc chắn để sinh ra một con quái kẹt vĩnh viễn.
		if g.stallWouldTrapMonster(sh) {
			continue
		}
		g.pushOffStall(p, sh)
		g.stallSeq++
		g.stalls = append(g.stalls, tempStall{
			ID: g.stallSeq, Owner: p.ID, Tier: tier, Shape: sh, Until: g.Time + secs,
		})
		placed++
		mine = append(mine, c)
		g.emit(Event{Kind: "StallUp", A: p.ID, B: g.stallSeq, N: tier, F: float32(w),
			X: float32(c.X), Y: float32(c.Y)})
	}
	// Vỏ chuối dọc CHÂN dãy, tức phía Cô Ba: con mồi bị dãy sạp chặn, quay lại thì dẫm vỏ.
	// Dùng lại `g.slicks` của Gánh Lướt — cùng một luật "trượt theo quán tính + choáng", nên
	// viết bản thứ hai chỉ là thêm một chỗ nữa phải sửa khi spec đổi số giây.
	peelR := g.skillRange(p, "vo_chuoi_w")
	for i := 0; i < n; i++ {
		a := base + span*((float64(i)+0.5)/float64(n)-0.5)
		at := p.Pos.Add(Vec{X: math.Cos(a), Y: math.Sin(a)}.Mul(dist - peelR))
		g.slicks = append(g.slicks, slick{at: at, until: g.Time + secs, owner: p.ID, r: peelR})
	}
	g.tetherToStalls(p, mine)
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "bay_hang", N: placed, F: float32(secs),
		X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// stallWouldTrapMonster: chỗ này có quái đang đứng không (người thì bị đẩy, xem `pushOffStall`).
func (g *Game) stallWouldTrapMonster(sh world.Shape) bool {
	for _, m := range g.Monsters {
		if _, hit := sh.PushOut(m.Pos, m.R); hit {
			return true
		}
	}
	return false
}

// pushOffStall đẩy địch đang đứng đúng chỗ sắp mọc sạp ra NGOÀI, theo hướng xa Cô Ba.
//
// Đẩy ra xa `p` chứ không đẩy ra theo cạnh gần nhất của sạp: spec ghi rõ hướng, và đẩy
// theo cạnh gần nhất thì một người đứng hơi lệch sẽ bị hất về PHÍA Cô Ba — tức kỹ năng
// chặn đường tự kéo con mồi vào lòng mình.
func (g *Game) pushOffStall(p *Player, sh world.Shape) {
	push := g.skillRange(p, "bay_sap_push")
	for _, o := range g.PlayerList() {
		if o.Dead || o.ID == p.ID || o.InTransit(g.Time) {
			continue
		}
		if _, hit := sh.PushOut(o.Pos, g.BodyR(o)); !hit {
			continue
		}
		g.pushAway(p, o, push)
		g.emit(Event{Kind: "SkillEffect", A: p.ID, B: o.ID, S: "bay_hang_day",
			X: float32(o.Pos.X), Y: float32(o.Pos.Y)})
	}
}

// tetherToStalls buộc địch trong `bay_sap_tether_r` U quanh dãy sạp vào CÁI SẠP GẦN NHẤT
// trong `bay_sap_tether_secs` giây (v6.48 §6.3): họ chỉ đi được trong vòng
// `bay_sap_tether_leash` U quanh sạp đó, kiếm vẫn quay bình thường.
//
// Chọn sạp gần nhất TRONG SỐ SẠP VỪA BÀY, không phải mọi sạp trên bản đồ: hai Cô Ba bày
// chồng lên nhau thì mỗi người buộc theo dãy của mình.
func (g *Game) tetherToStalls(p *Player, mine []Vec) {
	if len(mine) == 0 {
		return
	}
	r := g.skillRange(p, "bay_sap_tether_r")
	leash := g.skillRange(p, "bay_sap_tether_leash")
	secs := g.skillNum("bay_sap_tether_secs")
	for _, o := range g.PlayerList() {
		if o.Dead || o.ID == p.ID || o.InTransit(g.Time) {
			continue
		}
		best, bd := Vec{}, math.Inf(1)
		for _, c := range mine {
			if d := o.Pos.Dist(c); d < bd {
				best, bd = c, d
			}
		}
		if bd > r {
			continue
		}
		o.tetherAt, o.tetherR = best, leash
		g.ApplyStatusFrom(p, o, StTether, secs)
		g.emit(Event{Kind: "SkillEffect", A: p.ID, B: o.ID, S: "buoc_sap", F: float32(secs),
			X: float32(best.X), Y: float32(best.Y)})
	}
}

// updateStalls dọn sạp hết hạn.
func (g *Game) updateStalls() {
	keep := g.stalls[:0]
	for _, st := range g.stalls {
		if g.Time >= st.Until {
			c := st.Shape.Center()
			g.emit(Event{Kind: "StallDown", B: st.ID, X: float32(c.X), Y: float32(c.Y)})
			continue
		}
		keep = append(keep, st)
	}
	g.stalls = keep
}

// stallBlocking đẩy `p` ra khỏi những cái sạp mà **họ** bị chặn bởi.
//
// Ba trường hợp đi xuyên được, mỗi cái là một dòng trong spec:
//  1. **Chủ sạp** (`bay_sap_owner_passes`) — "Cô Ba tự đi xuyên sạp của mình → bỏ chạy chắc
//     chắn trước người ngang cấp". Đây là toàn bộ giá trị tấn công của kỹ năng.
//  2. (v6.48 BỎ: sạp nay **bất hoại**, không ai chém qua, không ai đi xuyên vì cảnh giới
//     cao hơn — xem `hitStalls`.)
//  3. **Shipper đang Gom Đơn** (`StGhost`) — "Shipper đi xuyên được dãy sạp của Cô Ba → hai
//     người khắc nhau" (§6.3 nguyên tắc 4).
func (g *Game) stallBlocking(p *Player, pos Vec, r float64) Vec {
	if len(g.stalls) == 0 {
		return pos
	}
	if p.HasStatus(StGhost, g.Time) {
		return pos
	}
	ownerPasses := g.skillBool("bay_sap_owner_passes")
	for iter := 0; iter < 3; iter++ {
		moved := false
		for i := range g.stalls {
			st := &g.stalls[i]
			if ownerPasses && st.Owner == p.ID {
				continue
			}
			if np, hit := st.Shape.PushOut(pos, r); hit {
				pos, moved = np, true
			}
		}
		if !moved {
			break
		}
	}
	return pos
}

// hitStalls: v6.48 — sạp **BẤT HOẠI**, kiếm quét qua chỉ kêu "cạch" chứ không phá được.
//
// Trước đây ai hơn Cô Ba một cảnh giới là chém bay cả dãy trong một nhịp kiếm, nên kỹ năng
// "dựng địa hình" chỉ hiệu nghiệm với người yếu hơn — đúng những người cô không cần chặn.
// Giờ chỉ có hai đường qua: chính Cô Ba, và Shipper đang Gom Đơn (§6.3 nguyên tắc 4 — hai
// người khắc nhau). Sạp vẫn tự tan sau 6 giây.
func (g *Game) hitStalls(a *Player) {
	if len(g.stalls) == 0 || !g.canFight(a) {
		return
	}
	for i := range g.stalls {
		st := &g.stalls[i]
		if st.Owner == a.ID {
			continue
		}
		c := st.Shape.Center()
		if _, ok := g.swordTouchesCircle(a, c, math.Max(st.Shape.W, st.Shape.H)/2); ok {
			g.emit(Event{Kind: "SkillEffect", A: a.ID, B: st.ID, S: "sap_bat_hoai",
				X: float32(c.X), Y: float32(c.Y)})
		}
	}
}

// StallList cho room gửi xuống client (snapshot) và cho test đọc.
func (g *Game) StallList() []tempStall { return g.stalls }
