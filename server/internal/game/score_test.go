package game

import (
	"math"
	"testing"
)

// hitOnce places a next to b and ticks until a's ring lands one body hit.
func hitOnce(t *testing.T, g *Game, a, b *Player) {
	t.Helper()
	place(g, a, Vec{X: 6000, Y: 6000}, 0)
	place(g, b, Vec{X: 6000 + g.Geom(a).Inner, Y: 6000}, 0)
	before := a.Hits
	for i := 0; i < 90 && a.Hits == before; i++ {
		g.Tick(dt)
	}
	if a.Hits == before {
		t.Fatal("không chém trúng được sau 3 giây")
	}
}

// revive: nạn nhân hồi lại đúng cấp cũ để bị chém lần nữa (test chống cày điểm).
func revive(g *Game, b *Player) {
	g.Respawn(b, b.Level, 1)
	b.InvulnUntil = 0
}

// Spec §11.1 (v6.26): Trúc Cơ (r=1) chém trúng Nguyên Anh (r=3) → 100 × 4 × 2 = 800, và vì
// chém trúng nay luôn là chém chết nên +50 → 850 (CHANGES_v6_26 mục 4).
func TestScoreUpsetExample(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: 10, Swords: 20, GuestID: "ga"}) // Trúc Cơ 1
	b := g.AddPlayer(SpawnOptions{Name: "B", Level: 30, Swords: 1, GuestID: "gb"})  // Nguyên Anh 1
	hitOnce(t, g, a, b)
	// Chỉ xét nguồn "kill": EXP cướp được có thể đẩy A đột phá, cộng thêm điểm nguồn khác.
	if math.Abs(a.ScoreBy[ScoreKill]-850) > 1e-9 {
		t.Fatalf("điểm chém chết = %v, muốn 850", a.ScoreBy[ScoreKill])
	}
}

// Spec §11.1 ví dụ: Đại Thừa chém chết Luyện Khí → 100 × 1 × 1 + 50 = 150.
func TestScoreKillNoUpset(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: 80, Swords: 20, GuestID: "ga"}) // Đại Thừa 1
	b := g.AddPlayer(SpawnOptions{Name: "B", Level: 0, Swords: 1, GuestID: "gb"})   // Luyện Khí 1
	hitOnce(t, g, a, b)
	if !b.Dead {
		t.Fatal("Luyện Khí bị chém vào thân phải chết")
	}
	// Chỉ xét điểm từ nguồn "kill": A đang là Top 1 nên còn cộng thêm điểm giữ ngôi mỗi tick.
	if math.Abs(a.ScoreBy[ScoreKill]-150) > 1e-9 {
		t.Fatalf("điểm chém chết = %v, muốn 150", a.ScoreBy[ScoreKill])
	}
}

// Spec §11.1: chém lại cùng một người trong same_victim_cooldown_secs → 0 điểm.
func TestScoreSameVictimCooldown(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: 10, Swords: 20, GuestID: "ga"})
	b := g.AddPlayer(SpawnOptions{Name: "B", Level: 30, Swords: 1, GuestID: "gb"})
	hitOnce(t, g, a, b)
	first := a.ScoreBy[ScoreKill]
	revive(g, b)
	hitOnce(t, g, a, b)
	if a.ScoreBy[ScoreKill] != first {
		t.Fatalf("chém lại trong %vs vẫn cộng điểm: %v → %v",
			g.Cfg.Score.SameVictimCooldown, first, a.ScoreBy[ScoreKill])
	}
	if a.Hits != 2 {
		t.Fatalf("số lần chém trúng = %d, muốn 2 (bảng Sát Thần vẫn đếm)", a.Hits)
	}
	// Hết cooldown thì tính điểm lại.
	a.HitCooldown[victimKey(b)] = g.Time - 1
	revive(g, b)
	hitOnce(t, g, a, b)
	if a.ScoreBy[ScoreKill] <= first {
		t.Fatalf("hết cooldown vẫn không cộng điểm: %v", a.ScoreBy[ScoreKill])
	}
}

// Spec §11.1: nạn nhân là bot → 50% điểm.
func TestScoreBotVictimHalved(t *testing.T) {
	g := quietGame(t, nil)
	a := g.AddPlayer(SpawnOptions{Name: "A", Level: 10, Swords: 20, GuestID: "ga"})
	bot := g.AddPlayer(SpawnOptions{Name: "Bot", Level: 30, Swords: 1, IsBot: true})
	hitOnce(t, g, a, bot)
	if math.Abs(a.ScoreBy[ScoreKill]-425) > 1e-9 {
		t.Fatalf("điểm chém bot = %v, muốn 425 (một nửa của 850)", a.ScoreBy[ScoreKill])
	}
}

// Spec §11.1: đột phá cảnh giới mới được 20 × r, mỗi cảnh giới chỉ một lần trong mạng đó.
func TestScoreBreakthroughOncePerRealm(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "P", Level: 9, Swords: 1}) // Luyện Khí tầng 10
	g.GainExp(p, g.Prog.ExpNeeded(9), false)                       // → Trúc Cơ, r = 1
	if g.Prog.Realm(p.Level) != 1 {
		t.Fatalf("chưa lên Trúc Cơ: level %d", p.Level)
	}
	want := g.Cfg.Score.BreakthroughPerRealm
	if math.Abs(p.ScoreBy[ScoreBreakthrough]-want) > 1e-9 {
		t.Fatalf("điểm đột phá = %v, muốn %v", p.ScoreBy[ScoreBreakthrough], want)
	}
	// Rớt lại rồi lên lại cùng cảnh giới → không cộng thêm.
	g.setTotalExp(p, g.Prog.Threshold(5))
	g.setTotalExp(p, g.Prog.Threshold(12))
	if math.Abs(p.ScoreBy[ScoreBreakthrough]-want) > 1e-9 {
		t.Fatalf("lên lại cùng cảnh giới vẫn cộng điểm: %v", p.ScoreBy[ScoreBreakthrough])
	}
}

// Spec §11.1: giữ Top 1 phòng → +top1_per_sec mỗi giây, và đếm top1_secs.
func TestScoreTop1PerSecond(t *testing.T) {
	g := quietGame(t, nil)
	lead := g.AddPlayer(SpawnOptions{Name: "Lead", Level: 40, Swords: 1})
	g.AddPlayer(SpawnOptions{Name: "Other", Level: 5, Swords: 1})
	for i := 0; i < int(1/dt); i++ {
		g.Tick(dt)
	}
	if math.Abs(lead.Top1Secs-1) > 2*dt {
		t.Fatalf("top1_secs = %v, muốn ≈1", lead.Top1Secs)
	}
	if math.Abs(lead.ScoreBy[ScoreTop1]-g.Cfg.Score.Top1PerSec) > 0.05 {
		t.Fatalf("điểm giữ Top 1 = %v, muốn ≈%v", lead.ScoreBy[ScoreTop1], g.Cfg.Score.Top1PerSec)
	}
}

// Spec §11.2: bảng phòng xếp theo Uy Danh, KHÔNG theo cảnh giới; total_exp chỉ phá hòa.
func TestRankingByScoreNotRealm(t *testing.T) {
	g := quietGame(t, nil)
	rich := g.AddPlayer(SpawnOptions{Name: "CảnhGiớiCao", Level: 60, Swords: 1})
	fighter := g.AddPlayer(SpawnOptions{Name: "ÍtTuViNhiềuĐiểm", Level: 3, Swords: 1})
	g.creditScore(fighter, ScoreHit, 500)
	r := g.Ranking()
	if r[0] != fighter {
		t.Fatalf("đứng đầu là %s, phải là người nhiều Uy Danh hơn", r[0].Name)
	}
	if g.RankOf(fighter) != 1 || g.RankOf(rich) != 2 {
		t.Fatalf("hạng = %d, %d; muốn 1, 2", g.RankOf(fighter), g.RankOf(rich))
	}
	// Hòa điểm thì mới so tu vi.
	g.creditScore(rich, ScoreHit, 500)
	if r := g.Ranking(); r[0] != rich {
		t.Fatalf("hòa điểm thì người nhiều EXP phải lên trước, đang là %s", r[0].Name)
	}
}

// Spec §11.3: phải giữ hạng mới liên tục title_hold_secs mới đổi danh hiệu.
func TestTitleNeedsHold(t *testing.T) {
	g := quietGame(t, nil)
	first := g.AddPlayer(SpawnOptions{Name: "First", Level: 40, Swords: 1})
	second := g.AddPlayer(SpawnOptions{Name: "Second", Level: 20, Swords: 1})
	for i := 0; i < int(g.Cfg.TitleHold/dt)+2; i++ {
		g.Tick(dt)
	}
	if first.Title != 1 || second.Title != 2 {
		t.Fatalf("danh hiệu = %d, %d; muốn 1, 2", first.Title, second.Title)
	}
	// Second vượt lên nhưng chưa đủ 3 giây → danh hiệu chưa đổi.
	g.setTotalExp(second, g.Prog.Threshold(60))
	for i := 0; i < int(g.Cfg.TitleHold/dt)-2; i++ {
		g.Tick(dt)
	}
	if second.Title != 2 {
		t.Fatalf("đổi danh hiệu quá sớm: %d", second.Title)
	}
	for i := 0; i < 4; i++ {
		g.Tick(dt)
	}
	if second.Title != 1 || first.Title != 2 {
		t.Fatalf("sau %vs danh hiệu phải đổi: %d, %d", g.Cfg.TitleHold, second.Title, first.Title)
	}
}
