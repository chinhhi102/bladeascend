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

// MaxSkillCharges là trần số lượt Bí Kíp của người này (spec v6.5 §6.3).
//
// Người phàm giữ 2; **Tiên giới (Chân Tiên trở lên) giữ 3**.
func (g *Game) MaxSkillCharges(p *Player) int {
	if g.TienRealmOf(p) >= 0 && g.Cfg.Skills.MaxChargesTien > 0 {
		return g.Cfg.Skills.MaxChargesTien
	}
	return g.Cfg.Skills.MaxCharges
}

// AddSkillCharge cộng một lượt dùng kỹ năng khi nhặt Bí Kíp (spec §6.3).
//
// Đã đủ trần mà nhặt thêm thì đổi sang tu vi, không bỏ phí: +20 Linh khí, hoặc +2 Tiên khí
// với người đã lên Tiên giới. Trả về false khi đã đầy (người gọi lo phần quy đổi).
func (g *Game) AddSkillCharge(p *Player) bool {
	if p.SkillCharges >= g.MaxSkillCharges(p) {
		return false
	}
	p.SkillCharges++
	return true
}

// clampSkillCharges cắt lượt thừa khi trần tụt xuống — tức khi rớt từ Tiên giới về Đại Thừa
// đang cầm 3 lượt (spec v6.5 §6.3: "còn 2, lượt thứ 3 mất, hiệu ứng cuộn sách tan").
//
// Gọi ở MỌI chỗ cảnh giới có thể tụt, không chỉ lúc bị chém: Thiêu Đốt cũng kéo Tiên khí
// xuống, và cắt sót một đường là người chơi giữ được 3 lượt ở Đại Thừa mà không ai biết tại sao.
func (g *Game) clampSkillCharges(p *Player) {
	if maxC := g.MaxSkillCharges(p); p.SkillCharges > maxC {
		p.SkillCharges = maxC
		g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "charge_lost", N: maxC,
			X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
	}
}

// CanCastSkill: đủ điều kiện bấm kỹ năng chưa (spec §6.3).
func (g *Game) CanCastSkill(p *Player) bool {
	return !p.Dead && p.SkillCharges > 0 && g.Time >= p.SkillReadyAt &&
		!p.InTransit(g.Time) && !p.HasStatus(StStun, g.Time) && SkillOf(p.Char) != nil
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

// Bánh Bao Thần — hồi ngay 20% số kiếm TỐI ĐA của cảnh giới hiện tại (spec v5.5 §6.3).
//
// Bản v5.0 hồi theo "số kiếm vừa bể trong 10 giây", và cái đó có một lỗ hổng không chữa
// được bằng cách chỉnh số: ai chưa bị đánh thì ăn bánh bao chẳng được gì, nên kỹ năng vô
// dụng đúng lúc muốn dùng nhất là lúc mở đầu. Lấy mốc theo kiếm tối đa thì luôn có tác
// dụng, và vẫn tự giới hạn vì không vượt được trần của cảnh giới.
//
// `AddSwords` lấp vòng trong trước (xem `ring.go`), đúng ý spec — đó cũng là chỗ che chắn
// tốt nhất cho người vừa bị mài gần hết.
func castBanhBao(g *Game, p *Player, _ Vec) {
	maxSw := g.Prog.MaxSwords(g.Prog.Realm(p.Level))
	back := int(float64(maxSw) * g.Cfg.Skills.BanhBaoMaxShare)
	if room := maxSw - p.Ring.Count; back > room {
		back = room // không vượt trần cảnh giới
	}
	if back > 0 {
		g.AddSwords(p, back)
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "banh_bao", N: back, X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// Loa Phường — nón 60° dài 6 U: đẩy lùi 6 U và làm vòng kiếm địch đổi chiều (§6.3).
func castLoaPhuong(g *Game, p *Player, dir Vec) {
	sc := g.Cfg.Skills
	reach := g.skillRange(p, "loa_cone")
	push := g.skillRange(p, "loa_push")
	for _, o := range g.targetsNear(p, reach) {
		if !inCone(p, o.Pos, dir, sc.LoaConeDeg, reach) {
			continue
		}
		if o.HasStatus(StBrace, g.Time) || o.HasStatus(StShell, g.Time) {
			continue // đang gồng hoặc rụt mai thì không đẩy được
		}
		g.pushAway(p, o, push)
		o.Ring.Dir = -o.Ring.Dir
		g.emit(Event{Kind: "SkillEffect", A: p.ID, B: o.ID, S: "loa", X: float32(o.Pos.X), Y: float32(o.Pos.Y)})
	}
}

// Ru Ngủ — địch trong bán kính 5 U chậm 40% và vòng kiếm chậm 50% trong 3 giây (§6.3).
func castRuNgu(g *Game, p *Player, _ Vec) {
	sc := g.Cfg.Skills
	for _, o := range g.targetsNear(p, g.skillRange(p, "ru_ngu_radius")) {
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

// Phun Trân Châu — nón 45° dài 8 U, người trúng bị bịt mắt 5 giây (§6.3).
func castTranChau(g *Game, p *Player, dir Vec) {
	sc := g.Cfg.Skills
	reach := g.skillRange(p, "tran_chau_cone")
	for _, o := range g.targetsNear(p, reach) {
		if inCone(p, o.Pos, dir, sc.TranChauConeDeg, reach) {
			g.ApplyStatus(o, StBlind, sc.TranChauSecs)
		}
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "tran_chau", X: float32(p.Pos.X), Y: float32(p.Pos.Y), F: float32(p.Facing)})
}

// Gánh Lướt — lướt 7 U theo hướng đi trong 0,3 giây, rơi trái cây tạo vệt trơn 4 giây (§6.3).
func castGanhLuot(g *Game, p *Player, dir Vec) {
	sc := g.Cfg.Skills
	dash := g.skillRange(p, "dash")
	slickR := g.skillRange(p, "slick_width")
	p.KnockVel = dir.Mul(dash / math.Max(sc.DashSecs, 1e-6))
	p.KnockUntil = g.Time + sc.DashSecs
	g.ApplyStatus(p, StDash, sc.DashSecs)
	// Vệt trơn: một dãy vũng nhỏ dọc đường lướt.
	for i := 1; i <= sc.SlickCount; i++ {
		at := p.Pos.Add(dir.Mul(dash * float64(i) / float64(sc.SlickCount)))
		g.slicks = append(g.slicks, slick{at: at, until: g.Time + sc.SlickSecs, owner: p.ID, r: slickR})
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "ganh_luot", X: float32(p.Pos.X), Y: float32(p.Pos.Y), F: float32(p.Facing)})
}

// Lưới Mì — ném lưới tới điểm cách 4,5 U phía trước, bán kính 3 U, trói 5 giây (§6.3).
func castLuoiMi(g *Game, p *Player, dir Vec) {
	sc := g.Cfg.Skills
	at := p.Pos.Add(dir.Mul(g.skillRange(p, "luoi_mi_range")))
	radius := g.skillRange(p, "luoi_mi_radius")
	for _, o := range g.PlayerList() {
		if o.ID == p.ID || o.Dead || o.InTransit(g.Time) {
			continue
		}
		if o.Pos.Dist(at) <= radius {
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
	// Bật cờ ngay tại đây chứ không đợi updateSkillState — y hệt lý do của `shellOn`: nếu
	// kỹ năng hết hạn trước tick kế tiếp (hoặc test nhảy thẳng thời gian) thì cờ chưa kịp
	// bật, và việc đẩy ra khỏi vật cản không bao giờ chạy.
	p.ghostOn = true
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

// shellBurst là sóng đẩy khi Cụ Rùa ra khỏi mai (sóng 2 U, đẩy 3 U — §6.3), gọi lúc
// StShell vừa hết hạn.
func (g *Game) shellBurst(p *Player) {
	push := g.skillRange(p, "shell_push")
	for _, o := range g.targetsNear(p, g.skillRange(p, "shell_wave")) {
		if o.HasStatus(StBrace, g.Time) {
			continue
		}
		g.pushAway(p, o, push)
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "rut_co_burst", X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}
