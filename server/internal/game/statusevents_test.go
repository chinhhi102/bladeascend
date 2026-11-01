package game

import "testing"

// Spec v6.36 §6.5: server gửi `StatusStart/StatusEnd {id, status, secs, src}` — client
// không tự đoán trạng thái.

func statusEvents(g *Game, kind string) []Event {
	var out []Event
	for _, e := range g.DrainEvents() {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

// Phun Trân Châu: nạn nhân nhận StatusStart có đúng mã mù, số giây, người gây và chỗ người gây.
func TestStatusStartCarriesSourceAndSecs(t *testing.T) {
	g := skillGame(t)
	fat := put(g, "map_tra_sua", Vec{X: 3000, Y: 3000}, 10, 20)
	fat.Facing = 0
	vic := put(g, "a_hon", Vec{X: 3150, Y: 3000}, 10, 20)
	g.DrainEvents()
	giveCharge(fat)
	if !g.CastSkill(fat) {
		t.Fatal("không dùng được kỹ năng")
	}
	evs := statusEvents(g, "StatusStart")
	if len(evs) != 1 {
		t.Fatalf("muốn 1 StatusStart, có %d", len(evs))
	}
	e := evs[0]
	if e.A != vic.ID || e.B != fat.ID || e.N != StBlind || e.F != float32(g.Cfg.Skills.TranChauSecs) {
		t.Errorf("event sai: %+v", e)
	}
	if e.X != float32(fat.Pos.X) || e.Y != float32(fat.Pos.Y) {
		t.Errorf("X/Y phải là chỗ người gây (để vẽ mũi tên), có (%v,%v)", e.X, e.Y)
	}
}

// Hết giờ → đúng MỘT StatusEnd; bị xoá giữa chừng (Gồng Cơ) cũng báo hết.
func TestStatusEndOnExpiryAndOnClear(t *testing.T) {
	g := skillGame(t)
	p := put(g, "be_ti", Vec{X: 3000, Y: 3000}, 10, 20)
	g.ApplyStatus(p, StSleep, 1)
	g.DrainEvents()
	ends := 0
	for i := 0; i < 60; i++ {
		g.Time += 1.0 / 30
		g.updateSkillState(1.0 / 30)
		ends += len(statusEvents(g, "StatusEnd"))
	}
	if ends != 1 {
		t.Fatalf("ngủ 1s rồi hết: muốn 1 StatusEnd, có %d", ends)
	}

	g.ApplyStatus(p, StSleep, 5)
	giveCharge(p)
	p.Status[StSleep] = 0 // Gồng không dùng được lúc ngủ — xoá tay như cách Gồng/Gom Đơn xoá
	g.updateSkillState(1.0 / 30)
	evs := statusEvents(g, "StatusEnd")
	if len(evs) != 1 || evs[0].N != StSleep || evs[0].A != p.ID {
		t.Fatalf("bị xoá giữa chừng phải báo StatusEnd ngủ, có %+v", evs)
	}
}

// Đứng trong vũng vỏ chuối gia hạn mỗi tick — không được báo lại mỗi tick.
func TestStatusStartNotSpammedOnSmallExtension(t *testing.T) {
	g := skillGame(t)
	p := put(g, "a_hon", Vec{X: 3000, Y: 3000}, 10, 20)
	g.DrainEvents()
	for i := 0; i < 20; i++ {
		g.Time += 1.0 / 20
		g.ApplyStatus(p, StSlip, 1)
	}
	if n := len(statusEvents(g, "StatusStart")); n > 3 {
		t.Errorf("gia hạn từng chút 20 lần mà báo %d lần", n)
	}
}

// Đang gồng thì miễn khống chế → không báo gì cả.
func TestNoStatusStartWhenImmune(t *testing.T) {
	g := skillGame(t)
	p := put(g, "be_ti", Vec{X: 3000, Y: 3000}, 10, 20)
	g.ApplyStatus(p, StBrace, 4)
	g.DrainEvents()
	g.ApplyStatus(p, StSleep, 2)
	if n := len(statusEvents(g, "StatusStart")); n != 0 {
		t.Errorf("đang gồng mà vẫn báo %d StatusStart ngủ", n)
	}
}

// Chết: mọi trạng thái đã báo đều phải báo hết ngay, không chờ tick sau (người chết rời bàn).
func TestDeathEndsAllShownStatuses(t *testing.T) {
	g := skillGame(t)
	p := put(g, "a_hon", Vec{X: 3000, Y: 3000}, 10, 20)
	g.ApplyStatus(p, StBlind, 5)
	g.ApplyStatus(p, StStun, 1)
	g.DrainEvents()
	g.onDeathClear(p)
	if n := len(statusEvents(g, "StatusEnd")); n != 2 {
		t.Errorf("chết đang mù + choáng: muốn 2 StatusEnd, có %d", n)
	}
}

// Loa Phường: mỗi người bị đẩy nhận `push`; Ông Ba nằm võng nhận `steal`; đàn em nhận `minion`.
func TestDisplayOnlyStatuses(t *testing.T) {
	g := skillGame(t)
	loa := put(g, "ba_tam_loa", Vec{X: 3000, Y: 3000}, 10, 20)
	vic := put(g, "a_hon", Vec{X: 3100, Y: 3000}, 10, 20)
	g.DrainEvents()
	giveCharge(loa)
	g.CastSkill(loa)
	want := map[int]bool{}
	for _, e := range statusEvents(g, "StatusStart") {
		if e.A == vic.ID && e.B == loa.ID {
			want[e.N] = true
		}
	}
	if !want[StPush] {
		t.Error("người bị Loa đẩy không nhận StatusStart push")
	}

	ong := put(g, "ong_ba_vong", Vec{X: 6000, Y: 6000}, 10, 20)
	g.finishRuNgu(ong)
	got := false
	for _, e := range statusEvents(g, "StatusStart") {
		got = got || (e.A == ong.ID && e.N == StSteal && e.F == float32(g.skillNum("ru_ngu_steal_secs")))
	}
	if !got {
		t.Error("Ông Ba nằm võng không nhận StatusStart steal")
	}

	boss := put(g, "dai_ca_dep", Vec{X: 8000, Y: 8000}, 10, 20)
	m := g.spawnMinion(boss, Vec{X: 8200, Y: 8000}, 8)
	got = false
	for _, e := range statusEvents(g, "StatusStart") {
		got = got || (e.A == m.ID && e.B == boss.ID && e.N == StMinion && e.F == 8)
	}
	if !got {
		t.Error("đàn em không nhận StatusStart minion")
	}
	if EventStatusNames[StPush] != "push" || EventStatusNames[StBlind] != "blind" {
		t.Error("EventStatusNames lệch thứ tự")
	}
}
