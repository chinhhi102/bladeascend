package room

import (
	"testing"

	"bladeascend/server/internal/game"
	"bladeascend/server/internal/world"
)

// Spec §3.4: Đạo Tổ hiện trên minimap của cả phòng; Ẩn Thân thì giấu (giống Top 1).
func TestDaoToShowsOnMinimap(t *testing.T) {
	_, _, cfg := testManager(t)
	g := game.New(cfg, world.EmptyMap(cfg.Room.MapSize), 1, 0)
	p := g.AddPlayer(game.SpawnOptions{Name: "T", Char: "a_hon", Swords: 3})
	if n := len(daoToMarks(g)); n != 0 {
		t.Fatalf("người mới vào đã hiện %d chấm Đạo Tổ", n)
	}
	p.Level = g.Prog.TienLevel(g.Prog.TienCount() - 1)
	p.Bush = -1
	if n := len(daoToMarks(g)); n != 1 {
		t.Fatalf("Đạo Tổ không hiện trên minimap (%d chấm)", n)
	}
	p.Buffs[game.BuffStealth] = g.Time + 10
	if n := len(daoToMarks(g)); n != 0 {
		t.Fatal("Đạo Tổ đang Ẩn Thân mà vẫn lộ trên minimap")
	}
}
