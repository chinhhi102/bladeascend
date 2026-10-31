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
	"ong_ba_vong": {Char: "ong_ba_vong", Name: "Ru Ngủ Trưa", Line: "À ơi… ngủ đi con…", cast: castRuNgu},
	"dai_ca_dep":  {Char: "dai_ca_dep", Name: "Gọi Đàn Em", Line: "Tụi bay đâu, lên đây!", cast: castGoiDanEm},
	"map_tra_sua": {Char: "map_tra_sua", Name: "Phun Trân Châu", Line: "Nếm thử trân châu nhà tao!", cast: castTranChau},
	"co_ba_ganh":  {Char: "co_ba_ganh", Name: "Bày Hàng Chặn Lối", Line: "Hàng về! Tránh ra cho cô bày!", cast: castBayHang},
	"su_phu_mi":   {Char: "su_phu_mi", Name: "Hút Mì Đoạt Kiếm", Line: "Sùn sụt… kiếm mày ngon đấy!", cast: castHutMi},
	"be_ti":       {Char: "be_ti", Name: "Gồng Cơ", Line: "Gồng! Không đứa nào đẩy nổi!", cast: castGongCo},
	"shipper":     {Char: "shipper", Name: "Gom Đơn Hỏa Tốc", Line: "Đơn gấp, xin phép đi xuyên!", cast: castGomDon},
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
// `reach` là khoảng cách THẬT, đã tính scale xong — đừng nhân thêm ở đây.
//
// Bản v6.5 nhân `reach * p.Scale` tại chỗ này, từ thời `reach` còn là số thuần trong config.
// Sau khi phạm vi chuyển sang `ranges_u × U` thì U đã chứa scale, nên hai lần nhân thành
// **scale bình phương**: nón của Bà Tám ở Đạo Tổ (scale 6) dài 6 × 3 240 = 19 440, tức dài
// hơn cả bản đồ 12 000 — mọi người trong khu đều bị đẩy lùi cùng lúc.
func inCone(p *Player, q Vec, dir Vec, deg, reach float64) bool {
	to := q.Sub(p.Pos)
	d := to.Len()
	if d > reach || d < 1e-6 {
		return false
	}
	cos := geom.Clamp(to.Mul(1/d).Dot(dir), -1, 1)
	return math.Acos(cos)*180/math.Pi <= deg/2
}

// targetsNear trả về đối thủ còn sống trong bán kính `r`.
//
// `r` là khoảng cách THẬT (`skillRange` đã nhân U vào) — xem `inCone` để biết vì sao chỗ này
// không được nhân scale lần nữa.
func (g *Game) targetsNear(p *Player, r float64) []*Player {
	var out []*Player
	for _, o := range g.PlayerList() {
		if o.ID == p.ID || o.Dead || o.InTransit(g.Time) {
			continue
		}
		if o.Pos.Dist(p.Pos) <= r {
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

// Ru Ngủ Trưa (spec v6.9 §6.3) — hát ru 0.5s rồi mọi địch trong 4 U ngủ gục 2.5s.
//
// Hát ru là một **chiêu phải chờ**: Ông Ba Võng đứng yên `ru_ngu_channel` giây, bị chém
// trúng thân trong lúc đó là mất chiêu (lượt đã trừ rồi, không hoàn). Đó là "cách chống
// lại" mà §6.3 mục 3 đòi mỗi kỹ năng phải có — nếu ru được tức thì thì người mở giao tranh
// luôn thắng và không ai chọn nhân vật khác.
//
// Bản cũ chỉ làm CHẬM 40% + chậm vòng kiếm 50%: địch vẫn mài kiếm bình thường nên không đổi
// kết quả trận, đúng cái §6.3 "Nguyên tắc thiết kế" mục 1 gọi là hiệu ứng phụ.
func castRuNgu(g *Game, p *Player, _ Vec) {
	secs := g.skillNum("ru_ngu_channel")
	p.channel = channelState{until: g.Time + secs, kind: "ru_ngu", at: p.Pos}
	// Đặt thẳng `Status` chứ không qua `ApplyStatus`: `ApplyStatus` lấy mốc XA HƠN giữa cũ
	// và mới (chống khoá cứng), nhưng hát ru phải khớp CHÍNH XÁC với `p.channel.until` —
	// lệch nhau là client giữ chân người chơi lâu hơn server, hoặc ngược lại.
	p.Status[StSing] = p.channel.until
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "ru_ngu_hat", F: float32(secs),
		X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// finishRuNgu chạy khi câu hát xong — bán kính tính lúc KẾT THÚC, không lúc bắt đầu.
//
// Khác biệt quan trọng: ai chạy ra khỏi 4 U trong nửa giây hát thì thoát được. Đó là phần
// "nhìn thấy trước mà tránh" của kỹ năng, và client vẽ vòng báo trước đúng bán kính này.
func (g *Game) finishRuNgu(p *Player) {
	secs := g.skillNum("ru_ngu_secs")
	for _, o := range g.targetsNear(p, g.skillRange(p, "ru_ngu_r")) {
		g.ApplyStatus(o, StSleep, secs)
		g.emit(Event{Kind: "SkillEffect", A: p.ID, B: o.ID, S: "ru_ngu", F: float32(secs),
			X: float32(o.Pos.X), Y: float32(o.Pos.Y)})
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

// Hút Mì Đoạt Kiếm (spec v6.9 §6.3) — cướp 25% kiếm của MỘT địch gần nhất trong nón 60°
// dài 5 U; trần cướp là 20% kiếm tối đa của Sư Phụ, phần dư rơi xuống đất thành Bó Kiếm.
//
// Bản cũ (Lưới Mì) trói 5 giây: địch đứng im nhưng vòng kiếm vẫn quay, nên hai người vẫn
// mài nhau y như không có kỹ năng. Bản này đụng thẳng vào **số kiếm** — thứ quyết định mọi
// cuộc mài — nên nó mới đáng một lượt Bí Kíp.
//
// Hụt thì vẫn mất lượt (§6.3 mục 3): không ai trong nón → chỉ có sợi mì bay ra rồi rụng.
func castHutMi(g *Game, p *Player, dir Vec) {
	reach := g.skillRange(p, "hut_mi_cone")
	var target *Player
	best := math.Inf(1)
	for _, o := range g.targetsNear(p, reach) {
		if !inCone(p, o.Pos, dir, g.Cfg.Skills.LoaConeDeg, reach) {
			continue
		}
		if d := o.Pos.Dist(p.Pos); d < best {
			target, best = o, d
		}
	}
	if target == nil {
		g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "hut_mi_miss",
			X: float32(p.Pos.X), Y: float32(p.Pos.Y), F: float32(p.Facing)})
		return
	}
	g.stealSwords(p, target)
}

// stealSwords chuyển kiếm từ `from` sang `to` theo luật Hút Mì.
//
// Thứ tự ba phép cắt ở đây không đổi được:
//  1. cướp `hut_mi_steal` phần số kiếm HIỆN CÓ của địch (làm tròn xuống, tối thiểu 1 nếu
//     họ còn kiếm) — tỉ lệ theo địch nên càng đông kiếm càng mất nhiều;
//  2. cắt theo trần `hut_mi_cap_of_own_max` × kiếm tối đa của MÌNH — không thì Sư Phụ cấp
//     thấp hút một Đạo Tổ là nhảy vọt lên 100 kiếm mà không tu luyện gì;
//  3. phần vượt trần vòng của mình thì RƠI XUỐNG ĐẤT thành Bó Kiếm, không bốc hơi — người
//     khác tranh được, nên hút đầy tay vẫn có cái giá.
func (g *Game) stealSwords(to, from *Player) {
	if from.Ring.Count <= 0 {
		return
	}
	ownMax := g.Prog.MaxSwords(g.Prog.Realm(to.Level))
	want := int(float64(from.Ring.Count) * g.skillNum("hut_mi_steal"))
	if want < 1 {
		want = 1
	}
	if lim := int(float64(ownMax) * g.skillNum("hut_mi_cap_of_own_max")); want > lim {
		want = lim
	}
	// `LoseSwords` không bao giờ hạ địch xuống dưới `sword.min_count`, nên số cướp THẬT
	// phải lấy từ nó, không phải từ ý muốn của mình: cướp 25 mà địch chỉ nhả 18 thì không
	// được tạo ra 7 thanh kiếm từ không khí.
	got := g.LoseSwords(from, want)
	if got <= 0 {
		return
	}
	// `LoseSwords` bỏ từ cuối mảng vòng, tức **vòng ngoài cùng** (spec) — vòng ngoài là
	// vòng đang chạm địch, nên mất nó là mất tầm chém ngay, thấy được liền.
	//
	// Phần vượt trần vòng của mình thì RƠI XUỐNG ĐẤT thành Bó Kiếm, không bốc hơi: người
	// khác tranh được, nên hút đầy tay vẫn có cái giá. Dùng `dropBundles` chứ không
	// `AddSwords` cho phần dư, vì `AddSwords` quy đổi phần thừa thành EXP — mà §6.3 cấm
	// kỹ năng đổi EXP trực tiếp.
	room := ownMax - to.Ring.Count
	gain := min(got, max(room, 0))
	if gain > 0 {
		to.Ring.Count += gain
		g.refreshDerived(to)
	}
	if over := got - gain; over > 0 {
		g.dropBundles(to.Pos, over, to.Geom.Reach)
	}
	g.emit(Event{Kind: "SkillEffect", A: to.ID, B: from.ID, S: "hut_mi", N: got,
		X: float32(from.Pos.X), Y: float32(from.Pos.Y)})
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

// Gom Đơn Hỏa Tốc (spec v6.9 §6.3) — 4s: tốc chạy ×3, xuyên mọi vật cản trừ biên khu,
// MIỄN mọi khống chế, và hút sạch mọi viên / vật phẩm / Bó Kiếm trong 3 U dọc đường chạy.
//
// Ba thứ v6.9 thêm so với bản cũ, mỗi thứ sửa một lỗ hổng thật:
//   - **Miễn và XOÁ khống chế**: bản cũ Shipper đang chạy bị trói là đứng chết tại chỗ với
//     cờ "đi xuyên vật cản" — tức đứng trong lòng một bức tường, vô hình, không làm gì
//     được. Giờ dùng được cả lúc đang trói/ngủ/choáng và thoát ngay (checklist §17).
//   - **Hút đồ 3 U**: cho Shipper một việc để LÀM khi chạy, chứ không chỉ chạy nhanh.
//   - **Xuyên trừ biên khu**: xem `resolveBorders` — không thì Shipper đi tắt sang khu
//     cảnh giới cao hơn mà không qua cổng.
func castGomDon(g *Game, p *Player, _ Vec) {
	g.ApplyStatus(p, StGhost, g.skillNum("gom_don_secs"))
	// Xoá khống chế TRƯỚC khi đặt cờ: thứ tự không quan trọng ở đây nhưng việc xoá thì có —
	// §17 đòi "đang bị trói/ngủ/choáng dùng được và thoát ngay".
	for _, st := range controlStatuses {
		p.Status[st] = 0
	}
	// Bật cờ ngay tại đây chứ không đợi updateSkillState — y hệt lý do của `shellOn`: nếu
	// kỹ năng hết hạn trước tick kế tiếp (hoặc test nhảy thẳng thời gian) thì cờ chưa kịp
	// bật, và việc đẩy ra khỏi vật cản không bao giờ chạy.
	p.ghostOn = true
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "gom_don", X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
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
	push := g.skillRange(p, "rut_co_push")
	for _, o := range g.targetsNear(p, g.skillRange(p, "rut_co_wave")) {
		if o.HasStatus(StBrace, g.Time) {
			continue
		}
		g.pushAway(p, o, push)
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "rut_co_burst", X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}
