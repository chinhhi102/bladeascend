package room

import (
	"testing"

	"bladeascend/server/internal/netproto"
)

// Spec v6.1 §9.4: một phần bot vừa vào trận đứng yên 3–8 s như người thật đang nhìn quanh.
// Ép xác suất = 1 để mọi bot đều đứng yên, rồi kiểm: 2 s đầu không nhúc nhích, qua 9 s thì
// có con đã đi.
func TestBotIdlesAfterJoin(t *testing.T) {
	m, _, cfg := testManager(t)
	cfg.Bots.IdleAfterJoinChance = 1
	cfg.Bots.IdleAfterJoinSecs = []float64{3, 8}
	m.Join(NewClient(&recConn{}, false), netproto.Join{Name: "NguoiThat", Char: "a_hon"})
	r := m.AllRooms()[0]
	z := r.Zones()[0]
	p := r.spawnBotIn(z, 0, 1)
	start := p.Pos
	dt := 1 / float64(cfg.Net.TickHz)
	for i := 0; i < int(2/dt); i++ {
		r.Step(dt)
	}
	if q, ok := z.G.Players[p.ID]; ok && !q.Dead && q.Pos.Dist(start) > 1 {
		t.Fatalf("bot vừa vào đã đi %.1f đơn vị trong 2 s đầu — phải đứng yên", q.Pos.Dist(start))
	}
	for i := 0; i < int(7/dt); i++ {
		r.Step(dt)
	}
	q, ok := z.G.Players[p.ID]
	if !ok || q.Dead {
		t.Skip("bot đã chết/rời trước khi kiểm được")
	}
	if q.Pos.Dist(start) < 1 {
		t.Fatal("qua 9 s bot vẫn đứng yên — khoảng đứng yên không hết hạn")
	}
}

// Xác suất 0 (tắt) thì không bot nào đứng yên.
func TestBotIdleOffByDefaultChanceZero(t *testing.T) {
	m, _, cfg := testManager(t)
	cfg.Bots.IdleAfterJoinChance = 0
	m.Join(NewClient(&recConn{}, false), netproto.Join{Name: "NguoiThat", Char: "a_hon"})
	r := m.AllRooms()[0]
	z := r.Zones()[0]
	p := r.spawnBotIn(z, 0, 1)
	if z.brains[p.ID].Idle(z.G.Time) {
		t.Fatal("xác suất 0 mà bot vẫn đứng yên")
	}
}
