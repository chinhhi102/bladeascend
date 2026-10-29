package game

import (
	"math"

	"kiemxoay/server/internal/geom"
)

// Kỹ năng nhân vật, mở bằng vật phẩm Bí Kíp (spec §6.3).
//
// Mỗi nhân vật có đúng MỘT kỹ năng và không tự dùng được: phải nhặt Bí Kíp mới có lượt.
// Bí Kíp là vật phẩm chung, ai nhặt cũng được lượt kỹ năng của CHÍNH MÌNH — nên nó không
// bao giờ rơi "sai người", và người ta vẫn phải tranh nhau một món đồ chung.
//
// Ba luật cân bằng của spec, áp cho mọi kỹ năng:
//   - không trực tiếp đổi cảnh giới hay EXP;
//   - không chém thân thay kiếm;
//   - không xuyên luật "còn ≥ 2 vòng thì kiếm địch không chạm tới thân" (§5.2).
//
// Bán kính và tầm xa **nhân theo scale** người dùng; thời gian hiệu ứng thì không.

// Skill mô tả một kỹ năng.
type Skill struct {
	// Char: nhân vật sở hữu; Name: tên hiển thị; Line: câu thoại tấu hài khi dùng.
	Char string
	Name string
	Line string
	// cast thi hành kỹ năng. `dir` là hướng nhìn của người dùng.
	cast func(g *Game, p *Player, dir Vec)
}

// Skills theo đúng bảng §6.3, khớp tên nhân vật trong config.characters.
var Skills = map[string]*Skill{
	"a_hon":       {Char: "a_hon", Name: "Bánh Bao Thần", Line: "Ăn miếng bánh bao, kiếm về như mới!", cast: castBanhBao},
	"ba_tam_loa":  {Char: "ba_tam_loa", Name: "Loa Phường", Line: "A lô a lô, tránh ra cho bà đi!", cast: castLoaPhuong},
	"ong_ba_vong": {Char: "ong_ba_vong", Name: "Ru Ngủ", Line: "À ơi… ngủ đi con…", cast: castRuNgu},
	"dai_ca_dep":  {Char: "dai_ca_dep", Name: "Gọi Đàn Em", Line: "Tụi bay đâu, lên đây!", cast: castGoiDanEm},
	"map_tra_sua": {Char: "map_tra_sua", Name: "Phun Trân Châu", Line: "Nếm thử trân châu nhà tao!", cast: castTranChau},
	"co_ba_ganh":  {Char: "co_ba_ganh", Name: "Gánh Lướt", Line: "Tránh đường, hàng nóng!", cast: castGanhLuot},
	"su_phu_mi":   {Char: "su_phu_mi", Name: "Lưới Mì", Line: "Mì gói ba phút, trói mày ba giây!", cast: castLuoiMi},
	"be_ti":       {Char: "be_ti", Name: "Gồng Cơ", Line: "Gồng! Không đứa nào đẩy nổi!", cast: castGongCo},
	"shipper":     {Char: "shipper", Name: "Giao Hỏa Tốc", Line: "Đơn gấp, xin phép đi xuyên!", cast: castGiaoHoaToc},
	"cu_rua":      {Char: "cu_rua", Name: "Rụt Cổ", Line: "Thiền… các ngươi cứ chém đi.", cast: castRutCo},
}

// SkillOf trả về kỹ năng của nhân vật này, nil nếu nhân vật chưa có kỹ năng.
func SkillOf(char string) *Skill { return Skills[char] }

// AddSkillCharge cộng một lượt dùng kỹ năng khi nhặt Bí Kíp (spec §6.3).
//
// Đã đủ 2 lượt mà nhặt thêm thì đổi sang tu vi, không bỏ phí: +20 Linh khí, hoặc +2 Tiên khí
// với người đã lên Tiên giới. Trả về false khi đã đầy (người gọi lo phần quy đổi).
func (g *Game) AddSkillCharge(p *Player) bool {
	if p.SkillCharges >= g.Cfg.Skills.MaxCharges {
		return false
	}
	p.SkillCharges++
	return true
}

// CanCastSkill: đủ điều kiện bấm kỹ năng chưa (spec §6.3).
func (g *Game) CanCastSkill(p *Player) bool {
	return !p.Dead && p.SkillCharges > 0 && g.Time >= p.SkillReadyAt &&
		!p.InTransit(g.Time) && SkillOf(p.Char) != nil
}

// CastSkill dùng kỹ năng của nhân vật. Trả về false nếu chưa dùng được.
func (g *Game) CastSkill(p *Player) bool {
	if !g.CanCastSkill(p) {
		return false
	}
	sk := SkillOf(p.Char)
	p.SkillCharges--
	p.SkillReadyAt = g.Time + g.Cfg.Skills.LockSecs
	// Đang Ẩn Thân mà dùng kỹ năng thì lộ ngay (spec §6.3).
	p.Buffs[BuffStealth] = 0
	dir := geom.FromAngle(p.Facing)
	sk.cast(g, p, dir)
	g.emit(Event{Kind: "SkillCast", A: p.ID, S: sk.Char, N: p.SkillCharges,
		X: float32(p.Pos.X), Y: float32(p.Pos.Y), F: float32(p.Facing)})
	return true
}

// ---- từng kỹ năng ---------------------------------------------------------

// inCone: q có nằm trong hình nón đỉnh tại p.Pos, hướng dir, góc mở `deg`, dài `reach`?
//
// Tầm xa nhân theo scale người dùng (spec §6.3) — người to thì tay dài, đó cũng là cách duy
// nhất để kỹ năng của Đạo Tổ không thành vô dụng khi vòng kiếm của họ đã rộng 1 900.
func inCone(p *Player, q Vec, dir Vec, deg, reach float64) bool {
	to := q.Sub(p.Pos)
	d := to.Len()
	if d > reach*p.Scale || d < 1e-6 {
		return false
	}
	cos := geom.Clamp(to.Mul(1/d).Dot(dir), -1, 1)
	return math.Acos(cos)*180/math.Pi <= deg/2
}

// targetsNear trả về đối thủ còn sống trong bán kính r (đã nhân scale).
func (g *Game) targetsNear(p *Player, r float64) []*Player {
	var out []*Player
	for _, o := range g.PlayerList() {
		if o.ID == p.ID || o.Dead || o.InTransit(g.Time) {
			continue
		}
		if o.Pos.Dist(p.Pos) <= r*p.Scale {
			out = append(out, o)
		}
	}
	return out
}

// Bánh Bao Thần — hồi 30% số kiếm đã bể trong 10 giây gần nhất.
func castBanhBao(g *Game, p *Player, _ Vec) {
	sc := g.Cfg.Skills
	lost := 0
	for _, e := range p.lostSwords {
		if g.Time-e.at <= sc.BanhBaoWindowSecs {
			lost += e.n
		}
	}
	back := int(float64(lost) * sc.BanhBaoShare)
	if back > 0 {
		g.AddSwords(p, back)
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "banh_bao", N: back, X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// Loa Phường — nón 60° dài 500: đẩy lùi 300 và làm vòng kiếm địch đổi chiều.
func castLoaPhuong(g *Game, p *Player, dir Vec) {
	sc := g.Cfg.Skills
	for _, o := range g.targetsNear(p, sc.LoaReach) {
		if !inCone(p, o.Pos, dir, sc.LoaConeDeg, sc.LoaReach) {
			continue
		}
		if o.HasStatus(StBrace, g.Time) || o.HasStatus(StShell, g.Time) {
			continue // đang gồng hoặc rụt mai thì không đẩy được
		}
		away := o.Pos.Sub(p.Pos).Norm()
		o.KnockVel = away.Mul(sc.LoaPush / math.Max(sc.KnockSecs, 1e-6))
		o.KnockUntil = g.Time + sc.KnockSecs
		o.Ring.Dir = -o.Ring.Dir
		g.emit(Event{Kind: "SkillEffect", A: p.ID, B: o.ID, S: "loa", X: float32(o.Pos.X), Y: float32(o.Pos.Y)})
	}
}

// Ru Ngủ — địch trong bán kính 350 chậm 40% và vòng kiếm chậm 50% trong 3 giây.
func castRuNgu(g *Game, p *Player, _ Vec) {
	sc := g.Cfg.Skills
	for _, o := range g.targetsNear(p, sc.RuNguRadius) {
		g.ApplyStatus(o, StSlow, sc.RuNguSecs)
		g.emit(Event{Kind: "SkillEffect", A: p.ID, B: o.ID, S: "ru_ngu", X: float32(o.Pos.X), Y: float32(o.Pos.Y)})
	}
}

// Gọi Đàn Em — 2 NPC nhỏ cùng cảnh giới chạy theo 8 giây, lao vào mài kiếm địch gần nhất.
func castGoiDanEm(g *Game, p *Player, _ Vec) {
	sc := g.Cfg.Skills
	for i := 0; i < sc.DanEmCount; i++ {
		a := p.Facing + float64(i)*math.Pi*2/float64(sc.DanEmCount)
		at := p.Pos.Add(geom.FromAngle(a).Mul(g.Geom(p).Reach + 80))
		m := g.spawnMinion(p, at, sc.DanEmSecs)
		if m == nil {
			break
		}
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "dan_em", N: sc.DanEmCount, X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// Phun Trân Châu — nón 45° dài 700, người trúng bị bịt mắt 3,5 giây.
func castTranChau(g *Game, p *Player, dir Vec) {
	sc := g.Cfg.Skills
	for _, o := range g.targetsNear(p, sc.TranChauReach) {
		if inCone(p, o.Pos, dir, sc.TranChauConeDeg, sc.TranChauReach) {
			g.ApplyStatus(o, StBlind, sc.TranChauSecs)
		}
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "tran_chau", X: float32(p.Pos.X), Y: float32(p.Pos.Y), F: float32(p.Facing)})
}

// Gánh Lướt — lướt 600 theo hướng đi trong 0,3 giây, rơi trái cây tạo vệt trơn 4 giây.
func castGanhLuot(g *Game, p *Player, dir Vec) {
	sc := g.Cfg.Skills
	p.KnockVel = dir.Mul(sc.DashDist * p.Scale / math.Max(sc.DashSecs, 1e-6))
	p.KnockUntil = g.Time + sc.DashSecs
	g.ApplyStatus(p, StDash, sc.DashSecs)
	// Vệt trơn: một dãy vũng nhỏ dọc đường lướt.
	for i := 1; i <= sc.SlickCount; i++ {
		at := p.Pos.Add(dir.Mul(sc.DashDist * p.Scale * float64(i) / float64(sc.SlickCount)))
		g.slicks = append(g.slicks, slick{at: at, until: g.Time + sc.SlickSecs, owner: p.ID})
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "ganh_luot", X: float32(p.Pos.X), Y: float32(p.Pos.Y), F: float32(p.Facing)})
}

// Lưới Mì — ném lưới tới điểm cách 400 phía trước, bán kính 250, trói 1,5 giây.
func castLuoiMi(g *Game, p *Player, dir Vec) {
	sc := g.Cfg.Skills
	at := p.Pos.Add(dir.Mul(sc.LuoiMiRange * p.Scale))
	for _, o := range g.PlayerList() {
		if o.ID == p.ID || o.Dead || o.InTransit(g.Time) {
			continue
		}
		if o.Pos.Dist(at) <= sc.LuoiMiRadius*p.Scale {
			g.ApplyStatus(o, StRoot, sc.LuoiMiSecs)
		}
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "luoi_mi", X: float32(at.X), Y: float32(at.Y)})
}

// Gồng Cơ — 4 giây độ bền kiếm ×2 và miễn mọi khống chế.
func castGongCo(g *Game, p *Player, _ Vec) {
	sc := g.Cfg.Skills
	g.ApplyStatus(p, StBrace, sc.BraceSecs)
	// Gồng là để thoát khống chế: xoá luôn những thứ đang dính.
	for _, s := range controlStatuses {
		p.Status[s] = 0
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "gong_co", X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// Giao Hỏa Tốc — tốc chạy ×2 trong 3 giây và đi xuyên vật cản PHÁ ĐƯỢC.
func castGiaoHoaToc(g *Game, p *Player, _ Vec) {
	sc := g.Cfg.Skills
	g.ApplyStatus(p, StGhost, sc.GhostSecs)
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "giao_hoa_toc", X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// Rụt Cổ — 2,5 giây đứng yên, không bị bể kiếm và không bị chém thân; ra khỏi mai bật sóng đẩy.
func castRutCo(g *Game, p *Player, _ Vec) {
	sc := g.Cfg.Skills
	g.ApplyStatus(p, StShell, sc.ShellSecs)
	// Bật cờ ngay tại đây, không đợi updateSkillState: nếu mai hết hạn trước tick kế tiếp
	// (hoặc test nhảy thẳng thời gian) thì cờ chưa kịp bật và sóng đẩy mất tăm.
	p.shellOn = true
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "rut_co", X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// shellBurst là sóng đẩy khi Cụ Rùa ra khỏi mai, gọi lúc StShell vừa hết hạn.
func (g *Game) shellBurst(p *Player) {
	sc := g.Cfg.Skills
	for _, o := range g.targetsNear(p, sc.ShellPushRadius) {
		if o.HasStatus(StBrace, g.Time) {
			continue
		}
		away := o.Pos.Sub(p.Pos).Norm()
		o.KnockVel = away.Mul(sc.ShellPush / math.Max(sc.KnockSecs, 1e-6))
		o.KnockUntil = g.Time + sc.KnockSecs
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "rut_co_burst", X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}
