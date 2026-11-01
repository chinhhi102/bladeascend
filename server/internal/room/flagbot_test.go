package room

import (
	"testing"

	"bladeascend/server/internal/game"
	"bladeascend/server/internal/netproto"
)

// Spec v6.1 §9.4: gói snapshot không được lộ ai là bot.
func TestSnapshotNeverFlagsBots(t *testing.T) {
	m, _, _ := testManager(t)
	m.Join(NewClient(&recConn{}, false), netproto.Join{Name: "NguoiThat", Char: "a_hon"})
	r := m.AllRooms()[0]
	z := r.Zones()[0]
	b := r.spawnBotIn(z, 0, 1)
	var viewer *game.Player
	for _, p := range z.G.Players {
		if !p.IsBot {
			viewer = p
		}
	}
	if viewer == nil {
		viewer = b
	}
	if playerFlags(z.G, viewer, b)&netproto.FlagBot != 0 {
		t.Fatal("snapshot gửi cờ FlagBot — người chơi đọc gói là biết đâu là bot")
	}
}
