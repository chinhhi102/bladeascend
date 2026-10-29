package game


// Trạng thái phụ trợ cho kỹ năng (spec §6.3): vệt trơn của Cô Ba, đàn em của Đại Ca, và
// sổ ghi kiếm vừa bể của A Hớn.

// slick là một vũng trái cây trên đất: ai đi vào thì trượt theo quán tính.
type slick struct {
	at    Vec
	until float64
	owner uint32
}

// lostSword ghi một lần bị bể kiếm, để Bánh Bao Thần biết hồi lại bao nhiêu.
//
// Phải ghi theo TỪNG LẦN kèm mốc thời gian, không cộng dồn vào một con số: kỹ năng chỉ hồi
// phần bể trong 10 giây gần nhất, mà một người Đạo Tổ có thể bể cả trăm kiếm trong một trận
// kéo dài — cộng dồn thì ăn một cái bánh bao là về đầy kiếm.
type lostSword struct {
	n  int
	at float64
}

// noteSwordLoss ghi lại số kiếm vừa bể và dọn những lần đã quá cũ.
func (g *Game) noteSwordLoss(p *Player, n int) {
	if n <= 0 {
		return
	}
	w := g.Cfg.Skills.BanhBaoWindowSecs
	keep := p.lostSwords[:0]
	for _, e := range p.lostSwords {
		if g.Time-e.at <= w {
			keep = append(keep, e)
		}
	}
	p.lostSwords = append(keep, lostSword{n: n, at: g.Time})
}

// spawnMinion thả một đàn em của Đại Ca Dép Tổ Ong (spec §6.3).
//
// Đàn em là **người chơi do máy điều khiển**, cùng cảnh giới với đại ca nhưng chỉ 1 vòng
// 10 kiếm, và **không chém thân được** — nên chúng chỉ mài kiếm đối thủ chứ không giết ai.
// Làm thành Player thay vì một loại thực thể mới để mọi luật va kiếm sẵn có chạy nguyên
// vẹn, khỏi phải viết lại một bản va chạm thứ hai.
func (g *Game) spawnMinion(boss *Player, at Vec, secs float64) *Player {
	m := g.AddPlayer(SpawnOptions{
		Name: boss.Name + " (đàn em)", Char: boss.Char, IsBot: true,
		Level: boss.Level, Swords: g.Cfg.Skills.DanEmSwords,
	})
	if m == nil {
		return nil
	}
	m.Pos = g.resolveBlocking(at, g.BodyR(m))
	m.IsMinion = true
	m.MinionOf = boss.ID
	m.MinionUntil = g.Time + secs
	m.InvulnUntil = 0
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
			if p.Pos.Dist(s.at) <= g.Cfg.Skills.SlickRadius {
				g.ApplyStatus(p, StSlip, g.Cfg.Skills.SlipSecs)
			}
		}
	}
	g.slicks = keep

	for _, p := range g.PlayerList() {
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

// minionTarget: đàn em lao vào đối thủ gần đại ca nhất (spec §6.3).
func (g *Game) minionTarget(m *Player) *Player {
	boss, ok := g.Players[m.MinionOf]
	if !ok {
		return nil
	}
	var best *Player
	bd := 0.0
	for _, o := range g.PlayerList() {
		if o.ID == boss.ID || o.ID == m.ID || o.Dead || o.IsMinion && o.MinionOf == boss.ID {
			continue
		}
		if d := o.Pos.Dist(boss.Pos); best == nil || d < bd {
			best, bd = o, d
		}
	}
	return best
}

// moveMinions lái đàn em: bám theo mục tiêu của đại ca, không thì bám đại ca.
func (g *Game) moveMinions(dt float64) {
	for _, m := range g.PlayerList() {
		if !m.IsMinion || m.Dead {
			continue
		}
		goal := g.minionTarget(m)
		at := m.Pos
		if goal != nil {
			at = goal.Pos
		} else if boss, ok := g.Players[m.MinionOf]; ok {
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
