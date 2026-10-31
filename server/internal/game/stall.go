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
	for i := 0; i < n; i++ {
		// `(i+0.5)/n - 0.5` rải tâm các sạp đều trên cung, đối xứng quanh hướng nhắm.
		a := base + span*((float64(i)+0.5)/float64(n)-0.5)
		c := p.Pos.Add(Vec{X: math.Cos(a), Y: math.Sin(a)}.Mul(dist))
		sh := world.Shape{X: c.X - w/2, Y: c.Y - w/2, W: w, H: w}
		// "Sạp không đặt đè lên người (đoạn nào trùng người thì bỏ trống đoạn đó)" — spec.
		//
		// Bỏ trống chứ không đẩy người ra: đẩy người là cho Cô Ba một đòn dịch chuyển cưỡng
		// chế miễn phí, và dãy sạp thành vũ khí thay vì địa hình.
		if g.stallWouldTrap(sh) {
			continue
		}
		g.stallSeq++
		g.stalls = append(g.stalls, tempStall{
			ID: g.stallSeq, Owner: p.ID, Tier: tier, Shape: sh, Until: g.Time + secs,
		})
		placed++
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
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "bay_hang", N: placed, F: float32(secs),
		X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// stallWouldTrap: chỗ này có người hay quái đang đứng không.
func (g *Game) stallWouldTrap(sh world.Shape) bool {
	for _, o := range g.PlayerList() {
		if o.Dead || o.InTransit(g.Time) {
			continue
		}
		if _, hit := sh.PushOut(o.Pos, g.BodyR(o)); hit {
			return true
		}
	}
	for _, m := range g.Monsters {
		if _, hit := sh.PushOut(m.Pos, m.R); hit {
			return true
		}
	}
	return false
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
//  2. **Người chém được bậc đó** — bậc sạp = cảnh giới Cô Ba + 1, nên người hơn cô một
//     cảnh giới chém qua. Chém sạp thì sạp mất ngay như mọi vật cản (§8.2); ở đây chỉ nói
//     về việc THÂN có bị chặn hay không.
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
			if g.CanCutObstacle(p, st.Tier) {
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

// hitStalls: kiếm của `a` quét trúng sạp nào mà họ đủ bậc chém thì sạp đó biến mất ngay.
//
// Chủ sạp KHÔNG chém được sạp của mình: nó đi xuyên người cô, nên kiếm cô cũng đi xuyên nó.
// Nếu không, Cô Ba vừa bày là tự xoá ngay trong nhịp kiếm kế tiếp.
func (g *Game) hitStalls(a *Player) {
	if len(g.stalls) == 0 || !g.canFight(a) {
		return
	}
	keep := g.stalls[:0]
	for _, st := range g.stalls {
		cut := st.Owner != a.ID && g.CanCutObstacle(a, st.Tier)
		if cut {
			c := st.Shape.Center()
			if _, ok := g.swordTouchesCircle(a, c, math.Max(st.Shape.W, st.Shape.H)/2); ok {
				g.emit(Event{Kind: "StallDown", A: a.ID, B: st.ID, X: float32(c.X), Y: float32(c.Y)})
				continue
			}
		}
		keep = append(keep, st)
	}
	g.stalls = keep
}

// StallList cho room gửi xuống client (snapshot) và cho test đọc.
func (g *Game) StallList() []tempStall { return g.stalls }
