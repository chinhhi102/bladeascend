package game

import (
	"math"

	"bladeascend/server/internal/geom"
)

// Kỹ năng nhân vật — tự hồi chiêu + tích lượt (spec v6.25 §6.3).
//
// Mỗi nhân vật có đúng MỘT kỹ năng, `cooldowns[char] = [giây hồi 1 lượt, lượt tối đa]`.
// Vật phẩm Bí Kíp đã bỏ (v6.25).
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

// skillCooldown: giây hồi 1 lượt và số lượt tích tối đa (thường) của nhân vật (spec v6.25
// §6.3). Nhân vật thiếu trong bảng thì không có kỹ năng (0, 0).
func (g *Game) skillCooldown(p *Player) (secs float64, maxC int) {
	cd, ok := g.Cfg.Skills.Cooldowns[p.Char]
	if !ok {
		return 0, 0
	}
	return cd[0], int(cd[1])
}

// SkillCooldownSecs: giây hồi 1 lượt của người này (client vẽ vòng hồi chiêu).
func (g *Game) SkillCooldownSecs(p *Player) float64 {
	secs, _ := g.skillCooldown(p)
	return secs
}

// MaxSkillCharges: trần lượt tích — theo nhân vật, Tiên giới +1 (spec v6.25 §6.3).
func (g *Game) MaxSkillCharges(p *Player) int {
	_, maxC := g.skillCooldown(p)
	if maxC > 0 && g.TienRealmOf(p) >= 0 {
		maxC += g.Cfg.Skills.TienExtraCharges
	}
	return maxC
}

// startSkillCharges: vào trận / hồi sinh có sẵn `start_charges` lượt, đồng hồ hồi chạy ngay.
func (g *Game) startSkillCharges(p *Player) {
	p.SkillCharges = min(g.Cfg.Skills.StartCharges, g.MaxSkillCharges(p))
	p.SkillRecharge = 0
	g.armRecharge(p)
}

// armRecharge bật đồng hồ hồi nếu còn chỗ trống mà đồng hồ đang dừng.
func (g *Game) armRecharge(p *Player) {
	if p.SkillRecharge <= 0 && p.SkillCharges < g.MaxSkillCharges(p) {
		p.SkillRecharge = g.SkillCooldownSecs(p)
	}
}

// rechargeSkills chạy mỗi tick: hồi lần lượt từng lượt, đầy thì dừng. Người chết hay đang
// dịch chuyển giữa hai khu không có mặt trong `Players` nên tự đứng đồng hồ.
func (g *Game) rechargeSkills(dt float64) {
	for _, p := range g.PlayerList() {
		if p.Dead || p.SkillRecharge <= 0 {
			continue
		}
		p.SkillRecharge -= dt
		if p.SkillRecharge > 0 {
			continue
		}
		p.SkillCharges++
		p.SkillRecharge = 0
		g.armRecharge(p)
		g.emit(Event{Kind: "SkillCharge", A: p.ID, To: p.ID, N: p.SkillCharges,
			X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
	}
}

// FillSkillCharges: hồi đầy lượt ngay — CHỈ cho lệnh debug.
func (g *Game) FillSkillCharges(p *Player) {
	p.SkillCharges = g.MaxSkillCharges(p)
	p.SkillRecharge = 0
}

// clampSkillCharges cắt lượt thừa khi trần tụt xuống — rớt từ Tiên giới về Đại Thừa.
//
// Gọi ở MỌI chỗ cảnh giới có thể tụt (qua `refreshDerived`): Thiêu Đốt cũng kéo Tiên khí
// xuống. Trần tăng (lên Tiên giới) thì bật lại đồng hồ nếu đang đầy.
func (g *Game) clampSkillCharges(p *Player) {
	maxC := g.MaxSkillCharges(p)
	if p.SkillCharges > maxC {
		p.SkillCharges = maxC
		g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "charge_lost", N: maxC,
			X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
	}
	if p.SkillCharges >= maxC {
		p.SkillRecharge = 0
		return
	}
	if !p.Dead {
		g.armRecharge(p)
	}
}

// CanCastSkill: đủ điều kiện bấm kỹ năng chưa (spec §6.3).
//
// Choáng và NGỦ khoá kỹ năng ("không di chuyển/kỹ năng/Thiêu Đốt" — Ru Ngủ, §6.3). Ngoại
// lệ: Shipper — Gom Đơn "miễn mọi khống chế", nên đó chính là nút thoát khi bị choáng/ngủ
// (§17; bot Shipper cũng dùng đúng lúc bị khống chế).
func (g *Game) CanCastSkill(p *Player) bool {
	if p.Dead || p.SkillCharges <= 0 || g.Time < p.SkillReadyAt || p.InTransit(g.Time) || SkillOf(p.Char) == nil {
		return false
	}
	return p.Char == "shipper" || !g.actionLocked(p)
}

// actionLocked: đang choáng hoặc ngủ gục — không kỹ năng, không Thiêu Đốt (spec §6.3).
func (g *Game) actionLocked(p *Player) bool {
	return p.HasStatus(StStun, g.Time) || p.HasStatus(StSleep, g.Time)
}

// CastSkill dùng kỹ năng của nhân vật. Trả về false nếu chưa dùng được.
func (g *Game) CastSkill(p *Player) bool {
	if !g.CanCastSkill(p) {
		return false
	}
	sk := SkillOf(p.Char)
	p.SkillCharges--
	p.SkillCasts++
	g.armRecharge(p) // đang đầy thì đồng hồ dừng — dùng 1 lượt là nó chạy lại
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

// Bánh Bao Thần (v6.48 §6.3) — hồi NGAY 10% số kiếm tối đa, rồi hồi dần 20% nữa trong 4s.
//
// Bản v5.5 hồi trọn 20% tức thì. Đổi vì nó làm A Hớn thành nút "huỷ đòn": sắp bị mài chết
// thì bấm một cái là đầy lại nửa vòng, đối thủ mất toi mười giây vừa mài. Giờ phần lớn số
// kiếm chỉ về nếu **sống thêm 4 giây** — chết là mất phần chưa hồi (xem `onDeathClear`).
//
// `AddSwords` lấp vòng trong trước (xem `ring.go`), đúng ý spec — đó cũng là chỗ che chắn
// tốt nhất cho người vừa bị mài gần hết.
func castBanhBao(g *Game, p *Player, _ Vec) {
	sc := g.Cfg.Skills
	maxSw := g.Prog.MaxSwords(g.Prog.Realm(p.Level))
	now := g.addUpTo(p, int(float64(maxSw)*sc.BanhBaoNow), maxSw)
	// Phần hồi dần: đặt mục tiêu theo kiếm TỐI ĐA, nhịp = tổng / số giây.
	secs := math.Max(sc.BanhBaoRegenSecs, 1e-6)
	p.regenLeft = int(float64(maxSw) * sc.BanhBaoRegen)
	p.regenRate = float64(p.regenLeft) / secs
	p.regenAcc = 0
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "banh_bao", N: now, F: float32(secs),
		X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// addUpTo thêm `n` kiếm nhưng không vượt trần cảnh giới; trả số thật sự thêm được.
func (g *Game) addUpTo(p *Player, n, maxSw int) int {
	if room := maxSw - p.Ring.Count; n > room {
		n = room
	}
	if n <= 0 {
		return 0
	}
	g.AddSwords(p, n)
	return n
}

// regenSwords chạy mỗi tick: nhả dần phần kiếm còn nợ của Bánh Bao Thần.
//
// Cộng dồn theo phần lẻ rồi mới nhả từng thanh: nhịp 5%/giây của Luyện Khí (20 kiếm tối đa)
// là 1 thanh/giây, nhưng của Đạo Tổ (100 kiếm) là 5 thanh/giây — làm tròn mỗi tick thì cấp
// thấp không bao giờ hồi được thanh nào.
func (g *Game) regenSwords(dt float64) {
	for _, p := range g.PlayerList() {
		if p.regenLeft <= 0 || p.Dead {
			continue
		}
		p.regenAcc += p.regenRate * dt
		n := int(p.regenAcc)
		if n <= 0 {
			continue
		}
		p.regenAcc -= float64(n)
		if n > p.regenLeft {
			n = p.regenLeft
		}
		maxSw := g.Prog.MaxSwords(g.Prog.Realm(p.Level))
		got := g.addUpTo(p, n, maxSw)
		p.regenLeft -= n
		if got > 0 {
			g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "banh_bao_hoi", N: got,
				X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
		}
	}
}

// Loa Phường (v6.48 §6.3) — hai nhịp: **lấy hơi 0.8s** hút sạch đồ trong 6 U về mình, rồi
// **hét** đẩy mọi địch trong 6 U ra 6 U và đổi chiều vòng kiếm của họ.
//
// Bản v6.30 chỉ có cú đẩy. Đẩy là hiệu ứng *phòng thủ*: dọn chỗ quanh mình rồi… hết, địch
// chạy lại ngay trong hai giây. Nhịp hút đồ cột kỹ năng vào vòng lặp chính (tranh Linh khí
// ở Đài, ở xác vừa nổ đồ) đúng như §6.3 "Nguyên tắc thiết kế" mục 1 đòi, và vì phải đứng
// yên 0,8 giây nên nó cũng có cách chống lại (mục 3): chém chết bà trong lúc lấy hơi.
func castLoaPhuong(g *Game, p *Player, _ Vec) {
	secs := g.skillNum("loa_channel")
	p.channel = channelState{until: g.Time + secs, kind: "loa", at: p.Pos}
	// Đặt thẳng `Status` chứ không qua `ApplyStatus` — xem ghi chú ở `castRuNgu`.
	p.Status[StSing] = p.channel.until
	g.statusStart(nil, p, StSing, secs)
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "loa_lay_hoi", F: float32(secs),
		X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// finishLoa chạy khi lấy hơi xong: hút đồ trước, rồi hét.
//
// Hút TRƯỚC khi đẩy, không ngược lại: đẩy trước thì Bó Kiếm rơi ra từ người bị đẩy đã nằm
// ngoài tầm hút, và cái hay nhất của kỹ năng (hét cho rơi đồ rồi hốt) không bao giờ xảy ra.
func (g *Game) finishLoa(p *Player) {
	// Chốt phạm vi TRƯỚC khi hút: hút viên Linh khí có thể làm bà lên cảnh giới ngay giữa
	// chiêu, và một cú hét tự dài ra vì vừa nhặt được mấy viên là thứ không ai đoán nổi.
	r := g.skillRange(p, "loa_radius")
	push := g.skillRange(p, "loa_push")
	got := g.vacuumInto(p, r)
	for _, o := range g.targetsNear(p, r) {
		if o.HasStatus(StBrace, g.Time) || o.HasStatus(StShell, g.Time) {
			continue // đang gồng hoặc rụt mai thì không đẩy được
		}
		g.pushAway(p, o, push)
		o.Ring.Dir = -o.Ring.Dir
		g.emit(Event{Kind: "SkillEffect", A: p.ID, B: o.ID, S: "loa", X: float32(o.Pos.X), Y: float32(o.Pos.Y)})
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "loa_het", N: got,
		X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// vacuumInto hút mọi thứ nhặt được trong bán kính `r` về `p` — theo ĐÚNG luật hấp thụ
// thường (§6.3: "nhặt luôn, theo luật hấp thụ thường; đang bị mù thì không hút được").
//
// Dùng lại `canAbsorb` + `pickup` của `resolvePickups` chứ không viết bản thứ hai: viên
// Linh khí và Tiên khí không lẫn sang nhau (§3.4), và đồ còn trong thời gian "của người
// khác" thì vẫn phải chờ — hai luật ấy mà chép lại là hai chỗ phải nhớ sửa.
func (g *Game) vacuumInto(p *Player, r float64) int {
	if !g.CanPickUp(p) {
		return 0
	}
	r2 := r * r
	var got []*Item
	g.ItemsNear(p.Pos, r, r, func(it *Item) {
		if it.Pos.Dist2(p.Pos) <= r2 && g.canAbsorb(p, it) && (it.Owner == 0 || it.Owner == p.ID || g.Time >= it.OwnerUntil) {
			got = append(got, it)
		}
	})
	for _, it := range got {
		g.pickup(p, it)
	}
	return len(got)
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
	g.statusStart(nil, p, StSing, secs)
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "ru_ngu_hat", F: float32(secs),
		X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// finishRuNgu chạy khi câu hát xong — bán kính tính lúc KẾT THÚC, không lúc bắt đầu.
//
// Khác biệt quan trọng: ai chạy ra khỏi 4 U trong nửa giây hát thì thoát được. Đó là phần
// "nhìn thấy trước mà tránh" của kỹ năng, và client vẽ vòng báo trước đúng bán kính này.
func (g *Game) finishRuNgu(p *Player) {
	// v6.30: hát xong thì nằm võng `ru_ngu_steal_secs` giây Trộm Kiếm (xem `clash`).
	steal := g.skillNum("ru_ngu_steal_secs")
	p.StealUntil = g.Time + steal
	g.statusStart(nil, p, StSteal, steal)
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "trom_kiem", F: float32(steal),
		X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
	secs := g.skillNum("ru_ngu_secs")
	for _, o := range g.targetsNear(p, g.skillRange(p, "ru_ngu_r")) {
		g.ApplyStatusFrom(p, o, StSleep, secs)
		g.emit(Event{Kind: "SkillEffect", A: p.ID, B: o.ID, S: "ru_ngu", F: float32(secs),
			X: float32(o.Pos.X), Y: float32(o.Pos.Y)})
	}
}

// Gọi Đàn Em (v6.48 §6.3) — 2 đàn em trong 8 giây, mỗi đứa mang **15% số kiếm hiện tại**
// của đại ca (tổng 30%, ngang mức lời của Trộm Kiếm) và **nhặt đồ hộ đại ca**.
//
// Gọi lần nữa khi còn đàn em thì **làm mới thời gian + kiếm**, không đẻ thêm: trần 2 đứa là
// luật của spec, mà nếu gọi lần hai sinh thêm hai đứa nữa thì Đại Ca 25 s × 2 lượt dựng
// được một đội bốn người.
func castGoiDanEm(g *Game, p *Player, _ Vec) {
	sc := g.Cfg.Skills
	have := g.minionsOf(p)
	for _, m := range have {
		m.MinionUntil = g.Time + sc.DanEmSecs
		g.statusStart(p, m, StMinion, sc.DanEmSecs)
		if want := g.minionSwords(p); m.Ring.Count < want {
			g.AddSwords(m, want-m.Ring.Count)
		}
	}
	for i := len(have); i < sc.DanEmCount; i++ {
		a := p.Facing + float64(i)*math.Pi*2/float64(sc.DanEmCount)
		at := p.Pos.Add(geom.FromAngle(a).Mul(g.Geom(p).Reach + 80))
		if m := g.spawnMinion(p, at, sc.DanEmSecs); m == nil {
			break
		}
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "dan_em", N: sc.DanEmCount, X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// minionsOf: đàn em còn sống của đại ca này.
func (g *Game) minionsOf(boss *Player) []*Player {
	var out []*Player
	for _, m := range g.PlayerList() {
		if m.IsMinion && m.MinionOf == boss.ID && !m.Dead {
			out = append(out, m)
		}
	}
	return out
}

// Phun Trân Châu — nón 45° dài 8 U, người trúng bị bịt mắt 5 giây (§6.3).
func castTranChau(g *Game, p *Player, dir Vec) {
	sc := g.Cfg.Skills
	reach := g.skillRange(p, "tran_chau_cone")
	for _, o := range g.targetsNear(p, reach) {
		if inCone(p, o.Pos, dir, sc.TranChauConeDeg, reach) {
			g.ApplyStatusFrom(p, o, StBlind, sc.TranChauSecs)
		}
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "tran_chau", X: float32(p.Pos.X), Y: float32(p.Pos.Y), F: float32(p.Facing)})
}

// Hút Mì Đoạt Kiếm (v6.48 §6.3) — húp 0.6s (đứng yên) rồi cướp **25% số kiếm hiện có của
// MỌI địch trong bán kính 5 U**, không giới hạn số cướp; phần vượt trần vòng của Sư Phụ rơi
// thành Bó Kiếm quanh ông.
//
// Ba thứ v6.48 đổi so với bản cũ, mỗi thứ một lý do:
//   - **Mọi địch thay vì một người gần nhất**: đây là kỹ năng mạnh nhất về cướp kiếm, nên
//     nó phải đáng với 35 giây hồi và chỉ 1 lượt. Một mục tiêu thì nó yếu hơn Trộm Kiếm của
//     Ông Ba (25s × 2) ở mọi tình huống đông người.
//   - **Bỏ trần 20% kiếm tối đa của mình**: trần ấy biến kỹ năng thành vô dụng đúng lúc
//     đáng dùng nhất (vây giữa ba người đầy kiếm). Phần dư rơi xuống đất nên vẫn có giá —
//     ai cũng nhặt được, kể cả nạn nhân.
//   - **Húp 0.6s**: "cách chống lại" mà §6.3 mục 3 đòi. Chém chết ông trong lúc húp là mất
//     cả chiêu lẫn lượt.
func castHutMi(g *Game, p *Player, _ Vec) {
	secs := g.skillNum("hut_mi_channel")
	p.channel = channelState{until: g.Time + secs, kind: "hut_mi", at: p.Pos}
	p.Status[StSing] = p.channel.until
	g.statusStart(nil, p, StSing, secs)
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "hut_mi_hup", F: float32(secs),
		X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// finishHutMi: húp xong thì kéo kiếm. Bán kính tính lúc KẾT THÚC — ai chạy ra khỏi 5 U
// trong 0,6 giây thì thoát, y như Ru Ngủ.
func (g *Game) finishHutMi(p *Player) {
	targets := g.targetsNear(p, g.skillRange(p, "hut_mi_r"))
	if len(targets) == 0 {
		g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "hut_mi_miss",
			X: float32(p.Pos.X), Y: float32(p.Pos.Y), F: float32(p.Facing)})
		return
	}
	for _, o := range targets {
		g.stealSwords(p, o)
	}
}

// stealSwords chuyển kiếm từ `from` sang `to` theo luật Hút Mì.
//
// Thứ tự ba phép cắt ở đây không đổi được:
//  1. cướp `hut_mi_steal` phần số kiếm HIỆN CÓ của địch (làm tròn xuống, tối thiểu 1 nếu
//     họ còn kiếm) — tỉ lệ theo địch nên càng đông kiếm càng mất nhiều;
//  2. (v6.48 bỏ trần "20% kiếm tối đa của mình" — xem `castHutMi`);
//  3. phần vượt trần vòng của mình thì RƠI XUỐNG ĐẤT thành Bó Kiếm, không bốc hơi — người
//     khác tranh được, nên hút đầy tay vẫn có cái giá.
func (g *Game) stealSwords(to, from *Player) {
	if from.Ring.Count <= 0 {
		return
	}
	want := int(float64(from.Ring.Count) * g.skillNum("hut_mi_steal"))
	if want < 1 {
		want = 1
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
	g.giveSwords(to, got)
	g.emit(Event{Kind: "SkillEffect", A: to.ID, B: from.ID, S: "hut_mi", N: got,
		X: float32(from.Pos.X), Y: float32(from.Pos.Y)})
}

// giveSwords nhập `n` kiếm cướp được vào vòng của `to`, tới trần cảnh giới; phần dư rơi
// xuống đất thành Bó Kiếm (không quy đổi EXP — §6.3 cấm kỹ năng đổi EXP trực tiếp).
func (g *Game) giveSwords(to *Player, n int) {
	ownMax := g.Prog.MaxSwords(g.Prog.Realm(to.Level))
	gain := min(n, max(ownMax-to.Ring.Count, 0))
	if gain > 0 {
		to.Ring.Count += gain
		g.refreshDerived(to)
	}
	if over := n - gain; over > 0 {
		g.dropBundles(to.Pos, over, to.Geom.Reach)
	}
}

// Gồng Cơ — Bung Cơ (v6.48 §6.3): gồng 4 giây (độ bền kiếm +1.0, miễn mọi khống chế), hết
// gồng thì BUNG — đẩy địch trong 2 U ra 3 U và mỗi người bị đẩy mất 10% số kiếm hiện có.
//
// Bản cũ chỉ có phần gồng: một kỹ năng phòng thủ thuần, không đổi kết quả cuộc mài kiếm,
// đúng cái §6.3 "Nguyên tắc thiết kế" mục 1 gọi là hiệu ứng phụ. Phần bung đặt ở CUỐI nên
// Bé Tí phải sống hết 4 giây mới ăn — và địch nhìn thấy trước mà tránh được.
func castGongCo(g *Game, p *Player, _ Vec) {
	sc := g.Cfg.Skills
	g.ApplyStatus(p, StBrace, sc.BraceSecs)
	// Gồng là để thoát khống chế: xoá luôn những thứ đang dính.
	for _, s := range controlStatuses {
		p.Status[s] = 0
	}
	// Bật cờ ngay tại đây chứ không đợi `updateSkillState` — cùng lý do với `shellOn`: gồng
	// có thể hết trước tick kế tiếp (hoặc test nhảy thẳng thời gian) và sóng bung mất tăm.
	p.braceOn = true
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "gong_co", F: float32(sc.BraceSecs),
		X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
}

// braceBurst là sóng "bung cơ" lúc hết gồng (sóng 2 U, đẩy 3 U, mất 10% kiếm — §6.3 v6.48).
//
// Kiếm mất RƠI QUANH NẠN NHÂN thành Bó Kiếm, không về Bé Tí: nếu về tay người bung thì đây
// thành bản thứ hai của Trộm Kiếm, mà §6.3 nguyên tắc 4 đòi mỗi nhân vật một vai riêng.
func (g *Game) braceBurst(p *Player) {
	push := g.skillRange(p, "bung_co_push")
	share := g.skillNum("bung_co_lose")
	for _, o := range g.targetsNear(p, g.skillRange(p, "bung_co_wave")) {
		if o.HasStatus(StBrace, g.Time) || o.HasStatus(StShell, g.Time) {
			continue
		}
		g.pushAway(p, o, push)
		lost := g.loseSwordsHit(o, int(float64(o.Ring.Count)*share), true)
		if lost > 0 {
			g.dropBundles(o.Pos, lost, o.Geom.Reach)
		}
		g.emit(Event{Kind: "SkillEffect", A: p.ID, B: o.ID, S: "bung_co", N: lost,
			X: float32(o.Pos.X), Y: float32(o.Pos.Y)})
	}
	g.emit(Event{Kind: "SkillEffect", A: p.ID, S: "bung_co_song", X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
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
