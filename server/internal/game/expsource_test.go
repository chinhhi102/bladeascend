package game

import "testing"

// EXP_AUDIT (CHANGES v6.21): mỗi lượng EXP phải đến tay `ExpTrace` kèm ĐÚNG nguồn — viên rơi
// từ quái và viên văng khi chém người trông giống hệt viên môi trường lúc nhặt.
func TestExpTraceTagsOrbSource(t *testing.T) {
	g := quietGame(t, nil)
	p := g.AddPlayer(SpawnOptions{Name: "A", Char: g.Cfg.Characters[0], Swords: 1})
	got := map[ExpSource]float64{}
	g.ExpTrace = func(_ *Player, src ExpSource, amount float64) { got[src] += amount }

	g.scatterOrbs(p.Pos, 20, 0, ExpOrbMonster)
	for _, it := range g.Items {
		g.pickup(p, it)
	}
	g.GainExpFrom(p, 7, false, ExpHitBot)

	if got[ExpOrbMonster] != 20 || got[ExpHitBot] != 7 || got[ExpOrbAmbient] != 0 {
		t.Fatalf("nguồn EXP ghi sai: %v; muốn quái 20, chém bot 7, môi trường 0", got)
	}
}
