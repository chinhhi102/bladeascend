package game

import (
	"math"

	"bladeascend/server/internal/geom"
)

// Trạng thái phụ trợ cho kỹ năng (spec §6.3): vệt trơn của Cô Ba, đàn em của Đại Ca, và
// sổ ghi kiếm vừa bể của A Hớn.

// slick là một vũng trái cây trên đất: ai đi vào thì trượt theo quán tính.
type slick struct {
	at    Vec
	until float64
	owner uint32
	// r: bán kính vũng, = `ranges_u.slick_width × U` của người ĐÃ TẠO ra nó, chốt ngay lúc
	// lướt (spec v6.5 §6.3). Lưu vào vũng chứ không tính lại mỗi tick: vũng sống 4 giây và
	// trong 4 giây đó Cô Ba có thể lên cảnh giới, rớt cảnh giới, hoặc rời phòng — tính lại
	// theo người thì vũng tự phình ra hoặc biến mất mà không có lý do nào nhìn thấy được.
	r float64
}

// spawnMinion thả một đàn em của Đại Ca Dép Tổ Ong (spec v5.4 §6.3).
//
// Đàn em là **người chơi do máy điều khiển**, cùng cảnh giới với đại ca và mang **50% số
// kiếm hiện tại của đại ca** (tối thiểu 1), xếp vòng y như người chơi. Làm thành Player
// thay vì một loại thực thể mới để mọi luật va kiếm sẵn có chạy nguyên vẹn, khỏi phải viết
// lại một bản va chạm thứ hai — và cũng để vòng kiếm của chúng dùng chung đúng đường vẽ của
// người chơi ở client, đúng ghi chú hiệu năng của spec.
//
// Đại ca KHÔNG bị trừ kiếm: đây là bản sao, không phải chia của.
func (g *Game) spawnMinion(boss *Player, at Vec, secs float64) *Player {
	sc := g.Cfg.Skills
	sw := int(float64(boss.Ring.Count) * sc.DanEmSwordShare)
	if sw < 1 {
		sw = 1
	}
	m := g.AddPlayer(SpawnOptions{
		Name: boss.Name + " (đàn em)", Char: boss.Char, IsBot: true,
		Level: boss.Level, Swords: sw,
	})
	if m == nil {
		return nil
	}
	m.Pos = g.resolveBlocking(at, g.BodyR(m))
	m.IsMinion = true
	m.MinionOf = boss.ID
	m.MinionUntil = g.Time + secs
	m.InvulnUntil = 0
	// Nhỏ hơn đại ca 30% (spec). Đặt SAU khi vào bàn cờ vì `AddPlayer` tự tính scale theo
	// cảnh giới, mà đàn em thì cùng cảnh giới với đại ca nên sẽ to bằng nếu không chỉnh.
	m.Scale *= sc.DanEmScale
	return m
}

// updateSkillState chạy mỗi tick: hết hạn vệt trơn, hết hạn đàn em, bắt người đi vào vệt
// trơn, và bật sóng đẩy khi Cụ Rùa ra khỏi mai.
func (g *Game) updateSkillState(dt float64) {
	// Vệt trơn: dọn vũng hết hạn rồi bắt người đi vào.
	keep := g.slicks[:0]
	for _, s := range g.slicks {
		if g.Time >= s.until {
			continue
		}
		keep = append(keep, s)
		for _, p := range g.PlayerList() {
			if p.Dead || p.ID == s.owner || p.InTransit(g.Time) {
				continue
			}
			if p.Pos.Dist(s.at) <= s.r {
				// Trượt VÀ choáng cùng lúc (spec v5.5): trượt theo quán tính, đồng thời
				// mất quyền điều khiển trong đúng khoảng đó.
				g.ApplyStatus(p, StSlip, g.Cfg.Skills.SlipSecs)
				g.ApplyStatus(p, StStun, g.Cfg.Skills.SlipSecs)
			}
		}
	}
	g.slicks = keep
	g.updateStalls()

	for _, p := range g.PlayerList() {
		g.updateChannel(p)
		g.updateUnstick(p, dt)
		// Cụ Rùa vừa ra khỏi mai → sóng đẩy. Dùng cờ riêng chứ không so thời gian, vì tick
		// có thể nhảy qua đúng khoảnh khắc hết hạn.
		if p.shellOn && !p.HasStatus(StShell, g.Time) {
			p.shellOn = false
			g.shellBurst(p)
		}
		if p.HasStatus(StShell, g.Time) {
			p.shellOn = true
		}
		// Đàn em hết giờ thì biến mất.
		if p.IsMinion && g.Time >= p.MinionUntil {
			g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "dan_em_gone", X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
			g.RemovePlayer(p.ID)
		}
	}
}

// updateChannel lo chiêu đang "hát" dở: tới hạn thì phát, đi khỏi chỗ thì mất.
//
// Huỷ vì DỊCH CHUYỂN cũng tính, không chỉ vì bị chém: `MoveSpeed` đã đặt tốc 0 trong lúc
// hát, nhưng đẩy lùi (Loa Phường, đòn bể kiếm), trượt trên băng và vỏ chuối đều dịch người
// mà không qua `MoveSpeed`. Spec nói "đứng yên 0.5s" — ai đẩy được Ông Ba Võng ra khỏi chỗ
// thì cũng phá được câu hát.
func (g *Game) updateChannel(p *Player) {
	if p.channel.kind == "" {
		return
	}
	if p.Dead || p.InTransit(g.Time) || p.Pos.Dist(p.channel.at) > g.BodyR(p) {
		g.cancelChannel(p, "moved")
		return
	}
	if g.Time < p.channel.until {
		return
	}
	kind := p.channel.kind
	p.channel = channelState{}
	p.Status[StSing] = 0
	if kind == "ru_ngu" {
		g.finishRuNgu(p)
	}
}

// cancelChannel ngắt câu hát. **Lượt Bí Kíp không được hoàn** — đã trừ lúc `CastSkill`.
//
// Mất lượt là cả giá trị của việc có giai đoạn hát: nếu huỷ mà được trả lượt thì không có
// rủi ro nào, và "bị chém là mất chiêu" của §6.3 chỉ còn là một chút chậm trễ.
func (g *Game) cancelChannel(p *Player, why string) {
	if p.channel.kind == "" {
		return
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: p.channel.kind + "_huy", To: p.ID,
		X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
	_ = why
	p.channel = channelState{}
	p.Status[StSing] = 0
}

// updateUnstick lo khoảnh khắc Giao Hỏa Tốc hết hạn khi Shipper đang đứng TRONG vật cản
// (spec v5.5 §6.3): đẩy mượt ra chỗ trống gần nhất trong 0,3 giây.
//
// Vì sao không bật ra ngay: nhảy một phát là client thấy đúng kiểu giật mà §8.2a vừa mất
// cả buổi để sửa — client dự đoán một đằng, server kéo về một nẻo. Trượt trong 0,3 giây thì
// bộ nội suy của client bám theo được.
//
// Chỗ trống gần nhất lấy bằng chính `resolveBlocking` — tức là **dùng lại đúng mã va chạm
// mà client cũng chạy** (spec đòi vậy), không viết một phép tìm chỗ trống thứ hai.
func (g *Game) updateUnstick(p *Player, dt float64) {
	// Đang trượt ra: nội suy tuyến tính từ chỗ kẹt tới chỗ trống.
	if p.unstickUntil > g.Time {
		total := math.Max(g.Cfg.Skills.GhostUnstickSecs, 1e-6)
		left := p.unstickUntil - g.Time
		t := geom.Clamp(1-left/total, 0, 1)
		p.Pos = p.unstickFrom.Add(p.unstickTo.Sub(p.unstickFrom).Mul(t))
		return
	}
	if p.ghostOn && !p.HasStatus(StGhost, g.Time) {
		p.ghostOn = false
		free := g.resolveBlocking(p.Pos, g.BodyR(p))
		if free.Dist2(p.Pos) > 1e-6 {
			p.unstickFrom, p.unstickTo = p.Pos, free
			p.unstickUntil = g.Time + g.Cfg.Skills.GhostUnstickSecs
			g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "ghost_unstick",
				X: float32(free.X), Y: float32(free.Y)})
		}
	}
	if p.HasStatus(StGhost, g.Time) {
		p.ghostOn = true
	}
}

// minionTarget: đàn em lao vào đối thủ gần đại ca nhất **trong tầm tìm 8 U** (spec v6.5 §6.3).
//
// Giới hạn tầm tìm là phần mới của v6.5. Bản trước lấy đối thủ gần nhất trong CẢ KHU, nên
// hai đàn em có thể bỏ đại ca mà chạy nửa bản đồ đuổi một người chẳng liên quan — trong khi
// kỹ năng chỉ sống 10 giây.
func (g *Game) minionTarget(m *Player) *Player {
	boss, ok := g.Players[m.MinionOf]
	if !ok {
		return nil
	}
	seek := g.skillRange(boss, "dan_em_seek")
	var best *Player
	bd := 0.0
	for _, o := range g.PlayerList() {
		if o.ID == boss.ID || o.ID == m.ID || o.Dead || o.IsMinion && o.MinionOf == boss.ID {
			continue
		}
		d := o.Pos.Dist(boss.Pos)
		if d > seek {
			continue
		}
		if best == nil || d < bd {
			best, bd = o, d
		}
	}
	return best
}

// moveMinions lái đàn em: bám theo mục tiêu của đại ca, không thì bám đại ca.
//
// **Vòng bám 3 U quanh đại ca là ràng buộc mạnh hơn mục tiêu** (spec v6.5 §6.3): ra ngoài
// vòng đó thì quay về, dù đang đuổi ai. Không có luật này thì đàn em bị kéo đi theo mục
// tiêu đang chạy và đại ca mất sạch lớp bảo vệ mà mình vừa gọi ra.
func (g *Game) moveMinions(dt float64) {
	for _, m := range g.PlayerList() {
		if !m.IsMinion || m.Dead {
			continue
		}
		boss, hasBoss := g.Players[m.MinionOf]
		goal := g.minionTarget(m)
		at := m.Pos
		switch {
		case hasBoss && m.Pos.Dist(boss.Pos) > g.skillRange(boss, "dan_em_follow"):
			at = boss.Pos
		case goal != nil:
			at = goal.Pos
		case hasBoss:
			at = boss.Pos
		}
		if d := at.Sub(m.Pos); d.Len() > 40 {
			m.InputAngle = d.Angle()
			m.Moving = true
		} else {
			m.Moving = false
		}
	}
}
