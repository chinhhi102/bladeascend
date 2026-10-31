package game

import (
	"math"

	"bladeascend/server/internal/geom"
	"bladeascend/server/internal/world"
)

// GateDest là đích của một lần dịch chuyển: khu nào, cổng nào, ở toạ độ nào.
type GateDest struct {
	Zone   int
	GateID int
	At     Vec
}

// checkGate starts a teleport when a player's centre enters a gate (spec §8.4).
func (g *Game) checkGate(p *Player) {
	if p.Transit != nil || g.Time < p.GateReadyAt || len(g.Map.Gates) == 0 {
		return
	}
	// Đàn em của Đại Ca không qua cổng (spec v5.4 §6.3). Cho đi thì chúng rơi sang khu khác
	// và sống tiếp ở đó mà không còn đại ca — vừa sai luật vừa thành NPC mồ côi lang thang.
	if p.IsMinion {
		return
	}
	gc := g.Cfg.Gates
	for _, gate := range g.Map.Gates {
		// Bến Thiên Môn chỉ là chỗ ĐẶT CHÂN, không phải cổng đi: đứng lên bến mà bị hút đi
		// thì người vừa lên Thiên Đình sẽ bị bắn ngược về khu khác ngay lập tức.
		if gate.Dock {
			continue
		}
		radius := gc.Radius
		if gate.Heaven {
			radius = gc.HeavenRadius // Thiên Môn to gấp đôi (spec §8.4a)
		}
		if p.Pos.Dist(Vec{X: gate.X, Y: gate.Y}) > radius {
			continue
		}
		d, ok := g.gateExit(gate)
		if !ok {
			return
		}
		half := gc.TransitSecs / 2
		p.Transit = &Transit{Dest: d.At, DestGate: d.GateID, DestZone: d.Zone, OutAt: g.Time + half, DoneAt: g.Time + gc.TransitSecs}
		g.setMood(p, MoodTeleportOut, half)
		g.emit(Event{Kind: "TeleportOut", A: p.ID, B: uint32(d.GateID), N: gate.ID, X: float32(gate.X), Y: float32(gate.Y)})
		// Người trong Top 3 bước qua Thiên Môn thì cả phòng biết (spec §8.4a) — đó là tin
		// đáng loan: khu trung tâm sắp có biến.
		if gate.Heaven && p.Title >= 1 && p.Title <= 3 {
			g.emit(Event{Kind: "KillFeed", Global: true, S: "heaven_gate", A: p.ID, N: p.Title})
		}
		return
	}
}

// gateExit chọn cổng ra.
//
// Mặc định là **ngẫu nhiên**: bước vào một cổng thì ra ở một cổng bất kỳ khác cùng mạng,
// nên không ai đoán trước được mình rơi vào đâu và không thể dùng cổng làm đường tẩu thoát
// quen thuộc. Đặt `gates.random_exit: false` trong config để quay về kiểu cũ (cổng đôi cố
// định, ra cổng nào thì quay lại đúng cổng đó).
func (g *Game) gateExit(gate world.Gate) (GateDest, bool) {
	// Thế giới 5 khu: cả 20 cổng nằm trong MỘT mạng chung (spec §8.4), nên đích có thể ở
	// khu khác. Phòng cài `GateRouter` để chọn giúp — khu không biết gì về khu khác.
	if g.GateRouter != nil {
		if d, ok := g.GateRouter(g.ZoneIndex, gate); ok {
			return d, true
		}
	}
	if !g.Cfg.Gates.RandomExit {
		d, ok := g.Map.GateExit(gate)
		return GateDest{Zone: g.ZoneIndex, GateID: d.ID, At: Vec{X: d.X, Y: d.Y}}, ok
	}
	net := g.Map.GateNetwork(gate)
	if len(net) == 0 {
		return GateDest{}, false
	}
	d := net[g.Rng.Intn(len(net))]
	return GateDest{Zone: g.ZoneIndex, GateID: d.ID, At: Vec{X: d.X, Y: d.Y}}, true
}

// updateTransit moves players through gates: warn flash, reappear, end immunity.
func (g *Game) updateTransit(p *Player) {
	t := p.Transit
	if t == nil {
		return
	}
	gc := g.Cfg.Gates
	if !t.WarnSent && g.Time >= t.OutAt-gc.ExitWarnSecs {
		t.WarnSent = true
		g.emit(Event{Kind: "GateWarn", N: t.DestGate, X: float32(t.Dest.X), Y: float32(t.Dest.Y)})
	}
	if !t.Arrived && g.Time >= t.OutAt {
		t.Arrived = true
		// Exit just outside the gate radius so the player does not re-enter immediately.
		off := geom.FromAngle(g.Rng.Float64() * 2 * math.Pi).Mul(gc.Radius + g.BodyR(p) + 10)
		if t.DestZone != g.ZoneIndex && g.MoveToZone != nil {
			// Đích ở KHU KHÁC: phòng bê nhân vật sang ở ranh giới tick. Chỗ đặt chân tính
			// theo bản đồ khu đích, nên không gọi resolveBlocking của khu này.
			g.MoveToZone(p, t.DestZone, t.Dest.Add(off))
			p.lastTeleportIn = g.Time
			g.setMood(p, MoodTeleportIn, gc.TransitSecs/2)
			return
		}
		p.Pos = g.resolveBlocking(t.Dest.Add(off), g.BodyR(p))
		p.lastTeleportIn = g.Time
		g.setMood(p, MoodTeleportIn, gc.TransitSecs/2)
		g.emit(Event{Kind: "TeleportIn", A: p.ID, N: t.DestGate, X: float32(p.Pos.X), Y: float32(p.Pos.Y)})
	}
	if g.Time >= t.DoneAt {
		p.Transit = nil
		p.GateReadyAt = g.Time + gc.CooldownSecs
	}
}
