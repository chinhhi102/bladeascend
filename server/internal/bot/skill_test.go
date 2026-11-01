package bot

import (
	"math"
	"testing"

	"bladeascend/server/internal/config"
	"bladeascend/server/internal/game"
	"bladeascend/server/internal/world"
)

// Luật bot dùng kỹ năng theo spec §9 (v6.9). Mỗi bài dựng một trận trống, đặt bot và một
// đối thủ đúng chỗ, rồi hỏi `wantsSkill` — không chạy cả vòng AI.

func quiet(t *testing.T) *game.Game {
	t.Helper()
	cfg := config.MustLoadDefault()
	cfg.Room.MonsterCount = 0
	cfg.Items.CountPerRoom = 0
	cfg.Items.ChestEverySecs = 1e9
	cfg.ExpOrbs.Cell = 0
	return game.New(cfg, world.EmptyMap(cfg.Room.MapSize), 1, 0)
}

func put(g *game.Game, char string, swords int, at game.Vec) *game.Player {
	p := g.AddPlayer(game.SpawnOptions{Name: char, Char: char, IsBot: true, Swords: swords})
	p.Pos = at
	return p
}

func brainFor(g *game.Game, p *game.Player) *Brain { return New(p.ID, 7, g) }

func TestBotOngBaSleepsOnlyWhenClashing(t *testing.T) {
	g := quiet(t)
	c := game.Vec{X: 2000, Y: 2000}
	me := put(g, "ong_ba_vong", 10, c)
	b := brainFor(g, me)
	foe := put(g, "a_hon", 10, c.Add(game.Vec{X: 5000}))
	if b.wantsSkill(g, me, foe, foe, nil) {
		t.Fatal("địch ở xa mà Ông Ba đã ru ngủ")
	}
	foe.Pos = c.Add(game.Vec{X: g.Geom(me).Reach * 0.8})
	if !b.wantsSkill(g, me, foe, foe, nil) {
		t.Fatal("địch đang va kiếm trong 4 U mà Ông Ba không ru ngủ")
	}
}

func TestBotSuPhuAimsAtSwordRichFoe(t *testing.T) {
	g := quiet(t)
	c := game.Vec{X: 2000, Y: 2000}
	me := put(g, "su_phu_mi", 5, c)
	me.Facing = 0
	b := brainFor(g, me)
	poor := put(g, "a_hon", 2, c.Add(game.Vec{X: g.SkillU(me) * 2}))
	if b.wantsSkill(g, me, poor, poor, poor) {
		t.Fatal("địch ít kiếm hơn mình mà Sư Phụ vẫn hút mì")
	}
	rich := put(g, "be_ti", 12, c.Add(game.Vec{X: g.SkillU(me) * 2, Y: g.SkillU(me) * 0.5}))
	if !b.wantsSkill(g, me, rich, rich, nil) {
		t.Fatal("địch nhiều kiếm trong nón mà Sư Phụ không hút mì")
	}
	want := rich.Pos.Sub(me.Pos).Angle()
	if math.Abs(me.Facing-want) > 1e-6 {
		t.Fatalf("Sư Phụ không xoay mặt về mục tiêu: facing %.3f, muốn %.3f", me.Facing, want)
	}
}

func TestBotShipperGrabsNearbyTienKhi(t *testing.T) {
	g := quiet(t)
	c := game.Vec{X: 2000, Y: 2000}
	me := put(g, "shipper", 5, c)
	b := brainFor(g, me)
	if b.wantsSkill(g, me, nil, nil, nil) {
		t.Fatal("không có gì mà Shipper vẫn dùng Gom Đơn")
	}
	g.SpawnItemForTest(game.ItemTienKhi, 10, c.Add(game.Vec{X: g.SkillU(me) * 4}))
	if !b.wantsSkill(g, me, nil, nil, nil) {
		t.Fatal("Tiên khí rơi trong 8 U mà Shipper không dùng Gom Đơn")
	}
}

func TestBotShipperEscapesControl(t *testing.T) {
	g := quiet(t)
	me := put(g, "shipper", 5, game.Vec{X: 2000, Y: 2000})
	b := brainFor(g, me)
	g.ApplyStatus(me, game.StSlow, 2)
	if !b.wantsSkill(g, me, nil, nil, nil) {
		t.Fatal("bị khống chế mà Shipper không dùng kỹ năng thoát")
	}
}

func TestBotCoBaBlocksBehindWhenFleeing(t *testing.T) {
	g := quiet(t)
	c := game.Vec{X: 2000, Y: 2000}
	me := put(g, "co_ba_ganh", 5, c)
	b := brainFor(g, me)
	b.State = StateFlee
	threat := put(g, "be_ti", 20, c.Add(game.Vec{X: -g.SkillRangeOf(me, "bay_sap_dist")}))
	if !b.wantsSkill(g, me, threat, threat, nil) {
		t.Fatal("đang bị đuổi sát mà Cô Ba không bày sạp chặn")
	}
	if want := threat.Pos.Sub(me.Pos).Angle(); math.Abs(me.Facing-want) > 1e-6 {
		t.Fatalf("Cô Ba phải bày sạp về phía kẻ đuổi (sau lưng): facing %.3f, muốn %.3f", me.Facing, want)
	}
}

func TestBotCoBaIgnoresIdlePrey(t *testing.T) {
	g := quiet(t)
	c := game.Vec{X: 2000, Y: 2000}
	me := put(g, "co_ba_ganh", 10, c)
	b := brainFor(g, me)
	prey := put(g, "a_hon", 1, c.Add(game.Vec{X: g.SkillU(me)}))
	if b.wantsSkill(g, me, prey, nil, prey) {
		t.Fatal("con mồi đứng yên mà Cô Ba đã bày sạp")
	}
	prey.Vel = game.Vec{X: 200}
	if !b.wantsSkill(g, me, prey, nil, prey) {
		t.Fatal("con mồi yếu đang chạy mà Cô Ba không bày sạp chặn trước")
	}
}
