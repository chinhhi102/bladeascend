package room

import (
	"strings"
	"testing"
	"time"

	"bladeascend/server/internal/safe"
)

// GAME_SPEC §13.8 + checklist §17: "panic giả trong tick 1 khu → khu đó bỏ 1 tick, các khu
// khác không ảnh hưởng".
func TestPanicTickMotKhuKhongAnhHuongKhuKhac(t *testing.T) {
	var caught []safe.Info
	safe.OnPanic(func(i safe.Info) { caught = append(caught, i) })
	defer safe.OnPanic(nil)

	m, cfg := zoneManager(t)
	joinN(m, 1)
	r := m.AllRooms()[0]
	settle(r, cfg)
	if len(r.Zones()) < 2 {
		t.Skip("cần phòng nhiều khu")
	}
	before := make([]uint32, len(r.Zones()))
	for i, z := range r.Zones() {
		before[i] = z.tick
	}
	r.Zones()[1].panicNextTick.Store(true)

	dt := 1 / float64(cfg.Net.TickHz)
	r.Step(dt) // tick có lỗi giả

	if len(caught) != 1 {
		t.Fatalf("bắt được %d panic, muốn 1 (%+v)", len(caught), caught)
	}
	if caught[0].Boundary != "tick" || !strings.Contains(caught[0].Detail, r.Zones()[1].Name) {
		t.Fatalf("bối cảnh panic không nói rõ khu nào: %+v", caught[0])
	}
	for i, z := range r.Zones() {
		want := before[i] + 1
		if i == 1 {
			want = before[i] // khu lỗi BỎ tick đó
		}
		if z.tick != want {
			t.Fatalf("khu %d (%s): tick %d, muốn %d", i, z.Name, z.tick, want)
		}
	}
	// Tick sau phòng chạy tiếp bình thường, kể cả khu vừa lỗi (nó chỉ mất đúng 1 tick).
	r.Step(dt)
	for i, z := range r.Zones() {
		want := before[i] + 2
		if i == 1 {
			want = before[i] + 1
		}
		if z.tick != want {
			t.Fatalf("khu %d không chạy tiếp sau lỗi: tick %d, muốn %d", i, z.tick, want)
		}
	}
	if r.closed {
		t.Fatal("phòng không được đóng vì một tick lỗi")
	}
}

// Panic ≥ 3 lần/phút ở một khu → đánh dấu khu "suy giảm" (§13.8).
func TestKhuSuyGiamSauBaPanicTrongMotPhut(t *testing.T) {
	safe.OnPanic(func(safe.Info) {})
	defer safe.OnPanic(nil)

	m, cfg := zoneManager(t)
	joinN(m, 1)
	r := m.AllRooms()[0]
	settle(r, cfg)
	z := r.Zones()[0]
	now := time.Now()
	if z.Health(now).Degraded {
		t.Fatal("khu mới dựng mà đã suy giảm")
	}
	dt := 1 / float64(cfg.Net.TickHz)
	for i := 0; i < 2; i++ {
		z.panicNextTick.Store(true)
		r.Step(dt)
	}
	if z.Health(time.Now()).Degraded {
		t.Fatal("2 lần chưa được coi là suy giảm")
	}
	z.panicNextTick.Store(true)
	r.Step(dt)
	h := z.Health(time.Now())
	if !h.Degraded || h.Panics != 3 {
		t.Fatalf("sau 3 lần: %+v", h)
	}
	// Im 5 phút thì cờ tự tắt — cờ đỏ vĩnh viễn là cờ không ai còn nhìn.
	if z.Health(time.Now().Add(6 * time.Minute)).Degraded {
		t.Fatal("cờ suy giảm phải tự tắt sau 5 phút không panic")
	}
}

// Việc trong hộp thư phòng cũng là một ranh giới: một việc hỏng không làm chết cả tick.
func TestPanicTrongHopThuKhongLamChetTick(t *testing.T) {
	var caught []safe.Info
	safe.OnPanic(func(i safe.Info) { caught = append(caught, i) })
	defer safe.OnPanic(nil)

	m, _, cfg := testManager(t)
	joinN(m, 1)
	r := m.AllRooms()[0]
	settle(r, cfg)
	before := r.Zones()[0].tick
	r.Do(func() { panic("việc hỏng") })
	r.Step(1 / float64(cfg.Net.TickHz))
	if len(caught) != 1 || caught[0].Value != "việc hỏng" {
		t.Fatalf("%+v", caught)
	}
	if r.Zones()[0].tick != before+1 {
		t.Fatal("khu phải vẫn tick dù một việc trong hộp thư hỏng")
	}
}
