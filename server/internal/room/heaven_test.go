package room

import (
	"testing"

	"kiemxoay/server/internal/game"
)

// Thiên Môn là cổng MỘT CHIỀU luôn ra ở Thiên Đình (spec §8.4a), khác hẳn 4 cổng thường
// vốn bắn ngẫu nhiên khắp thế giới.
func TestHeavenGateAlwaysLeadsToHeaven(t *testing.T) {
	m, cfg := zoneManager(t)
	joinN(m, 1)
	r := m.AllRooms()[0]
	settle(r, cfg)
	if len(r.Zones()) < 2 {
		t.Skip("phòng chỉ có một khu")
	}
	// Dựng một Thiên Môn ở khu 2 và một bến cho nó ở khu 0.
	z := r.Zone(2)
	z.G.Map.Gates = append(z.G.Map.Gates, worldGate(900, z.G.Map.Size/2, z.G.Map.Size/2, true, false, 0))
	heaven := r.Zone(0)
	heaven.G.Map.Gates = append(heaven.G.Map.Gates, worldGate(901, 1000, 2000, false, true, 2))

	d, ok := r.routeGate(2, z.G.Map.Gates[len(z.G.Map.Gates)-1])
	if !ok {
		t.Fatal("Thiên Môn không chọn được đích")
	}
	if d.Zone != 0 {
		t.Fatalf("Thiên Môn đưa tới khu %d, phải luôn là Thiên Đình (khu 0)", d.Zone)
	}
	if d.At.X != 1000 || d.At.Y != 2000 {
		t.Fatalf("ra ở (%.0f,%.0f), phải đúng Bến Thiên Môn (1000,2000)", d.At.X, d.At.Y)
	}
	// Chạy nhiều lần: đích phải CỐ ĐỊNH, không ngẫu nhiên như cổng thường.
	for i := 0; i < 50; i++ {
		again, _ := r.routeGate(2, z.G.Map.Gates[len(z.G.Map.Gates)-1])
		if again.Zone != d.Zone || again.At != d.At {
			t.Fatal("Thiên Môn ra chỗ khác nhau giữa các lần — phải là đích cố định")
		}
	}
}

// Cổng thường KHÔNG bao giờ ném người vào Thiên Môn hay Bến Thiên Môn (spec §8.4a): hai
// thứ đó không nằm trong bể ngẫu nhiên.
func TestNormalGatesSkipHeavenGates(t *testing.T) {
	m, cfg := zoneManager(t)
	joinN(m, 1)
	r := m.AllRooms()[0]
	settle(r, cfg)
	z := r.Zone(1)
	z.G.Map.Gates = append(z.G.Map.Gates, worldGate(950, 5000, 5000, true, false, 0))
	heaven := r.Zone(0)
	heaven.G.Map.Gates = append(heaven.G.Map.Gates, worldGate(951, 1000, 2000, false, true, 1))

	normal := r.Zone(0).G.Map.Gates[0]
	for i := 0; i < 500; i++ {
		d, ok := r.routeGate(0, normal)
		if !ok {
			t.Fatal("cổng thường không chọn được đích")
		}
		if d.GateID == 950 || d.GateID == 951 {
			t.Fatal("cổng thường ném người vào Thiên Môn / Bến Thiên Môn")
		}
	}
}

// Bến Thiên Môn chỉ là chỗ đặt chân: đứng lên đó không được hút đi tiếp, nếu không người
// vừa lên Thiên Đình sẽ bị bắn ngược về khu khác ngay lập tức.
func TestDockDoesNotTeleport(t *testing.T) {
	m, cfg := zoneManager(t)
	c := joinN(m, 1)[0]
	r := m.AllRooms()[0]
	settle(r, cfg)
	z := r.ZoneOf(c.PlayerID)
	z.G.Map.Gates = append(z.G.Map.Gates, worldGate(960, 3000, 3000, false, true, 1))
	p := z.G.Players[c.PlayerID]
	p.Pos = game.Vec{X: 3000, Y: 3000}
	p.GateReadyAt = 0
	dt := 1 / float64(cfg.Net.TickHz)
	for i := 0; i < 20; i++ {
		r.Step(dt)
	}
	if r.ZoneOf(c.PlayerID) != z {
		t.Fatal("đứng trên Bến Thiên Môn mà bị hút sang khu khác")
	}
}
