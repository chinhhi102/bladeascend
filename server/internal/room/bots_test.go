package room

import (
	"testing"

	"bladeascend/server/internal/bot"
	"bladeascend/server/internal/game"
)

// Spec v6.22 §9.3 / §17: bot sinh NHƯ NGƯỜI MỚI — Luyện Khí 1, 1 kiếm, không buff — kể cả
// khi phòng đang có người đứng đầu ở Đạo Tổ (bản cũ chép cảnh giới người hạng 11).
func TestBotSpawnsAsNewPlayer(t *testing.T) {
	m, _, _ := testManager(t)
	cs := joinN(m, 1)
	r := m.AllRooms()[0]
	r.Step(1.0 / 30) // người chơi chỉ thật sự vào bàn cờ sau một nhịp Step
	prog := r.zones[0].G.Prog
	forceTopPlayer(t, r, cs[0].PlayerID, prog.TienLevel(prog.TienCount()-1))
	for i := 0; i < 20; i++ {
		p := r.spawnBot()
		if p.Level != 0 || p.Ring.Count != 1 {
			t.Fatalf("bot sinh ở level %d với %d kiếm, muốn Luyện Khí 1 + 1 kiếm", p.Level, p.Ring.Count)
		}
		for b, until := range p.Buffs {
			if until > r.zones[0].G.Time {
				t.Fatalf("bot vừa sinh đã có buff %v", b)
			}
		}
	}
}

// forceTopPlayer đẩy một người lên cảnh giới `level` VÀ lên hạng 1 (ghim cả Uy Danh: phòng
// test có sẵn bot, chỉ nâng cảnh giới thì hạng 1 vẫn có thể là bot).
func forceTopPlayer(t *testing.T, r *Room, id uint32, level int) {
	t.Helper()
	p := r.Zone(0).G.Players[id]
	if p == nil {
		t.Fatalf("không tìm thấy người chơi %d trong khu 0", id)
	}
	r.Zone(0).G.ForceState(p, level, 20, p.Pos)
	p.Score = 1e9
}

// §9.3 v6.28: tự động tắt khi người thật ≥ ngưỡng (production 40), bật lại khi < ngưỡng − 5
// — ở giữa thì giữ trạng thái cũ (trễ), để phòng quanh ngưỡng không bật/tắt bot liên tục.
func TestBotAutoHysteresis(t *testing.T) {
	m, _, cfg := testManager(t)
	joinN(m, 1)
	r := m.AllRooms()[0]
	r.latchSettings()
	if r.set.BotAutoOffAt != 40 || cfg.Bots.AutoOffHysteresis != 5 {
		t.Fatalf("mặc định production 40, trễ 5 đã đổi: %d / %d", r.set.BotAutoOffAt, cfg.Bots.AutoOffHysteresis)
	}
	steps := []struct {
		humans int
		want   bool
	}{{10, true}, {39, true}, {40, false}, {38, false}, {35, false}, {34, true}, {38, true}}
	for _, s := range steps {
		r.applyBotSwitch(s.humans)
		if r.botsOn != s.want {
			t.Fatalf("%d người: botsOn = %v, muốn %v", s.humans, r.botsOn, s.want)
		}
	}
}

// Chế độ bật/tắt tay thắng chế độ tự động; tắt một phòng không đụng phòng khác; tắt toàn
// server thì phòng có ghi đè "on" vẫn bật.
func TestBotModeOverrides(t *testing.T) {
	m, _, _ := testManager(t)
	joinN(m, 1)
	a := m.AllRooms()[0]
	m.mu.Lock()
	b := m.createLocked()
	m.mu.Unlock()
	a.botOverride = BotOff
	a.applyBotSwitch(1)
	b.applyBotSwitch(1)
	if a.botsOn || !b.botsOn {
		t.Fatalf("tắt phòng A: A=%v B=%v, muốn A tắt, B bật", a.botsOn, b.botsOn)
	}
	if err := m.SetBotMode(BotOff); err != nil {
		t.Fatal(err)
	}
	b.botOverride = BotOn
	a.botOverride = BotFollow
	a.applyBotSwitch(1)
	b.applyBotSwitch(1)
	if a.botsOn || !b.botsOn {
		t.Fatalf("toàn server tắt, B ghi đè bật: A=%v B=%v", a.botsOn, b.botsOn)
	}
	a.botOverride = BotOn
	a.applyBotSwitch(100)
	if !a.botsOn {
		t.Fatal("bật tay mà vẫn bị chế độ tự động tắt vì đông người")
	}
	if m.SetBotMode("bogus") == nil || ValidBotMode(BotFollow, false) {
		t.Fatal("chế độ lạ / \"\" cho toàn server phải bị từ chối")
	}
}

// Bot tắt: không sinh mới, và bot chết thì rời hẳn chứ không hồi sinh.
func TestBotsOffNoSpawnNoRespawn(t *testing.T) {
	m, _, _ := testManager(t)
	joinN(m, 1)
	r := m.AllRooms()[0]
	stepSecs(r, 3) // để updateBots sinh lứa bot đầu
	if r.Bots() == 0 {
		t.Fatal("thiết lập sai: chưa có bot nào")
	}
	r.botOverride = BotOff
	stepSecs(r, 0.1)
	before := r.Bots()
	z := r.zones[0]
	var victim *game.Player
	for id := range z.brains {
		victim = z.G.Players[id]
		break
	}
	r.onBotDeath(z, victim)
	if r.Bots() != before-1 || z.respawnQueued(victim.ID) {
		t.Fatalf("bot tắt mà bot chết vẫn chờ hồi sinh (bot %d → %d)", before, r.Bots())
	}
	stepSecs(r, 5)
	if r.Bots() > before-1 {
		t.Fatalf("bot tắt mà vẫn sinh thêm: %d → %d", before-1, r.Bots())
	}
}

// "Rút bot ngay" (§9.5 / §17): bot rảnh rời ngay, bot đang giao tranh ở lại tới khi rảnh.
func TestWithdrawSkipsBotsInCombat(t *testing.T) {
	m, _, _ := testManager(t)
	joinN(m, 1)
	r := m.AllRooms()[0]
	r.Step(1.0 / 30)
	z := r.zones[0]
	for id := range z.brains { // dọn bot có sẵn cho bài test gọn
		z.G.RemovePlayer(id)
		delete(z.brains, id)
	}
	idle, busy := r.spawnBot(), r.spawnBot()
	z.brains[idle.ID].State, z.brains[busy.ID].State = bot.StateWander, bot.StateWander
	busy.CombatAt = z.G.Time
	r.withdraw = map[uint32]bool{idle.ID: true, busy.ID: true}
	r.withdrawIdle()
	if z.G.Players[idle.ID] != nil {
		t.Fatal("bot rảnh không rời")
	}
	if z.G.Players[busy.ID] == nil {
		t.Fatal("bot đang giao tranh bị rút giữa trận")
	}
	z.G.Time += r.cfg.Bots.WithdrawIdleSecs + 0.1
	r.withdrawIdle()
	if z.G.Players[busy.ID] != nil || len(r.withdraw) != 0 {
		t.Fatal("bot hết giao tranh mà vẫn chưa rút")
	}
}
