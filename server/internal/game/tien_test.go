package game

import (
	"math"
	"testing"
)

// hitLand makes a land one body hit on b and returns whether it connected.
func hitLand(t *testing.T, g *Game, a, b *Player) {
	t.Helper()
	place(g, a, Vec{X: 6000, Y: 6000}, 0)
	place(g, b, Vec{X: 6000 + g.Geom(a).Inner, Y: 6000}, 0)
	b.InvulnUntil = 0
	before := a.Hits
	for i := 0; i < 90 && a.Hits == before; i++ {
		g.Tick(dt)
		if !b.Dead {
			place(g, b, Vec{X: 6000 + g.Geom(a).Inner, Y: 6000}, 0)
			b.InvulnUntil = 0
		}
	}
	if a.Hits == before {
		t.Fatal("không chém trúng được")
	}
}

// Spec §3.4: Đại Thừa tầng 10 chém trúng người đủ 10 lần → lên Chân Tiên.
func TestTienAscendsOnKillCount(t *testing.T) {
	g := quietGame(t, nil)
	top := g.Prog.TopMortalLevel()
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: top, Swords: 60, GuestID: "ga"})
	if !g.CanEarnTienKills(a) {
		t.Fatal("Đại Thừa tầng 10 phải bắt đầu đếm sát mạng")
	}
	need := g.Prog.TienEntry(0)
	for i := 0; i < int(need); i++ {
		v := g.AddPlayer(SpawnOptions{Name: "V", Level: 30, Swords: 1, GuestID: "v" + string(rune('a'+i))})
		hitLand(t, g, a, v)
	}
	if got := g.TienRealmOf(a); got != 0 {
		t.Fatalf("sau %v sát mạng vẫn ở Tiên giới %d, muốn Chân Tiên (0); kills = %v", need, got, a.TienKills)
	}
	if name := g.Prog.RealmName(g.Prog.Realm(a.Level)); name != "Chân Tiên" {
		t.Fatalf("tên cảnh giới = %q, muốn Chân Tiên", name)
	}
}

// Spec §3.4: chưa tới Đại Thừa tầng 10 thì chém trúng không được tính trước.
func TestTienKillsOnlyCountAtTopMortal(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: 50, Swords: 60, GuestID: "ga"})
	v := g.AddPlayer(SpawnOptions{Name: "V", Level: 30, Swords: 1, GuestID: "gv"})
	hitLand(t, g, a, v)
	if a.TienKills != 0 {
		t.Fatalf("sát mạng = %v khi mới Luyện Hư, muốn 0", a.TienKills)
	}
}

// Spec §3.4: nạn nhân là bot chỉ tính 0.5 sát mạng.
func TestTienBotVictimCountsHalf(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: g.Prog.TopMortalLevel(), Swords: 60, GuestID: "ga"})
	bot := g.AddPlayer(SpawnOptions{Name: "Bot", Level: 30, Swords: 1, IsBot: true})
	hitLand(t, g, a, bot)
	if want := g.Cfg.Cultivation.TienBotKillValue; a.TienKills != want {
		t.Fatalf("chém bot được %v sát mạng, muốn %v", a.TienKills, want)
	}
}

// Spec §3.4: chém lại cùng một người trong 60s không tính thêm sát mạng.
func TestTienSameVictimDoesNotCountTwice(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: g.Prog.TopMortalLevel(), Swords: 60, GuestID: "ga"})
	v := g.AddPlayer(SpawnOptions{Name: "V", Level: 40, Swords: 1, GuestID: "gv"})
	hitLand(t, g, a, v)
	first := a.TienKills
	hitLand(t, g, a, v)
	if a.TienKills != first {
		t.Fatalf("chém lại cùng người vẫn cộng sát mạng: %v → %v", first, a.TienKills)
	}
}

// Spec §3.4 ví dụ: Kim Tiên với 45 sát mạng bị chém → Chân Tiên, bộ đếm về 10.
func TestTienRealmDropResetsCounter(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Swords: 1})
	p.Level = g.Prog.TienLevel(1) // Kim Tiên
	p.TienKills = 45
	g.refreshDerived(p)
	g.dropTienRealm(p)
	if got := g.TienRealmOf(p); got != 0 {
		t.Fatalf("rớt xuống Tiên giới %d, muốn Chân Tiên (0)", got)
	}
	if p.TienKills != 10 {
		t.Fatalf("bộ đếm = %v, muốn về 10 (mức khởi điểm Chân Tiên)", p.TienKills)
	}
	// Chân Tiên bị chém nữa → về Đại Thừa tầng 10, hết Tiên giới.
	g.dropTienRealm(p)
	if g.TienRealmOf(p) != -1 || p.Level != g.Prog.TopMortalLevel() || p.TienKills != 0 {
		t.Fatalf("Chân Tiên rớt phải về Đại Thừa tầng 10: level %d, kills %v", p.Level, p.TienKills)
	}
}

// Spec §3.4: bảng số vòng / kiếm tối đa / scale của Tiên giới.
func TestTienRealmTable(t *testing.T) {
	g := quietGame(t, nil)
	want := []struct {
		name   string
		swords int
		rings  int
		scale  float64
	}{
		{"Chân Tiên", 550, 10, 5.2},
		{"Kim Tiên", 660, 11, 5.4},
		{"Thái Ất Ngọc Tiên", 780, 12, 5.6},
		{"Đại La Kim Tiên", 910, 13, 5.8},
		{"Đạo Tổ", 1050, 14, 6.0},
	}
	for i, w := range want {
		L := g.Prog.TienLevel(i)
		r := g.Prog.Realm(L)
		if got := g.Prog.RealmName(r); got != w.name {
			t.Errorf("Tiên %d tên %q, muốn %q", i, got, w.name)
		}
		if got := g.Prog.MaxSwords(r); got != w.swords {
			t.Errorf("%s kiếm tối đa %d, muốn %d", w.name, got, w.swords)
		}
		if got := g.Prog.Scale(L); math.Abs(got-w.scale) > 1e-9 {
			t.Errorf("%s scale %v, muốn %v", w.name, got, w.scale)
		}
		geom := ComputeRingGeom(g.Cfg.Sword, g.Cfg.Player.BodyR, 1, w.swords, 1)
		if geom.Rings != w.rings {
			t.Errorf("%s có %d vòng, muốn %d", w.name, geom.Rings, w.rings)
		}
	}
}

// EXP không bao giờ đẩy người chơi vào Tiên giới — chỉ sát mạng mới được.
func TestExpNeverReachesTien(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Swords: 1})
	g.GainExp(p, 1e12, false)
	if p.Level != g.Prog.TopMortalLevel() {
		t.Fatalf("EXP đẩy lên level %d, phải dừng ở Đại Thừa tầng 10 (%d)", p.Level, g.Prog.TopMortalLevel())
	}
	if g.TienRealmOf(p) != -1 {
		t.Fatal("EXP không được đưa ai vào Tiên giới")
	}
}

// Thanh Tiên khí hiện đúng "đang có / cần" của chặng hiện tại.
func TestTienProgressBar(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Swords: 1})
	p.Level = g.Prog.TienLevel(1) // Kim Tiên: vào ở 30, lên Thái Ất ở 60
	p.TienKills = 45
	g.refreshDerived(p)
	have, need := g.TienProgress(p)
	if have != 15 || need != 30 {
		t.Fatalf("Tiên khí %v/%v, muốn 15/30", have, need)
	}
	// Đạo Tổ là trần: không còn chặng nào nữa.
	p.Level = g.Prog.TienLevel(g.Prog.TienCount() - 1)
	g.refreshDerived(p)
	if _, need := g.TienProgress(p); need != 0 {
		t.Fatalf("Đạo Tổ vẫn còn chặng %v, muốn 0", need)
	}
}
